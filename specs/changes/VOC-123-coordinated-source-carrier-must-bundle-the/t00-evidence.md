# VOC-123-T00 — Evidence

Task: `VOC-123-T00` — Bundle the coordinated source-carrier committed head
through a named ref.

Do not record secrets, credentials, session values, OAuth material, personal
data, or complete CI logs.

## Discovery recorded at planning time (issue #1005)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1005 |
| Blocked adopted task | #1003 (`VOC-122-T00`) |
| Pipeline run / job | `32915078678` / `98017696468` |
| Nested commit on the runner | `db31cc9` (`VOC-122: VOC-122-T00 coordinated source carrier (attempt 1)`) |
| Nested diff | 4 files, 506 insertions, 73 deletions, including `tests/test_voc122_actions_check_recovery.py` |
| Error | `fatal: Refusing to create empty bundle.` |
| Artifacts / carrier PRs | none — commit step failed before upload |
| Defect | `git -C karsift-ai-infra bundle create /tmp/implementer-source.bundle "${{ steps.infra-checkout.outputs.base_sha }}..$SOURCE_HEAD_SHA"` after `SOURCE_HEAD_SHA=$(git -C karsift-ai-infra rev-parse HEAD)` |
| Root cause | Git bundle create requires a named positive revision; a raw SHA advertises no bundle head |
| Local reproduction | `git bundle create raw.bundle "$base_sha..$head_sha"` exits 128; `git branch carrier "$head_sha"` then `git bundle create ref.bundle "$base_sha..carrier"` succeeds and list-heads advertises `refs/heads/carrier` |
| Existing test gap | VOC-121 nested-bundle tests use `source_base..HEAD` or `base_sha..$branch`, which are named tips and do not reproduce production |

## Chosen delivery path

| Item | Value |
|------|-------|
| Named-ref mechanism | Record during implementation: exact temporary ref name/namespace, and whether create/verify/cleanup is inline in `implement.yml` or extracted |
| Advertised-head check | Record the exact `git bundle list-heads` assertion used in production |
| Temp-ref cleanup | Record that the ref is deleted after create, including on verification failure |
| Caller recovery `integration_sha..HEAD` | pending implementation — prove safe or repair |
| Planner recovery `base_sha..HEAD` | pending implementation — prove safe or repair |

## Changed surfaces

Fill in during implementation. Expected:

**Infrastructure (`KARSIFT/karsift-ai-infra`):**

- `.github/workflows/implement.yml` nested source-bundle create/verify/cleanup
- `.github/workflows/plan.yml` only if `..HEAD` reproduces the defect
- `config/implementer_source_carrier.py` only if a helper is extracted
- current-state comments and `README.md`
- deterministic real-repository bundle tests

**Caller (`vocanova-platform-sandbox`):**

- `tooling/governance/fixtures/karsift-ai-infra/` pin and consumed files
- caller fixture regressions and any `scripts/foundation/*` pin literals
- this evidence file

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra PR | pending implementation |
| Independently reviewed infra head SHA | pending implementation |
| Exact infra merge SHA | pending implementation |
| Pin applicable? | pending implementation — expected yes because the fixture mirrors `implement.yml` |
| `PINNED_SHA.txt` after source merge | pending implementation |
| Drafting-time pin | `99476c2a1018e42d4bd442657b5257885ac9f1c9` (VOC-121) |

## Dependent #1003 (not implemented by this task)

| Item | Value |
|------|-------|
| VOC-122-T00 / issue #1003 | Distinct already-authorized promotion-recovery outcome |
| This package's duty | Repair carrier integrity; record the exact reviewed infra SHA that #1003 should be re-dispatched or reconciled against |
| Reconstruct `db31cc9` by hand? | No |
| Treat VOC-123 merge as VOC-122 completion? | No |

## Validation commands

```bash
# Infrastructure (exact reviewed source worktree)
python3 -m unittest discover -s tests -p 'test_*.py'

# Caller
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
git diff --check
```

| Command | Result |
|---------|--------|
| `python3 -m unittest discover -s tests -p 'test_*.py'` (reviewed infrastructure head) | pending implementation |
| `bash scripts/governance/validate-governance.sh` | pending implementation |
| `bash scripts/governance/classify-change-risk.sh` | pending implementation |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | pending implementation |
| targeted foundation pin tests if pin literals change | pending implementation |
| `git diff --check` | pending implementation |

## Acceptance mapping

- `VOC-123-AC-00` / `VOC-123-EV-00` — pending implementation
- `VOC-123-AC-01` / `VOC-123-EV-00` — pending implementation
- `VOC-123-AC-02` / `VOC-123-EV-00` — pending implementation
- `VOC-123-AC-03` / `VOC-123-EV-00` — pending implementation
- `VOC-123-AC-04` / `VOC-123-EV-00` — pending implementation
- `VOC-123-AC-05` / `VOC-123-EV-00` — pending implementation
- `VOC-123-AC-06` / `VOC-123-EV-00` — pending implementation
