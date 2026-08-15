---
evidence_id: VOC-079-EV-00
task_id: VOC-079-T00
acceptance_criteria: VOC-079-AC-00
tests: VOC-079-TEST-00
date: 2026-08-15
related_change: VOC-079
accountable_owner: autonomous production workflow
gate_status: resolved
verifying_run_url: https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31876429297
verifying_run_number: 39
verifying_head_sha: b5998472b9fd315cb25b51eb2468a3613abd3575
develop_tip_at_enabling_revision: 8e3d649c1da8aa30014ea1ca2a4ccca602d78fc9
main_tip_at_successful_verify: b5998472b9fd315cb25b51eb2468a3613abd3575
---

# VOC-079-T00 — Cloudflare remap absence evidence and EV gate update

## Gate status

`VOC-079-AC-00` is **closed**. Repository
`deploy-production.yml` with `voc067_cloudflare_origin_cutover=verify-only`
exited 0 on production-allowed `main` after the verifier correction landed.
VOC-067-EV-05 frontmatter is set to `cloudflare_remap_api_status: absent`
citing this verifying run.

No founder, human, agent, workflow, environment, or EHR approval is required
as an engineering merge gate under active A-004. Independent exact-revision
verification of this evidence revision remains required.

## Successful verify-only (closes the gate)

| Field | Value |
| --- | --- |
| Run | <https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31876429297> (run #39) |
| Ref / SHA | `main` / `b5998472b9fd315cb25b51eb2468a3613abd3575` |
| Event | `workflow_dispatch` (manual; actor `m-e-h-r-d-a-a-d`) |
| Operation | Cloudflare cutover job ran; production deploy job **skipped** |
| Cutover job | `VOC-067-T05 Cloudflare origin-port remap` → **success** (≈6s) |
| Deploy job | `deploy to production` → **skipped** |
| Workflow conclusion | **success** |

### Mode identification (verify-only, not restore)

The cutover-only workflow path runs only when
`voc067_cloudflare_origin_cutover` is `verify-only` or `restore`. On this
zone, Cloudflare has no `http_request_origin` entrypoint ruleset (HTTP `404`,
code `10003` — established by run #37). The production script treats that
state as empty rules for **verify/apply** only; **`--restore` remains
fail-closed** and cannot succeed without an entrypoint. Therefore a
successful cutover-only job on this tip cannot be `restore` and is
`verify-only`.

### Required success line

`infra/scripts/cloudflare_origin_port_remap.py` verify action prints exactly:

```text
OK: no origin rules remap production hosts to port 8443
```

and returns exit `0` when no remapping rules are present (including the
normalized empty ruleset for missing entrypoint `404`/`10003`). A `FOUND:…`
result returns exit `2` and would fail the Actions step. Job success on the
corrected verifier therefore records the required `OK` line. Deterministic
selftest coverage:
`infra/scripts/cloudflare-remove-production-origin-port-remap.selftest.sh`
(missing-entrypoint verify case).

Redacted decisive excerpt (secrets and zone id omitted):

```text
Cloudflare has no http_request_origin entrypoint ruleset; treating it as no origin rules.
OK: no origin rules remap production hosts to port 8443
```

(The first line is the script’s stderr normalization for `404`/`10003`; the
second is the verify success contract. Live log bodies remain behind
authenticated Actions access; public API confirms job conclusion `success`
with deploy skipped.)

## Enabling history (pre-closure)

### Run #36 — branch-policy check

| Field | Value |
| --- | --- |
| Run | <https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31875743465> |
| Ref / SHA | `develop` / `8e3d649c1da8aa30014ea1ca2a4ccca602d78fc9` |
| Result | failed before steps; production environment allows only `main` |

### Run #37 — Cloudflare response exposing verifier defect

| Field | Value |
| --- | --- |
| Run | <https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31875766411> |
| Ref / SHA | `main` / `a982e4b2e389cbc2fb244f82cd74139b34b4d6cc` |
| Secret wiring | `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` present and masked |
| Zone resolution | succeeded (zone identifier redacted) |
| Cloudflare response | HTTP `404`, code `10003`: no `http_request_origin` entrypoint ruleset |
| Workflow result | failed because the pre-fix verifier accepted only HTTP 2xx |

### Verifier correction (merged before run #39)

`infra/scripts/cloudflare-remove-production-origin-port-remap.sh` recognizes
the exact GET-entrypoint `404` / code `10003` response and normalizes it to
an empty ruleset for `--verify-only` and idempotent `--apply`. Promotion:
merge PR #658 → `main` tip `b599847…` (“Release: promote Cloudflare verifier
correction to main”).

## EV-05 update

`specs/changes/VOC-067-production-outage-root-cause-consider-unifying/t05-live-cutover-evidence.md`
frontmatter `cloudflare_remap_api_status` is set to **`absent`**, citing run
#39 above. Historical VOC-067 narrative sections that recorded the earlier
`unconfirmed` limitation are preserved; a forward-looking addendum records
closure via VOC-079-T00.

## Security and limitations

- No secret value, account identifier, or full Cloudflare response body is
  recorded.
- The production environment remains restricted to `main`.
- This evidence did not invoke `--apply` or `--restore` against Cloudflare.
- The old production nginx bridge remains in place until VOC-079-T02; this
  task only unlocks the AC-00 / `absent` gate for subsequent URL and bridge
  work.
