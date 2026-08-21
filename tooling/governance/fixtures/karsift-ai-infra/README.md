# Pinned karsift-ai-infra contract fixtures (VOC-080-T05, VOC-097-T03)

These copies are deterministic fixtures for caller-repo policy regressions.
They mirror `KARSIFT/karsift-ai-infra` at the SHA in `PINNED_SHA.txt` so
`tooling/governance/tests/test_voc080_*.py` and
`tooling/governance/tests/test_voc097_*.py` can assert merge/adopt/release/
remediate/plan-review/live-evidence/role contracts without cloning the infra
repository in CI.

They are not a second runtime source of truth. Callers still `uses:`
`KARSIFT/karsift-ai-infra/...@main`. Update the fixtures when VOC-080- or
VOC-097-related infra contracts change and record the new pin in evidence.
