# VOC-143 — Promotion provenance validation rejects legitimate AGENTS.md change under squash-safe-push: Specification

## Objective and requirement source

Stop promotion-path VOC-112 capture provenance from rejecting a legitimate
current `AGENTS.md` documentation update while the historical benchmark
fixture remains intentionally unmodified. Keep `local`, `pr-ancestry`, and
ordinary merge-base-anchored `pr-validation` fail-closed. Do not recapture
fixtures, switch promotion check identity, or bypass required checks.

**Requirement source:** [GitHub issue #1120](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1120).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1120)

| Item | Value |
|------|-------|
| Repository | `KARSIFT/vocanova-platform-sandbox` |
| Promotion PR | [#1119](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1119) |
| Exact promotion head | `376e00dd769253d7a255660f5391fb208781e2f3` |
| Failing required checks | `validate`, `ci / ci` |
| Release audit | [#1118](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1118), open |
| Prior causal package | VOC-142 changed `AGENTS.md` and correctly did not recapture VOC-112 fixtures (`VOC-142-DEP-07`) |
| Live pin | `8993e867640dfb604dec0466c4e0787e68d8e258` (not this defect) |

Reproduction from the issue:

1. Merge an otherwise governed implementation PR that changes `AGENTS.md`
   without recapturing the historical VOC-112 benchmark fixture.
2. Let automatic `develop` → `main` promotion run.
3. Run `node --test scripts/foundation/voc112-navigation-benchmark.test.mjs`
   in that promotion validation context.

Expected: `squash-safe-push` validates the historical capture against its
correct immutable provenance anchor and permits an unrelated current
`AGENTS.md` documentation update.

Actual: the benchmark still requires the fixture historical `agents_sha256` to
equal the current working-tree `AGENTS.md` hash.

Live caller contract at issue creation:

- `assertCapturedRevision` in
  `scripts/foundation/voc112-navigation-benchmark.test.mjs` skips
  captured-revision `AGENTS.md` binding for `squash-safe-push` and
  `pr-validation`, then still requires
  `evidence.source_hashes.agents_sha256 == sha256("AGENTS.md")` unless mode is
  exactly `pr-validation`.
- `.github/workflows/repository-governance.yml` selects `squash-safe-push` for
  same-repository `main` ← `develop` promotion PRs and for push/non-PR events.
  On pull_request it still exports `PR_BASE_SHA` / `PR_HEAD_SHA`.
- Reusable `ci.yml` uses `--promotion-pr` for that same promotion pair, so
  `ci / ci` runs `pr-validation` with `VOC112_PROMOTION_PR=true`.
  `assertPrValidationMergeBase` then requires fixture `agents_sha256` to equal
  `AGENTS.md` at `PR_HEAD_SHA`.
- `resolveFixtureAgentsBase` already walks from a tip to the ancestor whose
  `AGENTS.md` matches the fixture hash (used today to reconstruct a
  merge-base for ordinary `pr-validation` tests).

## Scope and non-goals

### In scope

1. `squash-safe-push` must bind fixture `agents_sha256` to an immutable Git
   ancestor of the validation tip whose `AGENTS.md` bytes match the fixture
   hash. The current working tree may differ. A hash that cannot be found in
   that ancestry fails closed.
2. Promotion `pr-validation` (`VOC112_PROMOTION_PR=true`) must use the same
   historical-ancestor bind for `agents_sha256` so `ci / ci` on a promotion
   PR also accepts an unmodified historical fixture after a later `AGENTS.md`
   documentation update. Navigator skill hashes remain HEAD-bound on
   promotion PRs.
3. Ordinary (non-promotion) `pr-validation` remains merge-base anchored for
   both hashed sources. `local` and `pr-ancestry` retain working-tree
   `AGENTS.md` equality.
4. Deterministic regressions for: `squash-safe-push` with changed working-tree
   `AGENTS.md` and unchanged historical fixture; tampered/unfound fixture
   hash fail-closed; unchanged `local` / `pr-ancestry` / ordinary
   `pr-validation`; promotion `pr-validation` historical-fixture success;
   retained VOC-139 navigator HEAD-binding and non-ancestor-base rejection.
5. Exhaustive current-state source/doc search and updates so no live document
   claims `squash-safe-push` requires working-tree `AGENTS.md` equality, or
   that promotion `pr-validation` requires fixture `agents_sha256` to equal
   current HEAD `AGENTS.md` after a legitimate documentation update.
6. After exact-SHA review and merge, ordinary `reconcile-release` for #1118
   must be able to re-evaluate the live same-repository promotion. Closure of
   #1120 binds allowlisted metadata from a successful recovery/release run.

### Non-goals / explicitly excluded

- Recapturing or editing VOC-112 JSON fixtures or hashed sources.
- Switching the promotion PR application check (`ci.yml`) to
  `--squash-safe-push`.
- Switching `repository-governance.yml` promotion validate away from
  `squash-safe-push`.
- Weakening `local`, `pr-ancestry`, or ordinary merge-base-anchored
  `pr-validation`.
- Weakening navigator HEAD/working-tree binding.
- Adding fetch/hydrate/recapture helpers, or changing `package.json`.
- Pin-advance, mirrored fixture replacement, or a coordinated
  `karsift-ai-infra` PR unless implementation proves an infra contract change
  is required.
- Editing `.github/workflows/repository-governance.yml` unless SHA passing is
  proven insufficient (that path is an R4 floor).
- Weakening the production merge guard, adding bypass actors, fabricating
  statuses, or manually merging #1119.
- Snapshotting the current develop/main gap (`karsift-ai-infra#15`).
- A VOC-097 operator-owned live-evidence second task.
- Changing `config/roles.yml`, adding OpenAI execution, or requesting
  `OPENAI_API_KEY`.
- Rewriting VOC-142, VOC-141, VOC-140, VOC-139, VOC-138, VOC-114, VOC-113, or
  VOC-112 package records under `specs/changes/`.
- Application runtime, deployment topology, credential-value, provider, or
  monitor-inventory changes.
- Self-adoption or self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R3**.
- Protected areas: VOC-112 capture provenance used by required `validate` and
  `ci / ci` before develop-to-main promotion. Path floor for
  `scripts/foundation/*.mjs` is R1; semantic CI/CD / governance-enforcement
  effect is R3. Editing `repository-governance.yml` would raise the path floor
  to R4 and is out of scope unless SHA passing is proven insufficient.
- Protected technical effect: whether promotion-path required checks accept a
  historical VOC-112 fixture after a later `AGENTS.md` documentation update.
  No application runtime effect is intended.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but provenance used as a required check still
  requires exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-143-D00`: This is one outcome-sized provenance repair. Use one
end-to-end implementation task covering `squash-safe-push` ancestor binding,
promotion `pr-validation` `AGENTS.md` ancestor binding, tests, current-state
docs, evidence, and release handoff. Validate versus `ci / ci`, tests versus
docs, and later `reconcile-release` of #1118 are not split reasons. Later
promotion of this repair is evidence of the outcome, not a second task and
not a snapshot of the develop/main gap.

`VOC-143-D01`: Under `VOC112_CAPTURE_PROVENANCE_MODE=squash-safe-push`, do
not require fixture `agents_sha256` to equal the current working-tree
`AGENTS.md` hash. Bind that hash to an immutable Git ancestor of the
validation tip (`PR_HEAD_SHA` when it is a 40-character SHA, otherwise
`HEAD`) using the existing `resolveFixtureAgentsBase` walk (or equivalent).
The class of promotion `validate` failure on PR #1119 / head `376e00dd…`
cannot recur for a legitimate `AGENTS.md` documentation update with an
unmodified historical fixture.

`VOC-143-D02`: A fixture `agents_sha256` that cannot be found as `AGENTS.md`
at an ancestor of the validation tip within the existing walk bound fails
closed. Tampered, truncated, or unrelated hashes must not pass
`squash-safe-push`. Navigator skill hashes remain bound to the working tree
in this mode unless they already have a separate immutable contract; this
package does not recapture or retarget navigator hashes.

`VOC-143-D03`: Promotion `pr-validation` (`VOC112_PROMOTION_PR=true`) must
bind fixture `agents_sha256` to the same historical ancestor of `PR_HEAD_SHA`,
not to current HEAD `AGENTS.md` bytes. This is required for `ci / ci` on the
same promotion PR. VOC-139 tests that supply `agents_sha256` equal to HEAD
continue to pass because HEAD is an ancestor of itself. Navigator skill
hashes remain anchored at `PR_HEAD_SHA`. Non-ancestor promotion bases still
fail closed.

`VOC-143-D04`: Do not weaken `local`, `pr-ancestry`, or ordinary
(non-promotion) `pr-validation`. `local` and `pr-ancestry` retain working-tree
`AGENTS.md` equality. Ordinary `pr-validation` retains merge-base anchoring
for both hashed sources. A unit test that only skips the working-tree
equality line without an ancestor bind is not sufficient coverage of #1120.

`VOC-143-D05`: Do not change promotion check identity. Keep
`repository-governance.yml` selecting `squash-safe-push` for same-repository
`main` ← `develop` promotion PRs (`VOC-114`). Keep reusable `ci.yml` using
`--promotion-pr` / `pr-validation` for that pair (`VOC-138` / `VOC-139`). Do
not treat switching either check's mode as the fix. Do not bypass, skip, or
rename required `validate` or `ci / ci`.

`VOC-143-D06`: Do not recapture or edit
`scripts/foundation/fixtures/voc112-*.json`, the navigator skill,
`package.json`, or VOC-112 hashed sources. Do not add fetch/hydrate helpers.
`PINNED_SHA.txt` stays at `8993e867640dfb604dec0466c4e0787e68d8e258` unless
implementation proves an infra contract change is required. Prefer not to
edit `.github/workflows/repository-governance.yml`; SHAs are already exported
on pull_request. An untracked local `karsift-ai-infra/` checkout is not this
repository's tracked tree.

`VOC-143-D07`: Tests must exercise the live `assertCapturedRevision` path, not
only comments. Include at least: `squash-safe-push` with current working-tree
`AGENTS.md` different from the historical fixture hash and the fixture hash
present on an ancestor; `squash-safe-push` fail-closed on an unfound/tampered
`agents_sha256`; `local` still requiring working-tree equality; `pr-ancestry`
retention; ordinary `pr-validation` still requiring merge-base hashes;
promotion `pr-validation` accepting historical fixture `agents_sha256` when
HEAD `AGENTS.md` differs; promotion still rejecting merge-base-only navigator
hashes and a non-ancestor base. Existing VOC-139 test expected messages may
be updated only as required by D03 (navigator assertion may fire first).

`VOC-143-D08`: Docs in the same PR. Before editing, exhaustively search
tracked source and current documentation for claims that `squash-safe-push`
requires working-tree `AGENTS.md` equality, that promotion `pr-validation`
requires fixture `agents_sha256` to equal current HEAD `AGENTS.md` after a
documentation update, and for VOC-142's "other modes retain working-tree
binding" wording where it would remain live-false. Record searched patterns
and path disposition in `t00-evidence.md`. Update every current-state
document that would otherwise remain false, including
`docs/operations/11-devops-and-ci-cd.md` if it omits the `validate`
`squash-safe-push` contract or the historical-ancestor `AGENTS.md` bind.
Preserve clearly labeled historical VOC-142/VOC-139/VOC-114 records. Do not
rewrite those package directories.

`VOC-143-D09`: No VOC-097 live-evidence second task, no snapshot-gap task, and
no manual merge of #1119. After this repair is live, ordinary
`gh workflow run pipeline.yml --ref develop -f action=reconcile-release -f release_issue_number=1118`
must be able to re-evaluate the live same-repository promotion. Do not
snapshot the develop/main gap (`karsift-ai-infra#15`). Do not create a
duplicate promotion PR or release audit. Live success after this repair is
ordinary release closure evidence recorded with allowlisted metadata only.

`VOC-143-D10`: Roles and credentials. Fixture `config/roles.yml` remains
implementer / `implementer_escalation` `cursor/composer-2.5` and planner /
reviewer / `reviewer_fast_retry` / `plan_reviewer`
`cursor/grok-4.6[effort=high,fast=false]`. No OpenAI route. No
`OPENAI_API_KEY` request. Do not print credential values. Preserve promotion
PR #1119, head `376e00dd…`, and audit #1118 as audit evidence only (no raw
logs).

`VOC-143-D11`: Validation after the repair is tracked and committed:

- `bash scripts/governance/validate-governance.sh` with exact base/head;
- `bash scripts/governance/classify-change-risk.sh` with exact base/head
  (expect at least the path floor; builder/verifier class is R3 unless
  `repository-governance.yml` is edited);
- `node --test scripts/foundation/voc112-navigation-benchmark.test.mjs`
  under `local`, `pr-ancestry`, `pr-validation`, promotion `pr-validation`,
  and `squash-safe-push`;
- `node --test scripts/foundation/voc114-actions-check-recovery.test.mjs`
  (mode-selection contract unchanged);
- `git diff --check`.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

`VOC-143-D12`: Feasible exact-revision evidence. The App-authored
independent-review comment/check must bind the live PR head exactly and must
explicitly evaluate: `squash-safe-push` historical-ancestor `AGENTS.md` bind;
promotion `pr-validation` historical-ancestor `AGENTS.md` bind; retained
`local` / `pr-ancestry` / ordinary `pr-validation`; retained navigator
binding; unchanged promotion check identity; unre-captured fixtures; and
current-state docs. Merge-gate must reject any mismatch. Committed
`t00-evidence.md` records the implementation PR base and the contract that
later exact-head binding is published as review/check metadata. A tracked
file must not be required to contain the SHA of the same commit that contains
it.

`VOC-143-D13`: Protected comparison versus implementation PR base.
Issue-creation promotion head is `376e00dd769253d7a255660f5391fb208781e2f3`.
Implementation must resolve current `develop` to a 40-character SHA before
any in-scope edit and record that SHA as the implementation PR base. Fail
closed on unrelated/material movement of `develop` (any tree change outside
this package directory, the provenance test, and named current-state docs).
This package's own plan/adoption/roster commits after `376e00dd…` are
governance-only and do not count as protected-file drift. VOC-142 and earlier
package files under `specs/changes/` are out of scope and must not be edited.

`VOC-143-D14`: Release handoff. After the exact reviewed caller merge,
ordinary `reconcile-release` for #1118 may re-evaluate #1119 or the live
same-repository `main` ← `develop` promotion at the then-current `develop`
tip. Ordinary later promotion of this package uses `release.yml` once
required checks pass. `develop` is advanced to the exact promotion merge SHA
before audit close. Every promotion merge push to `main` triggers automatic
production deployment, whose exact-SHA result is verified. Do not snapshot
the current gap. Closed state alone is not completion proof. Root issue
#1120 closes only after allowlisted metadata from a successful
recovery/release run exists. Do not create a duplicate promotion PR or
release audit.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. The repair prevents a
promotion deadlock without letting a tampered fixture hash pass, without
switching promotion check identity, and without skipping protected checks.

Abuse/process risks:

1. Skipping working-tree `AGENTS.md` equality under `squash-safe-push` without
   an immutable ancestor bind — forbidden by `VOC-143-D01` and `VOC-143-D02`.
2. Allowing a fixture hash that is not present on an ancestor of the
   validation tip — forbidden by `VOC-143-D02`.
3. Leaving promotion `pr-validation` HEAD-bound for `agents_sha256`, so
   `ci / ci` on #1119 still fails — forbidden by `VOC-143-D03`.
4. Weakening `local`, `pr-ancestry`, or ordinary merge-base `pr-validation` —
   forbidden by `VOC-143-D04`.
5. Weakening navigator HEAD/working-tree binding — forbidden by
   `VOC-143-D03` and `VOC-143-D04`.
6. Switching promotion `ci.yml` to `--squash-safe-push` or switching
   promotion `validate` away from `squash-safe-push` — forbidden by
   `VOC-143-D05`.
7. Recapturing VOC-112 fixtures or adding fetch/hydrate helpers — forbidden
   by `VOC-143-D06`.
8. Manually merging, closing, recreating, or bypassing #1119 — forbidden by
   `VOC-143-D09` and `VOC-143-DEP-04`.
9. Weakening the production merge guard or adding bypass actors — forbidden
   by `VOC-143-DEP-04`.
10. Snapshotting the develop/main gap or adding a self-invalidating
    snapshot-then-check task — forbidden by `VOC-143-D09` and
    `karsift-ai-infra#15`.
11. Changing `roles.yml` or adding an OpenAI route — forbidden by
    `VOC-143-D10`.
12. Printing credentials or copying full CI logs into evidence — forbidden.
13. Requiring a commit to contain its own SHA — forbidden by `VOC-143-D12`.
14. Pin-advancing or treating the untracked `karsift-ai-infra/` checkout as
    tracked evidence — forbidden by `VOC-143-D06`.

## Contradictions and open questions

1. **VOC-142 exception versus promotion `pr-validation` HEAD bind:** VOC-142
   skipped working-tree `AGENTS.md` equality only for `pr-validation`.
   Promotion `pr-validation` still binds `agents_sha256` to `PR_HEAD_SHA`, so
   `ci / ci` on #1119 fails even if only `squash-safe-push` is fixed. This
   package treats D03 as in-scope causal repair of the same outcome, not as a
   rewrite of VOC-139 package records. VOC-139-TEST-00 remains valid when the
   supplied hash equals HEAD. VOC-139-TEST-05 may need its expected error
   updated if the navigator assertion fires first after D03.
2. **`validate` uses `squash-safe-push`; `ci / ci` uses `--promotion-pr`:**
   VOC-114 selected `squash-safe-push` for promotion `validate`; VOC-138/VOC-139
   require promotion `ci` to stay `pr-validation`. D05 keeps both selections
   and repairs assertions underneath them. Do not "simplify" to one mode.
3. **VOC-139 merge-base rejection versus historical-ancestor bind:** Walking
   from HEAD would eventually find `main`'s `AGENTS.md` if that is where the
   historical hash still lives. Navigator remains HEAD-bound so merge-base-only
   *pair* hashes still fail closed (D03). Implementer must not drop navigator
   HEAD-binding to make TEST-05 pass.
4. **`repository-governance.yml` is an R4 path.** Prefer not to edit it. If
   SHA passing is proven insufficient for the ancestor walk, record that
   exception in `t00-evidence.md` and expect the path classifier to report R4.
5. **Pin `8993e867…` is not this defect.** Do not pin-advance as a substitute
   for the caller assertion repair.
6. **Untracked local `karsift-ai-infra/` checkout:** if present in the
   workspace, it is not this repository's tracked tree and is not
   implementation evidence.
