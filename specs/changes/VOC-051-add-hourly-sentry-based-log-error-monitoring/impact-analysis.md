# VOC-051 — Impact Analysis

## Security and privacy

- **New secret**: a Sentry API auth token stored as a GitHub Actions secret
  (proposed name: `SENTRY_API_AUTH_TOKEN`, final naming decided at
  implementation time consistent with this repo's existing
  `PRODUCTION_SENTRY_DSN`/`STAGING_SSH_HOST`-style secret-naming convention).
  It must be read-only and least-privilege-scoped — see `specification.md`'s
  open question 2 for why the exact Sentry scope combination is not pinned
  down by this drafting pass. This token is a CI/CD secret only; it must never
  be injected into `apps/api` or `apps/web`'s own runtime environment (it has
  no application-layer use — only the monitoring workflow needs it).
- **New GitHub API write path**: the scheduled workflow calls the GitHub API
  to search and create issues. Its identity (default `GITHUB_TOKEN` vs. a
  GitHub App token) is an open question (`specification.md`'s open question 1)
  with real authorization-boundary consequences — a GitHub App token can be
  scoped narrower (e.g. `issues: write` only) than the default `GITHUB_TOKEN`,
  which inherits the workflow's full declared `permissions:` block. Whichever
  is chosen, the workflow's `permissions:` block must be scoped to the
  minimum needed (`issues: write`, `contents: read` for checkout if any) —
  the same minimum-permissions discipline `.github/workflows/lighthouse.yml`
  already demonstrates (`permissions: contents: read`).
- **No SSH credential added.** Explicit non-goal (see `specification.md`'s
  scope item 5). This is the central security property the issue itself
  asked for: an unattended hourly job must not hold a credential that grants
  raw host access to production. Confirmed as its own acceptance criterion
  (`VOC-051-AC-05`) and test (`VOC-051-TEST-06`), not left as an implicit
  assumption.
- **New third-party data flow (browser → Sentry)**: `apps/web`'s Sentry SDK
  will, once wired, send browser-side error events (including stack traces,
  the URL the error occurred on, and — if the implementer sets Sentry user
  context, which this package does not require or forbid but flags here —
  potentially a user identifier) to Sentry's servers. This is a new
  data-processor relationship for browser-originated data specifically (the
  backend, `apps/api`, already sends server-side error data to the same
  Sentry organization, so the vendor relationship itself is not new — only
  the browser-data flow is). `VOC-051-T01`'s implementer must confirm whether
  any existing privacy policy or DPA covers this addition, or flag it as an
  open item for the reviewing human if it does not — this package does not
  assume it is already covered.
- **Least-privilege for the monitoring workflow overall**: read-only against
  Sentry, `issues: write`-scoped (at most) against GitHub, no SSH, no
  database credential, no deploy credential.

## Data and migrations

No database schema change. The scheduled workflow's own "what have I already
scanned" state (`specification.md`'s open question 4) is CI-infrastructure
state (e.g. a workflow-run artifact, GitHub Actions cache, or a small
committed marker file under this package's own implementation, not the
application database) — it has no migration, integrity, or application-level
rollback concern of its own. If the implementer chooses a committed marker
file, its own update-and-commit step must not touch any file outside its own
narrow purpose (comparable in spirit to `karsift-ai-infra#15`'s "a snapshot
commit under the workflow's own accounting is not new unreviewed scope" logic
that this prompt's own instructions already call out for a different case).

## Analytics and accessibility

- **Analytics**: Not applicable. This package adds error-monitoring
  instrumentation only; `VOC-051-T01`'s implementer must not configure the
  Sentry SDK to also capture performance/session-replay/analytics data beyond
  what's needed for error tracking unless a separate, explicitly approved
  package authorizes that broader scope.
- **Accessibility**: Not applicable — no user-facing UI changes. The one
  concrete accessibility-adjacent risk is Sentry's own client-side error
  overlay (a debug UI `@sentry/nextjs` can render) leaking into production
  builds; `VOC-051-TEST-01` checks this explicitly is disabled in the
  production/staging build configuration.

## Risks, dependencies, and evidence

- `VOC-051-R00`: The exact Sentry API auth token scope available may not map
  cleanly onto "read-only, least-privilege, both environments" — Sentry's own
  scope granularity may force a choice between "broader than strictly needed"
  and "insufficient for the actual API calls." Mitigated by requiring the
  implementer to record and justify the exact chosen scope
  (`specification.md`'s open question 2) rather than silently picking the
  broadest available scope for convenience.
- `VOC-051-R01`: A scheduled workflow that autonomously opens GitHub issues
  every hour, if the duplicate-check guard has a bug (e.g. a title-normalization
  mismatch that fails to recognize a re-worded but identical Sentry issue),
  could spam the issue tracker and, per `AGENTS.md`, each such issue
  auto-triggers `plan-from-issue` — meaning a duplicate-check bug here has a
  real downstream cost (spurious planning runs), not just tracker noise.
  Mitigated by `VOC-051-AC-04`/`VOC-051-TEST-05`'s explicit repeated-run
  non-duplication requirement and by embedding a stable Sentry issue-ID marker
  in the created issue's body (not relying on title-text matching alone) as
  the primary duplicate-check key, per `specification.md`'s scope item 4.
- `VOC-051-R02`: A Sentry outage or API-shape change could make the workflow
  fail silently (e.g. an empty/error response treated as "no new issues"
  rather than surfaced as a monitoring-workflow failure). Mitigated by
  requiring the workflow to fail its own GitHub Actions run (non-zero exit)
  on a genuine Sentry API error, distinct from a legitimate "zero new issues
  found" result, so a broken monitoring workflow is itself visible (via
  normal GitHub Actions run-failure visibility) rather than silently going
  dark — the same "never treat a missing integration ... as a pass" principle
  `CLAUDE.md` already states.
- `VOC-051-DEP-00`, `VOC-051-DEP-01`, `VOC-051-DEP-02`: see `change.yaml`.
- `VOC-051-EV-00`: `apps/web`'s Sentry wiring evidence (SDK present, DSN
  read from environment, disabled-when-unset behavior confirmed, per-environment
  DSN distinctness confirmed) — produced by `VOC-051-T01`.
- `VOC-051-EV-01`: the scheduled workflow's own run evidence (a real or
  simulated run showing correct issue creation and duplicate-suppression) —
  produced by `VOC-051-T02`.
- `VOC-051-EV-02`: the DOC-11 documentation-consistency update — produced by
  `VOC-051-T04`.
- `VOC-051-EV-03`: deterministic validation command output — produced across
  `VOC-051-T01`, `VOC-051-T02`, `VOC-051-T04`.
