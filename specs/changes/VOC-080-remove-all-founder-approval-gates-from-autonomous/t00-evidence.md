# VOC-080-EV-00 — T00 successor amendment and transition scaffolding

Evidence for `VOC-080-AC-09` (partial — full doc reconciliation is `VOC-080-T04`;
post-activation final check is `VOC-080-T07`). Tests: `VOC-080-TEST-08` (partial).

## Task outcome

`VOC-080-T00` authored the A-004 successor amendment and inactive transition
scaffolding. **A-003 remains the effective authority model.** No workflow,
merge-gate, adopt, release, or activation flip occurred in this task.

## Artifacts added or extended

| Path | Role |
|------|------|
| `docs/governance/amendments/A-004-remove-founder-approval-gates-from-autonomous-engineering-workflows.md` | Proposed A-004 text: removes founder `approved`-comment gates from engineering workflows after activation; preserves non-founder controls; cannot self-authorize |
| `docs/governance/a004-transition-state.yaml` | Machine-readable inactive scaffolding (`authority_model: a003-active`, `effective_activation_status: inactive`, `transition_stage: pre-activation-scaffolding`) |
| `docs/governance/a003-transition-state.yaml` | Extended with successor pointers only; `authority_model` unchanged (`a003-active`) |
| `docs/governance/README.md` | Index entries for A-004 amendment and transition state |

## Key design choices (VOC-080-DEP-00 / DEP-01 / DEP-02)

1. **Successor vehicle (DEP-00):** New A-004 amendment rather than editing A-003's
   frozen body. A-003 historical text and activation evidence preserved.
2. **Clarification vs gates (DEP-01):** A-004 §7 separates founder requirement
   clarification before stable AC from merge/adopt/release `approved` comments.
3. **`automatic_merge_allowed` (DEP-02):** A-004 §9 records neutralize semantics
   (draft `true` for all including R4; merge-gate must not treat `false` as
   founder-attention) — runtime change is `VOC-080-T01`.
4. **Supersession citation:** Issue #627 and VOC-075 workflow-gate supersession
   cited in A-004 frontmatter and §1; VOC-075 historical records not rewritten.

## Authority state after this task

| Field | Value | Notes |
|-------|-------|-------|
| Active authority model | `a003-active` | Unchanged in `a003-transition-state.yaml` |
| A-004 `status` (frontmatter) | `proposed` | Not adopted canonical governance yet |
| A-004 `effective_activation_status` | `inactive` | Activation is `VOC-080-T07` only |
| `successor_authority_model` (a004 state) | `a004-active` | Documented target; not active |
| Final founder transition approval | `pending-exact-revision-github-evidence` | One-time T07 gate under pre-A-004 authority |

## Validation

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Record command outcomes in the implementing PR. Expected: governance validation
passes without flipping A-003 active-state checksums (A-003 body untouched;
`a003-transition-state.yaml` only gained successor pointer fields).

## Explicitly not done (other tasks)

- `merge-gate.yml` / founder override removal (`VOC-080-T01`)
- Autonomous adopt + reconcile dispatch (`VOC-080-T02`)
- Release/deploy founder-gate removal (`VOC-080-T03`)
- AGENTS.md / CLAUDE.md / DOC-15 full reconciliation (`VOC-080-T04`)
- Activation flip (`VOC-080-T07`)

## Limitations

- `frozen_source_sha256` in `a004-transition-state.yaml` remains `null` until the
  activation revision is bound at T07.
- `validate_repository_foundation.py` does not yet assert A-004 lifecycle fields;
  harness extension is expected in `VOC-080-T05` / T07 if required.
