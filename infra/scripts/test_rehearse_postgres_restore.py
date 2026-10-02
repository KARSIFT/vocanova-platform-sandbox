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
import sys
import tempfile
import time
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

    def test_signal_during_serialization_publishes_complete_failure_report(self):
        # Separate process: the historical bug restores the default SIGTERM
        # action, so reproducing it in this test process would kill the suite.
        program = '''
import json, os, signal, sys
sys.path.insert(0, sys.argv[1])
import rehearse_postgres_restore as restore
restore.Rehearsal.run = lambda self: {"status":"PASS", "failure_stage":None, "failure_code":None}
original = restore.json.dump
sent = False
def interrupted_dump(*args, **kwargs):
    global sent
    if sent:
        return original(*args, **kwargs)
    if sys.argv[3] == "after":
        original(*args, **kwargs)
    if not sent:
        sent = True
        os.kill(os.getpid(), signal.SIGTERM)
        os.kill(os.getpid(), signal.SIGINT)
    if sys.argv[3] == "before":
        return original(*args, **kwargs)
restore.json.dump = interrupted_dump
raise SystemExit(restore.main(["--report-dir", sys.argv[2]]))
'''
        for timing in ("before", "after"):
            with self.subTest(timing=timing):
                directory = self.directory / ("interrupted-report-" + timing)
                process = subprocess.run([sys.executable, "-c", program, str(ROOT / "infra/scripts"), str(directory), timing],
                                         capture_output=True, timeout=10)
                self.assertEqual(process.returncode, 1, "serialization interruption was not recorded")
                report = json.loads((directory / "report.json").read_text())
                self.assertEqual(report["status"], "FAIL")
                self.assertEqual(report["failure_stage"], "report_persistence")
                self.assertEqual(report["failure_code"], "interrupted")
                self.assertEqual(sorted(p.name for p in directory.iterdir()), ["report.json"])

    def test_failed_serialization_never_publishes_partial_json(self):
        directory = self.directory / "write-error"
        def broken_dump(data, handle, **kwargs):
            handle.write('{"status":')
            raise OSError("synthetic-private-disk-diagnostic")
        with patch.object(restore.Rehearsal, "run", return_value={"status": "PASS"}), \
             patch.object(restore.json, "dump", side_effect=broken_dump):
            status, output = self.cli("--report-dir", str(directory))
        self.assertEqual(status, 2)
        self.assertNotIn("synthetic-private-disk-diagnostic", output)
        self.assertFalse((directory / "report.json").exists())
        self.assertEqual(list(directory.iterdir()), [])

    def test_atomic_publication_refuses_a_report_created_during_execution(self):
        directory = self.directory / "existing-report"
        def execute(runner):
            (directory / "report.json").write_text("existing evidence")
            return {"status": "PASS"}
        with patch.object(restore.Rehearsal, "run", new=execute):
            status, _ = self.cli("--report-dir", str(directory))
        self.assertEqual(status, 2)
        self.assertEqual((directory / "report.json").read_text(), "existing evidence")
        self.assertEqual(sorted(p.name for p in directory.iterdir()), ["report.json"])

    def test_failed_interruption_rewrite_removes_owned_stale_pass(self):
        directory = self.directory / "failed-correction"
        original_link, original_dump = restore.os.link, restore.json.dump
        dumps = []
        def interrupted_link(*args, **kwargs):
            original_link(*args, **kwargs)
            os.kill(os.getpid(), signal.SIGTERM)
        def failing_second_dump(*args, **kwargs):
            dumps.append(True)
            if len(dumps) == 2:
                raise OSError("synthetic failed interruption rewrite")
            return original_dump(*args, **kwargs)
        with patch.object(restore.Rehearsal, "run", return_value={"status": "PASS"}), \
             patch.object(restore.os, "link", side_effect=interrupted_link), \
             patch.object(restore.json, "dump", side_effect=failing_second_dump):
            status, _ = self.cli("--report-dir", str(directory))
        self.assertEqual(status, 2)
        self.assertEqual(len(dumps), 2)
        self.assertFalse((directory / "report.json").exists(), "failed correction must not retain a stale PASS")
        self.assertEqual(list(directory.iterdir()), [])

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

    def test_terminal_group_interrupt_does_not_kill_cleanup_command(self):
        program = '''
import importlib.util, sys
spec = importlib.util.spec_from_file_location("restore", sys.argv[1])
restore = importlib.util.module_from_spec(spec)
spec.loader.exec_module(restore)
restore.Rehearsal._execute = lambda self: self.report.update(status="PASS")
def cleanup(self):
    helper = "import pathlib,sys,time; pathlib.Path(sys.argv[1]).write_text('started'); time.sleep(0.6); print('finished')"
    result = self._command([sys.executable, "-c", helper, sys.argv[3]], timeout=3)
    if result.strip() != b"finished":
        raise restore.RehearsalError("cleanup_child_did_not_finish")
    return {"status":"PASS", "verified_absent":True}
restore.Rehearsal._cleanup = cleanup
raise SystemExit(restore.main(["--report-dir", sys.argv[2]]))
'''
        directory = self.directory / "group-interrupt"
        marker = self.directory / "cleanup-started"
        process = subprocess.Popen([sys.executable, "-c", program, restore.__file__, str(directory), str(marker)],
                                   stdout=subprocess.PIPE, stderr=subprocess.PIPE, start_new_session=True)
        try:
            deadline = time.monotonic() + 5
            while not marker.exists() and process.poll() is None and time.monotonic() < deadline:
                time.sleep(0.01)
            self.assertTrue(marker.exists(), "cleanup subprocess did not start")
            # Match terminal Ctrl-C delivery, not just an os.kill of the parent.
            os.killpg(process.pid, signal.SIGINT)
            process.communicate(timeout=5)
            self.assertEqual(process.returncode, 1)
            report = json.loads((directory / "report.json").read_text())
            self.assertEqual(report["status"], "FAIL")
            self.assertEqual(report["failure_code"], "interrupted")
            self.assertEqual(report["cleanup"]["status"], "PASS", "terminal interrupt aborted a cleanup subprocess")
            self.assertTrue(report["cleanup"]["temporary_files_removed"])
        finally:
            if process.poll() is None:
                process.kill()
                process.communicate(timeout=5)
            process.stdout.close()
            process.stderr.close()

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

    def test_first_interrupt_during_cleanup_finishes_cleanup_but_cannot_pass(self):
        cleanup_finished = []
        def execute():
            self.runner.report["status"] = "PASS"
        def cleanup():
            os.kill(os.getpid(), signal.SIGTERM)
            os.kill(os.getpid(), signal.SIGINT)
            cleanup_finished.append(True)
            return {"status": "PASS", "verified_absent": True}
        with patch.object(self.runner, "_execute", side_effect=execute), \
             patch.object(self.runner, "_cleanup", side_effect=cleanup):
            report = self.runner.run()
        self.assertEqual(cleanup_finished, [True])
        self.assertEqual(report["cleanup"]["status"], "PASS")
        self.assertEqual(report["status"], "FAIL")
        self.assertEqual(report["failure_stage"], "cleanup")
        self.assertEqual(report["failure_code"], "interrupted")

    def test_cleanup_interrupt_preserves_an_earlier_restore_failure(self):
        def execute():
            self.runner.stage = "restore"
            raise restore.RehearsalError("command_failed")
        def cleanup():
            os.kill(os.getpid(), signal.SIGTERM)
            return {"status": "PASS", "verified_absent": True}
        with patch.object(self.runner, "_execute", side_effect=execute), \
             patch.object(self.runner, "_cleanup", side_effect=cleanup):
            report = self.runner.run()
        self.assertEqual(report["status"], "FAIL")
        self.assertEqual(report["failure_stage"], "restore")
        self.assertEqual(report["failure_code"], "command_failed")
        self.assertTrue(report["cleanup"]["temporary_files_removed"])

    def test_interrupt_during_temporary_file_cleanup_also_prevents_pass(self):
        remove_temporary = restore.tempfile.TemporaryDirectory.cleanup
        def execute():
            self.runner.report["status"] = "PASS"
        def cleanup(directory):
            os.kill(os.getpid(), signal.SIGINT)
            remove_temporary(directory)
        with patch.object(self.runner, "_execute", side_effect=execute), \
             patch.object(self.runner, "_cleanup", return_value={"status": "PASS", "verified_absent": True}), \
             patch.object(restore.tempfile.TemporaryDirectory, "cleanup", new=cleanup):
            report = self.runner.run()
        self.assertEqual(report["status"], "FAIL")
        self.assertEqual(report["failure_stage"], "cleanup")
        self.assertEqual(report["failure_code"], "interrupted")
        self.assertTrue(report["cleanup"]["temporary_files_removed"])

    def test_unsupported_cached_tool_is_rejected_before_migrations(self):
        def docker(*args, **kwargs):
            if args[:2] == ("image", "inspect"):
                return ("sha256:" + "a" * 64).encode()
            if args[0] == "version":
                return b"29.8.1"
            if args[-1] == "--version":
                return f"{args[-2]} (PostgreSQL) 16.9".encode()
            if args[-1] == "--help":
                return b"synthetic old pg_dump help without repeatable-dump capability"
            self.fail("unexpected Docker call before capability rejection")
        with patch.object(self.runner, "_command", side_effect=[b"a" * 40, b""]), \
             patch.object(self.runner, "_docker", side_effect=docker), \
             patch.object(self.runner, "_create"), \
             patch.object(self.runner, "_psql", return_value=b"160009") as psql:
            with self.assertRaises(restore.RehearsalError) as error:
                self.runner._execute()
        self.assertEqual(str(error.exception), "pg_dump_restrict_key_unsupported")
        psql.assert_called_once_with(self.runner.names[0], "SHOW server_version_num;")

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
