"""Synthetic command/probe tests only; never inspect a host or invoke systemctl."""
import contextlib
import copy
import datetime
import importlib.util
import io
import json
from pathlib import Path
import subprocess
from types import SimpleNamespace
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location("discovery", Path(__file__).with_name("discover_backup_metadata.py"))
discovery = importlib.util.module_from_spec(spec)
spec.loader.exec_module(discovery)
NOW = datetime.datetime(2026, 10, 2, 12, 0, 0, tzinfo=datetime.timezone.utc)
START = "Fri 2026-10-02 10:00:00 UTC"
FINISH = "Fri 2026-10-02 10:01:00 UTC"
NEXT = "Sat 2026-10-03 10:00:00 UTC"
PRIVATE_TIMER = "vocanova-private-host-backup.timer"
PRIVATE_SERVICE = "private-path-backup.service"


class FakeSystem:
    def __init__(self):
        self.installed = f"{PRIVATE_TIMER} disabled enabled\nordinary.timer enabled enabled\n"
        self.loaded = "restic-secret-destination.timer loaded inactive dead private-description\nordinary.timer loaded active waiting private-description\n"
        self.timer = {"UnitFileState": "disabled", "ActiveState": "inactive", "LastTriggerUSec": FINISH, "NextElapseUSecRealtime": NEXT, "Unit": PRIVATE_SERVICE}
        self.service = {"ActiveState": "inactive", "Result": "success", "ExecMainStartTimestamp": START, "ExecMainExitTimestamp": FINISH}
        self.statuses = {}
        self.calls = []

    def command(self, args, timeout):
        self.calls.append((args, timeout))
        if "list-unit-files" in args:
            kind, output = "installed", self.installed
        elif "list-units" in args:
            kind, output = "loaded", self.loaded
        elif args[-1].endswith(".timer"):
            kind, output = "timer", "\n".join(f"{key}={value}" for key, value in self.timer.items())
        else:
            kind, output = "service", "\n".join(f"{key}={value}" for key, value in self.service.items())
        return self.statuses.get(kind, ("observed", output.encode(), False))


class DiscoveryTests(unittest.TestCase):
    def collect(self, fake=None):
        fake = fake or FakeSystem()
        with patch.object(discovery, "_command", side_effect=fake.command), \
             patch.object(discovery, "_presence", return_value="not_found_on_path"), \
             patch.object(discovery, "utc_now", return_value=NOW):
            return discovery.collect()

    def test_installed_disabled_and_loaded_only_timers_are_both_observed(self):
        fake = FakeSystem()
        report = self.collect(fake)
        self.assertEqual(report["collection_status"], "complete")
        inventory = report["timer_inventory"]
        self.assertEqual(inventory["candidate_count"], 2)
        self.assertEqual([c["id"] for c in inventory["candidates"]], ["timer-001", "timer-002"])
        self.assertTrue(all(c["schedule_state"] == "disabled" for c in inventory["candidates"]))
        self.assertTrue(all(c["service_result"] == "success" for c in inventory["candidates"]))
        text = json.dumps(report)
        for private in (PRIVATE_TIMER, PRIVATE_SERVICE, "secret-destination", "private-host", "private-path", "private-description"):
            self.assertNotIn(private, text)
        self.assertEqual(report["production_backup"], discovery.UNKNOWN_BACKUP)
        self.assertEqual(report["coverage_gaps"], discovery.GAPS)
        self.assertTrue(any("--all" in args for args, _ in fake.calls if "list-units" in args))
        self.assertTrue(all(args[0] == "/usr/bin/systemctl" and 0 < timeout <= 4 for args, timeout in fake.calls))
        self.assertTrue(all("ExecStart" not in " ".join(args) and "Environment" not in " ".join(args) for args, _ in fake.calls))

    def test_inventory_is_deduplicated_and_hints_are_bounded_tokens(self):
        fake = FakeSystem()
        fake.loaded += f"{PRIVATE_TIMER} loaded inactive dead private\n"
        fake.installed += "backupper.timer enabled enabled\nmyvocanova.timer disabled enabled\npg_dump-private.timer disabled disabled\nwal-g-private.timer disabled disabled\npostgresql@16.timer static -\n"
        report = self.collect(fake)
        self.assertEqual(report["timer_inventory"]["candidate_count"], 5)
        hints = {hint for item in report["timer_inventory"]["candidates"] for hint in item["name_hints"]}
        self.assertTrue({"vocanova", "backup", "pg_dump", "wal_g", "postgresql", "restic"}.issubset(hints))

    def test_default_success_and_last_trigger_never_prove_execution(self):
        for missing in ("", "n/a", "0"):
            fake = FakeSystem()
            fake.service.update(ExecMainStartTimestamp=missing, ExecMainExitTimestamp=missing)
            for candidate in self.collect(fake)["timer_inventory"]["candidates"]:
                self.assertIsNotNone(candidate["last_trigger_at"])
                self.assertEqual(candidate["service_result"], "unknown")
                self.assertIsNone(candidate["service_finished_at"])

    def test_running_service_cannot_reuse_stale_success(self):
        for active in ("active", "activating", "deactivating", "reloading"):
            fake = FakeSystem()
            fake.service["ActiveState"] = active
            for candidate in self.collect(fake)["timer_inventory"]["candidates"]:
                self.assertEqual(candidate["service_result"], "unknown")
                self.assertIsNone(candidate["service_finished_at"])

    def test_inconsistent_or_malformed_service_metadata_stays_partial_unknown(self):
        for changes in ({"ActiveState": "failed", "Result": "success"},
                        {"ActiveState": "private-host-state"}, {"Result": "private-result-value"}):
            fake = FakeSystem()
            fake.service.update(changes)
            report = self.collect(fake)
            self.assertEqual(report["collection_status"], "partial")
            for candidate in report["timer_inventory"]["candidates"]:
                self.assertEqual(candidate["service_result"], "unknown")
                self.assertEqual(candidate["metadata_status"], "partial")
            self.assertNotIn("private-host-state", json.dumps(report))
            self.assertNotIn("private-result-value", json.dumps(report))

    def test_invalid_missing_future_and_reversed_timestamps_stay_unknown(self):
        variations = [
            {"ExecMainStartTimestamp": ""},
            {"ExecMainStartTimestamp": "private-credential-like-text"},
            {"ExecMainExitTimestamp": "Sat 2026-10-03 10:01:00 UTC"},
            {"ExecMainExitTimestamp": "Fri 2026-10-02 09:00:00 UTC"},
            {"ExecMainExitTimestamp": "Mon 2026-10-02 10:01:00 UTC"},
            {"ExecMainStartTimestamp": "Thu 1970-01-01 00:00:00 UTC", "ExecMainExitTimestamp": "Thu 1970-01-01 00:00:00 UTC"},
        ]
        for updates in variations:
            fake = FakeSystem()
            fake.service.update(updates)
            report = self.collect(fake)
            for candidate in report["timer_inventory"]["candidates"]:
                self.assertEqual(candidate["service_result"], "unknown")
                self.assertIsNone(candidate["service_finished_at"])
            if updates != {"ExecMainStartTimestamp": ""}:
                self.assertEqual(report["collection_status"], "partial")
            self.assertNotIn("private-credential", json.dumps(report))

    def test_service_failure_requires_valid_finished_execution(self):
        fake = FakeSystem()
        fake.service.update(ActiveState="failed", Result="exit-code")
        self.assertEqual(self.collect(fake)["timer_inventory"]["candidates"][0]["service_result"], "failure")
        fake.service["ExecMainExitTimestamp"] = ""
        self.assertEqual(self.collect(fake)["timer_inventory"]["candidates"][0]["service_result"], "unknown")

    def test_timestamp_anomalies_and_hostile_service_names_do_not_expand_reads(self):
        fake = FakeSystem()
        fake.timer.update(Unit="--property=Environment.service", LastTriggerUSec=NEXT)
        report = self.collect(fake)
        self.assertEqual(report["collection_status"], "partial")
        self.assertTrue(all(c["last_trigger_at"] is None for c in report["timer_inventory"]["candidates"]))
        self.assertFalse(any(args[-1].endswith(".service") for args, _ in fake.calls))
        self.assertNotIn("Environment", json.dumps(report))

    def test_partial_inventory_permissions_and_unknown_count_are_explicit(self):
        fake = FakeSystem()
        fake.statuses["installed"] = ("permission_denied", b"raw-private-path", False)
        report = self.collect(fake)
        self.assertEqual(report["collection_status"], "partial")
        self.assertEqual(report["timer_inventory"]["candidate_count"], 1)
        fake.statuses["loaded"] = ("timed_out", b"raw-private-host", False)
        report = self.collect(fake)
        self.assertIsNone(report["timer_inventory"]["candidate_count"])
        self.assertEqual(report["timer_inventory"]["candidates"], [])
        self.assertNotIn("raw-private", json.dumps(report))

    def test_malformed_inventory_never_becomes_complete_zero_coverage(self):
        fake = FakeSystem()
        fake.installed = "ssh-banner-private-token\n../backup.timer enabled enabled\n"
        fake.loaded = ""
        report = self.collect(fake)
        self.assertEqual(report["timer_inventory"]["installed_status"], "invalid_output")
        self.assertEqual(report["collection_status"], "partial")
        self.assertEqual(report["timer_inventory"]["candidate_count"], 0)
        self.assertEqual(len(fake.calls), 2)

    def test_inventory_and_candidate_caps_preserve_partial_coverage(self):
        fake = FakeSystem()
        fake.installed = "".join(f"backup-{index:04d}.timer disabled disabled\n" for index in range(discovery.MAX_ROWS + 1))
        fake.loaded = ""
        report = self.collect(fake)
        inventory = report["timer_inventory"]
        self.assertTrue(inventory["truncated"])
        self.assertEqual(inventory["candidate_count"], discovery.MAX_ROWS)
        self.assertEqual(len(inventory["candidates"]), 20)
        self.assertEqual(report["collection_status"], "partial")
        self.assertEqual(len(fake.calls), 42)

    def test_overall_deadline_stops_new_commands(self):
        fake = FakeSystem()
        with patch.object(discovery, "time") as timer, patch.object(discovery, "_command", side_effect=fake.command), patch.object(discovery, "_presence", return_value="not_found_on_path"):
            timer.monotonic.side_effect = [0, discovery.TOTAL_SECONDS + 1, discovery.TOTAL_SECONDS + 2]
            report = discovery.collect()
        self.assertEqual(fake.calls, [])
        self.assertEqual(report["timer_inventory"]["installed_status"], "timed_out")
        self.assertEqual(report["collection_status"], "partial")

    def test_presence_checks_fixed_path_without_executing_tools(self):
        with patch.object(discovery.shutil, "which", return_value="/private/path/pg_dump") as probe:
            self.assertEqual(discovery._presence("pg_dump"), "found_on_path")
        probe.assert_called_once_with("pg_dump", path=discovery.FIXED_PATH)
        with patch.object(discovery.shutil, "which", side_effect=PermissionError("private-path")):
            self.assertEqual(discovery._presence("pg_dump"), "unavailable")

    def test_command_missing_or_permission_failure_is_sanitized_and_environment_minimal(self):
        for error, expected in ((FileNotFoundError(), "unavailable"), (PermissionError(13, "private-token"), "permission_denied")):
            with patch.object(discovery.subprocess, "Popen", side_effect=error) as command:
                result = discovery._command(["/usr/bin/systemctl", "list-units"], 1)
            self.assertEqual(result, (expected, b"", False))
            kwargs = command.call_args.kwargs
            self.assertEqual(kwargs["env"], discovery.ENV)
            self.assertNotIn("HOME", kwargs["env"])
            self.assertNotIn("SSH_AUTH_SOCK", kwargs["env"])
            self.assertTrue(kwargs["start_new_session"])
            self.assertFalse(kwargs.get("shell", False))

    def stream_command(self, stdout, stderr=b"", code=0, expired=False):
        # Fake pipes/select events exercise the actual streaming loop without a
        # subprocess, host inventory, temporary remote files or systemctl call.
        class Pipe:
            def __init__(self, descriptor):
                self.descriptor, self.closed = descriptor, False
            def fileno(self):
                return self.descriptor
            def close(self):
                self.closed = True
        class Process:
            def __init__(self):
                self.stdout, self.stderr = Pipe(31), Pipe(32)
                self.killed, self.finished = False, False
            def poll(self):
                return code if self.finished else None
            def kill(self):
                self.killed = True
            def wait(self, timeout):
                self.finished = True
                return code
        class Selector:
            def __init__(self):
                self.items = {}
            def register(self, pipe, _events, role):
                self.items[pipe] = SimpleNamespace(fileobj=pipe, data=role)
            def unregister(self, pipe):
                self.items.pop(pipe)
            def get_map(self):
                return self.items
            def select(self, _timeout):
                return [(next(iter(self.items.values())), 1)]
            def close(self):
                self.items.clear()
        process = Process()
        chunks = {31: [stdout[i:i + 16384] for i in range(0, len(stdout), 16384)] + [b""], 32: [stderr, b""]}
        with patch.object(discovery.subprocess, "Popen", return_value=process), \
             patch.object(discovery.selectors, "DefaultSelector", Selector), \
             patch.object(discovery.os, "read", side_effect=lambda fd, _size: chunks[fd].pop(0)), \
             patch.object(discovery.time, "monotonic", side_effect=([0, 2] if expired else lambda: 0)):
            result = discovery._command(["/usr/bin/systemctl", "list-units"], 1)
        self.assertTrue(process.stdout.closed and process.stderr.closed)
        return result, process

    def test_stream_output_cap_kills_child_and_keeps_only_complete_rows(self):
        row = b"backup.timer disabled disabled\n"
        result, process = self.stream_command(row * (discovery.MAX_OUTPUT // len(row) + 2))
        self.assertEqual(result[0], "observed")
        self.assertTrue(result[2])
        self.assertLessEqual(len(result[1]), discovery.MAX_OUTPUT)
        self.assertTrue(result[1].endswith(b"\n"))
        self.assertEqual(set(result[1].splitlines()), {row.strip()})
        self.assertTrue(process.killed)
        result, _ = self.stream_command(b"x" * (discovery.MAX_OUTPUT + 1))
        self.assertEqual(result, ("observed", b"", True))

    def test_stream_timeout_and_error_output_are_bounded_and_private(self):
        result, process = self.stream_command(b"", expired=True)
        self.assertEqual(result, ("timed_out", b"", False))
        self.assertTrue(process.killed)
        result, _ = self.stream_command(b"raw-private-name", b"Permission denied: /private/token", code=1)
        self.assertEqual(result, ("permission_denied", b"", False))
        result, process = self.stream_command(b"", b"private" * 1000, code=1)
        self.assertEqual(result, ("invalid_output", b"", True))
        self.assertTrue(process.killed)

    def test_truncated_and_unexpected_property_output_stays_partial(self):
        fake = FakeSystem()
        fake.statuses["timer"] = ("observed", b"Unit=private.service\n", True)
        report = self.collect(fake)
        self.assertEqual(report["collection_status"], "partial")
        self.assertTrue(all(c["metadata_status"] == "partial" for c in report["timer_inventory"]["candidates"]))
        fake.statuses["timer"] = ("observed", b"Environment=private-credential\n", False)
        report = self.collect(fake)
        self.assertEqual(report["collection_status"], "partial")
        self.assertNotIn("private-credential", json.dumps(report))

    def test_validator_rejects_extra_fields_types_claims_and_timestamp_contradictions(self):
        valid = self.collect()
        self.assertIs(discovery.validate_report(valid), valid)
        changes = [
            lambda r: r.update(schema_version=True),
            lambda r: r.update(extra="private-data"),
            lambda r: r.update(observed_at="2026-02-30T00:00:00Z"),
            lambda r: r["production_backup"].update(association="verified"),
            lambda r: r["coverage_gaps"].pop(),
            lambda r: r["tool_presence"].update(pg_dump=[]),
            lambda r: r["timer_inventory"].update(candidate_count=True),
            lambda r: r["timer_inventory"].update(candidate_count=99999),
            lambda r: r["timer_inventory"].update(loaded_status="unavailable"),
            lambda r: r["timer_inventory"]["candidates"][0].update(unit_name=PRIVATE_TIMER),
            lambda r: r["timer_inventory"]["candidates"][0].update(id="timer-002"),
            lambda r: r["timer_inventory"]["candidates"][0].update(name_hints=["backup", "backup"]),
            lambda r: r["timer_inventory"]["candidates"][0].update(service_finished_at=None),
            lambda r: r["timer_inventory"]["candidates"][0].update(service_finished_at="2026-10-03T10:00:00Z"),
            lambda r: r["timer_inventory"]["candidates"][0].update(next_trigger_at="2026-10-01T10:00:00Z"),
        ]
        for mutate in changes:
            changed = copy.deepcopy(valid)
            mutate(changed)
            with self.assertRaisesRegex(ValueError, "^invalid_backup_discovery_report$"):
                discovery.validate_report(changed)

    def test_main_internal_failure_emits_only_fixed_validated_schema(self):
        stdout = io.StringIO()
        with patch.object(discovery, "collect", side_effect=RuntimeError("private-secret-path")), patch.object(discovery.sys, "argv", ["-"]), contextlib.redirect_stdout(stdout):
            discovery.main()
        report = json.loads(stdout.getvalue())
        self.assertEqual(report["collection_status"], "failed")
        discovery.validate_report(report)
        self.assertNotIn("private-secret", stdout.getvalue())


if __name__ == "__main__":
    unittest.main()
