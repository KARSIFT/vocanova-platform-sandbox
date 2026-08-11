---
evidence_id: VOC-067-EV-04
task_id: VOC-067-T04
acceptance_criteria: VOC-067-AC-05
tests: VOC-067-TEST-05
date: 2026-08-11
related_change: VOC-067
status: blocked
---

# VOC-067-T04 — Port-normalization evidence (blocked; attempt-2 remediation)

## Verdict

**T04 cannot be completed in this run.** Founder-accepted cutover order in
`t00-edge-architecture-decision-record.md` (VOC-067-DEP-03 steps 146–159)
requires:

1. T02 shared edge live  
2. T03 deploy isolation  
3. Origin `:443` verification  
4. **T05** — Cloudflare origin-port remap removal + `VOC-067-EV-05`  
5. **T04** — remove `:8443` URL/doc workarounds **only after step 4 succeeds**

There is no `VOC-067-EV-05` / `t05-*` artifact. Live Cloudflare remap removal
and external `:443` proof are founder/ops-held (T05). This implementer role
has no authority to execute Cloudflare API cutover or invent that evidence.

Shipping T04-shaped URL normalization and retiring the production `:8443`
bridge while Cloudflare still remaps to origin `:8443` recreates the #485
outage class (edge points at a listener that no longer exists). Attempt 1
(`f0403d4`) did exactly that and correctly failed independent verification.

## Remediation of attempt-1 Critical/High findings

Working-tree restoration of the T03 cutover-safe repository shape (same as
`develop` / post-T03):

| Finding | Fix in this revision |
| --- | --- |
| Critical — T04 ahead of T00/DEP-03; no EV-05 | No `:8443` URL/doc normalization landed; task marked **blocked** pending T05 |
| Critical — retired `vocanova-production-nginx` `:8443` bridge before T05 | Bridge service + deploy `nginx -t`/reload restored |
| High — docs claimed remap-free steady state as current fact | `infra/README.md` / compose headers restored to cutover dual-publish wording (remap still required until T05) |
| Medium — live `:443` / OAuth checks deferred | Acknowledged; those checks belong to EV-05 then a future T04 re-dispatch |

Foundation guards remain the VOC-041/VOC-042 cutover expectations (`:8443`
required) until a post-T05 T04 re-run flips them under `VOC-067-TEST-05`.

## What T04 will do after T05 succeeds

When `VOC-067-EV-05` records remap removal and external `:443` success,
re-dispatch T04 to:

- Set `API_BASE_URL` / deploy-emitted `BASE_URL` / OAuth / health / smoke URLs
  to ordinary hostnames (no `:8443`)
- Retire the temporary production bridge **only if** T05’s cutover evidence
  authorizes bridge removal (or coordinate retirement in the same release
  train as documented in EV-05)
- Update operator docs so remap is no longer described as current ops
- Flip foundation guards to `VOC-067-TEST-05` “must not emit `:8443`”
- Record live curl / OAuth spot-checks in an updated EV-04

## Package graph note (follow-up; out of T04 scope)

`.karsift/tasks.json` currently has `T05 depends_on T04`, which contradicts
the founder-accepted T00 order (T05 before T04). Do not “fix” that by
landing T04 early. A planner/ops follow-up should reconcile the task graph
with DEP-03 before the next dispatch.

## Rollback / last-known-good

Public production path during cutover remains: Cloudflare remap → origin
`:8443` via `vocanova-production-nginx`, with shared edge on `80`/`443` for
Host routing proof. That is the T02/T03 last-known-good until T05 completes.
