# VOC-104 — Impact Analysis

## Security and privacy

- **Secrets:** No new secrets. No broadening of implementer token scopes.
- **Verification trust:** Reuse is allowed only for App-signed exact-head PASS
  verdicts already accepted by merge-gate. Human and implementer comments are
  never reusable authority.
- **Skip surface:** Skipping CI and model review is a deliberate, fail-closed
  optimization. Any missing/non-successful check, verdict defect, attestation
  gap, or base/head drift returns to the full path rather than inventing a merge.
- **Signals:** Proof and decision evidence are allowlisted metadata only (run IDs,
  job conclusions, SHAs, boolean reuse decision, package/task IDs). Forbidden:
  logs, artifacts, secrets, OAuth/session/cookie/token material, user identifiers.
- **Proof boundary:** The post-transition verifier is read-only. It reads only
  Actions run/job and PR metadata; never logs or artifacts; and receives no App
  write token, inherited/model/deploy/application secrets, or Actions-write
  permission.
- **Residual risk:** A too-permissive reuse predicate could skip verification on
  unsafe PRs. Mitigation: D02 precondition matrix, fail-closed defaults, and
  deterministic negative tests. A too-strict predicate only reintroduces today's
  redundant cost (fail-visible, non-unsafe).

## Application and operational surface

- **Application code:** No intentional change.
- **Operational effect:** Unchanged-SHA draft→ready transitions avoid duplicate
  CI and model/tool compute while still re-entering merge-gate once the PR is
  non-draft.
- **Latency/compute:** Expected reduction of several minutes and one model review
  per safe ready_for_review event (issue #872 observations on PRs #868 / #869).
- **Cross-repo:** Primary behavior change is in `KARSIFT/karsift-ai-infra`
  pipeline template / reuse helper and related reusable-job conditions; the
  calling repo wires the same decision and the narrow read-only proof action.
- **Protected paths:** The calling repository's
  `.github/workflows/pipeline.yml` is an R3 CI/governance surface. Its mandatory
  `docs/operations/15-ai-native-product-and-engineering-operating-model.md`
  §17.3 update is an R4 protected path and sets the package floor. Shared-infra
  pipeline/reuse workflow files and their policy tests are also explicitly in
  T00 scope.
- **Release:** Unchanged final-roster / develop→main promotion semantics.

## Data and migrations

- No application schema migration.
- No database mutation.
- Rollback reverts the reuse-eligibility wiring; pre-fix behavior (full CI +
  review on every ready_for_review) would return — known redundant but safe.

## Analytics and accessibility

- No analytics change — evidence-backed non-applicability.
- No product UI change — evidence-backed non-applicability.
- Accessibility — evidence-backed non-applicability.

## Risks, dependencies, and evidence

- `VOC-104-R00`: **Unsafe reuse** if preconditions are incomplete.
  Mitigation: D02 checklist; negative tests for drift, bad checks, bad verdicts,
  missing attestation, and non-App comments.
- `VOC-104-R01`: **Merge-gate starvation** if skipped review siblings prevent
  merge-gate from running.
  Mitigation: preserve existing `always()` / skipped-sibling reachability; TEST-02
  asserts merge-gate still selected.
- `VOC-104-R02`: **Draft auto-merge regression.**
  Mitigation: AC-00 / TEST-00; merge-gate draft blocking unchanged.
- `VOC-104-R03`: **Docs drift** claiming every ready_for_review always re-runs
  full CI and model review.
  Mitigation: AC-07 / TEST-11; AGENTS.md doc-consistency rule.
- `VOC-104-R04`: **Scope creep** into deprecated inputs / Node warnings /
  dependency alerts / remediation preflight.
  Mitigation: D09; explicit non-goals in README and tasks.
- `VOC-104-R05`: **Same-head check supersession or ruleset starvation** if the
  current ready run's check names obscure the completed prior run's successes in
  GitHub's latest-name rollup, or omitting the reusable CI caller fails to emit
  the ruleset's exact required `ci / ci` context.
  Mitigation: D02/D03 select and carry a distinct prior run/check identity,
  exclude the current run, always invoke the CI caller with a named reuse marker
  that skips checkout/application validation, and make merge-gate consume that
  validated reuse decision; TEST-02 and TEST-04A cover both directions.
- `VOC-104-DEP-00`: Issue #872 incident (PRs #868 / #869 duplicate ready_for_review
  CI+review).
- `VOC-104-DEP-01`: Existing draft-aware ready_for_review + merge-gate draft block.
- `VOC-104-DEP-02`: Cross-repo infra change ownership pattern.
- `VOC-104-DEP-03`: Resolved during draft refinement — one reusable, read-only
  eligibility workflow/helper is exposed as a caller job; caller conditions
  consume its machine decision.
- `VOC-104-DEP-04`: Resolved during draft refinement — optimize both `agent/` and
  `plan/` PRs, each with its matching trusted publisher check and identity
  binding.
- `VOC-104-EV-00`: T00 evidence — reuse-policy summary, deterministic test output,
  doc alignment notes (no secrets).
- `VOC-104-EV-01`: T01 evidence — scrubbed ready_for_review run metadata proving
  a successful required CI reuse marker, skipped full validation/model review,
  and successful merge-gate re-evaluation on unchanged exact SHA
  (operator-owned live evidence).
