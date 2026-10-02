"""Failure and isolation contracts; no daemon, provider, or credentials needed."""
import contextlib
import importlib.util
import io
import json
import os
from pathlib import Path
import signal
import stat
import subprocess
import tempfile
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location("restore", Path(__file__).with_name("rehearse_postgres_restore.py"))
restore = importlib.util.module_from_spec(spec)
spec.loader.exec_module(restore)
ROOT = Path(__file__).resolve().parents[2]


class RehearsalContracts(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.directory = Path(self.temporary.name)
        self.runner = restore.Rehearsal(ROOT, self.directory / "report")
        self.runner.temp_dir = self.directory
        self.old_umask = os.umask(0o022)
        self.addCleanup(os.umask, self.old_umask)

    def cli(self, *args):
        output = io.StringIO()
        with contextlib.redirect_stdout(output), contextlib.redirect_stderr(output):
            try:
                status = restore.main(list(args))
            except SystemExit as error:
                status = error.code
        return status, output.getvalue()

    def test_existing_output_and_symlink_are_preserved_without_work(self):
        original = self.directory / "original"
        original.mkdir()
        sentinel = original / "report.json"
        sentinel.write_text("previous evidence")
        alias = self.directory / "alias"
        alias.symlink_to(original, target_is_directory=True)
        with patch.object(restore.Rehearsal, "run") as run:
            for path in (original, alias):
                self.assertEqual(self.cli("--report-dir", str(path))[0], 2)
            run.assert_not_called()
        self.assertEqual(sentinel.read_text(), "previous evidence")

    def test_connection_archive_and_arbitrary_fault_arguments_cannot_start_work(self):
        with patch.object(restore.Rehearsal, "run") as run:
            for args in (("--host", "production.invalid"), ("--dsn", "postgres://example.invalid/db"),
                         ("--archive", "external.dump"), ("--fault", "arbitrary-sql")):
                self.assertEqual(self.cli("--report-dir", str(self.directory / "new"), *args)[0], 2)
            run.assert_not_called()
        self.assertFalse((self.directory / "new").exists())

    def test_rejected_connection_value_is_not_echoed(self):
        status, output = self.cli("--report-dir", str(self.directory / "new"),
                                  "--dsn", "postgres://synthetic-secret@remote.invalid/db")
        self.assertEqual(status, 2)
        self.assertNotIn("synthetic-secret", output)

    def test_cli_reserves_private_report_and_signals_failure(self):
        for result, expected in (("PASS", 0), ("FAIL", 1)):
            directory = self.directory / result
            with patch.object(restore.Rehearsal, "run", return_value={"status": result}):
                self.assertEqual(self.cli("--report-dir", str(directory))[0], expected)
            self.assertEqual(stat.S_IMODE(directory.stat().st_mode), 0o700)
            self.assertEqual(stat.S_IMODE((directory / "report.json").stat().st_mode), 0o600)
            self.assertEqual(json.loads((directory / "report.json").read_text())["status"], result)

    def test_commands_use_bounded_minimal_environment_and_hide_raw_errors(self):
        secret = "synthetic-secret-not-for-report"
        with patch.dict(os.environ, {"DOCKER_HOST": "tcp://remote.invalid:2375", "DOCKER_CONTEXT": "remote",
                                     "PGPASSWORD": secret, "DATABASE_URL": secret}), \
             patch.object(restore.subprocess, "run", return_value=subprocess.CompletedProcess([], 1, b"", secret.encode())) as run:
            with self.assertRaises(restore.RehearsalError) as error:
                self.runner._command(["docker", "version"], timeout=7)
        self.assertNotIn(secret, str(error.exception))
        self.assertEqual(run.call_args.kwargs["timeout"], 7)
        for key in ("DOCKER_HOST", "DOCKER_CONTEXT", "PGPASSWORD", "DATABASE_URL"):
            self.assertNotIn(key, run.call_args.kwargs["env"])
        self.assertFalse(run.call_args.kwargs.get("shell", False))

    def test_timeout_hides_captured_command_output(self):
        with patch.object(restore.subprocess, "run", side_effect=subprocess.TimeoutExpired(
                ["synthetic-private-command"], 1, output=b"synthetic-secret", stderr=b"synthetic-secret")):
            with self.assertRaises(restore.RehearsalError) as error:
                self.runner._command(["docker", "version"], timeout=1)
        self.assertEqual(str(error.exception), "command_timeout")

    def test_create_is_isolated_and_tracks_a_request_before_daemon_failure(self):
        observed = []
        def fail(args, **kwargs):
            observed.append(args)
            self.assertIn(self.runner.names[0], self.runner.requested_names)
            raise restore.RehearsalError("command_timeout")
        with patch.object(self.runner, "_command", side_effect=fail):
            with self.assertRaises(restore.RehearsalError):
                self.runner._create(self.runner.names[0], "sha256:" + "a" * 64)
        command = observed[0]
        self.assertEqual(command[:3], ["docker", "--host", "unix:///var/run/docker.sock"])
        for option, value in (("--pull", "never"), ("--network", "none")):
            self.assertEqual(command[command.index(option) + 1], value)
        self.assertFalse(set(command) & {"--publish", "-p", "--volume", "-v", "--mount", "--privileged"})

    def test_cleanup_never_removes_unowned_resource(self):
        self.runner.requested_names.append(self.runner.names[0])
        calls = []
        def docker(*args, **kwargs):
            calls.append(args)
            if args[0] == "ps":
                return ("a" * 64 + "\n").encode()
            if args[0] == "inspect":
                return b"/unrelated-production wrong-owner\n"
            self.fail("unexpected mutation of unowned container")
        with patch.object(self.runner, "_docker", side_effect=docker):
            with self.assertRaises(restore.RehearsalError) as error:
                self.runner._cleanup()
        self.assertEqual(str(error.exception), "cleanup_ownership_mismatch")
        self.assertFalse(any(call[0] == "rm" for call in calls))

    def test_interrupt_during_readiness_is_never_consumed_as_a_retry(self):
        with patch.object(self.runner, "_docker", return_value=b"postgres\n"), \
             patch.object(self.runner, "_psql", side_effect=restore.RehearsalError("interrupted")) as psql, \
             patch.object(restore.time, "sleep"):
            with self.assertRaises(restore.RehearsalError) as error:
                self.runner._create(self.runner.names[0], "sha256:" + "a" * 64)
        self.assertEqual(str(error.exception), "interrupted")
        psql.assert_called_once()

    def test_unobserved_timed_out_create_cannot_claim_cleanup_success(self):
        self.runner.requested_names.append(self.runner.names[0])
        self.runner.unconfirmed_creates.add(self.runner.names[0])
        with patch.object(self.runner, "_docker", return_value=b""), patch.object(restore.time, "sleep"):
            with self.assertRaises(restore.RehearsalError) as error:
                self.runner._cleanup()
        self.assertEqual(str(error.exception), "creation_completion_unconfirmed")

    def test_timed_out_create_is_cleaned_only_after_ownership_is_verified(self):
        name = self.runner.names[0]
        self.runner.requested_names.append(name)
        self.runner.unconfirmed_creates.add(name)
        container_id = "a" * 64
        removed = []
        def docker(*args, **kwargs):
            if args[0] == "ps":
                return b"" if removed else container_id.encode()
            if args[0] == "inspect":
                return f"/{name} {self.runner.run_id}".encode()
            if args[0] == "rm":
                self.assertEqual(args, ("rm", "--force", "--volumes", container_id))
                removed.append(container_id)
                return b""
            self.fail("unexpected Docker operation")
        with patch.object(self.runner, "_docker", side_effect=docker), patch.object(restore.time, "sleep"):
            cleanup = self.runner._cleanup()
        self.assertEqual(removed, [container_id])
        self.assertEqual(cleanup["status"], "PASS")
        self.assertTrue(cleanup["verified_absent"])

    def test_cleanup_failure_prevents_a_success_report(self):
        def execute():
            self.runner.report["status"] = "PASS"
        with patch.object(self.runner, "_execute", side_effect=execute), \
             patch.object(self.runner, "_cleanup", side_effect=restore.RehearsalError("command_timeout")):
            report = self.runner.run()
        self.assertEqual(report["status"], "FAIL")
        self.assertEqual(report["failure_stage"], "cleanup")
        self.assertFalse(report["cleanup"]["verified_absent"])

    def test_interruption_preserves_failure_and_runs_cleanup(self):
        def execute():
            self.runner.stage = "source_create"
            os.kill(os.getpid(), signal.SIGTERM)
        with patch.object(self.runner, "_execute", side_effect=execute), \
             patch.object(self.runner, "_cleanup", return_value={"status": "PASS", "verified_absent": True}) as cleanup:
            report = self.runner.run()
        cleanup.assert_called_once()
        self.assertEqual(report["status"], "FAIL")
        self.assertEqual(report["failure_code"], "interrupted")
        self.assertEqual(report["failure_stage"], "source_create")
        self.assertFalse(self.runner.temp_dir.exists())

    def test_invariant_results_must_be_nonempty_and_all_true(self):
        for invalid in (b"{}", b'{"saved_word_schedule":false}', b'{"saved_word_schedule":1}'):
            with patch.object(self.runner, "_psql", return_value=invalid):
                with self.assertRaises(restore.RehearsalError):
                    self.runner._checks("synthetic", b"checks")

    def test_seed_escaping_retains_quotes_unicode_nulls_and_rejects_extra_sql_columns(self):
        seed = json.loads((ROOT / "apps/api/cmd/seed/voc026-p1.json").read_bytes())
        seed["canonical_words"][0]["text"] = "learner's café\\path"
        sql = restore.seed_sql(json.dumps(seed).encode()).decode()
        self.assertIn("'learner''s café\\path'", sql)
        self.assertIn("SET LOCAL standard_conforming_strings = on", sql)
        self.assertEqual(restore.sql_literal(None), "NULL")
        self.assertEqual(restore.sql_literal(False), "FALSE")
        seed["canonical_words"][0]["unexpected); DROP TABLE users; --"] = "bad"
        with self.assertRaises(restore.RehearsalError):
            restore.seed_sql(json.dumps(seed).encode())


if __name__ == "__main__":
    unittest.main()
