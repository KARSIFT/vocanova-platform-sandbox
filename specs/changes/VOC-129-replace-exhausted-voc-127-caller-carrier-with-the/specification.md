# VOC-129 — Replace exhausted VOC-127 caller carrier with the exact infrastructure #164 fixture and pin: Specification

## Objective and requirement source

Deliver the remaining VOC-127 caller-side release-synchronization contract
through one governed replacement carrier. The final caller must consume the
exact authoritative infrastructure merge
`863fc1f35b1d35e4981a59166b0e939be1a2b681` (KARSIFT/karsift-ai-infra#164), pass
independent exact-revision review and all protected checks, supersede
unpublishable PR #1041 without treating the replacement as a third VOC-127
implementation attempt, and preserve the original VOC-127 audit trail.

**Requirement source:** [GitHub issue #1042](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1042).
The original VOC-127 outcome
([issue #1035](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1035),
package `specs/changes/VOC-127-converge-develop-to-the-exact-promotion-merge-sha`)
remains the approved release-synchronization contract. This package is the
caller replacement after VOC-127-T00 exhausted attempt `2/2`.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1042)

| Item | Value |
|------|-------|
| Exhausted VOC-127 task | [#1039](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1039) (`VOC-127-T00`) |
| Origin requirement | [#1035](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1035) |
| Unpublishable caller PR | [#1041](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1041) at `9d459813a733f9c6d58ad3352df0db27d33ee7f4` |
| Attempt-2/2 pipeline run | `33058603158` |
| Attempt-2 artifact | `implement-VOC-127-VOC-127-T00-attempt2` (id `9640920826`) |
| Artifact bundle SHA-256 | `724e5547e29283b2701b70b72e000253a5100edf77545807610fce17151f7906` |
| Artifact exact head | `bbdb93aadec461830435771490f5c79ba524fed9` |
| Stale pin on that artifact | `a9df74a63976d5239b84151fd01310835c999e7c` |
| Authoritative infra merge | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (#164) |
| Current `develop` pin | `60afda3a44fd06b8c00b219771de7112f1aded6e` |
| Why VOC-127 cannot retry | final allowed implementation retry already consumed; attempt `3` is forbidden |
| Why #1041 cannot be the carrier | publisher refused workflow-file publication; pin/evidence still name #163 |

## Scope and non-goals

### In scope

1. Recreate the complete intended VOC-127 caller diff from current `develop`:
   live `pipeline.yml` exceptional production-change dispatch, staging
   path-selection proof or narrow skip, current-state docs, caller fixture,
   pin, tests, and evidence.
2. Mirror every in-scope infrastructure file needed by the caller from exact
   merge `863fc1f35b1d35e4981a59166b0e939be1a2b681`, including `release.yml`,
   `reconcile-production-change.yml`, `release-checkout-ref-runner.py`,
   `branch_sync.py` / `branch-sync-runner.py`, project-repo `pipeline.yml`
   template, and the #164 policy/helper tests those files require.
3. Set `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` and every
   live pin assertion (caller governance tests and matching
   `scripts/foundation/*` pin literals) to
   `863fc1f35b1d35e4981a59166b0e939be1a2b681`.
4. Consume the live #164 identity `reconcile-production-change` on caller
   `pipeline.yml`. Wire `release` and `auto-advance` to wait on that job the
   same way the #164 project-repo template does. Do not add operator-typed SHA
   inputs. Keep every `workflow_dispatch` block at most 25 inputs. Preserve
   live `live_evidence_mode` rather than drive-by normalizing it to the
   template's `live_evidence_dispatch`.
5. Add deterministic caller regressions for (a) exact pin consistency with
   #164 and inequality with `a9df74a6…` / `60afda3a…`, and (b) the #164
   checkout-ref ordering / missing-`develop` fallback path present in the
   mirrored fixture.
6. Preserve VOC-113 through VOC-127 fail-closed contracts already on the
   caller: roster markers, required-check recovery, independent review, retry
   caps, App-token isolation, Cursor Composer implementer, Cursor Grok review,
   unchanged `config/roles.yml`, no OpenAI route.
7. Update current-state docs that would otherwise still omit exact-merge-SHA
   develop sync, `reconcile-release` retry of that sync, or exceptional
   governed production-change reconciliation.
8. After the exact reviewed caller replacement is merged and promoted, close
   #1041 as superseded (never merged), then close #1039 and root #1035 with
   audit comments naming the VOC-129 replacement merge. Do not manufacture a
   VOC-127 completion marker from #1041.

### Non-goals / explicitly excluded

- Re-implementing or replacing already-merged infrastructure #164.
- Opening a new `KARSIFT/karsift-ai-infra` PR unless a later exact-revision
  review proves the caller cannot consume `863fc1f…` as merged. That
  exception is not expected and is not this package's design.
- Publishing, merging, or composing PR #1041, artifact
  `implement-VOC-127-VOC-127-T00-attempt2`, or unpublished head
  `9d459813a733f9c6d58ad3352df0db27d33ee7f4`.
- Dispatching VOC-127-T00 as attempt `3`, resetting attempt `2` to attempt
  `1`, or treating this package as a third VOC-127 implementation attempt.
- Manufacturing a VOC-127 completion marker from #1041.
- Changing application runtime behavior, deployment topology, product
  permissions, or monitor inventory, except the staging skip for
  tree-equivalent develop-sync already required by VOC-127-D05.
- Snapshotting the current develop/main gap as a later drift gate
  (`karsift-ai-infra#15`).
- Fast-forwarding `main` instead of preserving `--merge` promotion commits.
- Normalizing direct-to-main as the ordinary workflow.
- Operator-typed SHA inputs; overloading `existing_pr_number`.
- Reviving VOC-127-DEP-07 preferred name `reconcile-main-to-develop`. The
  live #164 identity is `reconcile-production-change`.
- Changing `config/roles.yml`, adding OpenAI execution, or requesting
  `OPENAI_API_KEY`.
- Weakening exact-SHA review, risk floors, protected checks, App-token
  isolation, force-with-lease, retry caps, or fail-closed missing-binding
  behavior.
- Implementing VOC-122 promotion-recovery replan, merging unrelated open
  carriers, or rotating `KARSIFT_BOT_*` secrets / App installation
  permissions.
- A supervised bootstrap exception.
- Rewriting historical CHANGELOG, A-003, VOC-075, or VOC-127 audit records
  except for a new current-state note where a live contract would otherwise
  stay false.
- Splitting workflow logic, tests, docs, fixture/pin, or evidence into
  separate tasks.
- Self-adoption or self-authorization of this package.
- Operator-owned live-evidence contracts: acceptance is deterministic tests
  plus exact-SHA review. The #1041/#1039/#1035 close order is recorded
  post-promotion evidence, not a VOC-097 live-evidence gate.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: caller `tooling/governance/` fixtures and tests; live
  `.github/workflows/pipeline.yml` exceptional production-change dispatch;
  `deploy-staging.yml` path selection if a job-level skip is required;
  current-state release and branch-behavior documentation.
- Protected technical effect: whether the caller fixture and pin equal the
  authoritative #164 merge; whether live dispatch can reconcile an authorized
  main-only change; whether tree-equivalent ref sync deploys staging. No
  application runtime effect is intended for ordinary promotions.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but workflow and governance-fixture changes still
  require exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-129-D00`: This is one outcome-sized caller replacement. Use one
end-to-end implementation task covering live pipeline dispatch, staging
path-selection proof or narrow skip, current-state docs, caller fixture/pin,
tests, and evidence. Coordinated infrastructure work is already merged as
#164; this task consumes that exact merge rather than re-implementing it.
Repository count, file count, and workflow-versus-tests-versus-docs are not
split reasons. Superseding #1041 and closing #1039 / #1035 are evidence of
this outcome, not additional VOC-129 tasks and not a third VOC-127 roster
entry.

`VOC-129-D01`: Recreate the intended VOC-127 caller diff from current
`develop`. Do not compose unpublished #1041 files, the attempt-2 artifact, or
a cherry-pick of `9d459813a733f9c6d58ad3352df0db27d33ee7f4`. The replacement
carrier is a new VOC-129 branch/PR.

`VOC-129-D02`: Pin and fixture identity is exact merge
`863fc1f35b1d35e4981a59166b0e939be1a2b681`. `PINNED_SHA.txt` and every live
pin assertion must equal that SHA. They must not equal
`a9df74a63976d5239b84151fd01310835c999e7c` or
`60afda3a44fd06b8c00b219771de7112f1aded6e`. Mirror every in-scope
infrastructure file needed by the caller from that merge. If a file is not in
the fixture subset, do not copy it merely to force a pin; record
non-consumption. Consumption is expected.

`VOC-129-D03`: Exceptional main-only reconciliation on the caller uses the
live #164 identity `reconcile-production-change`. It requires an adopted
package/task authority issue and a merged main-targeting change already on
`main`. Caller `workflow_dispatch` still must not expose operator-typed SHA
inputs. `existing_pr_number` remains implement-only. Keep every
`workflow_dispatch` block at most 25 inputs. Preserve live
`live_evidence_mode` unless adding the new action itself requires touching
that input.

`VOC-129-D04`: Preserve the already-approved VOC-127 behavioral contract in
the mirrored fixture and live caller docs: after `--merge`, `develop`
advances to `mergeCommit.oid` before audit close; `reconcile-release` retries
that sync; unique develop commits, moved `main`, malformed merges, and
missing/unbound refs fail closed; auto-deleted `develop` is recreated at the
merge SHA; equal tips do not open a second promotion PR; tree-equivalent sync
must not keep staging scheduled.

`VOC-129-D05`: Deterministic tests must prove:

1. `PINNED_SHA.txt` equals `863fc1f35b1d35e4981a59166b0e939be1a2b681` and does
   not equal `a9df74a6…` or `60afda3a…`;
2. matching foundation pin literals equal the same SHA;
3. the mirrored fixture includes #164 `release-checkout-ref-runner.py` and
   its missing-`develop` fallback / checkout-ref ordering tests, and caller
   suites execute or otherwise fail closed if those files drift from the pin;
4. live `pipeline.yml` exposes `reconcile-production-change`, forwards
   `issue_number` as authority, does not expose operator SHA inputs, and stays
   at most 25 `workflow_dispatch` inputs;
5. `release` waits on `reconcile-production-change` and `auto-advance` does
   not proceed on a failed production-change reconcile;
6. fixture `release.yml` no longer restores `CHECKED_HEAD_SHA` as the
   post-merge integration tip;
7. VOC-111 allowlisted runtime/deploy paths still select staging; empty or
   specs-only diffs do not;
8. `config/roles.yml` is unchanged and no OpenAI route is added;
9. this package's caller PR `Closes` only its own VOC-129 task issue.

Tests must not mint real App tokens, use secrets, or use production data.

`VOC-129-D06`: Preserve VOC-113 through VOC-126 fail-closed contracts: roster
markers over issue-closed state, exact-head promotion checks, ruleset
attestation, required-check recovery, App-token mutation isolation, job-token
recovery reads, force-with-lease / match-head-commit, two-attempt implementer
bound, nested isolation, credential-free bundles, Cursor Composer
implementer, Cursor Grok exact-revision review, no OpenAI route, `roles.yml`
unchanged, no secrets in bundles/logs/fixtures, no credential values printed.
This package's caller PR `Closes` only its own VOC-129 task issue.

`VOC-129-D07`: Current-state comments in live `pipeline.yml`, fixture
`release.yml` / template / README, `AGENTS.md` (reconcile-release and release
authority), `docs/operations/11-devops-and-ci-cd.md`,
`docs/operations/10-development-workflow.md`, and current-state
branch/release paragraphs in
`docs/operations/15-ai-native-product-and-engineering-operating-model.md` and
`docs/governance/16-autonomous-development-operating-model.md` must say that
after a successful promotion merge, `develop` is advanced to that exact merge
SHA before audit close, that `reconcile-release` retries that sync, and that
exceptional governed production-change reconciliation is a separate adopted
exact-SHA path named `reconcile-production-change`. Historical amendment
records stay unchanged.

`VOC-129-D08`: This package is the governed replacement for unpublishable
VOC-127 caller PR #1041. Do not merge #1041. Do not dispatch VOC-127-T00 as
attempt `3`. After the exact reviewed caller replacement is merged and
promoted:

1. close #1041 as superseded, with an audit comment naming the VOC-129 merge
   SHA and stating that #1041 was never published because the pin/evidence
   still named infrastructure #163;
2. close VOC-127 task #1039 and origin #1035 only then, with audit comments
   that the remaining VOC-127 caller contract is delivered by this package's
   exact reviewed merge — not by a VOC-127 completion marker bound to #1041,
   and not by treating a VOC-129 PR as a third VOC-127 attempt.

`VOC-129-D09`: No bootstrap exception. VOC-124 already requested
`permission-workflows: write` on `publish-source`. T00's first run is attempt
`1` on a new VOC-129 carrier. Do not treat an untracked local
`karsift-ai-infra/` checkout as this repository's tracked tree. Do not treat
infrastructure merge `a9df74a6…` or `60afda3a…` as the pin target.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. Caller dispatch of
`reconcile-production-change` uses the existing App token with contents write
already required by release; it must not broaden installation permissions,
must not grant the model-controlled runner that token, and must not accept
free-form SHAs as authority.

Abuse/process risks:

1. Publishing stale #1041 or retrying VOC-127 as attempt `3` — forbidden by
   `VOC-129-D01` / `VOC-129-D08`.
2. Pinning #163 (`a9df74a6…`) or leaving `60afda3a…` — forbidden by
   `VOC-129-D02`.
3. Accepting operator SHAs or reviving `reconcile-main-to-develop` as a
   second mutating dialect — forbidden by `VOC-129-D03`.
4. Erasing unique develop commits — still forbidden by VOC-127-D02, preserved
   in the #164 fixture.
5. Unnecessary staging deploys from tree-equivalent ref sync — mitigated by
   `VOC-129-D04` / `VOC-129-DEP-07`.
6. Printing App tokens, private keys, or secret values — forbidden.

## Contradictions and open questions

1. **VOC-127-DEP-07 name:** required behavior is settled by merged #164. The
   live action identity is `reconcile-production-change`. Do not implement
   the earlier preferred name `reconcile-main-to-develop`.
2. **Staging skip mechanism (`VOC-129-DEP-07`):** required outcome is settled
   (tree-equivalent sync must not deploy; allowlisted runtime changes must).
   Whether VOC-111 `on.push.paths` already skips merge-commit fast-forwards,
   or a job-level selector must be added, is proven at implementation time.
   Do not broaden the allowlist.
3. **`live_evidence_mode` versus `live_evidence_dispatch`:** the live caller
   currently uses choice input `live_evidence_mode`; the #164 template uses
   boolean `live_evidence_dispatch`. That drift predates this defect and is
   not this package's objective. Preserve the live caller's existing
   reconcile-live-evidence input shape unless adding
   `reconcile-production-change` itself requires touching that input.
4. **Fixture pin applicability:** consumption is expected. Pin to
   `863fc1f35b1d35e4981a59166b0e939be1a2b681`. If some files are not in the
   mirrored subset, do not copy them merely to force a pin; record
   non-consumption.
5. **AGENTS.md / DOC-15 / DOC-16:** update current-state sentences that would
   become false. Do not expand them into a new release-internals runbook, and
   do not rewrite historical correction notes.
6. **Infrastructure already merged:** this package must not snapshot, re-open,
   or rewrite #164. If later exact-revision review finds the caller cannot
   consume `863fc1f…`, record that as a blocking finding rather than silently
   forking infrastructure inside the caller.
