# VOC-125 — Manual resume of an existing implementation PR must supply exact retry bindings: Specification

## Objective and requirement source

Make the documented caller `action=implement` entrypoint able to resume an
existing implementation PR after automatic remediation has stopped, by
deriving immutable recovery identity from that PR and forwarding verified
`expected_head_sha` / `expected_base_sha` into `implement.yml` before any
model or mutation step, without weakening attempt caps, carrier identity,
publication leases, or fail-closed review isolation.

**Requirement source:** [GitHub issue #1020](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1020).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1020)

| Item | Value |
|------|-------|
| Adopted task blocked | [#1003](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1003) (`VOC-122-T00`) |
| Existing caller PR | [#1012](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1012) |
| Pipeline run / job | `32966618512` / `98170418081` |
| Caller develop at dispatch | `f0f7a54283c3527f40544fd236516dc3a5f4dc82` |
| Infrastructure ref | `f406cc95a3f853e8aef5bf8bcf22d37a29d64547` |
| Existing branch/head | `agent/voc-122-voc-122-t00@0b7be8c531be8300d5a1d5534acc83bf4d6a1791` |
| Prior review base | `e910eb4a21d48bbb5b3e0c30b8ee647d64683dbe` |
| Dispatch | `action=implement`, `attempt=2` |
| Failure | existing remote branch found; exact-head remediation binding cannot be established; stop before model resolution |
| Defect locus | caller `pipeline.yml` dispatch/implement job omits recovery identity and does not forward `expected_head_sha` / `expected_base_sha` |

## Scope and non-goals

### In scope

1. Add one caller `workflow_dispatch` input for implement-only existing-carrier
   recovery whose value is an explicit existing PR number
   (`existing_pr_number`). Forward it from the live caller
   `.github/workflows/pipeline.yml` `implement` job and from the infrastructure
   project-repo pipeline template.
2. Keep `implement.yml` as the derivation and fail-closed locus (the caller
   stays a thin wire). On `attempt != 1`, require a verified existing-carrier
   identity before `Create implementation branch` and before model resolution:
   - if `expected_head_sha` / `expected_base_sha` are already supplied
     (automatic `remediate.yml` path), bind them to the live PR, branch,
     authority issue, task, package, repository, prior-review base/head when a
     review exists, and current remote ref;
   - if those SHAs are empty, require `existing_pr_number`, load that PR, and
     derive the immutable head/base from live PR metadata plus App-signed
     exact-revision review/metadata when a review exists;
   - if both PR number and SHAs are supplied, they must match the live PR.
3. Forward `remediate.yml`'s known `pr_number` into `implement.yml` as
   `existing_pr_number` so automatic retry shares the same binding. Keep
   passing event-derived `expected_head_sha` / `expected_base_sha`.
4. Resume the same task issue, deterministic task branch, and existing PR.
   Never create a replacement carrier, delete the branch, or open a second PR
   merely because automatic retry is exhausted.
5. Preserve attempt `1` or `2` only. Never allow attempt `3`. Never reclassify
   attempt `2` as attempt `1`. Attempt `1` with an existing open task PR or
   remote task branch fails closed.
6. Fail closed before any model or mutation step for the mismatch classes in
   `VOC-125-D03`.
7. Preserve VOC-121/VOC-123/VOC-124 publication leases, source/caller
   isolation, workflow-file permission isolation, risk classification,
   protected checks, two-attempt implementer bound, Cursor Composer
   implementer, Cursor Grok exact-revision review, and no credential fallback.
8. Add deterministic source and caller tests for a valid existing-carrier
   resume and every mismatch class.
9. Update current-state workflow comments/docs that describe implement
   dispatch or remediation retry identity, and pin
   `tooling/governance/fixtures/karsift-ai-infra/` to the exact reviewed infra
   merge SHA when the fixture consumes the change.
10. After that exact infra merge is live on `implement.yml@main` and the caller
    dispatch contract is merged, resume existing `VOC-122-T00` / #1003 / #1012
    through the repaired route at attempt `2` with `existing_pr_number=1012`.

### Non-goals / explicitly excluded

- Changing application runtime behavior, deployment topology, product
  permissions, or monitor inventory.
- Implementing VOC-122 promotion-recovery replan (`VOC-122-T00` / #1003)
  inside this package. That remains a distinct already-authorized outcome
  whose existing carrier is resumed after this repair is live.
- Merging or treating caller PR #1012 as this package's implementation PR.
- Adding operator-typed free-form SHA inputs to caller `workflow_dispatch`.
- Allowing attempt `3`, resetting attempt `2` to attempt `1`, or deleting the
  existing branch/PR to start a replacement carrier.
- Weakening exact-SHA review, risk floors, protected checks, App-token
  isolation, force-with-lease, retry caps, or fail-closed missing-binding
  behavior.
- Changing GitHub App installation permissions or rotating
  `KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY`.
- Reopening VOC-121/VOC-123/VOC-124 source-publication contracts.
- OpenAI credentials or execution routes.
- Rewriting historical CHANGELOG, A-003, or VOC-075 audit records.
- Splitting workflow logic, tests, docs, infrastructure, caller pin, or
  evidence into separate tasks.
- Self-adoption or self-authorization of this package.
- A supervised bootstrap exception: VOC-124 already published
  `permission-workflows: write` on `publish-source`. The first T00 run is
  attempt `1`.
- Operator-owned live-evidence contracts: acceptance is deterministic tests,
  exact-SHA review, and a recorded handoff to resume existing #1003 / #1012
  through the repaired `action=implement` route.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: reusable CI/CD implement and remediate workflows, caller
  pipeline dispatch contract, exact-head remediation bindings, attempt caps,
  implementation-carrier identity, and caller `tooling/governance/` fixtures
  and tests.
- Protected technical effect: whether an operator can resume an existing
  implementation PR after automatic remediation stops, and whether unverified
  SHA/PR identity can reach the model or mutate the carrier. No application
  runtime effect is intended.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but workflow and governance-fixture changes still
  require exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-125-D00`: This is one outcome-sized existing-carrier recovery change.
Use one end-to-end implementation task covering infrastructure source, caller
dispatch contract, tests, current-state docs/comments, caller fixture/pin, and
evidence. Coordinated pull requests in `KARSIFT/karsift-ai-infra` and this
caller remain one task. Repository count, file count, and
workflow-versus-tests-versus-docs are not split reasons. Resuming the existing
VOC-122 carrier is evidence of this outcome, not a second VOC-125 task and not
a replacement VOC-122 roster entry.

`VOC-125-D01`: Operator recovery identity is an explicit existing caller PR
number. The live caller `pipeline.yml` and the infrastructure project-repo
pipeline template add optional `workflow_dispatch` input `existing_pr_number`
(implement only) and forward it on the `implement` job. Do not add
operator-typed `expected_head_sha` or `expected_base_sha` inputs to caller
`workflow_dispatch`. Automatic `remediate.yml` continues to pass event-derived
SHAs and must also forward its known `pr_number` as `existing_pr_number`.

`VOC-125-D02`: `implement.yml` remains the derivation and fail-closed locus.
Add optional input `existing_pr_number` (default empty). Before
`Create implementation branch` and before model resolution:

- Attempt `1`: `existing_pr_number` must be empty; SHA inputs must be empty;
  an existing open task PR or remote deterministic task branch fails closed.
- Attempt `2`: a verified existing-carrier identity is required. If SHA
  inputs are empty, derive them from `existing_pr_number`. If SHA inputs are
  present, bind them to the live PR. If both are present, they must agree
  with the live PR head/base.
- Attempt other than `1` or `2` fails closed. There is no attempt `3`.

Derived or supplied SHAs are then the only values `verify-expected-head.py`,
prior-review evidence, and publication leases consume.

`VOC-125-D03`: Fail closed before any model or mutation step when any of the
following is true:

- missing, empty, or non-40-hex head or base after derivation;
- live PR head or base does not match the bound SHAs;
- live remote task-branch head does not match the bound head
  (`STALE` / `INVALID_*` from `verify-expected-head.py`);
- PR number missing on attempt `2` when SHAs are also empty;
- PR is not open, is merged, or is from another repository;
- PR head ref is not the deterministic task branch
  `agent/<change-id-lower>-<task-slug>`;
- PR body/metadata does not bind the dispatched `task_id`, `package_path`,
  and `issue_number`;
- authority issue is closed or the task is already completed;
- App-signed review evidence exists but is foreign (wrong author, wrong
  head/base, wrong task/package/issue) or malformed;
- dispatched attempt is `1` while an existing open task PR or remote task
  branch exists;
- dispatched attempt is not `1` or `2`.

Absent App-signed review remains allowed only for the existing CI-failure
class (no trusted review record for that failed head), matching current
`implement.yml` prior-review behavior. In that class, identity still comes
from the live open PR and remote branch, not from guessed SHAs.

`VOC-125-D04`: Resume the same task issue, branch, and PR. Publication still
uses force-with-lease against the bound expected heads. The publisher must
not open a second PR for the same task branch, silently reopen a closed PR,
or replace #1012. Closed or completed carriers fail closed
(`conflicting_existing_pr` class and the closed-task class in `VOC-125-D03`).

`VOC-125-D05`: Preserve VOC-121/VOC-123/VOC-124 fail-closed publication
contracts: exact base/head SHA binding, nested-repository isolation, no
gitlink, named-ref source bundle, credential-free bundles, clean
`publish-source` App-token separation with `permission-workflows: write` only
on that mint, caller `publish` still omitting workflow-write and still
refusing `.github/workflows/**`, force-with-lease, two-attempt implementer
bound, source PR `Relates to OWNER/CALLER#N` with no closing keyword, caller
PR `Closes #N`, Cursor Composer implementer, Cursor Grok exact-revision
review, no OpenAI route, no secrets in bundles/logs/fixtures, no credential
values printed.

`VOC-125-D06`: Deterministic tests must prove:

1. a valid attempt-2 resume with `existing_pr_number` matching the live open
   PR, branch, issue, task, package, repository, prior-review base/head, and
   current remote head populates `expected_head_sha` / `expected_base_sha` and
   proceeds past the exact-head guard without creating a replacement carrier;
2. the #1020 / job `98170418081` class (attempt `2`, existing branch, empty
   SHAs, no PR number) still fails closed;
3. every mismatch class in `VOC-125-D03` fails closed;
4. automatic `remediate.yml` retry still forwards event-derived SHAs and now
   also forwards `pr_number` as `existing_pr_number`;
5. caller `pipeline.yml` (live and template) exposes `existing_pr_number` and
   forwards it, and does not expose operator SHA inputs;
6. attempt `1` with an existing open carrier fails closed;
7. VOC-121/VOC-123/VOC-124 isolation, lease, retry, and permission contracts
   remain.

Tests must not mint real App tokens, use secrets, or use production data.
Positive cases may use fixture remotes and recorded metadata, not live
#1012 mutation.

`VOC-125-D07`: Current-state comments in `implement.yml`, `remediate.yml`,
caller and template `pipeline.yml`, and `karsift-ai-infra/README.md` must
describe the operator resume route: attempt `2` plus `existing_pr_number`,
derivation of immutable head/base, and fail-closed mismatch classes. They
must not present free-form SHA paste as the operator interface. Historical
CHANGELOG entries stay unchanged except for a new current-state note if that
file's current-state section is the live contract. After the exact reviewed
infrastructure merge SHA is known, pin
`tooling/governance/fixtures/karsift-ai-infra/` when the mirrored fixture
consumes the changed workflow, helper, tests, or comments, and advance
matching caller pin assertions.

`VOC-125-D08`: This package is a hard dependency for resuming #1003 / #1012
and a distinct recovery-identity outcome. Do not implement VOC-122
promotion-recovery replan behavior here. After the exact reviewed infra merge
is live on `implement.yml@main` and the caller dispatch contract is merged,
resume the existing `VOC-122-T00` carrier with:

```bash
gh workflow run pipeline.yml --repo KARSIFT/vocanova-platform-sandbox --ref develop \
  -f action=implement \
  -f change_id=VOC-122 \
  -f package_path=specs/changes/VOC-122-promotion-recovery-must-replan-required-checks \
  -f task_id=VOC-122-T00 \
  -f issue_number=1003 \
  -f attempt=2 \
  -f existing_pr_number=1012
```

Do not create a replacement VOC-122 task or PR. Do not merge #1012 from this
package. Publishing, independently reviewing, and merging VOC-122's own
authoritative work remain the existing VOC-122 roster's work and are not
VOC-125-T00 completion gates. Record the handoff in `t00-evidence.md`.

`VOC-125-D09`: No bootstrap exception. VOC-124 already requested
`permission-workflows: write` on `publish-source` (infra merge
`f406cc95a3f853e8aef5bf8bcf22d37a29d64547`). T00's first run is attempt `1`
on a new VOC-125 carrier and uses the normal coordinated source-publication
path. Do not treat an untracked local `karsift-ai-infra/` checkout as this
repository's tracked tree.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. Operator resume does not
mint a broader token, does not grant Actions-write to the model-controlled
runner, and does not accept free-form SHAs as authority.

Abuse/process risks:

1. Operator resume remaining impossible after automatic remediation stops —
   mitigated by `VOC-125-D01` / `VOC-125-D02`.
2. Free-form SHA paste binding a retry to the wrong head — forbidden as the
   operator interface by `VOC-125-D01`; remaining SHA inputs on `implement.yml`
   must match the live PR (`VOC-125-D02` / `VOC-125-D03`).
3. Reclassifying attempt `2` as attempt `1` or opening a replacement PR —
   forbidden by `VOC-125-D02` / `VOC-125-D04`.
4. Foreign or stale review evidence steering remediation — fail closed by
   `VOC-125-D03`.
5. Printing App tokens, private keys, or secret values in logs, tests, or
   evidence — forbidden.

## Contradictions and open questions

1. **Helper and test file layout (`VOC-125-DEP-07`):** the required behavior
   is settled; whether derivation lives in a new
   `config/bind-existing-carrier.py` (or equivalent) or extends
   `verify-expected-head.py`, and whether tests are new `test_voc125_*.py`
   files or extensions of existing policy tests, is an implementation choice.
2. **PR-to-task binding fields:** the required identity set is settled (PR,
   branch, repository, task, package, authority issue, prior-review base/head
   when a review exists, current remote ref). Exact GitHub API fields used to
   prove task/package/issue binding should reuse the existing implementer PR
   body conventions (`Closes #N`, package path, task id) already consumed
   elsewhere; do not invent a parallel metadata dialect.
3. **Fixture pin applicability:** pin
   `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` to the exact
   reviewed infra merge when the mirrored fixture consumes the changed
   workflow, helper, tests, or comments. `implement.yml` is in the policy
   fixture subset, so consumption is expected. If some files are not in that
   subset, do not copy them merely to force a pin; record non-consumption.
4. **AGENTS.md:** that file currently documents `reconcile` /
   `reconcile-release` dispatch, not `action=implement`. T00 updates it only
   if implementation adds an operator implement-resume command there or if an
   existing sentence would become false. Do not expand AGENTS.md into a new
   runbook. Workflow comments and `karsift-ai-infra/README.md` are the
   current-state contract.
