# Post-Merge Governance Activation Checklist

Status: Partially activated (repository-controlled merge/release/deploy enabled; RL1/RL2 technical activation disabled)

**Current state (2026-08-30, VOC-140):** A-004 is the active engineering-workflow authority model.
Automatic merge, the repository-controlled `develop` → `main` promotion path, and push-to-`main`
production deployment are enabled when promotion checks pass. Recovery never treats a still-running
release carrier as attestable `ci / ci`, and production merge-guard verification uses a separate
Administration-only guard token immediately before the unchanged mutation-token merge. RL1/RL2
technical activation remain disabled.

**Historical (pre-2026-08-08):** A-003 became effectively active at `2026-07-17T16:44:34Z`, but
automatic/autonomous merge, RL1/RL2 technical activation, production deployment, and autonomous
production release were still disabled or unimplemented until the 2026-08-08 delegation.

**Updated 2026-08-08**: for vocanova-platform-sandbox specifically, automatic/autonomous
merge, production deployment, and autonomous production release are no longer
disabled - the founder explicitly, twice-confirmed-live authorized all three (see
`AGENTS.md`'s "Release and deployment authority" section). RL1/RL2 technical
activation remain disabled; that authorization was not part of this request. The
sentence above is preserved as historically accurate for the period before this
date.

## A-003 adoption and effective-activation boundary

A-003 adoption and activation are completed historical facts:

- [x] Approved VOC-002 PR head SHA:
      `c858ebff3d97da88fea830bc32a74f69f59a9ad2`.
- [x] Distinct adopted `develop` SHA:
      `9d5b4bc1d4a72e313b013047601265ee837c34f2`.
- [x] Exact-SHA independent verification and one-time dual-capacity approval existed
      before adoption. Evidence: PR #8 comments `5005293621` and `5005389067`.
- [x] Deterministic post-merge validation passed on the adopted state. Evidence:
      https://github.com/KARSIFT/vocanova-platform/actions/runs/29597154713
- [x] Effective activation was recorded and the migration approval became exhausted
      and non-reusable. Evidence:
      https://github.com/KARSIFT/vocanova-platform/pull/8#issuecomment-5005456622
- [x] VOC-003 synchronizes canonical lifecycle fields without changing frozen A-003
      substantive policy.

Under active A-004, routine R3 does not require a standing
technical steward or founder approval merely because it is R3. **Historical (A-003 /
VOC-075):** R4 merge required founder approval before activation. **A-004 (active,
issue #627):** engineering-workflow gates require no founder `approved` comment; R4
remains a strengthened evidence class. EHR remains exceptional-only.

This checklist begins after the initial DOC-16/A-002 governance pull request merges.
The bootstrap exception expires on that merge. No unchecked item is implied to be
complete, and no production or autonomous-release authority is granted by the merge.
Record an evidence link, accountable human, and completion date for every item.

## Human authority and GitHub identities

- [x] Appoint a qualified, accountable human technical steward and record the scope.
      Evidence: [technical-steward-appointment.md](technical-steward-appointment.md)
      and the required final dual-capacity approval bound to the exact head revision
      of the appointment pull request before merge.
- [x] Preserve direct review routing without treating it as approval evidence. Do not
      create a replacement standing technical-steward team.
- [ ] Verify that no founder, steward, Codex, Claude, or automation identity
      placeholder remains in executable repository controls.
- [ ] Configure distinct, least-privilege Codex implementation and Claude Code
      independent-verifier identities. Neither may be represented as founder,
      pre-A-003 steward, or EHR-qualified human authority.

## Repository enforcement

- [ ] Protect `develop`: pull requests only, required policy/application checks,
      independent verification, stale-approval dismissal, conversation resolution,
      code-owner review, and no unaudited bypass.
- [ ] Protect `main`: release pull requests only, required release gates, no direct or
      force pushes, and conditional R3/R4 approvals.
- [ ] Configure a non-self-referential governance ruleset for the fixed R4 paths in
      [repository-settings.md](repository-settings.md).
- [ ] Test strengthened R3 gates without routine steward/founder approval and
      post-A-004 no-founder-comment merge/release behavior before enabling any
      technical automation.

## Engineering and deployment gates

- [ ] Add frozen pnpm installation and the real format, lint, type, unit, integration,
      build, security, migration, and accessibility commands when application code
      and package scripts exist.
- [ ] Configure isolated Cloudflare preview, staging, and production projects,
      credentials, environment bindings, and access boundaries.
- [ ] Configure monitoring, health checks, evidence retention, incident ownership,
      last-known-good deployment, and tested rollback/recovery.
- [ ] Configure preview/staging status and protected production release gates.

## Activation rehearsal

- [ ] Rehearse the full release gate in a non-production environment, including a
      forced failure and rollback.
- [ ] Independently verify ruleset behavior, identity separation, secret isolation,
      monitoring, and rollback evidence.
- [ ] Separately record technical authorization for any autonomous-release class only
      after every applicable prerequisite is implemented, tested, and proven. A-003
      policy permission alone is not technical activation.

Until every unchecked activation item below is evidenced, treat the remaining
checklist gaps as open operational risk. Repository-controlled automatic merge,
promotion, and production deployment are enabled under active A-004 when gates pass;
RL1/RL2 technical activation remain disabled.

**Updated 2026-08-08**: for vocanova-platform-sandbox, the founder explicitly
authorized automatic production release (see `AGENTS.md`'s "Release and deployment
authority") notwithstanding this checklist not being fully evidenced - the
"Activation rehearsal" items above (forced-failure/rollback rehearsal, independent
ruleset/identity/secret-isolation verification) genuinely have not all been done.
This was a deliberate founder decision to accept that gap, not a claim that the
checklist passed. Production rollback/migration safety remains a real, separately
tracked risk - see the founder's own conversation record and any linked follow-up
work for current status, not this checklist's own unchecked boxes.
