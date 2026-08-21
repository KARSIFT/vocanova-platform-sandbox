# VOC-098 — Remove unnecessary Actions permission from operational-failure observer App token: Specification

## Objective and requirement source

Restore the VOC-088-T02 operational-failure observer so it can mint a GitHub App
installation token and open or deduplicate sanitized issues when a watched
workflow ends with `failure`, `cancelled`, or `timed_out`, as recorded in
[GitHub issue #840](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/840).

Today the mint step requests `permission-actions: read` alongside
`permission-issues: write`. The App installation does not grant Actions, so
`actions/create-github-app-token` fails before `open-failure-issue.sh` runs.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

Drafting-time grounding:

| Item | Current state |
|------|---------------|
| Observer workflow | `.github/workflows/operational-failure-monitoring.yml` watches `scheduled-synthetics`, `deploy-staging`, `deploy-production` |
| App token mint | `permission-issues: write` **and** `permission-actions: read` on one mint step |
| Classifier | `infra/scripts/classify-deploy-concurrency-cancel.sh` uses `GH_TOKEN` against `/actions/runs/{id}/jobs` for deploy cancel benign-skip only |
| Issue writer | `infra/scripts/open-failure-issue.sh` App-token only; unlabeled; marker dedupe |
| Precedent | `.github/workflows/error-monitoring.yml` mints App token with `permission-issues: write` only |
| Tests asserting Actions on App token | `scripts/foundation/voc088-failure-to-issue.test.mjs`, `voc094-deploy-concurrency.test.mjs` |

## Scope and non-goals

In scope:

1. Remove Actions permission from the observer **App** installation-token request so
   mint succeeds with least privilege for issue write/dedupe.
2. Preserve App-only issue creation (no `GITHUB_TOKEN` issue create/dedupe),
   sanitized bodies, stable-marker deduplication, concurrency serialization, and
   fail-closed observer behavior when App credentials are missing.
3. Keep VOC-094 benign concurrency-supersession classification for deploy workflows
   without requiring Actions on the App installation (see `VOC-098-D01`).
4. Update deterministic foundation tests so they assert the least-privilege App
   permission set and the classifier/issue-writer token split.
5. Update operator docs only where they would otherwise remain false about which
   token performs Actions metadata reads versus issue creation.
6. Controlled live validation that a watched non-success conclusion invokes the
   observer successfully and creates or deduplicates exactly one sanitized
   App-authored issue; repeating the same marker creates no duplicate.

Non-goals / explicitly excluded:

- Expanding the GitHub App installation's granted permissions to include Actions.
- Using `GITHUB_TOKEN` for issue creation or marker-index mutation.
- Weakening sanitization, copying logs/secrets/sessions/OAuth/user identifiers into
  issues or evidence, or relaxing fail-closed classifier ambiguity handling.
- Changing watched workflow set, marker format, concurrency group, or Sentry/Kuma
  paths.
- Application, migration, signup-policy, or `infra/monitoring/` inventory ID changes.
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R3** (CI/CD observer auth and workflow permissions).
- **Measured path floor at drafting:** **R3** for `.github/workflows/` and
  `infra/scripts/`. Not proposed as R4; no authority-model or amendment docs.
- Protected areas: operational-failure observer auth path; App-token mutation for
  issues; benign-cancel classifier fail-closed semantics.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

The `risk: R3` value in `change.yaml` is a **draft proposal for the reviewing human
at adoption time, never a determination**. The path-based classifier and independent
verifier govern each task PR.

## Decisions

`VOC-098-D00`: The observer App installation token requests **only**
`permission-issues: write`. It must not request `permission-actions` (or any other
permission unused by issue create/list/dedupe). Mint must succeed against the
current App installation that lacks Actions.

`VOC-098-D01` (proposed default; confirm at adoption — open question 1): Bounded
Actions jobs-API reads for `classify-deploy-concurrency-cancel.sh` use the job's
default `GITHUB_TOKEN` (or equivalent non-App job token) with an explicit workflow
or job `permissions.actions: read` floor, **never** for issue create or open-issue
marker scans. `open-failure-issue.sh` continues to receive only the App token.
Fail-closed classifier semantics from VOC-094 remain unchanged (ambiguous/API-error
→ do not skip; proceed toward issue creation).

`VOC-098-D02`: Preserve VOC-088 invariants: App credentials required (fail closed if
missing); unlabeled App-authored issues; allowlisted workflow names and conclusions
only; canonical run URL only; stable HTML marker
`<!-- operational-failure:{workflow}:{conclusion} -->`; concurrency group
`operational-failure-monitoring` with `cancel-in-progress: false`.

`VOC-098-D03`: Deterministic tests must prove the App mint step does **not** request
Actions permission, does request `permission-issues: write`, does not wire
`GITHUB_TOKEN` into issue creation, and still invokes the classifier before
`open-failure-issue.sh` with a non-App token for Actions metadata when deploy
cancel classification runs.

`VOC-098-D04`: Live acceptance uses a **controlled** failed or cancelled watched
workflow (preferred: existing cancelled `scheduled-synthetics` fixture in
`docs/operations/staging-controlled-signup.md`) after the fix is live on the branch
the observer actually executes from (default-branch observer checkout). Evidence is
metadata-only: run IDs/URLs, conclusions, issue number, marker presence, authorship
identity class (App), and duplicate-check result — no logs or secrets.

`VOC-098-D05`: Do not weaken observer fail-closed behavior for missing App
credentials, unobserved workflow names, non-terminal conclusions, or non-canonical
run URLs.

## Open questions for the reviewing human

1. Confirm `VOC-098-D01` dual-token approach (App = issues only; job token = Actions
   jobs API for classifier). Alternative of granting Actions to the App installation
   is **out of scope** per issue #840 (“remove the unnecessary Actions permission”).
2. Confirm proposed **R3**, or raise in writing if observer token-scope changes are
   treated as R4.
3. Confirm T01 live proof may reuse the documented controlled cancel of
   `synthetic.staging.authenticated-core-journey` on the default branch, and that
   if marker `<!-- operational-failure:scheduled-synthetics:cancelled -->` already
   owns an open issue, dedupe-without-duplicate is the success path.
4. Confirm whether T00 must land on `main` (via normal develop→main promotion)
   before T01 live proof, given the observer checks out the default-branch workflow
   definition — package assumes yes.

## Data, migrations, analytics, and accessibility

- No application schema migration.
- No database mutation.
- No product UI change — evidence-backed non-applicability.
- No analytics change — evidence-backed non-applicability.
- Accessibility — evidence-backed non-applicability.
