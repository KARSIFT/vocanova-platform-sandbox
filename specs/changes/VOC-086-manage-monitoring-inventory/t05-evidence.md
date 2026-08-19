---
evidence_id: VOC-086-EV-05
task_id: VOC-086-T05
acceptance_criteria:
  - VOC-086-AC-05
  - VOC-086-AC-06
  - VOC-086-AC-07
  - VOC-086-AC-09
  - VOC-086-AC-10
tests:
  - VOC-086-TEST-10
  - VOC-086-TEST-12
  - VOC-086-TEST-13
  - VOC-086-TEST-17
date: 2026-08-19
related_change: VOC-086
accountable_owner: unassigned
gate_status: repository-complete-live-dispatch-blocked
live_socket_proof_claimed: false
live_synthetics_claimed: false
rollback_owner: revert repository inventory/workflow commits; re-run sync-monitoring with rolled-back inventory; explicit rotate_credentials only on compromise
last_known_good_sha: pre-VOC-086 tree (two unmanaged production Kuma monitors; deploy-only synthetics; no Kuma GitHub secrets) per issue #716
reviewed_sha: uncommitted-working-tree-bind-at-review
remediation_of: 4813868fda43b90d17bb46a588ae673e515a16e9
---

# VOC-086-T05 — Operator documentation and live verification proof

Attempt 2 remediates the High findings on `4813868fda43b90d17bb46a588ae673e515a16e9`
by (1) wiring read-only Socket.IO monitor-list proof into the normal
`sync-kuma-inventory.sh` path so a successful `sync-monitoring` run *is* TEST-10,
(2) refusing to label fixture tests as live TEST-10, and (3) actually attempting
`workflow_dispatch` of `sync-monitoring.yml` and `scheduled-synthetics.yml`.

Live apply and scheduled-synthetic greens **were not obtained**. The implementer
token cannot create `workflow_dispatch` events, and the GitHub `monitoring`
environment does not exist. Those are recorded as blockers, not as passes.

## Scope of this evidence

This task delivers operator documentation, read-only Socket.IO inventory proof
tooling wired into the normal sync host script, combined external verification,
deterministic harness coverage, DOC-11 uptime-monitoring amendment, and redacted
live external probes below.

**Socket.IO monitor-list closure (AC-05 / TEST-10 live)** and **scheduled-synthetic
green proof (AC-06 live)** are **not claimed**. `live_socket_proof_claimed` and
`live_synthetics_claimed` remain `false`. Harness tests fail closed if those
flags are set without Actions run URLs and PASS lines.

## Remediation (post independent FAIL on `4813868…`)

| Finding | Fix |
| --- | --- |
| High — live Kuma sync + Socket.IO inventory proof missing | Attempted `gh workflow run sync-monitoring.yml --ref develop`. Dispatch returned **HTTP 403** (`Resource not accessible by integration`; caller `pipeline.yml` grants `actions: read` only). GitHub environment `monitoring` is **absent** (HTTP 404), so `KUMA_*` secret names cannot exist. Wired `node prove-kuma-inventory.mjs` to run after `node sync-kuma.mjs` in `sync-kuma-inventory.sh` so the next authorized dispatch produces TEST-10 monitor-list metadata in the same log. |
| High — scheduled synthetics not proven green | Attempted `gh workflow run scheduled-synthetics.yml --ref develop`. Filename lookup 404s against default `main` (workflow is on `develop` only); API dispatch against the workflow ID also **403**. No run ID was created. |
| High — required work item 8 incomplete | Evidence records the attempted dispatch commands, the 403/404/environment-missing blockers, the latest healthy Sentry run URL, and external verifier output. It does **not** invent `sync-monitoring` or `scheduled-synthetics` success URLs. |
| Medium — evidence SHA not bound to reviewed head | `reviewed_sha` is `uncommitted-working-tree-bind-at-review`. Independent verification must bind to the exact commit SHA the implementer workflow creates from this working tree. |
| Medium — TEST-10 label overclaims | Fixture tests are named `VOC-086-T05 harness`. Live TEST-10/12 are evidence-gated: `live_*_claimed: true` requires Actions run URLs (and five `PASS:` monitor IDs for TEST-10). |

## Repository deliverables

| Artifact | Path |
| --- | --- |
| Operator documentation | `docs/operations/monitoring.md` |
| Operations index entry | `docs/operations/README.md` |
| DOC-11 uptime amendment | `docs/operations/11-devops-and-ci-cd.md` |
| External verification script | `infra/scripts/verify-voc086-monitoring.sh` |
| Disposable harness | `infra/scripts/verify-voc086-monitoring.selftest.sh` |
| Read-only Socket.IO proof (host) | `infra/scripts/prove-kuma-inventory.sh` |
| Proof logic (unit-testable) | `infra/monitoring/kuma-sync/prove-inventory.mjs` |
| Proof CLI | `infra/monitoring/prove-kuma-inventory.mjs` |
| Sync path now proves after apply | `infra/scripts/sync-kuma-inventory.sh` |
| Bundle includes prove script | `.github/workflows/sync-monitoring.yml` |
| Deterministic live harness | `scripts/foundation/voc086-live-verification.test.mjs` |
| Access-policy pointer | `infra/monitoring/access-policy.md` |
| Infra README pointers | `infra/README.md` |
| This evidence | `specs/changes/VOC-086-manage-monitoring-inventory/t05-evidence.md` |

## Acceptance mapping

| AC | Outcome this revision |
| --- | --- |
| AC-05 (live) | **Not met.** No authenticated Socket.IO monitor-list. Proof tooling is in the normal sync path; dispatch is blocked. |
| AC-06 (live) | **Not met.** No `scheduled-synthetics.yml` run. T03 wiring remains. |
| AC-07 | **Met (non-regression).** `error-monitoring.yml` unchanged by this task. Latest scheduled success on `main`: [run 32202272962](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32202272962) (`scanned=0, created=0`). |
| AC-09 (docs) | **Met.** `docs/operations/monitoring.md` covers add/change monitors, bootstrap/rotation, rollback, ownership, proof, governance. DOC-11 uptime row amended. |
| AC-10 | **Partial.** Monitor-host reachability, five public URLs, repository topology, and retired-bridge HTTP-2xx absence **PASS**. Live inventory/status and green synthetics **not met** (same blockers as AC-05/AC-06). |

## Live dispatch attempt (2026-08-19, implementer CI runner)

Caller: `pipeline.yml` `implement` job for PR #740. Token identity:
`github-actions[bot]`. Workflow permissions include `actions: read` (not write).

Commands:

```bash
gh workflow run sync-monitoring.yml --ref develop \
  -f rotate_credentials=true -f sync_inventory=true
# HTTP 403 Resource not accessible by integration

gh workflow run scheduled-synthetics.yml --ref develop
# HTTP 404 (workflow file not on default branch main), then API POST 403

gh workflow run error-monitoring.yml --ref develop
# HTTP 403 Resource not accessible by integration
```

| Check | Result |
| --- | --- |
| GitHub environment `monitoring` | **HTTP 404 — does not exist** |
| `KUMA_USERNAME` / `KUMA_PASSWORD` secret names | cannot exist without that environment |
| `sync-monitoring.yml` on `develop` | present (`c31220142a62`); **not on `main`** |
| `scheduled-synthetics.yml` on `develop` | present (`67b3025f48df`); **not on `main`** |
| Dispatched run IDs | **none** |

Pre-existing `sync-monitoring` runs are T02 push failures on task branches only
([31916545803](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31916545803),
[31916219840](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31916219840))
and are not live inventory apply.

## Operator closure steps (blocked for this implementer)

An identity with `actions: write` must, after creating the `monitoring`
environment:

```bash
gh workflow run sync-monitoring.yml --ref develop \
  -f rotate_credentials=true \
  -f sync_inventory=true

gh workflow run scheduled-synthetics.yml --ref develop
```

Then amend this evidence: set `live_socket_proof_claimed: true` and
`live_synthetics_claimed: true` only when the sync log contains `PASS:` for all
five `kuma.availability.*` IDs and the synthetics run is green for all five
stable IDs. Independent verification must re-bind that amended SHA.

## External verification (2026-08-19, implementer CI runner)

```bash
infra/scripts/verify-voc086-monitoring.sh --skip-socket-proof
```

```
VOC-086 monitoring verification — external + repository topology

VOC-081 monitor hostname and app-tier regression
PASS: staging web (https://staging.vocanova.site/) -> HTTP 200
PASS: staging api healthz (https://api-staging.vocanova.site/healthz) -> HTTP 200
PASS: production web (https://production.vocanova.site/) -> HTTP 200
PASS: production api healthz (https://api-production.vocanova.site/healthz) -> HTTP 200
PASS: monitor web (https://monitor.vocanova.site/ → final) -> HTTP 200
PASS: monitor body includes Uptime Kuma marker
PASS: unauthenticated SPA/entry-page boundary present (VOC-081-DEP-00; Kuma auth retained)
PASS: monitor socket.io polling handshake -> HTTP 200 (sid present; websocket upgrade advertised)

Canonical availability monitor URL probes
PASS: kuma.availability.staging.web -> HTTP 200
PASS: kuma.availability.staging.api-healthz -> HTTP 200
PASS: kuma.availability.production.web -> HTTP 200
PASS: kuma.availability.production.api-healthz -> HTTP 200
PASS: kuma.availability.monitor-host -> HTTP 200

Retired production bridge ports must not serve HTTP 2xx
PASS: production web :8081 -> HTTP 000 (not 2xx)
PASS: production api :8081 -> HTTP 000 (not 2xx)
PASS: production web :8443 -> HTTP 000 (not 2xx)
PASS: production api :8443 -> HTTP 400 (not 2xx)

Repository topology invariants
PASS: voc081-monitoring-topology repository assertions

Skipping Socket.IO inventory proof (no KUMA_USERNAME/KUMA_PASSWORD in environment)

All VOC-086 monitoring verification checks passed.
```

Retired-bridge probes (not HTTP 2xx):

| Probe | Result |
| --- | --- |
| `https://production.vocanova.site:8081/` | timeout, `http_code=000` |
| `http://production.vocanova.site:8081/` | timeout, `http_code=000` |
| `https://production.vocanova.site:8443/` | timeout, `http_code=000` |
| `http://production.vocanova.site:8443/` | HTTP 400 (TCP open; not HTTP 2xx) |

TCP 8443 is open on the public path with plaintext HTTP 400. That is **not** a
successful app response and is **not** a Cloudflare change (out of T05 scope).
Repository compose still does not publish `8081`/`8443`.

## Isolation and topology (TEST-17)

Repository assertions (`voc081-monitoring-topology.test.mjs`) confirm:

- single shared-edge nginx publishes host `80`/`443` only
- Kuma loopback-only `127.0.0.1:3001` on `vocanova-monitoring-net`
- no `8081`/`8443` publish in staging/production/monitoring compose
- staging deploy owns monitoring convergence; production deploy does not

External `:443` checks (no `:8443` in URLs) pass via `verify-voc067-cutover.sh`
inside the VOC-081 verifier.

## Deterministic validation

Commands (repo root):

```bash
bash infra/scripts/verify-voc086-monitoring.selftest.sh
node --test scripts/foundation/voc086-live-verification.test.mjs
node --test scripts/foundation/voc086-*.test.mjs
bash infra/scripts/sync-kuma-inventory.selftest.sh
bash scripts/governance/validate-governance.sh
```

Results on this working tree (bind to the committed SHA at review):

| Command | Result |
| --- | --- |
| `verify-voc086-monitoring.selftest.sh` | pass |
| `voc086-live-verification.test.mjs` | pass (9 tests) |
| `voc086-sync-monitoring-workflow.test.mjs` | pass (8 tests) |
| `voc086-kuma-sync.test.mjs` | pass (11 tests) |
| `sync-kuma-inventory.selftest.sh` | pass |
| `verify-voc086-monitoring.sh --skip-socket-proof` | pass |
| `validate-governance.sh` | pass |

## Rollback

1. Revert the responsible VOC-086 commit(s).
2. Re-run `sync-monitoring.yml` with `sync_inventory=true` from the rolled-back
   inventory (`rotate_credentials=false` unless compromise is suspected).
3. Re-run `verify-voc086-monitoring.sh` and confirm the sync log's
   `prove-kuma-inventory.mjs` PASS lines.
4. Preserve unrelated manual Kuma monitors.

## Secrets

None committed. No Kuma password, mint token, session cookie, or CSRF value
appears in this evidence.

## Follow-up notes (out of scope for T05 code)

- Grant `actions: write` to the implement caller, or dispatch from an operator
  / App token that already has it. Do not treat this implementer 403 as a
  green live proof.
- Create GitHub environment `monitoring` (T02 operator prerequisite) before
  the first `rotate_credentials=true` run.
- Promote `scheduled-synthetics.yml` / `sync-monitoring.yml` to `main` via the
  normal release path so `gh workflow run <filename>` resolves on the default
  branch.
- Package README front matter still says `draft` — housekeeping, not required
  for operator docs.
