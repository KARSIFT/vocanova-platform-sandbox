# VOC-078 — Acceptance Criteria

## VOC-078-AC-00 — Real staging run proves step 5 completes past the MC click

- Requirement source: issue #608; issue #575; `VOC-078-D00`; VOC-076-AC-03
- Tasks: `VOC-078-T00` (and `VOC-078-T01` if T00 failed)
- Tests: `VOC-078-TEST-00`, `VOC-078-TEST-02`
- Evidence: `VOC-078-EV-00` (and `VOC-078-EV-01` if remediation ran)
- Result: pending

Observable outcome:

1. A real `deploy-staging.yml` run URL is recorded against a `develop` tip
   that includes VOC-076 T01 and the PR #598 gap fix (or a later tip that
   still contains that fix, plus any VOC-078-T01 remediation).
2. That run executes `tests/staging-e2e/core-loop.staging.spec.ts` and
   completes step 5 ("work the review queue") without the prior
   disabled-button failure mode (neither the original 240s click timeout nor
   the run #227 `toBeEnabled` hang on a permanently disabled MC option).
3. MC coverage follows VOC-076-AC-03: prefer a run that exercises at least
   one multiple-choice card; if only self-check appears, evidence must show
   a subsequent MC-exercising run or an explicit forced-MC reproduction under
   the fixed revision — do not claim #575 closed on MC-disable without MC
   coverage.
4. Invented or missing run URLs fail this criterion. A FAIL run may be
   recorded honestly for T00 but does **not** satisfy this criterion until a
   later green run is recorded (via T01).

## VOC-078-AC-01 — VOC-076 package record matches the real staging outcome

- Requirement source: issue #608; `VOC-078-DEP-00`
- Tasks: `VOC-078-T00` (and `VOC-078-T01` if needed)
- Tests: `VOC-078-TEST-01`
- Evidence: `VOC-078-EV-00` / `VOC-078-EV-01`
- Result: pending

Observable outcome:

1. `specs/changes/VOC-076-…/t02-evidence.md` is updated so AC-03 is marked
   satisfied **only** when AC-00's green run exists; it must not claim PASS
   while citing only run #227 or "pending post-merge re-verification."
2. `VOC-076-AC-03` Result in VOC-076 `acceptance-criteria.md` is aligned
   with that evidence (satisfied when green; left unmet/pending if still
   failing — never "satisfied" on FAIL).
3. Historical FAIL narrative for run #227 may remain as baseline comparison;
   it must not be rewritten as a PASS.

## VOC-078-AC-02 — Issue #575 closed only after genuine staging PASS

- Requirement source: issue #608 suggested fix; issue #575
- Tasks: `VOC-078-T00` / `VOC-078-T01`
- Tests: `VOC-078-TEST-00`
- Evidence: `VOC-078-EV-00` / `VOC-078-EV-01`
- Result: pending

Observable outcome: GitHub issue #575 is closed only after AC-00 is met.
Closing #575 while staging still fails, or while evidence still says
pending/FAIL, fails this criterion. If T00 fails and T01 has not yet
produced PASS, #575 stays open.

## VOC-078-AC-03 — Package boundaries and honesty constraints respected

- Requirement source: issue #608; `specification.md` non-goals
- Tasks: `VOC-078-T00`, `VOC-078-T01`
- Tests: `VOC-078-TEST-03`
- Evidence: `VOC-078-EV-00`, `VOC-078-EV-01`
- Result: pending

Observable outcome:

1. No weakening of VOC-050's staging gate or VOC-074's `reviewedCards >= 1`
   hardening.
2. No silent expansion into `deploy-staging.yml` without adoption /
   `VOC-078` scope expansion (R3 path floor).
3. T01 source changes occur only when T00 recorded FAIL (or adoption
   explicitly ordered a preemptive fix); T01 N/A path makes no product edit.
4. Independent review of the exact revision reports `VERDICT: PASS` or
   `PASS WITH NON-BLOCKING FINDINGS` — a claim of staging PASS without a
   green run URL is a High/FAIL finding (same class as PR #598).
5. Diff contains no secrets or production credentials.
