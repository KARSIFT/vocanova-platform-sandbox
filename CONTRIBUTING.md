# Contributing

Vocanova uses two permanent branches:

- `main` contains production-ready history.
- `develop` is the base branch for ongoing development.

Create working branches from the appropriate protected branch using these prefixes:

- `feature/` for new capabilities
- `fix/` for corrections
- `docs/` for documentation changes
- `refactor/` for behavior-preserving code changes
- `infra/` for infrastructure changes
- `security/` for security changes
- `hotfix/` for an approved emergency path

Use a stable `VOC-###` identifier in the branch name when one exists. Work in an
isolated branch or worktree and target `develop`; release pull requests promote
`develop` to `main`. Working branches are normally squash-merged. Release promotions
use an identifiable merge commit.

Meaningful changes require a linked approved requirement or decision, risk
classification, applicable tests, independent verification, and a pull request.
Follow the [autonomous development model](docs/governance/16-autonomous-development-operating-model.md)
and [risk classification](docs/governance/change-risk-classification.md).

The pull-request template provides two paths:

- `Standard` for behavioral, protected, or otherwise meaningful changes.
- `Lightweight R0` for non-behavioral, non-policy documentation and small maintenance
  changes. It still records objective, scope, risk, relevant checks, and verifier
  evidence, but irrelevant sections may be marked `N/A` with a reason.

Run every installed validation relevant to the change. After the frozen installation
described in the [local development guide](docs/development.md), application-
foundation changes run the applicable root commands, normally beginning with:

```bash
pnpm validate
pnpm audit
```

Governance validation remains independently required where applicable:

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Use the exact checked-in tool versions and scripts with a frozen lockfile. Do not claim
an unavailable tool or external deployment passed.

Under active A-004, R0-R4 engineering workflows advance through applicable
deterministic controls and independent verification without a founder `approved`
comment on merge, adoption, release, deploy, or retry. R4 retains strengthened
evidence, rollout, monitoring, and rollback obligations. Founder input remains
requirement clarification for genuinely ambiguous product, legal, or strategy
questions before stable acceptance criteria; EHR remains exceptional, not a routine
approval layer. The independent reviewer is never human authority. Repository
protections apply to contributors and automation actors alike; never bypass failed
checks, required review, branch protection, or production gates.

**Historical bootstrap:** the one-time initial DOC-16/A-002 adoption merged with
founder approval, independent Claude Code verification, and passing repository
validation. It did not mark steward approval satisfied or authorize production, and
the exception expired on merge. It is not a current approval or release gate.

VOC-002 was not a bootstrap exception. It was the completed one-time A-003 migration
governed by pre-A-003 R4 founder and R3 technical-steward approval bound to its exact
revision. That approval is exhausted and cannot be reused - VOC-002 itself grants no
standing automatic-merge or autonomous-production-release authority. This does not mean
those capabilities are disabled system-wide: automatic merge into `develop` is a
separately implemented and proven gate, and repository-controlled promotion plus
push-to-`main` production deployment are enabled when their non-founder gates pass.
Those controls derive from current A-004 authority and live infrastructure, not
VOC-002. See AGENTS.md's
"Change workflow" section for the current, accurate state of that gate and of
autonomous production release.
