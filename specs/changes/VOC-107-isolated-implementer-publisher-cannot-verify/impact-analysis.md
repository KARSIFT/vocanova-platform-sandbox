# VOC-107 — Impact Analysis

## Security and privacy

- **Secrets:** No new secrets. Does not grant the implementer Actions credentials
  or move the App token onto the model-controlled runner.
- **Publish surface:** Broadens bundle contents from a thin remediating delta to
  the complete task-branch-only lineage relative to integration. That is required
  for correct isolated verification; it must not weaken exact-head checks,
  integration ancestry, workflow-path deny, or force-with-lease.
- **Signals / evidence:** Allowlisted metadata only (run IDs, SHAs, boolean
  outcomes, scrubbed reason codes). Forbidden: logs, artifacts, secrets,
  OAuth/session/cookie/token material, user identifiers.
- **Residual risk:** An incorrect integration-anchor selection could omit needed
  commits (fail closed at verify) or, if mis-wired to soft-reset, squash prior
  task commits. Mitigated by separating soft-reset tip from bundle base
  (`VOC-107-D01`) and by positive/negative Git fixtures (`VOC-107-D05`).

## Application and operational surface

- **Application code:** No intentional change.
- **Operational effect:** Remediation attempt-2 publications stop failing solely
  because the clean publisher lacks the remediating prerequisite object. Manual
  recovery of valid committed work for that failure class should no longer be
  required.
- **Cross-repo:** Primary behavior change is in `KARSIFT/karsift-ai-infra`
  `implement.yml` (and related fixtures/README). Calling-repo foundation tests
  or docs change only as needed for consumption/consistency.
- **Planner path:** `plan.yml` uses a similar thin-bundle pattern but is out of
  scope unless a separate issue records an identical live failure.

## Data and migrations

- No application schema migration.
- No database mutation.
- Rollback reverts to thin `base_sha..HEAD` bundles; the #891 publisher-reject
  class would return — known undesirable.

## Analytics and accessibility

- No analytics change — evidence-backed non-applicability.
- No product UI change — evidence-backed non-applicability.
- Accessibility — evidence-backed non-applicability.

## Risks, dependencies, and evidence

- `VOC-107-R00`: **Soft-reset accidentally uses integration tip** and squashes
  prior attempt commits on remediation. Mitigation: `VOC-107-D01`; TEST-00 asserts
  distinct tips.
- `VOC-107-R01`: **Workflow deny scans only the incremental delta** and misses
  workflow-path changes earlier in the task lineage. Mitigation: `VOC-107-D02`
  requires deny scan over integration-anchor..HEAD; TEST-03.
- `VOC-107-R02`: **Bundle still too thin after rebase** (anchor recorded before
  rebase, or old tip reused). Mitigation: record integration tip after
  rebase/fresh-after-conflict; TEST-02.
- `VOC-107-R03`: **Attempt-cap or model-rerun creep** for publisher-only
  failures. Mitigation: `VOC-107-D03`; TEST-06.
- `VOC-107-R04`: **Docs drift** claiming thin remediating bundles are always
  publisher-sufficient. Mitigation: AC-05 / TEST-07.
- `VOC-107-DEP-00`: Issue #891 incident (run 32539352323; thin-bundle reject).
- `VOC-107-DEP-01`: Current implement.yml bundle/publish wiring.
- `VOC-107-DEP-02`: Cross-repo infra change ownership pattern.
- `VOC-107-DEP-03`: Soft-reset tip must remain distinct from bundle/publish base.
- `VOC-107-EV-00`: T00 evidence — lineage-fix summary, deterministic Git fixture
  results, doc alignment notes (no secrets/logs/artifacts).
