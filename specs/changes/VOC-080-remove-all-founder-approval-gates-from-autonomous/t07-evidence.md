# VOC-080-EV-07 — T07 activation candidate and post-activation unblock

Evidence for `VOC-080-AC-00` through `VOC-080-AC-10`.

## Truthful activation semantics

This revision declares A-004 active only in the canonical repository tree produced
when the exact independently reviewed T07 head is merged. An unmerged PR is an
activation candidate, not active repository authority. The merge commit and merge
time are intentionally left to GitHub's immutable record; this file does not invent
a future timestamp or claim a verdict before the independent reviewer posts it.

The earlier package text requested one final founder transition approval. The later,
more specific founder direction revoked approval requirements for every condition,
including T07. That superseding requirement is recorded in
[issue #627 comment #5301333790](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/627#issuecomment-5301333790).
The comment is requirement provenance, not approval of this revision. T07 requires
deterministic gates and an independent exact-revision PASS, but no founder, human,
agent, workflow, environment, or exceptional-risk approval.

## Preconditions and live rehearsal evidence

| Requirement | Evidence | Result |
|---|---|---|
| Package adoption and T00–T06 | Plan PR [#628](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/628), task PRs #640–#645 | Complete |
| R4 automatic merge after non-human gates | Task PRs #640–#645 and their merge-gate runs | Pass |
| Unparseable risk fails closed | Disposable PR [#646](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/646), merge-gate comment [#5301291620](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/646#issuecomment-5301291620) | Blocked and closed unmerged |
| Release without founder interaction | Smoke run [31873626192](https://github.com/KARSIFT/karsift-ai-infra-smoke-test/actions/runs/31873626192), promotion PR [#12](https://github.com/KARSIFT/karsift-ai-infra-smoke-test/pull/12), release audit [#11](https://github.com/KARSIFT/karsift-ai-infra-smoke-test/issues/11) | Auto-promoted and audit closed |
| Reconcile is idempotent | Final sandbox run [31874129346](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31874129346) | 19-second success; no push, PR, duplicate issue, or root redispatch |
| Rehearsal defects fixed at source | Infra PRs [#41](https://github.com/KARSIFT/karsift-ai-infra/pull/41), [#42](https://github.com/KARSIFT/karsift-ai-infra/pull/42), [#43](https://github.com/KARSIFT/karsift-ai-infra/pull/43); caller PR [#647](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/647) | All exact heads passed installed CI before merge |
| Production environment reviewers | `VOC-080-EV-03` records `reviewers: null` | No environment approval gate |

## Activation candidate

| Item | Evidence / required gate |
|---|---|
| Task issue | [#637](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/637) |
| Activation PR | [#649](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/649) |
| Builder/verifier separation | Implementer created the revision; reusable `review` independently evaluates its exact SHA |
| Independent verification | Required external PASS on the exact final PR #649 head before merge; not pre-claimed in tracked state |
| Effective moment | GitHub merge time for the exact reviewed PR #649 head |
| Post-merge binding | GitHub PR merge commit/time and the exact-SHA review comment are canonical external evidence |

`approved_pr_head_sha` and `adopted_develop_sha` remain null in the candidate because a
commit cannot truthfully embed its own future SHA or merge commit. Validation permits
that self-reference-safe representation and requires the external exact-revision gate.

## Canonical activation artifacts

- `docs/governance/a004-transition-state.yaml`: `authority_model: a004-active`,
  active-on-canonical-merge lifecycle, no-approval clarification URL, rehearsal links.
- A-004 amendment: active canonical-tree notice and revoked transition-approval clause.
- A-003 successor pointer and protected-path policy: A-004 lockstep markers.
- `AGENTS.md`, `CLAUDE.md`, DOC-16, matrices, repository settings, templates: A-004
  is active after canonical merge; historical evidence remains historical.
- `tooling/governance/validate_repository_foundation.py`: enforces the active A-004
  state without fabricating approval or exact-revision evidence.

## VOC-079 / issue #624 unblock

After this exact revision passes independent review and merges, VOC-079 may resume on
the no-founder-approval engineering path. Its technical nginx cutover remains outside
T07. Recovery PR [#626](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/626)
is preserved for that package's own governed continuation.

## Acceptance closure

AC-00 through AC-09 are supported by `t01-evidence.md` through `t06-evidence.md` plus
the live rehearsal rows above. AC-10 becomes effective on canonical merge of this
exact independently reviewed activation revision. No item treats a missing check,
review, deployment, or external observation as passing.

## Deterministic validation

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_voc080*.py' -v
git diff --check
```

## Controls retained

- Independent exact-revision verification and builder/verifier separation
- CI, governance validation, protected-path floors, and fail-closed unknown risk
- Secrets isolation and least-privilege credentials
- Stronger R4 evidence, rollout, monitoring, rollback, and audit records
- Failed release/deploy remediation remains fail-closed
- RL1/RL2 technical activation remains disabled

## Post-merge actions

After PR #649 merges: record its immutable merge SHA/time in issue #627, close #627
only after the post-merge validation run passes, allow automatic develop-to-main
promotion, monitor production deployment, and verify application/monitoring endpoints.
