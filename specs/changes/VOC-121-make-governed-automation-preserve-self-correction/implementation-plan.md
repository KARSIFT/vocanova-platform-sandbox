# VOC-121 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `KARSIFT/karsift-ai-infra` implement/release/merge-gate/
  recovery workflows and Python policy modules; App-token publication;
  `prompts/implement.md`; infra README current-state comments; caller
  `tooling/governance/` fixtures and tests.
- Prerequisites: confirm the live `Commit implementer's work` /
  `Self-correct failed pre-push validation` ordering, the caller-only bundle
  publisher, and how `select_gate_evidence` /
  `suppress_active_or_successful_dispatches` /
  `promotion_status_attestation` treat cancelled check-runs versus statuses.
- Chicken-and-egg (`VOC-121-D10`): T00 cannot assume the not-yet-merged
  multi-carrier publisher already exists while implementing itself. Bootstrap
  recovery of this task's own infrastructure carrier is in-scope.
- Preserve two-attempt bounds, exact-SHA independent review, risk
  classification, fail-closed credentials, and runner/App-token isolation.
- Do not treat an untracked local `karsift-ai-infra/` checkout as this
  repository's tracked tree.

## File reconciliation and implementation sequence

### T00 — Preserve self-correction helpers, finish coordinated carriers without silent loss, and recover required checks from GitHub satisfaction state

| Target | Action | Notes |
|--------|--------|-------|
| `KARSIFT/karsift-ai-infra/.github/workflows/implement.yml` | modify | Preserve helpers before nested-checkout removal; stop silent discard of authorized nested edits; isolated second-carrier bundle/publish or fail-loud |
| `KARSIFT/karsift-ai-infra` publisher / token minting adjacent to the existing clean `publish` job | modify/add as needed | Clean runner only; least-privilege source-repo scope; no App token on the implementer runner |
| `KARSIFT/karsift-ai-infra/prompts/implement.md` | modify if needed | Keep coordinated-carrier permission aligned with the execution contract |
| `KARSIFT/karsift-ai-infra/config/actions_check_recovery.py` and `actions-check-recovery-runner.py` | modify | Do not treat alternate success or same-named status as satisfying a cancelled/failed required check-run |
| `KARSIFT/karsift-ai-infra/config/authoritative_checks.py` | modify if needed | Selection must not hide a ruleset-selected cancelled required check-run |
| `KARSIFT/karsift-ai-infra/config/promotion_status_attestation.py` and runner | modify if needed | Status attestation must not override a check-run GitHub still selects |
| `KARSIFT/karsift-ai-infra/.github/workflows/release.yml` and `merge-gate.yml` / `recover-actions-checks.yml` | modify as needed | Same satisfaction contract on every recovery caller |
| `KARSIFT/karsift-ai-infra/README.md` and implement.yml header comments | modify | Current-state: no silent discard; no false status-overrides-cancelled-check claim |
| `KARSIFT/karsift-ai-infra/tests/*` | create/extend | Reproduce all three live failures and publisher races/stale heads |
| `tooling/governance/fixtures/karsift-ai-infra/**` | sync/pin | Exact reviewed infra merge when consumed |
| `tooling/governance/tests/*` | modify/extend | Caller regressions for the same contracts |
| `specs/changes/VOC-121-.../t00-evidence.md` | create/update | Record path chosen, commands, results, infra SHA, pin applicability |

Ordered steps:

1. In `KARSIFT/karsift-ai-infra`, preserve every helper self-correction later
   invokes by copying it to an immutable location before `rm -rf
   karsift-ai-infra`. Point the self-correction step at those copies. Keep
   deleting the nested checkout before caller `git add -A`.
2. Detect authorized nested source edits. Implement isolated source
   commit/bundle/clean-publisher/PR with fully qualified non-closing caller
   references, independent review, merge-order, and exact SHA capture. If that
   general multi-carrier path cannot be made safe without weakening isolation
   or privilege, fail closed with precise recovery instructions instead of
   deleting the edits. Record the chosen path.
3. Change required-check recovery so a cancelled or failed required check-run
   on the unchanged exact head is rerun or redispatched even when another run
   or same-named status succeeded. Do not attest the context satisfied while
   GitHub still reports it unsatisfied.
4. Add deterministic tests for nested-edit discard, publisher races/stale
   heads, deletion-before-self-correction, and cancelled-check selection.
5. Update current-state comments/docs so they describe the new contract.
6. Run the infra unit/policy suite. Open one reviewed infra PR that
   `Relates to KARSIFT/vocanova-platform-sandbox#<task>` and does not use a
   closing keyword. Merge it first when the caller fixture consumes the
   change.
7. Sync and pin the caller fixture to that exact merge SHA when consumed;
   update caller governance tests; record evidence in `t00-evidence.md`.

## Validation and independent verification

Deterministic commands, as applicable to the final changed file set:

```bash
# In the checked-out primary KARSIFT/karsift-ai-infra source:
python3 -m unittest discover -s tests -p 'test_*.py'

# In this caller repository:
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
git diff --check
```

If implementation adds narrower targeted commands (for example bundle,
self-correction helper, or cancelled-check recovery tests), record the exact
commands in `t00-evidence.md` and run them in addition to the suite above.

Independent verifier (exact reviewed caller SHA, and exact reviewed infra SHA
when an infra PR is opened) should confirm:

- authorized nested edits are published in isolation or fail loudly, never
  silently discarded;
- caller commits contain no `karsift-ai-infra` gitlink;
- self-correction can resolve Cursor models after nested-checkout removal;
- cancelled required check-runs are recovered on the exact head despite
  alternate successes or statuses;
- no OpenAI/Codex path or `OPENAI_API_KEY` requirement was introduced;
- exact-SHA review, risk floors, protected checks, App-token isolation, and
  two-attempt limits remain;
- the caller fixture pin equals the exact reviewed infra merge when the
  fixture consumes the change, or evidence records why the pin was not
  applicable;
- the implementer did not approve or merge its own work on either carrier.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime.
- **Operational effect:** Future implement/release runs keep coordinated
  source work, self-correction helpers, and required-check recovery fail-closed.
- **Rollback trigger:** False-positive fail-loud blocks legitimate caller-only
  tasks; second-carrier publication is unsafe; self-correction cannot find
  helpers; or recovery redispatches required checks incorrectly.
- **Rollback mechanism:** Revert the infra and caller fixture/test/doc
  changes to the prior reviewed implement/recovery behavior.
- **Last-known-good reference:** Current `implement.yml` / recovery modules on
  `main`/`develop` before VOC-121 implementation lands.
