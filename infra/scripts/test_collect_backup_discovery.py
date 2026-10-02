"""Offline transport tests. Local Python children replace SSH; no remote calls."""
import contextlib
import copy
from datetime import datetime, timedelta, timezone
import hashlib
import importlib.util
import io
import json
import os
from pathlib import Path
import stat
import signal
import shutil
import subprocess
import time
import sys
import tempfile
import unittest
from unittest import mock

SPEC = importlib.util.spec_from_file_location("backup_transport", Path(__file__).with_name("collect_backup_discovery.py"))
transport = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(transport)

SECRET = "PRIVATE_SECRET_SENTINEL"
ENV = {
    "BACKUP_SSH_HOST": "backup.example.invalid", "BACKUP_SSH_USER": "ubuntu",
    "BACKUP_SSH_PRIVATE_KEY": SECRET + "-key", "BACKUP_SSH_KNOWN_HOSTS": SECRET + "-known-hosts",
    "GITHUB_SHA": "a" * 40,
}
REPORT = {
    "schema_version": 1, "scope": "backup_leads_only", "observed_at": "2026-10-02T12:00:00Z",
    "collection_status": "complete",
    "tool_presence": {name: "not_found_on_path" for name in ("pg_dump", "pgbackrest", "wal_g", "restic", "borg")},
    "timer_inventory": {"installed_status": "observed", "loaded_status": "observed", "truncated": False,
                        "candidate_count": 0, "candidates": []},
    "production_backup": {"association": "unverified", "retained_artifacts": None,
                          "separate_storage": "unknown", "restore_record": "unknown", "alert_receipt": "unknown"},
    "coverage_gaps": ["cron_not_inspected", "user_timers_not_inspected", "containers_not_inspected",
                      "provider_backups_not_inspected", "configuration_and_archive_contents_not_inspected",
                      "production_association_unverified"],
}


class TransportTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.root = Path(self.temporary.name)
        self.output = self.root / "report"
        self.source, self.validator = transport.load_collector()
        self.received_at = datetime(2026, 10, 2, 12, 0, tzinfo=timezone.utc)
        clock = mock.patch.object(transport, "utc_now", return_value=self.received_at)
        clock.start()
        self.addCleanup(clock.stop)

    def tearDown(self):
        self.temporary.cleanup()

    def run_main(self, response=None, error=None, environ=None, argv=None):
        if response is None:
            response = (0, json.dumps(REPORT).encode())
        output = io.StringIO()
        with mock.patch.dict(os.environ, ENV if environ is None else environ, clear=True), \
                mock.patch.object(transport, "capture_process", return_value=response, side_effect=error) as child, \
                contextlib.redirect_stdout(output), contextlib.redirect_stderr(output):
            code = transport.main(["--report-dir", str(self.output)] if argv is None else argv)
        return code, output.getvalue(), child

    def envelope(self):
        return json.loads((self.output / "report.json").read_text())

    def test_complete_report_provenance_privacy_and_ssh_contract(self):
        observed = {}
        def fake_child(command, source):
            observed["command"], observed["source"] = command, source
            key = Path(command[command.index("-i") + 1])
            known = Path(next(value.split("=", 1)[1] for value in command if value.startswith("UserKnownHostsFile=")))
            observed["temporary"] = key.parent
            for path in (key, known):
                self.assertEqual(stat.S_IMODE(path.stat().st_mode), 0o600)
                self.assertIn(SECRET, path.read_text())
            self.assertEqual(stat.S_IMODE(key.parent.stat().st_mode), 0o700)
            return 0, json.dumps(REPORT).encode()
        code, logs, _ = self.run_main(error=fake_child)
        self.assertEqual(code, 0)
        self.assertEqual(logs, "backup discovery: collection_complete\n")
        envelope = self.envelope()
        self.assertEqual(envelope, {"schema_version": 1, "collector_revision": ENV["GITHUB_SHA"],
                         "collector_sha256": hashlib.sha256(self.source).hexdigest(),
                         "transport_status": "complete", "failure_code": None, "report": REPORT})
        self.assertEqual(observed["source"], self.source)
        self.assertFalse(observed["temporary"].exists())
        self.assertEqual(stat.S_IMODE(self.output.stat().st_mode), 0o700)
        self.assertEqual(stat.S_IMODE((self.output / "report.json").stat().st_mode), 0o600)
        self.assertEqual([p.name for p in self.output.iterdir()], ["report.json"])
        command = observed["command"]
        self.assertEqual(command[:4], ["/usr/bin/ssh", "-F", "/dev/null", "-T"])
        self.assertEqual(command[-3:], ["--", ENV["BACKUP_SSH_HOST"], "/usr/bin/python3 -I -B -S -"])
        for option in ("StrictHostKeyChecking=yes", "GlobalKnownHostsFile=/dev/null", "KnownHostsCommand=none",
                       "VerifyHostKeyDNS=no", "UpdateHostKeys=no", "BatchMode=yes", "IdentitiesOnly=yes",
                       "IdentityAgent=none", "PasswordAuthentication=no", "KbdInteractiveAuthentication=no",
                       "ClearAllForwardings=yes", "ForwardAgent=no", "ForwardX11=no", "ProxyCommand=none",
                       "ProxyJump=none", "ControlPath=none", "ConnectTimeout=15", "ConnectionAttempts=1",
                       "ServerAliveInterval=15", "ServerAliveCountMax=2"):
            self.assertIn(option, command)
        self.assertNotIn(SECRET, logs + json.dumps(envelope) + " ".join(command))

    def test_cancellation_after_validation_preserves_observation_but_never_exits_success(self):
        publish = transport.publish_report
        def publish_then_cancel(fd, envelope):
            publish(fd, envelope)
            transport.request_cancel(signal.SIGTERM, None)
        previous_handlers = {number: signal.getsignal(number) for number in (signal.SIGINT, signal.SIGTERM)}
        with mock.patch.object(transport, "publish_report", side_effect=publish_then_cancel):
            code, logs, _ = self.run_main()
        self.assertEqual(code, 1)
        self.assertEqual(logs, "backup discovery: interrupted\n")
        self.assertEqual(self.envelope()["transport_status"], "complete")
        self.assertIsNone(self.envelope()["failure_code"])
        self.assertEqual(self.envelope()["report"], REPORT)
        self.assertEqual(previous_handlers, {number: signal.getsignal(number) for number in previous_handlers})

    def test_partial_collection_is_not_transport_failure_or_backup_absence(self):
        report = copy.deepcopy(REPORT)
        report["collection_status"] = "partial"
        report["timer_inventory"]["installed_status"] = "timed_out"
        code, _, _ = self.run_main(response=(0, json.dumps(report).encode()))
        self.assertEqual(code, 1)
        self.assertEqual(self.envelope()["transport_status"], "complete")
        self.assertIsNone(self.envelope()["failure_code"])
        self.assertEqual(self.envelope()["report"]["production_backup"]["association"], "unverified")

    def test_live_receipt_freshness_boundaries(self):
        for age, accepted in ((300, True), (300.000001, False), (-30, True), (-30.000001, False)):
            with self.subTest(age=age):
                report = copy.deepcopy(REPORT)
                report["observed_at"] = (self.received_at - timedelta(seconds=age)).isoformat().replace("+00:00", "Z")
                raw = json.dumps(report).encode()
                if accepted:
                    self.assertEqual(transport.decode_report(raw, self.validator, self.received_at), report)
                else:
                    with self.assertRaises(transport.TransportError) as error:
                        transport.decode_report(raw, self.validator, self.received_at)
                    self.assertEqual(error.exception.code, "clock_skew_or_stale_report")

    def test_fresh_observation_can_contain_historical_timer_events(self):
        report = copy.deepcopy(REPORT)
        report["timer_inventory"]["candidate_count"] = 1
        report["timer_inventory"]["candidates"] = [{
            "id": "timer-001", "name_hints": ["backup"], "metadata_status": "observed",
            "schedule_state": "enabled", "active_state": "active",
            "last_trigger_at": "2020-01-01T12:00:00Z", "next_trigger_at": "2026-10-03T12:00:00Z",
            "service_result": "success", "service_finished_at": "2020-01-01T12:00:05Z",
        }]
        self.assertEqual(transport.decode_report(json.dumps(report).encode(), self.validator, self.received_at), report)

    def test_stale_receipt_is_fixed_failure_without_retaining_remote_report(self):
        report = copy.deepcopy(REPORT)
        report["observed_at"] = "2026-10-02T11:54:59Z"
        code, logs, _ = self.run_main(response=(0, json.dumps(report).encode()))
        self.assertEqual(code, 1)
        self.assertEqual(self.envelope()["failure_code"], "clock_skew_or_stale_report")
        self.assertIsNone(self.envelope()["report"])
        self.assertEqual(logs, "backup discovery: clock_skew_or_stale_report\n")

    def test_failed_collection_preserves_validated_failure_report(self):
        report = copy.deepcopy(REPORT)
        report["collection_status"] = "failed"
        report["tool_presence"] = dict.fromkeys(report["tool_presence"], "unavailable")
        report["timer_inventory"].update(installed_status="unavailable", loaded_status="unavailable", candidate_count=None)
        code, _, _ = self.run_main(response=(0, json.dumps(report).encode()))
        self.assertEqual(code, 1)
        self.assertEqual(self.envelope()["report"], report)
        self.assertEqual(self.envelope()["transport_status"], "complete")

    def test_hostile_remote_output_is_never_logged_or_saved(self):
        malicious = copy.deepcopy(REPORT)
        malicious["private_detail"] = SECRET
        for raw in (SECRET.encode(), b"\xff" + SECRET.encode(), json.dumps(malicious).encode(),
                    b'{"schema_version":1,"schema_version":1,"secret":"' + SECRET.encode() + b'"}',
                    b'{"schema_version":NaN}', b"[" * 2000):
            with self.subTest(raw_length=len(raw)):
                self.output = self.root / str(len(list(self.root.iterdir())))
                code, logs, _ = self.run_main(response=(0, raw))
                self.assertEqual(code, 1)
                envelope = self.envelope()
                self.assertEqual(envelope["failure_code"], "invalid_report")
                self.assertIsNone(envelope["report"])
                self.assertNotIn(SECRET, logs + json.dumps(envelope))

    def test_ssh_failure_ignores_even_valid_stdout(self):
        code, logs, _ = self.run_main(response=(255, json.dumps(REPORT).encode() + SECRET.encode()))
        self.assertEqual(code, 1)
        self.assertEqual(self.envelope()["failure_code"], "ssh_failed")
        self.assertIsNone(self.envelope()["report"])
        self.assertNotIn(SECRET, logs + json.dumps(self.envelope()))

    def test_timeout_and_unexpected_exception_are_sanitized_and_clean_key_files(self):
        for error, expected in ((transport.TransportError("ssh_timeout"), "ssh_timeout"),
                                (RuntimeError(SECRET), "transport_error"),
                                (transport.TransportError(SECRET), "transport_error")):
            self.output = self.root / (expected + str(len(list(self.root.iterdir()))))
            paths = []
            def fail(command, source):
                paths.append(Path(command[command.index("-i") + 1]))
                raise error
            code, logs, _ = self.run_main(error=fail)
            self.assertEqual(code, 1)
            self.assertEqual(self.envelope()["failure_code"], expected)
            self.assertFalse(paths[0].exists())
            self.assertFalse(paths[0].parent.exists())
            self.assertNotIn(SECRET, logs + json.dumps(self.envelope()))

    def test_input_errors_make_no_ssh_call_and_do_not_echo_values(self):
        for name, value in (("BACKUP_SSH_HOST", "-oProxyCommand=" + SECRET),
                            ("BACKUP_SSH_HOST", "example.invalid;" + SECRET),
                            ("BACKUP_SSH_HOST", "user@example.invalid"),
                            ("BACKUP_SSH_USER", "root;" + SECRET),
                            ("GITHUB_SHA", SECRET), ("BACKUP_SSH_PRIVATE_KEY", ""),
                            ("BACKUP_SSH_KNOWN_HOSTS", "x" * (128 * 1024 + 1))):
            with self.subTest(name=name):
                environ = dict(ENV, **{name: value})
                code, logs, child = self.run_main(environ=environ)
                self.assertEqual(code, 2)
                child.assert_not_called()
                self.assertFalse(self.output.exists())
                self.assertNotIn(SECRET, logs)
        code, logs, child = self.run_main(argv=["--unknown", SECRET])
        self.assertEqual(code, 2)
        child.assert_not_called()
        self.assertNotIn(SECRET, logs)

    def test_valid_host_literals_and_dns_have_no_shell_interpretation(self):
        for host in ("192.0.2.1", "2001:db8::1", "backup.example.invalid", "host-1"):
            self.assertEqual(transport.validate_environment(dict(ENV, BACKUP_SSH_HOST=host))[0], host)
        for host in ("host\nsecret", "host/path", "[2001:db8::1]", "fe80::1%eth0", "-host"):
            with self.assertRaises(transport.InputError):
                transport.validate_environment(dict(ENV, BACKUP_SSH_HOST=host))

    def test_existing_directory_and_symlink_are_preserved_before_ssh(self):
        existing = self.root / "existing"
        existing.mkdir()
        (existing / "report.json").write_text("prior evidence")
        for path in (existing, self.root / "alias"):
            if path != existing:
                path.symlink_to(existing, target_is_directory=True)
            self.output = path
            code, _, child = self.run_main()
            self.assertEqual(code, 2)
            child.assert_not_called()
            self.assertEqual((existing / "report.json").read_text(), "prior evidence")

    def test_atomic_publication_never_overwrites_competing_evidence(self):
        fd = transport.reserve_directory(self.output)
        try:
            (self.output / "report.json").write_text("prior evidence")
            with self.assertRaises(transport.InputError):
                transport.publish_report(fd, {"report": "new"})
            self.assertEqual((self.output / "report.json").read_text(), "prior evidence")
            self.assertEqual([p.name for p in self.output.iterdir()], ["report.json"])
        finally:
            os.close(fd)

    def test_failed_publication_leaves_no_partial_json(self):
        fd = transport.reserve_directory(self.output)
        try:
            with mock.patch.object(transport.os, "link", side_effect=OSError(SECRET)):
                with self.assertRaises(transport.InputError) as error:
                    transport.publish_report(fd, {"report": "new"})
            self.assertEqual(error.exception.code, "report_write_failed")
            self.assertEqual(list(self.output.iterdir()), [])
        finally:
            os.close(fd)

    def test_open_directory_handle_prevents_symlink_redirection(self):
        fd = transport.reserve_directory(self.output)
        preserved = self.root / "preserved"
        preserved.mkdir()
        moved = self.root / "original"
        self.output.rename(moved)
        self.output.symlink_to(preserved, target_is_directory=True)
        try:
            transport.publish_report(fd, {"test": True})
            self.assertEqual(list(preserved.iterdir()), [])
            self.assertEqual(json.loads((moved / "report.json").read_text()), {"test": True})
        finally:
            os.close(fd)

    def test_collector_source_is_bounded_before_compile(self):
        with mock.patch.object(transport.Path, "open", return_value=io.BytesIO(b"x" * (transport.MAX_SOURCE_BYTES + 1))):
            with self.assertRaises(transport.InputError) as error:
                transport.load_collector()
        self.assertEqual(error.exception.code, "collector_unavailable")


class CancellationTests(unittest.TestCase):
    def test_main_signals_stop_owned_child_and_remove_private_credentials(self):
        harness = """
import importlib.util, pathlib, sys
spec = importlib.util.spec_from_file_location('cancel_transport', sys.argv[1])
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)
marker, key_marker = sys.argv[2], sys.argv[3]
def local_child(host, user, key_path, known_hosts_path):
    pathlib.Path(key_marker).write_text(str(key_path.parent))
    script = 'import os,pathlib,time; pathlib.Path(' + repr(marker) + ').write_text(str(os.getpid())); time.sleep(30)'
    return [sys.executable, '-I', '-c', script]
module.ssh_command = local_child
sys.exit(module.main(['--report-dir', sys.argv[4]]))
"""
        for requested_signal in (signal.SIGTERM, signal.SIGINT):
            with self.subTest(signal=requested_signal), tempfile.TemporaryDirectory() as temporary:
                root = Path(temporary)
                marker, key_marker, report = root / "child", root / "keydir", root / "report"
                process = subprocess.Popen(
                    [sys.executable, "-I", "-c", harness, str(Path(transport.__file__).resolve()),
                     str(marker), str(key_marker), str(report)],
                    stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                    env=dict(transport.SUBPROCESS_ENV, **ENV),
                )
                child_id = None
                child_gone = False
                credential_dir = None
                try:
                    deadline = time.monotonic() + 5
                    while not marker.exists() and process.poll() is None and time.monotonic() < deadline:
                        time.sleep(0.01)
                    self.assertTrue(marker.exists(), "local fake SSH child did not start")
                    child_id = int(marker.read_text())
                    credential_dir = Path(key_marker.read_text())
                    process.send_signal(requested_signal)
                    stdout, stderr = process.communicate(timeout=5)
                    self.assertEqual(process.returncode, 1)
                    self.assertEqual(stdout, b"backup discovery: interrupted\n")
                    self.assertEqual(stderr, b"")
                    self.assertFalse(credential_dir.exists())
                    with self.assertRaises(ProcessLookupError):
                        os.kill(child_id, 0)
                    child_gone = True
                    envelope = json.loads((report / "report.json").read_text())
                    self.assertEqual(envelope["failure_code"], "interrupted")
                    self.assertIsNone(envelope["report"])
                    self.assertNotIn(SECRET, stdout.decode() + stderr.decode() + json.dumps(envelope))
                finally:
                    # Reproduction failures must not themselves leave a fake child or key file.
                    if process.poll() is None:
                        process.kill()
                    process.communicate(timeout=5)
                    if child_id is not None and not child_gone:
                        try:
                            os.killpg(child_id, signal.SIGKILL)
                        except ProcessLookupError:
                            pass
                    if credential_dir is not None and credential_dir.exists():
                        self.assertEqual(credential_dir.parent, Path("/tmp"))
                        self.assertTrue(credential_dir.name.startswith("vocanova-backup-ssh-"))
                        shutil.rmtree(credential_dir)


class BoundedChildTests(unittest.TestCase):
    def child(self, script, source=b"", **kwargs):
        return transport.capture_process([sys.executable, "-I", "-c", script], source, **kwargs)

    def test_large_stdin_and_both_output_pipes_do_not_deadlock(self):
        source = b"s" * (192 * 1024)
        code, output = self.child("import sys; data=sys.stdin.buffer.read(); sys.stderr.write('discarded'); print(len(data))", source)
        self.assertEqual(code, 0)
        self.assertEqual(output, b"196608\n")

    def test_parent_secrets_and_ssh_agent_are_absent_from_child_environment(self):
        with mock.patch.dict(os.environ, {"BACKUP_SSH_PRIVATE_KEY": SECRET, "SSH_AUTH_SOCK": SECRET}):
            code, output = self.child("import os,json; print(json.dumps(dict(os.environ)))")
        self.assertEqual(code, 0)
        environment = json.loads(output)
        self.assertNotIn("BACKUP_SSH_PRIVATE_KEY", environment)
        self.assertNotIn("SSH_AUTH_SOCK", environment)
        self.assertNotIn(SECRET, output.decode())

    def test_stdout_and_stderr_floods_are_bounded(self):
        for descriptor in (1, 2):
            with self.subTest(descriptor=descriptor):
                with self.assertRaises(transport.TransportError) as error:
                    self.child(f"import os; os.write({descriptor}, b'x'*20000)", output_limit=1024)
                self.assertEqual(error.exception.code, "output_limit")

    def test_stalled_child_and_blocked_stdin_time_out(self):
        with self.assertRaises(transport.TransportError) as error:
            self.child("import time; time.sleep(10)", b"s" * (192 * 1024), timeout=0.1)
        self.assertEqual(error.exception.code, "ssh_timeout")

    def test_early_exit_during_stdin_write_is_captured(self):
        code, output = self.child("import sys; sys.exit(3)", b"s" * (192 * 1024))
        self.assertEqual(code, 3)
        self.assertEqual(output, b"")

    def test_completed_child_is_not_signalled_after_wait(self):
        with mock.patch.object(transport.os, "killpg") as kill_group:
            self.assertEqual(self.child("print('done')"), (0, b"done\n"))
        kill_group.assert_not_called()

    def test_missing_executable_has_fixed_failure(self):
        with self.assertRaises(transport.TransportError) as error:
            transport.capture_process(["/does-not-exist-" + SECRET], b"")
        self.assertEqual(error.exception.code, "ssh_unavailable")
        self.assertNotIn(SECRET, str(error.exception))


if __name__ == "__main__":
    unittest.main()
