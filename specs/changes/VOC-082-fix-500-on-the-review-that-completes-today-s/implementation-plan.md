# VOC-082 — Implementation Plan

## Preconditions and protected areas

Do not begin implementation until this package is adopted (`status:
adopted`, `approval_status` per house adoption convention,
`implementation_authorized: true` / `implementation.authorized: true` in
`change.yaml`). Under **active A-004**, adoption does not wait on a
founder `approved` comment; exact-revision plan review and deterministic
gates still apply.

Additionally:

- Treat issue #675's confirmed cause as the starting point
  (`VOC-082-DEP-00`). If T00 evidence shows a different primary cause,
  stop and record it before expanding scope.
- Do not touch VOC-081 monitor / shared-edge paths (`VOC-082-DEP-01`).
- Path floor R1; proposed package risk **R2** — staging evidence and
  rollback credibility required; no founder-comment merge gate under
  A-004.
- Default: no migrations; forward-fix only unless adoption expands
  stuck-row remediation.

## File reconciliation and implementation sequence

1. **`VOC-082-T00`** — Fix `ReconcileStreak` in
   `apps/api/business/gamification/streak.go` so
   `currentCompletion=true` + today already `completed` in the snapshot
   list is handled as the current completion (not
   `ErrInvalidStreakSnapshot`). Keep rejecting `lastGood` after today.
   Add regression coverage in `streak_test.go` (and transactional /
   repository coverage if the existing harness can prove completing-
   review atomic success without unrelated refactors). Confirm
   `applyP4ReviewWiring` call order in `postgres.go`; edit that file
   only if a call-site defect beyond streak reconciliation is proven.
2. **`VOC-082-T01`** — After T00 merges to `develop`, obtain a real
   staging core-loop run through the completing review; record PASS (or
   honest FAIL) evidence. No further product change expected unless
   verification surfaces a narrow residual gap attributable to this
   package.

Preserve existing streak rules (a)–(d), grace-day behavior, and ledger
idempotency. Do not invent staging PASS evidence.

## Validation and independent verification

Deterministic commands before claiming T00 complete (adjust package
filter to match repo conventions in `docs/development.md`):

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
cd apps/api && go test ./business/gamification/ ./business/reviews/ -count=1
```

Also run the workspace validation documented in `docs/development.md`
when web/E2E files change (`pnpm validate` or the narrower documented
subset). Missing credentials or inability to run staging is a
limitation, not a pass.

Independent verification (per `CLAUDE.md`) must bind the exact commit
SHA per task, confirm the implementer-role occupant did not
approve/merge its own work, confirm authority model **`a004-active`**,
escalate if semantic risk exceeds the R2 proposal, and report every
still-required evidence obligation. The implementer must not
self-approve.

## Deployment and rollback

Authorization: this package does **not** itself authorize production
deployment by draft or by adoption alone. After task merges, existing
auto-promotion / `deploy-production.yml` on `main` push behavior applies
per AGENTS.md under active A-004.

Rollout sequence:

1. Merge T00 → `develop` with independent verification PASS (or PASS
   WITH NON-BLOCKING FINDINGS).
2. Staging deploy + T01 core-loop evidence.
3. Package roster completion → promote `develop` → `main` → production
   deploy via existing automation.

Rollback: revert T00's commit(s) if completing reviews regress (new
5xx, incorrect streak breaks, double rewards). Last-known-good is the
tree immediately preceding T00's merge (known-broken completing-review
path). Validate rollback by confirming review submissions below target
still succeed and that the prior 500-at-target behavior is understood as
the pre-fix baseline if reintroduced.
