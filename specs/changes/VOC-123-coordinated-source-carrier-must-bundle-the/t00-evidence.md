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
| Named-ref mechanism | `refs/karsift/source-bundle-head` via `config/implementer_source_carrier.py create-bundle`; `implement.yml` calls the helper after the isolated nested commit |
| Bootstrap carrier | `VOC-123-D08` supervised bootstrap via infrastructure PR #158 (`agent/voc123-t00-bootstrap`); merged to `main` as `7500a4171d96a8e0d38889a9c92ad5dc092ad8dd`; no direct `main` push or runner-environment Git interception |
| Advertised-head check | `git bundle list-heads` must equal exactly `{head_sha} refs/karsift/source-bundle-head`; enforced by `verify_bundle_heads()` before upload |
| Temp-ref cleanup | `update-ref -d refs/karsift/source-bundle-head` in a `finally` block; bundle deleted on verification failure |
| Caller recovery `integration_sha..HEAD` | **proven safe, unchanged** — real-repo tests show non-empty bundle with `list-heads` `{head_sha} HEAD` on attached and detached HEAD |
| Planner recovery `base_sha..HEAD` | **proven safe, unchanged** — same proof as caller recovery |

## Changed surfaces

**Infrastructure (`KARSIFT/karsift-ai-infra`, PR #158):**

- `.github/workflows/implement.yml` — VOC-123 current-state comment; nested source bundle now calls `implementer_source_carrier.py create-bundle` instead of raw-SHA `bundle create`
- `config/implementer_source_carrier.py` — `create_verified_source_bundle()`, `verify_bundle_heads()`, `SOURCE_BUNDLE_REF = refs/karsift/source-bundle-head`, CLI `create-bundle` subcommand
- `README.md` — source-carrier paragraph describes temporary fixed-name ref binding before bundle create
- `tests/test_voc123_source_bundle.py` — real-Git regressions for raw-SHA failure, named-ref success, fail-closed heads/base/SHA/cleanup, caller/planner `..HEAD` proof, publisher-contract preservation
- `tests/test_voc121_implement_policy.py` — workflow assertions updated from raw `bundle create` to helper invocation

**Caller (`vocanova-platform-sandbox`):**

- `tooling/governance/fixtures/karsift-ai-infra/` — mirrored `implement.yml`, `implementer_source_carrier.py`, VOC-121/123 tests, and fixture README pin paragraph
- `tooling/governance/tests/test_voc121_implement_policy.py` — pin advanced; named-ref fixture regression added
- `scripts/foundation/voc097-fixture-matrix.test.mjs`, `scripts/foundation/voc104-ready-for-review-reuse.test.mjs`, `scripts/foundation/voc108-authoritative-lifecycle.test.mjs` — pin literals advanced to `7500a417…`
- this evidence file

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra PR | https://github.com/KARSIFT/karsift-ai-infra/pull/158 — merged |
| Independently reviewed infra head SHA | bound to PR #158 exact final revision (bootstrap carrier) |
| Exact infra merge SHA | `7500a4171d96a8e0d38889a9c92ad5dc092ad8dd` |
| Pin applicable? | **yes** — `implement.yml`, `implementer_source_carrier.py`, and VOC-123 tests are in the policy fixture subset |
| `PINNED_SHA.txt` after source merge | `7500a4171d96a8e0d38889a9c92ad5dc092ad8dd` |
| Fixture/source comparison | **PASS** — `implement.yml`, `implementer_source_carrier.py`, `test_voc123_source_bundle.py`, and `test_voc121_implement_policy.py` are byte-for-byte identical between the reviewed infra tree and the pinned fixture |

## Dependent #1003 (not implemented by this task)

| Item | Value |
|------|-------|
| VOC-122-T00 / issue #1003 | Distinct already-authorized promotion-recovery outcome |
| This package's duty | Repair carrier integrity; record the exact reviewed infra SHA that #1003 should be re-dispatched or reconciled against |
| Re-dispatch against | `7500a4171d96a8e0d38889a9c92ad5dc092ad8dd` (infra `main` after VOC-123 bootstrap merge) |
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
(cd tooling/governance/fixtures/karsift-ai-infra && \
  python3 -m unittest discover -s tests -p 'test_*.py')
node --test scripts/foundation/voc097-fixture-matrix.test.mjs \
  scripts/foundation/voc104-ready-for-review-reuse.test.mjs \
  scripts/foundation/voc108-authoritative-lifecycle.test.mjs
git diff --check
```

| Command | Result |
|---------|--------|
| `python3 -m unittest discover -s tests -p 'test_*.py'` (reviewed infrastructure head) | **PASS** (336 tests) |
| `python3 -m unittest discover -s tests -p 'test_*.py'` (pinned fixture) | **PASS** (255 tests) |
| `bash scripts/governance/validate-governance.sh` | **PASS** |
| `bash scripts/governance/classify-change-risk.sh` | **PASS** — R4 floor |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | **PASS** (185 tests) |
| `node --test scripts/foundation/voc097-fixture-matrix.test.mjs scripts/foundation/voc104-ready-for-review-reuse.test.mjs scripts/foundation/voc108-authoritative-lifecycle.test.mjs` | **PASS** — pin assertions match `7500a417…` |
| `git diff --check` | **PASS** |

Targeted VOC-123 tests:

```bash
python3 -m unittest tests.test_voc123_source_bundle  # infra and pinned fixture
```

| Command | Result |
|---------|--------|
| `python3 -m unittest tests.test_voc123_source_bundle` (infra) | **PASS** (8 tests) |
| `python3 -m unittest tests.test_voc123_source_bundle` (pinned fixture) | **PASS** (8 tests) |

## Acceptance mapping

- `VOC-123-AC-00` / `VOC-123-EV-00` — named-ref bundle advertises exact committed head via `refs/karsift/source-bundle-head`
- `VOC-123-AC-01` / `VOC-123-EV-00` — raw-SHA positive tip reproduces empty-bundle (exit 128) in `test_raw_sha_positive_tip_reproduces_empty_bundle`
- `VOC-123-AC-02` / `VOC-123-EV-00` — wrong/missing/multiple heads, malformed SHA, wrong base, cleanup mismatch fail closed
- `VOC-123-AC-03` / `VOC-123-EV-00` — caller/planner `..HEAD` paths proven safe; unchanged in production workflows
- `VOC-123-AC-04` / `VOC-123-EV-00` — VOC-121 isolation, App-token split, lease, retry limits, non-closing source PR preserved
- `VOC-123-AC-05` / `VOC-123-EV-00` — deterministic real-repository tests in `test_voc123_source_bundle.py`
- `VOC-123-AC-06` / `VOC-123-EV-00` — docs/comments updated; fixture pin equals infra merge `7500a417…`; #1003 recorded as distinct re-dispatch
