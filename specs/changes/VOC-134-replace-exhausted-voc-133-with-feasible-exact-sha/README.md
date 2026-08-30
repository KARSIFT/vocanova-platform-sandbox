# VOC-134 — Replace exhausted VOC-133 with feasible exact-SHA evidence and complete infra #166 release

| Field | Value |
|-------|-------|
| Package | `VOC-134` |
| Title | Replace exhausted VOC-133 with feasible exact-SHA evidence and complete infra #166 release |
| Path | `specs/changes/VOC-134-replace-exhausted-voc-133-with-feasible-exact-sha` |
| Status | `draft` |
| Risk | `R4` (draft proposal; caller fixture/pin and `tooling/governance/` tests) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1066](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1066) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

VOC-133 was adopted to pin exact infrastructure
[KARSIFT/karsift-ai-infra#166](https://github.com/KARSIFT/karsift-ai-infra/pull/166)
(`f3d79177bf8a9abe0dae550f39502165d494c576`) from current `develop` with a
complete VOC-112 no-change boundary. VOC-133-T00 issue
[#1063](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1063) /
PR [#1065](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1065)
exhausted attempts 1 and 2. Exact reviewed heads:

| Attempt | Head | Independent result |
|---------|------|--------------------|
| 1 | `e88fbda6274488bed898877151bbcc1714a45450` | FAIL: [comment 5441321518](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1065#issuecomment-5441321518) |
| 2 | `70930cf0572f73f254429a240c5328640846e0a9` | FAIL: [comment 5441659390](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1065#issuecomment-5441659390) |

Attempt 1 added `package.json` plus
`scripts/foundation/ensure-voc112-capture-commits.mjs`, fetching a missing
VOC-112 capture commit before `pnpm test` and making the required
full-checkout `local` fail-closed behavior unreachable. Supervisor finding:
[comment 5441214837](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1065#issuecomment-5441214837).

Attempt 2 removed the helper but still changed `package.json` to force
`VOC112_CAPTURE_PROVENANCE_MODE=pr-validation` with both PR SHAs set to
`HEAD`, again bypassing the default `local` fail-closed path. It also added
a mutable evidence-stamping helper and an in-memory rewrite that allowed
tests to pass while committed evidence named the prior failed head.

There is a specification defect in VOC-133-D12 / TEST-10 as interpreted by
those attempts: a tracked file in a commit cannot contain the SHA of that
same commit, because changing the file changes the commit SHA. Exact
reviewed-revision binding must remain fail-closed, but it must be represented
by the immutable App-authored independent-review comment/check and PR
metadata (and post-merge audit comment), not by an impossible
self-referential value in the same Git tree. The committed evidence may
identify the exact carrier base and infra merge and explicitly state where
the later exact-head binding will be published.

VOC-133 cannot be redispatched. Do not reuse PR #1065. This package is the
governed replacement authority for the same still-unfinished #166 pin plus
release-repair outcome, with a feasible exact-revision evidence contract.

The caller remains pinned to infrastructure #164 because VOC-130 / VOC-131
exhausted on unrelated VOC-112 changes, VOC-132 never published a caller PR,
and VOC-133 exhausted on provenance-bypass and self-referential evidence.

### Live reproduction

| Item | Value |
|------|-------|
| Exhausted VOC-130 task | [#1049](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1049) (`VOC-130-T00`) |
| Unmerged VOC-130 carrier | [#1051](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1051) |
| Exhausted VOC-131 carrier | [#1056](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1056) |
| Adopted VOC-132 plan | [#1058](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1058) |
| VOC-132 task that must not be redispatched | [#1059](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1059) (`VOC-132-T00`) |
| VOC-132 failed implementer run | [`33079499176`](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/33079499176), job `98543693163` |
| Adopted VOC-133 plan | [#1062](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1062) |
| Exhausted VOC-133 task | [#1063](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1063) (`VOC-133-T00`) |
| Exhausted VOC-133 carrier | [#1065](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1065) |
| VOC-133 attempt-1 FAIL head | `e88fbda6274488bed898877151bbcc1714a45450` |
| VOC-133 attempt-2 FAIL head | `70930cf0572f73f254429a240c5328640846e0a9` |
| VOC-129 caller merge | [#1046](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1046) at `429d8c6d49303148ca1cc14dba5f6768a7863346` |
| Original failing/no-op release wake-up | [`33066533397`](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/33066533397) |
| Authoritative infra merge | `KARSIFT/karsift-ai-infra@f3d79177bf8a9abe0dae550f39502165d494c576` (#166) |
| Failed initial infra review head | `ce86f5d77c733e0e9f30397c167fbd1dfc7c5a8f` |
| Independently reviewed PASS infra head | `1488619d0d37aaa179d8e739bfe931881d6c51aa` |
| Current `develop` pin at drafting | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (infrastructure #164) |
| Expected immutable carrier base at drafting | `95a779f9e62090f856ed03f389e7ac1d901aaa14` |
| VOC-112 `subject_revision` at current `develop` | `f9d11e232a07c7d7a9c433d02c9267912543ba10` |

Root cause of the VOC-133 exhaustion: the adopted VOC-133 evidence test was
interpreted as requiring the implementation tree to name its own head SHA,
which is Git-impossible. The attempts then bypassed VOC-112 `local`
fail-closed behavior (`package.json` provenance-mode wrapper / capture-commit
fetch) and mutated evidence at test time so a prior failed head could still
appear to pass.

Root cause of the VOC-132 publication failure: post-model lifecycle helpers
were still read from the nested checkout after the implementer had
legitimately removed it. #166 copies those helpers to immutable `/tmp` paths
before the unrestricted model runs, treats an absent nested checkout as no
infrastructure-source change while caller changes continue, and fails closed
on a surviving path that is not a distinct nested Git checkout.

Root cause of the earlier release no-op (run `33066533397`): shared policy
must be restored from the same immutable reusable-workflow revision after
caller checkout and before later lifecycle helpers. That #165 restore remains
required and is included in the #164→#166 fixture delta.

## Required outcome (summary)

Use one largest-safe coherent replacement task and one fresh implementation
PR from current `develop` (do not reuse PR #1051, #1056, or #1065, and do not
redispatch VOC-132-T00 or VOC-133-T00):

1. Pin `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` exactly to
   infra #166 merge `f3d79177bf8a9abe0dae550f39502165d494c576`. It must equal
   neither #164 `863fc1f35b1d35e4981a59166b0e939be1a2b681` nor #165
   `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`.
2. Mirror every necessary authoritative #164→#166 caller fixture surface
   byte-for-byte: `.github/workflows/implement.yml`,
   `.github/workflows/release.yml`, `config/implementer_nested_checkout.py`,
   `tests/test_release_policy.py`, `tests/test_voc121_implement_policy.py`,
   `tests/test_voc123_source_bundle.py`, and fixture `CHANGELOG.md`, with
   committed SHA-256 constants and no runtime `/tmp` checkout dependency.
3. Preserve #165 post-caller-checkout release-policy restoration in both
   `identify` and `converge` after caller checkout and before
   task-completion helpers, using exact `job.workflow_repository` /
   `job.workflow_sha` and `persist-credentials: false`.
4. Preserve #166 helper-copy-before-model and nested-checkout classification:
   absent continues; symlink, non-directory, and parent-Git inheritance fail
   closed.
5. Keep the five VOC-112 no-change paths byte-identical to the immutable
   carrier-base SHA and absent from the implementation diff. JSON
   `subject_revision` remains `f9d11e232a07c7d7a9c433d02c9267912543ba10`.
6. Leave `package.json` byte-identical to the carrier base. Do not add a
   capture-commit fetch helper, provenance-mode wrapper, evidence-stamping
   helper, test-time evidence mutation, or any invocation that masks default
   `local` fail-closed behavior.
7. Bind the no-change regression to the immutable carrier-base SHA selected
   before implementation (expected `95a779f9e62090f856ed03f389e7ac1d901aaa14`).
   Revalidate current `develop` at dispatch and fail closed if it moved
   unexpectedly. The test must require the exact commit and all five paths to
   resolve and must fail—not skip—if proof cannot run. Do not use a moving
   branch ref as the authority after merge.
8. Define feasible exact-revision evidence: committed package evidence records
   carrier base, exact infra merge, hashes, validation, and the contract that
   the final implementation head is bound by the App-authored
   independent-review comment/check. The final review comment must bind the
   live PR head exactly; merge gate must reject any mismatch. Post-merge
   audit may record reviewed head and merge SHA. Do not require a commit to
   contain its own SHA.
9. Preserve #164 recovery and exact-SHA synchronization, promotion checks,
   unique-develop fail-closed behavior, serialization/idempotency,
   review/implementer separation, retry caps, App-token isolation, sanitized
   errors, task completion markers, and release-loop/staging-deploy
   suppression. Leave `config/roles.yml` unchanged. Add no OpenAI route.
10. After exact-SHA PASS and protected checks, merge through KARSIFT, promote
    `develop` to `main`, advance `develop` to the exact resulting main merge
    SHA, avoid a staging deploy for tree-equivalent sync, complete production
    deployment where selected, and reconcile/close VOC-129 plus exhausted /
    superseded VOC-130 through VOC-133 without manufactured task markers.

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates. Treat this as one outcome-sized task.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Pin exact infra #166 from current develop with feasible exact-SHA evidence, complete VOC-112 no-change boundary, and complete the release repair | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or
implementation authority.

## Risk note

This package **proposes R4** because durable caller fixture/pin and
`tooling/governance/` test updates belong under the R4 path floor, and
because the change mutates the protected implementer nested-checkout /
source-carrier and release checkout-lifetime contracts (`.github/workflows/*`
is an R3 floor if a live caller workflow must change). The path-based
classifier and independent verifier remain authoritative; this draft proposal
is not a determination.
