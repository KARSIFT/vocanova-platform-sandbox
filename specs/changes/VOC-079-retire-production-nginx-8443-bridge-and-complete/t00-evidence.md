---
evidence_id: VOC-079-EV-00
task_id: VOC-079-T00
acceptance_criteria: VOC-079-AC-00
tests: VOC-079-TEST-00
date: 2026-08-15
related_change: VOC-079
accountable_owner: autonomous production workflow
gate_status: pending-verifier-fix-release-and-successful-rerun
develop_tip_at_execution: 8e3d649c1da8aa30014ea1ca2a4ccca602d78fc9
main_tip_at_execution: a982e4b2e389cbc2fb244f82cd74139b34b4d6cc
---

# VOC-079-T00 — Cloudflare remap absence evidence and verifier correction

## Current gate status

`VOC-079-AC-00` is not yet closed. Authenticated repository `verify-only`
execution reached the correct Cloudflare zone and Cloudflare reported that the
`http_request_origin` entrypoint ruleset does not exist. Cloudflare represents that
valid “no rules in this phase” state as HTTP `404`, error code `10003`; the existing
repository script incorrectly treated it as an API failure.

This revision corrects the verifier and adds deterministic regression coverage. It
is an enabling revision and intentionally does **not** close issue #652. After the
fix reaches the production-allowed `main` branch, a follow-up exact revision must
record a successful `verify-only` rerun, set this gate to `resolved`, and update
VOC-067-EV-05 to `cloudflare_remap_api_status: absent` before bridge retirement.

No founder, human, agent, workflow, environment, or EHR approval is required.
Independent exact-revision verification and deterministic checks remain required.

## Authenticated execution evidence

### Run #36 — branch-policy check

| Field | Value |
| --- | --- |
| Run | <https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31875743465> |
| Ref / SHA | `develop` / `8e3d649c1da8aa30014ea1ca2a4ccca602d78fc9` |
| Operation | `voc067_cloudflare_origin_cutover=verify-only` |
| Result | failed before steps; the production environment allows only `main` |

The environment policy was preserved. It was not weakened to execute this task.

### Run #37 — Cloudflare response exposing verifier defect

| Field | Value |
| --- | --- |
| Run | <https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31875766411> |
| Ref / SHA | `main` / `a982e4b2e389cbc2fb244f82cd74139b34b4d6cc` |
| Operation | `voc067_cloudflare_origin_cutover=verify-only` |
| Secret wiring | `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` present and masked |
| Zone resolution | succeeded (zone identifier redacted from this evidence) |
| Cloudflare response | HTTP `404`, code `10003`: no `http_request_origin` entrypoint ruleset |
| Workflow result | failed because the old verifier accepted only HTTP 2xx |

Redacted decisive excerpt:

```text
ERROR: Cloudflare API GET /zones/[redacted]/rulesets/phases/
http_request_origin/entrypoint returned HTTP 404
code: 10003
message: could not find entrypoint ruleset in the http_request_origin phase
```

This is not an unknown Cloudflare state and not evidence of an `:8443` remap. There
is no phase entrypoint in which such an Origin Rule could exist. The repository gate
still remains pending until its own corrected command exits successfully.

## Verifier correction in this revision

`infra/scripts/cloudflare-remove-production-origin-port-remap.sh` now:

1. recognizes only the exact GET-entrypoint `404` / code `10003` response;
2. normalizes it to an empty ruleset for `--verify-only` and idempotent `--apply`;
3. keeps `--restore` fail-closed because creating a missing entrypoint is a distinct
   mutation from updating an existing ruleset; and
4. resolves `ZONE_ID` outside command substitution so live apply/restore calls do
   not lose it in a subshell under `set -u`.

`infra/scripts/cloudflare-remove-production-origin-port-remap.selftest.sh` adds a
fake-API regression for missing-entrypoint verify and apply behavior.

Local deterministic results on this exact working revision:

```text
All cloudflare cutover selftests passed.
Repository foundation validation passed.
Governance structure validation passed.
Ran 95 tests ... OK
git diff --check: pass
```

## Required follow-up before T01

1. Merge this enabling revision to `develop` without closing issue #652.
2. Promote the verifier correction to `main` through checked branch promotion.
3. Dispatch `deploy-production.yml` on `main` with
   `voc067_cloudflare_origin_cutover=verify-only`.
4. Require workflow success and the exact status line:

   ```text
   OK: no origin rules remap production hosts to port 8443
   ```

5. Commit a follow-up evidence revision that records the successful run and updates
   VOC-067-EV-05 from `unconfirmed` to `absent`.
6. Close #652 only after the independently reviewed follow-up merges. T01 remains
   dependency-blocked until then.

## Security and limitations

- No secret value, account identifier, or full Cloudflare response body is recorded.
- The production environment remains restricted to `main`.
- This task has not invoked `--apply` or `--restore` against Cloudflare.
- The old production nginx bridge remains in place; no server mutation is authorized
  by this enabling revision.
