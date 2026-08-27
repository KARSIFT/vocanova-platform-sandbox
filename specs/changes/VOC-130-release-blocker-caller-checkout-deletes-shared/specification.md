# VOC-130 — Release blocker: caller checkout deletes shared lifecycle policy: Specification

## Objective and requirement source

Deliver the caller-side pin, fixture, tests, docs, and evidence that consume
the exact authoritative infrastructure merge
`8ce2b77a09a729e458a9f4cbea1ca26eb114d398` (KARSIFT/karsift-ai-infra#165). The
final caller must restore shared lifecycle policy after the root caller
checkout in both `identify` and `converge` before task-completion helpers run,
preserve the already-landed #164 contracts, pass independent exact-revision
review and all protected checks, and then complete ordinary develop-to-main
promotion for this package and the VOC-129 promotion that run `33066533397`
skipped.

**Requirement source:** [GitHub issue #1047](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1047).
VOC-129 (`specs/changes/VOC-129-replace-exhausted-voc-127-caller-carrier-with-the`)
already merged caller PR [#1046](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1046)
at `429d8c6d49303148ca1cc14dba5f6768a7863346`. This package is the governed
caller consumption of the #165 checkout-lifetime repair, not a second VOC-129
implementation.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1047)

| Item | Value |
|------|-------|
| VOC-129 caller PR | [#1046](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1046) at `429d8c6d49303148ca1cc14dba5f6768a7863346` |
| Failing/no-op release run | [`33066533397`](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/33066533397) |
| Selected checkout ref | `develop` |
| Failure | `karsift-ai-infra/config/task-completion-runner.py` missing after caller root checkout |
| Release result | missing validator treated as safe no-op; no audit/promotion; converge skipped |
| Authoritative infra merge | `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` (#165) |
| Independently reviewed infra head | `e33931d02f7bdbb094ae8177fd88324cd19ac5ce` |
| Infra verification | 429 policy tests plus hosted actionlint, shellcheck, YAML parsing, and policy checks |
| Current `develop` pin | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (#164) |
| Why VOC-129 is not retried | #1046 already merged; this is a new checkout-lifetime defect, not an unpublished VOC-129 carrier |

## Scope and non-goals

### In scope

1. Set `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` and every
   live pin assertion (caller governance tests and matching
   `scripts/foundation/*` pin literals) to
   `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`.
2. Mirror the in-scope #165 infrastructure files needed by the caller from
   that exact merge, including fixture `.github/workflows/release.yml` and the
   #165 policy tests that prove both `identify` and `converge` restore shared
   policy after caller checkout and before task-completion helpers.
3. Add deterministic caller regressions for (a) exact pin consistency with
   #165 and inequality with `863fc1f…`, and (b) restore ordering in both
   jobs using `job.workflow_repository` / `job.workflow_sha` / `path:
   karsift-ai-infra`.
4. Preserve VOC-113 through VOC-129 fail-closed contracts already on the
   caller: #164 missing-`develop` recovery, branch exact-SHA synchronization,
   unique-develop fail-closed behavior, promotion checks, release
   serialization, roster markers, required-check recovery, independent review,
   retry caps, App-token isolation, Cursor Composer implementer, Cursor Grok
   review, unchanged `config/roles.yml`, no OpenAI route.
5. Update current-state fixture README / comments that would otherwise still
   name the #164 pin or omit the post-caller-checkout restore.
6. After the exact reviewed caller pin is merged, complete ordinary
   develop-to-main promotion, exact develop synchronization, production
   deployment where selected, and audit reconciliation for VOC-129 and this
   package. This package's implementation PR `Closes` only its own VOC-130
   task issue.

### Non-goals / explicitly excluded

- Re-implementing or replacing already-merged infrastructure #165.
- Opening a new `KARSIFT/karsift-ai-infra` PR unless a later exact-revision
  review proves the caller cannot consume `8ce2b77…` as merged. That
  exception is not expected and is not this package's design.
- Re-implementing VOC-129, re-merging PR #1046, or manufacturing a second
  VOC-129 completion marker.
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
- Rewriting historical CHANGELOG, A-003, VOC-075, VOC-127, or VOC-129 audit
  records except for a new current-state note where a live contract would
  otherwise stay false.
- Changing application runtime behavior, deployment topology, product
  permissions, or monitor inventory.
- Splitting workflow logic, tests, docs, fixture/pin, or evidence into
  separate tasks.
- Self-adoption or self-authorization of this package.
- Operator-owned live-evidence contracts: acceptance is deterministic tests
  plus exact-SHA review. VOC-129 and VOC-130 promotion/closure are ordinary
  release-path evidence, not a VOC-097 live-evidence gate.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: caller `tooling/governance/` fixtures and tests; live
  `.github/workflows/*` only if implementation proves a caller dispatch file
  must change to consume #165; current-state fixture README / pin comments.
- Protected technical effect: whether the caller fixture and pin equal the
  authoritative #165 merge; whether both release jobs restore shared policy
  after caller checkout before task-completion helpers. No application
  runtime effect is intended for ordinary promotions.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but workflow and governance-fixture changes still
  require exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-130-D00`: This is one outcome-sized caller pin. Use one end-to-end
implementation task covering fixture/pin, #165 `release.yml` restore, tests,
docs, and evidence. Coordinated infrastructure work is already merged as
#165; this task consumes that exact merge rather than re-implementing it.
Repository count, file count, and workflow-versus-tests-versus-docs are not
split reasons. VOC-129 promotion after the repaired path is live is evidence
of this outcome, not an additional VOC-130 task and not a second VOC-129
roster entry.

`VOC-130-D01`: Consume already-merged infrastructure #165. Do not compose a
replacement infra PR. Do not treat the untracked local `karsift-ai-infra/`
checkout (if present) as this repository's tracked tree.

`VOC-130-D02`: Pin and fixture identity is exact merge
`8ce2b77a09a729e458a9f4cbea1ca26eb114d398`. `PINNED_SHA.txt` and every live
pin assertion must equal that SHA. They must not equal
`863fc1f35b1d35e4981a59166b0e939be1a2b681`. Mirror every in-scope
infrastructure file needed by the caller from that merge, including
`release.yml` and the #165 restore tests. If a file is not in the fixture
subset, do not copy it merely to force a pin; record non-consumption.
Consumption of the restore contract is required.

`VOC-130-D03`: Both `identify` and `converge` must restore shared lifecycle
policy after `Checkout caller release state` and before
`task-completion-runner.py` (`validate-task` in identify, `validate-roster`
in converge). The restore checkout must use
`repository: ${{ job.workflow_repository }}`,
`ref: ${{ job.workflow_sha }}`, and `path: karsift-ai-infra`.

`VOC-130-D04`: Preserve the already-approved #164 behavioral contract in the
mirrored fixture and live caller docs: absent-`develop` fallback / checkout-ref
ordering; after `--merge`, `develop` advances to `mergeCommit.oid` before
audit close; `reconcile-release` retries that sync; unique develop commits,
moved `main`, malformed merges, and missing/unbound refs fail closed;
auto-deleted `develop` is recreated at the merge SHA; equal tips do not open
a second promotion PR; tree-equivalent sync must not keep staging scheduled;
live `reconcile-production-change` remains the exceptional main-only identity.

`VOC-130-D05`: Deterministic tests must prove:

1. `PINNED_SHA.txt` equals `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` and does
   not equal `863fc1f35b1d35e4981a59166b0e939be1a2b681`;
2. matching foundation pin literals equal the same SHA;
3. fixture `release.yml` contains exactly two `Restore shared lifecycle policy
   after caller checkout` steps, one in `identify` and one in `converge`;
4. in each job, caller checkout precedes restore, and restore precedes the
   task-completion helper;
5. restore uses `job.workflow_repository`, `job.workflow_sha`, and
   `path: karsift-ai-infra`;
6. the #164 checkout-ref ordering / missing-`develop` fallback remains;
7. fixture `release.yml` still binds develop to `mergeCommit.oid` and does not
   restore `CHECKED_HEAD_SHA` as success;
8. live `pipeline.yml` still exposes `reconcile-production-change`, stays at
   most 25 `workflow_dispatch` inputs, and does not expose operator SHA
   inputs;
9. `config/roles.yml` is unchanged and no OpenAI route is added;
10. this package's caller PR `Closes` only its own VOC-130 task issue.

Tests must not mint real App tokens, use secrets, or use production data.

`VOC-130-D06`: Preserve VOC-113 through VOC-129 fail-closed contracts: roster
markers over issue-closed state, exact-head promotion checks, ruleset
attestation, required-check recovery, App-token mutation isolation, job-token
recovery reads, force-with-lease / match-head-commit, two-attempt implementer
bound, nested isolation, credential-free bundles, Cursor Composer
implementer, Cursor Grok exact-revision review, no OpenAI route, `roles.yml`
unchanged, no secrets in bundles/logs/fixtures, no credential values printed,
sanitized raw-error controls. This package's caller PR `Closes` only its own
VOC-130 task issue.

`VOC-130-D07`: Current-state comments in fixture `release.yml` / README and
any live caller docs that would otherwise still name pin `863fc1f…` or omit
the post-caller-checkout restore must be updated in the same task. Historical
amendment records stay unchanged.

`VOC-130-D08`: VOC-129 caller PR #1046 remains the VOC-129 carrier. Do not
re-implement VOC-129. After the repaired `release.yml@main` path is live and
this package's exact reviewed merge is promoted:

1. VOC-129 promotion that run `33066533397` skipped proceeds through ordinary
   release evaluation or `reconcile-release` when a valid App-authored
   completion marker exists;
2. this package promotes through the same repaired path;
3. audit comments name both exact promotion merges. Do not snapshot the
   current develop/main gap. Do not treat closed state alone as completion
   proof.

`VOC-130-D09`: No bootstrap exception. VOC-124 already requested
`permission-workflows: write` on `publish-source`. T00's first run is attempt
`1` on a new VOC-130 carrier. Do not treat an untracked local
`karsift-ai-infra/` checkout as this repository's tracked tree. Do not treat
infrastructure merge `863fc1f…` as the pin target.

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

1. Pinning #164 (`863fc1f…`) after #165 merged — forbidden by `VOC-130-D02`.
2. Covering only `identify` and leaving `converge` unrestored — forbidden by
   `VOC-130-D03`.
3. Re-implementing VOC-129 or snapshotting the develop/main gap — forbidden
   by `VOC-130-D08` and `karsift-ai-infra#15`.
4. Erasing unique develop commits — still forbidden by VOC-127-D02, preserved
   in the #164/#165 fixture.
5. Treating a missing task-completion helper as a successful no-op after this
   repair — the restore must make the helper present; tests must fail if it
   can still disappear before use.
6. Printing App tokens, private keys, or secret values — forbidden.

## Contradictions and open questions

1. **Live caller workflow edit (`VOC-130-DEP-07`):** required behavior is
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
3. **Fixture pin applicability:** consumption of `release.yml` restore is
   required. Pin to `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`. If some
   unrelated #165 files are not in the mirrored subset, do not copy them
   merely to force a pin; record non-consumption.
4. **AGENTS.md / DOC-15 / DOC-16:** update current-state sentences that would
   become false. Do not expand them into a new release-internals runbook, and
   do not rewrite historical correction notes.
5. **Infrastructure already merged:** this package must not snapshot, re-open,
   or rewrite #165. If later exact-revision review finds the caller cannot
   consume `8ce2b77…`, record that as a blocking finding rather than silently
   forking infrastructure inside the caller.
