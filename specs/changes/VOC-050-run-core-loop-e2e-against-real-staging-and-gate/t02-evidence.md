# VOC-050-EV-02 — T02 evidence and open-question resolutions

Evidence for `VOC-050-T02` (`VOC-050-AC-03`, test `VOC-050-TEST-03`).

This file carries the decisions `tasks.md`'s `VOC-050-T02` requires the
implementer to state explicitly (`specification.md`'s open questions 2 and 3).
They live in the repository rather than in the pull-request body because an
`implement.yml` run generates that body from fixed package metadata — the
implementer cannot author it. The previous revision's independent review
recorded the missing written resolution for exactly that reason; a committed
file is the durable record.

No secret value, no production data, and no real user data appears here.

## 1. Open question 3 — a separate spec file and a separate Playwright config

**Chosen: a sibling spec in its own directory,
`apps/web/tests/staging-e2e/core-loop.staging.spec.ts`, driven by
`apps/web/playwright.staging.config.ts`** — not an environment-aware variant of
`apps/web/tests/e2e/core-loop.spec.ts`.

Reasons, in order of weight:

1. The two runs cannot share a setup step. `core-loop.spec.ts` step 1 invents
   its session (`test-session-${randomUUID()}`) and relies on
   `mock-api-server.mjs` accepting any value; the real Go API only accepts a
   session row created by `CreateSession`. An environment-aware branch would
   fork at the first line and stay forked through onboarding and fixture
   selection.
2. It keeps the PR-time suite untouched and unchanged in behaviour, which
   `specification.md`'s "Out of scope" section requires.
3. It matches this repository's existing convention of separate,
   purpose-specific files (`accessibility.yml` vs `lighthouse.yml`) over one
   file serving every environment.

**Why a separate directory, not just a separate file under `tests/e2e/`:**
`apps/web/package.json`'s `test:e2e` script runs `playwright test` with the
default config, whose `testDir` is `./tests/e2e`. A staging spec placed there
would be collected by the PR-time accessibility/E2E run and fail immediately
(it requires a minted real-backend session, which that run has no way to
produce). `tests/staging-e2e/` keeps the two suites disjoint by construction
rather than by a `testIgnore` pattern a future edit could drop.

**Config placement, and the previous revision's blocking bug.** The earlier
attempt put the staging config at `apps/web/tests/e2e/playwright.staging.config.ts`
while keeping package-root-relative `testDir: "./tests/e2e"` and
`outputDir: "./test-results-staging"`. Playwright resolves both relative to the
**config file's own directory**, so they pointed at
`apps/web/tests/e2e/tests/e2e` (nonexistent) and
`apps/web/tests/e2e/test-results-staging` (not the path the workflow uploaded).
The config now sits beside `playwright.config.ts` at the package root, where
those relative paths mean what they read as. Verified rather than reasoned
about:

```
$ pnpm exec playwright test --config playwright.staging.config.ts --list
  [staging-desktop-1280] › core-loop.staging.spec.ts:168:3 › Core loop against real staging (VOC-050-T02) › ...
Total: 1 test in 1 file

$ pnpm exec playwright test --list        # the default, mock-backed suite
Total: 39 tests in 9 files                # unchanged; the staging spec is absent
```

## 2. Open question 2 — no test-only override cookies against the real backend

**Chosen: the option `specification.md` leaned toward.** The staging journey
uses no `e2e_onboarding_status` and no `e2e_unauthenticated` cookie. Nothing in
this task asks the real backend to honour a test-only override, so no new trust
surface is added to it.

Concretely:

- **Authenticated start:** the journey consumes a session minted by
  `VOC-050-T01`'s token-gated `POST /ops/synthetic-smoke-test/session` for the
  `VOC-050-T00`-seeded account, passed in as `E2E_SESSION_COOKIE` /
  `E2E_CSRF_TOKEN` and installed as the real `vocanova_session` /
  `vocanova_csrf` cookies.
- **Onboarding:** `apps/api/scripts/seed-synthetic-smoke-user.sql` seeds the
  account with `onboarding_status = 'completed'`, so the journey lands on
  `/home`. The onboarding form is still driven for real if the app redirects to
  `/onboarding` — a regression that unexpectedly resets onboarding fails a step
  instead of being masked by an override cookie.
- **Auth-gate rejection (step 10):** instead of forcing a 401 with
  `e2e_unauthenticated`, the journey relies on the logout it just performed
  having revoked the session server-side, then asserts the middleware redirects
  `/home` → `/signin?returnTo=%2Fhome`. This exercises the real gate reacting to
  a real 401.

## 3. Content-agnostic selectors, and rerun safety against a persistent account

`core-loop.spec.ts` can hard-code "Ordering at a cafe" and "pour" because the
mock serves fixed fixtures. Staging serves whatever canonical content is seeded,
and the synthetic account keeps its state between deploys, so the staging
journey selects the first situation, then prefers the first word in it the
account has **not** already saved (a newly-saved word has `next_review_at IS
NULL` and is therefore immediately due — see
`apps/api/business/reviews/postgres.go`'s `ListDueWords`, which is what keeps
the review, sentence-feedback, and progress steps exercised rather than skipped
on a typical run).

## 4. Journey coverage against `VOC-050-TEST-03`, including the two steps the previous revision omitted

| `VOC-050-TEST-03` step | Covered | How |
| --- | --- | --- |
| Onboarding-appropriate start | yes | seeded `completed`; real form driven if redirected |
| Discover | yes | `/discover` → first situation → a word detail page |
| Save | yes | asserts the save button's `aria-pressed` reaches `true` |
| Review | yes | works the due queue, handling both self-check and multiple-choice prompts |
| Sentence feedback | yes (conditional, see below) | real `POST /api/v1/sentence-feedback`, asserts HTTP 200 and a rendered outcome |
| Progress | yes (see below) | daily-mission counter re-read and compared against the pre-review baseline |
| Settings | yes | display-name write, confirmation, value round-trip |
| Logout | yes | real `POST /api/v1/auth/logout`, redirect to `/signin` |
| Auth-gate rejection | yes | post-logout `/home` → `/signin?returnTo=%2Fhome` |

Two honest limits on that coverage, both deliberate:

1. **The sentence-feedback step is reachable only when the review queue is
   emptied within `MAX_REVIEW_CARDS` (8).** The widget renders only in the
   caught-up state, for a card reviewed in that same session
   (`review-session.tsx` lines 187-203). A run that starts with a deeper
   backlog still reviews eight real cards and records a
   `skipped-step` annotation naming why, rather than failing or silently
   claiming coverage. Removing the bound would let an accumulated backlog
   consume the deploy budget.
2. **The verdict is not asserted, only that one was produced.** The staging
   evaluator is a real, non-deterministic provider. The step asserts the POST
   returns HTTP 200 (a 5xx, an auth break, or a routing break fails the deploy)
   and that the widget settled on a rendered outcome. Asserting a specific
   verdict would make a real deploy gate flake on provider variance.

The progress assertion compares against a baseline read before the review step
(`reviewedAfter >= reviewedBefore + reviewedCards`) rather than a fixed
"1 of 20", because the account's daily counter carries over between deploys on
the same day.

## 5. `VOC-050-AC-03` — fail-closed wiring in `deploy-staging.yml`

- New steps: mint session → set up Node/pnpm → install deps → install Chromium
  → run the journey → (on failure only) upload report and results.
- **No `continue-on-error` on any step in the job, and none at job level** —
  verified by parsing the committed YAML, not by reading it:

```
$ python3 -c "...yaml.safe_load('.github/workflows/deploy-staging.yml')..."
no continue-on-error anywhere
job-level continue-on-error: None
timeout: 40
```

- The mint step **exits non-zero** when `STAGING_SMOKE_TEST_SESSION_MINT_TOKEN`
  is unset, when the endpoint call fails (`curl -fsS`), or when the response
  carries an empty session or CSRF value. There is no path that skips the
  journey and reports success — the "unset secret" case is the one most likely
  to be quietly tolerated, so it is the case made loudest.
- Target: `https://staging.vocanova.site`, via `STAGING_WEB_BASE_URL`.
- `timeout-minutes` raised 30 → 40 to cover the added setup and run, with the
  file's existing budget comment updated to match.

**Placement, and the ordering finding from the previous review.** The steps sit
after *both* readiness polls rather than between them. `VOC-050-AC-03` asks for
the journey to run once the api `/healthz` poll passes, which it does; but the
journey drives the web tier, and starting it before `staging.vocanova.site` is
confirmed serving would report web cold-start as a core-loop failure. The
independent review raised exactly this risk. The workflow comment states the
reasoning in place so a later edit does not "restore" the fragile order.

## 6. Secret handling — minted credentials are masked before they are written

The minted `session_cookie` and `csrf_token` grant access to the synthetic
account until they expire, and GitHub step outputs are not secrets. Both values
are now registered with `::add-mask::` **before** being written to
`$GITHUB_OUTPUT`, so they cannot surface in this run's log or in a Playwright
trace uploaded on failure. They are passed to the test through the step's
`env:` block, matching the convention every other secret in this file already
uses. The previous revision wrote them to step outputs unmasked; that was the
review's Medium finding.

## 7. Risk classification

`change.yaml`'s `risk` was raised **R2 → R3** (never lowered) in this task.
`implement.yml` derives every task PR's `Risk classification:` line from that
field, and `governance-policy.yml` runs
`classify-change-risk.sh --require-declaration`, which fails closed when the
declared class is below the detected path floor. Measured floor for this task's
file list:

```
$ bash scripts/governance/classify-change-risk.sh --files-from <this task's files>
R3      .github/workflows/deploy-staging.yml
R1      apps/web/playwright.staging.config.ts
R1      apps/web/tests/staging-e2e/core-loop.staging.spec.ts
R1      apps/web/tests/tsconfig.json
R0      specs/.../t02-evidence.md
R1      specs/.../change.yaml
Detected path-based risk floor: R3
```

`T00`'s `apps/api/migrations` and `T01`'s `apps/api/business/auth` are R3 path
classes too, so R3 is also correct for the tasks still to come in this package.

## 8. Deterministic validation run against this working tree

| Command | Result |
| --- | --- |
| `pnpm --filter @vocanova/web typecheck:e2e` | exit 0 (after `pnpm --filter @vocanova/api-client build`, which the workspace requires first) |
| `pnpm --filter @vocanova/web lint` | exit 0 |
| `pnpm exec playwright test --config playwright.staging.config.ts --list` | 1 test in 1 file, discovered correctly |
| `pnpm exec playwright test --list` | 39 tests in 9 files — PR-time suite unchanged, staging spec absent |
| `bash scripts/governance/validate-governance.sh` | exit 0 |
| `bash scripts/governance/classify-change-risk.sh` | exit 0, floor R3 (§7) |
| YAML parse of `deploy-staging.yml` | parses; 23 steps; no `continue-on-error` (§5) |

`apps/web/tests/tsconfig.json` gained `../playwright.staging.config.ts` so the
new config is type-checked by the existing `typecheck:e2e` script rather than
being invisible to CI.

## 9. Limitations — what this revision does **not** prove

- **No live staging run exists for this revision.** The workflow change is not
  on `develop` until this PR merges, and this environment holds no staging
  secret, no `STAGING_SMOKE_TEST_SESSION_MINT_TOKEN`, and no route to
  `staging.vocanova.site`. `VOC-050-TEST-03`'s "passes against a real
  staging-shaped target" therefore remains outstanding and must be confirmed on
  the first post-merge staging deploy. Static verification (test discovery,
  typecheck, lint, YAML parse) is stronger than inspection but is not a live
  run.
- **Selector-level assumptions about seeded staging content are unverified**
  against the real staging database: that at least one journey situation
  exists, and that it contains at least one word. Both are implied by
  `VOC-050-T04`'s expectation that `GET /api/v1/journey-situations` returns 200
  for this account, but neither was observed here.
- **Operator prerequisite:** `STAGING_SMOKE_TEST_SESSION_MINT_TOKEN` must be
  provisioned as a repository secret **and** must equal the staging api
  container's `SMOKE_TEST_SESSION_MINT_TOKEN`. `VOC-050-T01` registers no mint
  route at all when the container-side value is unset. This was not verified
  from here; the first deploy after merge is where a mismatch shows up, and it
  shows up as a hard failure, which is the intended direction.
- **Cross-repo promotion gating is out of this task's scope** — whether
  `karsift-ai-infra`'s `release.yml` actually consults this workflow's
  conclusion is `VOC-050-T03`'s documented open item
  (`specification.md`'s open question 1).

## 10. Follow-ups noticed, deliberately not fixed here

1. **No deterministic way to put the synthetic account's SRS state into a known
   shape.** `ApplyReview` pushes `next_review_at` forward on every review
   (`apps/api/business/reviews/scheduling.go`), and re-saving a soft-deleted
   word restores it without clearing that timestamp
   (`apps/api/business/learning/postgres.go`), so "a card is due" cannot be
   guaranteed on every deploy. Handled here with rerun-safe branching; a real
   fix (a reset entry point scoped to the synthetic account) is new backend
   surface and belongs in its own package.
2. **The staging journey mutates staging data** (saves words, submits reviews,
   rewrites the display name) on every deploy. That is appropriate for staging
   and matches this task's scope; `VOC-050-T04`'s production check is
   separately constrained to read-only requests and must stay that way.
3. **`apps/web/eslint.config.mjs` ignores `playwright.config.ts` by name.** The
   new sibling config is not ignored and lints clean today, but the ignore
   entry is now inconsistent with the directory's contents. Left alone rather
   than widened, since narrowing lint coverage is not this task's business.
