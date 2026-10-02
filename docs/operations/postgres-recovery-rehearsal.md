# Synthetic PostgreSQL recovery rehearsal

This tool verifies a logical dump and restore of Vocanova's current schema,
canonical content and synthetic learning records. It creates two disposable local
PostgreSQL containers. A passing report is evidence about this rehearsal; it does
not establish that production backups exist or can be recovered.

## Run locally

Use Linux or WSL with Bash, Python 3, Git, Docker and access to
`unix:///var/run/docker.sock`. Run from a trusted checkout. The runner never pulls
images; explicitly cache the official PostgreSQL 16 image first. Its immutable
local image ID and tool versions are recorded in every successful run.

```bash
docker --host unix:///var/run/docker.sock pull postgres:16-alpine
report_parent="$(mktemp -d /tmp/vocanova-recovery-evidence.XXXXXX)"
bash infra/scripts/rehearse-postgres-restore.sh \
  --report-dir "$report_parent/round-trip"
cat "$report_parent/round-trip/report.json"
```

The output directory must be new and its parent must exist. Existing files,
directories and symlinks are preserved. The directory is private (`0700`), and
`report.json` is private (`0600`). Keep the report with the revision it describes;
the dump itself is temporary and removed during cleanup. A dirty checkout is
explicitly marked, with hashes of the migrations, seed and verification fixtures.

Only `--report-dir`, `--fault` and help are accepted. There are no host, DSN,
container, credentials or imported-archive options. Ambient Docker contexts,
remote endpoints and PostgreSQL credentials are not inherited. Both containers
use the same cached image ID, no network, no published ports, no host mounts and
temporary in-memory database storage. Do not repurpose the fixture SQL as an
operator command against another database.

## What a pass establishes

The runner applies every forward migration in filename order, loads the actual
canonical seed through fixed column mappings and inserts the
[synthetic learner fixture](../../infra/recovery/README.md). It creates a
custom-format archive and restores it into a separate empty database in one
transaction, with errors terminating the restore.

Verification compares the complete schema dump, every application table's row
count and sorted-content digest, and sequence state. Independent expectations
check nonempty saved vocabulary, review schedules and history, settings, sentence
history with prewritten synthetic feedback, missions, activity and points.
The target must also allow a valid settings write and reject selected invalid
writes through the restored constraints and append-only trigger. Those probes
roll back, and the original state is checked again before comparison.

The report contains identities, timestamps, durations, input/archive hashes,
counts, check outcomes and cleanup status. It excludes learner rows, SQL error
details, provider output and credentials. Each subprocess is time bounded. A
SIGINT or SIGTERM produces failure evidence and starts cleanup. Cleanup verifies
both the unique run label and requested container name before removing anything;
an unresolved creation or cleanup failure prevents PASS. SIGKILL, daemon failure
or machine loss can prevent cleanup and require local inspection using the
report's run ID; never use a broad container prune as recovery for this tool.

## Failure controls and automation

Run the complete five-case acceptance suite explicitly:

```bash
evidence_parent="$(mktemp -d /tmp/vocanova-recovery-tests.XXXXXX)"
VOCANOVA_REHEARSAL_REPORT_ROOT="$evidence_parent/cases" \
  python3 infra/scripts/test_rehearse_postgres_restore_live.py
```

It requires a real local Docker daemon and cached image; missing prerequisites
fail the suite instead of skipping it. Cases cover a successful restore,
corrupted archive, truncated archive, a removed activity record and SIGTERM
cleanup. The negative cases must produce the expected failed report while the
test suite itself succeeds. Without the report-root environment variable, the
test suite removes its reports after completion.

For a single negative control, pass `--fault corrupt-archive`,
`--fault truncated-archive` or `--fault missing-record` to the runner, using a new
report directory. The first two must fail at `restore`; the missing record must
fail at `invariants` and show the differing table count. Exit 0 means a successful
rehearsal and cleanup, exit 1 means a recorded rehearsal/cleanup failure, and exit
2 means invalid arguments, output reservation or report-writing failure.

Foundation tests exercise argument rejection, output preservation, isolation,
error redaction, interruption and cleanup failure without Docker. The
[recovery workflow](../../.github/workflows/recovery-rehearsal.yml) runs the real
suite for changes to migrations, canonical seed or this harness, and supports
manual dispatch. It retains only sanitized JSON reports for 14 days. It is a
supplemental check, not a required merge-queue status.

## Evidence still needed for production recovery

Production acceptance needs an identified backup mechanism, successful recent
backup, documented schedule and retention, separate storage, protected access,
and an isolated restore from a real recovery point. Verify the chosen release
against restored data, measure recovery time and data age against agreed targets,
and retain evidence of alert receipt and rollback. Use an approved private
environment for real learner records; this synthetic runner does not import them.

The rehearsal omits production roles, ownership, permissions, extensions beyond
the committed schema, point-in-time recovery and application/provider execution.
Local elapsed time is not a production RTO measurement. PostgreSQL documents
[logical dump scope and formats](https://www.postgresql.org/docs/16/app-pgdump.html)
and [transactional restore options](https://www.postgresql.org/docs/16/app-pgrestore.html).
Docker documents the [isolated network mode](https://docs.docker.com/engine/network/drivers/none/)
used here.
