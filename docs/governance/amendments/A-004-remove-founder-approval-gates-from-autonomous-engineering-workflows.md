---
id: A-004
title: Remove Founder Approval Gates from Autonomous Engineering Workflows
version: 1.0
document_type: governance-amendment
status: approved
owner: founder
canonical_path: docs/governance/amendments/A-004-remove-founder-approval-gates-from-autonomous-engineering-workflows.md

founder_direction_status: approved
formal_founder_approval_status: approved-exact-revision-github-evidence
repository_adoption_status: adopted
effective_activation_status: active

approved_at: 2026-08-15T08:30:00Z
adopted_at: 2026-08-15T08:30:00Z
effective_at: 2026-08-15T08:30:00Z
approved_pr_head_sha: null
adopted_develop_sha: 69b8cb98ea2c4e5726b67f901d35151ee0366e02
approval_evidence: "https://github.com/KARSIFT/vocanova-platform-sandbox/issues/627"
independent_verification_evidence: "https://github.com/KARSIFT/vocanova-platform-sandbox/issues/637"
repository_adoption_evidence: "https://github.com/KARSIFT/vocanova-platform-sandbox/pull/628"
activation_evidence: "specs/changes/VOC-080-remove-all-founder-approval-gates-from-autonomous/t07-evidence.md"

supersedes:
  - id: A-003
    scope: founder `approved`-comment gates on autonomous engineering workflows (merge, plan adoption, release promotion, deploy/retry paths, unparseable-risk override, and residual `automatic_merge_allowed: false` founder-attention semantics) only where A-003 or VOC-075 require such gates
  - id: VOC-075
    scope: approve-only-R4 engineering-workflow gate policy (issue #573); historical adoption evidence preserved

related_documents:
  - DOC-15
  - DOC-16
  - A-002
  - A-003

related_decisions:
  - VOC-080
  - VOC-080-D00
  - VOC-080-D01
  - VOC-080-D02
  - VOC-080-D03

requirement_source: "GitHub issue #627 (effective 2026-08-15; founder m-e-h-r-d-a-a-d)"
transition_package: VOC-080
---

# Amendment A-004 — Remove Founder Approval Gates from Autonomous Engineering Workflows

> **Effective authority notice:** A-004 is **effectively active** as of
> `2026-08-15T08:30:00Z` per `docs/governance/a004-transition-state.yaml`
> (activation task `VOC-080-T07`; requirement source issue #627). A-003 remains
> authoritative historical audit evidence; engineering-workflow founder `approved`-comment
> gates are superseded where A-004 governs. This amendment did not authorize its own
> adoption or activation; those occurred under pre-A-004 authority.

## 1. Purpose

This amendment removes every founder `approved`-comment gate from VocaNova's
autonomous **engineering workflows** so agents and workflows progress when
deterministic checks, independent verification, scope, governance, and other
non-founder gates pass — at **every risk class including R4** — without waiting
for a founder `approved` comment on merge, adoption, release promotion, deploy,
or remediation retry paths.

Requirement source: [GitHub issue #627](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/627)
(effective request date 2026-08-15). This supersedes VOC-075 / issue #573
“approve only R4” for **engineering-workflow gates** only. Historical VOC-075,
A-003, and audit records remain preserved; they are not rewritten as if prior
rules never existed.

This amendment **does not** remove or weaken:

- independent verification (plan_reviewer / review roles);
- deterministic CI and governance validation;
- risk classification and protected-path floors;
- rollback, monitoring, and audit-evidence requirements;
- secrets isolation and least-privilege credentials;
- builder/verifier separation (no self-review of the same exact revision);
- Exceptional Human Review (EHR) as an exceptional-only escalation;
- the obligation to obtain founder **requirement clarification** when product,
  legal, or strategy requirements are genuinely ambiguous **before** a change
  package has stable acceptance criteria.

The governing principle after effective activation:

> Engineering workflows advance on evidence and gates, not on a standing founder
> `approved` comment. R4 remains a strengthened evidence and control class — not
> a founder-comment merge gate.

---

## 2. Document Lifecycle

A-004 has three separate governance lifecycle stages, mirroring A-003's pattern.

### 2.1 Formal Founder Transition Approval

Founder direction in issue #627 authorizes preparation of this amendment but is not,
by itself, canonical repository approval or effective activation.

The **one final** founder transition approval under pre-A-004 authority (active
A-003 / VOC-075 policy) is required for the **exact activation revision** that
flips `effective_activation_status` to `active`. That approval binds to the exact
commit SHA recorded in `a004-transition-state.yaml` at activation (`VOC-080-T07`).

Until that evidence exists:

```text
formal_founder_transition_approval_status: pending-exact-revision-github-evidence
```

### 2.2 Repository Adoption

A-004 becomes adopted repository governance only when:

1. the final exact revision satisfies the governance rules effective before A-004
   (active A-003);
2. all required deterministic validation passes;
3. required independent verification passes on the adoption revision;
4. all currently required pre-transition approval evidence is recorded;
5. the approved revision is merged into the canonical `develop` branch.

Repository adoption does not retroactively validate actions taken before adoption.

### 2.3 Effective Activation

Repository adoption and effective activation are distinct states.

After adoption, A-004 becomes effective only when:

1. post-merge governance validation required for the transition succeeds;
2. rehearsal evidence recorded under VOC-080 is complete (`VOC-080-T06`);
3. the one-time exact-revision founder transition approval is recorded;
4. independent verification of the **exact activation revision** passes;
5. activation evidence is recorded in `a004-transition-state.yaml`.

During any state where A-004 is adopted but not yet effective, **A-003 continues
to govern** engineering-workflow founder gates.

Once effective activation is recorded:

```text
effective_activation_status: active
authority_model: a004-active
```

no autonomous engineering workflow may wait on a founder `approved` comment.

A-004's governance authority becoming effective does not by itself change RL1/RL2
technical activation, Control Plane implementation, or other gates outside this
amendment's scope.

---

## 3. Authority and Precedence

A-004 partially supersedes A-003 **only** where A-003 (or VOC-075 workflow-gate
interpretations) require a founder `approved` comment for engineering-workflow
progression.

A-004 specifically supersedes requirements for:

- founder `approved` comments to merge plan or implementation PRs at any risk class;
- founder `approved` comments to adopt independently reviewed plan packages;
- founder `approved` comments to override failed or missing deterministic checks or
  independent-verification verdicts;
- founder `approved` comments as the primary or retry gate for develop→main
  promotion or push-to-main production deploy on repository-controlled paths;
- founder `approved` comments as the escape hatch for unparseable or inconsistent
  risk classification;
- treating residual `automatic_merge_allowed: false` as a founder-attention merge
  gate (see §10).

All non-conflicting controls from DOC-15, DOC-16, A-001, A-002, and A-003 remain
in force, including:

- independent verification;
- risk classification and protected paths;
- protected branches and governed pull requests;
- least-privilege access and secrets isolation;
- traceability and rollback requirements;
- separation of implementation and verification;
- prohibition of implementation self-approval;
- EHR as exceptional-only escalation;
- emergency controls and kill switches.

Where A-004 conflicts with A-003's engineering-workflow founder-gate requirements,
A-004 takes precedence **only after** A-004 becomes effectively active.

A-003's substantive body and historical activation evidence remain authoritative
audit records. They are not deleted or falsified.

---

## 4. Engineering Governance Principle

After A-004 becomes effective:

1. AI agents and workflows perform routine engineering work under governed packages.
2. Deterministic systems verify mechanical correctness wherever practical.
3. Independent verification challenges implementation and planning artifacts.
4. Founder involvement for **product, legal, and strategy** matters occurs as
   **requirement clarification before stable acceptance criteria exist** — not as
   an `approved` comment on merge, adopt, release, or deploy gates.
5. Technical autonomy is proportional to evidence, reversibility, and risk class.
6. No builder approves or independently verifies its own exact revision.
7. No workflow uses a human comment to bypass a failed or missing gate.
8. EHR remains available when autonomous confidence is insufficient.

---

## 5. Risk Classification (R0–R4)

VocaNova retains the independent R0–R4 change-risk classification.

### R4 — Consequential Decision (post-A-004)

R4 continues to identify changes with consequential scope: governance authority,
protected effects, material autonomous-system expansion, and similarly strengthened
evidence requirements.

After A-004 activation, R4 **requires**:

- stronger specification, validation, independent verification, rollout, monitoring,
  and rollback evidence as defined by risk policy and protected-path floors;
- traceable audit records bound to exact revisions.

After A-004 activation, R4 **does not require**:

- a founder `approved` comment on merge, adoption, release promotion, deploy, or
  remediation retry paths when all applicable non-founder gates pass.

Unparseable or inconsistent risk classification **fails closed for correction**. It
does not wait for founder override and does not auto-merge.

---

## 6. Exceptional Human Review (EHR)

EHR remains an escalation condition, not a routine approval layer and not a
replacement for removed founder workflow gates.

EHR must not be converted into a standing human approval requirement for routine
engineering workflow progression.

---

## 7. Founder Authority (Clarification vs Workflow Gates)

After A-004 becomes effective, the founder retains authority to:

- answer genuine product, legal, strategy, pricing, business-model, and user-trust
  questions **before** planners draft stable acceptance criteria;
- define or clarify requirements for initial public launch and major launches as
  **specification inputs**, not as merge/adopt/release `approved` comments;
- exercise explicit protected policy conditions that are **not** engineering-workflow
  `approved`-comment gates (EHR, emergency policy, out-of-band repository settings
  documented per AGENTS.md).

The founder **does not** retain standing `approved`-comment authority to:

- merge PRs when CI, governance, scope, or independent verification fail or are
  missing;
- adopt plan packages without passing plan_reviewer / governance validation;
- promote to `main` or deploy when release/deploy gates fail;
- override unparseable risk classification.

---

## 8. Role Boundaries

Codex or another builder may implement changes but may not independently authorize
or verify its own exact revision.

Claude or another independent verifier may produce verification evidence and a
policy-recognized verdict but does not become the founder or a standing human
approval authority.

Deterministic systems may enforce policy but may not use founder-comment handlers
to bypass failed gates.

---

## 9. Merge Authority

### Merge into `develop` (plan and implementation PRs)

After A-004 becomes effective, a change may merge automatically into `develop`
when:

- the applicable change package or specification is valid for the reviewed revision;
- required deterministic checks pass;
- required independent verification passes (plan_reviewer or review as applicable);
- no blocking finding remains that policy treats as merge-blocking;
- no active EHR condition exists;
- risk classification is parseable and consistent;
- repository policy authorizes automatic merge for that change (`auto_merge_enabled`
  and merge-gate contract).

R0–R4 are all eligible for automatic merge when the above gates pass. R4 carries
stronger evidence obligations but **not** a founder-comment gate.

No merge path may treat a founder `approved` comment as sufficient to override a
failed or missing deterministic check or independent-verification verdict.

### `automatic_merge_allowed` (VOC-080-DEP-02)

After A-004 becomes effective:

- all new change packages draft `automatic_merge_allowed: true`, including R4;
- merge-gate ignores historical `automatic_merge_allowed: false` as a founder-attention
  mechanism (optional historical backfill may occur separately);
- the field may remain for audit compatibility but does not gate merge on founder
  attention;
- missing CI, failed governance validation, or non-PASS independent verification
  still fail closed.

---

## 10. Autonomous Package Adoption

After A-004 becomes effective, an independently reviewed plan package that passes
governance and deterministic validation must become `status: adopted` /
`implementation_authorized: true` (and matching nested fields) **automatically**
as part of the adoption path, recording:

- exact revision SHA;
- independent review evidence;
- resolved and explicitly deferred decisions;
- risk classification;
- authority provenance (pre-transition adoption under A-003 through VOC-080).

A bot-mediated plan PR merge must not leave a package silently merged as `draft` /
unauthorized. Idempotent reconcile dispatch must repair merged-but-unadopted
packages and missing task rosters without replaying stale GitHub events.

---

## 11. Promotion to `main` and Production Deploy

After A-004 becomes effective:

- develop→main promotion follows `auto_release_enabled` and release-gate policy
  without a required founder `approved` comment on the repository-controlled path;
- push-to-main production deployment follows existing delegation (2026-08-08) without
  a required founder `approved` comment retry gate;
- failed promotion or deploy attempts remain fail-closed until remediation checks
  pass;
- retry uses dispatch/remediation checks, not founder-comment override.

Residual GitHub environment reviewer requirements that demand founder click-approve
on repository-controlled paths must be removed or documented as out-of-band
operations with immediate doc follow-up per AGENTS.md.

---

## 12. Dependabot and Recognized Bots (VOC-080-DEP-05)

Recognized dependency bots (e.g. Dependabot) follow the documented green-CI exception
without founder approval. They retain the no-agent-authored-code review exception.
Unrecognized bot identities fail closed for correction.

---

## 13. Self-Modification and Governance Safety

Changes affecting merge-gate, adopt, release, remediate, governance amendments, and
related workflow authority remain protected. This amendment's own adoption and
activation were evaluated under active A-003 authority and cannot authorize
themselves.

A-004 does not authorize its own adoption.

---

## 14. Safe Governance Transition (A-004)

Governance changes are evaluated under the governance rules effective before the
proposed replacement becomes active.

Proposed replacement rules cannot authorize their own adoption or activation.

This rule applies to A-004.

Under governance effective before A-004 (active A-003):

- this transition is an R4 governance decision with R3 protected technical effects;
- VOC-080 was adopted under pre-transition authority;
- one final exact-revision founder transition approval is required for the activation
  revision (`VOC-080-T07`);
- post-transition rules do not authorize VOC-080 or A-004 retroactively.

At activation, the one-time founder transition approval is recorded against the
exact activation revision SHA. That approval is **exhausted** for this transition
and must not be reused as a standing engineering-workflow gate.

After valid effective activation of A-004:

- no later autonomous engineering workflow waits on a founder `approved` comment;
- non-founder controls listed in §1 remain mandatory;
- EHR remains exceptional-only;
- builder/verifier separation remains mandatory.

---

## 15. Continuing Principles

After A-004 becomes effective:

1. Engineering workflows progress on deterministic checks and independent verification.
2. R4 means stronger evidence — not a founder-comment gate.
3. Founder input happens upstream as requirement clarity, not downstream as merge
   authority.
4. Failed gates stay fail-closed until remediation succeeds.
5. Audit history under A-003 and VOC-075 remains preserved and citable.
6. VOC-079 and subsequent packages resume on the no-founder-gate path subject only to
   remaining non-founder gates.
