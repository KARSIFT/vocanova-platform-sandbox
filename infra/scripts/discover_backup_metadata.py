#!/usr/bin/env python3
"""Bounded read-only system timer leads; no production backup association."""
import datetime
import errno
import json
import os
import re
import selectors
import shutil
import subprocess
import sys
import time

FIXED_PATH = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV = {"PATH": FIXED_PATH, "LANG": "C", "LC_ALL": "C", "TZ": "UTC", "SYSTEMD_COLORS": "0"}
TOOLS = {"pg_dump": "pg_dump", "pgbackrest": "pgbackrest", "wal_g": "wal-g", "restic": "restic", "borg": "borg"}
GAPS = ["cron_not_inspected", "user_timers_not_inspected", "containers_not_inspected", "provider_backups_not_inspected", "configuration_and_archive_contents_not_inspected", "production_association_unverified"]
UNKNOWN_BACKUP = {"association": "unverified", "retained_artifacts": None, "separate_storage": "unknown", "restore_record": "unknown", "alert_receipt": "unknown"}
INVENTORY_STATUSES = {"observed", "unavailable", "permission_denied", "invalid_output", "timed_out"}
HINT_PATTERNS = {
    "vocanova": "vocanova", "postgresql": "postgres(?:ql)?", "pg_dump": "pg_?dump",
    "pgbackrest": "pgbackrest", "wal_g": "wal[-_]g", "restic": "restic", "borg": "borg", "backup": "backup",
}
UNIT_PATTERN = re.compile(r"[A-Za-z0-9][A-Za-z0-9_.:@-]{0,240}\.(timer|service)\Z")
MAX_OUTPUT = 128 * 1024
MAX_ROWS = 1024
MAX_CANDIDATES = 20
TOTAL_SECONDS = 45
COMMAND_SECONDS = 4
TIMER_PROPERTIES = ("UnitFileState", "ActiveState", "LastTriggerUSec", "NextElapseUSecRealtime", "Unit")
SERVICE_PROPERTIES = ("ActiveState", "Result", "ExecMainStartTimestamp", "ExecMainExitTimestamp")
ACTIVE_STATES = {"active", "reloading", "inactive", "failed", "activating", "deactivating", "maintenance", "refreshing"}
UNIT_FILE_STATES = {"", "enabled", "enabled-runtime", "linked", "linked-runtime", "alias", "static", "disabled", "masked", "masked-runtime", "generated", "transient", "indirect", "bad"}
FAILURE_RESULTS = {"exit-code", "signal", "core-dump", "watchdog", "timeout", "start-limit-hit", "resources", "protocol", "oom-kill", "exec-condition"}


def utc_now():
    return datetime.datetime.now(datetime.timezone.utc)


def iso(value):
    return value.isoformat(timespec="seconds").replace("+00:00", "Z")


def empty_report():
    return {
        "schema_version": 1, "scope": "backup_leads_only", "observed_at": iso(utc_now()),
        "collection_status": "failed", "tool_presence": {key: "unavailable" for key in TOOLS},
        "timer_inventory": {"installed_status": "unavailable", "loaded_status": "unavailable", "truncated": False, "candidate_count": None, "candidates": []},
        "production_backup": dict(UNKNOWN_BACKUP), "coverage_gaps": list(GAPS),
    }


def _command(args, timeout):
    """Return fixed status, bounded stdout, truncation; never expose raw errors."""
    process = None
    streams = selectors.DefaultSelector()
    output, errors = bytearray(), bytearray()
    deadline = time.monotonic() + timeout
    try:
        process = subprocess.Popen(args, stdin=subprocess.DEVNULL, stdout=subprocess.PIPE,
                                   stderr=subprocess.PIPE, env=ENV, start_new_session=True)
        streams.register(process.stdout, selectors.EVENT_READ, "stdout")
        streams.register(process.stderr, selectors.EVENT_READ, "stderr")
        while streams.get_map():
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                return "timed_out", b"", False
            for key, _ in streams.select(remaining):
                chunk = os.read(key.fileobj.fileno(), 16384)
                if not chunk:
                    streams.unregister(key.fileobj)
                    continue
                buffer = output if key.data == "stdout" else errors
                limit = MAX_OUTPUT if key.data == "stdout" else 4096
                remaining_bytes = limit - len(buffer)
                buffer.extend(chunk[:remaining_bytes])
                if len(chunk) > remaining_bytes:
                    if key.data == "stdout":
                        # Only complete leading rows are eligible for inventory.
                        return "observed", bytes(output[:output.rfind(b"\n") + 1]), True
                    return "invalid_output", b"", True
        code = process.wait(timeout=max(0.001, deadline - time.monotonic()))
        if code:
            denied = any(token in errors.lower() for token in (b"permission denied", b"access denied", b"not authorized", b"authentication is required"))
            return ("permission_denied" if denied else "unavailable"), b"", False
        return "observed", bytes(output), False
    except subprocess.TimeoutExpired:
        return "timed_out", b"", False
    except OSError as error:
        return ("permission_denied" if error.errno in (errno.EACCES, errno.EPERM) else "unavailable"), b"", False
    finally:
        streams.close()
        if process is not None:
            if process.poll() is None:
                process.kill()
            try:
                process.wait(timeout=1)
            except subprocess.TimeoutExpired:
                pass
            for pipe in (process.stdout, process.stderr):
                if pipe is not None:
                    pipe.close()


def _presence(binary):
    try:
        return "found_on_path" if shutil.which(binary, path=FIXED_PATH) else "not_found_on_path"
    except OSError:
        return "unavailable"


def _hints(unit):
    stem = unit[:-6].lower()
    return [name for name, pattern in HINT_PATTERNS.items() if re.search(r"(?:^|[._@-])(?:" + pattern + r")(?:$|[._@-])", stem)]


def _unit(value, suffix):
    return type(value) is str and UNIT_PATTERN.fullmatch(value) is not None and value.endswith(suffix)


def _schedule(value):
    return {"enabled": "enabled", "enabled-runtime": "enabled", "disabled": "disabled", "static": "static", "masked": "masked", "masked-runtime": "masked"}.get(value, "unknown")


def _active(value):
    return value if value in {"active", "inactive", "failed"} else "unknown"


def _inventory(kind, command):
    args = ["/usr/bin/systemctl", "list-unit-files" if kind == "installed" else "list-units", "--type=timer", "--no-legend", "--no-pager", "--plain"]
    if kind == "loaded":
        args.append("--all")
    status, output, truncated = command(args)
    if status != "observed":
        return status, {}, truncated
    try:
        lines = output.decode("utf-8", errors="strict").splitlines()
    except UnicodeError:
        return "invalid_output", {}, truncated
    truncated = truncated or len(lines) > MAX_ROWS
    units = {}
    for line in lines[:MAX_ROWS]:
        if not line.strip():
            continue
        fields = line.split(None, 4)
        if len(fields) < (2 if kind == "installed" else 4) or not _unit(fields[0], ".timer"):
            status = "invalid_output"
            continue
        units[fields[0]] = _schedule(fields[1]) if kind == "installed" else _active(fields[2])
    return status, units, truncated


def _properties(unit, properties, command):
    status, output, truncated = command(["/usr/bin/systemctl", "show", "--no-pager", "--property=" + ",".join(properties), "--", unit])
    if status != "observed":
        return None
    if truncated:
        return {}
    try:
        values = {}
        for line in output.decode("utf-8", errors="strict").splitlines():
            key, value = line.split("=", 1)
            if key not in properties or key in values:
                return {}
            values[key] = value
        return values
    except (UnicodeError, ValueError):
        return {}


def _timestamp(value, now, past_only):
    if value in ("", "n/a", "0"):
        return None, False
    if value is None:
        return None, True
    try:
        parsed = datetime.datetime.strptime(value, "%a %Y-%m-%d %H:%M:%S UTC").replace(tzinfo=datetime.timezone.utc)
        # strptime does not check that the stated weekday matches the date.
        if parsed.strftime("%a %Y-%m-%d %H:%M:%S UTC") != value or parsed.timestamp() <= 0 or (past_only and parsed > now):
            return None, True
        return parsed, False
    except ValueError:
        return None, True


def _candidate(unit, index, installed, loaded, command):
    result = {"id": f"timer-{index:03d}", "name_hints": _hints(unit), "metadata_status": "observed",
              "schedule_state": installed.get(unit, "unknown"), "active_state": loaded.get(unit, "unknown"),
              "last_trigger_at": None, "next_trigger_at": None, "service_result": "unknown", "service_finished_at": None}
    timer = _properties(unit, TIMER_PROPERTIES, command)
    if timer is None:
        result["metadata_status"] = "unavailable"
        return result
    partial = set(timer) != set(TIMER_PROPERTIES) or timer.get("ActiveState") not in ACTIVE_STATES or timer.get("UnitFileState") not in UNIT_FILE_STATES
    result["schedule_state"] = _schedule(timer.get("UnitFileState", ""))
    result["active_state"] = _active(timer.get("ActiveState", ""))
    now = utc_now()
    last, anomaly = _timestamp(timer.get("LastTriggerUSec"), now, True)
    following, next_anomaly = _timestamp(timer.get("NextElapseUSecRealtime"), now, False)
    partial = partial or anomaly or next_anomaly
    if last and following and following <= last:
        following, partial = None, True
    result["last_trigger_at"] = iso(last) if last else None
    result["next_trigger_at"] = iso(following) if following else None
    service_name = timer.get("Unit")
    if not _unit(service_name, ".service"):
        result["metadata_status"] = "partial"
        return result
    service = _properties(service_name, SERVICE_PROPERTIES, command)
    if service is None:
        result["metadata_status"] = "partial"
        return result
    partial = partial or set(service) != set(SERVICE_PROPERTIES) or service.get("ActiveState") not in ACTIVE_STATES or service.get("Result") not in FAILURE_RESULTS | {"", "success"}
    inconsistent_result = service.get("ActiveState") == "failed" and service.get("Result") == "success"
    partial = partial or inconsistent_result
    now = utc_now()
    start, start_anomaly = _timestamp(service.get("ExecMainStartTimestamp"), now, True)
    finish, finish_anomaly = _timestamp(service.get("ExecMainExitTimestamp"), now, True)
    partial = partial or start_anomaly or finish_anomaly
    if start and finish and finish < start:
        finish, partial = None, True
    if start and finish and service.get("ActiveState") in {"inactive", "failed"} and not inconsistent_result:
        result["service_finished_at"] = iso(finish)
        if service.get("Result") == "success":
            result["service_result"] = "success"
        elif service.get("Result") in FAILURE_RESULTS:
            result["service_result"] = "failure"
        else:
            partial = True
    result["metadata_status"] = "partial" if partial else "observed"
    return result


def collect():
    report = empty_report()
    deadline = time.monotonic() + TOTAL_SECONDS
    def command(args):
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            return "timed_out", b"", False
        return _command(args, min(COMMAND_SECONDS, remaining))
    report["tool_presence"] = {key: _presence(binary) for key, binary in TOOLS.items()}
    installed_status, installed, first_truncated = _inventory("installed", command)
    loaded_status, loaded, second_truncated = _inventory("loaded", command)
    names = sorted(unit for unit in set(installed) | set(loaded) if _hints(unit))
    truncated = first_truncated or second_truncated or len(names) > MAX_CANDIDATES
    # This is only the deduplicated candidate set seen within bounded inventories,
    # never a count of all backup jobs or established production mechanisms.
    count = len(names) if installed or loaded or "observed" in {installed_status, loaded_status} else None
    candidates = [_candidate(unit, index, installed, loaded, command) for index, unit in enumerate(names[:MAX_CANDIDATES], 1)]
    report["timer_inventory"] = {"installed_status": installed_status, "loaded_status": loaded_status, "truncated": truncated, "candidate_count": count, "candidates": candidates}
    complete = installed_status == loaded_status == "observed" and not truncated and "unavailable" not in report["tool_presence"].values() and all(item["metadata_status"] == "observed" for item in candidates)
    report["collection_status"] = "complete" if complete else "partial"
    report["observed_at"] = iso(utc_now())
    return validate_report(report)


def _require(condition):
    if not condition:
        raise ValueError("invalid_backup_discovery_report")


def _keys(value, keys):
    _require(type(value) is dict and set(value) == set(keys))


def _enum(value, allowed):
    _require(type(value) is str and value in allowed)


def _iso_timestamp(value):
    _require(type(value) is str and re.fullmatch(r"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]{1,6})?Z", value) is not None)
    try:
        return datetime.datetime.fromisoformat(value[:-1] + "+00:00")
    except ValueError:
        raise ValueError("invalid_backup_discovery_report") from None


def validate_report(value):
    """Reject unknown fields and contradictory evidence; return without coercion."""
    _keys(value, ("schema_version", "scope", "observed_at", "collection_status", "tool_presence", "timer_inventory", "production_backup", "coverage_gaps"))
    _require(type(value["schema_version"]) is int and value["schema_version"] == 1)
    _enum(value["scope"], {"backup_leads_only"})
    observed = _iso_timestamp(value["observed_at"])
    _enum(value["collection_status"], {"complete", "partial", "failed"})
    _keys(value["tool_presence"], TOOLS)
    for present in value["tool_presence"].values():
        _enum(present, {"found_on_path", "not_found_on_path", "unavailable"})
    _keys(value["production_backup"], UNKNOWN_BACKUP)
    _require(value["production_backup"] == UNKNOWN_BACKUP)
    _require(type(value["coverage_gaps"]) is list and value["coverage_gaps"] == GAPS)
    inventory = value["timer_inventory"]
    _keys(inventory, ("installed_status", "loaded_status", "truncated", "candidate_count", "candidates"))
    for key in ("installed_status", "loaded_status"):
        _enum(inventory[key], INVENTORY_STATUSES)
    _require(type(inventory["truncated"]) is bool)
    count, candidates = inventory["candidate_count"], inventory["candidates"]
    _require(count is None or (type(count) is int and 0 <= count <= 2 * MAX_ROWS))
    _require(type(candidates) is list and len(candidates) <= MAX_CANDIDATES)
    _require(not candidates if count is None else count >= len(candidates))
    _require(count is None or count <= MAX_CANDIDATES or inventory["truncated"])
    if not inventory["truncated"] and count is not None:
        _require(count == len(candidates))
    for index, candidate in enumerate(candidates, 1):
        _keys(candidate, ("id", "name_hints", "metadata_status", "schedule_state", "active_state", "last_trigger_at", "next_trigger_at", "service_result", "service_finished_at"))
        _require(candidate["id"] == f"timer-{index:03d}")
        hints = candidate["name_hints"]
        _require(type(hints) is list and 1 <= len(hints) <= len(HINT_PATTERNS))
        _require(all(type(hint) is str and hint in HINT_PATTERNS for hint in hints))
        _require(len(set(hints)) == len(hints))
        _enum(candidate["metadata_status"], {"observed", "partial", "unavailable"})
        _enum(candidate["schedule_state"], {"enabled", "disabled", "static", "masked", "unknown"})
        _enum(candidate["active_state"], {"active", "inactive", "failed", "unknown"})
        _enum(candidate["service_result"], {"success", "failure", "unknown"})
        times = {}
        for field in ("last_trigger_at", "next_trigger_at", "service_finished_at"):
            times[field] = None if candidate[field] is None else _iso_timestamp(candidate[field])
            if field != "next_trigger_at" and times[field] is not None:
                _require(times[field] <= observed)
        if times["last_trigger_at"] and times["next_trigger_at"]:
            _require(times["next_trigger_at"] > times["last_trigger_at"])
        if candidate["service_result"] != "unknown":
            _require(times["service_finished_at"] is not None and candidate["metadata_status"] != "unavailable")
    if value["collection_status"] == "complete":
        _require(inventory["installed_status"] == inventory["loaded_status"] == "observed" and not inventory["truncated"] and count is not None)
        _require("unavailable" not in value["tool_presence"].values() and all(c["metadata_status"] == "observed" for c in candidates))
    return value


def main():
    try:
        report = collect() if len(sys.argv) == 1 else empty_report()
        report = validate_report(report)
    except Exception:
        report = empty_report()
    print(json.dumps(report, separators=(",", ":"), sort_keys=True))


if __name__ == "__main__":
    main()
