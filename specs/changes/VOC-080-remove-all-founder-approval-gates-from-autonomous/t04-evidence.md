# VOC-080-EV-04 — T04 caller-repo wiring and documentation reconciliation

Evidence for `VOC-080-AC-09` and `VOC-080-AC-07`. Tests: `VOC-080-TEST-08`
(pre-activation; final post-activation settings check remains `VOC-080-T07`).

## Task outcome

`VOC-080-T04` reconciles caller `pipeline.yml` and canonical docs with the
post-T01/T03 infra contracts: no founder `approved` comment as an
engineering-workflow gate; R0–R4 auto-merge eligibility when non-founder
gates pass; autonomous adopt + `reconcile` / `reconcile-release` recovery;
`automatic_merge_allowed: true` for all new packages (`VOC-080-DEP-02`).

**A-003 remains effective until `VOC-080-T07`.** This task does not flip
`a004-transition-state.yaml` / `authority_model`.

## Files reconciled (attempt 1 + remediation)

| Path | Change summary |
|------|----------------|
| `.github/workflows/pipeline.yml` | Removed `founder_username` / `issue_comment` founder-merge paths; wired `reconcile` and `reconcile-release` |
| `AGENTS.md` | Post-A-004 drafting/`automatic_merge_allowed`; adopt/release reconcile; release authority without founder comment |
| `CLAUDE.md` | A-003-until-T07 / post-A-004 verifier guidance |
| `docs/operations/15-ai-native-product-and-engineering-operating-model.md` | §1 overview, §16 branch/PR rules, §17 pipeline, §18.3 release checks, §19.2 production flow, §20 protected controls, DG1/DG4/DG5/DG10, §27 superseded — live founder-gate claims removed or marked historical |
| `docs/governance/16-autonomous-development-operating-model.md` | Post-A-004 gate language; historical A-003/VOC-075 preserved |
| `docs/governance/approval-matrix.md` | R0–R4 no founder-comment engineering gates after activation |
| `docs/governance/change-risk-classification.md` | R4 evidence class; DEP-02 drafting |
| `docs/governance/repository-settings.md` | Production reviewers; historical R4 founder-comment clause clarified vs post-A-004 |
| `docs/governance/protected-areas.md` | Post-A-004 R4 path-floor note |
| `docs/governance/post-merge-activation-checklist.md` | No founder-comment engineering gates after activation |
| `.github/pull_request_template.md` | Risk / approval wording aligned |
| `specs/templates/change-package/*` | `automatic_merge_allowed: true` for all risk classes |

Infra reusable-workflow logic remains T01–T03 (out of scope here).

## `VOC-080-TEST-08` procedure 1 — grep trail (remediation)

Commands run against the working tree after DOC-15 remediation:

```bash
rg -n "Founder approval is required for|Requires founder approval|Founder approves develop|Publication to production requires founder|does not replace founder approval|cannot reach \`main\` or production without founder" \
  docs/operations/15-ai-native-product-and-engineering-operating-model.md \
  AGENTS.md CLAUDE.md \
  docs/governance/approval-matrix.md \
  docs/governance/change-risk-classification.md \
  docs/governance/repository-settings.md \
  docs/governance/protected-areas.md \
  docs/governance/post-merge-activation-checklist.md \
  docs/governance/16-autonomous-development-operating-model.md \
  .github/workflows/pipeline.yml \
  specs/templates/change-package/
```

Expected: **no live matches** requiring founder approval for R4/adopt/merge/release
on the repository-controlled path. Remaining founder-approval strings are either
(a) explicitly marked Historical / Correction, (b) product/legal requirement
clarification or exceptional waiver language, or (c) preserved A-003/A-002/VOC-075
amendment bodies.

Remediation closed the High finding from independent review of
`28bc862575271a44c772abeb0589753e2a4317f3`: live DOC-15 §16.1 / §18.3 (and related
overview / release / DG) founder-gate claims without historical marking.

## Settings documentation note

`repository-settings.md` documents that the production environment has no founder
environment-reviewer requirement on the repository-controlled path (aligned with
`VOC-080-EV-03` / T03 live API inspection: `reviewers: null`). Final post-activation
assert that documented settings match live config remains `VOC-080-T07` /
TEST-08 procedure 3.

## Non-founder controls preserved (AC-07)

- Builder/verifier separation; no implementer self-review of the same exact revision
- CI green + independent verification PASS / PASS WITH NON-BLOCKING FINDINGS as hard gates
- Unparseable / inconsistent risk fails closed (no founder override)
- EHR exceptional-only; not a standing approval layer
- Failed release/deploy remain fail-closed until remediation checks pass

## Validation

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Recorded in this remediation run: foundation + governance structure validation
passed; path floor **R4** (DOC-15); `git diff --check` clean.

## Explicitly not done

- Activation / `authority_model` flip (`VOC-080-T07`)
- Infra merge-gate / adopt / release logic (T01–T03)
- Live sandbox rehearsal (`VOC-080-T06`)
- Closing issue #627

## Limitations

- Live GitHub environment reviewer re-check deferred to T07 final settings assert
  (T03 already recorded `reviewers: null`).
- TEST-08 procedure 2 (post-activation claim absence) completes at T07.
