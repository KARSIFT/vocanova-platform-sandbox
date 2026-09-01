# VOC-143 — Test Plan

## VOC-143-TEST-00 — `squash-safe-push` accepts historical fixture `agents_sha256` when working-tree `AGENTS.md` differs

- Covers: `VOC-143-AC-00`
- Preconditions: live `assertCapturedRevision` callable; historical VOC-112
  fixture unmodified; current working-tree `AGENTS.md` hash differs from
  fixture `agents_sha256` (the VOC-142 documentation-update class); no secrets
  or production data
- Procedure: Invoke `assertCapturedRevision` on the stored
  `voc112-navigation-benchmark-traces.json` evidence under
  `VOC112_CAPTURE_PROVENANCE_MODE=squash-safe-push`. Use `PR_HEAD_SHA` when a
  40-character SHA is available, otherwise `HEAD`, as the walk tip. Assert the
  call succeeds. Assert fixture `agents_sha256` is not equal to
  `sha256("AGENTS.md")` in this repository state. Assert the fixture hash
  equals `AGENTS.md` at the ancestor returned by `resolveFixtureAgentsBase`.
- Expected result: the #1120 working-tree-equality class cannot recur under
  `squash-safe-push`.
- Evidence: `VOC-143-EV-00`

## VOC-143-TEST-01 — `squash-safe-push` fails closed on an unfound or tampered `agents_sha256`

- Covers: `VOC-143-AC-01`
- Preconditions: same harness as TEST-00
- Procedure: Repeat TEST-00 with fixture `agents_sha256` replaced by a
  64-character value that is not `AGENTS.md` at any ancestor of the tip within
  the existing walk bound. Assert `assertCapturedRevision` throws. Assert a
  truncated or malformed hash also fails closed.
- Expected result: skipping working-tree equality does not accept a hash that
  is not an immutable ancestor bind.
- Evidence: `VOC-143-EV-00`

## VOC-143-TEST-02 — `local` still requires working-tree `AGENTS.md` equality

- Covers: `VOC-143-AC-03`
- Preconditions: same stored fixture; current working-tree `AGENTS.md` hash
  differs from fixture `agents_sha256`
- Procedure: Invoke `assertCapturedRevision` under implicit or explicit
  `local` mode with no promotion signal. Assert the call throws because
  fixture `agents_sha256` does not equal `sha256("AGENTS.md")`.
- Expected result: `local` is not weakened.
- Evidence: `VOC-143-EV-00`

## VOC-143-TEST-03 — `pr-ancestry` still requires working-tree and captured-revision `AGENTS.md` binding

- Covers: `VOC-143-AC-03`
- Preconditions: existing VOC-113 `pr-ancestry` harness
- Procedure: Keep the existing positive `pr-ancestry` case that supplies
  current-head hashes. Add or retain a case where historical fixture
  `agents_sha256` differs from working-tree `AGENTS.md` and the captured
  subject is an ancestor of HEAD; assert `pr-ancestry` still fails closed on
  that mismatch. Do not change the non-ancestor rejection.
- Expected result: `pr-ancestry` is not weakened.
- Evidence: `VOC-143-EV-00`

## VOC-143-TEST-04 — Ordinary `pr-validation` still requires merge-base hashes

- Covers: `VOC-143-AC-03`
- Preconditions: existing VOC-139 ordinary-PR harness with distinct
  merge-base and head `AGENTS.md` hashes
- Procedure: Re-run VOC-139-TEST-02 (or equivalent): mode `pr-validation`, no
  promotion signal, `agents_sha256` equal to HEAD rather than merge base.
  Assert the call throws with merge-base anchoring. Assert a case that
  supplies merge-base `agents_sha256` still passes when navigator hashes also
  match the merge-base contract.
- Expected result: ordinary `pr-validation` remains merge-base anchored.
- Evidence: `VOC-143-EV-00`

## VOC-143-TEST-05 — Promotion `pr-validation` accepts historical fixture `agents_sha256` when HEAD `AGENTS.md` differs

- Covers: `VOC-143-AC-02`
- Preconditions: `VOC112_PROMOTION_PR=true`; mode `pr-validation`; exact
  40-character base/head pair where HEAD `AGENTS.md` differs from the stored
  fixture hash and that fixture hash exists on an ancestor of HEAD
- Procedure: Invoke `assertCapturedRevision` on the stored historical fixture
  (or an equivalent clone that keeps historical `agents_sha256` and sets
  navigator hashes to `PR_HEAD_SHA`). Assert the call succeeds. Assert
  navigator hashes remain equal to the skill file at `PR_HEAD_SHA`.
- Expected result: the #1120 `ci / ci` HEAD-equality class cannot recur for
  `AGENTS.md` documentation updates with an unmodified historical fixture.
- Evidence: `VOC-143-EV-00`

## VOC-143-TEST-06 — Promotion still rejects merge-base-only navigator hashes and a non-ancestor base

- Covers: `VOC-143-AC-04`
- Preconditions: existing VOC-139 promotion harness
- Procedure: Re-run VOC-139-TEST-00 (HEAD-matching `agents_sha256` still
  passes), VOC-139-TEST-05 (merge-base-only *pair* still fails closed; expected
  message may name navigator if that assertion fires first after D03), and
  VOC-139-TEST-04 (divergent non-ancestor base still fails closed).
- Expected result: VOC-139 navigator HEAD-binding and ancestry rejection
  remain; D03 does not accept merge-base-only navigator hashes.
- Evidence: `VOC-143-EV-00`

## VOC-143-TEST-07 — Fixtures, mode selection, pin, and current docs are unchanged where required

- Covers: `VOC-143-AC-05`, `VOC-143-AC-06`
- Preconditions: implementation diff against the recorded implementation PR
  base
- Procedure: Assert `scripts/foundation/fixtures/voc112-*.json` are
  byte-identical to the implementation PR base. Assert
  `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` still strips
  to `8993e867640dfb604dec0466c4e0787e68d8e258`. Assert
  `voc114-actions-check-recovery.test.mjs` still requires promotion
  `validate` to select `squash-safe-push` for same-repository `main` ←
  `develop`. Assert reusable `ci.yml` still passes `--promotion-pr` for that
  pair and `--squash-safe-push` only otherwise. Exhaustively search tracked
  source/current docs for working-tree-equality and HEAD-equality claims;
  assert evidence disposes every match as updated, intentionally historical,
  or irrelevant. Assert `git diff` against the implementation PR base does
  not name files under `specs/changes/VOC-142-…/` through
  `specs/changes/VOC-112-…/` except this package. Assert no
  `ensure-voc112-capture-commits` / `hydrate-voc112-git-objects` helper was
  added.
- Expected result: promotion check identity and historical fixtures stay;
  current-state docs match the ancestor-bind contract.
- Evidence: `VOC-143-EV-00`

## VOC-143-TEST-08 — Full suites, classification, and feasible exact-head evidence

- Covers: `VOC-143-AC-07`
- Preconditions: repair tracked and committed; exact implementation PR base
  and head
- Procedure: Run the commands in `implementation-plan.md` with exact
  base/head. Assert `validate-governance.sh`, `classify-change-risk.sh`, the
  VOC-112 suite under each required mode, VOC-114 mode-selection tests, and
  `git diff --check` pass. Assert `t00-evidence.md` names the implementation
  PR base, states that the live head is bound by the App-authored
  independent-review comment/check, and does not require a tracked file in
  the same commit to equal `HEAD`. Assert fixture `roles.yml` is unchanged.
  Assert no App ID/private-key secret changed. Assert classification is R3
  unless `repository-governance.yml` was edited (then R4 path floor).
- Expected result: hosted-equivalent deterministic checks pass; exact-head
  binding remains Git-feasible; fail-closed boundaries remain.
- Evidence: `VOC-143-EV-00`

## VOC-143-TEST-09 — No snapshot-gap; `reconcile-release` of #1118 can proceed after T00

- Covers: `VOC-143-AC-08`
- Preconditions: this package's implementation PR and `t00-evidence.md`
- Procedure: Assert no develop/main gap snapshot was committed. Assert no
  duplicate promotion-PR or release-audit creation helper was added. Assert
  no helper merges, closes, or recreates #1119. Assert evidence states that
  after this correction merges, ordinary `reconcile-release` for release
  issue #1118 may re-evaluate #1119 when it still matches, wait for complete
  required checks, merge, clean up, and close the audit. Assert incident PR
  #1119, head `376e00dd…`, and audit #1118 are preserved as audit evidence,
  not restaged as this package's implementation work. Assert closed state is
  not treated as completion proof. Assert root issue #1120 is documented to
  close only after allowlisted metadata from a successful recovery/release
  run exists.
- Expected result: repair is provenance-assertion plus ordinary release
  handoff, not a snapshot-then-check package and not a manual merge of
  #1119.
- Evidence: `VOC-143-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
