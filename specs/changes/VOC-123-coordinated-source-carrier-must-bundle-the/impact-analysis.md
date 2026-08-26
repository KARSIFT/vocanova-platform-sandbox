# VOC-123 — Impact Analysis

## Security and privacy

This package repairs governed coordinated source-carrier Git-bundle creation.
It does not introduce new secret values, OAuth/session material,
production-data access, or user-facing data flows.

Security controls that must remain:

- The model-controlled implementer runner never receives the GitHub App
  token. It uploads a credential-free Git bundle only.
- `publish-source` mints an infrastructure-scoped App token
  (`repositories: karsift-ai-infra`) with no caller-token fallback.
- Publisher `bundle verify` then fetches only
  `"$PUBLISH_HEAD_SHA:refs/heads/$PUBLISH_BRANCH"`.
- Force-with-lease continues to bind `EXPECTED_SOURCE_HEAD_SHA`.
- Nested `karsift-ai-infra/.git` is never staged as a caller gitlink.
- Temporary named refs are isolated, are not the publish branch, are not the
  remediation `source-carrier` branch, and are deleted after use.
- Unrelated refs, tags, remotes, and secrets are never advertised or
  published.
- Caller `.github/workflows/**` changes remain refused on the unreviewed
  automated caller publication path. Infrastructure workflow files belong to
  the infrastructure repository carrier and are independently reviewed there.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect.

## Risks, dependencies, and evidence

- `VOC-123-R00`: **High delivery risk** if source-bundle creation continues
  to use a raw-SHA positive tip. Every coordinated nested infrastructure
  change fails with empty-bundle after the nested commit, repeating #1003 /
  job `98017696468`, with no artifact and no carrier PR. Mitigation: named
  positive revision bound to the exact committed SHA (`VOC-123-D01`,
  `VOC-123-AC-00`).
- `VOC-123-R01`: **High publication risk** if the advertised bundle head is
  not the recorded `SOURCE_HEAD_SHA`, or if unrelated refs become
  fetchable. Mitigation: exact `list-heads` verification, isolated temp-ref,
  publisher SHA fetch only (`VOC-123-D01`, `VOC-123-D03`, `VOC-123-AC-02`).
- `VOC-123-R02`: **High isolation risk** if a temp-ref name collides with
  the publish branch or remediation `source-carrier` checkout, or if the
  named-ref step stages a gitlink. Mitigation: `VOC-123-D01` /
  `VOC-123-DEP-06` and existing gitlink tests.
- `VOC-123-R03`: **Medium coverage risk** if tests keep bundling `..HEAD` or
  `..$branch` and therefore never reproduce the production raw-SHA defect.
  Mitigation: `VOC-123-D04`, `VOC-123-TEST-00`, `VOC-123-TEST-01`.
- `VOC-123-R04`: **Medium documentation risk** if README/workflow comments
  continue to describe a raw-SHA exclusion range as current-state.
  Mitigation: same-task current-state updates (`VOC-123-D05`).
- `VOC-123-R05`: **High bootstrap risk** because the already-resolved
  `implement.yml@main` cannot consume a nested edit to itself and will repeat
  the raw-SHA failure before artifact upload. Mitigation: the bounded,
  independently reviewed, one-time supervised infra bootstrap in
  `VOC-123-D08`; normal source publication becomes mandatory after its exact
  merge and the exception then expires.
- `VOC-123-R06`: **Low release risk** because no application runtime
  deployment change is intended; rollback is workflow/config/test reversion.
  Restored source-carrier publication may open an already-authorized infra
  PR through existing gates.
- Protected surfaces: `KARSIFT/karsift-ai-infra` implement/plan workflows,
  App-token versus job-token split, caller `tooling/governance/` fixtures
  and tests, and this package directory.
- `VOC-123-DEP-00` through `VOC-123-DEP-07`: see `change.yaml`.
- `VOC-123-EV-00`: T00 evidence — named-ref mechanism, advertised-head
  proof, caller/planner inspection result, validation commands, exact infra
  SHA, pin applicability, and #1003 re-dispatch note.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required for
engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected CI/CD orchestration and
`tooling/governance/` fixtures, but the path classifier and independent
verifier remain authoritative.
