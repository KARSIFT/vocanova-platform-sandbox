# VOC-104 — Skip redundant ready-for-review CI and model review on unchanged exact SHA: Specification

## Objective and requirement source

Stop the caller pipeline from repeating full CI and independent model review on
`ready_for_review` when the PR base SHA and head SHA are unchanged and that exact
pair already has successful required checks plus a trusted App-authored PASS
verdict, as recorded in
[GitHub issue #872](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/872).

Today the caller subscribes to `ready_for_review` so a draft that already passed
CI and review can become merge-eligible without a new commit. Merge-gate correctly
blocks drafts and rechecks draft/open state plus the reviewed base/head before
merge. The redundant cost is that `ready_for_review` still starts a full `ci` and
`review` / `plan-review` path even when nothing about the reviewed revision
changed.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

Drafting-time grounding:

| Item              | Current state                                                                                                                       |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| Trigger           | Caller `.github/workflows/pipeline.yml` includes `ready_for_review` so unchanged-SHA draft completion can re-enter merge evaluation |
| CI / review       | `ci`, `review`, and `plan-review` jobs run for every non-closed PR action, including `ready_for_review`                             |
| Merge gate        | Blocks `isDraft=true`; rechecks live base/head and `--match-head-commit` before merge                                               |
| Verdict authority | Only App-signed exact-head independent verification comments (not human/implementer text)                                           |
| Incident          | PR #868 runs 32492018782 → 32492586621; PR #869 runs 32493543324 → 32494037984 (unchanged heads; full CI+review repeated)           |
| Deferred by       | VOC-102-D08 explicitly left duplicate exact-SHA reviews to a separate follow-up                                                     |

## Scope and non-goals

In scope:

1. Detect `ready_for_review` and evaluate whether prior exact-SHA evidence may be
   reused for that event.
2. Reuse only when the live base/head pair equals the expected pair from the
   triggering run, all required checks for that exact head are successful
   (excluding in-flight merge-gate/remediate self-checks as merge-gate already
   does), and a trusted App-authored PASS or PASS WITH NON-BLOCKING FINDINGS
   verdict is bound to that exact base/head plus the PR's task/package (and
   authority-issue, for agent PRs) identity.
3. In that safe case, skip full CI and model review; still run deterministic
   merge-gate re-evaluation so a newly non-draft PR can merge.
4. Fail closed to the normal full CI + review path when any reuse precondition
   fails (base/head drift, missing/non-successful required checks, missing /
   WAITING / FAIL / PENDING / malformed / untrusted verdict, required
   live-evidence attestation absent, or identity/scope metadata mismatch).
5. Preserve draft never auto-merges.
6. Never accept human or implementer comments as reusable verification authority.
7. Preserve exact-SHA stale-run protections and independent-role separation.
8. Deterministic shared-infra policy tests plus calling-repository
   fixture/foundation coverage.
9. Controlled live proof of the optimized draft→ready path with metadata-only
   evidence (run IDs and job conclusions).
10. Update infra README and calling-repo DOC-15 §17.3 so “fresh pipeline
    evaluation” explicitly distinguishes safe exact-SHA reuse from the full path.

Non-goals / explicitly excluded:

- Deprecated action inputs, Node runtime warnings, dependency alerts, and
  deterministic remediation preflight (separate focused roots per #872).
- Allowing draft auto-merge or weakening merge-gate draft blocking.
- Treating human/implementer comments as PASS authority.
- Application, migration, signup-policy, secrets, database, or
  `infra/monitoring/` inventory ID changes.
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Package classification:** **R4**.
- **Measured path floor:** **R3** for `.github/workflows/` and related governance
  automation, raised to **R4** because T00 must update the exact protected path
  `docs/operations/15-ai-native-product-and-engineering-operating-model.md`.
  The reuse behavior is semantically R3 CI/CD/governance enforcement, but the
  package takes the highest applicable class. Each task PR still uses the highest
  builder, path-classifier, or independent-verifier result.
- Protected areas: merge-gate draft blocking; App-only verdict trust; exact-SHA
  stale-run guards; independent implementer/reviewer separation; live-evidence
  attestation when required.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

The `risk: R4` value in `change.yaml` reflects the mandatory DOC-15 path floor.
The path-based classifier and independent verifier still govern each task PR.

## Decisions

`VOC-104-D00`: Draft pull requests MUST remain non-mergeable even when their
current exact SHA already has green required checks and a trusted App PASS.
`ready_for_review` is the intentional re-entry that allows merge-gate to see
`isDraft=false` without requiring a new commit. This package MUST NOT weaken that
rule.

`VOC-104-D01`: Reuse evaluation applies only to the `pull_request` activity
`ready_for_review`. Other activities (`opened`, `synchronize`, `reopened`) keep
today's full CI and applicable review path. The eligibility decision is computed
once per pipeline run and exposed as a machine-readable outcome
(`reuse-evidence` vs `full-path` vs `fail-closed-to-full-path`) consumed by `ci`,
`review` / `plan-review`, and `merge-gate` job conditions. Resolved placement: a
small reusable, read-only eligibility workflow/helper runs as a caller job,
inspects Actions/check/comment metadata only, and exposes the decision to caller
`if:` conditions. The full CI/reviewer reusable workflows do not each reimplement
the trust predicate. The outcome contract is exact: `reuse-evidence` means all
D02 predicates passed; `full-path` means the event is not eligible or a D02
predicate deterministically failed; `fail-closed-to-full-path` means the helper
could not make a trustworthy decision because metadata, API access, parsing, or
its own execution failed. Callers run the same full CI/review path for either
non-reuse outcome, but retain the distinct value for tests and diagnosis.

`VOC-104-D02`: Reuse is allowed only when all of the following hold for the live
PR at evaluation time:

1. Live `headRefOid` and `baseRefOid` equal the caller-supplied expected
   base/head pair for this run (same exact-SHA stale-run rule merge-gate already
   enforces).
2. Every required check for that exact head is `SUCCESS` or an explicitly
   allowed `SKIPPED` state under the same exclusion set merge-gate already uses
   for its own in-flight `merge-gate/*` and `remediate/*` names. Required CI and
   trusted review publisher checks must be `SUCCESS` in one completed prior
   pipeline run on this exact head. The helper identifies that prior run and its
   job/check identities explicitly, excludes the current `ready_for_review`
   `github.run_id`, and does not rely on an ambiguous latest-by-name PR rollup.
3. A trusted App-authored independent verification comment exists, signed by the
   dedicated publisher identity already used by merge-gate
   (`karsift-ai-infra-bot[bot]`), bound to the exact head SHA, exact base SHA, and
   the PR's package path (plus task ID and authority issue for `agent/` PRs).
4. The classified verdict is `PASS` or `PASS WITH NON-BLOCKING FINDINGS` (not
   `FAIL`, `WAITING`, `PENDING`, or malformed).
5. When live-evidence attestation is required for that PR/task state, the
   deterministic attestation must be present and true; absence fails closed.
6. Identity/scope metadata matches the PR body-bound package/task/authority
   fields used by merge-gate today.

`VOC-104-D03`: When reuse is allowed, the `ready_for_review` pipeline run MUST
skip full CI execution and skip model-invoking review / plan-review work, and MUST
still run deterministic merge-gate re-evaluation so auto-merge can proceed for a newly
non-draft PR. The reusable CI caller MUST emit the ruleset-required `ci / ci`
context through a successful named reuse-marker step while checkout and the full
application-validation step are skipped. Skipped jobs must not leave merge-gate waiting forever on a missing
review dependency — merge-gate already tolerates a skipped review sibling via
`always()`; the reuse path must preserve that reachability. The reuse decision
and validated prior-run identity are passed to merge-gate. On `reuse-evidence`,
merge-gate treats current intentional CI/review skips as replaced only by that
validated prior `SUCCESS` evidence and App-signed exact-SHA verdict. It must not
re-require a current-run publisher `SUCCESS` or let a current skipped check
supersede prior success in a name-only rollup. Without a valid reuse decision,
existing full-path check and verdict requirements remain unchanged.

`VOC-104-D04`: When any `VOC-104-D02` precondition deterministically fails, emit
`full-path` and run normal full CI + applicable review (not an invented merge).
Examples include: base or head changed; required checks missing or
non-successful; verdict missing / WAITING / FAIL / PENDING / malformed /
untrusted; live-evidence attestation required but absent; package / task /
authority identity mismatch. When evaluation itself is uncertain or fails — for
example an API, parsing, metadata-shape, permission, or helper execution error —
emit `fail-closed-to-full-path` and run that same full path. Invalid draft-state
handling already fails closed in merge-gate and must remain so. Do not guess.

`VOC-104-D05`: Human comments, implementer comments, and any non-App review text
MUST NEVER be treated as reusable verification authority. Only the existing
App-signed publisher comment shape used by merge-gate qualifies.

`VOC-104-D06`: Exact-SHA stale-run protections and independent-role separation
remain in force. A newer push still invalidates older runs. Implementer and
reviewer remain different roles/vendors. This package must not create a path where
the implementer can mint or impersonate the reusable PASS. Both source runs must
resolve the complete reuse-critical shared-workflow set to the same authenticated
policy SHA. The proof verifier must recompute the latest eligible prior run. If
GitHub omits a closed run's PR association, branch/head equality is insufficient:
the prior run requires its exact App review binding, and the ready run requires a
unique App-authored pre-merge transition attestation bound to repository, PR,
branch, base/head, ready/prior run IDs, and policy SHA.

`VOC-104-D07`: Deterministic shared-infra policy tests cover at least: (positive)
unchanged base/head + green required checks + trusted App PASS → skip CI/model
review and still evaluate merge-gate; (negative) any D02 failure → full path;
(draft) draft PRs still never enter auto-merge; (authority) human/implementer
comments rejected; (attestation) required live-evidence attestation absent → full
path / fail closed; (regression) `synchronize` / `opened` still run full CI and
review. Calling-repository fixture/foundation coverage asserts the caller
pipeline wires the reuse decision into `ci` / review / merge-gate conditions and
still lists `ready_for_review` in PR types. Adversarial proof fixtures cover
policy-SHA drift, a non-latest supplied prior run, missing/duplicate transition
attestations, and separate PRs that reuse the same branch/head identity.

The positive fixture must include two check suites on the same head: a completed
prior pipeline with successful CI/publisher checks and the current
`ready_for_review` run with a successful reuse-only `ci / ci` context plus the
publisher job intentionally skipped. It proves
prior evidence is selected by run/check identity, the current run is excluded
from eligibility evidence, and merge-gate still passes required-check and verdict
evaluation. The paired negative fixture omits prior success and must take the
full path.

`VOC-104-D08`: Controlled live proof uses a draft PR whose exact base/head already
has green required checks and a trusted App PASS, then marks it ready. Evidence
records allowlisted metadata only: prior run ID, ready_for_review run ID, job
names/conclusions (CI reuse marker successful, full validation/review skipped;
merge-gate success), base/head SHAs, and
boolean reuse decision. Forbidden: logs, artifacts, secrets, tokens, user
identifiers. Preferred dogfood: after T00 is live, a controlled draft→ready
transition (this package's own post-T00 carrier/proof PR or another short-lived
controlled PR). T00 also lands a read-only metadata verifier
(`verify-ready-for-review-reuse`) so T01 can bind proof to `exact_pr_head` without
claiming a pre-ready run as the ready_for_review evidence.

`VOC-104-D09`: Keep root scope focused. Deprecated action inputs, Node runtime
warnings, dependency alerts, and deterministic remediation preflight are out of
scope follow-ups per issue #872.

`VOC-104-D10`: The optimized path applies to both `agent/` task PRs and `plan/`
planning PRs, each
requiring its matching trusted publisher check (`review / publish-review` vs
`plan-review / publish-plan-review`) and identity fields already enforced by
merge-gate.

## Resolved draft-refinement decisions (2026-08-21)

1. Use one reusable, read-only eligibility workflow/helper surfaced as a caller
   job; caller conditions consume its machine decision.
2. Cover both `agent/` implementation PRs and `plan/` package PRs, with their
   distinct trusted publisher checks and identity fields.
3. Classify the package R4 because mandatory DOC-15 §17.3 work raises the R3
   semantic CI/CD effect to the protected path floor; task diffs are reclassified.
4. T01 will dogfood a controlled draft→ready transition after T00 is live, and
   records only run IDs, job conclusions, base/head SHAs, and the reuse boolean.

## Data, migrations, analytics, and accessibility

- Data / migrations: None — evidence-backed non-applicability.
- Analytics: None — evidence-backed non-applicability.
- Accessibility: None — evidence-backed non-applicability (no product UI).

## Security and privacy

- No new secrets. No broadening of implementer token scopes.
- Reuse decision and proof evidence are allowlisted metadata only (workflow/run
  IDs, job conclusions, SHAs, boolean reuse decision, package/task IDs).
  Forbidden: logs, artifacts, secrets, OAuth/session/cookie/token material, user
  identifiers.
- Verdict trust remains App-publisher only; human/implementer comments are never
  reusable authority.
- Skipping CI/model review is allowed only under the fail-closed D02 preconditions;
  any uncertainty returns to the full path.
