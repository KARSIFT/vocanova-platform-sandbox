---
evidence_id: VOC-072-EV-02
task_id: VOC-072-T02
acceptance_criteria: VOC-072-AC-02
tests: VOC-072-TEST-03
date: 2026-08-13
related_change: VOC-072
accountable_owner: m-e-h-r-d-a-a-d (founder/ops; VOC-067-DEP-03)
gate_status: pending_operator_execution
t01_merge_commit: e768ca853acf
t01_merge_pr: 580
---

# VOC-072-T02 — Production workflow `--verify-only` evidence

## Gate status: PENDING — founder/ops dispatch required

**`VOC-072-AC-02` is NOT satisfied at this revision.** This implementer run has
no GitHub authentication (`gh` / Actions API write access), cannot dispatch
`deploy-production.yml` against the **production** environment, and cannot read
job logs (unauthenticated log download returns `403`). No qualifying
`workflow_dispatch` run exists after `VOC-072-T01` merged. The operator must
execute §3 and fill §5 before this gate closes.

| Requirement | Status |
| --- | --- |
| `VOC-072-T01` merged (cutover job wired to DEP-01 secret) | **Pass** — commit `e768ca8`, PR #580, merged 2026-08-13T22:40:32Z to `develop` |
| `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` in production env | **Unverified here** — `VOC-072-EV-00` §5 still pending; operator must confirm before dispatch |
| Post-T01 `workflow_dispatch` with `verify-only` on wired revision | **Pending** — operator §3 |
| Job `VOC-067-T05 Cloudflare origin-port remap` succeeds (exit 0) | **Pending** |
| Redacted log shows zone resolution + remap status (not `zone not found`) | **Pending** — operator §5 |

## 1. Preconditions verified (read-only)

| Item | Status | Notes |
| --- | --- | --- |
| Package `VOC-072` adopted, T02 authorized | OK | `change.yaml`: `status: adopted`, `implementation.authorized: true` |
| T01 dependency issue #544 closed | OK | Roster dependency satisfied |
| T01 wiring on `develop` | OK | `voc067-cloudflare-cutover` binds `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` (commit `e768ca8`) |
| T01 wiring on `main` | **Not yet** | `origin/main` (`26e9076`) still binds `PRODUCTION_CLOUDFLARE_API_TOKEN` only — dispatch **must** target `develop`, not `main` |
| Prior failed verify-only (pre-T01 baseline) | Recorded | Run #28 below — expected failure with Workers-AI token |

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
T02 must produce a **new** successful run on the T01 revision with the DEP-01
secret.

## 2. Cross-reference — `VOC-067-EV-05` §3 (credential unblock)

[`t05-live-cutover-evidence.md`](../VOC-067-production-outage-root-cause-consider-unifying/t05-live-cutover-evidence.md)
§3 records that live `--verify-only` was not executed and that missing or
mis-scoped credentials are a **fail-closed** VOC-067-TEST-06 outcome, not a pass.

**When this gate closes (§5 filled with a successful run):**

- The **credential unblock** clause in VOC-067-TEST-06 is satisfied — zone
  resolution succeeded with the wired token, and remap status was reported.
- If verify output is `OK: no origin rules remap production hosts to port 8443`,
  founder/ops may update `cloudflare_remap_api_status: absent` in
  `VOC-067-EV-05` in a follow-up VOC-067 revision (out of scope for this task).
- If output is `FOUND: … origin rule(s) still remap to port 8443`, credential
  unblock is still satisfied; `--apply` remains a separate authorized step
  (VOC-072 open question 3 — do not dispatch `--apply` from this task).
- The `:8443` bridge must **remain** until remap absence is confirmed via
  verify output (`voc067-cutover-bridge-gate.test.mjs`).

## 3. Operator runbook — production `workflow_dispatch`

Accountable owner: `m-e-h-r-d-a-a-d` (or delegate with production environment
approval rights).

**Before dispatch:**

1. Confirm `VOC-072-EV-00` §5 — `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN`
   exists in GitHub → Settings → Environments → **production** (never paste
   the value into git).
2. Confirm you dispatch from branch **`develop`** at or after `e768ca8` (T01
   wiring). Dispatch from `main` still uses the old Workers-AI token binding
   and will fail zone lookup even if the new secret exists.

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

**Success criteria:**

- Workflow conclusion: `success`
- Job **VOC-067-T05 Cloudflare origin-port remap** conclusion: `success`
- Step **Run Cloudflare origin-port remap action** exit 0
- Log includes zone resolution (ruleset fetch after `resolve_zone_id`) and **one
  of**:
  - `OK: no origin rules remap production hosts to port 8443`
  - `FOUND: N origin rule(s) still remap to port 8443:` (with rule summary)
- Log must **not** include:
  - `ERROR: zone not found` (empty zone list / wrong token)
  - `ERROR: CLOUDFLARE_API_TOKEN … is required` (secret missing from env)

**Do not** set `voc067_cloudflare_origin_cutover=apply` in this task.

## 4. Monitor command (operator / reviewer)

```bash
# Latest deploy-production workflow_dispatch runs
curl -fsS -H "Accept: application/vnd.github+json" \
  "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/actions/workflows/deploy-production.yml/runs?event=workflow_dispatch&per_page=5" \
  | jq '.workflow_runs[] | {run_number, id, head_branch, head_sha, conclusion, html_url, created_at}'

# After a run id is known (requires auth for job logs):
gh run view <run-id> --repo KARSIFT/vocanova-platform-sandbox --log-failed
```

As of this evidence revision, the most recent `workflow_dispatch` verify-only
attempt is run **#28** (`31580068543`, **failure**, pre-T01). No post-`e768ca8`
verify-only run exists yet.

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
| Branch / commit | _pending_ (expect `develop` / `e768ca8` or later) |
| Job conclusion | _pending_ (expect `success`) |
| Remap status line | _pending_ (`OK: no origin rules…` or `FOUND: …`) |

### Redacted log excerpt

```text
(paste step "Run Cloudflare origin-port remap action" output here — no secrets)
```

When complete, update frontmatter `gate_status` to `resolved`.

## 6. Acceptance mapping (when §5 complete)

| Criterion | Expected |
| --- | --- |
| `VOC-072-AC-02` | Satisfied when §5 records successful verify-only with zone resolution |
| `VOC-072-TEST-03` | Satisfied by this document (`VOC-072-EV-02`) |
| VOC-067-TEST-06 credential clause | Unblocked (remap absence may still require separate `--apply`) |

## 7. Method and limitations

- Repository facts read from `develop` tip `e768ca8` and public GitHub Actions
  API (run metadata only; logs require auth).
- No `workflow_dispatch`, no production secret read/write, and no Cloudflare
  API call from this implementer environment.
- No secret values appear in this file.
- Independent verification should bind to the exact commit that adds §5 and
  confirm the quoted run URL and log lines against Actions log access.
