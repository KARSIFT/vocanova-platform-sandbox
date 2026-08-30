---
id: DOC-19
title: VocaNova Governance Reconciliation Notes
version: 1.0
document_type: governance-orientation
status: proposed
owner: founder
canonical_path: docs/operations/19-governance-reconciliation-notes.md
approved_at: null
last_reviewed_at: 2026-08-30
review_cycle: quarterly
supersedes: null
related_documents:
  - DOC-10
  - DOC-11
  - DOC-15
  - DOC-16
  - DOC-17
  - DOC-18
related_decisions:
  - A-002
  - A-003
  - A-004
adoption_change: VOC-007
source_files:
  - path: 12-governance-and-automation.md
    sha256: 7fda3a4b20321f3c741c8fedddbe26b13ffed64064a09202ce5df0e2fb8fc2a1
---

# 19 — VocaNova Governance Reconciliation Notes

> **Authority boundary:** This proposed note is an orientation aid. It does not amend or replace
> canonical governance. For an authority question, read the linked governance sources directly.

## 1. Current authority in plain terms

VocaNova classifies changes R0–R4 and releases RL1–RL3. These are separate axes. **A-004** is
the active engineering-workflow authority model (activated by canonical merge of VOC-080-T07; see
[the A-004 transition state](../governance/a004-transition-state.yaml)). Routine R0–R3 work uses
proportionate checks and independent verification without standing founder or technical-steward
approval merely because work is R3. R4 carries stronger evidence obligations but no founder
`approved`-comment merge gate on engineering workflows. EHR is an exceptional stop-and-seek-
expertise condition, not a standing approval layer or a risk class.

**Historical (A-003):** Before A-004 activation, effectively active A-003 retired standing
founder or technical-steward approval merely because work was R3; R4 still required founder
approval on merge. That model is preserved in amendment records as audit evidence only.

Canonical sources: [DOC-16](../governance/16-autonomous-development-operating-model.md),
[A-002](../governance/amendments/A-002-governed-autonomous-releases.md),
[A-003](../governance/amendments/A-003-governed-autonomous-engineering-authority.md),
[A-004](../governance/amendments/A-004-remove-founder-approval-gates-from-autonomous-engineering-workflows.md),
[risk classification](../governance/change-risk-classification.md), and the
[approval matrix](../governance/approval-matrix.md).

## 2. Permission is not technical activation

Governance may permit an action class while the repository lacks the technical controls needed to
perform it automatically. This table is a point-in-time snapshot and goes stale as capabilities are
actually activated - treat
[the A-004 transition state](../governance/a004-transition-state.yaml) as the live source of truth,
not this copy. Last corrected 2026-08-30 (the `2026-07-19` snapshot below was stale on the
production-release rows - see that file's own correction note for why):

| Capability | Recorded state |
|---|---|
| A-004 authority model | active |
| RL1 technical activation | `false` |
| RL2 technical activation | `false` |
| Automatic merge allowed (into `develop`) | `true` - live since VOC-012 |
| Production deployment (to `main`) | `enabled` on push when gates pass |
| Autonomous production release | `enabled` via repository-controlled promotion path |
| Control Plane implementation | `false` |

These are implementation-time facts, not permanent promises. The transition YAML and
[activation checklist](../governance/post-merge-activation-checklist.md) remain the sources of truth.

Promotion recovery does not trust a required `ci / ci` result from a parent workflow
that is still running or otherwise not completed successfully. It also rejects any
run that actually executed a non-skipped `release / converge` job, even when the run
title appears dedicated. When no completed non-carrier result exists, a dedicated
`promotion-pr-validation PR #<n>` run must complete successfully for the exact
promotion PR and head before attestation; a skipped shared `release / converge`
definition does not by itself make the run a release carrier.

Production merge uses a strict two-token contract. The mutation App token grants
exactly Contents, Issues, and Pull requests write and is the sole token for merge and
repository mutations. Immediately before the guard verification and merge decision,
the workflow mints a current-caller-repository-scoped, Administration-write-only
guard token used only by `verify-production-merge-guard.sh`. It is never used for
statuses, Issues, Pull requests, Contents, or merge operations, and the verifier
accepts bypass configuration only as an explicit empty `bypass_actors` array.

## 3. What the stale source got right

The source comparison usefully recorded that earlier Documents 10 and 14, DOC-15/A-001, and the
DOC-17/DOC-18/A-003 track described conflicting merge and release models. It also correctly
emphasized GitHub traceability, separation between builder and verifier, approved change packages,
least privilege, kill switches, rollback, and the need to resolve contradictions explicitly.

## 4. What is rejected

The source's final recommendation—immediate Claude-gated automatic merge for every risk class and
one blanket founder approval covering release-PR merge plus production publication—is not current
governance. It lacked visibility into approved A-002 and the later active A-004 model. It must not
be used to interpret DOC-15/A-001 as the latest rule, bypass current exact-SHA, risk, or independent-
verification gates, recreate a standing founder-approval requirement for engineering workflows, or
claim automation is active outside the recorded transition and repository settings.

## 5. Historical approvals are not reusable

The initial DOC-16/A-002 bootstrap exception expired with PR #3. The one-time dual-capacity
VOC-002 approval used for the A-003 transition is exhausted and permanently non-reusable. Neither
record authorizes a later change. Current changes follow their own approved package, exact-revision
verification, risk gates, and any independently required authority.

## 6. Reader rule

Product and technical documents should link to governance rather than copy approval rules. If a
summary conflicts with canonical governance or the live transition state, the canonical source
wins and the summary must be corrected through a governed change.
