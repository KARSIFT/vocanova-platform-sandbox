# VOC-038 — Acceptance Criteria

## VOC-038-AC-00 — Cohort/allowlist mechanism decided and documented

- Requirement source: `VOC-038-D00`
- Tasks: `VOC-038-T00`
- Result: pending
- Observable outcome: A decision record names the chosen allowlist mechanism and initial
  cohort composition, with explicit founder approval.

## VOC-038-AC-01 — Cohort/allowlist mechanism implemented

- Requirement source: `VOC-038-D00`'s accepted decision
- Tasks: `VOC-038-T01`
- Result: pending
- Observable outcome: A non-allowlisted account cannot sign up or log in while
  `NEW_USER_SIGNUP_ENABLED=false`; an allowlisted account can, verified live against production.

## VOC-038-AC-02 — Production smoke-test suite exists and passes

- Requirement source: DOC-12 §5
- Tasks: `VOC-038-T02`
- Result: pending
- Observable outcome: A scripted smoke-test suite runs against production, checks health,
  auth reachability, kill-switch state, and the core-loop happy path, and currently passes.

## VOC-038-AC-03 — Non-AI core loop validated with the real allowlisted cohort

- Requirement source: DOC-12 §5
- Tasks: `VOC-038-T03`
- Result: pending
- Observable outcome: At least one real allowlisted user (not a disposable test row) completes
  auth → discover/save → review → mission/progress with no critical defect, `AI_FEATURES_ENABLED=false`.

## VOC-038-AC-04 — AI feedback enabled and observed for the cohort

- Requirement source: DOC-12 §5
- Tasks: `VOC-038-T04`
- Result: pending
- Observable outcome: `AI_FEATURES_ENABLED=true` in production; at least one real cohort user
  receives real AI sentence feedback; Sentry/Kuma show no launch-blocking incident during the
  observation window.

## VOC-038-AC-05 — Expansion thresholds decided and documented

- Requirement source: `VOC-038-D01`
- Tasks: `VOC-038-T05`
- Result: pending
- Observable outcome: A decision record states the quantitative thresholds gating expansion
  beyond the initial cohort, with explicit founder approval.

## VOC-038-AC-06 — Rollback proven with real cohort/AI state present

- Requirement source: DOC-12 §5
- Tasks: `VOC-038-T06`
- Result: pending
- Observable outcome: A genuine rollback rehearsal completes against production with the
  cohort mechanism and AI enabled, with no data loss or corruption for real cohort users.

## VOC-038-AC-07 — L1 gate passes and founder records hold/expand

- Requirement source: DOC-12 §5
- Tasks: `VOC-038-T07`
- Result: pending
- Observable outcome: The L1 release PR/issue shows all applicable checks passing, and a
  founder-authored hold/expand record exists, stating explicitly whether L1 remains at its
  current cohort, expands, or pauses, and what (if anything) remains as a tracked follow-up.
