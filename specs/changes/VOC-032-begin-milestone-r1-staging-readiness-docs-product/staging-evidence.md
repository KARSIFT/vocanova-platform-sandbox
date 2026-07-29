# VOC-032 — R1 Staging Readiness: Staging and Rollback Evidence

## Purpose

This document records the evidence required by
`VOC-032-AC-09`/`AC-10`/`AC-12`..`AC-15` (via
`VOC-032-TEST-21`/`TEST-22`/`TEST-24`..`TEST-29`). It is drafted before
adoption/implementation; it is updated as `T00`–`T12` merge to record the
in-repository evidence actually produced, and — unlike every prior
milestone's `staging-evidence.md`, which has been blocked outright by a
non-existent F3 staging environment — this document is where that blockage
is finally supposed to end: `T09`/`T10` are this package's own tasks to
actually stand up and exercise the real environment, not evidence deferred to
some future package.

## Current status

As of adoption (2026-07-28), no task has been implemented. Once `T00`–`T08`,
`T11`, `T13`, and `T12` merge, their in-repository evidence can be recorded
here without further external dependency; `T14`/`T15`'s unit-tested code
paths can too. `T09` (migration/rollback rehearsal) and `T10` (live
AI-evaluation pass) require the founder-provisioned credentials named in
`VOC-032-DEP-00` (SSH access), `VOC-032-DEP-01` (Cloudflare certificate/DNS),
and `VOC-032-DEP-03` (AI-provider staging credentials); `T14`/`T15`'s one
live send/exchange require `VOC-032-DEP-07` (email-provider/Google-OAuth-
client credentials) — these are narrower, concrete, credential-shaped
blockers, not the open-ended "F3 does not exist" blocker every prior
milestone recorded. No R1-gate-complete declaration is made here, and none
can be made until every task is implemented, the founder-provisioned
credentials exist, `T09`/`T10`/`T14`/`T15` actually run, and the founder
completes staging acceptance per DOC-12 §5.

## Planned in-repository evidence (produced by `T00`–`T12`, recorded as each task merges)

| Evidence | Requirement | Status / source |
| --- | --- | --- |
| `EV-00`..`EV-04` | Real DB-backed API server, health endpoint, kill switches, graceful shutdown | **Produced by `T00`** — `apps/api/cmd/api/main.go`, new production-wiring source file(s) in `apps/api/app/api/`. |
| `EV-05` | `.env.example` completeness | **Produced by `T01`** — `apps/api/.env.example`, `apps/web/.env.example`. |
| `EV-06`..`EV-07` | `apps/api` Dockerfile builds and serves healthy | **Produced by `T02`** — `apps/api/Dockerfile`. |
| `EV-08`..`EV-09` | `apps/web` Dockerfile builds and serves | **Produced by `T03`** — `apps/web/Dockerfile`, `apps/web/next.config.ts`. |
| `EV-10`..`EV-11` | Compose stack validates and comes up healthy locally | **Produced by `T04`** — `docker-compose.yml`. |
| `EV-12`..`EV-13` | nginx config valid, routes correctly, real-IP scoped | **Produced by `T05`** — nginx configuration file(s). Live Cloudflare-certificate verification recorded separately below (blocked). |
| `EV-14`..`EV-15` | Atlas applies the full migration set; re-apply is a no-op; down-files not auto-discovered | **T06 tooling delivered** — `apps/api/atlas.hcl`, `apps/api/scripts/migrate.sh`, `apps/api/migrations/atlas.sum`. **Two pre-existing migration-file issues block `EV-14`'s end-to-end pass against the actual migration set; see "T06 follow-ups" section below.** The pre-flight-validation half of `EV-15` (down-files not auto-discovered) is exercised by `apps/api/migrations/atlas_tooling_test.go::TestMigrationsDirectoryHasNoForwardDiscoveredDownFiles`; the end-to-end apply half is blocked until the two follow-ups close. |
| `EV-16`..`EV-17` | Deploy workflow valid, build/push succeeds, fails closed on bad health check | **Produced by `T07`** — `.github/workflows/deploy-staging.yml` (or equivalent). Live SSH-deploy execution recorded separately below (blocked). |
| `EV-18`..`EV-20` | AI-evaluation gate passes at thresholds, fails on violation, wired as required check | **Produced by `T08`** — evaluation-gate command + CI wiring. |
| `EV-21` | Live migration and rollback rehearsal | **Blocked by `VOC-032-DEP-00`/`DEP-01`** — real server/credentials do not yet exist. Procedure documented below; live execution recorded as blocked. |
| `EV-22` | Live-provider AI evaluation pass | **Blocked by `VOC-032-DEP-03`** — staging AI-provider credentials do not yet exist. Procedure documented below; live execution recorded as blocked. |
| `EV-23` | `infra/README.md` accuracy | **Produced by `T11`.** |
| `EV-24`..`EV-25` | Installed-suite pass at final SHA; mock-inventory confirmation | **Produced by `T12` — see `T12 evidence` below.** |
| `EV-26` | DOC-11 §1 amendment accuracy | **Produced by `T13`** — `docs/operations/11-devops-and-ci-cd.md`. |
| `EV-27` | Real email sender: request-construction unit test | **Produced by `T14`.** `apps/api/foundation/email/http.go` (`HTTPSender`, provider-agnostic JSON POST with `Authorization: Bearer <key>`) and `apps/api/foundation/email/http_test.go` (16 tests covering URL/APIKey/From validation, recipient/subject/body construction, Bearer auth, 2xx/non-2xx handling, context cancellation, the never-log-APIKey contract) all pass against a fake `httptest.NewServer` transport; no real provider call is made from CI. The production wiring (`apps/api/app/api/production.go::buildEmailSender`) constructs the real `HTTPSender` when `EMAIL_PROVIDER_API_KEY`/`URL`/`FROM` are set and the `EMAIL_MAGIC_LINK_ENABLED` kill switch is on, and falls back to `email.Fake{}` otherwise - covered by `apps/api/app/api/production_test.go::TestBuildEmailSender_*` (5 tests, all pass). `EMAIL_PROVIDER_URL`/`EMAIL_PROVIDER_API_KEY`/`EMAIL_FROM`/`EMAIL_PROVIDER_TIMEOUT` are documented in `apps/api/.env.example` and read by `LoadProductionConfig`; the timeout defaults to `10s` and `TestLoadProductionConfig_EmailProviderTimeoutDefaults` asserts a positive default. |
| `EV-28` | Real email sender: one live staging delivery | **Blocked by `VOC-032-DEP-07`** — no email-provider account exists yet. |
| `EV-29` | Real Google OAuth provider: token/userinfo unit test | **Produced by `T15`.** |
| `EV-30` | Real Google OAuth provider: one live staging exchange | **Blocked by `VOC-032-DEP-07`** — no Google Cloud OAuth client exists yet. |
| `EV-31` | Exact-SHA independent verification (per PR) | **Performed by Claude Code (different model binding) at each PR's exact final SHA** — this is not the implementer's evidence to record; it is produced by the independent-verification role and reported per-PR. |

## Staging exercise plan (blocked by founder-provisioned credentials, not by a missing environment)

Once `VOC-032-DEP-00`/`DEP-01`/`DEP-03` resolve and `T00`–`T08` are deployed
to the real server at least once, the following exercises must be executed
and their results appended to this document.

### `EV-21` — Migration and rollback rehearsal

1. Confirm the real staging server is reachable via the founder-provisioned
   SSH credentials and that `T07`'s workflow has successfully deployed
   `T00`–`T06` at least once.
2. Apply the current migration set via `T06`'s tooling against the real
   staging Postgres.
3. Create a disposable non-production identity; exercise a handful of
   core-loop writes (save a word, complete a review, submit a sentence
   against the real AI provider).
4. Take a disposable copy of the staging database (`pg_dump` or an
   equivalent snapshot) — never operate on the live staging database
   directly for the remainder of this procedure.
5. Apply each in-scope migration's `.down.sql.example` in reverse order
   against the disposable copy; confirm the resulting schema is consistent
   with the prior release (no orphaned constraint, no dangling foreign key).
6. Re-apply the forward migrations against the disposable copy; confirm no
   data loss beyond what the down-scripts intentionally revert.
7. Record every command, its timestamp, and its outcome below this list once
   run.

### `EV-22` — Live-provider AI evaluation pass

1. Confirm founder-provisioned staging AI-provider credentials exist and an
   explicit cost/rate ceiling has been agreed for this run.
2. Run `T08`'s evaluation harness (or a staging-specific invocation of the
   same dataset) against `aifeedback.NewOpenCodeFeedbackProvider`.
3. Record the resulting scores against every DOC-09 §23 threshold, plus
   total cost and latency, below once run. Any threshold miss is a
   release-blocking finding for R1, not a warning.

### `EV-28` — Real email sender: one live staging delivery

1. Confirm `VOC-032-DEP-07`'s email-provider account/API key exists in
   staging.
2. Trigger one real magic-link send to a founder-controlled test inbox.
3. Confirm the email is actually received; record the send below once run.

### `EV-30` — Real Google OAuth provider: one live staging exchange

1. Confirm `VOC-032-DEP-07`'s Google Cloud OAuth client exists and its
   redirect URI matches `api-staging.vocanova.site`.
2. Perform one real sign-in against Google's actual OAuth flow in staging.
3. Confirm the exchange succeeds and returns a real identity; record the
   attempt below once run.

## Rollback triggers

Per this package's implementation-plan.md §Deployment and rollback /
release-plan.md §Rollback, initiate application rollback on:

- A deploy that leaves `/healthz` unhealthy past the workflow's timeout.
- A migration apply failure against the real staging database.
- An nginx misconfiguration exposing an internal service or misrouting
  traffic between the two staging subdomains.
- Any credential or secret value found in a committed file.
- A rollback-rehearsal result showing a down-file no longer matches its
  forward migration (a finding to fix, not to silently route around).

## Rollback procedure

Redeploy the previous known-good image tag via `T07`'s workflow (or
manually, over the same SSH access, if the workflow itself is what is
broken). Never automate a database rollback — prefer a corrective forward
migration; restore from backup only when data integrity is genuinely at
risk, and only ever against a disposable copy first to confirm the
corrective action's effect, exactly as the rehearsal procedure above
demonstrates. The last-known-good revision is recorded at the future release
decision.

## Relationship to the F3 dependency chain

Every P1–P5 package (VOC-025 through VOC-031) has carried forward the same
open dependency: "the F3 staging environment does not yet exist." This
package's `T09` is the first task in this repository's history to actually
close that dependency — but closing it here does not retroactively declare
any prior milestone's own gate passed. Each of P1–P5 would still need its own
staging exercise, run against this now-real environment and recorded in its
own `staging-evidence.md`, before that milestone's DOC-12 §5 gate can be
declared complete. This document records VOC-032's own R1 evidence only.

## T06 follow-ups: pre-existing migration-file issues block `EV-14` end-to-end

`T06` delivers the Atlas tooling (`apps/api/atlas.hcl`,
`apps/api/scripts/migrate.sh`, `apps/api/migrations/atlas.sum`) and
the in-process pre-flight-validation tests for the wrapper. Two
pre-existing issues in the protected `apps/api/migrations/*.sql`
files were discovered while implementing the apply end-to-end
test; per the implementer prompt's protected-area rule, `T06`
records them as follow-ups rather than silently editing the
migrations in scope. **Both are small, narrowly-scoped fixes**
that can each be a single follow-up PR (or a single combined PR
of ~5 lines diff per file plus one new atlas.sum).

### Follow-up 1: `-- atlas:txmode transaction` is not a valid Atlas directive

Every file in `apps/api/migrations/*.sql` (13 files, VOC-025
through VOC-031) starts with the comment:

    -- atlas:txmode transaction

Atlas v1.x's `atlas:txmode` directive accepts exactly three
values: `none`, `file`, and `all` (and `all` is rejected inside a
per-file directive — it is global-only). The literal value
`transaction` is not in that set; Atlas errors with
`unknown txmode "transaction" found in file directive "<name>.sql"`
and `atlas migrate apply` aborts before running any SQL.
Confirmed locally against Atlas v1.2.0-canary (downloaded
2026-07-29 from `https://release.ariga.io/atlas/atlas-linux-amd64-latest.zip`)
and a fresh disposable Postgres 16.

Fix (minimum diff): either delete the `-- atlas:txmode
transaction` line from every file, or change it to
`-- atlas:txmode file`. The current intent of the comment
("wrap this file in its own transaction") is the default Atlas
behavior with no directive at all, so deleting the line is the
cleanest fix and matches the actual default.

### Follow-up 2: duplicate `streak_states_user_id_key` index in `20260725130002_voc030_p4_gamification_tables.sql`

`apps/api/migrations/20260725130002_voc030_p4_gamification_tables.sql`
defines `streak_states` at line 31-45 with `user_id uuid NOT NULL
UNIQUE` (line 33), and then explicitly creates the same index at
lines 47-48:

    CREATE UNIQUE INDEX streak_states_user_id_key
      ON streak_states (user_id);

Postgres auto-creates a unique index named
`<table>_<column>_key` for any inline `UNIQUE` column
constraint, so the explicit `CREATE UNIQUE INDEX` collides with
the auto-generated one and the apply errors with
`relation "streak_states_user_id_key" already exists (42P07)`
on the very first apply against a fresh database. Confirmed
locally against a fresh disposable Postgres 16 with the
directive from follow-up 1 also fixed (so the only remaining
error is the duplicate index).

Fix (minimum diff): remove either the inline `UNIQUE` from line
33 (keep just `NOT NULL`) or remove the explicit
`CREATE UNIQUE INDEX streak_states_user_id_key` on lines 47-48.
Removing the explicit CREATE UNIQUE INDEX is the smaller diff
(the inline UNIQUE documents the column-level constraint
clearly and Postgres enforces the same invariant either way).

### Combined fix size and what T06 already guarantees

Both follow-ups are mechanical, file-local edits that do not
change any column or constraint semantics. Combined, the diff
is:

  * `apps/api/migrations/*.sql` (13 files) — 1 line each
    deleted (the invalid directive) for follow-up 1.
  * `apps/api/migrations/20260725130002_voc030_p4_gamification_tables.sql`
    — 2 lines deleted (the explicit CREATE UNIQUE INDEX) for
    follow-up 2.
  * `apps/api/migrations/atlas.sum` — regenerated by
    `atlas migrate hash --dir file://migrations` and committed.
  * No changes to any column, type, constraint, index name, or
    row-data semantics; the existing `migration_test.go` string-
    invariant tests still pass unchanged.

`T06`'s own deliverables — the `atlas.hcl` config, the
`migrate.sh` wrapper, the `atlas.sum` integrity file, and the
new `atlas_tooling_test.go` Go tests — are independent of
these two follow-ups and remain in place. The wrapper's
pre-flight-validation behavior (missing DATABASE_URL, missing
Atlas binary, missing atlas.sum, missing migrations dir, non-
file:// URL) is verified by ten passing Go tests; the wrapper
also successfully applies a clean migration set end-to-end
against a fresh disposable Postgres when both follow-ups are
fixed (manually verified locally; an integration test that
exercises the wrapper against a real Postgres is not added
here because the follow-ups are not yet applied — adding such a
test now would re-fail the same way Atlas does against the
current migrations, providing no additional signal over the
upstream-bug evidence already documented above).

### Recommended handling for the founder

Either of the two patterns below is acceptable; the first is
the minimum-blast-radius option and is what `T06`'s implementer
recommends:

  1. **T06 follow-up PR (single, ~5-line diff per file).** Open
     a new change package (e.g. `VOC-033`) with one task that
     applies the two minimum-diff fixes above, regenerates
     `atlas.sum`, and re-runs `T06`'s tests. The new package
     re-uses the existing migration-test invariants in
     `apps/api/migrations/migration_test.go` and T06's new
     `atlas_tooling_test.go` as its regression suite, and adds
     a single end-to-end test that exercises the wrapper
     against a fresh disposable Postgres to confirm the full
     apply path now succeeds. This is a small, low-risk,
     well-scoped PR that does not change any product behavior.

  2. **Narrow exception in T06's scope (founder decision).**
     The implementer's protection-against-scope-expansion rule
     is conservative; if the founder judges the two fixes to
     be obviously within the spirit of T06 ("add Atlas
     tooling that applies the existing migrations"), the
     fixes can be folded into T06 as a single, clearly-
     disclosed commit. This path requires the founder to
     explicitly grant the scope exception (the implementer
     does not grant their own), and the PR description
     should name the two follow-ups and explain why the
     exception is being applied rather than spinning out a
     new package.

Either way, `T06` itself is now in a state where the founder
can review and merge the tooling as-is, and the two follow-ups
can be handled in a tightly-scoped follow-up. `EV-14`'s
end-to-end pass can be recorded as complete once both follow-
ups are applied and the integration test (added in the
follow-up package) passes.

## T14 evidence: real email sender

`T14` adds `apps/api/foundation/email.HTTPSender`, a
provider-agnostic JSON-POST `email.Sender` implementation that
authenticates with `Authorization: Bearer <EMAIL_PROVIDER_API_KEY>`
and posts a `{"from","to","subject","text","html"}` body. This
shape is compatible with Resend, SendGrid v3, and Postmark's
token-auth mode out of the box; providers with a different wire
format can be supported by a future, narrower T14-follow-up
package rather than blocking R1.

The production wiring in `apps/api/app/api/production.go`
(`buildEmailSender`) applies the T14 fallback rules:

1. `EMAIL_MAGIC_LINK_ENABLED=false` → always `email.Fake{}`
   (the auth service rejects magic-link requests outright, so
   the real sender is never reached).
2. `EMAIL_PROVIDER_API_KEY` unset (and kill switch on) → always
   `email.Fake{}`. Staging can run with magic-link delivery off
   at the provider layer rather than crashing at startup.
3. All of `EMAIL_PROVIDER_URL`, `EMAIL_PROVIDER_API_KEY`,
   `EMAIL_FROM` set → real `HTTPSender`. A half-configured
   sender (key set but URL or From missing) is a hard startup
   error, not a silent `Fake{}` fallback; silently falling
   back would hide a real configuration mistake.

The selected sender is logged to stderr (`api: email
sender=HTTPSender url=... from=...` or
`api: email sender=Fake (reason)`) so an operator running the
binary interactively can confirm which path is wired without
reading the env file. The log line never includes the API key.

### Unit-tested evidence (passes locally, in CI)

`apps/api/foundation/email/http_test.go` exercises:

- `NewHTTPSender` validation: missing `URL`/`APIKey`/`From` each
  return a descriptive error (`TestNewHTTPSender_RequiresURL`,
  `_RequiresAPIKey`, `_RequiresFrom`).
- `NewHTTPSender` timeout default: a zero/negative
  `cfg.Timeout` produces a positive `Client.Timeout`
  (`TestNewHTTPSender_DefaultsTimeout`); an explicit
  `cfg.Timeout` is honored (`TestNewHTTPSender_HonorsCustomTimeout`).
- `HTTPSender.Send` request construction: a representative
  magic-link message produces the correct JSON body
  (`{"from","to","subject","text","html"}` matching the source
  message's recipient, subject, and both text/HTML bodies), the
  correct `Authorization: Bearer <key>` header, the correct
  `Content-Type: application/json`, and a stable
  `User-Agent: vocanova-api/1.0`
  (`TestHTTPSender_BuildsCorrectMagicLinkRequest`).
- Multi-recipient path: every recipient is forwarded in the
  JSON array (`TestHTTPSender_MultipleRecipients`).
- Input validation: empty recipients, empty subject, empty
  body, and recipient missing email each return a descriptive
  error before any HTTP call (`TestHTTPSender_ReturnsErrorOnEmptyRecipients`,
  `_ReturnsErrorOnEmptySubject`, `_ReturnsErrorOnEmptyBody`,
  `_RejectsRecipientMissingEmail`).
- Response handling: any 2xx is treated as success
  (`TestHTTPSender_Treats2xxAsSuccess`, parameterised over
  200/201/202/204); any 3xx/4xx/5xx is surfaced as an error
  whose message intentionally does NOT include the response
  body (a provider may return debug details we do not want to
  echo) (`TestHTTPSender_TreatsNon2xxAsError`, parameterised
  over 400/401/403/422/500/502/503).
- Context cancellation: a cancelled context aborts the
  in-flight send (`TestHTTPSender_RespectsContext`).
- The never-log-APIKey contract: HTTPSender has no logging
  paths; the regression test `TestHTTPSender_NeverLogsAPIKey`
  exists as a load-bearing reminder that a future refactor
  adding a logging call must remove the API key from any
  logged payload.

`apps/api/app/api/production_test.go` exercises the wiring
(`TestBuildEmailSender_FallsBackToFakeWhenKillSwitchOff`,
`_FallsBackToFakeWhenNoAPIKey`,
`_BuildsHTTPSenderWhenFullyConfigured`,
`_HardErrorsOnMisconfiguredHTTPSender`,
`_HardErrorsOnMissingFrom`) and the env-var reads
(`TestLoadProductionConfig_ReadsEmailProviderEnvVars`,
`_EmailProviderTimeoutDefaults`).

All `T14` tests pass against a fake `httptest.NewServer`
transport; no real provider call is made from CI. This satisfies
`EV-27`.

### Live-evidence block (recorded as blocked)

`EV-28` — one real magic-link email actually delivered to a
founder-controlled test inbox in staging — remains blocked by
`VOC-032-DEP-07` (no email-provider account/API key exists yet).
The unit-tested code path above is the load-bearing piece; the
one live send is recorded, not asserted, and is exercised once
during `T09`'s rehearsal or later founder staging acceptance.
The procedure is the same as the staging-evidence plan above
(section "`EV-28` — Real email sender: one live staging
delivery").
