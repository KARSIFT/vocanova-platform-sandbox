# VOC-094 — Fix deploy-staging cancellation from concurrency queue supersession: Specification

## Objective and requirement source

Remediate the operational failure recorded in
[GitHub issue #781](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/781):
`deploy-staging` ended `cancelled` because GitHub Actions superseded a **pending**
queued run when a newer push entered the `staging-deploy` concurrency group.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

Primary evidence (issue #781 + public run page and Actions REST metadata for run
32290409156):

| Item | Value |
|------|-------|
| Workflow | `deploy-staging` |
| Run | [#299](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32290409156), push-triggered 2026-08-19 18:58 UTC |
| Head branch / SHA | `develop` @ `411fb60157d437e495fc07f599e268669f139e5a` |
| Conclusion | `cancelled` (workflow wall clock ~2m 28s) |
| Public annotation | Canceling since a higher priority waiting request for **staging-deploy** exists |
| Jobs started | **0** (`/actions/runs/32290409156/jobs` reports `total_count: 0`) |
| Issue origin | VOC-088-T02 `operational-failure-monitoring.yml` sanitized issue body |

Drafting-time repo read of `deploy-staging.yml`:

```yaml
concurrency:
  group: staging-deploy
  cancel-in-progress: false
```

The workflow header documents why `cancel-in-progress: false` is required: cancelling
an in-progress deploy after `docker compose pull` but before `up -d` would leave
new images local but not running. GitHub's **default** concurrency queue, however,
permits only one pending run; when a third event arrives, the older pending run is
cancelled — producing the annotation above without any deploy step executing.

This is distinct from VOC-090 (scheduled-synthetics job **timeout** cancellation)
and from a true deploy failure (SSH, migration, health check, or core-loop exit
non-zero).

## Scope and non-goals

In scope:

1. Add `queue: max` to the `concurrency` block in `.github/workflows/deploy-staging.yml`
   (with `cancel-in-progress: false` unchanged) so up to GitHub's documented limit
   of pending staging deploys queue sequentially instead of cancelling superseded
   pending runs.
2. Add the same `queue: max` posture to `.github/workflows/deploy-production.yml`
   on group `production-deploy` for parity (`VOC-094-D04`).
3. Extend the operational-failure observer path so **benign**
   concurrency-superseded `cancelled` conclusions on `deploy-staging` and
   `deploy-production` do **not** open governed issues. Detection must use bounded
   GitHub API metadata only — run conclusion, job count, and/or the stable
   annotation substring — never job logs, step output, secrets, or session values
   in issue bodies (`VOC-094-D02`).
4. Preserve observer behavior for actionable conclusions: `failure`, `timed_out`, and
   `cancelled` runs that are **not** classified as concurrency supersession (manual
   cancel, in-progress cancel if ever enabled, deploy job started then cancelled,
   etc.).
5. Add deterministic tests locking queue wiring and benign-cancel classification.
6. Live verification (T01): after T00 merges, demonstrate that a controlled
   supersession scenario no longer leaves a spurious open
   `deploy-staging:cancelled` issue and that the latest commit's deploy still
   reaches `success`.

Non-goals / explicitly excluded:

- Setting `cancel-in-progress: true` on deploy workflows (would violate
  VOC-032-T07 fail-closed deploy safety documented in the workflow header).
- Weakening deploy steps, health checks, OAuth guards, core-loop gates, or SSH
  semantics.
- Modifying application code, migrations, Kuma inventory, or signup policy.
- Changing `scheduled-synthetics.yml` (VOC-090 owns staging synthetic timeouts).
- Removing the operational-failure observer or deduplication markers for real failures.
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R3** (`.github/workflows/`, `infra/scripts/`).
- **Measured path floor at drafting:** **R3**. Not R4 unless a task touches
  `scripts/governance/*`.
- Protected areas: `.github/workflows/deploy-staging.yml`,
  `.github/workflows/deploy-production.yml`,
  `.github/workflows/operational-failure-monitoring.yml`,
  `infra/scripts/open-failure-issue.sh`.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

## Decisions

`VOC-094-D00`: Run 32290409156 cancelled because GitHub Actions replaced a
**pending** run in concurrency group `staging-deploy` when a higher-priority
waiting request arrived. The remediation target is queue depth and observer
classification, not a defect in the deploy script itself.

`VOC-094-D01`: `cancel-in-progress: false` on deploy workflows remains mandatory.
The fix is **`queue: max`**, not enabling in-progress cancellation.

`VOC-094-D02`: Benign-cancel detection stays within VOC-088's sanitization boundary:
bounded workflow-run metadata via GitHub API or workflow_run event fields. No job
log ingestion, no step output, no secret values in issues or evidence.

`VOC-094-D03`: Classifier must remain **fail-closed toward opening issues** when
metadata is ambiguous. If the API is unavailable or classification is inconclusive,
the observer keeps today's behavior (open or deduplicate the issue) rather than
silently dropping a real cancellation.

`VOC-094-D04`: `deploy-production.yml` receives the same `queue: max` change so
rapid `main` promotions do not hit the identical single-pending-slot failure mode.
Production deploy semantics otherwise stay unchanged.

`VOC-094-D05`: `scheduled-synthetics:cancelled` and other non-deploy fingerprints
are out of scope unless T00 introduces a shared helper explicitly documented and
tested; default is deploy workflows only.

## Open questions for the reviewing human

1. Accept proposed **R3**, or raise in writing if the adopting human treats deploy
   concurrency/observer changes as R4 operational risk.
2. Confirm classifier signals: is **zero jobs + cancelled + deploy workflow** sufficient
   for T00, or should the implementer also require the stable annotation substring
   from the Checks/Jobs API when available?
3. If T01 cannot safely reproduce a triple-push supersession on the live staging
   host, is evidence of (a) a green subsequent `deploy-staging` run for the latest
   `develop` SHA plus (b) a deterministic fixture proving the classifier skip path
   acceptable for AC-05?

## Data, migrations, analytics, and accessibility

- No application schema migration.
- No database mutation.
- No product UI change — evidence-backed non-applicability.
- No analytics change — evidence-backed non-applicability.
- No accessibility change — evidence-backed non-applicability.
