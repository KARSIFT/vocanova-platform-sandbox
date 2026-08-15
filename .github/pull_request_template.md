## Objective and traceability

- Business/product objective:
- Approved Vocanova requirement or decision:
- Linked issue/specification (`VOC-###` where applicable):
- Change-package status and canonical path:
- Requirement source:
- Stable acceptance-criteria mapping:
- Implementation tasks/commits:

Change mode: <!-- Standard | Lightweight R0 -->
Risk classification: <!-- Replace with exactly R0, R1, R2, R3, or R4 -->

Use `Lightweight R0` only for a non-behavioral, non-policy documentation or small
maintenance change. For that path, complete the objective, scope, risk, checks, and
verifier result; mark genuinely irrelevant standard sections `N/A` with one reason.

## Summary and scope

- What changed:
- In scope:
- Out of scope:

## Existing-file reconciliation

| Path | Classification | Preserved content | Reconciliation |
|---|---|---|---|
| path | present-compatible / present-needs-reconciliation / absent-approved-to-create / material-conflict |  |  |

- Previous governance control:
- Proposed governance control:
- Active authority model (`A-004 active` since `VOC-080-T07`):
- Governance lifecycle impact (`none` or direction/approval/adoption/activation/sync):

## Risk and approvals

- Risk rationale:
- CI-detected risk floor:
- Affected protected areas (or `None`):
- Required approval class:
  - [ ] R0-R2 — independent verifier and applicable gates
  - [ ] R3 — strengthened applicable controls and independent verification; no
        standing steward/founder approval solely because work is R3
  - [ ] R4 — strengthened evidence, validation, verification, rollout, monitoring,
        rollback; **under active A-004:** no founder `approved` comment on engineering-workflow
        gates when non-founder gates pass
  - Historical VOC-002 migration — exhausted and permanently non-reusable
  - Historical initial DOC-16/A-002 bootstrap — expired with PR #3 and unavailable
    to later changes; no checkbox or waiver exists
- Exceptional-human-review evidence or `N/A — no EHR trigger`:
- Founder requirement-clarification evidence or `N/A — stable AC from approved source`:
- Exact reviewed head SHA:
- Adopted `develop` SHA or `N/A — pre-merge`:
- Effective-activation evidence or `N/A — inactive`:

## Acceptance-criteria evidence

| Criterion | Test or observable evidence | Result |
|---|---|---|
| AC-## |  |  |

## Validation evidence

- Commands executed and results:
- CI run:
- Preview deployment URL/status or `N/A` with reason:
- Independent-verifier report/result:
- Implementer provenance:
- Verifier provenance:

## Impact assessments

- Security and privacy:
- Migration/data integrity:
- Rollback trigger, mechanism, and owner:
- Analytics/telemetry:
- Accessibility:
- Documentation:
- Cloudflare/deployment/operations:

## Release and outcome

- Release/feature-flag plan:
- Monitoring and post-release outcome owner:
- Known limitations/follow-up issues:
- Hosted activation status:
- Automatic-merge status:
- Autonomous-production-release status:
- RL1/RL2 technical-activation status:
- Package closure status:

## Author checklist

- [ ] The effective risk is not below the CI-detected floor.
- [ ] The change stays within the approved scope and contains no unrelated cleanup.
- [ ] All installed checks relevant to this change pass; unavailable checks are
      disclosed rather than represented as passing.
- [ ] No secrets, credentials, or unnecessary personal data are included.
- [ ] Material changes after review will dismiss or renew approvals and verification.
- [ ] AI implementation/review provenance is disclosed above.
