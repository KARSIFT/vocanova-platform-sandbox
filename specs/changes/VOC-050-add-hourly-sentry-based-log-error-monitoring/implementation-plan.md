# VOC-050 — Implementation Plan

## Preconditions and protected areas

Do not begin until this package is adopted (`change.yaml`'s `status: adopted`
and `implementation_authorized: true`, set by a human, never by this
package's own drafting or by any agent). Once adopted:

- `VOC-050-T00` must complete and its finding reviewed before `VOC-050-T01`
  or `VOC-050-T02` begin — both depend on the confirmed Sentry
  project/DSN layout.
- `VOC-050-T01` and `VOC-050-T02` may proceed in parallel once `T00`
  completes (see `tasks.md`); `VOC-050-T03` is a documentation-only
  companion to `T02` (the implementer does not obtain the real secret
  value — see `AGENTS.md`'s "Safety" section on agents not receiving
  production secrets); `VOC-050-T04` lands the documentation-consistency
  update in the same PR as `T02` or immediately after.
- Protected areas: `.github/workflows/` (new file + edits to
  `deploy-staging.yml`/`deploy-production.yml`), GitHub Actions secrets
  (new secret), `docs/operations/11-devops-and-ci-cd.md` (approved
  governance document — amend, do not silently rewrite). No database
  migration path, no authentication/authorization code path is touched.

## File reconciliation and implementation sequence

1. **`VOC-050-T00`**: confirm Sentry organization/plan capacity and record
   the exact per-environment DSN/project layout. No file change beyond this
   task's own evidence record.
2. **`VOC-050-T01`**: add `@sentry/nextjs` to `apps/web/package.json`; add
   Sentry instrumentation files consistent with the Next.js App Router
   integration shape; wire DSN/environment/release env vars parallel to
   `apps/api/app/api/production.go`'s existing `SentryDSN`/
   `SentryEnvironment`/`SentryRelease` fields; update
   `apps/web/.env.example`; update `.github/workflows/deploy-staging.yml`
   and `.github/workflows/deploy-production.yml` to inject the correct
   per-environment DSN, mirroring `deploy-production.yml`'s existing
   `PRODUCTION_SENTRY_DSN` injection pattern (including its
   both-or-neither partial-configuration guard).
3. **`VOC-050-T02`** (may run in parallel with `T01`): add
   `.github/workflows/error-monitoring.yml` with an hourly `schedule:`
   trigger plus `workflow_dispatch:` for manual verification; implement the
   Sentry API query, duplicate-check guard, and unlabeled-issue creation
   logic; declare a minimum `permissions:` block; declare and use no SSH
   secret.
4. **`VOC-050-T03`**: document the required `SENTRY_API_AUTH_TOKEN` secret
   name/scope in the workflow file's own comments and this package's
   README/impact-analysis; do not attempt to provision the real value.
5. **`VOC-050-T04`**: amend `docs/operations/11-devops-and-ci-cd.md` §1's
   "Error monitoring" row, following the existing
   `VOC-032-§1-amendment`-style annotation convention, in the same PR as
   `T02` or an immediate follow-up.

No existing file is deleted. `apps/web`'s existing build/test configuration
is extended, not replaced. `.github/workflows/deploy-staging.yml` and
`.github/workflows/deploy-production.yml` gain new env-injection steps
alongside their existing ones, following the same pattern already proven for
`apps/api`'s Sentry wiring — no existing step in either file is removed or
restructured by this package.

## Validation and independent verification

- Deterministic: `pnpm validate` (or the narrower `pnpm lint` / `pnpm
  typecheck` / `pnpm test` / `pnpm build` for `apps/web`) for `VOC-050-T01`;
  a workflow-syntax check (`actionlint` if installed, or GitHub's own
  workflow-validation on push) plus a `workflow_dispatch` dry run against a
  disposable/test GitHub issue tracker or a scoped test repository for
  `VOC-050-T02`, since a scheduled workflow's actual cron trigger cannot be
  deterministically tested pre-merge; `bash scripts/governance/validate-governance.sh`
  and `bash scripts/governance/classify-change-risk.sh` for `VOC-050-T04`'s
  documentation amendment and for the overall task-scoped file list, to
  confirm the real detected risk floor against this draft's proposed R3
  (see `change.yaml`'s `blocking_reasons`).
- Exact-SHA independent verification: Claude Code (per `CLAUDE.md`) confirms,
  against the final revision of each task's PR: (a) the Sentry DSN wiring is
  genuinely disabled when unset and does not leak a DSN value into client
  bundles beyond the intended `NEXT_PUBLIC_`-prefixed variable; (b) the new
  Sentry API auth token's documented scope is genuinely read-only and
  least-privilege, not merely described as such; (c) the scheduled
  workflow's `permissions:` block is minimum-necessary and it declares no
  SSH secret; (d) the duplicate-check guard is keyed on a stable identifier,
  not fragile title-text matching alone; (e) DOC-11's amendment note follows
  the existing annotation convention rather than silently rewriting history.

## Deployment and rollback

- Authorization boundary: `VOC-050-T01`'s `apps/web` changes deploy through
  the existing, already-governed `deploy-staging.yml`/`deploy-production.yml`
  paths — no new deployment mechanism is introduced. `VOC-050-T02`'s
  scheduled workflow has no "deployment" of its own beyond merging to the
  default branch, after which its `schedule:` trigger activates
  automatically; this package does not request or need any manual dispatch
  step for normal operation (only for verification, via
  `workflow_dispatch:`).
- Rollout sequence: `T01` ships with Sentry disabled (no DSN secret set)
  until the human adopting this package provisions the real per-environment
  DSN values (`VOC-050-T03`'s documented precondition) — so merging `T01`'s
  code has no live effect until secrets are set, a safe default rollout.
  `T02`'s scheduled workflow similarly has no effect until
  `SENTRY_API_AUTH_TOKEN` is provisioned; until then, the workflow should
  fail clearly (per `tasks.md`'s `VOC-050-T03`) rather than silently no-op,
  making the missing precondition visible in Actions run history rather than
  invisible.
- Rollback trigger: the scheduled workflow spamming duplicate or
  low-signal issues (per `impact-analysis.md`'s `VOC-050-R01`), or the
  `apps/web` Sentry SDK causing any client-side performance regression or
  the debug overlay leaking into production.
- Rollback mechanism: disable the scheduled workflow (delete/disable the
  `.github/workflows/error-monitoring.yml` trigger, or revert its merge
  commit) for the monitoring workflow; revert `VOC-050-T01`'s merge commit,
  or simply unset the DSN secret (reverting to the existing no-op-when-unset
  default), for the `apps/web` wiring. No data migration is introduced, so
  no data-compatibility rollback work is required.
- Owner: unassigned; to be recorded at adoption time.
- Last-known-good reference: `develop`'s tip immediately prior to this
  package's first merged task PR.
