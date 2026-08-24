# VOC-114-T01 — Operator live recovery evidence

Date: `2026-08-24`

This record contains allowlisted metadata only. No Actions logs, credentials,
token values, or application data are included. Auto-advance created draft
carrier PR `#969`; no implementer run was started for this operator-owned task.

## Exact identities

- Evidence carrier: PR `#969`, branch
  `agent/voc-114-voc-114-t01`.
- Initial live-proof integration and promotion head:
  `c718f6b49ad6a9a4f1d26eb4319347f6220a8d54`.
- Final trusted caller and promotion head after consuming all causal fixes:
  `0b0d866533e0100f6dfe37e3109f040ddde37bd6`.
- Promotion PR: `#947`, `develop` → `main`, open and mergeable when observed.
- T00 shared-infra baseline: PR `#136`, reviewed head
  `72b3742d5cd3ed1561534908d34286869befbe53`, merge
  `30cc0a6f443b95e45527b03094767b8357b0a2dc`.
- T00 caller baseline: PR `#961`, reviewed head
  `675cb8448f33f9226e1eb0b874a4e4407d1d321a`, merge
  `c718f6b49ad6a9a4f1d26eb4319347f6220a8d54`.

## Live-proof corrections kept in the same outcome

Hosted proof exposed additional defects that were causal to this exact recovery
outcome. They were corrected under VOC-114 rather than split into new plans or
tasks:

- shared-infra PR `#137`: reviewed head
  `227ab75db1aef59e9ef1ec2cb64ffcd880652823`, merge
  `053ad6f396113b306822d749ae8db26194a00ec6`; removed unsupported `--repo`
  from the recovery runner's `gh api` calls;
- shared-infra PR `#138`: reviewed head
  `7fa9a328e332628c0162aefc6a247e500c001929`, merge
  `3f4745006cb86eb766913896a20fd399c539c72e`; bound promotion suppression
  to required contexts and made completion independent of unrelated checks;
- shared-infra PR `#139`: reviewed head
  `b5f3847826f59b3890679009ebd42c93dc90a117`, merge
  `2562b5463248308f285c35cf26aa838e2d3215e2`; corrected both hosted
  verifier adapters to use valid `gh api` repository context;
- shared-infra PR `#140`: reviewed head
  `4af16640091099821befec2df8497c4c5ed73f71`, merge
  `da61963aeaa0e566e499e63139132cbe86c3cd6b`; exposed the existing
  integration resolver/recovery pair to bounded operator dispatch without a
  free-form target SHA;
- shared-infra PR `#141`: reviewed head
  `37fadcaf7c10b1b73ee0463d31a7310c0d2985d4`, merge
  `4c0395aff2a4599160308f7f37c593b75c7394b6`; corrected the resolver's
  invalid `gh api --slurp --jq` combination;
- shared-infra PR `#142`: reviewed head
  `041d912c58d335ce3faea0d309b52e6d7d0389b3`, merge
  `bdc6736568827103b48255521f4bc83d5103bd3b`; made the final read-only
  verifier judge only the three authoritative promotion contexts instead of
  allowing unrelated failed or pending workflows to veto exact-head proof.
- shared-infra PR `#143`: reviewed head
  `b712752cbd20da6ae5d91853f593f989872f092a`, merge
  `9d7e334f917643c42bb4b7a062c8fcddecc7927f`; after release run
  `32724415871` failed at `workflow_dispatch_failed`, separated the recovery
  Actions token from the App mutation token. Recovery now uses job-scoped
  Actions write plus Checks/Statuses/Contents/Pull requests read, while the App
  token remains limited to PR/issue/content mutations.
- shared-infra PR `#144`: reviewed head
  `46c4cc1ceae5845b49e57e38d6d7fa399ed73ff4`, merge
  `6999e2beda5bbf00028fae04ca0e65324fc59afa`; added the D07
  ruleset-attestation bridge after genuine recovery passed but GitHub still
  reported the promotion PR blocked.
- shared-infra PR `#145`: reviewed head
  `73a63d42345ff54619d257e4857fc4166a2785af`, merge
  `c5d8bccfa8676bd367b53ad5f6f9a51a40c99405`; corrected the project template
  so only release receives Statuses write and added job-scoped negative tests.

The caller fixture first pinned token separation at shared-infra merge
`9d7e334f917643c42bb4b7a062c8fcddecc7927f` and now pins the complete D07
contract, including the template permission correction, at
`c5d8bccfa8676bd367b53ad5f6f9a51a40c99405`. Caller corrective PR `#970`
merged as `c3455941463c0ded5630ea309b50f94a6cd546af`; final pin-sync PR `#971`
merged as `172648f555b0eacedeb44fef707e6edf3cc60372`; promotion recovery provenance
PR `#972` merged as `51c4d261d940c0e96a66238992c5380384729bb2`; final
push-provenance refinement PR `#973` merged as
`0b0d866533e0100f6dfe37e3109f040ddde37bd6`; token-sync PR `#975` merged as
`f18acca8322131eadfaf9bc963352f8980e9d6f7`.
The caller recovery workflows set repository context before their pre-checkout
PR reads. Exact-`develop` promotion recovery and the canonical same-repository
`develop` → `main` promotion PR use `squash-safe-push`, matching the immutable
integration tip. Ordinary pull-request events retain `pr-validation` and strict
`pr-ancestry` for original fixture-changing changes; forks and other branch pairs
cannot select the promotion exception.

## Integration-push proof

Operator workflow-dispatch run `32715579496` ran on carrier head
`a58f654ff0d8c629146080184b0d9750ab95f45c` and resolved the current
`develop` head internally; the operator supplied no target SHA.

- `resolve-integration-recovery-target`, job `97395972560`: `success`;
- `recover-integration-push / recover`, job `97396003037`: `success`;
- legacy App token mint step: `success` (this idempotent no-op did not exercise
  the later-failing Actions dispatch endpoint and is superseded by PR `#143`);
- exact-SHA recovery step: `success`;
- observed integration SHA:
  `c718f6b49ad6a9a4f1d26eb4319347f6220a8d54`.

The recovery preflight observed the already-successful genuine workflow evidence
for that exact head and completed without fabricating a status. This is the
intended idempotent no-op result after the required integration workflows are
already green.

## Promotion recovery and authoritative selection

Operator reconcile run `32714169687` completed successfully:

- release identify job `97391797255`: `success`;
- release converge job `97391833289`: `success`;
- legacy App-backed recovery step: `success` because required evidence was already
  present; it did not prove an App-authenticated missing-check dispatch;
- authoritative promotion-check selection step: `success`.

The newest authoritative genuine GitHub Actions checks selected for promotion
head `c718f6b49ad6a9a4f1d26eb4319347f6220a8d54` were:

- `governance-policy`, check/job `97386188937`: `success`;
- `validate`, check/job `97391497933`: `success`;
- `ci / ci`, check/job `97386226716`: `success`.

Earlier failed `validate` attempts remain visible as history and were not
rewritten. The authoritative selector chose the newer successful exact-head
attempt. PR `#947` remained open and mergeable; no manual merge or synthetic
status was performed.

## Sanitized intermediate findings

- run `32712323463`: recovery stopped with the sanitized metadata-read failure
  before the invalid `gh api --repo` correction;
- run `32713089936`: cancelled after promotion dispatch suppression waited on
  its own recovery workflow rather than the required context;
- run `32713755797`, job `97390553833`: `validate` failed because pre-checkout
  PR resolution lacked repository context;
- verifier run `32714394902`, job `97392466710`: failed in the read-only verify
  step before the hosted verifier CLI correction;
- integration run `32715323592`, resolver job `97395209442`: failed before
  recovery because `gh api` rejects combined `--slurp` and `--jq` flags;
- verifier run `32719387468`, job `97407364112`: failed closed after the
  caller correction advanced the promotion head and invalidated the older
  exact-head check set;
- verifier run `32719947496`, job `97409037642`: failed closed because the
  verifier still counted an unrelated failed `release / converge` workflow,
  leading to shared-infra PR `#142` and caller pin-sync PR `#971`;
- verifier run `32721444405`, job `97413542354`: failed closed because the
  recovered `validate` context used original-PR ancestry rules against an
  already-squashed promotion history, leading to caller PR `#972`;
- recovery validate run `32722352070`, job `97416266201`: failed closed after
  merge-base PR validation proved unsuitable when the old production merge base
  predates the navigator skill, leading to caller refinement PR `#973`;
- final promotion recovery validate run `32723140094`, job `97418637552`:
  `success` on exact promotion head
  `0b0d866533e0100f6dfe37e3109f040ddde37bd6` using the corrected
  `squash-safe-push` path.
- post-carrier release run `32724415871`, release-converge job `97423057295`:
  failed closed at `workflow_dispatch_failed`. Repository installation metadata
  confirmed `karsift-ai-infra-bot` has Contents/Issues/Pull requests/Workflows
  permissions but no Actions permission. Manual diagnostic dispatches
  `32724752006`, `32724766278`, and `32724769924` all succeeded on exact head
  `a07ea8c0cdaf060ff8a75db2b1436eebeecf2d52`, isolating authorization—not the
  allowlist, workflow inputs, or target identity—as the remaining defect.
- caller correction PR `#975` passed exact-head CI and independent review, then
  merged to `develop` as `f18acca8322131eadfaf9bc963352f8980e9d6f7`.
  Its merge-gate recovery step successfully used the separated job token; only
  the pre-existing duplicate immutable task marker failed after the merge.
- release run `32727465136`, converge job `97432766968`: job permissions exposed
  Actions write plus Checks/Statuses read, the App remained mutation-only, and
  promotion recovery successfully dispatched exact-head Repository Governance
  run `32727780544` and pipeline run `32727783329` for `f18acca…`; both passed.
  Recovery and newest-authoritative-check selection succeeded. The final merge
  failed because GitHub's ruleset did not associate those successful
  `workflow_dispatch` runs with promotion PR #947's required contexts. No manual
  merge, ruleset bypass, or unbacked status was performed.
- shared-infra PR `#144` then implemented D07 in the same VOC-114-T00 outcome and
  merged as `6999e2beda5bbf00028fae04ca0e65324fc59afa`: release-only Statuses
  write, exact expected-workflow/PR/SHA validation, same-SHA ruleset
  attestations, future-selector exclusion, and ruleset-propagation retries.
- caller PR `#976` passed fresh exact-head CI and independent review after the
  reviewer caught and the same PR corrected a template permission inversion. It
  merged to `develop` as `dd7383ff4257632078bff46eebbbcfa7f2f1f451` and its
  integration-push recovery succeeded; shared-infra PR `#145` is the pinned
  immutable source correction for that finding.
- release run `32733575823`, converge job `97451846259`: genuine exact-head
  recovery, authoritative selection, and all three D07 ruleset attestations
  succeeded. GitHub still rejected the merge because the PR-associated
  Repository Governance `validate` check had already failed under original-PR
  ancestry semantics. A success status cannot erase a failed check run with the
  same required context. The same task therefore now selects `squash-safe-push`
  directly for the canonical same-repository `develop` → `main` PR while
  preserving strict modes for ordinary and fixture-changing PRs.

Each failure remained fail-closed and produced no fabricated check or manual
promotion merge.

## Deterministic validation

- shared-infra focused VOC-113/VOC-114 suite: `48` tests, pass;
- shared-infra full suite after the final correction: `265` tests, pass;
- caller focused Node recovery/pin/policy suite: `25` tests, pass;
- caller full foundation suite after package build: `338` tests, pass;
- pinned shared-infra fixture suite: `181` tests, pass;
- caller governance suite: `160` tests, pass;
- repository governance validation: pass;
- `git diff --check`: pass.

The contract-bound `verify-promotion-check-recovery / verify` dispatch is made
after this evidence commit so GitHub can bind the run to the carrier's exact PR
head. Its final run metadata is intentionally external to this self-referential
file and is consumed by `.karsift/live-evidence/VOC-114-T01.yaml`.
