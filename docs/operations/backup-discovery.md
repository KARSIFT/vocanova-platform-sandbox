# Read-only backup discovery

This manual check gathers leads when the production backup mechanism is not yet
identified. It observes system timer metadata and the presence of a fixed set of
backup tools on the host. It does not inspect or run a backup, identify which
database a job protects, or establish production recovery readiness.

The [synthetic recovery rehearsal](postgres-recovery-rehearsal.md) separately
tests the committed schema and synthetic learning records. Neither check proves
that production recovery points exist.

## Run the reviewed collector

After the workflow is merged to `main`, an operator with the existing production
environment access can run:

```bash
gh workflow run backup-discovery.yml --ref main
```

The [workflow](../../.github/workflows/backup-discovery.yml) accepts no inputs,
runs only from `main`, checks out the dispatched revision and uses the existing
production SSH secrets through their normal environment injection. It grants
only repository read permission. There is no scheduled run.

The transport requires the supplied known-host keys and uses strict host
verification, a temporary private identity file and bounded batch authentication.
It does not scan or accept a new host key. A missing secret, wrong host key,
connection failure or unavailable remote Python produces a fixed failure outcome.
Investigate configuration privately; do not paste private keys or raw server
diagnostics into an issue or workflow log.

The fixed Python collector is sent through SSH standard input. It does not upload
a bundle, create remote files, install tools, invoke Docker, use `sudo`, run a
backup or restore, change a service, query learner records, read configuration or
archive contents, or send test notifications. Installed tools are detected, never
executed. Commands and the overall collection have time limits.

## Read the evidence

The retained JSON identifies the exact collector revision and source hash. Raw
SSH output and errors are captured privately and never published. The transport
validates every report key, type, enum and timestamp before atomically publishing
the private report. A live observation timestamp must be within five minutes
before receipt and 30 seconds after receipt; stale responses or excessive clock
skew produce `clock_skew_or_stale_report`. Historical job timestamps remain
historical and are allowed inside a fresh observation. The workflow retains only
that sanitized report for seven
days. Existing output is preserved; partial JSON is not published.

`transport_status` describes whether valid evidence was received.
`collection_status` describes coverage of this collector's supported observations.
Neither is a backup-health result. Exit 0 means transport and supported collection
completed; exit 1 means a transport or collection gap; exit 2 means invalid local
inputs or an output problem.

SIGINT/SIGTERM cancellation stops the owned local SSH process and allows
temporary credentials to be removed. Cancellation before valid metadata is
received records the fixed `interrupted` transport failure. If a complete,
validated observation was already received, that evidence remains in the atomic
report, but the command still exits 1 and reports interruption. A complete
transport envelope therefore describes evidence receipt, not the overall run's
success. Forced process or machine loss can prevent local cleanup; the workflow
uses an ephemeral runner.

The report includes:

- Presence on a fixed executable search path for `pg_dump`, `pgbackrest`,
  `wal-g`, `restic` and `borg`. Missing host binaries do not rule out container,
  remote or provider-managed backups.
- Bounded installed and loaded system timer inventories, including disabled
  candidates. Separate coverage statuses distinguish unavailable access, bad
  output, timeouts and truncation from a fully observed inventory.
- Candidates selected by a fixed set of name hints. Tool or product names are
  leads only: a `postgresql` timer need not be a backup timer, and a `backup`
  timer need not protect Vocanova. Arbitrary unit names, paths and descriptions
  are excluded; candidate identifiers are opaque and specific to the observation.
- Selected timer state and activation timestamps, plus limited reported service
  results where an actual completed execution timestamp exists. Timer activation
  does not establish backup completion; a default `Result=success` without a
  completed execution is reported as unknown. Active or inconsistent execution
  metadata cannot establish a finished job result.

Candidate counts describe only the bounded inventory observed. Unknown counts
are `null`; zero means no name-hint candidates were observed in the stated scope.
It does not mean no backups exist. Inspect coverage and truncation alongside
the count. Historical timestamps stay historical; a report does not declare a
retention schedule or freshness target satisfied.

The report deliberately keeps production association, retained artifacts,
separate storage, restore evidence and alert receipt unknown. Cron, user timers,
container schedules, provider-managed backups, configuration contents and
archive contents are outside this collector's coverage.

## Follow a lead

Identify the actual production backup mechanism and its bounded inventory scope
through an authorized private operator source. Then establish its configured
schedule and owner, observed job executions, retained artifact metadata and
separate storage. Keep those claims distinct: a successful job exit or a file's
timestamp does not prove that a recovery point is usable.

Production acceptance still needs an isolated restore of a real recovery point,
application checks against the restored data, measured recovery time/data age,
and authoritative alert receipt. This discovery workflow provides no command to
perform those operations. Record gaps in the [release-readiness plan](../product/release-readiness.md).

## Verification boundary

Offline contracts exercise coverage gaps, observation freshness, timestamp and default-success handling,
strict schema rejection, private output preservation, SSH host verification,
bounded failures and error redaction. These tests do not prove access to the live
host or the presence of any backup. Record the manual run's sanitized artifact
and exact revision separately after dispatch.
