# Approval Matrix

Verification answers whether evidence supports the change. Approval answers whether
an accountable authority authorizes it. Both may be required.

## Current active A-003 authority assignment

- Founder: `@m-e-h-r-d-a-a-d`
- Historical Qualified Human Technical Steward: `@m-e-h-r-d-a-a-d`
- Historical relationship: the same verified human served in two explicitly separate
  capacities for the completed VOC-002 migration, as recorded in
  [technical-steward-appointment.md](technical-steward-appointment.md)

A-003 is effectively active. The historical steward role is retired as routine R3
approval authority. Routine R3 does not require founder approval merely because it is
R3. R4 founder authority remains independently required, and EHR applies only when an
actual exceptional trigger exists.

Builders may implement approved work but cannot self-approve or merge it. Independent
verification remains separate from every human authority.

| Change/action | Independent verification and controls | Human authority | Automation permission |
|---|---|---|---|
| Routine R0-R2 | Proportionate deterministic and independent verification | No founder or standing technical-steward approval | Only where separately implemented and proven |
| Routine R3 protected technical work | Strengthened risk-specific controls and independent verification | No standing technical-steward approval; no founder approval merely because work is R3 | Only where every applicable technical gate is implemented and proven |
| R4 consequential decision/change | R3 controls too when technically protected | Founder approval required | Cannot bypass founder authority |
| Initial public or predefined major launch | Independent release review and every applicable technical gate | Founder go/no-go required | Publish only after recorded founder approval and technical activation |
| Emergency protective rollback using approved runbook | Post-action verification and permanent evidence | Pre-authorized runbook/incident authority; founder only for a new R4 decision | Only a predefined safer-than-waiting action may execute automatically |
| Change to CI/CD, ownership, approval, agent authority, or this matrix | Independent governance/security review and fail-closed validation | Founder approval when the change is R4; no standing steward approval solely for R3 | Cannot self-modify into effect |
| EHR escalation | Operation stops and qualified expertise is recorded | Exceptional qualified human review for the triggered matter only | Not a routine approval layer |

The one-time VOC-002 migration approval is exhausted and must never be reused.
CODEOWNERS remains review routing and is not approval evidence.
R4 founder authority remains unchanged. The migration record must never be reused as
approval for later work.

## Approval evidence

An approval is valid only when it is attributable to the configured GitHub identity,
bound to the reviewed commit or pull-request revision, recorded in GitHub, and not
dismissed by later material changes. Private chat approval is insufficient.

The completed VOC-002 approval explicitly named both founder and technical-steward
capacities and remains historical evidence only; it cannot approve later work.

The historical initial DOC-16/A-002 bootstrap row was the only exception to the
steward requirement then applicable. It expired when that initial pull request merged.
The later appointment of a steward does not retroactively change PR #3 evidence and
does not convert Claude Code review into steward approval.

VOC-002 was not an exception: it used the pre-A-003 requirements in full. Its
technical-steward approval was a one-time migration requirement, is now exhausted,
and is not a standing future rule.

## `develop` merge and `automatic_merge_allowed`

Under active A-003, routine R0–R3 work does not require founder approval
merely because of risk class. Merge-gate may auto-merge an implementation PR
into `develop` when CI is green, independent verification passed, the project
auto-merge switch is on, and the package has `automatic_merge_allowed: true`
(required for R0–R3 drafting per `AGENTS.md` and
[issue #573](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/573)).
R4 always requires founder approval; R4 packages must set
`automatic_merge_allowed: false`. The founder's literal `approved` comment
remains a valid override when the gate requires approval (R4, unparseable risk,
or a residual `false` on an existing package).

## Independent verifier result

The verifier reports `PASS`, `PASS WITH NON-BLOCKING FINDINGS`, or `FAIL`. Any open
Critical, High, or required Medium finding blocks merge. The builder cannot be the
independent verifier. If the verifier authors a substantial correction, all checks
rerun and a separate independent review is required for the correction.
