# VOC-108 — Impact Analysis

## Direct impact

- Shared reusable workflows for adoption, task closure/advancement, merge gates,
  and release promotion.
- The caller event surface that wakes a cheap release re-evaluation when a
  required external check or workflow reaches a terminal conclusion; this may
  add or extend a `check_run`/`workflow_run`-class entrypoint while keeping the
  expensive CI and reviewer jobs undispatched.
- Shared deterministic state-selection/completion helpers and self-CI fixtures.
- Cross-repository PR text policy in the implementer prompt/workflow paths that
  can describe work outside the caller repository.
- Caller foundation contract tests or evidence only where needed to prove the
  shared `@main` behavior is consumed.

## Risk

The proposed risk is R3 because an incorrect decision could block valid work or
promote incomplete integration state. The design reduces that risk with exact
repository/PR/base/head binding, latest-attempt semantics, App-authored immutable
completion markers, per-release serialization, negative fixtures, and
fail-closed ambiguity handling.

## Security and privacy

No secrets or application data are needed. GitHub App credentials remain scoped
to clean publisher/merge jobs and are never exposed to model runners. Issue and
PR evidence remains sanitized metadata. Untrusted titles, bodies, and comments
are parsed only against strict machine-readable contracts.

## Operational impact

The intended effect is less duplicated CI/model work and deterministic recovery
from delayed checks or duplicate events. No container, network, port, database,
deployment, public-host, OAuth, Kuma, or Sentry configuration changes.

## Rollback impact

Revert shared-infra changes through its normal PR path and any caller contract
updates through the governed caller path. Rollback restores the prior race-prone
behavior, so exact regression fixtures must be rerun and the limitation noted.
