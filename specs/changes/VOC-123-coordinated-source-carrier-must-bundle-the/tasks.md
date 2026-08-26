# VOC-123 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
This package intentionally defaults to **one task** because issue #1005 is one
source-carrier integrity outcome. Coordinated caller and infrastructure
pull requests remain one task; repository count, workflow-versus-tests-versus-docs,
and fixture/pin work are not split reasons.

Cross-repo note: T00 changes `KARSIFT/karsift-ai-infra` for named-ref source
bundle creation. Because the broken running workflow cannot publish its own
repair, the initial infra PR uses the bounded supervised bootstrap in
`VOC-123-D08`; this package is the authorizing change package for the required
outcome. Do not
treat the untracked local `karsift-ai-infra/` checkout (if present) as this
repo's tracked tree. Caller fixture/pin, tests, and evidence land in this
repository under the same task. Infra PRs must say
`Relates to KARSIFT/vocanova-platform-sandbox#<task>` and MUST NOT use a
closing keyword.

## VOC-123-T00 — Bundle the coordinated source-carrier committed head through a named ref

- Requirement source: issue #1005; `VOC-123-D00` through `VOC-123-D08`
- Acceptance criteria: `VOC-123-AC-00` through `VOC-123-AC-06`
- Tests: `VOC-123-TEST-00` through `VOC-123-TEST-06`
- Evidence: `VOC-123-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1005 live failure in `t00-evidence.md` (task #1003,
   pipeline run `32915078678` / job `98017696468`, nested commit `db31cc9`,
   `fatal: Refusing to create empty bundle.`, no uploaded artifacts, defect
   in `implement.yml` `base_sha..$SOURCE_HEAD_SHA`).
2. In `KARSIFT/karsift-ai-infra/.github/workflows/implement.yml` `Commit
   implementer's work`, after the isolated nested source commit, bind the
   exact 40-character `SOURCE_HEAD_SHA` to an isolated temporary named ref
   that is not the publish branch and not the remediation `source-carrier`
   branch. Create `/tmp/implementer-source.bundle` from
   `"${{ steps.infra-checkout.outputs.base_sha }}..<that-ref>"`. Verify the
   bundle is non-empty and `git bundle list-heads` advertises exactly the
   expected committed head. Remove the temporary ref afterward, including on
   verification failure. Continue to emit `source_head_sha=$SOURCE_HEAD_SHA`
   for the existing clean publisher.
3. Inspect caller recovery (`git bundle create /tmp/implementer-work.bundle
   "${{ steps.branch.outputs.integration_sha }}..HEAD"`) and planner recovery
   (`git bundle create /tmp/planner-work.bundle "${{ steps.branch.outputs.base_sha }}..HEAD"`).
   Prove with real-repository tests whether `HEAD` is advertised safely.
   Change only paths that reproduce empty-bundle or unsafe advertised-head
   behavior.
4. Preserve VOC-121 fail-closed contracts: exact base/head binding,
   nested-repository isolation, no gitlink, credential-free source bundle,
   clean `publish-source` App-token separation, `bundle verify` plus
   authorized-SHA fetch, force-with-lease, two-attempt bound, non-closing
   source PR, caller `Closes #N`. Never bundle unrelated refs or secrets.
5. Add deterministic tests that create real temporary Git repositories and
   prove:
   - raw-SHA positive tip reproduces empty-bundle (exit 128);
   - the fixed source-carrier path produces a non-empty bundle advertising
     exactly the expected committed head;
   - wrong/missing/multiple advertised heads, wrong prerequisite/base,
     malformed SHA, and cleanup/publish mismatches fail closed;
   - no unrelated refs or objects become publishable;
   - existing caller/planner bundle paths remain correct or are fixed with
     equivalent regression coverage.
   Do not treat existing VOC-121 `..HEAD` / `..$branch` bundle tests as
   covering the production raw-SHA defect.
6. Update current-state comments/docs (`implement.yml` current-state /
   commit-step comments, `plan.yml` only if that path changes,
   `karsift-ai-infra/README.md` source-carrier paragraph) so they no longer
   describe a raw-SHA exclusion range as a working bundle contract.
7. Bootstrap the initial infra change through one clean isolated branch from
   current infra `main` under `VOC-123-D08`, because the already-compiled
   `implement.yml@main` cannot consume its own nested edit. Open one reviewed
   non-closing infra PR, run full source validation, bind independent review
   to its exact final SHA, and require a different actor to merge. Do not
   mutate PATH/Git, push to `main`, reuse `db31cc9`, or widen the patch. After
   that exact merge is live, use the normal coordinated carrier for remaining
   caller work. Pin
   `tooling/governance/fixtures/karsift-ai-infra/` to that exact merge SHA
   when the fixture consumes the change. Update caller fixture regressions
   and any `scripts/foundation/*` pin literals that still assert
   `99476c2a1018e42d4bd442657b5257885ac9f1c9`, in the same task.
8. Record in `t00-evidence.md` that #1003 (`VOC-122-T00`) is a distinct
   already-authorized task to re-dispatch or reconcile against the exact
   reviewed infra merge. Do not implement VOC-122 promotion-recovery replan
   here and do not reconstruct runner-only commit `db31cc9` by hand.
9. Run applicable validation and record results in `t00-evidence.md`:
   - `python3 -m unittest discover -s tests -p 'test_*.py'` in the primary
     `KARSIFT/karsift-ai-infra` checkout;
   - `bash scripts/governance/validate-governance.sh`;
   - `bash scripts/governance/classify-change-risk.sh`;
   - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
   - targeted foundation pin tests if those files change;
   - `git diff --check`;
   - exact reviewed infra SHA and pin applicability;
   - any narrower targeted commands added by the implementation.
10. Preserve independent exact-SHA review for each carrier, risk
    classification, protected checks, and App-token isolation.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider, or
  monitor-inventory changes.
- Implementing VOC-122 promotion-recovery replan or otherwise delivering
  #1003 inside this package.
- Hand-patching or pushing nested commit `db31cc9`; bootstrap is limited to
  the newly implemented, validated, independently reviewed source repair in
  `VOC-123-D08`.
- Weakening exact-SHA review, risk floors, protected checks, retry caps, or
  App-token isolation.
- Splitting workflow logic, tests, docs, infrastructure, caller pin, or
  evidence into separate tasks.
- Operator-owned live evidence contracts: acceptance is deterministic tests
  plus exact-SHA review. The #1003 empty-bundle incident is already recorded;
  the next live implement run is not an implementer-dispatched VOC-097
  evidence obligation.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the live defect, the named-ref contract, the tests, the docs,
  and the caller pin are one carrier-integrity outcome.
- Infra should merge first when the caller fixture/pin consumes that change;
  otherwise the two reviewed PRs may complete under the same task without a
  pin bump.
- The initial infra repair cannot consume itself: GitHub resolves the reusable
  workflow from old infra `main` before the job starts. Use only the bounded
  `VOC-123-D08` supervised bootstrap for that first infra PR. Once merged,
  normal `implement.yml@main` publication is mandatory and the bootstrap
  authority is exhausted.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
