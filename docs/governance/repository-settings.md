# Repository and External Settings

Files in this repository describe policy but cannot enable GitHub organization
settings, create Cloudflare projects, or provision credentials. A repository
administrator must configure and record the following before autonomous merge or
release is enabled.

A-003 governance authority is active. That activation must not be represented as
hosted or technical activation. Automatic merge into `develop` (A-003 §10) is
implemented, tested, and proven - live since VOC-012 (see
`docs/governance/a003-transition-state.yaml`). RL1/RL2 technical activation,
production deployment, and autonomous production release (A-003 §11/12, a distinct,
narrower gate than develop-merge) remain disabled or unimplemented. *(Corrected
2026-07-24 - previously conflated develop-merge with production-release authority.)*

## GitHub rulesets

Configure `develop`:

- require pull requests and block direct pushes, force pushes, and branch deletion;
- require `policy / governance-policy` and every installed application CI check;
- require the independent Claude Code verifier status check;
- require conversation resolution and dismiss stale approvals;
- require code-owner review for protected paths;
- allow squash merge; and
- restrict bypass to a small human incident-administrator group with audited use.

Configure a non-self-referential R4 control for these exact paths:

```text
/.github/workflows/governance-policy.yml
/.github/CODEOWNERS
/scripts/governance/
/docs/operations/15-ai-native-product-and-engineering-operating-model.md
/docs/governance/approval-matrix.md
/docs/governance/change-risk-classification.md
/docs/governance/protected-areas.md
/docs/governance/post-merge-activation-checklist.md
/docs/governance/amendments/
/docs/governance/a003-transition-state.yaml
/docs/governance/16-autonomous-development-operating-model.md
/docs/architecture/17-autonomous-development-architecture.md
/docs/planning/18-autonomous-development-implementation-roadmap.md
/specs/changes/VOC-002-a003-governance-transition/
/specs/changes/VOC-003-a003-lifecycle-sync/
/specs/changes/VOC-004-canonical-adoption-doc-17-doc-18/
```

Under active A-003 until A-004 activation, that ruleset must continue non-self-referential
verification and R4 path-floor enforcement but must not impose routine standing steward
or founder approval merely because a change is R3. **After A-004 activation**, R4
engineering-workflow gates do not require founder click-approve on merge/adopt/release;
strengthened evidence and independent verification remain required. Where the GitHub
plan supports organization-required workflows, run the policy gate from a separately
protected default-branch or organization source. A status name produced solely by a workflow that the same pull
request can rewrite is not sufficient protection.

Configure `main`:

- include all `develop` protections;
- accept only release pull requests from `develop` or the documented emergency path;
- require release, staging, migration, rollback, and health-check gates;
- enforce strengthened R3 gates without a standing steward/founder requirement solely
  for R3; **post-A-004:** no founder environment-reviewer requirement on the
  repository-controlled release/deploy path when promotion checks pass;
- use merge commits for release promotion; and
- prevent an AI or release-bot identity from bypassing required non-founder checks.

GitHub cannot natively express every conditional R0-R4 gate combination using
CODEOWNERS alone. Use separate protected teams/environments and a reviewed gate that
validates the effective risk class and attributable evidence. Keep autonomous merge
disabled until that gate is tested.

Multiple owners on one CODEOWNERS pattern are alternatives: one matching owner can
satisfy GitHub's native code-owner review requirement. They do not mean that every
listed owner must approve. Under active A-003 until A-004 activation, enforce R4
founder authority on merge (historical). **Post-A-004:** enforce strengthened R3/R4
evidence gates independently; do not recreate a combined standing founder-comment
merge requirement.

Enable repository security settings when available:

- secret scanning and push protection;
- Dependabot alerts and security updates after a dependency manifest exists;
- private vulnerability reporting;
- Actions restricted to reviewed, immutable action SHAs; and
- minimal default workflow token permissions.

The verified human repository identity `@m-e-h-r-d-a-a-d` is formally recorded as
founder and as the historical pre-A-003 qualified human technical steward in
[technical-steward-appointment.md](technical-steward-appointment.md). Direct account
routing is currently used. A steward team is not a prerequisite to A-003 activation
and must not be created as a replacement permanent authority. Direct routing may
remain for review routing, but never proves conditional approval. Never use an AI or
bot identity as human authority.

## Required identities and credentials

- Distinct implementer-role and independent-reviewer-role identities (currently both
  Cursor-backed, per karsift-ai-infra's `config/roles.yml` - configurable, not a
  permanent Codex/Claude assignment).
- The recorded human founder identity and preserved historical technical-steward
  identity; no replacement standing steward team is required.
- GitHub App or OIDC-based credentials with least privilege and short expiry.
- Separate Cloudflare preview, staging, and production projects/accounts or clearly
  isolated environments.
- Environment-scoped Cloudflare tokens; production credentials are unavailable to
  pull-request and implementation-agent contexts.

No credential value belongs in the repository.

## Cloudflare and release configuration

Before deployment automation is added, record and validate:

- approved build and deploy commands from the future package scripts;
- preview, staging, and production project identifiers and domains;
- environment bindings, data stores, migration order, and secret isolation;
- preview cleanup behavior and access restrictions;
- staging and production smoke/health endpoints;
- monitoring alerts, responsible responder, and evidence retention;
- last-known-good artifact or commit redeployment procedure;
- feature-flag or traffic-shift rollback where relevant;
- database backup, restore test, recovery point objective, and recovery time objective;
  and
- production environment approval rules matching R3/R4.

## Current blockers (rewritten 2026-07-24 - the paragraph below was written during the
original repository bootstrap, before any application code existed, and had never been
updated since; it's preserved as history for the paragraph after it, which is still
accurate)

*Historical, no longer true:* "The repository currently has no application, package
manifest, pnpm lockfile, workspace, test/build scripts... only the dependency-free
governance policy check can run today; application CI, previews, staging, production,
and rollback cannot truthfully be automated yet."

**Current reality:** `apps/web` and `apps/api` are real, working applications;
`package.json`/`pnpm-lock.yaml`/the pnpm workspace exist; deterministic CI (format,
lint, typecheck, test, build) runs on every PR and has passed across 22+ shipped
packages (VOC-010 through VOC-022 at minimum) via `karsift-ai-infra`'s `ci.yml`.
`docs/migration-manifest.yaml` and `docs/document-graph.yaml` were migrated
(VOC-007/VOC-008) and later archived to `docs/archive/` as historical evidence
trails (2026-07-24) - they are available, just not filed as live/current
documentation. Verified sources for DOC-00 through DOC-13 are canonical and adopted;
DOC-14 was deliberately reconciled but not adopted (see `docs/README.md`'s index).

**Updated again (2026-08-08):** staging and production deployment are no longer
unbuilt - `deploy-staging.yml` and `deploy-production.yml` exist, have both run
successfully many times against real infrastructure (a real server plus
vocanova.site/Cloudflare DNS), and production deploys are now restricted to `main`
only via the `production` environment's branch policy. Rollback is a manual,
proven procedure (redeploy the previous immutable image digest - DOC-11 §3), not
one-click automation. Per-PR Cloudflare previews genuinely remain unbuilt - that
part of the original blockers list still holds for previews specifically, not for
staging/production deployment anymore.

The initial governance bootstrap merged through PR #3 and its one-time exception has
expired. The historical technical-steward appointment and completed dual-capacity
VOC-002 approval remain permanent evidence, but the role is retired as routine R3
authority and that migration approval cannot be reused. Under active A-003 (until `VOC-080-T07`), routine R3 uses strengthened technical
gates and independent verification. **Historical (A-003 / VOC-075):** R4 required
a founder `approved` comment on engineering-workflow merge gates. Automatic merge
into `develop` is live (see above); RL1/RL2 technical activation remain disabled
until separately implemented, tested, and proven. **Post-A-004 activation**
removes founder-comment gates on engineering-workflow merge/release/deploy paths
at every risk class including R4; see `a004-transition-state.yaml` and issue #627.
