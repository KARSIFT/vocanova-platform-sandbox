"""Explicit local-Docker recovery acceptance; never silently skips missing Docker.

Run separately from foundation tests. The optional report root must be new;
only sanitized reports survive. No external archive or connection is accepted.
"""
import contextlib
import io
import json
import os
from pathlib import Path
import re
import signal
import subprocess
import sys
import tempfile
import time
import unittest
from unittest.mock import patch

import rehearse_postgres_restore as restore
from rehearse_postgres_restore import COMMAND_ENV, DOCKER_ENDPOINT, LABEL

ROOT = Path(__file__).resolve().parents[2]
RUNNER = ROOT / "infra/scripts/rehearse-postgres-restore.sh"


class RecoveryAcceptance(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.scratch = tempfile.TemporaryDirectory(prefix="vocanova-recovery-tests-")
        cls.addClassCleanup(cls.scratch.cleanup)
        selected = os.environ.get("VOCANOVA_REHEARSAL_REPORT_ROOT")
        cls.reports = Path(selected) if selected else Path(cls.scratch.name) / "reports"
        cls.reports.mkdir(mode=0o700, parents=False, exist_ok=False)
        cls.docker_config = Path(cls.scratch.name) / "docker-config"
        cls.docker_config.mkdir(mode=0o700)
        cls.docker("version", "--format", "{{.Server.Version}}")

    @classmethod
    def docker(cls, *args):
        result = subprocess.run(
            ["docker", "--host", DOCKER_ENDPOINT, "--config", str(cls.docker_config), *args],
            env=COMMAND_ENV, capture_output=True, timeout=15,
        )
        if result.returncode:
            raise AssertionError("local Docker verification command failed")
        return result.stdout.decode().strip()

    def read_report(self, directory):
        raw = (directory / "report.json").read_text()
        for forbidden in ("restore-rehearsal@vocanova.invalid", "I caught up with Maya", "sentinel-do-not-inherit"):
            self.assertNotIn(forbidden, raw)
        report = json.loads(raw)
        self.assertIn("synthetic-only", report["scope"])
        self.assertEqual(report["cleanup"]["status"], "PASS", report["cleanup"])
        self.assertTrue(report["cleanup"]["verified_absent"])
        self.assertRegex(report["run_id"], r"^[0-9a-f]{32}$")
        remaining = self.docker("ps", "--all", "--filter", f"label={LABEL}={report['run_id']}", "--format", "{{.ID}}")
        self.assertEqual(remaining, "", "rehearsal left its own containers behind")
        self.assertEqual(sorted(p.name for p in directory.iterdir()), ["report.json"])
        return report

    def run_case(self, fault=None):
        directory = self.reports / (fault or "success")
        command = ["bash", str(RUNNER), "--report-dir", str(directory)]
        if fault:
            command.extend(["--fault", fault])
        # Deliberately hostile ambient defaults must have no effect on the
        # local-only tool. This sentinel is synthetic and is never a credential.
        environment = {**COMMAND_ENV, "DOCKER_HOST": "tcp://remote.invalid:2375",
                       "DOCKER_CONTEXT": "sentinel-do-not-inherit",
                       "PGPASSWORD": "sentinel-do-not-inherit",
                       "DATABASE_URL": "postgres://sentinel-do-not-inherit@remote.invalid/db"}
        process = subprocess.Popen(command, cwd=ROOT, env=environment,
                                   stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        try:
            try:
                process.communicate(timeout=180)
            except subprocess.TimeoutExpired:
                process.terminate()
                try:
                    process.communicate(timeout=45)
                except subprocess.TimeoutExpired:
                    process.kill()
                    process.communicate(timeout=5)
                self.fail("rehearsal exceeded its acceptance-test time limit")
        finally:
            process.stdout.close()
            process.stderr.close()
        report = self.read_report(directory)
        self.assertEqual(process.returncode, 1 if fault else 0, {
            "status": report["status"], "failure_stage": report["failure_stage"], "failure_code": report["failure_code"],
        })
        return report

    def test_success_restores_actual_learning_state_and_constraints(self):
        report = self.run_case()
        self.assertEqual(report["status"], "PASS")
        self.assertIsNone(report["failure_stage"])
        self.assertTrue(report["schema"]["match"])
        self.assertTrue(report["sequences"]["match"])
        self.assertGreater(len(report["tables"]), 10)
        self.assertTrue(all(table["match"] for table in report["tables"].values()))
        for table, count in (("user_words", 1), ("review_attempts", 1), ("learner_sentences", 1),
                             ("ai_feedback_attempts", 1), ("daily_mission_snapshots", 1),
                             ("daily_activity_summaries", 1), ("confidence_point_ledger", 4)):
            self.assertEqual(report["tables"]["public." + table]["target"]["count"], count)
        for side in ("source", "target"):
            checks = report["independent_checks"][side]
            self.assertTrue(checks)
            self.assertTrue(all(value is True for value in checks.values()))
            self.assertIn("saved_word_schedule", checks)
            self.assertIn("learning_relationships", checks)
            self.assertIn("restored_write_and_protections", checks)
        self.assertEqual(report["archive"]["sha256"], report["archive"]["restore_input_sha256"])

    def test_corrupt_archive_fails_during_restore(self):
        report = self.run_case("corrupt-archive")
        self.assertEqual(report["status"], "FAIL")
        self.assertEqual(report["failure_stage"], "restore")
        self.assertEqual(report["failure_code"], "command_failed")
        self.assertNotEqual(report["archive"]["sha256"], report["archive"]["restore_input_sha256"])

    def test_truncated_archive_fails_during_restore(self):
        report = self.run_case("truncated-archive")
        self.assertEqual(report["status"], "FAIL")
        self.assertEqual(report["failure_stage"], "restore")
        self.assertEqual(report["failure_code"], "command_failed")
        self.assertLess(report["archive"]["restore_input_size_bytes"], report["archive"]["size_bytes"])

    def test_missing_learning_record_is_detected_independently(self):
        report = self.run_case("missing-record")
        self.assertEqual(report["status"], "FAIL")
        self.assertEqual(report["failure_stage"], "invariants")
        self.assertEqual(report["failure_code"], "invariant_check_failed")
        table = report["tables"]["public.daily_activity_summaries"]
        self.assertEqual(table["source"]["count"], 1)
        self.assertEqual(table["target"]["count"], 0)
        self.assertFalse(table["match"])

    def assert_delete_protection_is_required(self, table):
        directory = self.reports / ("weakened-" + table)
        original_checks = restore.Rehearsal._checks
        weakened = []
        def checks(runner, container, sql):
            if container == runner.names[0]:
                # These two fixed test cases mutate only the synthetic source
                # created by this runner. No runtime or CLI table input exists.
                self.assertIn(table, ("confidence_point_ledger", "grace_day_ledger"))
                trigger = table + "_append_only"
                runner._psql(container, f"""BEGIN;
                    DROP TRIGGER {trigger} ON public.{table};
                    CREATE TRIGGER {trigger} BEFORE UPDATE ON public.{table}
                    FOR EACH ROW EXECUTE FUNCTION public.vocanova_reject_learning_ledger_mutation();
                    COMMIT;
                """)
                bits = runner._psql(container, f"SELECT tgtype FROM pg_trigger WHERE tgrelid='public.{table}'::regclass AND tgname='{trigger}';")
                self.assertEqual(bits.strip(), b"19", "test must establish UPDATE-only trigger before checking it")
                weakened.append(table)
            return original_checks(runner, container, sql)
        with patch.object(restore.Rehearsal, "_checks", new=checks), \
             contextlib.redirect_stdout(io.StringIO()):
            status = restore.main(["--report-dir", str(directory)])
        self.assertEqual(weakened, [table])
        self.assertEqual(status, 1, "a weakened source schema must not pass merely because the restore matches")
        report = self.read_report(directory)
        self.assertEqual(report["status"], "FAIL")
        self.assertEqual(report["failure_stage"], "fixture")
        self.assertEqual(report["failure_code"], "command_failed")

    def test_point_ledger_missing_delete_protection_is_rejected(self):
        self.assert_delete_protection_is_required("confidence_point_ledger")

    def test_grace_ledger_missing_delete_protection_is_rejected(self):
        self.assert_delete_protection_is_required("grace_day_ledger")

    def test_named_but_ineffective_feedback_constraint_is_rejected(self):
        directory = self.reports / "weakened-feedback-constraint"
        original_checks = restore.Rehearsal._checks
        weakened = []
        def checks(runner, container, sql):
            if container == runner.names[0]:
                runner._psql(container, """BEGIN;
                    ALTER TABLE public.ai_feedback_attempts
                      DROP CONSTRAINT ai_feedback_attempts_feedback_json_required_on_success;
                    ALTER TABLE public.ai_feedback_attempts
                      ADD CONSTRAINT ai_feedback_attempts_feedback_json_required_on_success CHECK (true);
                    COMMIT;
                """)
                definition = runner._psql(container, """SELECT pg_get_constraintdef(oid)
                    FROM pg_constraint WHERE conrelid='public.ai_feedback_attempts'::regclass
                    AND conname='ai_feedback_attempts_feedback_json_required_on_success';""")
                self.assertEqual(definition.strip(), b"CHECK (true)")
                weakened.append(True)
            return original_checks(runner, container, sql)
        with patch.object(restore.Rehearsal, "_checks", new=checks), \
             contextlib.redirect_stdout(io.StringIO()):
            status = restore.main(["--report-dir", str(directory)])
        self.assertEqual(weakened, [True])
        self.assertEqual(status, 1, "a matching constraint name alone must not establish its protection")
        report = self.read_report(directory)
        self.assertEqual(report["status"], "FAIL")
        self.assertEqual(report["failure_stage"], "fixture")
        self.assertEqual(report["failure_code"], "command_failed")

    def test_signal_cleans_up_isolated_container_and_preserves_failed_report(self):
        directory = self.reports / "interrupted"
        def containers():
            output = self.docker("ps", "--all", "--filter", f"label={LABEL}", "--format", "{{.Names}}")
            return set(output.splitlines())
        before = containers()
        process = subprocess.Popen(["bash", str(RUNNER), "--report-dir", str(directory)],
                                   cwd=ROOT, env=COMMAND_ENV, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        self.addCleanup(process.stdout.close)
        self.addCleanup(process.stderr.close)
        try:
            deadline = time.monotonic() + 45
            observed = set()
            while process.poll() is None and time.monotonic() < deadline:
                observed = containers() - before
                if observed:
                    break
                time.sleep(0.05)
            self.assertTrue(observed, "no disposable container observed before interruption")
            source = next(name for name in observed if re.fullmatch(r"vocanova-restore-[0-9a-f]{32}-source", name))
            ownership = self.docker("inspect", "--format", '{{index .Config.Labels "' + LABEL + '"}}', source)
            self.assertRegex(ownership, r"^[0-9a-f]{32}$")
            config = json.loads(self.docker("inspect", "--format", "{{json .HostConfig}}", source))
            self.assertEqual(config["NetworkMode"], "none")
            self.assertFalse(config["PortBindings"])
            self.assertFalse(config["Binds"])
            self.assertFalse(config["Privileged"])
            process.send_signal(signal.SIGTERM)
            process.communicate(timeout=45)
            self.assertEqual(process.returncode, 1)
            report = self.read_report(directory)
            self.assertEqual(report["run_id"], ownership)
            self.assertEqual(report["status"], "FAIL")
            self.assertEqual(report["failure_code"], "interrupted")
        finally:
            if process.poll() is None:
                process.terminate()
                try:
                    process.communicate(timeout=45)
                except subprocess.TimeoutExpired:
                    process.kill()
                    process.communicate(timeout=5)


if __name__ == "__main__":
    unittest.main(verbosity=2)
