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
gate_status: live-complete-awaiting-independent-review
live_socket_proof_claimed: true
live_synthetics_claimed: true
live_workflow_sha: f59c6a3f8fb2e8e886a32c343d79775276c9800d
reviewed_sha: bind-at-independent-review
rollback_owner: revert repository inventory/workflow commits; re-run sync-monitoring with the rolled-back inventory; rotate credentials only on compromise
last_known_good_sha: f06061628930fc9b9164369005f05430a4c353de
remediation_of: 4813868fda43b90d17bb46a588ae673e515a16e9
---

# VOC-086-T05 — Live monitoring verification

The repository inventory is applied to live Kuma through its supported
Socket.IO protocol. Five canonical availability monitors are active, a second
reconciliation is a zero-change no-op, every canonical synthetic ID has a green
live execution, Sentry monitoring is green, and the expected shared-edge and
network isolation remain intact.

No SQLite file was used as a deployment mechanism. No credential, session,
mint token, or CSRF value is present in this evidence.

## Live workflow evidence

| Purpose | Revision | Result | Run |
| --- | --- | --- | --- |
| Production release containing scheduled workflows and VOC-087 safety corrections | `f06061628930fc9b9164369005f05430a4c353de` | success | [deploy-production 32205116790](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32205116790) |
| Kuma inventory apply and read-only proof | `53e9401f0fe23b54064528cb64c4f94148ef5930` | success; five monitors | [sync-monitoring 32206113488](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32206113488) |
| Kuma idempotence and repeated proof | `53e9401f0fe23b54064528cb64c4f94148ef5930` | success; `0 monitor change(s)` | [sync-monitoring 32206161250](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32206161250) |
| Four production/OAuth synthetics on protected `main` | `f06061628930fc9b9164369005f05430a4c353de` | four jobs success; staging job found the deterministic-state gap described below | [scheduled-synthetics 32206210907](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32206210907) |
| Corrected staging core-loop synthetic | `f59c6a3f8fb2e8e886a32c343d79775276c9800d` | success | [scheduled-synthetics 32206440090](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32206440090) |
| Sentry issue monitor | `f06061628930fc9b9164369005f05430a4c353de` | success | [error-monitoring 32206700297](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32206700297) |

The production environment permits `main`, not task branches. Therefore the
pre-merge evidence uses the four green protected-production jobs from the
`main` run and the corrected green staging job from the task revision. After
normal promotion, the operator reruns the full suite once on `main` and records
that combined run on the root issue before closure.

### Authenticated Socket.IO inventory proof

The first successful apply log contains:

```text
PASS: kuma.availability.staging.web (kuma_id=3, name=Staging Web, url=https://staging.vocanova.site/, interval=60, timeout=30, retries=2, statuscodes=200, repo_managed=true, active=true)
PASS: kuma.availability.staging.api-healthz (kuma_id=4, name=Staging API /healthz, url=https://api-staging.vocanova.site/healthz, interval=60, timeout=30, retries=2, statuscodes=200, repo_managed=true, active=true)
PASS: kuma.availability.production.web (kuma_id=2, name=VocaNova Production Web, url=https://production.vocanova.site, interval=60, timeout=30, retries=2, statuscodes=200, repo_managed=true, active=true)
PASS: kuma.availability.production.api-healthz (kuma_id=1, name=VocaNova Production API, url=https://api-production.vocanova.site/healthz, interval=60, timeout=30, retries=2, statuscodes=200, repo_managed=true, active=true)
PASS: kuma.availability.monitor-host (kuma_id=5, name=Monitor host reachability, url=https://monitor.vocanova.site/, interval=60, timeout=30, retries=2, statuscodes=200, repo_managed=true, active=true)
All canonical availability monitors match inventory.
```

The second run contains the same five `PASS:` lines plus:

```text
Kuma inventory already matches repository inventory (no-op).
Kuma sync completed (0 monitor change(s)).
Managed monitors in Kuma: 5
All canonical availability monitors match inventory.
```

The two original production monitors retained Kuma IDs 1 and 2 and were
adopted in place. Staging web/API and monitor-host were created as IDs 3–5.
Unrelated manually owned monitors are excluded from repository mutation by the
synchronizer's ownership/collision rules.

### Scheduled synthetic proof

The following canonical IDs have successful live jobs:

| Stable ID | Live result |
| --- | --- |
| `synthetic.staging.oauth-expected-state` | success in run 32206210907 |
| `synthetic.production.oauth-expected-state` | success in run 32206210907 |
| `synthetic.production.journey-content` | success in run 32206210907 |
| `synthetic.production.authenticated-route-content-sweep` | success in run 32206210907 |
| `synthetic.staging.authenticated-core-journey` | success in run 32206440090 |

The production route sweep is GET-only and uses the reserved production
synthetic account. The staging core journey is explicitly mutating and uses
only the reserved `.invalid` staging account.

## Live-found failures and repository remediations

| Finding | Repository correction and evidence |
| --- | --- |
| GitHub App lacks `Environments: write` | Added explicit `preprovisioned_credentials` recovery mode. The `monitoring` environment and both named secrets were provisioned outside the server without logging values. Normal runs only read them. |
| Kuma reset utility catches internal errors and exits 0 | The wrapper now requires Kuma's exact reset-success and immediate authenticated-login markers before writing reset proof. Deterministic fixture covers a caught error with exit 0. |
| Blindly piping both reset answers caused `Error: readline was closed` | Replaced preload/EOF behavior with a prompt-driven, non-echoing coprocess exchange over `docker exec -i`. Run 32205903380 proved reset and login success. |
| Disposable Node container left root-owned npm files | Run it as the deploy user's UID/GID with `HOME=/tmp`; cleanup is now deterministic. |
| Kuma 1.23 rejected newer `conditions` payload field | Removed the unsupported add/edit/rollback field and added a protocol regression test. |
| Normal sync reused a stale extracted bundle | Every sync now removes and extracts the exact uploaded reviewed bundle before execution. |
| Repeated staging browser synthetic had no due review cards | Before the journey, the workflow reruns the same idempotent deployment seed over SSH. It refreshes only the reserved synthetic account and marks one saved word due. Run 32206440090 passed. |

Failed bootstrap/sync attempts were fail-closed:

- [32205157088](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32205157088): App permission rejected before SSH/reset.
- [32205377420](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32205377420): exposed the false reset result; credential material was scrubbed only after the preprovisioned pair had been confirmed.
- [32205730469](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32205730469): marker guard rejected `readline was closed`; no reset proof was created and unused copies were removed.
- [32205903380](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32205903380): corrected reset/login succeeded; inventory then failed before mutation on the unsupported schema field.
- [32206036881](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32206036881): stale extracted code was detected; the first create failed and compensation left the pre-existing inventory intact.

Read-only verification after recovery confirmed that no
`/tmp/kuma-{new-password,reset-applied,rotate-metadata}*` material remains.

## External verification through Cloudflare

Command:

```bash
infra/scripts/verify-voc086-monitoring.sh --skip-socket-proof
```

The Socket.IO proof is skipped only in this local command because credentials
are not exported locally; runs 32206113488 and 32206161250 perform the
authenticated proof inside the workflow.

```text
PASS: staging web -> HTTP 200
PASS: staging api healthz -> HTTP 200
PASS: production web -> HTTP 200
PASS: production api healthz -> HTTP 200
PASS: monitor web -> HTTP 200
PASS: monitor body includes Uptime Kuma marker
PASS: monitor socket.io polling handshake -> HTTP 200
PASS: kuma.availability.staging.web -> HTTP 200
PASS: kuma.availability.staging.api-healthz -> HTTP 200
PASS: kuma.availability.production.web -> HTTP 200
PASS: kuma.availability.production.api-healthz -> HTTP 200
PASS: kuma.availability.monitor-host -> HTTP 200
PASS: production web :8081 -> HTTP 000 (not 2xx)
PASS: production api :8081 -> HTTP 000 (not 2xx)
PASS: production web :8443 -> HTTP 000 (not 2xx)
PASS: production api :8443 -> HTTP 400 (not 2xx)
PASS: voc081-monitoring-topology repository assertions
All VOC-086 monitoring verification checks passed.
```

## Live topology and isolation

Read-only host inspection on 2026-08-19 showed:

- staging and production Postgres/API/web containers healthy;
- `vocanova-uptime-kuma` healthy and published only on
  `127.0.0.1:3001`;
- exactly one nginx container, `vocanova-shared-edge-nginx`, owning host
  `80/443`;
- shared edge attached to `vocanova-net`, `vocanova-production-net`, and
  `vocanova-monitoring-net`;
- Kuma attached only to `vocanova-monitoring-net`;
- no host listeners on `8081` or `8443`;
- no retained credential recovery files.

## Acceptance mapping

The operator procedure and ownership model are maintained in
`docs/operations/monitoring.md`.

| Acceptance criterion | Result |
| --- | --- |
| AC-05 / TEST-10 | Met: apply plus two authenticated Socket.IO proofs; second run is a no-op. |
| AC-06 / TEST-12 | Met: all five stable synthetic IDs have green live executions with masked session/token handling. |
| AC-07 / TEST-13 | Met: Sentry remains separate and run 32206700297 is green. |
| AC-09 | Met: operator docs cover ownership, add/update flow, bootstrap, rollback, and feature-monitor mappings. |
| AC-10 / TEST-17 | Met: Cloudflare-facing URLs, monitor host, inventory proof, synthetics, topology, and retired listeners verified. |

## Deterministic validation

Run from repository root:

```bash
pnpm test
pnpm run format:check
bash infra/scripts/verify-voc086-monitoring.selftest.sh
bash infra/scripts/sync-kuma-inventory.selftest.sh
bash scripts/governance/validate-governance.sh
```

The final independent review must bind to the committed evidence SHA. CI and
governance results are recorded on PR #740.

## Rollback

1. Revert the responsible VOC-086 commits through the normal PR path.
2. Run `sync-monitoring.yml` with `rotate_credentials=false` and
   `sync_inventory=true` from the rolled-back inventory.
3. Confirm the authenticated `prove-kuma-inventory` output and preserve any
   unrelated manually owned monitors.
4. Run `scheduled-synthetics.yml` and
   `verify-voc086-monitoring.sh --skip-socket-proof`.
5. Rotate Kuma credentials only when compromise is suspected; an inventory
   rollback does not require rotation.
