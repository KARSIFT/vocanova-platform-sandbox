# VOC-121 — Impact Analysis

## Security and privacy

This package repairs governed automation. It does not introduce new secret
values, OAuth/session material, production-data access, or user-facing data
flows.

Security controls that must remain:

- The model-controlled implementer runner never receives the GitHub App token.
- Caller publication stays on a clean runner consuming a credential-free bundle.
- A second-repository publisher, if landed, mints least-privilege credentials
  on a clean runner scoped only to that source repository.
- Cursor-backed paths require `CURSOR_API_KEY` and fail closed when it is
  absent. Credentials are never printed.
- No OpenAI/Codex path or `OPENAI_API_KEY` requirement is added.
- Nested `karsift-ai-infra/.git` is never staged into the caller as a gitlink.
- Caller `.github/workflows/**` changes remain refused on the unreviewed
  automated caller publication path.

Helper copies used by self-correction are policy scripts, not credentials, and
must live outside the deleted nested checkout without reintroducing that
checkout into the caller commit.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect.

## Risks, dependencies, and evidence

- `VOC-121-R00`: **High delivery risk** if authorized nested infrastructure
  edits continue to be deleted after a successful implementer step. Mitigation:
  isolated source publication or fail-loud detection (`VOC-121-D01`,
  `VOC-121-D02`, `VOC-121-AC-00`).
- `VOC-121-R01`: **High privilege risk** if a second-repository publisher
  shares a runner with the implementer or broadens the caller App token.
  Mitigation: clean-runner isolation and least-privilege repository scope
  (`VOC-121-D02`).
- `VOC-121-R02`: **High self-correction risk** if staging still deletes
  `prepare_cursor_model.py` (or other required helpers) before the retry step.
  Mitigation: immutable helper copies before nested-checkout removal
  (`VOC-121-D05`, `VOC-121-AC-03`).
- `VOC-121-R03`: **High merge-safety risk** if recovery again treats a
  successful same-named status or alternate run as satisfying a cancelled
  required check-run. Mitigation: GitHub actual satisfaction state and
  rerun/redispatch of the cancelled exact-head run (`VOC-121-D07`,
  `VOC-121-AC-04`).
- `VOC-121-R04`: **Medium documentation risk** if README/workflow comments
  continue to describe the buggy discard and status-attestation bridge as
  current-state. Mitigation: same-task current-state updates (`VOC-121-D09`).
- `VOC-121-R05`: **Low release risk** because no application runtime
  deployment change is intended; rollback is workflow/config/test reversion.
  Restored promotion recovery may merge an already-reviewed promotion PR
  through existing gates after genuine checks.
- Protected surfaces: `KARSIFT/karsift-ai-infra` implement/release/merge-gate
  workflows and recovery modules, App-token publication, caller
  `tooling/governance/` fixtures and tests, and this package directory.
- `VOC-121-DEP-00` through `VOC-121-DEP-06`: see `change.yaml`.
- `VOC-121-EV-00`: T00 evidence — chosen multi-carrier or fail-loud path,
  helper-preservation mechanism, recovery probe, validation commands, exact
  infra SHA, and pin applicability.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required for
engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected CI/CD orchestration and
`tooling/governance/` fixtures, but the path classifier and independent
verifier remain authoritative.
