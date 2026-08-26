# VOC-121 — Make governed automation preserve self-correction helpers and finish coordinated cross-repository tasks: Specification

## Objective and requirement source

Make the live governed implementation and promotion path finish authorized work
without silent loss or manual recovery. One adopted task must be able to complete
coordinated repository carriers, retain self-correction helpers after caller
staging, and recover required checks from GitHub's actual ruleset satisfaction
state.

**Requirement source:** [GitHub issue #994](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/994).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #994)

| Item | Value |
|------|-------|
| Adopted task that exposed the failures | `VOC-120-T00` |
| Implementation run | https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32899479806 |
| Implementation job | `97970681892` |
| Discard mechanism | `implement.yml` copies `config/run-app-checks.sh`, then `rm -rf karsift-ai-infra`, then a caller-only Git bundle |
| Self-correction failure | `python3: can't open file .../karsift-ai-infra/config/prepare_cursor_model.py` after the nested checkout was deleted |
| Promotion PR | [#993](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/993) |
| Cancelled required check-run | exact-head `governance-policy` pull-request run, cancelled by concurrency |
| Alternate successful evidence | another workflow-dispatch run and a published status of the same context |
| Release run | https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32902418610 (`actions-check-recovery: ... dispatched=none`) |
| Ruleset reality | `gh pr checks --required` still reported `governance-policy` failed; three merge attempts failed; rerunning the cancelled exact-head run succeeded |

## Scope and non-goals

### In scope

1. Give a single adopted task a fail-closed way to complete all authorized
   coordinated repository carriers, including a separate source PR and caller PR
   when the repository boundary requires them.
2. Commit and publish source work from isolated repository state. Never smuggle
   the nested checkout into the caller as a gitlink. Never silently discard
   authorized nested edits.
3. If general multi-carrier support cannot be made safe in this change without
   weakening runner isolation, App-token least privilege, exact-SHA review, or
   protected checks, detect authorized edits in the disposable policy checkout
   and fail closed with precise recovery instructions. Silent loss is
   unacceptable either way.
4. Keep cross-repository PR bodies on fully qualified non-closing caller
   references (`Relates to OWNER/CALLER#N`).
5. Preserve independent review, exact-head checks, merge order, exact source
   merge SHA capture, caller fixture/pin reconciliation, evidence, one bounded
   remediation retry, and caller completion as fail-closed controls.
6. Preserve immutable access to every model-resolution, retry, and check helper
   self-correction needs after caller staging, without allowing a nested
   checkout/gitlink into caller commits.
7. Make required-check recovery follow GitHub's actual branch/ruleset
   satisfaction state. Rerun or redispatch a cancelled or failed required
   check-run on the unchanged exact head even when another run or same-named
   status is successful. Do not treat status attestation as overriding a
   check-run when GitHub does not.
8. Add deterministic tests that reproduce all three live failures and prove the
   corrected fail-closed behavior, including cross-repository publisher
   races/stale heads, deletion-before-self-correction, and cancelled-check
   selection.
9. Update current-state workflow comments/docs and the caller fixture/pin after
   the authoritative infrastructure merge SHA is known.
10. Run the complete relevant infrastructure policy suite and caller
    governance/fixture validation, with independent exact-revision review for
    each repository carrier.

### Non-goals / explicitly excluded

- Changing application runtime behavior, deployment topology, product
  permissions, or monitor inventory.
- Introducing an OpenAI/Codex route or an `OPENAI_API_KEY` requirement.
- Weakening missing-Cursor-credential handling, invalid model configuration
  handling, unsupported-provider fail-closed behavior, secrets handling, the
  existing two-attempt bound, exact-SHA protections, or review independence.
- Splitting the three findings into one issue or task per workflow or file.
- Treating historical logs as permission to weaken any gate.
- Self-adoption or self-authorization of this package.
- Using closed issue state as completion proof.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: reusable CI/CD workflows (`implement.yml`, `release.yml`,
  `merge-gate.yml`, recovery runners), App-token issuance for publication,
  `tooling/governance/` fixtures and tests, and shared-infra policy modules
  that select required-check evidence.
- Protected technical effect: how authorized implementer work is committed,
  published, self-corrected, and how promotion recovery decides a required
  check is satisfied. No application runtime effect is intended.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but workflow and governance-fixture changes still
  require exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-121-D00`: This is one outcome-sized reliability change. Use one end-to-end
implementation task covering infrastructure source, caller fixture/pin,
tests, current-state docs/comments, and evidence. Coordinated pull requests in
`KARSIFT/karsift-ai-infra` and this caller remain one task. Repository count,
workflow count, and file count are not split reasons.

`VOC-121-D01`: Preferred contract: a single adopted task completes every
authorized coordinated carrier end to end. Source work is isolated, bundled,
published, independently reviewed, merged, and reconciled by exact source merge
SHA into the caller fixture/pin when the fixture consumes the change.

If implementation proves that general multi-carrier support cannot be made safe
in this change without weakening runner isolation, App-token least privilege,
exact-SHA review, or protected checks, the same task must detect authorized
edits in the disposable policy checkout and fail closed with precise recovery
instructions. That fallback still forbids silent deletion. Record the chosen
path and the concrete safety constraint in `t00-evidence.md`. Do not guess a
plausible-looking publisher design that mixes the model-controlled runner with
an App token or forges a second-repo PR from caller gitlinks.

`VOC-121-D02`: Source commits are created from isolated repository state for
that source. The nested `karsift-ai-infra/` checkout remains disposable policy
input for the caller worktree. `git add -A` in the caller must never stage that
nested `.git` as a gitlink. Caller publication remains a credential-free bundle
consumed by a clean publisher job.

`VOC-121-D03`: The caller implementation PR keeps the local `Closes #N` binding
expected by merge-gate. Any PR, evidence, commit, or comment produced in a
different repository must use `Relates to OWNER/CALLER#N` and must not use a
GitHub closing keyword before that caller issue reference.

`VOC-121-D04`: Independent exact-SHA review is required for each repository
carrier. Merge order is infrastructure first when the caller fixture/pin
consumes the change. Capture the exact source merge SHA. Bounded retry remains
two attempts. Caller completion stays fail-closed and cannot be manufactured
from a foreign-repository closing keyword.

`VOC-121-D05`: Before the nested infrastructure checkout is removed from the
caller worktree, copy every helper later self-correction or re-validation
invokes to an immutable location outside that checkout. At minimum this
includes `config/run-app-checks.sh` (already copied) and
`config/prepare_cursor_model.py` (currently invoked after deletion). If
self-correction also needs `config/retry-helpers.sh` or another
model-resolution/retry/check helper, preserve those the same way. After
staging, self-correction must still fail closed on missing Cursor credentials,
invalid model configuration, and unsupported providers. Nested checkout and
gitlink must still be removed before caller `git add -A`.

`VOC-121-D06`: Do not introduce an OpenAI/Codex execution path or an
`OPENAI_API_KEY` requirement. Do not print credentials. Missing `CURSOR_API_KEY`
on Cursor-backed paths continues to fail closed. Existing two-attempt bounds
and exact-SHA protections remain unchanged or stronger.

`VOC-121-D07`: Required-check recovery must use GitHub's actual branch/ruleset
satisfaction state, not "newest same-named success" and not a published commit
status that GitHub does not treat as overriding a check-run. A cancelled or
failed required check-run on the unchanged exact head must be rerun or
redispatched even when:

- another workflow-dispatch or non-PR run of the same workflow/context
  succeeded; or
- a same-named commit status is successful.

Status attestation may continue only where GitHub actually treats it as
satisfying the ruleset. It must not be treated as sufficient while
`gh pr checks --required` (or the equivalent authenticated required-check
view GitHub uses for merge) still reports that context failed or cancelled.
The #993 recovery that worked — rerunning the cancelled exact-head
`governance-policy` run — is the required behavior class.

`VOC-121-D08`: Deterministic tests must reproduce:

1. authorized nested infrastructure edits being discarded by caller-only
   staging, including publisher races and stale heads for a second carrier;
2. deletion of the nested checkout before self-correction invokes
   `prepare_cursor_model.py` (and any other preserved helper);
3. a cancelled exact-head required check-run remaining unsatisfied when an
   alternate successful run or same-named status exists.

Positive cases must prove the corrected fail-closed behavior. Tests must not
use secrets or production data.

`VOC-121-D09`: Current-state comments in `implement.yml`, recovery/release
docs, and `karsift-ai-infra/README.md` must stop describing the nested checkout
as safely disposable when authorized edits exist, and must stop claiming that
same-SHA status attestation satisfies a cancelled required check-run if GitHub
does not. After the authoritative infrastructure merge SHA is known, pin
`tooling/governance/fixtures/karsift-ai-infra/` when the fixture consumes the
change, or record explicit non-consumption.

`VOC-121-D10`: T00 is itself a coordinated infrastructure-plus-caller change.
Until this package's infrastructure merge is live on `@main`, current
`implement.yml` may still discard nested source edits (chicken-and-egg).
Bootstrap recovery of this task's own infrastructure carrier is in-scope causal
work under the same task, not a new package. After T00 merges, future adopted
tasks must not require that manual path.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. Cursor paths continue to
require `CURSOR_API_KEY`. The model-controlled implementer runner must still
never receive the GitHub App token. A second-repository publisher, if landed,
must mint least-privilege credentials on a clean runner scoped only to the
source repository it publishes, with the same bundle/isolation pattern as the
existing caller `publish` job.

Abuse/process risks:

1. Silent discard of authorized source edits — mitigated by isolated
   publication or fail-loud detection (`VOC-121-D01`, `VOC-121-D02`).
2. App token or workflow-file publication from a model-controlled runner —
   mitigated by keeping publication on a clean runner and preserving the
   existing caller refusal to publish caller `.github/workflows/**` from that
   unreviewed path. Infrastructure workflow files belong to the infrastructure
   repository carrier and are independently reviewed there.
3. Treating a successful status as a required check-run — mitigated by
   `VOC-121-D07` and cancelled-check tests.
4. Helper preservation accidentally reintroducing a gitlink — mitigated by
   copying helpers out, then still deleting the nested checkout before caller
   staging (`VOC-121-D05`).

## Contradictions and open questions

1. **Full multi-carrier versus fail-loud floor (`VOC-121-DEP-06`):** the required
   user outcome is end-to-end completion of authorized carriers. The issue also
   states a fail-loud floor if general multi-carrier support cannot be made
   safe in one change. T00 must prefer the full path. The fallback is allowed
   only with a recorded safety constraint; it does not count as "carriers
   completed automatically."
2. **Second-repository App token scope:** the current caller publisher mints a
   token for `${{ github.event.repository.name }}` only. A source-repo
   publisher needs a least-privilege token for `KARSIFT/karsift-ai-infra`
   without broadening the caller token or returning any App token to the
   implementer runner. Exact job graph is an implementation choice inside
   `VOC-121-D01` / `VOC-121-D02`.
3. **Ruleset satisfaction probe:** implementation may use `gh pr checks
   --required`, the authenticated checks API constrained to required contexts,
   or another probe that matches GitHub merge/ruleset evaluation. The contract
   is behavioral: if GitHub still selects a cancelled or failed required
   check-run, recovery must rerun or redispatch that exact-head run.
4. **Chicken-and-egg for this package:** T00 cannot assume the not-yet-merged
   publisher already exists while implementing itself (`VOC-121-D10`).
5. **Fixture pin applicability:** pin
   `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` to the exact
   reviewed infra merge when the mirrored fixture consumes the changed
   workflows, recovery modules, or tests. If some files are not in the policy
   fixture subset, do not copy them merely to force a pin; record
   non-consumption. Workflow and recovery-module changes in this package are
   expected to be consumed.
