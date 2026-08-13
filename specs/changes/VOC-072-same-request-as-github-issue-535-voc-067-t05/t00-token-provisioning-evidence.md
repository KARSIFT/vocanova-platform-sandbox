---
evidence_id: VOC-072-EV-00
task_id: VOC-072-T00
acceptance_criteria: VOC-072-AC-00
tests: VOC-072-TEST-00
date: 2026-08-13
related_change: VOC-072
accountable_owner: m-e-h-r-d-a-a-d (founder/ops; VOC-067-DEP-03)
gate_status: resolved
voc072_dep_00: dedicated_secret
voc072_dep_01_secret_name: PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN
---

# VOC-072-T00 — Cloudflare zone/Origin-Rules token provisioning evidence

Evidence for `VOC-072-T00`, `VOC-072-AC-00`, and `VOC-072-TEST-00`.

History: an earlier revision claimed AC-00 from a **repository**-scoped
secret (rejected on review); the next revision correctly withdrew that claim
and left AC-00 open pending an environment-scoped secret. A later revision
(`f49ffc50`) recorded founder provisioning into the GitHub **production**
environment. Commit `4b021050` regressed this file to
`gate_status: pending_operator_execution`; `VOC-077-T00` restores the
confirmed provisioned state below.

## Gate status: RESOLVED — production-environment secret confirmed present

**`VOC-072-AC-00` is satisfied. `VOC-072-TEST-00` passes.**

Founder moved the secret into the GitHub **production** environment and
confirmed via:

```bash
gh secret list --env production --repo KARSIFT/vocanova-platform-sandbox \
  | grep -E 'PRODUCTION_CLOUDFLARE'
```

Redacted excerpt (name + updated-at only; **never** the token value):

```
PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN    2026-08-13T21:24:26Z
PRODUCTION_CLOUDFLARE_API_TOKEN                    (present, unchanged)
PRODUCTION_CLOUDFLARE_ACCOUNT_ID                 (present, unchanged)
```

This is exactly TEST-00 step 1's required listing (`--env production`, not a
bare repository listing). Initial provisioning confirmed 2026-08-13;
`VOC-077-T00` (2026-08-13) restored this record after an implementer
regression overwrote it with a stale pending template.

| Requirement | Status |
| --- | --- |
| `VOC-072-DEP-00` resolved (dedicated vs reuse) | **Resolved** — dedicated secret (§1) |
| `VOC-072-DEP-01` resolved (secret name + location) | **Resolved** — `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` in production environment (§1, §5) |
| Cloudflare API token created with correct scopes | **Confirmed by founder** — Zone Read + Origin Rules Edit on `vocanova.site` (§5) |
| GitHub production environment secret set | **Confirmed** — present per redacted listing above |
| Redacted audit record (no secret values in git) | **This file** — §5 complete |

| AC-00 element | Standing after this evidence |
| --- | --- |
| Production GitHub environment holds a zone-capable secret | **SATISFIED** — `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` present in the `production` environment per the `--env production` listing above |
| Token scoped to Zone Read + Origin Rules Edit for `vocanova.site` | **CONFIRMED BY FOUNDER** against the §3 procedure — not independently re-readable from outside the Cloudflare dashboard; taken on direct founder confirmation, consistent with human-production-authority precedent elsewhere in this package |
| Evidence documents token label, scopes, secret name (no values) | **PARTIAL** — secret name, location, and permission summary recorded; exact Cloudflare dashboard display name remains unconfirmed (Low finding, non-blocking per prior review) |
| If reuse path: Workers AI sync re-verified | **N/A** — dedicated-secret path chosen (§1). `PRODUCTION_CLOUDFLARE_API_TOKEN` and `PRODUCTION_CLOUDFLARE_ACCOUNT_ID` confirmed still present and unmodified |

`VOC-072-T01` may wire the name from DEP-01 against the real secret.
`VOC-072-T02` `--verify-only` may now proceed.

## 1. Adoption decisions (`VOC-072-DEP-00` / `VOC-072-DEP-01`)

Recorded at T00 per `tasks.md` (DEP decisions are not resolved in
`change.yaml`'s drafting-time dependency block; T00 is the authoritative
record for T01 wiring).

Adoption accepted this package as proposed with the **recommended
least-privilege path** from `specification.md` open questions 1–2. Plan PR
**#537**; adopt PR **#541** (2026-08-12, founder-delegate under standing
founder-gate delegation and the founder's 2026-08-12 trip authorization for
infra fix work).

| Decision | Resolution |
| --- | --- |
| **VOC-072-DEP-00** | **Dedicated secret** for zone read + Origin Rules edit on `vocanova.site`. Do **not** broaden `PRODUCTION_CLOUDFLARE_API_TOKEN` (Workers AI sync keeps its existing narrow scope). **Resolved.** |
| **VOC-072-DEP-01** | Secret **name** **`PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN`**, required **location** the GitHub **production** environment on `KARSIFT/vocanova-platform-sandbox`. Both name and environment-scoped presence are confirmed (§5). |

**Reuse path rejected** at adoption: widening `PRODUCTION_CLOUDFLARE_API_TOKEN`
would increase blast radius if that token leaks and couples Workers AI sync to
zone/Origin-Rules authority.

### `VOC-072-DEP-00` — Dedicated secret (chosen)

**Decision:** Add a **dedicated** production GitHub secret for Cloudflare zone
read and Origin Rules edit. Do **not** broaden `PRODUCTION_CLOUDFLARE_API_TOKEN`.

**Rationale:**

- Issue #535 and `VOC-037-EV-06` confirm the existing
  `PRODUCTION_CLOUDFLARE_API_TOKEN` is scoped for **Workers AI provider sync**
  (`deploy-production.yml` writes `AI_PROVIDER_API_KEY` from that secret).
- `infra/scripts/cloudflare-remove-production-origin-port-remap.sh` calls
  `GET /zones?name=vocanova.site` before reading or mutating the
  `http_request_origin` ruleset. The Workers-AI-scoped token returns an empty
  `result` array → `ERROR: zone not found` in CI (fail-closed per
  `VOC-067-TEST-06`).
- Least privilege: cutover tooling needs Zone Read + Origin Rules Edit on
  `vocanova.site` only; Workers AI sync should keep its narrow scope
  (`VOC-072-R00` / `VOC-072-R01`).

**Rejected alternative:** Reuse/broaden `PRODUCTION_CLOUDFLARE_API_TOKEN` —
higher blast radius if leaked; requires re-verifying Workers AI sync after
scope change; not chosen.

### `VOC-072-DEP-01` — Secret name

**Decision:** `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN`

`VOC-072-T01` must bind `voc067-cloudflare-cutover` and the post-smoke
`--apply` step to this secret. Workers AI sync continues using
`PRODUCTION_CLOUDFLARE_API_TOKEN` and `PRODUCTION_CLOUDFLARE_ACCOUNT_ID`
unchanged.

## 2. Repository context (verified read-only)

| Artifact | Current binding | Notes |
| --- | --- | --- |
| `deploy-production.yml` job `voc067-cloudflare-cutover` | `PRODUCTION_CLOUDFLARE_API_TOKEN` | Fails zone lookup with Workers-AI token (issue #535) |
| `deploy-production.yml` step `VOC-067-T05 … (apply after healthy deploy)` | `PRODUCTION_CLOUDFLARE_API_TOKEN` | Same — T01 retargets to DEP-01 secret |
| `deploy-production.yml` AI provider sync | `PRODUCTION_CLOUDFLARE_API_TOKEN` + `PRODUCTION_CLOUDFLARE_ACCOUNT_ID` | **Unchanged** on dedicated path |
| `cloudflare-remove-production-origin-port-remap.sh` | `CLOUDFLARE_API_TOKEN` or `PRODUCTION_CLOUDFLARE_API_TOKEN` | T01 adds DEP-01 env precedence |

**Minimum Cloudflare API permissions** for the new token:

| Permission | Purpose |
| --- | --- |
| Zone → Read | `GET /zones?name=vocanova.site` (`resolve_zone_id`) |
| Zone → Origin Rules → Edit (or Account → Origin Rules → Edit scoped to zone) | `GET` / `PUT` `/zones/{id}/rulesets/phases/http_request_origin/entrypoint` and ruleset updates |

**Zone resources:** Include → Specific zone → `vocanova.site` only.

## 3. Operator runbook — Cloudflare dashboard

Accountable owner: `m-e-h-r-d-a-a-d` (or delegate with dashboard access).

1. Sign in to Cloudflare → **My Profile** → **API Tokens** → **Create Token**.
2. Use **Create Custom Token** (not a template — templates may omit Origin Rules).
3. **Token name** (suggested): `vocanova-production-origin-rules-cutover`
   — record the exact display name in §5.
4. **Permissions** (minimum):
   - Zone → Zone → Read
   - Zone → Origin Rules → Edit  
     (If the dashboard offers Account-level Origin Rules Edit instead, restrict
     zone resources to `vocanova.site` only.)
5. **Zone Resources:** Include → Specific zone → `vocanova.site`.
6. **Account Resources:** None required unless the dashboard forces an account
   scope for Origin Rules — if so, limit to the account holding `vocanova.site`.
7. Create the token. Copy the value **once** — it will not be shown again.
   **Do not paste the token into git, PR comments, or this evidence file.**

## 4. Operator runbook — GitHub production environment

Repository: `KARSIFT/vocanova-platform-sandbox`

1. **Settings** → **Environments** → **production** → **Environment secrets**.
2. **New secret**
   - Name: `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` (per DEP-01)
   - Value: the Cloudflare token from §3
3. Confirm `PRODUCTION_CLOUDFLARE_API_TOKEN` is **unchanged** (Workers AI sync).
4. Redacted audit (names only, no values):

```bash
gh secret list --env production --repo KARSIFT/vocanova-platform-sandbox \
  | grep -E 'PRODUCTION_CLOUDFLARE'
```

Expect `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` in **that** listing.
`gh secret list --repo ...` without `--env production` is **not** TEST-00.

### Do not run `--apply` in this task

`VOC-072-T02` proves `--verify-only` only. Remap removal (`--apply`) stays under
existing VOC-067-T05 / deploy-production dispatch authority after verify-only
passes.

## 5. Operator confirmation (resolved)

Complete. **Never record the token string.**

| Field | Value |
| --- | --- |
| Operator GitHub handle | `m-e-h-r-d-a-a-d` (founder) |
| Provisioning date (UTC) | 2026-08-13 |
| Date moved to production environment | 2026-08-13 |
| Cloudflare token display name | Not independently re-read from the dashboard. Suggested name: `vocanova-production-origin-rules-cutover`. Non-blocking Low finding, unresolved. |
| Cloudflare permissions summary | Zone Read + Origin Rules Edit on `vocanova.site` (founder confirmation against §3) |
| GitHub secret name | `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` |
| `PRODUCTION_CLOUDFLARE_API_TOKEN` left unchanged | Yes — `PRODUCTION_CLOUDFLARE_API_TOKEN` / `PRODUCTION_CLOUDFLARE_ACCOUNT_ID` still present, unmodified |
| Redacted `gh secret list` excerpt | `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` listed in production environment (updated-at `2026-08-13T21:24:26Z` per package confirmation; see gate status above) |

`VOC-072-AC-00` / `VOC-072-TEST-00` satisfied as of this record.

## 6. Post-provisioning verification (operator)

Run **after** the secret exists. These steps do not mutate Origin Rules.

**Local or bastion** (token via env, not committed):

```bash
export PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN='…'   # from GitHub secret
export CLOUDFLARE_API_TOKEN="$PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN"

infra/scripts/cloudflare-remove-production-origin-port-remap.sh --verify-only
```

**Success criteria:**

- Exit code `0`
- **No** `ERROR: zone not found` (empty `GET /zones` result)
- Output includes either `OK: no origin rules remap production hosts to port 8443`
  or an explicit `FOUND` message — both prove zone resolution succeeded

**Not in T00 scope:** `workflow_dispatch` with
`voc067_cloudflare_origin_cutover=verify-only` — that is `VOC-072-T02` after
`VOC-072-T01` wires the workflow to DEP-01.

## 7. Downstream tasks

| Task | Dependency on T00 |
| --- | --- |
| `VOC-072-T01` | Requires DEP-00/DEP-01 (§1) and secret present (§5) before merge is meaningful |
| `VOC-072-T02` | Requires T01 merged + live `--verify-only` in production CI |

## 8. Method and limitations

- Repository facts in §2 read from this branch's tip:
  `.github/workflows/deploy-production.yml`,
  `infra/scripts/cloudflare-remove-production-origin-port-remap.sh`,
  `specs/changes/VOC-067-production-outage-root-cause-consider-unifying/t05-live-cutover-evidence.md`,
  `specs/changes/VOC-037-begin-milestone-r2-production-readiness-docs/t06-production-provisioning-evidence.md`.
- No Cloudflare API call, no GitHub secret read/write, and no production
  `workflow_dispatch` from this implementer environment.
- No secret values appear in this file or elsewhere in the diff.
- Human production authority stayed with founder/ops (`m-e-h-r-d-a-a-d` per
  VOC-067-DEP-03) throughout.
