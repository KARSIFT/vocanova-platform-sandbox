#!/usr/bin/env python3
"""Bounded SSH transport for sanitized backup leads; never establishes backups."""

import argparse
from datetime import datetime, timezone
import hashlib
import ipaddress
import json
import os
from pathlib import Path
import re
import selectors
import signal
import subprocess
import sys
import tempfile
import time
import types

MAX_SOURCE_BYTES = 256 * 1024
MAX_OUTPUT_BYTES = 128 * 1024
SSH_TIMEOUT_SECONDS = 90
SUBPROCESS_ENV = {"PATH": "/usr/bin:/bin", "LANG": "C.UTF-8", "LC_ALL": "C.UTF-8", "TZ": "UTC"}
TRANSPORT_FAILURE_CODES = frozenset({
    "ssh_unavailable", "ssh_failed", "ssh_timeout", "output_limit", "invalid_report",
    "clock_skew_or_stale_report", "interrupted", "transport_error",
})


class DiscoveryError(Exception):
    def __init__(self, code):
        self.code = code
        super().__init__(code)


class InputError(DiscoveryError):
    pass


class TransportError(DiscoveryError):
    pass


_cancel_requested = False


def request_cancel(signum, frame):
    # Never raise asynchronously across Popen setup, atomic publication or cleanup.
    global _cancel_requested
    _cancel_requested = True


def check_cancelled():
    if _cancel_requested:
        raise TransportError("interrupted")


class QuietParser(argparse.ArgumentParser):
    def error(self, message):
        # argparse's normal error includes supplied arguments, which may be private.
        raise InputError("invalid_arguments")


def validate_environment(environ):
    names = ("BACKUP_SSH_HOST", "BACKUP_SSH_USER", "BACKUP_SSH_PRIVATE_KEY", "BACKUP_SSH_KNOWN_HOSTS", "GITHUB_SHA")
    if any(not isinstance(environ.get(name), str) or not environ[name].strip() for name in names):
        raise InputError("missing_environment")
    host, user, key, known_hosts, revision = (environ[name] for name in names)
    try:
        ipaddress.ip_address(host)
        valid_host = "%" not in host
    except ValueError:
        valid_host = bool(re.fullmatch(r"(?=.{1,253}\Z)[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?(?:\.[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?)*", host))
    if (not valid_host or not re.fullmatch(r"[A-Za-z_][A-Za-z0-9_-]{0,31}", user)
            or not re.fullmatch(r"[0-9a-fA-F]{40}", revision)
            or len(key.encode("utf-8")) > 64 * 1024
            or len(known_hosts.encode("utf-8")) > 128 * 1024
            or "\0" in key or "\0" in known_hosts):
        raise InputError("invalid_environment")
    return host, user, key, known_hosts, revision


def load_collector():
    path = Path(__file__).resolve().with_name("discover_backup_metadata.py")
    try:
        with path.open("rb") as stream:
            source = stream.read(MAX_SOURCE_BYTES + 1)
        if not source or len(source) > MAX_SOURCE_BYTES:
            raise ValueError("source bounds")
        # Execute only the trusted sibling source, never remote output. Validating
        # with these exact bytes binds the schema to the transmitted source hash.
        module = types.ModuleType("backup_discovery_validator")
        module.__file__ = str(path)
        exec(compile(source, "discover_backup_metadata.py", "exec"), module.__dict__)
        validator = module.validate_report
        if not callable(validator):
            raise ValueError("validator absent")
        return source, validator
    except Exception:
        raise InputError("collector_unavailable") from None


def ssh_command(host, user, key_path, known_hosts_path):
    options = (
        "BatchMode=yes", "StrictHostKeyChecking=yes",
        "UserKnownHostsFile=" + str(known_hosts_path), "GlobalKnownHostsFile=/dev/null",
        "KnownHostsCommand=none", "VerifyHostKeyDNS=no", "UpdateHostKeys=no",
        "IdentitiesOnly=yes", "IdentityAgent=none", "PreferredAuthentications=publickey",
        "PasswordAuthentication=no", "KbdInteractiveAuthentication=no", "NumberOfPasswordPrompts=0",
        "ClearAllForwardings=yes", "ForwardAgent=no", "ForwardX11=no", "Tunnel=no",
        "ProxyCommand=none", "ProxyJump=none", "PermitLocalCommand=no",
        "CanonicalizeHostname=no", "ControlMaster=no", "ControlPath=none",
        "ConnectionAttempts=1", "ConnectTimeout=15", "ServerAliveInterval=15",
        "ServerAliveCountMax=2", "LogLevel=ERROR",
    )
    command = ["/usr/bin/ssh", "-F", "/dev/null", "-T"]
    for option in options:
        command.extend(("-o", option))
    command.extend(("-i", str(key_path), "-l", user, "--", host, "/usr/bin/python3 -I -B -S -"))
    return command


def capture_process(command, source, timeout=SSH_TIMEOUT_SECONDS, output_limit=MAX_OUTPUT_BYTES):
    """Drain both pipes incrementally; output/time limits also cover hostile children."""
    check_cancelled()
    deadline = time.monotonic() + timeout
    process = None
    selector = selectors.DefaultSelector()
    output = bytearray()
    captured = 0
    sent = 0
    try:
        try:
            process = subprocess.Popen(command, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                       stderr=subprocess.PIPE, env=SUBPROCESS_ENV, start_new_session=True)
        except OSError:
            raise TransportError("ssh_unavailable") from None
        for stream, event in ((process.stdin, selectors.EVENT_WRITE),
                              (process.stdout, selectors.EVENT_READ),
                              (process.stderr, selectors.EVENT_READ)):
            os.set_blocking(stream.fileno(), False)
            selector.register(stream, event)
        while selector.get_map():
            check_cancelled()
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                raise TransportError("ssh_timeout")
            for selected, event in selector.select(min(remaining, 0.1)):
                stream = selected.fileobj
                if event == selectors.EVENT_WRITE:
                    try:
                        sent += os.write(stream.fileno(), source[sent:sent + 65536])
                    except BrokenPipeError:
                        sent = len(source)
                    except BlockingIOError:
                        continue
                    if sent == len(source):
                        selector.unregister(stream)
                        stream.close()
                else:
                    try:
                        chunk = os.read(stream.fileno(), min(65536, output_limit + 1))
                    except BlockingIOError:
                        continue
                    if not chunk:
                        selector.unregister(stream)
                        stream.close()
                        continue
                    captured += len(chunk)
                    if captured > output_limit:
                        raise TransportError("output_limit")
                    if stream is process.stdout:
                        output.extend(chunk)
                    # stderr contributes to the bound but is never retained or logged.
        while True:
            check_cancelled()
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                raise TransportError("ssh_timeout")
            try:
                return process.wait(timeout=min(0.1, remaining)), bytes(output)
            except subprocess.TimeoutExpired:
                continue
    finally:
        selector.close()
        if process is not None:
            # Kill the entire isolated local process group, including a child that
            # inherited a pipe. Remote command is read-only and has its own bounds.
            if process.returncode is None:
                try:
                    os.killpg(process.pid, signal.SIGKILL)
                except ProcessLookupError:
                    pass
                process.wait()
            for stream in (process.stdin, process.stdout, process.stderr):
                if stream is not None:
                    stream.close()


def reject_duplicates(pairs):
    value = {}
    for key, item in pairs:
        if key in value:
            raise ValueError("duplicate key")
        value[key] = item
    return value


def reject_constant(value):
    raise ValueError("non-finite JSON")


def utc_now():
    return datetime.now(timezone.utc)


def decode_report(raw, validator, received_at=None):
    try:
        if len(raw) > MAX_OUTPUT_BYTES:
            raise ValueError("output bounds")
        parsed = json.loads(raw.decode("utf-8"), object_pairs_hook=reject_duplicates,
                            parse_constant=reject_constant)
        report = validator(parsed)
    except Exception:
        raise TransportError("invalid_report") from None
    # Schema validation remains usable for historical evidence; only this live
    # transport enforces observation freshness against the local receipt clock.
    observed = datetime.fromisoformat(report["observed_at"].replace("Z", "+00:00"))
    age = ((received_at if received_at is not None else utc_now()) - observed).total_seconds()
    if age < -30 or age > 300:
        raise TransportError("clock_skew_or_stale_report")
    return report


def reserve_directory(path):
    try:
        os.mkdir(path, 0o700)  # Exclusive: existing evidence, symlinks and files all fail.
        os.chmod(path, 0o700, follow_symlinks=False)
        return os.open(path, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW)
    except OSError:
        raise InputError("output_reservation") from None


def publish_report(directory_fd, envelope):
    # A hard link publishes the complete private file atomically and refuses any
    # existing report.json. Directory-relative operations resist path replacement.
    temporary = ".report-" + os.urandom(16).hex()
    try:
        fd = os.open(temporary, os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_NOFOLLOW,
                     0o600, dir_fd=directory_fd)
        with os.fdopen(fd, "wb") as stream:
            os.fchmod(stream.fileno(), 0o600)
            stream.write((json.dumps(envelope, sort_keys=True, separators=(",", ":")) + "\n").encode("utf-8"))
            stream.flush()
            os.fsync(stream.fileno())
        os.link(temporary, "report.json", src_dir_fd=directory_fd, dst_dir_fd=directory_fd,
                follow_symlinks=False)
        os.fsync(directory_fd)
    except OSError:
        raise InputError("report_write_failed") from None
    finally:
        try:
            os.unlink(temporary, dir_fd=directory_fd)
        except FileNotFoundError:
            pass


def collect_over_ssh(host, user, key, known_hosts, source, validator):
    check_cancelled()
    with tempfile.TemporaryDirectory(prefix="vocanova-backup-ssh-", dir="/tmp") as directory:
        directory = Path(directory)
        for name, contents in (("key", key), ("known_hosts", known_hosts)):
            fd = os.open(directory / name, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
            with os.fdopen(fd, "w", encoding="utf-8") as stream:
                os.fchmod(stream.fileno(), 0o600)
                stream.write(contents + ("" if contents.endswith("\n") else "\n"))
        code, output = capture_process(ssh_command(host, user, directory / "key", directory / "known_hosts"), source)
        if code != 0:
            raise TransportError("ssh_failed")
        return decode_report(output, validator)


def run(argv=None):
    directory_fd = None
    try:
        parser = QuietParser(prog="collect-backup-discovery", description=__doc__, allow_abbrev=False)
        parser.add_argument("--report-dir", required=True)
        args = parser.parse_args(argv)
        host, user, key, known_hosts, revision = validate_environment(os.environ)
        source, validator = load_collector()
        check_cancelled()
        directory_fd = reserve_directory(args.report_dir)
        envelope = {"schema_version": 1, "collector_revision": revision,
                    "collector_sha256": hashlib.sha256(source).hexdigest(),
                    "transport_status": "failed", "failure_code": None, "report": None}
        try:
            report = collect_over_ssh(host, user, key, known_hosts, source, validator)
            envelope.update(transport_status="complete", report=report)
        except TransportError as error:
            envelope["failure_code"] = error.code if error.code in TRANSPORT_FAILURE_CODES else "transport_error"
        except Exception:
            envelope["failure_code"] = "transport_error"
        publish_report(directory_fd, envelope)
        if _cancel_requested:
            print("backup discovery: interrupted")
            return 1
        if envelope["transport_status"] == "failed":
            print("backup discovery: " + envelope["failure_code"])
            return 1
        status = report["collection_status"]
        print("backup discovery: collection_" + status)
        return 0 if status == "complete" else 1
    except TransportError:
        print("backup discovery: interrupted")
        return 1
    except InputError as error:
        print("backup discovery: " + error.code)
        return 2
    except Exception:
        print("backup discovery: local_error")
        return 2
    finally:
        if directory_fd is not None:
            os.close(directory_fd)


def main(argv=None):
    global _cancel_requested
    previous_cancel = _cancel_requested
    _cancel_requested = False
    previous_handlers = {signum: signal.signal(signum, request_cancel)
                         for signum in (signal.SIGINT, signal.SIGTERM)}
    try:
        code = run(argv)
    finally:
        # During all resource cleanup handlers still only set a flag. Restore
        # caller handlers after cleanup, then preserve cancellation in the exit.
        for signum, handler in previous_handlers.items():
            signal.signal(signum, handler)
        cancelled = _cancel_requested
        _cancel_requested = previous_cancel
    if cancelled:
        if code != 1:
            print("backup discovery: interrupted")
        return 1
    return code


if __name__ == "__main__":
    sys.exit(main())
