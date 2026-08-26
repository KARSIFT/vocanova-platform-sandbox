# VOC-126 — VOC-125 caller pipeline exceeds GitHub workflow_dispatch input limit: Specification

## Objective and requirement source

Make the live caller `pipeline.yml` and the authoritative project-repo
template GitHub-valid under the platform `workflow_dispatch` maximum of 25
inputs, while preserving VOC-125's `existing_pr_number` operator-resume
interface and every existing recovery/verifier capability, so adopted
VOC-125-T00 can be closed as superseded-unusable-carrier after a governed
replacement lands and existing VOC-122-T00 / #1003 / #1012 can be resumed
through a live `action=implement` route.

**Requirement source:** [GitHub issue #1025](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1025).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1025)

| Item | Value |
|------|-------|
| Adopted VOC-125 task blocked | [#1022](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1022) (`VOC-125-T00`); origin [#1020](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1020) |
| Unusable caller PR | [#1024](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1024) at `8621f12dd466edab37fddb86d4e5e0a348ed3609` |
| Infrastructure merge | `KARSIFT/karsift-ai-infra@1f1705dbad41729563b0ad1e878e4154e5511e93` |
| Actions run | `32977045898` — no jobs or logs |
| Annotation | `Invalid workflow file: .github/workflows/pipeline.yml#L1 — you may only define up to 25 inputs for a workflow_dispatch event` |
| VOC-125 template input count | 26 keys under `on.workflow_dispatch.inputs`, including `existing_pr_number` |
| Live caller on this drafting ref | 25 keys; `existing_pr_number` is not yet present |
| VOC-125 retry | final allowed retry already consumed remediating the source carrier |
| Downstream blocked carrier | VOC-122-T00 #1003 / draft PR #1012 |

## Scope and non-goals

### In scope

1. Keep `existing_pr_number` as the implement-only operator resume identity on
   the live caller `pipeline.yml` and the infrastructure project-repo pipeline
   template. Forward it on the `implement` job. Do not add operator-typed
   `expected_head_sha` or `expected_base_sha` inputs to any caller
   `workflow_dispatch`.
2. Relocate the coherent read-only verifier dispatch surface into a dedicated
   caller workflow so every live `workflow_dispatch` block, including the
   remaining `pipeline.yml` block, has at most 25 inputs:
   - `verify-auto-advance-live-evidence`
   - `verify-ready-for-review-reuse`
   - `verify-remediate-operator-ownership`
   - `verify-promotion-check-recovery`
   - `verify-post-promotion-workflow`
   and the dedicated inputs those jobs currently consume.
3. Keep mutating operator-loop actions on `pipeline.yml`: `implement`, `plan`,
   `reconcile`, `reconcile-release`, `reconcile-live-evidence`,
   `recover-integration-push`, and `recover-promotion-pr-checks`.
4. Preserve every existing fail-closed exact-head/base, two-attempt, review,
   risk, protected-check, App-token, Cursor-only, and publication-lease
   contract already required by VOC-121 through VOC-125. Do not delete an
   active recovery or verifier capability. Do not silently drop an input
   merely to get below 26.
5. Add deterministic source and caller tests, including an explicit
   maximum-input-count regression for every project-repo template and live
   caller workflow that declares `workflow_dispatch`.
6. Land the infrastructure template/test/doc repair through the normal
   coordinated source carrier first. Independently review and merge it.
   Record the exact merge SHA. Then update the live caller workflow, fixture,
   exact pin, tests, documentation, and evidence to that merge.
7. Replace/supersede caller PR #1024 only because its definition cannot enter
   the governed review pipeline. Preserve its audit trail. Close it only after
   the governed replacement exists and is reconciled. Do not merge #1024.
8. Close VOC-125 task #1022 / origin #1020 only when the live caller route is
   valid, reviewed, merged, and promoted. Then resume existing VOC-122
   #1003 / #1012 at attempt `2` with `existing_pr_number=1012`.
9. Update current-state workflow comments/docs that describe verifier dispatch
   through `pipeline.yml` so they name the dedicated workflow. Update
   `AGENTS.md` only if an existing sentence would become false.

### Non-goals / explicitly excluded

- Changing application runtime behavior, deployment topology, product
  permissions, or monitor inventory.
- Implementing VOC-122 promotion-recovery replan (`VOC-122-T00` / #1003)
  inside this package. That remains a distinct already-authorized outcome
  whose existing carrier is resumed after this repair is live.
- Merging or treating caller PR #1024 or #1012 as this package's
  implementation PR.
- Dispatching VOC-125-T00 as attempt `3`, resetting attempt `2` to attempt
  `1`, or deleting the existing VOC-125 or VOC-122 branches to start
  replacement carriers for those packages.
- Silently dropping, renaming-away, or repurposing an active recovery or
  verifier input merely to get below 26.
- Replacing named verifier scalars with a packed JSON/string interface in this
  package (`VOC-126-D01`).
- Normalizing live `live_evidence_mode` versus template
  `live_evidence_dispatch` as a drive-by unless one of those files must change
  for the relocation.
- Adding operator-typed free-form SHA inputs to any caller
  `workflow_dispatch`.
- Weakening exact-SHA review, risk floors, protected checks, App-token
  isolation, force-with-lease, retry caps, or fail-closed missing-binding
  behavior.
- Changing GitHub App installation permissions or rotating
  `KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY`.
- Reopening VOC-121/VOC-123/VOC-124 source-publication contracts or the
  VOC-125 `implement.yml` bind helper beyond documenting that they remain.
- OpenAI credentials or execution routes.
- Rewriting historical CHANGELOG, A-003, or VOC-075 audit records.
- Splitting workflow logic, tests, docs, infrastructure, caller pin, or
  evidence into separate tasks.
- Self-adoption or self-authorization of this package.
- A supervised bootstrap exception: VOC-124 already published
  `permission-workflows: write` on `publish-source`. The first T00 run is
  attempt `1`.
- Operator-owned live-evidence contracts: acceptance is deterministic tests,
  exact-SHA review, and recorded handoff for #1024 / #1022 / #1020 /
  #1003 / #1012.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: caller and template CI/CD dispatch contracts, read-only
  verifier routing, operator implement-resume identity, and caller
  `tooling/governance/` fixtures and tests.
- Protected technical effect: whether GitHub will accept the caller pipeline
  definition, whether an operator can resume an existing implementation PR,
  and whether an existing read-only verifier or mutating recovery capability
  remains reachable. No application runtime effect is intended.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but workflow and governance-fixture changes still
  require exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-126-D00`: This is one outcome-sized caller-dispatch validity change.
Use one end-to-end implementation task covering infrastructure template,
source tests, caller workflows, tests, current-state docs/comments, caller
fixture/pin, and evidence. Coordinated pull requests in
`KARSIFT/karsift-ai-infra` and this caller remain one task. Repository count,
file count, and workflow-versus-tests-versus-docs are not split reasons.
Superseding #1024 and handing off VOC-125 / VOC-122 are evidence of this
outcome, not additional VOC-126 tasks and not replacement VOC-125 or VOC-122
roster entries.

`VOC-126-D01`: Relocate the five read-only verifier jobs and their dedicated
inputs into a dedicated caller workflow (preferred file name
`pipeline-verify.yml`; exact name is `VOC-126-DEP-07`). Keep mutating
operator-loop actions on `pipeline.yml`. Do not pack related verifier scalars
into one unstructured JSON/string input in this package: GitHub does not
schema-validate JSON `workflow_dispatch` values, and a packed interface would
replace already-tested named inputs with a new parser. Do not silently drop
an input. After relocation, every live `workflow_dispatch` block MUST have at
most 25 inputs.

`VOC-126-D02`: Preserve VOC-125 operator resume identity. Live caller
`pipeline.yml` and the infrastructure project-repo pipeline template expose
implement-only `existing_pr_number` and forward it on the `implement` job.
Caller `workflow_dispatch` still does not expose operator-typed SHA inputs.
`implement.yml@main` already declares `existing_pr_number` at infrastructure
merge `1f1705d…` and remains the derivation and fail-closed locus. This
package does not reopen `VOC-125-D02` / `VOC-125-D03` mismatch classes.

`VOC-126-D03`: The dedicated verifier workflow is a thin wire, same as
`pipeline.yml`. It calls the existing reusable verifier workflows at `@main`
with the same named inputs those jobs currently forward. It uses read-only
job permissions (`actions: read`, `contents: read`, and `issues` /
`pull-requests` read where those jobs already declare them). It does not use
`secrets: inherit`, does not mint App tokens, does not grant `actions: write`,
and does not expose model credentials. Missing or malformed required inputs
for a selected verify action continue to fail closed inside the existing
reusable workflows.

`VOC-126-D04`: Keep `recover-integration-push` and
`recover-promotion-pr-checks` on `pipeline.yml`. Those jobs mutate
(Actions-write recovery and promotion-PR CI rerun). Moving them would mix
write-capable recovery into the read-only verifier workflow. `promotion_pr_number`
may exist on both workflows when recover stays on `pipeline.yml` and the two
VOC-113 verifiers move; that duplication is not an input drop.

`VOC-126-D05`: Preserve VOC-121 through VOC-125 fail-closed publication and
resume contracts: exact base/head SHA binding, nested-repository isolation,
no gitlink, named-ref source bundle, credential-free bundles, clean
`publish-source` App-token separation with `permission-workflows: write` only
on that mint, caller `publish` still omitting workflow-write and still
refusing `.github/workflows/**`, force-with-lease, two-attempt implementer
bound, source PR `Relates to OWNER/CALLER#N` with no closing keyword, caller
PR `Closes #N` only for this package's own task issue, Cursor Composer
implementer, Cursor Grok exact-revision review, no OpenAI route, no secrets
in bundles/logs/fixtures, no credential values printed.

`VOC-126-D06`: Deterministic tests must prove:

1. every project-repo template and live caller workflow that declares
   `workflow_dispatch` has at most 25 keys under `on.workflow_dispatch.inputs`
   (the #1025 / run `32977045898` class);
2. `existing_pr_number` remains on `pipeline.yml` (live and template) and is
   forwarded on `implement`, and neither file exposes operator SHA inputs;
3. the five read-only verifier jobs exist on the dedicated workflow, still
   call the same reusable workflows, and still forward the same named inputs;
4. `pipeline.yml` still exposes and routes `implement`, `plan`, `reconcile`,
   `reconcile-release`, `reconcile-live-evidence`,
   `recover-integration-push`, and `recover-promotion-pr-checks`;
5. the dedicated verifier workflow is read-only (no `secrets: inherit`, no
   `actions: write`, no App-token mint);
6. VOC-125 attempt caps, empty-binding fail-closed behavior, and
   `remediate.yml` SHA / `existing_pr_number` forwards remain;
7. VOC-121/VOC-123/VOC-124 isolation, lease, retry, and permission contracts
   remain.

Tests must not mint real App tokens, use secrets, or use production data.
Input-count assertions must inspect YAML keys, not only comments.

`VOC-126-D07`: Current-state comments in the template, live caller workflows,
tests that currently hard-code the `pipeline.yml` action-options list, and
`karsift-ai-infra/README.md` must describe operator resume as attempt `2`
plus `existing_pr_number` on `pipeline.yml`, and must describe the five
read-only verifiers as dispatching through the dedicated workflow. They must
not present a 26-input `pipeline.yml` as valid. Historical CHANGELOG entries
stay unchanged except for a new current-state note if that file's
current-state section is the live contract. After the exact reviewed
infrastructure merge SHA is known, pin
`tooling/governance/fixtures/karsift-ai-infra/` when the mirrored fixture
consumes the changed template, tests, or comments, and advance matching
caller pin assertions.

`VOC-126-D08`: This package is the governed replacement for unusable VOC-125
caller PR #1024 and a hard dependency for resuming #1003 / #1012. Do not
implement VOC-122 promotion-recovery replan behavior here. Do not merge
#1024 or #1012. Do not dispatch VOC-125-T00 as attempt `3`. This package's
implementation PR `Closes` only its own VOC-126 task issue. After the exact
reviewed infra merge is live and the caller dispatch contract is merged and
promoted:

1. close #1024 as superseded, with an audit comment naming the VOC-126 SHA
   and stating that #1024 was never reviewable because GitHub rejected its
   workflow definition;
2. close VOC-125 task #1022 and origin #1020 only then, with audit comments
   that the remaining VOC-125 caller contract is delivered by this package's
   exact reviewed merge — not by a VOC-125 completion marker bound to #1024,
   and not by treating a VOC-126 PR as a third VOC-125 attempt;
3. resume the existing `VOC-122-T00` carrier with:

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

Do not create a replacement VOC-122 task or PR. Publishing, independently
reviewing, and merging VOC-122's own authoritative work remain the existing
VOC-122 roster's work and are not VOC-126-T00 completion gates. Record the
handoff in `t00-evidence.md`.

`VOC-126-D09`: No bootstrap exception. VOC-124 already requested
`permission-workflows: write` on `publish-source`. T00's first run is attempt
`1` on a new VOC-126 carrier and uses the normal coordinated source-publication
path. Do not treat an untracked local `karsift-ai-infra/` checkout as this
repository's tracked tree. Do not treat infrastructure merge `1f1705d…` as
the pin target: that revision encodes the invalid 26-input template even
though `implement.yml` there is otherwise the VOC-125 bind contract.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. Relocating read-only
verifiers does not mint a broader token, does not grant Actions-write to a
verifier runner, and does not accept free-form SHAs as authority.

Abuse/process risks:

1. Caller pipeline remaining GitHub-invalid after adding `existing_pr_number`
   — mitigated by `VOC-126-D01` / `VOC-126-D06`.
2. Silently dropping a recovery or verifier input to get below 26 —
   forbidden by `VOC-126-D01` / `VOC-126-D04`.
3. Moving mutating recovery onto a read-only workflow or granting the
   verifier workflow `secrets: inherit` / `actions: write` — forbidden by
   `VOC-126-D03` / `VOC-126-D04`.
4. Retrying unusable PR #1024 or dispatching VOC-125 as attempt `3` —
   forbidden by `VOC-126-D08`.
5. Printing App tokens, private keys, or secret values in logs, tests, or
   evidence — forbidden.

## Contradictions and open questions

1. **Dedicated workflow file name (`VOC-126-DEP-07`):** the required split
   (read-only verifiers versus mutating operator loop) is settled; the exact
   file name is an implementation choice. Preferred: `pipeline-verify.yml`.
2. **Shared input names on the verifier workflow:** `verify-remediate-operator-ownership`
   currently reuses `change_id`, `task_id`, `package_path`,
   `live_evidence_run_id`, and `live_evidence_pr_number` from `pipeline.yml`.
   Preserve those names on the dedicated workflow rather than inventing a
   parallel dialect, unless a collision with `verify_change_id` /
   `verify_task_id` / `verify_package_path` would exceed 25 inputs — it will
   not, at the counts recorded in `VOC-126-DEP-02`.
3. **`live_evidence_mode` versus `live_evidence_dispatch`:** the live caller
   currently uses a choice input `live_evidence_mode`; the VOC-125 template
   uses boolean `live_evidence_dispatch`. That drift predates this defect and
   is not this package's objective. Preserve each file's existing
   reconcile-live-evidence input shape unless the relocation itself requires
   touching that input.
4. **Fixture pin applicability:** pin
   `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` to the exact
   reviewed infra merge when the mirrored fixture consumes the changed
   template, tests, or comments. The project-repo `pipeline.yml` template is
   in the fixture, so consumption is expected. Pin to this package's infra
   merge, not to `1f1705d…`. If some files are not in that subset, do not
   copy them merely to force a pin; record non-consumption.
5. **AGENTS.md:** that file currently documents `reconcile` /
   `reconcile-release` dispatch via `pipeline.yml`, which remains true. T00
   updates it only if an existing sentence would become false after verifier
   relocation. Do not expand AGENTS.md into a new verifier-dispatch runbook.
   Workflow comments and `karsift-ai-infra/README.md` are the current-state
   contract.
