#!/usr/bin/env python3
"""Rehearse recovery of disposable synthetic PostgreSQL data on local Docker only."""
from __future__ import annotations

import argparse
from contextlib import contextmanager
import datetime
import hashlib
import json
import os
from pathlib import Path
import re
import signal
import subprocess
import sys
import tempfile
import time
import uuid

IMAGE = "postgres:16-alpine"
DATABASE = "vocanova_rehearsal"
LABEL = "site.vocanova.restore-rehearsal"
DOCKER_ENDPOINT = "unix:///var/run/docker.sock"
# Deliberately do not inherit Docker contexts, hosts, TLS settings, credentials,
# PostgreSQL options, or shell startup configuration from the invoking process.
COMMAND_ENV = {"PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", "LANG": "C.UTF-8", "LC_ALL": "C.UTF-8"}
SEED_COLUMNS = {
    "journey_situations": ("id", "slug", "title", "short_description", "level_band", "category", "status", "display_order"),
    "canonical_words": ("id", "text", "normalized_text", "word_type", "language_code", "status", "difficulty_level", "frequency_rank"),
    "word_meanings": ("id", "word_id", "part_of_speech", "short_definition", "learner_definition", "meaning_order", "status", "difficulty_level"),
    "word_examples": ("id", "meaning_id", "example_text", "example_order", "difficulty_level", "situation_label", "status"),
    "usage_notes": ("id", "meaning_id", "note_type", "note_text", "note_order", "status"),
    "journey_words": ("id", "journey_situation_id", "meaning_id", "relevance_score", "display_order", "is_core"),
}


class SafeArgumentParser(argparse.ArgumentParser):
    def error(self, _message):
        # Rejected values may be an accidentally pasted DSN or credential.
        self.exit(2, "Restore rehearsal: invalid arguments; use --help.\n")


class RehearsalError(Exception):
    """Only fixed, non-sensitive failure codes cross the command boundary."""

    def __init__(self, code: str):
        super().__init__(code)
        self.code = code


def sha256(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def sql_literal(value):
    if value is None:
        return "NULL"
    if isinstance(value, bool):
        return "TRUE" if value else "FALSE"
    if isinstance(value, int):
        return str(value)
    if isinstance(value, str) and "\x00" not in value:
        # standard_conforming_strings is explicitly enabled before seed inserts.
        return "'" + value.replace("'", "''") + "'"
    raise RehearsalError("unsupported_seed_value")


def seed_sql(data: bytes) -> bytes:
    seed = json.loads(data)
    if not isinstance(seed, dict) or set(seed) != set(SEED_COLUMNS):
        raise RehearsalError("unexpected_seed_tables")
    lines = ["BEGIN;", "SET LOCAL standard_conforming_strings = on;"]
    for table, columns in SEED_COLUMNS.items():
        rows = seed[table]
        if not isinstance(rows, list) or not rows:
            raise RehearsalError("empty_seed_table")
        for row in rows:
            if not isinstance(row, dict) or set(row) - set(columns) or not row.get("id"):
                raise RehearsalError("unexpected_seed_columns")
            # Fixed columns, preserved canonical IDs, deterministic synthetic times.
            values = [sql_literal(row.get(column)) for column in columns]
            values += ["'2026-01-01T00:00:00Z'", "'2026-01-01T00:00:00Z'"]
            names = ", ".join(columns + ("created_at", "updated_at"))
            lines.append(f"INSERT INTO {table} ({names}) VALUES ({', '.join(values)});")
    lines.append("COMMIT;")
    return ("\n".join(lines) + "\n").encode()


def quote_identifier(value: str) -> str:
    return '"' + value.replace('"', '""') + '"'


class Rehearsal:
    def __init__(self, repo_root: Path, report_dir: Path, fault: str = "none"):
        self.repo_root = Path(repo_root)
        self.report_dir = Path(report_dir)
        self.fault = fault
        self.run_id = uuid.uuid4().hex
        self.names = [f"vocanova-restore-{self.run_id}-{role}" for role in ("source", "target")]
        self.requested_names = []
        self.unconfirmed_creates = set()
        self.stage = "preflight"
        self.temp_dir = None
        self.report = {
            "schema_version": 1,
            "scope": "synthetic-only local Docker logical dump/restore rehearsal",
            "not_evidenced": ["production backups", "production roles", "off-host retention", "PITR", "alert delivery"],
            "run_id": self.run_id,
            "started_at": datetime.datetime.now(datetime.timezone.utc).isoformat(),
            "status": "FAIL", "failure_stage": None, "failure_code": None,
            "fault": fault, "commit": "unknown", "dirty": None,
            "identities": {}, "inputs": {}, "archive": {}, "durations": {},
            "tables": {}, "schema": {}, "sequences": {},
            "independent_checks": {}, "cleanup": {"status": "pending"},
        }

    def _command(self, args, *, input_data=None, timeout=60) -> bytes:
        try:
            # Terminal Ctrl-C targets the foreground process group. Keep each
            # bounded child outside that group so deferred cleanup can finish.
            # subprocess.run still kills/waits for its child on timeout or when
            # our active-work signal handler raises; no signal mask is changed.
            result = subprocess.run(args, input=input_data, stdout=subprocess.PIPE,
                                    stderr=subprocess.PIPE, timeout=timeout,
                                    env=COMMAND_ENV, check=False, start_new_session=True)
        except subprocess.TimeoutExpired:
            raise RehearsalError("command_timeout") from None
        except OSError:
            raise RehearsalError("command_unavailable") from None
        if result.returncode:
            raise RehearsalError("command_failed")
        return result.stdout

    def _docker(self, *args, input_data=None, timeout=60):
        return self._command(["docker", "--host", DOCKER_ENDPOINT, "--config",
                              str(self.temp_dir / "docker-config"), *args],
                             input_data=input_data, timeout=timeout)

    @contextmanager
    def _stage(self, name):
        self.stage = name
        start = time.monotonic()
        try:
            yield
        finally:
            self.report["durations"][name] = round(time.monotonic() - start, 3)

    def _psql(self, container, sql, timeout=60):
        if isinstance(sql, str):
            sql = sql.encode()
        return self._docker("exec", "-i", "--env", "PGOPTIONS=-c timezone=UTC -c extra_float_digits=3",
                            container, "psql", "-X", "-q", "-t", "-A",
                            "--set", "ON_ERROR_STOP=1", "--username", "postgres",
                            "--dbname", DATABASE, input_data=sql, timeout=timeout)

    def _create(self, name, image_id):
        # Register intent before the daemon sees the create request: timeout or
        # interruption can leave a container even without a returned ID.
        self.requested_names.append(name)
        self.unconfirmed_creates.add(name)
        self._docker("create", "--pull", "never", "--network", "none",
                     "--name", name, "--label", f"{LABEL}={self.run_id}",
                     "--tmpfs", "/var/lib/postgresql/data:rw,nosuid,noexec,size=512m",
                     "--env", "POSTGRES_HOST_AUTH_METHOD=trust",
                     "--env", f"POSTGRES_DB={DATABASE}", image_id, timeout=60)
        self.unconfirmed_creates.discard(name)
        self._docker("start", name, timeout=30)
        for attempt in range(30):
            try:
                # initdb briefly starts a temporary server. Wait until the
                # entrypoint has exec'd the final postmaster before migrating.
                process = self._docker("exec", name, "cat", "/proc/1/comm", timeout=3)
                if process.strip() != b"postgres":
                    raise RehearsalError("postgres_initializing")
                self._psql(name, "SELECT 1;", timeout=3)
                return
            except RehearsalError as error:
                if error.code == "interrupted":
                    raise
                if attempt == 29:
                    raise RehearsalError("postgres_not_ready") from None
                time.sleep(0.2)

    def _checks(self, container, sql):
        checks = json.loads(self._psql(container, sql))
        if not isinstance(checks, dict) or not checks or any(value is not True for value in checks.values()):
            raise RehearsalError("invariant_results_invalid")
        if any(not re.fullmatch(r"[a-z][a-z0-9_]{0,79}", key) for key in checks):
            raise RehearsalError("invariant_names_invalid")
        return checks

    def _snapshot(self, container, restrict_key):
        # A common per-run restrict key makes the complete schema dump comparable
        # without stripping constraints, functions, triggers or arbitrary SQL.
        schema = self._docker("exec", container, "pg_dump", "--schema-only", "--no-owner",
                              "--no-privileges", "--restrict-key", restrict_key,
                              "--username", "postgres", "--dbname", DATABASE)
        inventory = json.loads(self._psql(container, """
            SELECT coalesce(json_agg(json_build_array(n.nspname, c.relname)
              ORDER BY n.nspname, c.relname), '[]'::json)
            FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
            WHERE c.relkind IN ('r', 'p') AND n.nspname NOT LIKE 'pg_%'
              AND n.nspname <> 'information_schema';
        """))
        if not inventory:
            raise RehearsalError("empty_table_inventory")
        tables = {}
        for namespace, table in inventory:
            qualified = quote_identifier(namespace) + "." + quote_identifier(table)
            # COPY escapes newlines inside values. Each output line represents
            # one complete row; duplicate rows remain in the digest and count.
            rows = self._psql(container, f'COPY (SELECT row_to_json(t)::text FROM {qualified} t ORDER BY row_to_json(t)::text COLLATE "C") TO STDOUT;')
            tables[f"{namespace}.{table}"] = {"count": rows.count(b"\n"), "sha256": sha256(rows)}
        sequence_inventory = json.loads(self._psql(container, """
            SELECT coalesce(json_agg(json_build_array(schemaname, sequencename)
              ORDER BY schemaname, sequencename), '[]'::json)
            FROM pg_sequences WHERE schemaname NOT LIKE 'pg_%'
              AND schemaname <> 'information_schema';
        """))
        sequences = {}
        for namespace, name in sequence_inventory:
            qualified = quote_identifier(namespace) + "." + quote_identifier(name)
            state = self._psql(container, f"SELECT last_value, is_called FROM {qualified};")
            sequences[f"{namespace}.{name}"] = {"sha256": sha256(state)}
        return {"schema_sha256": sha256(schema), "tables": tables, "sequences": sequences}

    def _execute(self):
        with self._stage("preflight"):
            try:
                commit = self._command(["git", "--no-optional-locks", "-C", str(self.repo_root), "rev-parse", "HEAD"]).decode().strip()
            except (RehearsalError, ValueError) as error:
                if isinstance(error, RehearsalError) and error.code == "interrupted":
                    raise
                # Preserve the explicit unknown metadata and fail closed; never
                # substitute a guessed revision for reproducibility evidence.
                raise RehearsalError("commit_identity_unavailable") from None
            if not re.fullmatch(r"[0-9a-f]{40,64}", commit):
                raise RehearsalError("invalid_commit_identity")
            self.report["commit"] = commit
            self.report["dirty"] = bool(self._command(["git", "--no-optional-locks", "-C", str(self.repo_root), "status", "--porcelain"]))
            image_id = self._docker("image", "inspect", IMAGE, "--format", "{{.Id}}").decode().strip()
            if not re.fullmatch(r"sha256:[0-9a-f]{64}", image_id):
                raise RehearsalError("invalid_local_image_identity")
            self.report["identities"]["image"] = {"requested": IMAGE, "local_id": image_id}
            for role in ("Client", "Server"):
                version = self._docker("version", "--format", "{{." + role + ".Version}}").decode().strip()
                if not re.fullmatch(r"[0-9][0-9A-Za-z.+_-]{0,79}", version):
                    raise RehearsalError("invalid_docker_identity")
                self.report["identities"]["docker_" + role.lower()] = version
            paths = sorted((self.repo_root / "apps/api/migrations").glob("*.sql"))
            paths = [path for path in paths if not path.name.endswith(".down.sql")]
            if not paths:
                raise RehearsalError("no_forward_migrations")
            migrations = [(path.name, path.read_bytes()) for path in paths]
            seed = (self.repo_root / "apps/api/cmd/seed/voc026-p1.json").read_bytes()
            recovery = self.repo_root / "infra/recovery"
            fixture = (recovery / "rehearsal-fixture.sql").read_bytes()
            checks = (recovery / "rehearsal-check.sql").read_bytes()
            fault_sql = (recovery / "rehearsal-missing-record.sql").read_bytes()
            self.report["inputs"] = {
                "migrations": [{"file": name, "sha256": sha256(data)} for name, data in migrations],
                "seed_sha256": sha256(seed), "fixture_sha256": sha256(fixture),
                "check_sha256": sha256(checks), "missing_record_sha256": sha256(fault_sql),
            }
            inserts = seed_sql(seed)
        source, target = self.names
        with self._stage("source_create"):
            self._create(source, image_id)
            version = self._psql(source, "SHOW server_version_num;").decode().strip()
            if not re.fullmatch(r"16[0-9]{4}", version):
                raise RehearsalError("postgres_major_mismatch")
            self.report["identities"]["postgres_server_version_num"] = version
            for tool in ("pg_dump", "pg_restore", "psql"):
                version = self._docker("exec", source, tool, "--version").decode().strip()
                if not re.fullmatch(r"[A-Za-z_]+ \(PostgreSQL\) 16\.[0-9]+(?: [A-Za-z0-9().+ _-]+)?", version):
                    raise RehearsalError("postgres_tool_identity_invalid")
                self.report["identities"][tool] = version
            # Inspect the cached binary's capability, not just its major version.
            # Older cached minor releases may lack this schema-dump option.
            if b"--restrict-key" not in self._docker("exec", source, "pg_dump", "--help"):
                raise RehearsalError("pg_dump_restrict_key_unsupported")
            self.report["identities"]["pg_dump_restrict_key_supported"] = True
        with self._stage("migrate"):
            for _, migration in migrations:
                self._psql(source, migration)
        with self._stage("seed"):
            self._psql(source, inserts)
        with self._stage("fixture"):
            self._psql(source, fixture)
            self.report["independent_checks"]["source"] = self._checks(source, checks)
        restrict_key = uuid.uuid4().hex
        with self._stage("source_snapshot"):
            before = self._snapshot(source, restrict_key)
            # Preserve measured source evidence even if archive/restore fails.
            self.report["schema"] = {"source_sha256": before["schema_sha256"], "target_sha256": None, "match": None}
            self.report["tables"] = {name: {"source": value, "target": None, "match": None} for name, value in before["tables"].items()}
            self.report["sequences"] = {"source": before["sequences"], "target": None, "match": None}
        with self._stage("archive"):
            archive = self._docker("exec", source, "pg_dump", "--format=custom", "--no-owner",
                                   "--no-privileges", "--username", "postgres", "--dbname", DATABASE)
            if not archive.startswith(b"PGDMP") or len(archive) < 16:
                raise RehearsalError("invalid_dump_archive")
            self.report["archive"] = {"sha256": sha256(archive), "size_bytes": len(archive)}
            if self.fault == "corrupt-archive":
                archive = b"BROKEN" + archive[6:]
            elif self.fault == "truncated-archive":
                archive = archive[:len(archive) // 2]
            self.report["archive"].update(restore_input_sha256=sha256(archive), restore_input_size_bytes=len(archive))
            archive_path = self.temp_dir / "synthetic.dump"
            descriptor = os.open(archive_path, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
            with os.fdopen(descriptor, "wb") as handle:
                handle.write(archive)
        with self._stage("target_create"):
            self._create(target, image_id)
        with self._stage("restore"):
            self._docker("exec", "-i", target, "pg_restore", "--single-transaction", "--exit-on-error",
                         "--no-owner", "--no-privileges", "--username", "postgres", "--dbname", DATABASE,
                         input_data=archive_path.read_bytes(), timeout=120)
        with self._stage("fault_injection"):
            if self.fault == "missing-record":
                self._psql(target, fault_sql)
        with self._stage("invariants"):
            invariant_failure = False
            try:
                self.report["independent_checks"]["target"] = self._checks(target, checks)
            except (RehearsalError, ValueError) as error:
                if isinstance(error, RehearsalError) and error.code == "interrupted":
                    raise
                self.report["independent_checks"]["target"] = {"status": "FAIL", "code": "invariant_check_failed"}
                invariant_failure = True
        with self._stage("target_snapshot"):
            after = self._snapshot(target, restrict_key)
        with self._stage("compare"):
            self.report["schema"] = {"source_sha256": before["schema_sha256"], "target_sha256": after["schema_sha256"], "match": before["schema_sha256"] == after["schema_sha256"]}
            for table in sorted(set(before["tables"]) | set(after["tables"])):
                left, right = before["tables"].get(table), after["tables"].get(table)
                self.report["tables"][table] = {"source": left, "target": right, "match": left == right}
            self.report["sequences"] = {"source": before["sequences"], "target": after["sequences"], "match": before["sequences"] == after["sequences"]}
            if invariant_failure:
                self.stage = "invariants"
                raise RehearsalError("invariant_check_failed")
            if before != after:
                raise RehearsalError("restore_comparison_failed")
            if self.fault != "none":
                raise RehearsalError("fault_not_detected")
        self.report["status"] = "PASS"

    def _cleanup(self):
        removed = 0
        if not self.requested_names:
            return {"status": "PASS", "removed_containers": 0, "verified_absent": True}
        # Query only this unpredictable ownership label. Never delete by a
        # generic prefix, image, age, or an unverified name.
        for attempt in range(3):
            ids = self._docker("ps", "--all", "--no-trunc", "--filter", f"label={LABEL}={self.run_id}", "--format", "{{.ID}}", timeout=15).decode().split()
            for container_id in ids:
                if not re.fullmatch(r"[0-9a-f]{64}", container_id):
                    raise RehearsalError("cleanup_identity_invalid")
                metadata = self._docker("inspect", "--format", '{{.Name}} {{index .Config.Labels "' + LABEL + '"}}', container_id, timeout=15).decode().strip().split()
                if len(metadata) != 2 or metadata[0].lstrip("/") not in self.requested_names or metadata[1] != self.run_id:
                    raise RehearsalError("cleanup_ownership_mismatch")
                self._docker("rm", "--force", "--volumes", container_id, timeout=30)
                removed += 1
                self.unconfirmed_creates.discard(metadata[0].lstrip("/"))
            if attempt < 2:
                time.sleep(0.2)
        remaining = self._docker("ps", "--all", "--filter", f"label={LABEL}={self.run_id}", "--format", "{{.ID}}", timeout=15).strip()
        if remaining:
            raise RehearsalError("cleanup_resources_remaining")
        if self.unconfirmed_creates:
            # A timed-out create may still be completing in the daemon. No
            # matching container observed is insufficient to certify cleanup.
            raise RehearsalError("creation_completion_unconfirmed")
        return {"status": "PASS", "removed_containers": removed, "verified_absent": True}

    def run(self):
        started = time.monotonic()
        old_handlers = {}
        temporary = None
        cleaning_up = False
        interruption_pending = False

        def interrupt(_signum, _frame):
            nonlocal interruption_pending
            # Cleanup must finish, but a first cancellation arriving during
            # container or temporary-file cleanup must still prevent PASS.
            self.report["status"] = "FAIL"
            if self.report["failure_stage"] is None:
                self.report.update(failure_stage="cleanup" if cleaning_up else self.stage,
                                   failure_code="interrupted")
            if cleaning_up or interruption_pending:
                return
            interruption_pending = True
            raise RehearsalError("interrupted")

        for sig in (signal.SIGINT, signal.SIGTERM):
            old_handlers[sig] = signal.signal(sig, interrupt)
        try:
            temporary = tempfile.TemporaryDirectory(prefix="vocanova-restore-")
            self.temp_dir = Path(temporary.name)
            (self.temp_dir / "docker-config").mkdir(mode=0o700)
            self._execute()
        except RehearsalError as error:
            self.report.update(status="FAIL", failure_stage=self.stage, failure_code=error.code)
        except (OSError, ValueError, TypeError, KeyError):
            self.report.update(status="FAIL", failure_stage=self.stage, failure_code="invalid_local_input_or_result")
        finally:
            # Keep the same handler installed throughout cleanup: cancellation
            # is recorded, while further exceptions are deferred until it ends.
            cleaning_up = True
            try:
                with self._stage("cleanup"):
                    try:
                        self.report["cleanup"] = self._cleanup()
                    except (RehearsalError, OSError, ValueError, TypeError, KeyError) as error:
                        code = error.code if isinstance(error, RehearsalError) else "cleanup_result_invalid"
                        self.report["cleanup"] = {"status": "FAIL", "code": code, "verified_absent": False}
                        if self.report["failure_stage"] is None:
                            self.report.update(failure_stage="cleanup", failure_code=code)
                        self.report["status"] = "FAIL"
                    try:
                        if temporary is not None:
                            temporary.cleanup()
                        self.report["cleanup"]["temporary_files_removed"] = True
                    except OSError:
                        self.report["cleanup"].update(status="FAIL", temporary_files_removed=False)
                        self.report["status"] = "FAIL"
                        if self.report["failure_stage"] is None:
                            self.report.update(failure_stage="cleanup", failure_code="temporary_cleanup_failed")
                # Keep deferred cancellation active through final metadata too.
                self.report["durations"]["total"] = round(time.monotonic() - started, 3)
                self.report["finished_at"] = datetime.datetime.now(datetime.timezone.utc).isoformat()
            finally:
                for sig, handler in old_handlers.items():
                    signal.signal(sig, handler)
        return self.report


class ReportCancellation:
    """Defer late CLI cancellation until a complete failed report is published."""

    def __init__(self):
        self.report = None
        self.interrupted = False
        self.complete = False

    def bind(self, report):
        self.report = report
        if self.interrupted:
            self.record()

    def record(self, _signum=None, _frame=None):
        if self.complete:
            return
        self.interrupted = True
        if self.report is not None:
            self.report["status"] = "FAIL"
            if self.report.get("failure_stage") is None:
                self.report.update(failure_stage="report_persistence", failure_code="interrupted")


def write_report_atomic(report_dir: Path, report: dict, cancellation: ReportCancellation):
    """Publish only complete JSON; cancellation can require at most one rewrite.

    The first publication is exclusive. A retry may replace only the file this
    call published, never a preexisting report. Signals remain deferred during
    serialization and fsync. A short masked section checks pending signals on
    both sides of atomic publication, then closes the completed operation.
    """
    cancellation.bind(report)
    destination = report_dir / "report.json"
    published_identity = None
    watched = {signal.SIGINT, signal.SIGTERM}
    try:
        for _ in range(2):
            interrupted_before = cancellation.interrupted
            descriptor, filename = tempfile.mkstemp(prefix=".report-", suffix=".json", dir=report_dir)
            temporary = Path(filename)
            try:
                with os.fdopen(descriptor, "w", encoding="utf-8") as handle:
                    json.dump(report, handle, indent=2, sort_keys=True)
                    handle.write("\n")
                    handle.flush()
                    os.fsync(handle.fileno())
                previous_mask = signal.pthread_sigmask(signal.SIG_BLOCK, watched)
                try:
                    if signal.sigpending() & watched:
                        cancellation.record()
                    if interrupted_before != cancellation.interrupted:
                        # Do not publish a snapshot serialized before cancellation.
                        continue
                    identity = temporary.stat()
                    if published_identity is None:
                        os.link(temporary, destination)
                    else:
                        existing = destination.lstat()
                        if (existing.st_dev, existing.st_ino) != published_identity:
                            raise RehearsalError("report_identity_changed")
                        os.replace(temporary, destination)
                    published_identity = (identity.st_dev, identity.st_ino)
                    # Finish removing our sibling before declaring persistence done;
                    # cancellation during this final cleanup is still recorded.
                    temporary.unlink(missing_ok=True)
                    temporary = None
                    if signal.sigpending() & watched:
                        cancellation.record()
                    if interrupted_before != cancellation.interrupted:
                        # A signal during publication needs one atomic FAIL rewrite.
                        continue
                    # Publication is complete. Signals arriving after this point
                    # cannot alter a completed report or interrupt its persistence.
                    cancellation.complete = True
                    return
                finally:
                    signal.pthread_sigmask(signal.SIG_SETMASK, previous_mask)
            finally:
                if temporary is not None:
                    temporary.unlink(missing_ok=True)
        raise RehearsalError("report_publication_incomplete")
    except (OSError, ValueError, TypeError, RehearsalError):
        # A signal after the first publication can require a corrective FAIL
        # rewrite. If that write fails, do not leave our earlier PASS behind.
        # Never remove an output whose identity this helper did not publish.
        if published_identity is not None and not cancellation.complete:
            try:
                existing = destination.lstat()
                if (existing.st_dev, existing.st_ino) == published_identity:
                    destination.unlink()
            except FileNotFoundError:
                pass
        raise


def main(argv=None):
    parser = SafeArgumentParser(description=__doc__, allow_abbrev=False)
    parser.add_argument("--report-dir", required=True, type=Path, help="new private output directory; never overwrites an existing path")
    parser.add_argument("--fault", choices=("corrupt-archive", "truncated-archive", "missing-record"), default="none")
    args = parser.parse_args(argv)
    os.umask(0o077)
    try:
        args.report_dir.mkdir(mode=0o700, parents=False, exist_ok=False)
    except OSError:
        print("Restore rehearsal: output directory must be new and writable.", file=sys.stderr)
        return 2
    cancellation = ReportCancellation()
    old_handlers = {sig: signal.signal(sig, cancellation.record) for sig in (signal.SIGINT, signal.SIGTERM)}
    try:
        runner = Rehearsal(Path(__file__).resolve().parents[2], args.report_dir, args.fault)
        cancellation.bind(runner.report)
        report = runner.run()
        write_report_atomic(args.report_dir, report, cancellation)
        print(f"Restore rehearsal: {report['status']} (synthetic data only).")
        return 0 if report["status"] == "PASS" else 1
    except (OSError, ValueError, TypeError, RehearsalError):
        print("Restore rehearsal: could not complete or write the private report.", file=sys.stderr)
        return 2
    finally:
        for sig, handler in old_handlers.items():
            signal.signal(sig, handler)


if __name__ == "__main__":
    raise SystemExit(main())
