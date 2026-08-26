# VOC-125 — Impact Analysis

## Security and privacy

This package repairs the documented operator implement-dispatch contract so
an existing implementation PR can be resumed after automatic remediation
stops. It does not introduce new secret values, OAuth/session material,
production-data access, or user-facing data flows. It does not rotate
`KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY` and does not change the
`karsift-ai-infra-bot` installation.

Security controls that must remain:

- Operator identity is an existing PR number, not free-form SHAs on caller
  `workflow_dispatch`.
- Derived or supplied SHAs are bound to the live PR, branch, repository,
  task, package, authority issue, prior-review evidence when present, and
  current remote ref before any model or mutation step.
- The model-controlled implementer runner never receives the GitHub App
  token and still has no general `actions` permission.
- `publish-source` remains an infrastructure-scoped App token with no
  caller-token fallback.
- Caller `publish` still omits `permission-workflows` and still rejects
  `.github/workflows/**` before push.
- Force-with-lease continues to bind expected heads.
- Nested `karsift-ai-infra/.git` is never staged as a caller gitlink.
- No credential values are printed in logs, tests, or evidence.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect.

## Risks, dependencies, and evidence

- `VOC-125-R00`: **High delivery risk** if operator resume still cannot
  supply verified `expected_head_sha` / `expected_base_sha`. Every
  post-remediation resume of an existing carrier repeats #1003 / job
  `98170418081`. Mitigation: `VOC-125-D01`, `VOC-125-D02`, `VOC-125-AC-00`.
- `VOC-125-R01`: **High identity-spoofing risk** if caller dispatch accepts
  free-form SHAs, or if supplied SHAs are not bound to the live PR/branch/
  issue/task/package/review/remote head. Mitigation: `VOC-125-D01`,
  `VOC-125-D03`, `VOC-125-AC-03`.
- `VOC-125-R02`: **High carrier-rewrite risk** if resume is allowed as
  attempt `1`, attempt `3`, or a replacement PR/branch. Mitigation:
  `VOC-125-D02`, `VOC-125-D04`, `VOC-125-AC-01`, `VOC-125-AC-02`.
- `VOC-125-R03`: **High automatic-retry regression risk** if `remediate.yml`
  stops forwarding event-derived SHAs or starts requiring an operator PR
  paste. Mitigation: `VOC-125-D01`, `VOC-125-AC-04`.
- `VOC-125-R04`: **Medium coverage risk** if tests cover only the happy path
  or only empty-SHA failure, omitting wrong-PR, stale-head, foreign-review,
  closed-task, and attempt-1-with-existing-carrier classes. Mitigation:
  `VOC-125-D06`, `VOC-125-TEST-04`.
- `VOC-125-R05`: **Low release risk** because no application runtime
  deployment change is intended; rollback is workflow/config/test reversion.
  Restored operator resume may continue an already-authorized VOC-122
  carrier through existing gates.
- Protected surfaces: `KARSIFT/karsift-ai-infra` implement and remediate
  workflows, caller pipeline dispatch, exact-head bindings, attempt caps,
  caller `tooling/governance/` fixtures and tests, and this package
  directory.
- `VOC-125-DEP-00` through `VOC-125-DEP-07`: see `change.yaml`.
- `VOC-125-EV-00`: T00 evidence — recovery-identity mechanism, mismatch
  fail-closed proof, validation commands, exact infra SHA, pin
  applicability, and #1003 / #1012 resume handoff.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required for
engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected CI/CD orchestration and
`tooling/governance/` fixtures, but the path classifier and independent
verifier remain authoritative.
