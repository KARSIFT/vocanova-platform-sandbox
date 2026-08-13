# VOC-077 — Implementation Plan

## Preconditions and protected areas

Do not begin until this package is adopted and implementation is
authorized. Adoption should record `VOC-077-DEP-00` and `VOC-077-DEP-01`
so implementers are not guessing about PR #558 disposition or pipeline
follow-up scope.

In-scope paths (ordinary `specs/changes/` package files; measured floor
**R1**):

- `specs/changes/VOC-072-same-request-as-github-issue-535-voc-067-t05/t00-token-provisioning-evidence.md`
- `specs/changes/VOC-072-same-request-as-github-issue-535-voc-067-t05/change.yaml`
  (DEP status text only)
- `specs/changes/VOC-072-same-request-as-github-issue-535-voc-067-t05/acceptance-criteria.md`
  (`VOC-072-AC-00` Result only)
- This package's `t00-evidence.md` (created by the implementer)

Explicitly out of scope: `.github/workflows/`, `infra/`, `apps/`,
`packages/`, migrations, production secret create/delete APIs,
`karsift-ai-infra` implementer/reviewer workflow semantics (unless
adoption expands `VOC-077-DEP-01` into a separate package).

## File reconciliation and implementation sequence

1. Confirm adoption decisions for DEP-00 / DEP-01.
2. Re-verify secret name presence via redacted `gh secret list --env
   production`. Stop if missing.
3. Restore VOC-072 `t00-token-provisioning-evidence.md` to confirmed state
   (substance of `f49ffc50`, not regressed pending).
4. Update VOC-072 `change.yaml` DEP-00/DEP-01 status text; leave adoption /
   authorization / `automatic_merge_allowed` alone.
5. Set `VOC-072-AC-00` Result to satisfied when evidence supports it.
6. Write `t00-evidence.md` under this VOC-077 package with commands,
   redacted excerpts, and SHA.
7. Open/land the task PR; wait for independent review PASS; merge via
   normal merge-gate (no FAIL override).
8. If adoption required closing PR #558 after this merges, do so and note
   it in evidence.

## Validation and independent verification

Deterministic commands before claiming complete:

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Independent verification (per `CLAUDE.md`) must confirm against the exact
implemented revision:

- Evidence gate_status and narrative match live redacted secret list.
- DEP text no longer claims presence outstanding.
- AC-00 Result aligned; no secret values in diff; no out-of-scope files.
- Declared risk meets or exceeds path floor.
- Implementer-role occupant did not approve or merge its own
  implementation.
- Active authority model remains `a003-active`.

## Deployment and rollback

No application deployment effect is intended. Rollback is a documentation /
metadata revert of the VOC-072 evidence and DEP/AC text (and this package's
evidence file).

Rollback trigger: evidence claims resolved while the secret is absent; or
secret values appear in git; or out-of-scope workflow/infra edits landed.

Rollback mechanism: revert the implementation commit(s). Last-known-good:
VOC-072 files immediately preceding this package's implementation merge
(note: pre-merge state may itself be the incorrect pending regression —
rollback returns to that pending state and re-blocks T01/T02; prefer
forward-fix over rollback unless the correction introduced a worse error).

Owner: implementer of `VOC-077-T00`.
