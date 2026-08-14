# VOC-078 — VOC-076-T02: No Real Staging Proof the Review-Queue Button Fix Works: Specification

## Objective and requirement source

Close the verification gap reported in
[GitHub issue #608](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/608):
VOC-076-T02's PR
[#598](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/598) merged
to `develop` despite an independent-verification `VERDICT: FAIL` because there
was **no real staging proof** that step 5 of the staging core-loop journey
completes past the multiple-choice review-queue click on the fixed revision.

Primary evidence (issue #608 and VOC-076 `t02-evidence.md`):

| Item | Value |
|------|--------|
| Predecessor | VOC-076 T00/T01 passed; T02 PR #598 merged with FAIL |
| Cited failing staging run | [#227](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31791701520) (T01 tip `d305632`) |
| Failure | `expect(…).toBeEnabled()` on MC option — stayed `disabled` / `aria-pressed="false"` for 20s |
| Gap fix landed via #598 | `shouldShowReviewCardPrompt` busy state during `isRefetching`; E2E prompt-ready waits (`disabled: false`) — on `develop` as of drafting |
| Original bug issue | [#575](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/575) — still open |

**Objective:** after this package's implementation, (a) a real
`deploy-staging.yml` run of `tests/staging-e2e/core-loop.staging.spec.ts`
against the fixed `develop` tip completes step 5 past the MC answer
interaction without the prior disabled-button failure mode, (b) that PASS is
recorded as evidence for VOC-076-AC-03 / this package's acceptance criteria,
and (c) issue #575 is closed only once that PASS is genuine.

## Confirmed findings (from issue #608 and drafting-time repo read)

- VOC-076-T02 never produced a green staging run URL that satisfies AC-03;
  `t02-evidence.md` explicitly marks AC-03 unmet / pending post-merge
  re-verification.
- PR #598 merged anyway with `VERDICT: FAIL` (High: missing staging proof).
  That merge defect is a process concern (`VOC-078-DEP-01`); this package's
  product obligation is the missing proof (and remediation if still broken).
- Drafting-time tree contains the #598 gap fix
  (`shouldShowReviewCardPrompt`, "Loading next reviews…", E2E
  `disabled: false` prompt-ready waits). Tip merge
  `26d85c1` is PR #598. Whether a **post-merge** `deploy-staging` run is
  green is **not** confirmed at drafting — `VOC-078-DEP-02` / T00 must settle
  it with a concrete run URL.
- A local/mock suite or evidence prose without a green Actions run does not
  close #575 or VOC-076-AC-03.

## Scope and non-goals

In scope:

- `VOC-078-T00`: Identify the `develop` tip under test (must include VOC-076
  T01 + #598 gap fix). Locate or obtain a real `deploy-staging.yml` run of
  `core-loop.staging.spec.ts` against that tip (or a later tip that still
  contains the fix). Record PASS or FAIL honestly in this package's
  `t00-evidence.md`. On PASS: update VOC-076 `t02-evidence.md` and
  `VOC-076-AC-03` Result; do not close #575 until PASS is recorded.
- `VOC-078-T01`: Only if T00 recorded FAIL — diagnose the remaining cause
  against the new run (without re-deriving VOC-076-T00 from scratch unless
  contradicted), apply the narrowest product and/or E2E fix, add regression
  coverage as needed, then obtain and record a green staging run. If T00
  PASSed, mark T01 N/A with citation — no source change.

Non-goals / explicitly excluded:

- Not redispatches of VOC-076-T02 under exhausted attempt budget (fresh
  package/tasks instead — this package).
- Not weakening VOC-050's staging gate or VOC-074's `reviewedCards >= 1`
  hardening.
- Not assumed `deploy-staging.yml` edits (R3 path floor if expanded).
- Not force-merging or inventing green run URLs.
- Not closing #575 on evidence that still says pending/FAIL.
- Not pipeline/merge-gate hardening for FAIL merges unless adoption expands
  `VOC-078-DEP-01`.
- Not adopting, authorizing, implementing, or merging this package from
  within the draft itself.

## Risk and protected areas

Builder assessment: default happy path is `specs/changes/` evidence and
VOC-076 AC Result updates (path floor **R1**). Conditional remediation
touches `apps/web` review-session / staging-e2e (also path floor **R1**).
`deploy-staging.yml` is out of default scope (**R3** if included).

This package proposes **R2** because the staging core-loop gate is still
unproven green for the original #575 failure mode and a residual product
hang may affect real learners if T01 must change review-session again
(VOC-076 precedent). The independent verifier must re-run
`classify-change-risk.sh` against the real task file list and may raise or
lower per T00 outcome and whether the workflow file is touched.

No governance, secret-handling, or migration area is in default scope. EHR
is not triggered. Under active A-003, routine R3 (if the path floor rises)
does not require standing technical-steward or founder approval merely for
being R3; strengthened verification still applies.

## Decisions, contradictions, security, and privacy

`VOC-078-D00` (recorded here for traceability; formal decision numbering
applies after adoption): A green local or unit suite does not satisfy
VOC-076-AC-03 or close issue #575. Closure requires a real
`deploy-staging.yml` run URL where step 5 completes past the MC click (with
MC coverage per VOC-076-AC-03's rule) on a revision that includes the
VOC-076 fix tip. Independent review must FAIL any task PR that claims PASS
without that URL.

No contradiction with VOC-076's cause analysis: this package assumes T00/T01
findings and the #598 gap fix remain the baseline, and only re-opens product
work if a new staging run still fails.

Open questions for the reviewing human:

1. **`VOC-078-DEP-00` — VOC-076-T02 disposition.** Recommended default: once
   VOC-078 records real staging PASS, update VOC-076 `t02-evidence.md` /
   AC-03 Result and treat VOC-076-AC-03 as satisfied by VOC-078 evidence; do
   not redispatch VOC-076-T02. Confirm at adoption.
2. **`VOC-078-DEP-01` — FAIL-merge process follow-up.** Recommended default:
   out of this package; open a separate unlabeled issue if merge-gate /
   human override of FAIL needs hardening. Confirm at adoption.
3. **`VOC-078-DEP-02` / risk.** Accept proposed R2, or adopt as R1 if scoped
   evidence-only with T01 explicitly deferred to a follow-up package on
   FAIL.

Security / privacy: staging verification continues to use only the existing
synthetic smoke-test account (VOC-050). No new secrets or personal data.

## Data, migrations, analytics, and accessibility

- **Data / migrations:** None expected.
- **Analytics:** None expected.
- **Accessibility:** If T01 changes review-session disabled/busy timing,
  preserve fieldset/legend, `aria-pressed`, focus-visible, and the
  `max-w-[28rem]` workaround (`.karsift/lessons.md`). Evidence-only path:
  explicit non-applicability.
