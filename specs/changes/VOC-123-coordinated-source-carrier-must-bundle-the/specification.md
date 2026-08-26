# VOC-123 — Coordinated source carrier must bundle the committed head through a named ref: Specification

## Objective and requirement source

Make the governed coordinated source-carrier publisher able to deliver nested
`KARSIFT/karsift-ai-infra` changes by advertising the exact committed source
head through a named Git-bundle tip, without weakening VOC-121 isolation,
lineage, lease, App-token, or fail-closed contracts.

**Requirement source:** [GitHub issue #1005](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1005).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1005)

| Item | Value |
|------|-------|
| Adopted task blocked | [#1003](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1003) (`VOC-122-T00`) |
| Pipeline run / job | `32915078678` / `98017696468` |
| Nested commit on the runner | `db31cc9` |
| Error | `fatal: Refusing to create empty bundle.` |
| Artifacts / carrier PRs | none uploaded; no implementation carrier PR for #1003 |
| Defect locus | `KARSIFT/karsift-ai-infra/.github/workflows/implement.yml` `Commit implementer's work`: `base_sha..$SOURCE_HEAD_SHA` |
| Caller fixture mirror | `tooling/governance/fixtures/karsift-ai-infra/.github/workflows/implement.yml` |
| Root cause | Git bundle create requires a named positive revision; a raw SHA advertises no bundle head |

## Scope and non-goals

### In scope

1. Before nested source-bundle creation, bind the exact committed
   `SOURCE_HEAD_SHA` to an isolated temporary named ref (or an equivalently
   safe named ref), create `/tmp/implementer-source.bundle` from
   `infra-checkout.base_sha..that-ref`, verify the bundle advertises exactly
   the expected committed head/ref, and remove the temporary ref afterward.
2. Inspect caller recovery (`implement.yml`
   `integration_sha..HEAD`) and planner recovery (`plan.yml` `base_sha..HEAD`).
   Prove with deterministic tests whether `HEAD` is advertised safely. Change
   those paths only if they reproduce empty-bundle or unsafe advertised-head
   behavior.
3. Keep publication fail-closed:
   - exact 40-character base and head binding;
   - nested-repository isolation (no caller gitlink);
   - bundle verify plus fetch of the authorized head only;
   - artifact integrity (upload the verified bundle, refuse a missing bundle);
   - force-with-lease publishing and App-token separation;
   - two-attempt implementer bound;
   - independent exact-SHA review and protected checks;
   - never bundle unrelated refs or secrets.
4. Add deterministic tests that create real temporary Git repositories for
   the empty-bundle failure class, named-ref success, and named fail-closed
   cases.
5. Update current-state comments/docs so they no longer describe a raw-SHA
   positive tip as a working source-bundle contract.
6. Land the infrastructure change through one reviewed infra PR, then pin
   `tooling/governance/fixtures/karsift-ai-infra/` to that exact merge SHA when
   the fixture consumes the change, and update caller fixture/pin tests in the
   same task.
7. Record how active #1003 delivery is reconciled against that exact
   revision. Do not implement VOC-122 promotion-recovery behavior in this
   package.

### Non-goals / explicitly excluded

- Changing application runtime behavior, deployment topology, product
  permissions, or monitor inventory.
- Implementing VOC-122 promotion-recovery replan (`VOC-122-T00` / #1003).
  That remains a distinct already-authorized outcome and a hard dependent of
  this carrier-integrity repair.
- Hand-patching, force-pushing, or reconstructing runner-only nested commit
  `db31cc9` outside the governed implement path.
- Reopening VOC-121 source-publication semantics: isolated nested commit,
  clean `publish-source` job, infrastructure-scoped App token, non-closing
  `Relates to` source PR, caller `Closes #N`, and no gitlink remain.
- Weakening exact-SHA review, risk floors, protected checks, App-token
  isolation, force-with-lease, retry caps, or fail-closed missing-bundle
  behavior.
- Bundling unrelated refs, tags, remotes, or secrets.
- Splitting workflow logic, tests, docs, infrastructure, caller pin, or
  evidence into separate tasks.
- Self-adoption or self-authorization of this package.
- Operator-owned live evidence contracts: acceptance is deterministic tests
  plus exact-SHA review.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: reusable CI/CD implement/plan workflows, coordinated
  second-repository publication, and caller `tooling/governance/` fixtures
  and tests.
- Protected technical effect: whether an authorized nested infrastructure
  commit can be advertised in a Git bundle and published. No application
  runtime effect is intended.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but workflow and governance-fixture changes still
  require exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-123-D00`: This is one outcome-sized carrier-integrity change. Use one
end-to-end implementation task covering infrastructure source, tests,
current-state docs/comments, caller fixture/pin, and evidence. Coordinated
pull requests in `KARSIFT/karsift-ai-infra` and this caller remain one task.
Repository count, file count, and workflow-versus-tests-versus-docs are not
split reasons.

`VOC-123-D01`: The nested source-carrier bundle must be created from a named
positive revision bound to the exact committed `SOURCE_HEAD_SHA`, not from
that raw object ID. After the isolated nested commit:

1. Capture `SOURCE_HEAD_SHA` as a 40-character object ID and fail closed if
   it is malformed.
2. Bind that object ID to an isolated temporary named ref that is not the
   publish branch and not the remediation `source-carrier` branch.
3. Create `/tmp/implementer-source.bundle` from
   `"${{ steps.infra-checkout.outputs.base_sha }}..<that-ref>"`.
4. Verify the bundle is non-empty and `git bundle list-heads` advertises
   exactly the expected committed head (and the intended ref). Wrong,
   missing, or multiple advertised heads fail closed.
5. Remove the temporary ref afterward (including on verification failure).
6. Continue to record `source_head_sha=$SOURCE_HEAD_SHA` for the existing
   clean publisher, which fetches
   `"$PUBLISH_HEAD_SHA:refs/heads/$PUBLISH_BRANCH"`.

Switching the range to `..HEAD` without binding and verifying the exact SHA
is not sufficient to satisfy this decision. Control-flow shape (inline shell
versus a small helper; exact ref namespace) is an implementation choice
(`VOC-123-DEP-06`); the observable contract is not.

`VOC-123-D02`: Inspect every active bundle-creation path in `implement.yml`
and `plan.yml`. At drafting time those additional paths are:

- caller recovery: `git bundle create /tmp/implementer-work.bundle "${{ steps.branch.outputs.integration_sha }}..HEAD"`
- planner recovery: `git bundle create /tmp/planner-work.bundle "${{ steps.branch.outputs.base_sha }}..HEAD"`

`HEAD` is a named revision. T00 must prove with real-repository tests whether
those paths advertise a safe bundle head matching the recorded `head_sha`.
Change a path only if it reproduces empty-bundle or unsafe advertised-head
behavior. Do not "fix" working `..HEAD` paths merely for symmetry.

`VOC-123-D03`: Preserve VOC-121 fail-closed publication contracts:

- exact base/head SHA binding and integration-ancestor checks;
- nested checkout isolation and refusal to stage `karsift-ai-infra` as a
  gitlink;
- credential-free source bundle consumed by a clean `publish-source` job;
- infrastructure-scoped App token with no caller-token fallback;
- `bundle verify` then fetch of the authorized head only;
- force-with-lease against `EXPECTED_SOURCE_HEAD_SHA`;
- two-attempt implementer bound;
- source PR `Relates to OWNER/CALLER#N` with no closing keyword;
- caller PR keeps local `Closes #N`;
- no secrets in bundles, logs, or fixtures.

`VOC-123-D04`: Deterministic tests must create real temporary Git
repositories and prove:

1. a raw-SHA positive tip reproduces `fatal: Refusing to create empty bundle`
   (exit 128);
2. the fixed source-carrier path produces a non-empty bundle advertising
   exactly the expected committed head;
3. wrong, missing, or multiple advertised heads fail closed;
4. wrong prerequisite/base and malformed SHA fail closed;
5. cleanup/publish mismatches fail closed (advertised object ≠ recorded
   `SOURCE_HEAD_SHA`, or publisher would not fetch that object);
6. no unrelated refs or objects become publishable;
7. existing caller/planner bundle paths remain correct, or are fixed with
   equivalent regression coverage.

Positive cases must prove the corrected behavior. Tests must not use secrets
or production data. Existing VOC-121 tests that bundle `..HEAD` or
`..$branch` do not satisfy (1) or (2) by themselves.

`VOC-123-D05`: Current-state comments in `implement.yml` (commit/bundle
header and VOC-121 current-state block), `plan.yml` only if that path
changes, and `karsift-ai-infra/README.md` source-carrier paragraph must stop
implying that a raw-SHA exclusion range is a working bundle contract. After
the authoritative infrastructure merge SHA is known, pin
`tooling/governance/fixtures/karsift-ai-infra/` when the mirrored fixture
consumes the changed workflow, tests, or comments. Advance matching caller
pin assertions, including `tooling/governance/tests/test_voc121_implement_policy.py`
and the `scripts/foundation/voc097-fixture-matrix.test.mjs`,
`voc104-ready-for-review-reuse.test.mjs`, and
`voc108-authoritative-lifecycle.test.mjs` pin literals when those tests still
assert the previous infra merge (`99476c2a1018e42d4bd442657b5257885ac9f1c9`
at drafting time).

`VOC-123-D06`: VOC-121's isolated source publisher remains the normal delivery
path after this repair is live on infrastructure `main`. Silent discard of
nested infrastructure edits remains forbidden. Infra PRs must say
`Relates to KARSIFT/vocanova-platform-sandbox#<task>` and must not use a
GitHub closing keyword. The caller implementation PR keeps local `Closes #N`.

`VOC-123-D07`: This package is a hard dependency for #1003 and a distinct
carrier-integrity outcome. Do not implement VOC-122 promotion-recovery replan
behavior here. After the exact reviewed infra merge is pinned, record in
`t00-evidence.md` that #1003 should be re-dispatched or reconciled against
that revision. Do not reconstruct runner-only commit `db31cc9` by hand.

`VOC-123-D08`: T00 cannot publish its own initial infrastructure repair with
the broken path. The caller resolves
`KARSIFT/karsift-ai-infra/.github/workflows/implement.yml@main` before the
model job starts, so nested edits cannot recompile the running commit step;
the old raw-SHA bundle command will deterministically fail first. As in
`VOC-121-D10`, supervised bootstrap recovery of this task's own source PR is
therefore in-scope under T00, not a second package or task.

The bootstrap is limited to one clean isolated infrastructure branch based on
the current protected `main`, the exact named-ref repair, its tests/docs, and
no caller pin. It must open one non-closing infra PR linked to the T00
authority issue, pass source self-CI, receive independent review bound to its
exact final SHA, and be merged by someone other than the implementer. No
model-runner credential, PATH/Git interception, ruleset bypass, direct push to
`main`, or reuse of `db31cc9` is allowed. After that exact infra merge is live,
all remaining caller fixture/pin/evidence work and the later #1003 re-dispatch
must use the normal governed `implement.yml@main` path. This exception expires
with the bootstrap infra merge and cannot be reused by later tasks.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. Source publication
continues to use a credential-free bundle on the model-controlled runner and
an infrastructure-scoped GitHub App token only on the clean `publish-source`
job. The model-controlled implementer runner still never receives the GitHub
App token. Caller publication remains a separate credential-free bundle.

Abuse/process risks:

1. Empty source bundle discarding authorized nested work — mitigated by
   `VOC-123-D01` and raw-SHA versus named-ref tests.
2. Advertising or publishing unrelated refs/objects — mitigated by isolated
   temp-ref, exact `list-heads` verification, and publisher fetch of the
   authorized SHA only (`VOC-123-D01`, `VOC-123-D03`, `VOC-123-D04`).
3. Broadening recovery credentials or mixing App token onto the model runner
   — out of scope and forbidden.
4. Using a temp ref name that collides with the publish branch or remediation
   `source-carrier` checkout — forbidden by `VOC-123-D01` / `VOC-123-DEP-06`.

## Contradictions and open questions

1. **Temporary ref namespace (`VOC-123-DEP-06`):** the required behavior is
   settled; the exact ref name is not. T00 may use a `refs/karsift/…` ref or
   a throwaway local branch, provided it is isolated, not the publish branch,
   not `source-carrier`, deleted after use, and proven by tests.
2. **Inline versus helper:** T00 may keep create/verify/cleanup in
   `implement.yml` or extract a small helper (shell or Python next to
   `config/implementer_source_carrier.py`), as long as tests pin the live
   empty-bundle class and the publisher still consumes
   `/tmp/implementer-source.bundle` with the existing SHA fetch refspec.
3. **Caller/planner `..HEAD` paths:** drafting-time reading suggests `HEAD`
   is a named revision and those paths have published successfully. That is
   not a determination. T00 must prove it. If tests show `HEAD` is advertised
   unsafely (for example detached-HEAD or multiple-head cases), those paths
   receive the same named-ref contract. If they are safe, leave them and
   record the proof.
4. **Fixture pin applicability:** pin
   `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` to the exact
   reviewed infra merge when the mirrored fixture consumes the changed
   workflow, tests, or comments. `implement.yml` is in the policy fixture
   subset, so consumption is expected. If some files are not in that subset,
   do not copy them merely to force a pin; record non-consumption.
