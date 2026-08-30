# Change Risk Classification

Risk is classified by consequence, reversibility, blast radius, data sensitivity,
and authority impact—not by diff size alone. The effective class is the highest class
identified by the builder, path-based policy check, independent verifier, technical
steward, or founder.

| Class | Objective criteria and examples | Required verification | Approval | Release and rollback |
|---|---|---|---|---|
| R0 | Documentation, comments, formatting, tests, or metadata with no behavioral, policy, authority, security, or release effect | Policy check; applicable document/link checks; independent verification proportionate to the change | No founder or technical-steward approval unless a protected policy is affected | May release automatically; revert commit or redeploy prior artifact |
| R1 | Low-risk behavioral implementation; small blast radius; no sensitive data; backward compatible; independently reversible | Full installed baseline CI, acceptance evidence, independent verifier, preview when relevant | No founder or technical-steward approval | May release automatically after gates; tested revert, flag, or prior-artifact redeploy |
| R2 | Moderate blast radius, cross-component change, non-destructive schema addition, API addition, significant dependency update, or operational change with tested recovery | R1 plus integration/contract, accessibility, migration, security, performance, or staging checks as applicable | Independent verifier; designated domain review may be required | May release automatically only with staged evidence, monitoring, named rollback owner, and tested recovery |
| R3 | Authentication/authorization; sensitive-data handling; schema migrations; billing implementation; secrets; production infrastructure; AI-provider controls; audio/voice storage; backups; CI/CD, rollback, security, governance enforcement, or agent authority | All applicable CI; specialist security/architecture/migration/deployment review; independent verifier; explicit protected-area evidence | No standing technical-steward or founder approval solely because work is R3; EHR only when exceptionally triggered | Strengthened controls required; controlled rollout and tested rollback/recovery required; destructive operations require restore evidence |
| R4 | New or changed strategy, pricing, financial commitment, legal position, privacy policy, material product direction, public promise, user-trust posture, difficult-to-reverse action, initial public launch, or major launch | R3 checks when technical areas are affected; decision record; impact analysis; independent verification | **Active A-004:** strengthened evidence and controls; no founder `approved` comment on engineering-workflow gates. Founder clarifies product/legal/strategy requirements before stable AC. **Pre-A-004 (historical):** founder approval on merge. | Explicit go/no-go evidence, rollback or contingency plan; no founder-comment merge gate under active A-004 |

The table above reflects the active A-004 authority reconciled in VOC-080-T04 and
activated in VOC-080-T07. **A-004 is the effective authority model** for
engineering-workflow gates. The R0-R4 risk definitions and verification requirements
remain. Routine R3 does not require standing technical-steward approval or founder
approval merely because it is R3. R4 does not require a founder `approved` comment
on merge, adoption, release, or deploy when non-founder gates pass. EHR is an
exceptional escalation condition, not a routine approval layer or risk class.

## `automatic_merge_allowed` drafting

Package-level `automatic_merge_allowed` in each change package's `change.yaml` is
retained for audit compatibility. **Under active A-004** (`VOC-080-DEP-02`):

- **R0–R4:** draft `true`. Do not set `false` to require founder attention on merge.
- merge-gate ignores historical `false` as a founder-attention mechanism.
- Missing CI, failed governance validation, non-PASS independent verification, or
  unparseable risk still fail closed.

**Historical (VOC-075 / issue #573, superseded by issue #627):** Under active A-003
before activation, R0–R3 set `true` and R4 set `false` with merge-gate R4 founder block.

The completed A-003 transition was fixed at R4 with an R3 protected
governance/authority effect and was approved under pre-A-003 governance. Its one-time
migration approval is exhausted and cannot authorize another change. Canonical A-003
lifecycle state and amendment paths retain an R4 floor.

## Classification tests

Use the highest `Yes` answer:

1. Does this decide an R4 matter, commit the company publicly or financially, change
   user rights/trust, or authorize the initial/major launch? Classify R4.
2. Does it touch a protected technical path or change security, data, infrastructure,
   deployment, rollback, governance enforcement, or agent authority? Classify at
   least R3.
3. Does it have moderate blast radius, require coordinated migration, or require
   specialized validation even though it is not protected? Classify at least R2.
4. Does it change behavior but remain small, backward compatible, and independently
   reversible? Classify at least R1.
5. Is it demonstrably non-behavioral and non-policy documentation or maintenance?
   Classify R0.

Uncertainty raises the class until resolved. Splitting a change does not lower the
classification if the parts produce one combined consequence.

## Historical initial-governance bootstrap classification

The initial DOC-16/A-002 adoption was R4 because it established consequential
governance. It required founder approval, independent Claude Code verification, and
passing repository validation. The one-time bootstrap exception permitted that
initial governance pull request to merge without a nonexistent technical steward,
but did not lower its risk, satisfy steward approval, authorize production, or apply
to any later R3/R4 change. The then-current steward requirement became effective
after merge and remained so until A-003 retired it. Active A-004 now governs
engineering-workflow gates.

## Automated risk floor

`scripts/governance/classify-change-risk.sh` computes a floor from changed paths and
the pull-request declaration. The workflow fails when the declared class is below
the detected floor. The classifier cannot identify semantic R4 decisions reliably;
the independent verifier must compare the diff with this document and escalate.

False positives are corrected by changing the classifier in the same independently
reviewed pull request. A comment, label, or approval cannot simply suppress the
detected floor.

Path classification remains a risk floor, not proof of a human approval requirement.
Under active A-004, neither an R3 nor R4 path floor creates a founder-comment or
standing technical-steward approval gate; the applicable strengthened evidence,
independent verification, rollout, monitoring, and rollback controls still apply.

## Waivers

Required deterministic security checks are not builder-waivable. **Under active A-004,** no
founder `approved` comment may waive a failed or missing gate. Under active A-003
until activation, R4 founder approval on merge was required (historical). No waiver
may recreate routine steward approval; EHR and independently applicable controls
remain mandatory when triggered.
A time-limited waiver for another blocking check must name
the authority, reason, scope, expiry, compensating control, follow-up issue, and
rollback condition. The release record links the waiver.
