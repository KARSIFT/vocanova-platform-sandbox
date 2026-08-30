# Governance

This directory contains the current controls for Vocanova's autonomous development
workflow. Read these documents together with
[DOC-15](../operations/15-ai-native-product-and-engineering-operating-model.md).
A-004 is the effective engineering-workflow authority model after canonical merge of
`VOC-080-T07`; it removes founder `approved`-comment gates at every risk class while
retaining deterministic checks, independent verification, risk floors, EHR, and
stronger R4 evidence. A-003 remains frozen historical audit evidence, and its
non-conflicting controls remain effective through A-004.

## Current documents

| Document | Purpose |
|---|---|
| [DOC-16](16-autonomous-development-operating-model.md) | Approved canonical autonomous-development operating model |
| [Amendment A-002](amendments/A-002-governed-autonomous-releases.md) | Approved canonical release-authority amendment |
| [Amendment A-003](amendments/A-003-governed-autonomous-engineering-authority.md) | Approved historical predecessor; substantive body remains frozen |
| [Amendment A-004](amendments/A-004-remove-founder-approval-gates-from-autonomous-engineering-workflows.md) | Approved and effectively active successor (VOC-080); removes engineering-workflow founder gates |
| [A-003 transition state](a003-transition-state.yaml) | Machine-readable A-003 approval, adoption, activation, and operational truth (historical) |
| [A-004 transition state](a004-transition-state.yaml) | VOC-080 successor; **effective** engineering-workflow authority (`a004-active`) |
| [Technical-steward appointment](technical-steward-appointment.md) | Permanent historical evidence; retired as routine R3 approval authority |
| [Change risk classification](change-risk-classification.md) | R0-R4 classification and verification requirements |
| [Protected areas](protected-areas.md) | Sensitive paths and change types |
| [Approval matrix](approval-matrix.md) | Required decision, technical, verification, and release authorities |
| [Repository settings](repository-settings.md) | Required GitHub and external configuration |
| [Post-merge activation checklist](post-merge-activation-checklist.md) | Tracked steps required before protected or autonomous releases |

Governance changes are protected changes. An author or implementation agent cannot
be the sole approver of a governance change.
