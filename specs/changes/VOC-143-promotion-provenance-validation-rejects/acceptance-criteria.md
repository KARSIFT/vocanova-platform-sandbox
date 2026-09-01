# VOC-143 — Acceptance Criteria

## VOC-143-AC-00 — `squash-safe-push` accepts a historical fixture after a current `AGENTS.md` documentation update

- Requirement source: `VOC-143-D01`
- Tasks: `VOC-143-T00`
- Tests: `VOC-143-TEST-00`
- Evidence: `VOC-143-EV-00`
- Result: pending

Under `VOC112_CAPTURE_PROVENANCE_MODE=squash-safe-push`, the VOC-112
benchmark suite passes when fixture `agents_sha256` equals `AGENTS.md` at an
immutable ancestor of the validation tip and the current working-tree
`AGENTS.md` hash differs. The class of promotion `validate` failure on PR
#1119 / head `376e00dd769253d7a255660f5391fb208781e2f3` cannot recur for that
documentation-update class.

## VOC-143-AC-01 — `squash-safe-push` fails closed when the fixture `AGENTS.md` hash is not on an ancestor of the tip

- Requirement source: `VOC-143-D02`
- Tasks: `VOC-143-T00`
- Tests: `VOC-143-TEST-01`
- Evidence: `VOC-143-EV-00`
- Result: pending

A tampered, truncated, or otherwise unfound `agents_sha256` does not pass
`squash-safe-push`. Skipping working-tree equality without an ancestor bind
is not sufficient.

## VOC-143-AC-02 — Promotion `pr-validation` accepts the same historical `AGENTS.md` fixture so `ci / ci` can pass

- Requirement source: `VOC-143-D03`
- Tasks: `VOC-143-T00`
- Tests: `VOC-143-TEST-05`
- Evidence: `VOC-143-EV-00`
- Result: pending

With `VOC112_PROMOTION_PR=true` and mode `pr-validation`, fixture
`agents_sha256` may match `AGENTS.md` at an ancestor of `PR_HEAD_SHA` rather
than at current HEAD. Navigator skill hashes remain anchored at `PR_HEAD_SHA`.
The class of promotion `ci / ci` failure on PR #1119 cannot recur for the
same documentation-update class.

## VOC-143-AC-03 — `local`, `pr-ancestry`, and ordinary `pr-validation` stay fail-closed

- Requirement source: `VOC-143-D04`
- Tasks: `VOC-143-T00`
- Tests: `VOC-143-TEST-02`, `VOC-143-TEST-03`, `VOC-143-TEST-04`
- Evidence: `VOC-143-EV-00`
- Result: pending

`local` and `pr-ancestry` still require fixture `agents_sha256` to equal the
current working-tree `AGENTS.md` hash. Ordinary (non-promotion)
`pr-validation` still requires merge-base-anchored hashes for both hashed
sources.

## VOC-143-AC-04 — VOC-139 navigator HEAD-binding and non-ancestor-base rejection remain

- Requirement source: `VOC-143-D03`, `VOC-143-D05`
- Tasks: `VOC-143-T00`
- Tests: `VOC-143-TEST-06`
- Evidence: `VOC-143-EV-00`
- Result: pending

Promotion `pr-validation` still rejects merge-base-only navigator hashes and
still rejects a promotion base that is not an ancestor of its head. Existing
VOC-139 tests that supply `agents_sha256` equal to HEAD still pass.

## VOC-143-AC-05 — Promotion check identity and historical fixtures are unchanged

- Requirement source: `VOC-143-D05`, `VOC-143-D06`
- Tasks: `VOC-143-T00`
- Tests: `VOC-143-TEST-07`
- Evidence: `VOC-143-EV-00`
- Result: pending

`repository-governance.yml` still selects `squash-safe-push` for
same-repository `main` ← `develop` promotion PRs. Reusable `ci.yml` still
uses `--promotion-pr` / `pr-validation` for that pair. VOC-112 JSON fixtures
are byte-identical to the implementation PR base. `PINNED_SHA.txt` is
unchanged. No fetch/hydrate helper is added.

## VOC-143-AC-06 — Current-state docs match the live ancestor-bind contract

- Requirement source: `VOC-143-D08`
- Tasks: `VOC-143-T00`
- Tests: `VOC-143-TEST-07`
- Evidence: `VOC-143-EV-00`
- Result: pending

An exhaustive tracked-source search identifies every current claim that
`squash-safe-push` requires working-tree `AGENTS.md` equality or that
promotion `pr-validation` requires fixture `agents_sha256` to equal current
HEAD `AGENTS.md` after a documentation update. Those live documents are
updated. Clearly marked historical VOC-142/VOC-139/VOC-114 records remain
historical. Those package directories are not rewritten.

## VOC-143-AC-07 — Deterministic suites and exact-SHA review pass

- Requirement source: `VOC-143-D10`, `VOC-143-D11`, `VOC-143-D12`
- Tasks: `VOC-143-T00`
- Tests: `VOC-143-TEST-08`
- Evidence: `VOC-143-EV-00`
- Result: pending

After the repair is tracked and committed, governance validation, risk
classification, the VOC-112 and VOC-114 Node suites, `git diff --check`, and
independent exact-revision review that binds the live head all pass.
`roles.yml` is unchanged. Evidence does not require a commit to contain its
own SHA.

## VOC-143-AC-08 — Documented `reconcile-release` can recover #1118; no snapshot-gap

- Requirement source: `VOC-143-D09`, `VOC-143-D14`
- Tasks: `VOC-143-T00`
- Tests: `VOC-143-TEST-09`
- Evidence: `VOC-143-EV-00`
- Result: pending

No snapshot of the current develop/main gap is committed. No duplicate
promotion PR or release audit is created. #1119 is not manually merged,
closed, recreated, or bypassed by this package. After exact-SHA review and
merge into `develop`, ordinary `reconcile-release` for release issue #1118
can re-evaluate the live same-repository promotion. Root issue #1120 closes
only after allowlisted metadata from a successful recovery/release run
exists. Closed state alone is not completion proof. Named incident
PR/SHA/issue IDs remain audit evidence.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
