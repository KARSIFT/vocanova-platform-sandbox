# VOC-110 — Fix deploy-staging failure after dependabot integration merge: Specification

## Objective and requirement source

Remediate the operational failure recorded in
[GitHub issue #911](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/911):
`deploy-staging` ended `failure` after merge of Dependabot PR #859 to `develop`.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

Primary evidence (issue #911 + public run page and Actions REST metadata for run
32566405628):

| Item | Value |
|------|-------|
| Workflow | `deploy-staging` |
| Run | [#364](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32566405628), push-triggered 2026-08-22 09:59 UTC |
| Head branch / SHA | `develop` @ `f25e4ccf5fc28dcc5b14a438fbdc4f93e5c53a46` |
| Conclusion | `failure` (workflow wall clock ~8m 28s) |
| Job | `deploy to staging` (~8m 23s, exit code 1) |
| Trigger commit | Merge PR **#859** — Dependabot `minor-and-patch` npm group (eight updates) |
| Public artifacts | Two Docker build attestations produced during the run |
| Public annotations | Process exit code 1; Playwright report/test-results paths missing on failure upload; benign Go cache restore warning |
| Issue origin | VOC-088-T02 `operational-failure-monitoring.yml` sanitized issue body |

Drafting-time bounding (not a root-cause determination):

- **Not VOC-094:** job ran ~8m with conclusion `failure`, not `cancelled` with
  zero jobs and a concurrency-supersession annotation.
- **Not VOC-095:** total duration is well below the 40-minute job timeout; public
  signals point past image build/push rather than an unbounded Playwright install stall.
- **Likely post-build phase:** Docker build artifacts exist; duration fits build/push
  plus SSH deploy and post-deploy gates (health checks, OAuth start check, staging
  core-loop Playwright journey per VOC-050-T02).
- **Dependabot context:** the failing run's head is the first integration commit
  carrying eight npm minor/patch bumps — dependency regression is a leading hypothesis
  until logs name the failing step.

The exact failing workflow step and remediation surface are **open until T00 reads
job logs** at implementation time.

## Scope and non-goals

In scope:

1. Read run 32566405628 job logs for `deploy to staging` at implementation time.
   Record the failing step name and sanitized failure class in `t00-evidence.md`
   (no secrets, SSH output, session cookies, OAuth state, tokens, or personal data).
2. Fix the identified root cause in the smallest correct surface:
   - **Application / dependency regression** — fix `apps/web/`, shared `packages/`,
     or lockfile-resolved behavior when evidence shows staging runtime or UI broke
     after the Dependabot merge.
   - **Staging core-loop test** — update `apps/web/tests/staging-e2e/` when evidence
     shows the product is correct but the journey assertion is stale.
   - **Deploy harness / script** — fix `infra/scripts/` (OAuth verify, Playwright
     install, session mint helpers) when evidence shows wiring or harness defect.
   - **Workflow wiring** — minimal `deploy-staging.yml` change only when evidence
     shows misconfiguration unrelated to the dependency bump.
3. Preserve fail-closed deploy semantics: no `continue-on-error`, no skipped health
   checks, no weakened OAuth or core-loop ordering (VOC-050, VOC-084, VOC-088).
4. Extend deterministic tests to lock the fix or a regression fixture matching T00
   evidence.
5. Live verification (T01): after T00 merges to `develop`, record a `deploy-staging`
   run for a revision containing the fix that reaches conclusion `success`.

Non-goals / explicitly excluded:

- Weakening or removing the staging core-loop gate, OAuth-start check, health polls,
  or deploy concurrency posture (VOC-094 `queue: max` remains).
- Changing operational-failure observer behavior (VOC-088 worked as designed by opening
  issue #911).
- Reverting Dependabot routing to `main` or bypassing governed integration for future
  dependency updates.
- Broad dependency rollback without evidence; if the fix is pinning or reverting one
  package, T00 must document which bump caused the failure.
- Modifying production deploy semantics, signup policy, Kuma inventory, or migrations
  unless T00 proves a separate defect requiring its own package.
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R3** (`.github/workflows/`, `infra/scripts/`,
  possible `apps/web/` / `packages/`).
- **Measured path floor at drafting:** **R3**. Not R4 unless a task touches
  `scripts/governance/*`.
- Protected areas: `.github/workflows/deploy-staging.yml`, staging SSH deploy path,
  `apps/web/tests/staging-e2e/core-loop.staging.spec.ts`, repository secrets consumed
  by deploy-staging (read-only use in CI only).
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

## Decisions

`VOC-110-D00`: Run 32566405628 is an **actionable deploy-staging failure** (job
started, exit code 1, ~8m duration), not a benign concurrency supersession or a
Playwright-install timeout class failure.

`VOC-110-D01`: T00 must name the failing workflow step from job logs before choosing
the remediation surface. Drafting-time hypotheses (dependency regression vs harness
vs deploy step) are not decisions.

`VOC-110-D02`: Evidence and fixes stay within VOC-088 sanitization boundaries: bounded
metadata in issues and package evidence; job logs inform T00 but are not copied into
issue bodies or committed evidence files.

`VOC-110-D03`: The fix must preserve VOC-050's post-deploy core-loop gate unless T00
proves the gate assertion is wrong — in that case adjust tests or product behavior,
do not skip the gate.

`VOC-110-D04`: If logs show SSH/migration/health failure unrelated to PR #859's npm
bumps, T00 still fixes that defect but must record the causal link (or lack thereof)
explicitly so reviewers can scope escalation if it is environmental.

`VOC-110-D05`: Dependabot PR #859 is context, not automatic blame. The task PR must
record whether the fix is a direct consequence of a listed package bump, a latent
defect exposed by timing, or an unrelated staging host issue.

## Open questions for the reviewing human

1. Accept proposed **R3**, or raise in writing if the adopting human treats
   deploy-staging remediation touching `apps/web/` after a dependency merge as R4
   operational risk.
2. After T00 log review, confirm whether the fix belongs in application code,
   staging E2E tests, infra scripts, or workflow wiring — the task PR must record
   the chosen branch.
3. If the root cause is a Dependabot bump that cannot be safely adapted in-repo,
   is a targeted revert of that bump within scope, or does it require a separate
   dependency-governance package?

## Data, migrations, analytics, and accessibility

- No application schema migration unless T00 proves an unrelated data defect requiring
  a separate governed package.
- No intentional staging database mutation beyond existing deploy seed/migrate steps.
- Possible product UI fix if core-loop failure is a rendering or routing regression.
- No analytics change — evidence-backed non-applicability unless a route fix
  incidentally touches analytics (record in task PR if so).
- Accessibility impact follows any UI fix; no standalone a11y scope unless required
  by the chosen fix (record in task PR).
