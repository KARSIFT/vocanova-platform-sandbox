---
evidence_id: VOC-079-EV-00
task_id: VOC-079-T00
acceptance_criteria: VOC-079-AC-00
tests: VOC-079-TEST-00
date: 2026-08-15
related_change: VOC-079
accountable_owner: m-e-h-r-d-a-a-d (founder/ops; VOC-067-DEP-03)
gate_status: pending_operator_execution
develop_tip_at_implement: 8e3d649
main_tip_at_implement: a982e4b
voc072_t01_merge_commit: e768ca8
---

# VOC-079-T00 — Cloudflare remap absence evidence and EV gate update

## Gate status: PENDING — founder/ops dispatch required

**`VOC-079-AC-00` is NOT satisfied at this revision.** This implementer run has
no GitHub authentication (`gh` / Actions API write access), cannot dispatch
`deploy-production.yml` against the **production** environment, cannot list
production environment secrets, and cannot download job logs (unauthenticated log
download returns `403`). No qualifying post-T01 `workflow_dispatch` run with
`voc067_cloudflare_origin_cutover=verify-only` exists yet. The operator must
execute §3 and fill §5 before this gate closes.

**Do not** set `cloudflare_remap_api_status: absent` in
[`t05-live-cutover-evidence.md`](../VOC-067-production-outage-root-cause-consider-unifying/t05-live-cutover-evidence.md)
until §5 records `OK: no origin rules remap production hosts to port 8443` from
a successful repository run.

| Requirement | Status |
| --- | --- |
| Package `VOC-079` adopted, T00 authorized | **Pass** — `change.yaml`: `status: adopted`, `implementation.authorized: true` |
| `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` in production env | **Unverified here** — founder confirmed provision in [`VOC-072-EV-00`](../../VOC-072-same-request-as-github-issue-535-voc-067-t05/t00-token-provisioning-evidence.md) §Gate status; operator must re-confirm before dispatch |
| VOC-072-T01 wiring on dispatched ref | **Pass (read-only)** — `deploy-production.yml` binds `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` on `develop` (`8e3d649`) and `main` (`a982e4b`); T01 merge `e768ca8` is ancestor of both |
| Post-T01 `workflow_dispatch` with `verify-only` | **Pending** — operator §3 |
| Job `VOC-067-T05 Cloudflare origin-port remap` succeeds (exit 0) | **Pending** |
| Log includes zone resolution + `OK: no origin rules remap…` | **Pending** — operator §5 |
| VOC-067-EV-05 `cloudflare_remap_api_status` → `absent` | **Pending** — only after §5 `OK` line |

## 1. Preconditions verified (read-only)

| Item | Status | Notes |
| --- | --- | --- |
| Requirement source issue #624 | OK | Package `requirement_source` cites #624 |
| VOC-079-DEP-02 (verify-only mandatory) | OK | Default adoption path; dashboard-only is insufficient |
| Bridge-retention gate still active | OK | [`t05-live-cutover-evidence.md`](../VOC-067-production-outage-root-cause-consider-unifying/t05-live-cutover-evidence.md) frontmatter `cloudflare_remap_api_status: unconfirmed` |
| Cutover script + workflow wiring | OK | `infra/scripts/cloudflare-remove-production-origin-port-remap.sh --verify-only`; `deploy-production.yml` job `voc067-cloudflare-cutover` |
| Prior failed verify-only (pre-T01) | Recorded | Run **#28** below — expected failure with Workers-AI token on `main` |
| Post-T01 successful verify-only | **None yet** | Public API shows no `workflow_dispatch` verify-only after `e768ca8` |

### Baseline failure (issue #535 — pre-T01 wiring)

| Field | Value |
| --- | --- |
| Workflow | `deploy-production` |
| Run number | **#28** |
| Run id | `31580068543` |
| URL | <https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31580068543> |
| Job URL | <https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31580068543/job/94060874049> |
| Event | `workflow_dispatch` |
| Branch | `main` |
| Commit | `6af09be` (pre-T01) |
| Job conclusion | `failure` |
| Failing step | **Run Cloudflare origin-port remap action** |

That run used `PRODUCTION_CLOUDFLARE_API_TOKEN` (Workers-AI scope). Zone lookup
returned an empty `GET /zones` result → `ERROR: zone not found` (issue #535).
T00 must produce a **new** successful run on a T01-wired revision with the
zone-scoped secret.

## 2. Cross-reference — VOC-072 and VOC-067 gates

| Artifact | Role |
| --- | --- |
| [`VOC-072-EV-00`](../../VOC-072-same-request-as-github-issue-535-voc-067-t05/t00-token-provisioning-evidence.md) | Token provisioning (`gate_status: resolved`) — name/location only; not remap absence |
| [`VOC-072-EV-02`](../../VOC-072-same-request-as-github-issue-535-voc-067-t05/t02-verify-only-evidence.md) | Parallel pending operator dispatch template (same limitation as this file) |
| [`VOC-067-EV-05`](../VOC-067-production-outage-root-cause-consider-unifying/t05-live-cutover-evidence.md) | Bridge-retention frontmatter; update to `absent` only after this task's §5 `OK` |

**When this gate closes (§5 filled with successful verify-only and `OK` output):**

1. Set `cloudflare_remap_api_status: absent` in `t05-live-cutover-evidence.md`
   and cite the verifying run URL in that file's summary table.
2. `scripts/foundation/voc067-cutover-bridge-gate.test.mjs` will then allow
   bridge retirement in a follow-up task (`VOC-079-T02`).
3. If output is `FOUND: N origin rule(s) still remap to port 8443`, do **not**
   set `absent`; run authorized `apply` / `--apply` per
   [`VOC-079-D00`](specification.md) and re-verify before bridge retirement.

## 3. Operator runbook — production `workflow_dispatch`

Accountable owner: `m-e-h-r-d-a-a-d` (or delegate with production environment
approval rights).

**Before dispatch:**

1. Confirm `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` exists in GitHub →
   Settings → Environments → **production** (never paste the value into git):

   ```bash
   gh secret list --env production --repo KARSIFT/vocanova-platform-sandbox \
     | grep -E 'PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN'
   ```

2. Dispatch from **`develop`** at or after `e768ca8` (T01 wiring), or from
   **`main`** at or after `a982e4b` (2026-08-15 develop→main promotion). Both
   branches now bind the zone-scoped secret; pre-T01 `main` revisions still fail
   zone lookup even if the secret exists.

**Dispatch (GitHub CLI):**

```bash
gh workflow run deploy-production.yml \
  --repo KARSIFT/vocanova-platform-sandbox \
  --ref develop \
  -f production_web_host=production.vocanova.site \
  -f production_api_host=api-production.vocanova.site \
  -f production_api_base_url=https://api-production.vocanova.site \
  -f new_user_signup_allowlist= \
  -f voc067_cloudflare_origin_cutover=verify-only
```

**Or:** Actions → **deploy-production** → **Run workflow** → branch
**develop** → set `voc067_cloudflare_origin_cutover` to **verify-only** → Run.

**Success criteria (AC-00):**

- Workflow conclusion: `success`
- Job **VOC-067-T05 Cloudflare origin-port remap** conclusion: `success`
- Step **Run Cloudflare origin-port remap action** exit 0
- Log includes zone resolution (ruleset fetch after `resolve_zone_id`) and:

  ```text
  OK: no origin rules remap production hosts to port 8443
  ```

- Log must **not** include:
  - `ERROR: zone not found` (empty zone list / wrong token)
  - `ERROR: CLOUDFLARE_API_TOKEN … is required` (secret missing from env)

**If log shows `FOUND:`** — do not set `absent`. Either dispatch with
`voc067_cloudflare_origin_cutover=apply` (authorized cutover path) or run
`infra/scripts/cloudflare-remove-production-origin-port-remap.sh --apply` with
the same token, then repeat verify-only until `OK`.

**Do not** dispatch `apply` from this evidence task unless remap is still present.

## 4. Monitor command (operator / reviewer)

```bash
# Latest deploy-production workflow_dispatch runs
curl -fsS -H "Accept: application/vnd.github+json" \
  "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/actions/workflows/deploy-production.yml/runs?event=workflow_dispatch&per_page=5" \
  | jq '.workflow_runs[] | {run_number, id, head_branch, head_sha, conclusion, html_url, created_at}'

# After a run id is known (requires auth for job logs):
gh run view <run-id> --repo KARSIFT/vocanova-platform-sandbox --log-failed
```

As of this evidence revision (`2026-08-15`), the most recent `workflow_dispatch`
verify-only attempt is run **#28** (`31580068543`, **failure**, pre-T01). No
post-`e768ca8` verify-only run exists yet.

## 5. Qualifying run log excerpt (fill after dispatch)

Complete this section after §3 succeeds. **Redact** any token material, full
account IDs, or raw API error bodies beyond what audit needs.

| Field | Value |
| --- | --- |
| Operator GitHub handle | _pending_ |
| Dispatch date (UTC) | _pending_ |
| Workflow run number | _pending_ |
| Workflow run id | _pending_ |
| Run URL | _pending_ |
| Branch / commit | _pending_ (expect `develop` / `8e3d649` or later, or `main` / `a982e4b` or later) |
| Job conclusion | _pending_ (expect `success`) |
| Remap status line | _pending_ (expect `OK: no origin rules remap production hosts to port 8443`) |

### Redacted log excerpt

```text
(paste step "Run Cloudflare origin-port remap action" output here — no secrets)
```

**After §5 records the `OK` line:**

1. Update frontmatter `gate_status` here to `resolved`.
2. Update [`t05-live-cutover-evidence.md`](../VOC-067-production-outage-root-cause-consider-unifying/t05-live-cutover-evidence.md) frontmatter `cloudflare_remap_api_status: absent` and note the verifying run in §Summary.

## 6. Acceptance mapping (when §5 complete)

| Criterion | Expected |
| --- | --- |
| `VOC-079-AC-00` | Satisfied when §5 records successful verify-only with `OK: no origin rules…` |
| `VOC-079-TEST-00` | Satisfied by this document (`VOC-079-EV-00`) with cited run URL and log excerpt |
| VOC-067-EV-05 gate | `cloudflare_remap_api_status: absent` only after §5 `OK` (same revision or immediate follow-up commit) |
| `VOC-079-T01` unblock | Depends on `absent` frontmatter per `implementation-plan.md` |

## 7. Method and limitations

- Repository facts read from `develop` tip `8e3d649` and `main` tip `a982e4b`,
  plus public GitHub Actions API (run metadata only; logs require auth).
- No `workflow_dispatch`, no production secret read/write, and no Cloudflare API
  call from this implementer environment (`GITHUB_TOKEN` / `GH_TOKEN` not set in
  the Cursor implementer shell).
- No secret values appear in this file.
- Founder dashboard confirmation (issue #624 context) is supporting context only;
  it does not satisfy AC-00 without the repository transcript (`VOC-079-DEP-02`).
- Independent verification should bind to the exact commit that completes §5
  (including `absent` frontmatter update when applicable) and confirm the quoted
  run URL and log lines against Actions log access.
