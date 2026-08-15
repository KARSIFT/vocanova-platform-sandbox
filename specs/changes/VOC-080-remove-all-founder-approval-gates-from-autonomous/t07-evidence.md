# VOC-080-EV-07 — T07 transition activation and post-activation unblock

Evidence for `VOC-080-AC-00` through `VOC-080-AC-10` (activation closure).
Tests: `VOC-080-TEST-06`, `VOC-080-TEST-08`.

## Task outcome

`VOC-080-T07` records the one-time A-004 transition activation under pre-A-004
authority, flips `authority_model` to `a004-active`, synchronizes canonical docs and
protected-path policy markers, and confirms VOC-079 / issue #624 can resume on the
no-founder-gate engineering-workflow path.

**Authority state after this task:** `a004-active` (`effective_activation_at`:
`2026-08-15T08:30:00Z`). The one-time founder transition approval for this revision
is **exhausted** and must not be reused as a standing engineering-workflow gate.

## Preconditions verified

| Prerequisite | Evidence |
|--------------|----------|
| VOC-080 adopted under A-003 | PR [#628](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/628); `change.yaml` `status: adopted` |
| T00–T05 merged with independent verification | PRs [#640](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/640)–[#644](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/644) |
| T06 rehearsal PASS | [t06-evidence.md](t06-evidence.md) |
| Infra merge-gate / adopt / release contracts | `VOC-080-EV-01`–`EV-03`; infra pin `489dd82` |
| Production environment reviewers | `reviewers: null` (`VOC-080-EV-03`) |

## Activation revision binding

| Field | Value / note |
|-------|----------------|
| Pre-task `develop` tip (`adopted_develop_sha`) | `69b8cb98ea2c4e5726b67f901d35151ee0366e02` |
| `approved_pr_head_sha` | Binds to the **exact implementing PR head** at independent verification and founder transition approval; `null` in transition-state until the workflow commit SHA is recorded on the PR |
| `frozen_source_sha256` (A-004) | `6668b49477549680193945cf7146ff0babc2c4a61c3a5b06f7da8cf48ad53c3d` |
| Task issue | [#637](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/637) |
| Requirement source | [#627](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/627) |

Independent verifier and founder transition approval must bind to the same exact
commit SHA as this task's implementing PR head. Post-merge, update
`approved_pr_head_sha`, `approved_adopted_tree_sha`, `post_merge_validation_status`,
and GitHub comment URLs in `a004-transition-state.yaml` if they differ from the
pre-merge scaffold.

## One-time founder transition approval

| Item | Evidence |
|------|----------|
| Founder direction (requirement) | Issue [#627](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/627) (`m-e-h-r-d-a-a-d`, effective 2026-08-15) |
| One-time transition approval (exact revision) | Recorded on this task's implementing PR (`VOC-080-T07`) before merge — binds activation revision SHA |
| `migration_approval_status` | `exhausted-non-reusable` |
| `migration_approval_exhausted` | `true` |

Issue #627 authorizes preparation and adoption under pre-A-004 authority; it is not
by itself the canonical exact-revision activation approval. The T07 PR founder
`approved` comment (or equivalent recorded approval) on the activation revision
satisfies A-004 §2.1 / §14.

## Independent verification

| Item | Evidence |
|------|----------|
| T00–T06 task PRs | PASS / PASS WITH NON-BLOCKING FINDINGS on exact SHAs (see `t01-evidence.md`–`t06-evidence.md`) |
| T07 activation revision | Independent `review` / `plan_reviewer` PASS bound to implementing PR head SHA (task issue [#637](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/637)) |
| Builder/verifier separation | Implementer did not approve or merge its own work (`VOC-080-TEST-07`) |

## Activation artifacts

| Path | Change |
|------|--------|
| `docs/governance/a004-transition-state.yaml` | `authority_model: a004-active`, `effective_activation_status: active`, rehearsal + adoption evidence |
| `docs/governance/amendments/A-004-…md` | `status: approved`, effective notice, synchronized frontmatter |
| `docs/governance/a003-transition-state.yaml` | `successor_effective_activation_status: active` pointer only |
| `.github/approved-policy/protected-paths.yaml` | `authority_model: a004-active`, `status: approved-a004-active` |
| `AGENTS.md`, `CLAUDE.md`, DOC-16, matrices, templates, PR template | Post-activation authority language |
| `tooling/governance/validate_repository_foundation.py` | `validate_a004_lifecycle` + protected-path lockstep |
| `specs/changes/VOC-080…/change.yaml` | `authority_model: a004-active` |

## VOC-080-TEST-08 — post-activation doc and settings check

### Procedure 1 — live founder-gate phrase grep

```bash
rg -n "Founder approval is required for|Requires founder approval|Founder approves develop|Publication to production requires founder|does not replace founder approval|cannot reach \`main\` or production without founder|reply \`approved\`" \
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

Expected: no **live** engineering-workflow founder-gate claims; historical sections
explicitly marked.

### Procedure 2 — post-activation authority claims

Canonical docs assert **A-004 active** for engineering-workflow gates; A-003/VOC-075
founder-merge requirements appear only in historical context.

### Procedure 3 — repository settings

Production environment `reviewers: null` (recorded `VOC-080-EV-03`; no founder
environment-reviewer on repository-controlled deploy path). `auto_merge_enabled` and
`auto_release_enabled` remain `true` on this sandbox (`pipeline.yml`).

## VOC-079 / issue #624 unblock (`VOC-080-AC-10`)

After this activation:

- [VOC-079](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/624) plan PR
  [#625](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/625) merged-as-draft
  failure class is recoverable via `workflow_dispatch` reconcile
  (`VOC-080-EV-02` / `t06-evidence.md`).
- R4 merge, adopt, release, and deploy paths no longer require founder `approved`
  comments when non-founder gates pass.
- VOC-079 **implementation** (nginx cutover) remains out of scope; only gate clearance
  for post-activation progression is claimed here.

## Acceptance criteria closure

| AC | Result | Primary evidence |
|----|--------|------------------|
| AC-00 | **pass** | `t06-evidence.md` §1; `t02-evidence.md` |
| AC-01 | **pass** | `t06-evidence.md` §2; `t01-evidence.md` |
| AC-02 | **pass** | `t06-evidence.md` §4; `t05-evidence.md` |
| AC-03 | **pass** | `t06-evidence.md` §5; `t05-evidence.md` |
| AC-04 | **pass** | `t06-evidence.md` §1; `t02-evidence.md` |
| AC-05 | **pass** | `t06-evidence.md` §3; `t02-evidence.md` |
| AC-06 | **pass** | `t06-evidence.md` §6; `t03-evidence.md` |
| AC-07 | **pass** | `t04-evidence.md`; `t05-evidence.md` |
| AC-08 | **pass** | `t05-evidence.md` |
| AC-09 | **pass** | `t04-evidence.md`; this file TEST-08 |
| AC-10 | **pass** | § VOC-079 unblock above |

## Deterministic validation (this task)

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_voc080*.py' -v
git diff --check
```

## Non-founder controls confirmed post-activation

- Independent verification (plan_reviewer / review) remains mandatory
- CI + governance validation remain fail-closed
- Unparseable risk fails closed; no founder override
- Builder/verifier separation; no self-review of same exact revision
- EHR exceptional-only
- Failed release/deploy remain fail-closed until remediation
- RL1/RL2 technical activation remain **disabled** (unchanged by A-004)

## Explicitly not done

- VOC-079 technical cutover implementation
- Closing issue [#627](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/627)
  (requires merge of this activation PR + recorded post-merge validation)
- Full VOC-080 develop→main promotion (roster closes when this task issue closes)

## Limitations

| Limitation | Impact |
|------------|--------|
| `approved_pr_head_sha` null until PR head bound | Verifier/founder must bind exact SHA on implementing PR; post-merge sync may update transition-state |
| `post_merge_validation_status: pending` | Updates to `passed` after merge governance run URL recorded |
| GitHub comment URLs | Scaffold uses issue links; exact comment URLs recorded at PR review time |

## Overall T07 result

**PASS pending exact-revision founder transition approval and independent verification
on the implementing PR head SHA.** Activation markers, canonical docs, protected-path
policy, and validation harness are flipped to `a004-active`. VOC-079 may resume on the
no-founder-gate path subject only to remaining non-founder gates.
