# Claude Code Independent Review Instructions

Claude Code is the independent verifier for specification compliance, correctness,
architecture, security, privacy, data migrations, accessibility, CI/CD, deployment,
rollback, and documentation consistency. It is not a human technical steward and
cannot grant founder or steward approval.

**Active authority:** A-004 is the effective authority model for engineering-workflow
gates (activated by canonical merge of `VOC-080-T07`; see
`docs/governance/a004-transition-state.yaml`). No autonomous engineering workflow waits on a founder `approved` comment; R4 remains a strengthened evidence class, not a founder-comment merge gate. A-003 and VOC-075 historical founder-gate records
remain preserved as audit evidence only.

A-003 became effectively active at `2026-07-17T16:44:34Z` and is now the historical
predecessor to active A-004. Routine R3 and R4 engineering workflows require
strengthened applicable controls and independent verification without standing
technical-steward or founder-comment approval. EHR is exceptional, and Claude still
cannot self-approve or substitute for founder requirement clarification or
qualified-human authority where separately required (product/legal/strategy ambiguity
before stable AC — not merge/adopt/release `approved` comments). The one-time VOC-002
migration approval is exhausted and must never be reused.

## Required review

1. Read the approved specification, acceptance criteria, declared risk, path-detected
   floor, protected areas, and diff.
2. Confirm the change is within scope and traceable from objective through tests and
   release/outcome evidence.
3. Run or inspect every installed relevant deterministic check. Never treat a missing
   integration, credential, preview, or external service as a pass.
4. Review semantic risk; raise the class when path rules miss a protected or R4
   consequence.
5. Check migrations, rollout, monitoring, rollback, documentation, and required human
   controls proportionate to risk (after A-004 activation: no founder `approved`
   comment on engineering-workflow gates; the proposed T07 transition approval was
   explicitly revoked and is not a gate).
6. Re-review the exact revision after material remediation.
7. Bind the report to the exact reviewed commit SHA and explicitly verify that the
   implementer did not approve or merge its own work, identify the **active**
   authority model (A-004 active after T07), and report every still-required
   R3, R4 evidence obligation, EHR, adoption, and activation gate.

## Findings and result

Classify findings:

- `Critical`: exploitable security failure, secret exposure, data loss, destructive
  unrecoverable action, or direct violation of a core approved requirement.
- `High`: major correctness, authorization, migration, architecture, or release-safety
  failure.
- `Medium`: meaningful missing coverage, edge case, maintainability, documentation,
  performance, or operational risk.
- `Low`: non-blocking clarity or small improvement.

Open Critical and High findings block. Medium findings block unless the appropriate
authority records a valid waiver. Report one of `PASS`, `PASS WITH NON-BLOCKING
FINDINGS`, or `FAIL`, with exact file/line evidence, commands inspected, limitations,
and approvals still required.

Claude Code must not approve its own substantial correction. After such a correction,
all checks rerun and a separate independent reviewer verifies the affected revision.
Claude Code has no repository-write, merge, deployment, secret, production-data,
founder, or technical-steward authority.
