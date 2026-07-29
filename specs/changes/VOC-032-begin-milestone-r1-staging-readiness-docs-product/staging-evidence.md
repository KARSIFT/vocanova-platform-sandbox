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
| `EV-23` | `infra/README.md` accuracy | **DIVERGENT — see `T11` follow-up note below.** `infra/README.md` still contains the pre-VOC-005 placeholder text ("This directory is a non-deploying structural boundary. VOC-005 authorizes no Cloudflare, staging, production, release, or autonomous-development infrastructure."); `T11`'s rewrite to the AC-11 description of the docker-compose / nginx / Atlas layout has not been applied. Detected and t.Log-reported by `apps/api/gate_readiness/gate_readiness_test.go::TestT11InfraReadmeIsNotThePlaceholder`. |
| `EV-24`..`EV-25` | Installed-suite pass at final SHA; mock-inventory confirmation | **Produced by `T12` — see `T12 evidence` and `R1 gate-readiness summary` sections below.** |
| `EV-26` | DOC-11 §1 amendment accuracy | **DIVERGENT — see `T13` follow-up note below.** `docs/operations/11-devops-and-ci-cd.md` §1's target-infrastructure table still describes the pre-amendment "Cloudflare Workers via OpenNext + Render Web Service + Render PostgreSQL + vocanova.com" target; `T13`'s amendment to the package's real, built shape (self-hosted Docker Compose + nginx on the founder's server, vocanova.site) has not been applied. Detected and t.Log-reported by `apps/api/gate_readiness/gate_readiness_test.go::TestT13Doc11AmendmentApplied`. |
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

## T15 evidence: real Google OAuth provider

`T15` adds `apps/api/business/auth.GoogleOAuthProvider`, a
real OAuth 2.0 / OpenID Connect implementation that
exchanges an authorization code with Google's
`https://oauth2.googleapis.com/token` endpoint and
fetches the identity from
`https://openidconnect.googleapis.com/v1/userinfo`.
The implementation is provider-shaped after Google's
documented endpoints but takes both URLs from the
config struct so a fake HTTP transport (the unit-test
pattern `T14` established with `email.HTTPSender`) can
intercept every call without reaching a real Google
endpoint from CI. The constructor defaults both URLs
to the real Google endpoints when left empty, so the
production wiring never has to set them.

The production wiring in `apps/api/app/api/production.go`
(`buildOAuthProvider`) applies the T15 fallback rules
(mirroring T14's `buildEmailSender` shape):

1. `GOOGLE_OAUTH_ENABLED=false` → always
   `auth.NewFakeOAuthProvider`. The auth service's
   kill switch short-circuits `OAuthStart` with
   `ErrOAuthDisabled` before the provider is reached,
   so the fake is never exercised in practice; the
   wiring still has to return a non-nil provider so
   `NewService` can be constructed.
2. `GOOGLE_OAUTH_CLIENT_ID` unset (or empty) →
   `auth.NewFakeOAuthProvider`. The real provider
   requires a client ID; falling back keeps staging
   runnable with Google sign-in disabled at the
   provider layer rather than crashing at startup.
3. `GOOGLE_OAUTH_CLIENT_ID` set + secret set →
   real `GoogleOAuthProvider`. A misconfigured
   credential (ID set but secret missing) is a hard
   startup error, not a silent `FakeOAuthProvider`
   fallback: silently falling back would hide a real
   configuration mistake.

The selected provider is logged to stderr (`api: oauth
provider=FakeOAuthProvider (...)` or `api: oauth
provider=GoogleOAuthProvider redirect=...`) so an
operator running the binary interactively can confirm
which path is wired without reading the env file. The
log line never includes the client secret.

### Unit-tested evidence (passes locally, in CI)

`apps/api/business/auth/google_oauth_test.go`
exercises (against a fake `httptest.NewServer`
transport):

- `NewGoogleOAuthProvider` validation: missing
  `ClientID`/`ClientSecret`/`RedirectURI` each return
  a descriptive error
  (`TestNewGoogleOAuthProvider_Requires*`).
- Default URL selection: a config with empty
  `TokenURL`/`UserinfoURL` is filled in with the
  real Google endpoints
  (`TestNewGoogleOAuthProvider_DefaultsTokenAndUserinfoURLs`).
- Default scope set: a config with empty `Scopes` is
  filled in with `auth.DefaultGoogleOAuthScopes`
  (`TestNewGoogleOAuthProvider_DefaultsScopes`).
- Default timeout: a config with zero `Timeout`
  produces a positive `Client.Timeout`
  (`TestNewGoogleOAuthProvider_DefaultsTimeout`).
- `Exchange` happy path: a successful token response
  yields the parsed `Token` and the subsequent
  `Identity` call yields the correct `OAuthIdentity`
  shape (subject, email, email-verified, display name,
  avatar URL); no real Google call is made
  (`TestGoogleOAuthProvider_Exchange*`).
- `Exchange` failure paths: a non-2xx token response
  surfaces a descriptive error; a non-2xx userinfo
  response surfaces a descriptive error; a malformed
  JSON token response surfaces a parse error
  (`TestGoogleOAuthProvider_ExchangeReturnsErrorOn*`).
- `Identity` direct-call path: a successful userinfo
  response yields the correct `OAuthIdentity`; a
  non-2xx response surfaces a descriptive error
  (`TestGoogleOAuthProvider_Identity*`).
- Context cancellation: a cancelled context aborts
  the in-flight call
  (`TestGoogleOAuthProvider_RespectsContext`).
- The never-log-secret contract: the implementation
  has no logging paths; the regression test
  `TestGoogleOAuthProvider_NeverLogsSecret` exists
  as a load-bearing reminder that a future refactor
  adding a logging call must remove the client
  secret from any logged payload.
- Production wiring's `TestBuildOAuthProvider_*`
  tests (in `apps/api/app/api/production_test.go`)
  exercise the four fallback rules above; the
  `_HardErrorsOnMissingSecret` test guards the
  "half-configured" path against silent fallback.

All `T15` tests pass against a fake
`httptest.NewServer` transport; no real call to
Google is made from CI. This satisfies `EV-29`.

### Live-evidence block (recorded as blocked)

`EV-30` — one real sign-in against Google's actual
OAuth flow in staging — remains blocked by
`VOC-032-DEP-07` (no Google Cloud OAuth 2.0 client
ID/secret/redirect URI exists yet). The
unit-tested code path above is the load-bearing
piece; the one live exchange is recorded, not
asserted, and is exercised once during `T09`'s
rehearsal or later founder staging acceptance. The
procedure is the same as the staging-evidence plan
above (section "`EV-30` — Real Google OAuth
provider: one live staging exchange").

## T11 follow-up note: `infra/README.md` is still the placeholder

`T11`'s job was to replace
`infra/README.md`'s pre-VOC-005 placeholder text
("This directory is a non-deploying structural
boundary. VOC-005 authorizes no Cloudflare, staging,
production, release, or autonomous-development
infrastructure.") with an accurate description of
the docker-compose / nginx / Atlas layout `T00`–`T09`
built, plus an explicit note that this reflects
VOC-032's founder-directed shape for the staging
tier, which contradicts DOC-11 §1's still-approved
target infrastructure (`VOC-032-D02`) pending a
founder-approved DOC-11 amendment.

That rewrite has not been applied. The file's
content at the `T12` implementation base SHA is
exactly the pre-`T11` placeholder. `T12` does NOT
silently rewrite the file in `T12`'s scope: doing so
would be (a) a package-scope question (`T11` owns
the rewrite), not an implementation judgment call
`T12` is permitted to make, and (b) would hide a
real divergence the gate-readiness summary below
exists to make visible.

The divergence is detected and reported as a
non-failing `t.Log` line by
`apps/api/gate_readiness/gate_readiness_test.go::TestT11InfraReadmeIsNotThePlaceholder`.
The intended handling is a tightly-scoped
`VOC-032-T11` (or `VOC-033`-T11) follow-up PR that
applies the rewrite; the follow-up is a single,
small, well-scoped PR that does not change any
product behavior. `EV-23` is recorded as
"divergent, not satisfied at the final SHA" — `T12`
does not silently mark it as passing.

## T13 follow-up note: DOC-11 §1 amendment is not applied

`T13`'s job was to amend
`docs/operations/11-devops-and-ci-cd.md` §1's
target-infrastructure table: replace the
"Cloudflare Workers via OpenNext" / "Render Web
Service" / "Render PostgreSQL" rows and the
`vocanova.com` domain set with this package's
real, built shape (self-hosted Docker Compose +
nginx on the founder's own server, `vocanova.site`,
Cloudflare for DNS/TLS/WAF/CDN only — not compute),
and add an inline amendment note matching
DOC-15 §17's amendment-note style.

That amendment has not been applied. `DOC-11` §1
still describes the pre-amendment
Render / Cloudflare-Workers / `vocanova.com` target
at the `T12` implementation base SHA. `T12` does
NOT silently amend an approved document in `T12`'s
scope: doing so would be (a) a package-scope
question (`T13` owns the amendment), and (b) would
touch a protected area under
`docs/governance/protected-areas.md`'s
"Repository governance" row (an approved document's
amendment is R3-protected, and the
package-internal implementer does not have the
authority to amend an approved document on their
own — that authority sits with the founder, per
DOC-12 §11's change-control rule).

The divergence is detected and reported as a
non-failing `t.Log` line by
`apps/api/gate_readiness/gate_readiness_test.go::TestT13Doc11AmendmentApplied`.
The intended handling is a tightly-scoped
`VOC-032-T13` (or `VOC-033`-T13) follow-up PR that
applies the amendment, in the same PR review track
as `T11`'s rewrite or as a separate small PR. `EV-26`
is recorded as "divergent, not satisfied at the
final SHA" — `T12` does not silently mark it as
passing.

## T12 evidence: installed-suite pass at the final SHA

`T12` added `apps/api/gate_readiness/`, a new
cross-cutting Go subpackage whose `_test.go` is the
load-bearing machine-readable half of the R1 gate.
The package has four tests:

- **`TestInRepoEvidencePresent`** — the strict
  half. Walks every in-repo `EV-*` claim in the
  table above and asserts each file exists at its
  claimed path. A missing file is a hard test
  failure (not a logged warning), because a
  missing file is either an accidental removal
  (a real bug) or a prior task that did not
  actually merge its deliverable; both are
  release blockers for the R1 gate. At the
  implementation base SHA, every in-repo
  `EV-*` path resolves to a real file (the
  `T11` and `T13` divergent paths are included
  in the walk — the file exists, even if its
  content does not yet match the AC). This
  test exercises `EV-24` (the in-repo-evidence-
  present half).
- **`TestT11InfraReadmeIsNotThePlaceholder`** —
  the second half of the `T11` evidence check.
  Logs (not fails on) the pre-`T11` placeholder
  text if it is still present, so the divergence
  cannot silently disappear between `T12`'s PR
  review and a future PR that does own `T11`.
- **`TestT13Doc11AmendmentApplied`** — the
  second half of the `T13` evidence check.
  Logs (not fails on) the pre-amendment
  Render / Cloudflare-Workers / `vocanova.com`
  table contents if still present, so the
  amendment-missing state cannot silently
  disappear. The log line names the specific
  pre-amendment signal phrases still present,
  giving an independent reviewer immediate
  visibility into which DOC-11 rows have not
  been touched.
- **`TestFullInstalledSuiteRuns`** — the broadest
  half. Asserts that `apps/api/go.mod` exists at
  the resolved `apps/api` root, so the
  `apps/api/...` test tree is reachable from
  `go test ./...`; the test's own PASS/FAIL
  result then encodes "the `go test ./...` run
  this test is part of passed at the final SHA"
  (any failure in any other package's test in
  this CI run fails this test by transitivity
  — the gate cannot be green if the suite is
  red). This test exercises `EV-25` (the
  installed-suite-passes half).

The four tests pass at the implementation base
SHA. The full `apps/api` test suite
(`go test ./...` from `apps/api/`) also passes;
the `go vet ./...` and `gofmt -l .` checks pass
with no output. The `apps/api/cmd/api` binary
builds cleanly and starts (it correctly errors
out with `api: load config: DATABASE_URL is
required` when run without env, confirming
`T00`'s kill-switch-respecting startup path
is wired).

`T12` is not the implementer that produces
`EV-23` or `EV-26`; those are `T11` and `T13`'s
deliverables respectively. The `T12` evidence is
`EV-24` (the in-repo-evidence-present half) and
`EV-25` (the installed-suite-passes half), both
satisfied at the final SHA. The `mock-inventory.md`
final state (this package introduces no product
mock) is recorded separately in that file's
"Final `T12` re-confirmation" section, and is
itself verified by the suite-pass half of `EV-25`.

## R1 gate-readiness summary (T12 deliverable)

Per `DOC-12` §5, the R1 gate is: "stable in
staging, no unresolved critical/high blocker, all
required tests pass, migration + rollback
rehearsed, AI evaluation thresholds pass, founder
completes staging acceptance, scope is frozen."
`T12`'s job is to summarize which of those gate
items are satisfied by this package's evidence
and which remain founder-owned and cannot be
satisfied by any package. The table below is that
summary, exactly as `AC-12` requires.

| R1 gate item (DOC-12 §5) | Status at the final SHA | How it is satisfied / what remains |
| --- | --- | --- |
| **Stable in staging** | **Recorded, not satisfied.** No real staging environment exists yet; `VOC-032-DEP-00` (SSH credentials) and `VOC-032-DEP-01` (Cloudflare certificate / DNS) are still open. The four-service docker-compose stack (`T04`–`T05`) has never been brought up against the real provisioned server. | Founder-owned: requires `DEP-00` and `DEP-01` resolution, then `T07`'s first end-to-end deploy run, then a non-trivial observation period of `/healthz`, `wget` probes, and the core-loop's non-AI paths. Cannot be demonstrated by a code package; requires the real environment. |
| **No unresolved critical/high blocker** | **Partial.** No `P0`/`P1` defect in this package's own work, but the `T06` follow-ups (invalid `-- atlas:txmode transaction` directive; duplicate `streak_states_user_id_key` index) are recorded as known blockers to `EV-14`'s end-to-end pass. The `T11` and `T13` follow-ups are also recorded as blockers to `AC-11` and `AC-13` respectively. | The `T06` follow-ups are small, file-local edits and are recommended as a tightly-scoped `VOC-033`-T06 PR (see the "T06 follow-ups" section above for the minimum diff). The `T11` and `T13` follow-ups are similarly small. None of these is a critical/high blocker for the `T00`–`T09` build/CI-wiring path, but they are blockers for declaring specific ACs fully satisfied. |
| **All required tests pass** | **Satisfied (in-repo).** `go test ./...` from `apps/api/`, `go vet ./...`, `gofmt -l .`, `go build ./...`, and `bash scripts/governance/validate-governance.sh` all pass at the final SHA. The `apps/web` and `infra` trees are exercised by the `karsift-ai-infra` `ci.yml` reusable workflow, not by `T12`'s in-process check, and their evidence is per-PR. | In-repo check is `T12`'s. The `karsift-ai-infra ci.yml` results are produced by CI, not by `T12`. |
| **Migration + rollback rehearsed** | **Procedure documented; live execution recorded as blocked.** `T09`'s procedure (apply → exercise → snapshot → roll back on the disposable copy → verify → re-apply forward → record) is fully specified in this document's "Staging exercise plan" section. The live execution itself is blocked on `VOC-032-DEP-00` / `DEP-01` (real server, real DNS). | Founder-owned: requires the real staging server to exist and the deploy to land at least once before the rehearsal can run. The `T06` follow-ups (above) also block the apply half from succeeding against the current migration set; both must close before `T09`'s forward-apply step is non-blocking. |
| **AI evaluation thresholds pass** | **Mock-provider half satisfied; live-provider half recorded as blocked.** `T08`'s `RunGoldenGate` against `aifeedback.NewMockProvider()` passes every DOC-09 §23 threshold that the current `EvaluationResult` shape can measure (the `StructuredOutputValidAfterOneRepair` and `MeaningPreservation` thresholds are correctly reported as "not tracked" because the data shape cannot measure them yet, per `GoldenThresholdSpec.NotTracked` — this is the documented `T08` boundary, not a regression). `T10`'s live-provider pass is blocked on `VOC-032-DEP-03`. | Mock half: `T08`'s `TestGoldenSetThresholdsAgainstMockProvider` + `TestGoldenGateEnforcesViolatedThreshold` + `TestGoldenGatePassesOnCleanFixture` all pass. Live half: founder-owned, requires `DEP-03` resolution. |
| **Founder completes staging acceptance** | **Not started — founder-owned, out of scope of any package.** | This is a single human decision by the founder, recorded as a comment on the package's release issue (per the standard `karsift-ai-infra release.yml` flow). It cannot be satisfied by any code change. |
| **Scope is frozen** | **Satisfied (in this package).** No new product scope introduced; only the `T00`–`T15` and `T12` deliverables. | Inherently a property of the package's own change-control discipline; `T12`'s role is to confirm no PR in this package introduced unapproved scope. |

### Items this gate-readiness summary explicitly does NOT mark as passing

The following items are not "passing" at the final SHA and are not
the founder's call to wave through. Each is a recorded, factual
limitation:

- **No live staging environment has been brought up.** `T07`'s
  deploy workflow has not been triggered end-to-end against a
  real server. `EV-16`/`EV-17` are satisfied in the "YAML
  valid, build/push succeeds" half only; the "fails closed
  on bad health check" half is unit-tested but not
  end-to-end-verified against a real server.
- **No live Cloudflare TLS verification.** `T05`'s `EV-12`
  /`EV-13` are satisfied in the "nginx -t validates, Host
  header routes correctly" half only; the live
  Cloudflare-issued-certificate verification is blocked on
  `DEP-01`.
- **No live migration-and-rollback rehearsal.** `EV-21` is
  blocked. Procedure documented; execution pending `DEP-00` /
  `DEP-01` resolution and the `T06` follow-ups closing.
- **No live AI-evaluation pass.** `EV-22` is blocked on
  `DEP-03`.
- **No live email send.** `EV-28` is blocked on `DEP-07`.
- **No live Google OAuth exchange.** `EV-30` is blocked on
  `DEP-07`.
- **`T11`'s `infra/README.md` rewrite is missing.** `EV-23`
  is divergent; `AC-11` is not satisfied. T12 logs the
  divergence; the rewrite is a `T11` follow-up.
- **`T13`'s DOC-11 §1 amendment is missing.** `EV-26` is
  divergent; `AC-13` is not satisfied. T12 logs the
  divergence; the amendment is a `T13` follow-up.
- **The `T06` follow-ups are open.** The apply command
  cannot succeed against the current migration set until
  the `-- atlas:txmode transaction` directive and the
  duplicate `streak_states_user_id_key` index are fixed.
  `T12` records this; the fixes are a `T06` follow-up
  (recommended as a single `VOC-033`-T06 PR — see the "T06
  follow-ups" section above).
- **The `T08` "not tracked" thresholds
  (`StructuredOutputValidAfterOneRepair`,
  `MeaningPreservation`)** are reported as not-measurable
  from the current `EvaluationResult` shape. This is the
  documented `T08` boundary, not a regression; closing
  the gap requires wiring repair-attempt tracking and
  corrected-sentence text comparison into the eval
  pipeline, which is a future scope expansion, not a
  `T12` fix.

### Founder-owned items no package can satisfy

The following gate items are by definition not satisfiable
by any code change and remain founder-owned at the
final SHA:

- **Staging acceptance itself.** A single human
  decision by the founder, recorded per the standard
  release.yml flow. `T12` produces this summary so
  the founder has a one-document reference for
  which R1 gate items are package-satisfied and
  which are blocked-on-founder action; the actual
  "approved" decision is not `T12`'s to make.
- **Resolution of `VOC-032-DEP-00` / `DEP-01` /
  `DEP-03` / `DEP-07`.** These are credential and
  DNS provisioning actions the planner/implementer
  cannot perform; the founder is the only actor
  with the relevant accounts.
- **The `VOC-032-D03` (staging subdomain naming)
  and `D04` (F3/R1 scope-folding) open decisions
  recorded at adoption.** `D03` was accepted as-is
  in the adoption delegation; `D04` was resolved
  as part of adoption. Neither is a `T12` item;
  both are recorded here so the R1 gate-readiness
  summary is the single document a future
  independent reviewer can cite to confirm the
  package's full disposition.

### Cross-references

- `mock-inventory.md` — this package's no-product-mock
  disposition, finalized at `T12`.
- `apps/api/gate_readiness/gate_readiness_test.go` —
  the machine-readable half of the gate; runs as part
  of `go test ./...` from `apps/api/`.
- `change.yaml` — the package's own adoption-time
  status, dependencies, and per-DEP resolution text.
- `docs/product/12-mvp-implementation-plan.md` §5 —
  the canonical R1 gate text this summary maps to.
- `docs/operations/11-devops-and-ci-cd.md` §1 — the
  approved document `T13`'s amendment must align
  with (the amendment itself is a `T13` follow-up,
  not a `T12` item).
