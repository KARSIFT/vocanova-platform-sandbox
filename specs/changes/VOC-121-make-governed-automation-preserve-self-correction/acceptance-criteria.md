# VOC-121 — Acceptance Criteria

## VOC-121-AC-00 — Authorized nested source edits are never silently discarded

- Requirement source: `VOC-121-D01`, `VOC-121-D02`, `VOC-121-D10`
- Tasks: `VOC-121-T00`
- Tests: `VOC-121-TEST-00`, `VOC-121-TEST-01`
- Evidence: `VOC-121-EV-00`
- Result: pending

On the exact reviewed infrastructure revision, an implementer that makes
authorized edits in the nested `karsift-ai-infra` checkout plus caller edits
does not lose the source edits. Either:

1. the source carrier is committed and published from isolated repository
   state as its own PR; or
2. the workflow fails closed before deleting those edits and prints precise
   recovery instructions.

A caller-only bundle after `rm -rf karsift-ai-infra` with no source PR and no
failure is a failing result.

## VOC-121-AC-01 — Source publication is isolated and never a caller gitlink

- Requirement source: `VOC-121-D02`, `VOC-121-D04`
- Tasks: `VOC-121-T00`
- Tests: `VOC-121-TEST-01`, `VOC-121-TEST-02`
- Evidence: `VOC-121-EV-00`
- Result: pending

Source work is not smuggled into the caller commit as a gitlink, submodule, or
nested `.git`. Caller publication remains a credential-free Git bundle consumed
by a clean publisher. When a second carrier is published, that publisher
rejects stale heads and races fail closed. Independent exact-SHA review is
required for each repository carrier. Infrastructure merges first when the
caller fixture/pin consumes the change, and the exact source merge SHA is
captured.

## VOC-121-AC-02 — Cross-repository references never close the caller task

- Requirement source: `VOC-121-D03`
- Tasks: `VOC-121-T00`
- Tests: `VOC-121-TEST-04`
- Evidence: `VOC-121-EV-00`
- Result: pending

The caller implementation PR keeps local `Closes #N`. Any PR body, comment, or
evidence produced in `KARSIFT/karsift-ai-infra` uses
`Relates to KARSIFT/vocanova-platform-sandbox#N` and does not use a GitHub
closing keyword before that caller issue reference.

## VOC-121-AC-03 — Self-correction retains immutable helpers after caller staging

- Requirement source: `VOC-121-D05`, `VOC-121-D06`
- Tasks: `VOC-121-T00`
- Tests: `VOC-121-TEST-03`
- Evidence: `VOC-121-EV-00`
- Result: pending

After caller staging removes the nested `karsift-ai-infra` checkout,
self-correction can still invoke every model-resolution, retry, and check
helper it needs, including `prepare_cursor_model.py` and `run-app-checks.sh`.
Those helpers are served from an immutable copy outside the deleted checkout.
Missing `CURSOR_API_KEY`, invalid model configuration, and unsupported
providers still fail closed. No nested checkout or gitlink enters the caller
commit.

## VOC-121-AC-04 — Required-check recovery follows GitHub ruleset satisfaction

- Requirement source: `VOC-121-D07`
- Tasks: `VOC-121-T00`
- Tests: `VOC-121-TEST-05`, `VOC-121-TEST-06`
- Evidence: `VOC-121-EV-00`
- Result: pending

When an exact-head required check-run is cancelled or failed, recovery reruns
or redispatches that check on the unchanged exact head even if another run of
the same context succeeded or a same-named commit status is successful. Status
attestation is not treated as overriding a check-run when GitHub does not.
`dispatched=none` is incorrect while GitHub still reports the required
context unsatisfied (`gh pr checks --required` or the equivalent merge/ruleset
view).

## VOC-121-AC-05 — Existing safety gates remain unchanged or stronger

- Requirement source: `VOC-121-D04`, `VOC-121-D06`
- Tasks: `VOC-121-T00`
- Tests: `VOC-121-TEST-07`
- Evidence: `VOC-121-EV-00`
- Result: pending

The change does not weaken exact-SHA independent verification, deterministic
risk floors, protected checks, review independence, secrets handling, or the
existing two-attempt remediation bound. It does not introduce an OpenAI/Codex
route or an `OPENAI_API_KEY` requirement. The model-controlled implementer
runner still never receives the GitHub App token.

## VOC-121-AC-06 — Deterministic tests reproduce the three live failures

- Requirement source: `VOC-121-D08`
- Tasks: `VOC-121-T00`
- Tests: `VOC-121-TEST-00` through `VOC-121-TEST-06`
- Evidence: `VOC-121-EV-00`
- Result: pending

Infrastructure and caller suites include deterministic coverage for:

1. silent nested-edit discard versus isolated publish or fail-loud;
2. cross-repository publisher races and stale heads;
3. deletion-before-self-correction helper loss;
4. cancelled required check-run selection despite a successful alternate run
   or same-named status.

Tests do not use secrets or production data.

## VOC-121-AC-07 — Current-state docs and caller pin follow the reviewed infra merge

- Requirement source: `VOC-121-D09`, `VOC-121-D04`
- Tasks: `VOC-121-T00`
- Tests: `VOC-121-TEST-08`
- Evidence: `VOC-121-EV-00`
- Result: pending

Current-state workflow comments and infra README/docs no longer claim that
authorized nested edits are safely discarded, or that same-SHA status
attestation satisfies a cancelled required check-run if GitHub does not. Both
repositories are validated on their exact reviewed revisions. If the caller
fixture consumes the infrastructure change, `PINNED_SHA.txt` equals that exact
infra merge SHA. If not, the pin is unchanged and non-consumption is recorded
in `t00-evidence.md`.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
