# Pinned karsift-ai-infra contract fixtures (VOC-080-T05, VOC-097-T03, VOC-102-T00)

These copies are deterministic fixtures for caller-repo policy regressions.
They mirror `KARSIFT/karsift-ai-infra` at the SHA in `PINNED_SHA.txt` so
`tooling/governance/tests/test_voc080_*.py`,
`tooling/governance/tests/test_voc097_*.py`, and
`tooling/governance/tests/test_auto_advance_ownership.py` can assert
merge/adopt/release/remediate/plan-review/live-evidence/auto-advance/role
contracts without cloning the infra repository in CI.

They are not a second runtime source of truth. Callers still `uses:`
`KARSIFT/karsift-ai-infra/...@main`. Update the fixtures when VOC-080-,
VOC-097-, or VOC-102-related infra contracts change and record the new pin in
evidence.

## auto-advance ownership (VOC-102)

Adoption starts the first task automatically. The adopted roster records an explicit
`depends_on` edge from every later task to its predecessor, and `auto-advance.yml` releases the
next task only after the preceding task's implementation PR merges and its tracking issue closes.
For ordinary implementation tasks it dispatches `implement.yml` attempt 1. When the next roster
task is operator-owned or live-actions-only (declared in
`<package>/.karsift/live-evidence/<task_id>.yaml`), auto-advance instead prepares a deterministic
draft evidence-carrier PR and posts a sanitized waiting marker without starting the implementer;
the existing live-evidence reconciler remains the operator path. Implementer jobs are serialized per change package.
