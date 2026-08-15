# Pinned karsift-ai-infra contract fixtures (VOC-080-T05)

These copies are deterministic fixtures for caller-repo policy regressions.
They mirror `KARSIFT/karsift-ai-infra` at the SHA in `PINNED_SHA.txt` so
`tooling/governance/tests/test_voc080_*.py` can assert merge/adopt/release/
remediate/plan-review/role contracts without cloning the infra repository in
CI.

They are not a second runtime source of truth. Callers still `uses:`
`KARSIFT/karsift-ai-infra/...@main`. Update the fixtures when VOC-080-related
infra contracts change and record the new pin in evidence.
