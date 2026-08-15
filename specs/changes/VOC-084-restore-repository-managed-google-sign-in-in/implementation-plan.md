# VOC-084 — Implementation Plan

## Preconditions and protected areas

Do not begin implementation until this package is adopted (`status:
adopted`, `approval_status` per house adoption convention,
`implementation_authorized: true` / `implementation.authorized: true` in
`change.yaml`). Under **active A-004**, adoption does not wait on a founder
`approved` comment; exact-revision plan review and deterministic gates still
apply.

Additionally:

- Treat issue #691's verified root cause as the starting point
  (`VOC-084-DEP-00`). If implementer evidence shows a different primary cause,
  stop and record it before expanding scope.
- Do not mutate Google Cloud unless access already exists; otherwise record
  the exact external callback registration (`VOC-084-DEP-01`).
- Resolve or accept the recommended default for staging allowlist control
  (`VOC-084-DEP-02`) at adoption.
- Path floor **R3**; proposed package risk **R3** — strengthened evidence and
  rollback credibility required; no founder-comment merge gate under A-004.
- Preserve staging/production isolation and shared-edge topology (VOC-067).
- Never commit secrets or print credential values.

## File reconciliation and implementation sequence

1. **`VOC-084-T00` — Staging deploy OAuth sync + config + allowlist + tests**
   - Add Google OAuth credential synchronization to
     `.github/workflows/deploy-staging.yml` mirroring
     `deploy-production.yml`'s safe pattern (both unset → coherent skip /
     disabled; partial → fail before convergence; both present → write to
     `/opt/vocanova/infra/secrets/api.env` with `chmod 600`).
   - Write/overwrite staging application OAuth-related config on each
     deploy: exact
     `OAUTH_REDIRECT_URI=https://api-staging.vocanova.site/api/v1/auth/oauth/google/callback`,
     matching allowlist/redirect-allowlist shape appropriate to staging web
     host, `GOOGLE_OAUTH_ENABLED` derived from pair availability, 
     `NEW_USER_SIGNUP_ENABLED=false`, and
     `NEW_USER_SIGNUP_ALLOWLIST` from the chosen repository-workflow control
     (default empty).
   - Update staging deploy header/docs comments that still claim every
     non-AI secret is exclusively founder-populated if that claim becomes
     false for the Google pair / allowlist keys.
   - Add deterministic workflow/config tests (both-present, both-absent,
     partial rejection, canonical callback, allowlist default/control).
   - Do not edit production OAuth behavior.

2. **`VOC-084-T01` — Capability-gated Google button**
   - Teach `apps/web` sign-in to consume a deploy-derived capability signal
     (prefer existing unauthenticated `/healthz`
     `kill_switches.oauth_enabled`).
   - Render "Continue with Google" only when capability reports enabled.
   - Add deterministic UI tests for disabled-method rendering.
   - Preserve existing sign-in a11y and the `max-w-[28rem]` workaround.

3. **`VOC-084-T02` — Post-deploy OAuth-start check + Google callback disposition**
   - Add a post-deploy step in `deploy-staging.yml` that POSTs OAuth start
     with `redirectUri=https://staging.vocanova.site/home` (or the
     documented staging home redirect), asserts HTTP 200, Google host, and
     exact staging callback `redirect_uri`, and does **not** follow Google.
   - When credentials are both absent, the check must expect coherent
     disabled behavior rather than inventing a green OAuth start.
   - Determine from available access/evidence whether the Google client
     authorizes the staging callback; if unavailable, record the exact
     external configuration requirement in `t02-evidence.md`.
   - Confirm public health endpoints / shared-edge remain healthy after
     deploy.

Preserve compatible existing work (production pattern, `/healthz`
kill-switch reporting, VOC-038 allowlist semantics). Prefer minimal diffs.

## Validation and independent verification

Deterministic commands before claiming tasks complete (adjust to
`docs/development.md` and package filters):

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
# foundation / workflow tests for deploy-staging OAuth branches
# (exact path chosen in T00; follow existing scripts/foundation style)
pnpm validate   # or narrower lint/typecheck/test when apps/web changes
```

Missing credentials or inability to inspect Google Cloud is a limitation to
record under AC-07, not a silent pass for Console authorization.

Independent verification (per `CLAUDE.md`) must bind the exact commit SHA
per task, confirm the implementer-role occupant did not approve/merge its
own work, confirm authority model **`a004-active`**, escalate if semantic
risk exceeds the R3 proposal, and report every still-required evidence
obligation. The implementer must not self-approve.

## Deployment and rollback

Authorization: this package does **not** itself authorize production
deployment by draft or by adoption alone. Staging deploy follows merge to
`develop` via existing `deploy-staging.yml`. Production promotion remains the
normal develop→main path and must retain existing production OAuth behavior.

Rollout sequence:

1. Merge T00 → `develop` with independent verification PASS (or PASS WITH
   NON-BLOCKING FINDINGS); staging deploy converges OAuth config.
2. Merge T01 → capability-gated UI on staging.
3. Merge T02 → live OAuth-start gate + Google callback disposition evidence.
4. Later production promotion must not alter production OAuth sync semantics.

Rollback:

- Trigger: partial-pair writes, credential log exposure, wrong callback URI,
  signup kill switch flipped on, UI advertising disabled Google, production
  OAuth regression, or topology/health breakage.
- Mechanism: revert the package task revision(s) and redeploy staging via
  the repository path. Absent-pair state must fail coherently disabled.
- Last-known-good: tree immediately preceding the first merged task of this
  package (known-broken staging OAuth start / advertise-while-disabled per
  issue #691).
