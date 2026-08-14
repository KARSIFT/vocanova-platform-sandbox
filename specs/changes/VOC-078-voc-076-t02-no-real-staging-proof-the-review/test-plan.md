# VOC-078 — Test Plan

## VOC-078-TEST-00 — Real staging core-loop step 5 / MC proof

- Covers: `VOC-078-AC-00`, `VOC-078-AC-02`
- Preconditions: Package adopted; `develop` tip includes VOC-076 T01 + #598
  gap fix (or later remediation); ability to read Actions
  `deploy-staging.yml` run logs.
- Procedure:
  1. Open the recorded `deploy-staging` run URL from `t00-evidence.md`
     (or `t01-evidence.md` if remediation ran).
  2. Confirm the run executed `tests/staging-e2e/core-loop.staging.spec.ts`.
  3. Confirm step 5 completed without the disabled-button failure mode
     (no 240s MC click timeout; no run #227-style `toBeEnabled` hang on a
     permanently disabled option).
  4. Confirm MC coverage per VOC-076-AC-03 (or the explicit exception path).
  5. Confirm issue #575 is closed only if steps 3–4 pass; otherwise still
     open.
- Expected result: Green staging proof with MC coverage rule satisfied;
  #575 closed iff PASS.
- Evidence: `VOC-078-EV-00` / `VOC-078-EV-01`

## VOC-078-TEST-01 — VOC-076 evidence and AC-03 Result alignment

- Covers: `VOC-078-AC-01`
- Preconditions: VOC-076 `t02-evidence.md` and `acceptance-criteria.md`
  edited in the same PR tip that claims PASS (or left unmet on FAIL).
- Procedure:
  1. Read VOC-076 `t02-evidence.md`.
  2. If claiming PASS: assert a green run URL is present and AC-03 is marked
     satisfied; assert run #227 is not rewritten as PASS.
  3. If still FAIL/pending: assert AC-03 is not marked satisfied.
  4. Confirm `VOC-076-AC-03` Result matches the evidence file.
- Expected result: No contradictory "PASS without URL" or "satisfied on
  FAIL" state.
- Evidence: `VOC-078-EV-00` / `VOC-078-EV-01`

## VOC-078-TEST-02 — Remediation regression (FAIL path only)

- Covers: `VOC-078-AC-00` (after T01)
- Preconditions: T00 recorded FAIL; T01 produced a product/E2E diff.
- Procedure:
  1. Run applicable web tests for touched files (at minimum
     `review-session-prompt-readiness` tests if product changed; typecheck /
     lint / e2e typecheck as touched).
  2. Review diff: intentional disabled states preserved; VOC-074
     `reviewedCards >= 1` intact; `max-w-[28rem]` workaround preserved if
     `review-session.tsx` touched.
  3. Confirm a **new** green staging run URL (distinct from the T00 FAIL
     run) is recorded.
- Expected result: Local regression green; new staging PASS recorded.
  If T01 is N/A (T00 PASS), record N/A with citation — do not invent
  remediation.
- Evidence: `VOC-078-EV-01`

## VOC-078-TEST-03 — Boundaries, review PASS, governance checks

- Covers: `VOC-078-AC-03`
- Preconditions: Full PR diff available.
- Procedure:
  1. Assert file list stays within package scope (VOC-076 evidence/AC Result;
     this package's evidence; conditional `apps/web` review/staging-e2e /
     readiness tests). No unauthorized `deploy-staging.yml` edit.
  2. Confirm independent-review comment on the tip is PASS or PASS WITH
     NON-BLOCKING FINDINGS — and that a staging-PASS claim without a green
     URL is treated as blocking if present.
  3. Run:

     ```bash
     bash scripts/governance/validate-governance.sh
     bash scripts/governance/classify-change-risk.sh
     git diff --check
     ```

     For `apps/web` remediation, also run the applicable commands from
     `docs/development.md` (do not invent unavailable checks as passing).
- Expected result: Validation passes; declared risk meets or exceeds
  detected floor; scope and honesty constraints hold; review PASS.
- Evidence: `VOC-078-EV-00` / `VOC-078-EV-01`

## Rollback coverage

Rolling back means reverting the task PR commit(s) (evidence updates and any
T01 product/E2E changes). Validation: governance scripts pass on the
reverted tree; applicable web tests pass if product was touched. Staging may
return to the pre-VOC-078 unverified or failing state — that is the known
pre-package condition for AC-03.

## Constraints

No test in this plan uses secrets or production user data.
Staging runs use only the existing synthetic smoke-test account already
provisioned for staging E2E by VOC-050. Missing credentials or a missing
staging run must not be treated as PASS (CLAUDE.md).
