# VOC-050 — Release Plan

## Release and deployment authorization

Not authorized by this package. A merged package does not itself authorize
production deployment. `change.yaml`'s `release.deployment` and
`release.mode` are both set to `not-authorized-by-this-package`; a separate
release decision, made under the repository's active authority model
(A-003), is required before any task's change reaches production. Once
implemented and released, this package's own effect is to add a *new gate*
ahead of the existing, already-authorized R0-R2 `develop` → `main`
auto-promotion path (per `AGENTS.md`'s "Release and deployment authority"
section) - it does not itself change who may authorize that promotion or
under what risk class, per the issue's own stated constraint.

## Preconditions, monitoring, and outcome

- Exact revision: to be recorded as the merged commit SHA for each of
  `VOC-050-T00` through `VOC-050-T04` once implemented.
- Checks: `pnpm validate` (or its narrower scoped equivalents),
  `go test ./apps/api/business/auth/...`, `go test ./apps/api/cmd/seed/...`
  (or the new seed package's tests), and the live-observed workflow runs
  described in `test-plan.md`'s `TEST-03`/`TEST-04`/`TEST-05`.
- Approvals: independent verification by Claude Code against this package's
  specification and acceptance criteria, per `CLAUDE.md`; human approval
  proportionate to whichever risk floor each task's own path-based
  classification (`scripts/governance/classify-change-risk.sh`) actually
  computes at implementation time - likely R3 for `T00` (migrations), `T01`
  (authentication), and any workflow-file task, not the package-level R2
  proposed in `change.yaml`.
- Staged evidence: none exists yet; this package is unadopted and
  unimplemented.
- Monitoring: after release, watch for (a) any staging core-loop failure
  that does NOT block promotion (would mean `VOC-050-R02`'s cross-repo gap
  is real and unresolved), (b) any sign of the synthetic account's identity
  appearing in real user-facing surfaces, and (c) any increase in production
  deploy duration/failure rate attributable to the new smoke-test check.
- Outcome owner: unassigned; to be recorded at adoption time.

## Rollback

- Trigger: any evidence the synthetic account is reachable via a real
  signup path, any evidence the session-minting mechanism is reachable
  without its gating token, any evidence the production core-loop check
  performs a state-mutating request against real infrastructure, or the
  staging/production core-loop check itself becoming a source of
  false-positive deploy failures that block legitimate releases.
- Mechanism: revert the merged commit/PR for the affected task. `T00`'s
  seeded row should be deleted if its migration is reverted, to avoid an
  orphaned synthetic account with no corresponding code expecting it. No
  other task introduces stateful data requiring migration-compatible
  rollback.
- Accountable owner: unassigned; to be recorded at adoption time, per this
  repository's active authority model.
- Validation: re-run the applicable `go test`/`pnpm validate` scope and
  confirm `deploy-staging.yml`/`deploy-production.yml` return to their
  pre-VOC-050 behavior (healthz/`/`-poll only; SKIP-based production
  core-loop check) after any rollback.
- Last-known-good reference: the `develop` branch commit immediately prior
  to the affected task's merge.

## Independent verification, human approvals, and closure

- Independent verification: Claude Code reviews the exact final revision of
  each task's pull request, binds its report to that commit SHA, and
  confirms Codex (or whichever implementer) did not approve or merge its
  own work, per `CLAUDE.md`. For `T03` specifically, independent
  verification must also confirm the pull request accurately discloses the
  cross-repo `release.yml` dependency rather than overstating what was
  proven.
- Human approvals: routine R3 under active A-003 does not require standing
  technical-steward or founder approval merely for being R3; strengthened
  applicable controls and independent verification remain required. The
  reviewing human should separately decide, at adoption time, whether `T01`
  (a mechanism capable of minting authenticated sessions outside the normal
  flow) warrants review beyond routine R3 given its authentication-bypass
  shape, even though it is scoped to a single synthetic account.
- Closure evidence: this package closes only once all five tasks are
  implemented, independently verified, and `acceptance-criteria.md`'s
  criteria are all marked satisfied with linked evidence - repository merge,
  release, activation, and closure remain distinct events and must not be
  conflated. Closure explicitly does NOT certify that `karsift-ai-infra`'s
  `release.yml` gates on the new staging signal (see `VOC-050-AC-04`'s own
  scoping) - that remains a separately-tracked, human-confirmed item.
