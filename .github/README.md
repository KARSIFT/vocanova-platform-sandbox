# GitHub Configuration

This directory contains repository contribution and governance controls:

- `pull_request_template.md` records traceability, risk, evidence, impact, verification,
  and approvals with a lightweight R0 path.
- `ISSUE_TEMPLATE/` provides governed change intake and private security routing.
- `CODEOWNERS` uses the verified human repository identity for review routing. It is
  not approval evidence and does not create a standing authority under active A-004.
- `workflows/governance-policy.yml` validates the governance structure and prevents a
  pull request from declaring a risk below its changed-path floor.

The policy workflow is one of several automated checks now live - see
`docs/governance/repository-settings.md`'s "Current reality" section for what
actually runs today (application CI, independent review, staging and production
deployment) versus what's still genuinely unbuilt (per-PR Cloudflare previews,
one-click rollback automation). Treat that section, not this paragraph, as the
source of truth for current automation state - this file only describes what
lives in `.github/` itself. See
[`docs/governance/repository-settings.md`](../docs/governance/repository-settings.md)
for the required administrator settings and credentials.

A-004 is the active engineering-workflow authority. These files do not themselves
activate automation, but automatic merge into `develop`, repository-controlled
`develop` → `main` promotion, and push-to-`main` production deployment are implemented
and enabled through karsift-ai-infra when their gates pass. RL1/RL2 technical
activation remains disabled. See `docs/governance/repository-settings.md` and
`docs/governance/a004-transition-state.yaml` for current activation state.
