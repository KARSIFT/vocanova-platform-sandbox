# VOC-131 — Replacement release blocker: consume infra #165 without rewriting VOC-112 evidence: Specification

## Objective and requirement source

Deliver the caller-side pin, fixture, tests, docs, and evidence that consume
the exact authoritative infrastructure merge
`8ce2b77a09a729e458a9f4cbea1ca26eb114d398` (KARSIFT/karsift-ai-infra#165)
without rewriting VOC-112 evidence. The final caller must restore shared
lifecycle policy after the root caller checkout in both `identify` and
`converge` before task-completion helpers run, keep the two named VOC-112
fixtures byte-identical to the new carrier's `develop` base, preserve the
already-landed #164 contracts, pass independent exact-revision review and all
protected checks, and then complete ordinary develop-to-main promotion for
this package and the VOC-129 promotion that run `33066533397` skipped.

**Requirement source:** [GitHub issue #1052](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1052).
VOC-129 (`specs/changes/VOC-129-replace-exhausted-voc-127-caller-carrier-with-the`)
already merged caller PR [#1046](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1046)
at `429d8c6d49303148ca1cc14dba5f6768a7863346`. VOC-130
(`specs/changes/VOC-130-release-blocker-caller-checkout-deletes-shared`)
adopted the same #165 pin outcome, then exhausted VOC-130-T00 (#1049 / PR
#1051) because both reviewed heads retargeted VOC-112 evidence. This package
is the governed replacement carrier, not a second VOC-129 implementation and
not a third VOC-130 attempt on #1051.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1052)

| Item | Value |
|------|-------|
| Exhausted VOC-130 task | [#1049](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1049) (`VOC-130-T00`) |
| Unmerged VOC-130 carrier | [#1051](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1051) |
| Attempt-1 reviewed head | `a04a41a19176d1322faee1a3365d82153d61af1b` |
| Attempt-1 review | [comment 5438961796](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1051#issuecomment-5438961796) |
| Attempt-2 reviewed head | `e846cc2f62556c2d6616282f2e9c4929c00655e7` |
| Attempt-2 review | [comment 5439141140](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1051#issuecomment-5439141140) |
| Durable exhaustion marker | [issue #1049 comment 5439143793](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1049#issuecomment-5439143793) |
| VOC-129 caller PR | [#1046](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1046) at `429d8c6d49303148ca1cc14dba5f6768a7863346` |
| Original failing/no-op release run | [`33066533397`](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/33066533397) |
| Selected checkout ref | `develop` |
| Failure | `karsift-ai-infra/config/task-completion-runner.py` missing after caller root checkout |
| Release result | missing validator treated as safe no-op; no audit/promotion; converge skipped |
| Authoritative infra merge | `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` (#165) |
| Independently reviewed infra head | `e33931d02f7bdbb094ae8177fd88324cd19ac5ce` |
| Current `develop` pin | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (#164) |
| VOC-112 `subject_revision` at current `develop` | `f9d11e232a07c7d7a9c433d02c9267912543ba10` |
| Why VOC-130 cannot retry on #1051 | both allowed attempts exhausted; both reviewed heads retained out-of-scope VOC-112 retargets |
| Why VOC-129 is not retried | #1046 already merged; this is a checkout-lifetime pin plus a VOC-112 identity constraint |

## Scope and non-goals

### In scope

1. Set `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` and every
   live pin assertion (caller governance tests and matching
   `scripts/foundation/*` pin literals) to
   `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`.
2. Mirror from exact infra merge #165 only the two changed authoritative
   files into the caller fixture, byte-for-byte:
   `.github/workflows/release.yml` and `tests/test_release_policy.py`.
3. Add deterministic caller regressions for (a) exact pin consistency with
   #165 and inequality with `863fc1f…`, (b) restore ordering in both jobs
   using `job.workflow_repository` / `job.workflow_sha` / `path:
   karsift-ai-infra` with `persist-credentials: false`, and (c) VOC-112
   fixture byte-identity with the carrier's `develop` base.
4. Preserve VOC-113 through VOC-130 fail-closed contracts already on the
   caller: #164 missing-`develop` recovery, branch exact-SHA
   synchronization, unique-develop fail-closed behavior, promotion checks,
   release serialization, roster markers, required-check recovery, independent
   review, retry caps, App-token isolation, Cursor Composer implementer,
   Cursor Grok review, unchanged `config/roles.yml`, no OpenAI route.
5. Update current-state fixture README / comments that would otherwise still
   name the #164 pin or omit the post-caller-checkout restore. Do not edit
   `AGENTS.md` or `.agents/skills/vocanova-repo-navigator/SKILL.md` for this
   pin; those files are hashed by the VOC-112 fixtures that must stay
   unchanged.
6. After the exact reviewed caller pin is merged, complete ordinary
   develop-to-main promotion, exact develop synchronization, production
   deployment where selected, and audit reconciliation for VOC-129, VOC-130,
   and this package. This package's implementation PR `Closes` only its own
   VOC-131 task issue.

### Non-goals / explicitly excluded

- Re-implementing or replacing already-merged infrastructure #165.
- Opening a new `KARSIFT/karsift-ai-infra` PR unless a later exact-revision
  review proves the caller cannot consume `8ce2b77…` as merged. That
  exception is not expected and is not this package's design.
- Reusing, merging, modifying, or composing PR #1051 or its branch.
- Dispatching VOC-130-T00 as attempt `3`, resetting attempt `2` to attempt
  `1`, or treating this package as a third VOC-130 implementation attempt on
  the exhausted carrier.
- Re-implementing VOC-129, re-merging PR #1046, or manufacturing a VOC-129
  or VOC-130 completion marker.
- Retargeting, recapturing, or otherwise changing
  `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json` or
  `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`. Their
  `subject_revision` values must remain
  `f9d11e232a07c7d7a9c433d02c9267912543ba10`. The implementation PR diff must
  not contain either path.
- Editing `AGENTS.md` or `.agents/skills/vocanova-repo-navigator/SKILL.md`
  as a way to "keep VOC-112 hashes current."
- Snapshotting the current develop/main gap as a later drift gate
  (`karsift-ai-infra#15`).
- Fast-forwarding `main` instead of preserving `--merge` promotion commits.
- Normalizing direct-to-main as the ordinary workflow.
- Operator-typed SHA inputs; overloading `existing_pr_number`.
- Changing `config/roles.yml`, adding OpenAI execution, or requesting
  `OPENAI_API_KEY`.
- Weakening exact-SHA review, risk floors, protected checks, App-token
  isolation, force-with-lease, retry caps, unique-develop fail-closed
  behavior, or fail-closed missing-binding behavior.
- Implementing VOC-122 promotion-recovery replan, merging unrelated open
  carriers, or rotating `KARSIFT_BOT_*` secrets / App installation
  permissions.
- A supervised bootstrap exception.
- Rewriting historical CHANGELOG, A-003, VOC-075, VOC-127, VOC-129, or
  VOC-130 audit records except for a new current-state note where a live
  contract would otherwise stay false.
- Changing application runtime behavior, deployment topology, product
  permissions, or monitor inventory.
- Splitting workflow logic, tests, docs, fixture/pin, VOC-112 identity
  regression, or evidence into separate tasks.
- Self-adoption or self-authorization of this package.
- Operator-owned live-evidence contracts: acceptance is deterministic tests
  plus exact-SHA review. VOC-129, VOC-130, and VOC-131 promotion/closure are
  ordinary release-path evidence, not a VOC-097 live-evidence gate.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: caller `tooling/governance/` fixtures and tests; live
  `.github/workflows/*` only if implementation proves a caller dispatch file
  must change to consume #165; current-state fixture README / pin comments.
  VOC-112 evidence fixtures are protected *against* change: they must not
  appear in the implementation diff.
- Protected technical effect: whether the caller fixture and pin equal the
  authoritative #165 merge; whether both release jobs restore shared policy
  after caller checkout before task-completion helpers; whether VOC-112
  capture identity remains the published `develop` bytes. No application
  runtime effect is intended for ordinary promotions.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but workflow and governance-fixture changes still
  require exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-131-D00`: This is one outcome-sized caller replacement. Use one
end-to-end implementation task covering fixture/pin, the two #165 files,
restore coverage, VOC-112 identity regression, tests, docs, and evidence.
Coordinated infrastructure work is already merged as #165; this task consumes
that exact merge rather than re-implementing it. Repository count, file
count, and workflow-versus-tests-versus-docs are not split reasons. VOC-129
and exhausted VOC-130 promotion/closure after the repaired path is live are
evidence of this outcome, not additional VOC-131 tasks and not a VOC-130
attempt `3`.

`VOC-131-D01`: Consume already-merged infrastructure #165. Do not compose a
replacement infra PR. Do not reuse, merge, modify, or compose PR #1051 or
its branch. Do not treat the untracked local `karsift-ai-infra/` checkout
(if present) as this repository's tracked tree. Start a fresh VOC-131
carrier from current `develop`.

`VOC-131-D02`: Pin and fixture identity is exact merge
`8ce2b77a09a729e458a9f4cbea1ca26eb114d398`. `PINNED_SHA.txt` and every live
pin assertion must equal that SHA. They must not equal
`863fc1f35b1d35e4981a59166b0e939be1a2b681`. Mirror from that merge only the
two changed authoritative files, byte-for-byte:
`tooling/governance/fixtures/karsift-ai-infra/.github/workflows/release.yml`
and
`tooling/governance/fixtures/karsift-ai-infra/tests/test_release_policy.py`.
If exact comparison proves another #165 file differs from the current caller
fixture, record that finding in `t00-evidence.md` rather than silently
copying extra files; do not copy unrelated files merely to force a pin.
Consumption of the restore contract in those two files is required.

`VOC-131-D03`: Both `identify` and `converge` must restore shared lifecycle
policy after `Checkout caller release state` and before
`task-completion-runner.py` (`validate-task` in identify, `validate-roster`
in converge). The restore checkout must use
`repository: ${{ job.workflow_repository }}`,
`ref: ${{ job.workflow_sha }}`, `path: karsift-ai-infra`, and
`persist-credentials: false`.

`VOC-131-D04`: Preserve the already-approved #164 behavioral contract in the
mirrored fixture and live caller docs: absent-`develop` fallback / checkout-ref
ordering; after `--merge`, `develop` advances to `mergeCommit.oid` before
audit close; `reconcile-release` retries that sync; unique develop commits,
moved `main`, malformed merges, and missing/unbound refs fail closed;
auto-deleted `develop` is recreated at the merge SHA; equal tips do not open
a second promotion PR; tree-equivalent sync must not keep staging scheduled;
live `reconcile-production-change` remains the exceptional main-only identity.

`VOC-131-D05`: Deterministic tests must prove:

1. `PINNED_SHA.txt` equals `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` and does
   not equal `863fc1f35b1d35e4981a59166b0e939be1a2b681`;
2. matching foundation pin literals equal the same SHA;
3. fixture `release.yml` and `tests/test_release_policy.py` are byte-identical
   to those paths at infra merge `8ce2b77…`;
4. fixture `release.yml` contains exactly two `Restore shared lifecycle policy
   after caller checkout` steps, one in `identify` and one in `converge`;
5. in each job, caller checkout precedes restore, and restore precedes the
   task-completion helper;
6. restore uses `job.workflow_repository`, `job.workflow_sha`,
   `path: karsift-ai-infra`, and `persist-credentials: false`;
7. the #164 checkout-ref ordering / missing-`develop` fallback remains;
8. fixture `release.yml` still binds develop to `mergeCommit.oid` and does not
   restore `CHECKED_HEAD_SHA` as success;
9. live `pipeline.yml` still exposes `reconcile-production-change`, stays at
   most 25 `workflow_dispatch` inputs, and does not expose operator SHA
   inputs;
10. `config/roles.yml` is unchanged and no OpenAI route is added;
11. the two named VOC-112 fixtures are byte-identical to the carrier's
    `develop` base, keep `subject_revision`
    `f9d11e232a07c7d7a9c433d02c9267912543ba10`, and do not appear in
    `git diff` against that base;
12. this package's caller PR `Closes` only its own VOC-131 task issue.

Tests must not mint real App tokens, use secrets, or use production data.

`VOC-131-D06`: Preserve VOC-113 through VOC-130 fail-closed contracts: roster
markers over issue-closed state, exact-head promotion checks, ruleset
attestation, required-check recovery, App-token mutation isolation, job-token
recovery reads, force-with-lease / match-head-commit, two-attempt implementer
bound, nested isolation, credential-free bundles, Cursor Composer
implementer, Cursor Grok exact-revision review, no OpenAI route, `roles.yml`
unchanged, no secrets in bundles/logs/fixtures, no credential values printed,
sanitized raw-error controls. This package's caller PR `Closes` only its own
VOC-131 task issue.

`VOC-131-D07`: Current-state comments in fixture `release.yml` / README that
would otherwise still name pin `863fc1f…` or omit the post-caller-checkout
restore must be updated in the same task. Historical amendment records stay
unchanged. Do not edit `AGENTS.md` or the navigator skill to satisfy VOC-112
hash checks; those hashed sources plus the two VOC-112 fixtures are out of
scope.

`VOC-131-D08`: VOC-129 caller PR #1046 remains the VOC-129 carrier. Exhausted
VOC-130-T00 (#1049 / PR #1051) is not retried. Do not manufacture a VOC-129
or VOC-130 completion marker. After the repaired `release.yml@main` path is
live and this package's exact reviewed merge is promoted:

1. VOC-129 promotion that run `33066533397` skipped proceeds through ordinary
   release evaluation or `reconcile-release` when a valid App-authored
   completion marker exists;
2. this package promotes through the same repaired path;
3. audit comments name the exact promotion merges and close VOC-129,
   VOC-130, and this replacement with precise links. Do not snapshot the
   current develop/main gap. Do not treat closed state alone as completion
   proof.

`VOC-131-D09`: No bootstrap exception. VOC-124 already requested
`permission-workflows: write` on `publish-source`. T00's first run is attempt
`1` on a new VOC-131 carrier from current `develop`. Do not treat an
untracked local `karsift-ai-infra/` checkout as this repository's tracked
tree. Do not treat infrastructure merge `863fc1f…` as the pin target. Do not
treat PR #1051 as a publishable or composable source.

`VOC-131-D10`: The two named VOC-112 fixtures must remain byte-for-byte
identical to the new carrier's `develop` base:

- `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`
- `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`

At current `develop` their `subject_revision` values remain
`f9d11e232a07c7d7a9c433d02c9267912543ba10`. Do not retarget or recapture
them, including as a "provenance fix" for `pr-ancestry` mode. Changing
`voc112-navigation-benchmark-traces.json` also switches
`repository-governance.yml` into `pr-ancestry` mode; this package must not
take that path. A deterministic regression must fail if either file differs
from the adopted package / carrier `develop` base.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. The #165 restore uses
the existing reusable-workflow checkout (`job.workflow_repository` /
`job.workflow_sha`) with `persist-credentials: false`. It must not broaden
installation permissions, must not grant the model-controlled runner the App
token, and must not accept free-form SHAs as authority.

Abuse/process risks:

1. Pinning #164 (`863fc1f…`) after #165 merged — forbidden by `VOC-131-D02`.
2. Covering only `identify` and leaving `converge` unrestored — forbidden by
   `VOC-131-D03`.
3. Reusing PR #1051 or dispatching VOC-130 attempt `3` — forbidden by
   `VOC-131-D01` and `VOC-131-D09`.
4. Re-implementing VOC-129, manufacturing a VOC-129 or VOC-130 completion
   marker, or snapshotting the develop/main gap — forbidden by
   `VOC-131-D08` and `karsift-ai-infra#15`.
5. Retargeting VOC-112 evidence to make provenance tests pass — forbidden by
   `VOC-131-D10`. That is the VOC-130 exhaustion cause.
6. Erasing unique develop commits — still forbidden by VOC-127-D02, preserved
   in the #164/#165 fixture.
7. Treating a missing task-completion helper as a successful no-op after this
   repair — the restore must make the helper present; tests must fail if it
   can still disappear before use.
8. Printing App tokens, private keys, or secret values — forbidden.

## Contradictions and open questions

1. **Live caller workflow edit (`VOC-131-DEP-07`):** required behavior is
   settled (pin/fixture equal #165; both jobs restore shared policy). Whether
   any live `.github/workflows/*.yml` file besides the fixture must change is
   proven at implementation time. Expected: no live `pipeline.yml` edit
   because the caller already calls `release.yml@main`.
2. **VOC-129 promotion timing:** hosted `release.yml@main` may already contain
   the #165 restore before this caller pin merges. `reconcile-release` for
   VOC-129 may therefore succeed independently of this package's merge. That
   does not remove this package's pin/test/doc obligation. Do not block
   VOC-129 promotion on this fixture sync if the live reusable workflow
   already has the repair, and do not treat VOC-129 promotion as this
   package's implementation diff.
3. **Fixture pin applicability:** consumption of the two named #165 files is
   required. Pin to `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`. If exact
   comparison proves another #165 file differs from the current caller
   fixture, record that as implementation evidence rather than silently
   expanding the mirror set; do not copy unrelated files merely to force a
   pin.
4. **AGENTS.md / DOC-15 / DOC-16:** do not update those hashed or historical
   documents to name pin `8ce2b77…` if doing so would require VOC-112
   recapture. Put the current-state pin sentence in fixture README and this
   package's evidence. Historical correction notes stay unchanged.
5. **Infrastructure already merged:** this package must not snapshot, re-open,
   or rewrite #165. If later exact-revision review finds the caller cannot
   consume `8ce2b77…`, record that as a blocking finding rather than silently
   forking infrastructure inside the caller.
6. **PR #1051:** the pin/restore portion of that unmerged carrier is not
   authority to cherry-pick. A fresh carrier from current `develop` is
   required so the VOC-112 files cannot ride along from the exhausted heads.
