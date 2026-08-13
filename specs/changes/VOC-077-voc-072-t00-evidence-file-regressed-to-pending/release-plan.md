# VOC-077 — Release Plan

## Release and deployment authorization

This package requests no new deployment authority and performs no
production secret or cutover action. Once adopted and implemented, the
task PR follows the existing governed path: independent review, then merge
into `develop` per `merge-gate.yml`, then — per AGENTS.md's "Release and
deployment authority" section — automatic promotion to `main` and automatic
production deployment once this package's task roster closes. Repository
text under `specs/changes/` has no user-facing runtime effect.

This package does not authorize VOC-072-T01 wiring or VOC-072-T02
`--verify-only` / `--apply`. Those remain VOC-072 tasks after the evidence
gate is accurate.

## Preconditions, monitoring, and outcome

Preconditions:

- Package adopted with `VOC-077-DEP-00` and `VOC-077-DEP-01` settled in
  writing (or explicitly deferred as follow-ups).
- Implementation authorized.
- Redacted `gh secret list` still shows
  `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` at implementation time.
- Independent verification PASS (or PASS WITH NON-BLOCKING FINDINGS) on the
  exact implemented revision.
- No founder approval required solely for proposed R1 under active A-003;
  merge-gate still requires positive PASS proof (founder `approved` alone
  does not override FAIL — as demonstrated on PR #558).

Monitoring after merge:

- VOC-072-T01 may proceed from an evidence-gate perspective (secret present
  + evidence resolved).
- PR #558 / issue #543 disposition matches adoption's DEP-00 choice.
- No secret values appear in the merged commit.

Outcome owner: implementer records `VOC-077-EV-00`; adopting human owns
DEP decisions.

## Rollback

Trigger: resolved evidence while secret absent; secret leakage; or
out-of-scope production-path edits.

Mechanism: revert the implementation commit(s) for `VOC-077-T00`.

Validation: governance scripts pass on the reverted tree; targeted VOC-072
files match the intended rollback tip.

Accountable owner: implementer of `VOC-077-T00`.

Last-known-good reference: VOC-072 evidence / `change.yaml` /
`acceptance-criteria.md` revisions immediately preceding this package's
implementation merge (with the caveat in `implementation-plan.md` that the
pre-change tip may be the incorrect pending regression).

## Independent verification, human approvals, and closure

Independent verification must confirm, against the exact implemented
revision's commit SHA:

- Adoption decisions for open questions were followed.
- `VOC-077-AC-00` through `VOC-077-AC-02` hold with linked evidence.
- Implementer-role occupant did not approve or merge its own
  implementation.
- Active authority model remains `a003-active`.
- No still-required R4 / EHR gate applies to this evidence-only change.
- Codex (or current implementer-role occupant) did not self-approve.

Under active A-003, no standing technical-steward approval is assumed for
proposed R1. Repository merge into `develop` is not the same event as
VOC-072 package closure — VOC-072-T01/T02 remain open until those tasks
complete under VOC-072's own acceptance criteria.

Closure of **this** package requires VOC-077 acceptance criteria recorded
as passing with linked evidence (`VOC-077-EV-00`).
