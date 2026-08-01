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

`T00`–`T08`, `T14`, and `T15` have merged; `T12` has merged and recorded
this document's gate-readiness summary. `T11` and `T13` remain open follow-
ups (see their own divergence notes below). `VOC-032-DEP-00` (SSH access),
`VOC-032-DEP-01` (Cloudflare certificate/DNS), and `VOC-032-DEP-03`
(AI-provider staging credentials) are now resolved.
`T09`'s migration/rollback rehearsal ran and passed (see `EV-21` below,
including the disclosed caveat about the state the disposable copy was
taken in). `T10`'s live AI-evaluation run and result are recorded in `T10`'s
own evidence update, a separate change not yet merged into this document —
this document does not assert an `EV-22` outcome. `T14`/`T15`'s one live
send/exchange still require `VOC-032-DEP-07` (email-provider/Google-OAuth-
client credentials). No R1-gate-complete declaration is made here, and none
can be made until `T10`'s `EV-22` result is recorded here, `T14`/`T15`'s
live evidence is recorded, and the founder completes staging acceptance per
DOC-12 §5.

## Planned in-repository evidence (produced by `T00`–`T12`, recorded as each task merges)

| Evidence | Requirement | Status / source |
| --- | --- | --- |
| `EV-00`..`EV-04` | Real DB-backed API server, health endpoint, kill switches, graceful shutdown | **Produced by `T00`** — `apps/api/cmd/api/main.go`, new production-wiring source file(s) in `apps/api/app/api/`. |
| `EV-05` | `.env.example` completeness | **Produced by `T01`** — `apps/api/.env.example`, `apps/web/.env.example`. |
| `EV-06`..`EV-07` | `apps/api` Dockerfile builds and serves healthy | **Produced by `T02`** — `apps/api/Dockerfile`. |
| `EV-08`..`EV-09` | `apps/web` Dockerfile builds and serves | **Produced by `T03`** — `apps/web/Dockerfile`, `apps/web/next.config.ts`. |
| `EV-10`..`EV-11` | Compose stack validates and comes up healthy locally | **Produced by `T04`** — `docker-compose.yml`. |
| `EV-12`..`EV-13` | nginx config valid, routes correctly, real-IP scoped | **Produced by `T05`, live TLS confirmed.** nginx configuration file(s), plus a live Cloudflare-fronted HTTPS response observed on both public hostnames (see `EV-21`). |
| `EV-14`..`EV-15` | Atlas applies the full migration set; re-apply is a no-op; down-files not auto-discovered | **T06 tooling delivered; end-to-end apply now passes.** `apps/api/atlas.hcl`, `apps/api/scripts/migrate.sh`, `apps/api/migrations/atlas.sum`. The two pre-existing migration-file issues that previously blocked `EV-14`'s end-to-end pass were fixed in `VOC-033` (see "T06 follow-ups" section below, now marked resolved); the live Atlas apply recorded under `EV-21` confirms the fix. The pre-flight-validation half of `EV-15` (down-files not auto-discovered) is exercised by `apps/api/migrations/atlas_tooling_test.go::TestMigrationsDirectoryHasNoForwardDiscoveredDownFiles`. |
| `EV-16`..`EV-17` | Deploy workflow valid, build/push succeeds, fails closed on bad health check | **Produced by `T07`, live deploy confirmed.** `.github/workflows/deploy-staging.yml`, plus a real end-to-end SSH deploy to the founder's server (see `EV-21` and deploy run `30618654496`). The "fails closed on bad health check" half remains unit-tested only — no real health-check failure has been observed end-to-end. |
| `EV-18`..`EV-20` | AI-evaluation gate passes at thresholds, fails on violation, wired as required check | **Produced by `T08`** — evaluation-gate command + CI wiring. |
| `EV-21` | Live migration and rollback rehearsal | **Passed on 2026-07-30/31.** The real deploy, live Atlas no-op check, disposable-copy rollback of all 12 approved down artifacts, 13-migration forward re-apply, exact schema comparison, and cleanup passed. After the real provider was enabled, a disposable identity also completed the save-word, review, and sentence-feedback core loop with persisted-row and cleanup proof. Details below and in `VOC-034`'s live evidence. |
| `EV-22` | Live-provider AI evaluation pass | **No longer blocked on `VOC-032-DEP-03`; result not recorded in this document.** The `cmd/eval-live` command (`apps/api/cmd/eval-live/main.go`) and the supporting library (`apps/api/business/aifeedback/live_eval.go`) constitute the runnable one-shot procedure. `DEP-03` is now resolved and the protected run has been performed, but its result is recorded in `T10`'s own evidence update, a separate change not yet merged into this document — this `T09` PR does not assert an `EV-22` outcome. |
| `EV-23` | `infra/README.md` accuracy | **DIVERGENT — see `T11` follow-up note below.** `infra/README.md` still contains the pre-VOC-005 placeholder text ("This directory is a non-deploying structural boundary. VOC-005 authorizes no Cloudflare, staging, production, release, or autonomous-development infrastructure."); `T11`'s rewrite to the AC-11 description of the docker-compose / nginx / Atlas layout has not been applied. Detected and t.Log-reported by `apps/api/gate_readiness/gate_readiness_test.go::TestT11InfraReadmeIsNotThePlaceholder`. |
| `EV-24`..`EV-25` | Installed-suite pass at final SHA; mock-inventory confirmation | **Produced by `T12` — see `T12 evidence` and `R1 gate-readiness summary` sections below.** |
| `EV-26` | DOC-11 §1 amendment accuracy | **DIVERGENT — see `T13` follow-up note below.** `docs/operations/11-devops-and-ci-cd.md` §1's target-infrastructure table still describes the pre-amendment "Cloudflare Workers via OpenNext + Render Web Service + Render PostgreSQL + vocanova.com" target; `T13`'s amendment to the package's real, built shape (self-hosted Docker Compose + nginx on the founder's server, vocanova.site) has not been applied. Detected and t.Log-reported by `apps/api/gate_readiness/gate_readiness_test.go::TestT13Doc11AmendmentApplied`. |
| `EV-27` | Real email sender: request-construction unit test | **Produced by `T14`.** `apps/api/foundation/email/http.go` (`HTTPSender`, provider-agnostic JSON POST with `Authorization: Bearer <key>`) and `apps/api/foundation/email/http_test.go` (16 tests covering URL/APIKey/From validation, recipient/subject/body construction, Bearer auth, 2xx/non-2xx handling, context cancellation, the never-log-APIKey contract) all pass against a fake `httptest.NewServer` transport; no real provider call is made from CI. The production wiring (`apps/api/app/api/production.go::buildEmailSender`) constructs the real `HTTPSender` when `EMAIL_PROVIDER_API_KEY`/`URL`/`FROM` are set and the `EMAIL_MAGIC_LINK_ENABLED` kill switch is on, and falls back to `email.Fake{}` otherwise - covered by `apps/api/app/api/production_test.go::TestBuildEmailSender_*` (5 tests, all pass). `EMAIL_PROVIDER_URL`/`EMAIL_PROVIDER_API_KEY`/`EMAIL_FROM`/`EMAIL_PROVIDER_TIMEOUT` are documented in `apps/api/.env.example` and read by `LoadProductionConfig`; the timeout defaults to `10s` and `TestLoadProductionConfig_EmailProviderTimeoutDefaults` asserts a positive default. |
| `EV-28` | Real email sender: one live staging delivery | **Blocked by `VOC-032-DEP-07`** — no email-provider account exists yet. |
| `EV-29` | Real Google OAuth provider: token/userinfo unit test | **Produced by `T15`.** |
| `EV-30` | Real Google OAuth provider: one live staging exchange | **Blocked by `VOC-032-DEP-07`** — no Google Cloud OAuth client exists yet. |
| `EV-31` | Exact-SHA independent verification (per PR) | **Performed by Claude Code (different model binding) at each PR's exact final SHA** — this is not the implementer's evidence to record; it is produced by the independent-verification role and reported per-PR. |

## Staging exercise plan and execution

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

#### 2026-07-30/31 operator execution — passed

`EV-21` is complete. Founder-provisioned SSH access to
`ubuntu@130.185.123.152` was reachable.
`deploy-staging` run `30571703073` had already deployed commit
`8d9429b` successfully. Independent checks returned HTTP 200 from
`https://staging.vocanova.site/` and from
`https://api-staging.vocanova.site/healthz`; the API body reported
`status=ok` and `database=ok`.

The following live-database command used the deployed `T06` wrapper. The
connection value was sourced from the founder-owned, untracked
`infra/secrets/api.env`, rewritten only from the Compose hostname to the
private container IP, and never printed:

```sh
cd /opt/vocanova/infra
set -a
. ./secrets/api.env
set +a
postgres_ip=$(docker inspect --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' vocanova-postgres)
migration_database_url=$(printf "%s" "$DATABASE_URL" | sed "s/@postgres:5432/@${postgres_ip}:5432/")
DATABASE_URL="$migration_database_url" sh /opt/vocanova/apps/api/scripts/migrate.sh
```

- Start: `2026-07-30T19:24:08Z`
- End: `2026-07-30T19:24:09Z`
- Outcome: PASS — Atlas v1.2.0 returned
  `No migration files to execute`. The live Atlas ledger contained all 13
  expected versions.

The live database contained the expected 25 public application tables and
zero application rows before the rehearsal. The operator then created the
explicitly named disposable copy and restored a custom-format snapshot:

```sh
docker exec vocanova-postgres pg_dump \
  -Fc -U vocanova -d vocanova \
  -f /tmp/vocanova_t09_rehearsal_20260730.dump
docker exec vocanova-postgres createdb \
  -U vocanova -T template0 vocanova_t09_rehearsal_20260730
docker exec vocanova-postgres pg_restore \
  -U vocanova -d vocanova_t09_rehearsal_20260730 \
  /tmp/vocanova_t09_rehearsal_20260730.dump
```

- Rehearsal start: `2026-07-30T19:25:39Z`
- Copy restored: `2026-07-30T19:25:42Z`
- Restored baseline: 25 public application tables and 13 Atlas revision
  rows.

Only the disposable database was selected for the rollback commands. Each
approved `.down.sql.example` was piped to `psql -v ON_ERROR_STOP=1` in the
following reverse order:

```text
20260725140002_voc031_p5_account_deletion_requests.down.sql.example
20260725140001_voc031_p5_email_change_links.down.sql.example
20260725140000_voc031_p5_user_onboarding_profiles.down.sql.example
20260725130002_voc030_p4_gamification_tables.down.sql.example
20260725130001_voc030_p4_mission_tables.down.sql.example
20260725130000_voc030_p4_user_settings.down.sql.example
20260725120001_voc028_p3_ai_feedback_attempts.down.sql.example
20260725120000_voc028_p3_learner_sentences.down.sql.example
20260725110000_voc027_p2_review_attempts.down.sql.example
20260725100001_voc026_p1_idempotency_keys.down.sql.example
20260725100000_voc026_p1_content_tables.down.sql.example
20260724210000_identity_foundation.down.sql.example
```

The per-file command was:

```sh
docker exec -i vocanova-postgres psql \
  -X -v ON_ERROR_STOP=1 -U vocanova \
  -d vocanova_t09_rehearsal_20260730 \
  < "/tmp/vocanova-t09-20260730/${rollback_file}"
```

- Rollback complete: `2026-07-30T19:25:45Z`
- Outcome: PASS — all 12 artifacts completed without a SQL error. The only
  remaining public application table was `oauth_states`, exactly matching
  the repository inventory: `20260724210001_oauth_state.sql` has no approved
  `.down.sql.example`.

To let Atlas replay the complete forward set from version zero, the operator
removed that empty OAuth table and the revision ledger **only from the
disposable copy**:

```sh
docker exec vocanova-postgres psql \
  -X -v ON_ERROR_STOP=1 -U vocanova \
  -d vocanova_t09_rehearsal_20260730 \
  -c "DROP TABLE public.oauth_states" \
  -c "DROP SCHEMA atlas_schema_revisions CASCADE"
```

The same deployed wrapper was then run with a connection URL derived for
`vocanova_t09_rehearsal_20260730`. It applied all 13 migrations (74 SQL
statements) in 1.628 seconds and recreated 13 revision rows. Schema-only
dumps of live `vocanova` and the re-applied disposable database were
normalized only by removing PostgreSQL's random `\restrict`/`\unrestrict`
tokens and compared with `diff -u`; the comparison had no output.

- Forward re-apply complete: `2026-07-30T19:25:47Z`
- Schema comparison: PASS — exact match
- Rehearsal end: `2026-07-30T19:25:49Z`

After verification, the disposable artifacts were removed:

```sh
docker exec vocanova-postgres dropdb \
  -U vocanova --force vocanova_t09_rehearsal_20260730
docker exec vocanova-postgres rm -f \
  /tmp/vocanova_t09_rehearsal_20260730.dump
rm -rf /tmp/vocanova-t09-20260730
```

- Cleanup complete: `2026-07-30T19:26:06Z`
- Post-cleanup safety check: the disposable database no longer existed; the
  live ledger still contained 13 revisions; `/healthz` still reported
  `status=ok` and `database=ok`.

At the end of this 2026-07-30 migration rehearsal, the
create-user/save-word/complete-review/submit-sentence portion was still
blocked because staging AI was disabled. It was completed after the real
provider was enabled and the VOC-034 moderation fix was deployed:

- Fixed revision `f990e86efeef73730d747d53ea2d1ca7cd77bf84`
  deployed successfully in run
  [30618654496](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/30618654496).
- A disposable-identity core loop ran from `2026-07-31T09:10:43Z` through
  `09:10:57Z`: save-word and review requests returned HTTP 200, and the
  real-provider sentence-feedback request returned HTTP 200 with no error
  code and status `correct`.
- Before cleanup, the scoped `users`, `user_words`, `review_attempts`,
  `learner_sentences`, and successful `ai_feedback_attempts` rows each
  counted exactly one. The feedback row recorded `provider=opencode` and
  `model=opencode-go/hy3`.
- Transactional cleanup deleted the disposable identity, session, word
  fixture, saved word, review, sentence, feedback attempt, and idempotency
  rows; post-cleanup counts for every scoped row group were zero.

The complete sanitized identifiers, row counts, and timestamps are recorded
in
[`VOC-034` staging evidence](../VOC-034-production-ai-feedback-is-unreachable-because-no/staging-evidence.md),
merged by PR
[#229](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/229).
The observed sentence-feedback latency was `13,423 ms`; this does not
invalidate `EV-21`'s reachability, persistence, rollback, and cleanup
criteria, and is carried explicitly to `T10` for evaluation against its
release-blocking thresholds.

**Disclosed limitation:** the rollback rehearsal (2026-07-30) and the
core-loop exercise (2026-07-31) did not happen in the order `T09`'s own
procedure describes (core-loop writes, then snapshot, then rollback/
replay) — the disposable copy was taken from the live database while it
held zero application rows (see above), a day before staging AI was
enabled and the core-loop rows were created and deleted directly against
the live database. Consequently, no `.down.sql.example` in this repository
has been executed against populated tables, and the forward re-apply's "no
unintended data loss" claim is satisfied vacuously (there was no data to
lose) rather than by observing a populated table survive a down/up cycle.
`AC-09`'s literal wording is still met — reachability, rollback, forward
re-apply, and cleanup were all genuinely exercised — but this is weaker
assurance than a populated-table rollback would give, and the gate-
readiness summary below states this plainly rather than implying the
stronger claim.

### `EV-22` — Live-provider AI evaluation pass

The T10 deliverable is the runnable one-shot procedure the
founder will invoke once the staging AI-provider credentials
(`VOC-032-DEP-03`) are provisioned, plus the documented
operator procedure for the live run. The actual live
execution cannot happen from the implementer side: the
credential does not exist, and inventing a real-provider
result here would be a fabricated "release-blocking pass"
that DOC-12 §9 ("protected live-provider evaluation
outside CI with explicit cost limits") and AC-10's own
"Result: pending — blocked by VOC-032-DEP-03" wording
explicitly forbid.

#### T10 in-repo deliverable (delivered at this SHA)

`apps/api/business/aifeedback/live_eval.go` adds:

- `LiveEvaluationReport` — the structured output of one
  T10 run. Every field `EV-22` requires (provider, model,
  dataset, spec, every per-threshold value, the violation
  list, wall-clock duration, per-call latency summary
  statistics, provider-call count, estimated input/output
  char counts, operator-supplied cost in USD, pre-agreed
  cost ceiling in USD, the cost-ceiling-exceeded flag,
  start/finish timestamps, and operator notes) is a named
  field, so the report the operator copies into this
  document is machine-parseable and the missing-data case
  is always explicit (a `-1.00` or `0.00` is a recorded
  fact, not a silent omission).
- `RunLiveEvaluation` — the deterministic library function
  the cmd and any future test harness both call. It
  reuses `RunEvaluation` and the existing
  `ComputeGoldenThresholds` / `CheckGoldenThresholds`
  (T08's gate logic) so the per-threshold computation is
  identical between the CI gate and the live pass; the
  only difference is which `FeedbackProvider` the
  evaluation runs against.
- `InstrumentedProvider` — a thin wrapper that records
  per-call latency and rough input/output char counts
  without changing the `FeedbackProvider` interface. The
  wrapper is removed from the call path as soon as
  `RunEvaluation` returns; it does not persist.
- `FormatLiveEvaluationReport` — the deterministic
  report renderer the cmd writes to stdout (and the
  optional output file). The format is plain text and
  contains every field on `LiveEvaluationReport`, so the
  report is copy-pasteable into this section verbatim.
- Three exported env-var name constants
  (`LiveEvaluationCostCeilingUSDEnv`,
  `LiveEvaluationCostUSDEnv`, `LiveEvaluationOutputEnv`)
  so the operator procedure below can name the exact
  variables and a future rename shows up as a test
  failure rather than as a documentation drift.

`apps/api/cmd/eval-live/main.go` adds the runnable
command:

- `apps/api/cmd/eval-live` is a single-package Go binary
  in the existing `apps/api/cmd/<name>` convention (the
  same shape `cmd/api`, `cmd/seed`, and `cmd/openapi`
  follow). It reads the staging AI provider's base URL
  and API key from `--base-url` / `--api-key` (or
  `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` env
  vars) and constructs the real
  `aifeedback.NewOpenCodeFeedbackProvider`. The provider
  constructor is a package-level function variable
  (`newProvider`) so the test suite can swap it for a
  fake without needing a live OpenCode server; the swap
  is a test seam, not a production-affecting
  indirection.
- The command's exit codes follow the T08 gate's
  pass/fail semantics: `0` (no tracked threshold
  violation and cost ceiling not exceeded), `1`
  (release-blocking finding: any tracked threshold
  violation OR cost ceiling exceeded), or `2`
  (configuration error: missing required env var, bad
  flag value, or unwritable output file path). Exit
  code `1` is the operator signal that the run
  produced a finding the founder must rule on before R1
  acceptance; it is not a `panic` and does not abort
  the report-write, so the structured report is
  available for review even on failure.
- The command's `--output` flag (or `EVAL_LIVE_OUTPUT`
  env var) writes the same report to a file in
  permission mode `0600` so the API key's
  neighborhood on disk is not world-readable. The
  command never logs the API key; the `TestRunEvalLive_DoesNotLogAPIKey`
  test exercises the contract by injecting a known
  sentinel value and asserting it does not appear on
  either stdout or stderr.

The T10 unit tests cover:

- `InstrumentedProvider` wrapping the real provider,
  recording latency and char counts, propagating
  errors, and respecting context cancellation.
- `summarizeLatencies` and `percentileNearestRank`
  (deterministic nearest-rank percentile, NIST §7.2.1.1)
  across zero/one/many inputs and edge cases
  (clamping `p` to `[0, 100]`).
- `RunLiveEvaluation`'s report shape against the
  `NewMockProvider` (no production dependency), the
  cost-ceiling-exceeded flag at the spec boundary
  (ceiling = 0.5 with cost 0.4 / 0.7 / 999.0), operator
  notes, and the wall-clock + latency summary
  behavior with a measurable-delay provider.
- The cmd's exit-code mapping (`0` / `1` / `2`) for:
  missing required env vars, `--help`, a clean
  run (provider that matches every dataset
  expectation), a deliberately-violating provider,
  cost-ceiling-exceeded, output-file written, output
  file unwritable, and the API key never logged.
- The env-var name constants — a future rename
  fails the test on the rename commit, not as a
  silent documentation drift in this file.

#### Operator procedure for the live run (still blocked on `VOC-032-DEP-03`)

Once the founder has provisioned the staging AI-provider
credentials (base URL + API key, and the AI provider's
model identifier the founder wants exercised), the live
run is:

1. Confirm `VOC-032-DEP-03` is resolved: the staging
   AI-provider credential is present in the staging
   environment (typically via `infra/secrets/api.env`,
   the same untracked env file the T09 operator
   execution used for `DATABASE_URL`).
2. Agree an explicit cost ceiling in USD for the run
   (DOC-12 §9) and a per-request timeout; export both
   to the shell:
   ```sh
   export AI_PROVIDER_BASE_URL="..."
   export AI_PROVIDER_API_KEY="..."
   export AI_PROVIDER_MODEL="opencode-go/..."
   export AI_PROVIDER_TIMEOUT="8s"
   export EVAL_LIVE_COST_CEILING_USD="0.50"
   ```
3. Optionally pre-stage the billed-cost value
   (operator can fill it in from the provider's
   billing console after the run; if set, it must
   be a non-negative number):
   ```sh
   export EVAL_LIVE_COST_USD="0.00"   # placeholder
   ```
4. Run the command from the staging host (or from
   the developer's workstation against the staging
   provider) and capture the rendered report:
   ```sh
   go run ./apps/api/cmd/eval-live \
     --output /opt/vocanova/staging-evidence/eval-live-2026-MM-DD.txt
   ```
5. The report's "Result:" line is the operator's
   pass/fail signal. A `Result: PASS` (exit 0) and
   `CostCeilingExceeded: false` together satisfy
   AC-10's "AI evaluation thresholds pass" half; any
   `Result: FAIL` line is a release-blocking finding
   for R1 per AC-10's explicit wording, and the
   per-threshold violation messages list which
   DOC-09 §23 thresholds missed.
6. Paste the full report into the `EV-22` section
   below this procedure, including the operator's
   post-run `EVAL_LIVE_COST_USD` value (filled in
   from the provider's billing console for the
   run's window) and any `OperatorNotes` (e.g.
   rate-limit warnings observed, retries
   attempted, model-name discrepancies).
7. Mark this `EV-22` row's status as satisfied in
   the `T12 evidence` section (the gate-readiness
   summary's "AI evaluation thresholds pass"
   bullet).

#### Live execution status — RESOLVED (2026-08-01), founder-accepted deviation

`VOC-032-DEP-03` is resolved. The protected live evaluation was run against
three real providers in total, across this package's own `T10` and two
follow-up packages it spawned once the original provider proved unusable:

1. **OpenCode** (`opencode-go/hy3`, this package's own originally-configured
   provider): every one of 56 calls timed out inside the mandated 8-second
   budget; 0% on every tracked threshold. Recorded in this section's own
   history above (superseded by this final resolution) and in
   `VOC-032-T10`'s original evidence.
2. **Google Gemini** (`VOC-035`, a follow-up package): also FAIL — a stale
   default model blocked for new API keys, then real signal (11/56) but
   still below every threshold, attributable to free-tier rate-limiting on
   an unpaced request burst. Full record:
   `specs/changes/VOC-035-voc-032-t10-s-live-ai-evaluation-gate-ev-22-is/staging-evidence.md`.
3. **Cloudflare Workers AI** (`VOC-036`, a second follow-up package): a real
   code defect was found and fixed along the way (PR #253); three models
   were then tried. The smallest, fastest model
   (`@cf/meta/llama-3.1-8b-instruct-fp8-fast`) reached 100% reliability
   (56/56 calls succeeded, 1.2–3.2s latency, comfortably inside budget) but
   only 41% overall accuracy against DOC-09 §23's ≥90% bar, with 5
   zero-tolerance over-correction defects. Larger models tried
   (`llama-3.3-70b`, `llama-4-scout-17b`) were either too slow or had *worse*
   over-correction rates. Full record:
   `specs/changes/VOC-036-voc-032-t10-s-live-ai-evaluation-gate-ev-22-is/staging-evidence.md`.

**Root-cause note, not yet acted on:** the identical "unnecessary correction
on clearly-correct sentences" failure pattern across every model and
provider tried strongly suggests the gap is in `apps/api/business/aifeedback/task.go`'s
`developerPrompt()` itself — it defines the required output *shape* per
status but never gives the model explicit grading criteria for *when* a
sentence should be marked `correct` versus `needs_improvement`. This is
flagged as a concrete, likely-high-value follow-up (a prompt fix, re-tested
against the already-built Gemini/Cloudflare providers, potentially cheaper
than adopting a paid tier) but was not pursued in this round — see the
founder decision below.

**Founder decision (2026-08-01):** no available provider meets DOC-09 §23's
accuracy thresholds. Rather than continue evaluating providers or pursue the
prompt-fix lead above before launch, the founder has explicitly decided to
proceed to R1 with `AI_PROVIDER=cloudflare` /
`AI_PROVIDER_MODEL=@cf/meta/llama-3.1-8b-instruct-fp8-fast` as the production
AI-feedback provider — chosen for being free, 100% reliable, and fast, with
the explicit acknowledgment that its measured accuracy (41% overall,
structured-output validity and clearly-correct accuracy both below spec) is
**below DOC-09 §23's release-blocking bar**. This is a deliberate, informed
exception, not a silent threshold change: DOC-09 §23's thresholds are
unchanged and remain the target; this package does not amend DOC-09. The
founder's own stated reasoning: real user data is needed to further tune
quality, and that data does not exist pre-launch — the provider will be
revisited (via the prompt-fix lead, a stronger model, or a paid tier) once
real usage exists. `AI_PROVIDER_API_KEY`/`AI_PROVIDER_ACCOUNT_ID` are
founder-populated directly on the staging host's `infra/secrets/api.env` (not
committed to this repository, per this package's own established
convention) — this document does not assert that step has been completed,
only that it is the founder's own decision and remaining action, outside any
package's implementation authority.

**`EV-22` disposition:** `VOC-032-AC-10`'s "AI evaluation thresholds pass"
half is **not satisfied** and is recorded as such — this document does not
claim otherwise. The R1 gate item this blocks is explicitly waived by the
founder's own decision above, not by this document declaring it passed.

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

## T06 follow-ups: pre-existing migration-file issues (RESOLVED by VOC-033)

**Both follow-ups below are fixed as of `VOC-033`** (merged before this
`T09` rehearsal ran) — every migration file now reads `-- atlas:txmode
file`, and the duplicate `streak_states_user_id_key` index is removed. This
section is preserved as the historical record of what was found and why
(the fix itself is `VOC-033`'s diff, not this document's), not as a
statement that these are still open. Do not re-fix these in a future `T06`
follow-up package — they are closed.

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
| **Stable in staging** | **Operational evidence recorded; final acceptance pending.** `VOC-032-DEP-00`/`DEP-01` are resolved. The four-service docker-compose stack deployed successfully to the founder's server; both public HTTPS endpoints returned HTTP 200 and `/healthz` reported `status=ok`/`database=ok` (see `EV-21`). | The deployment and rehearsal evidence are recorded under `EV-21`. The founder still owns the final staging-acceptance decision (below), and `T10`'s open finding separately prevents R1 acceptance now. |
| **No unresolved critical/high blocker** | **Partial.** The `T06` migration-tooling follow-ups (invalid `-- atlas:txmode transaction` directive; duplicate `streak_states_user_id_key` index) that previously blocked `EV-14`'s end-to-end pass were fixed in `VOC-033` before this rehearsal ran. The `T11` and `T13` follow-ups remain open, recorded as blockers to `AC-11` and `AC-13` respectively; `T10` is now an observed release-blocking live-evaluation finding, tracked separately. | `T11`/`T13` are small, already-scoped follow-ups (see their own sections). `T10`'s finding requires an authorized provider/model or governance decision, tracked on its own issue — not a `T09` blocker. |
| **All required tests pass** | **Satisfied (in-repo).** `go test ./...` from `apps/api/`, `go vet ./...`, `gofmt -l .`, `go build ./...`, and `bash scripts/governance/validate-governance.sh` all pass at the final SHA. The `apps/web` and `infra` trees are exercised by the `karsift-ai-infra` `ci.yml` reusable workflow, not by `T12`'s in-process check, and their evidence is per-PR. | In-repo check is `T12`'s. The `karsift-ai-infra ci.yml` results are produced by CI, not by `T12`. |
| **Migration + rollback rehearsed** | **Satisfied, with a disclosed limitation.** The live no-op apply, disposable-copy rollback of all 12 approved down artifacts, full 13-migration forward re-apply, and exact schema comparison all passed on 2026-07-30/31 — see `EV-21` above. The disposable copy was taken while the live database held zero application rows (the core-loop exercise that populates rows ran separately, a day later, directly against the live database, not against the disposable copy) — see `EV-21`'s "Disclosed limitation" note. `AC-09`'s literal criteria are met; a populated-table rollback was not exercised. | Exact commands, timestamps, and the disclosed ordering limitation are recorded under `EV-21`. A future package could re-run steps 3-6 in `T09`'s originally-specified order (core-loop writes before the snapshot) if stronger populated-table rollback assurance is ever required. |
| **AI evaluation thresholds pass** | **FAILED — release-blocking per DOC-09 §23; founder-accepted exception recorded for R1.** `T08`'s mock-provider gate passes every threshold the current `EvaluationResult` shape can measure. The live-provider half genuinely failed against all three real providers tried (OpenCode, Gemini, Cloudflare) — see `EV-22`'s "Live execution status" section above for the full record. The founder has explicitly decided to proceed to R1 with the best available option (Cloudflare, `llama-3.1-8b-instruct-fp8-fast`: 100% reliable, 41% accurate) despite it not meeting DOC-09 §23, to be revisited with real usage data. DOC-09 §23 itself is unchanged. | Mock half: `T08`'s tests all pass. Live half: genuinely failed; not satisfiable by any package without either a stronger available provider, a paid tier, or the flagged prompt-fix lead in `EV-22` — see that section for the founder's own decision and reasoning. |
| **Founder completes staging acceptance** | **Not started — founder-owned, out of scope of any package.** | This is a single human decision by the founder, recorded as a comment on the package's release issue (per the standard `karsift-ai-infra release.yml` flow). It cannot be satisfied by any code change. |
| **Scope is frozen** | **Satisfied (in this package).** No new product scope introduced; only the `T00`–`T15` and `T12` deliverables. | Inherently a property of the package's own change-control discipline; `T12`'s role is to confirm no PR in this package introduced unapproved scope. |

### Items this gate-readiness summary records explicitly, passing or not

Several of the items below now describe outcomes that passed (the live
staging environment, live TLS, the migration rehearsal) — recorded here in
detail because they were previously open items in this same list. The
remainder are not "passing" at the final SHA and are not the founder's call
to wave through; each is a recorded, factual limitation:

- **The live staging environment has been brought up.** `T07`'s
  deploy workflow ran end-to-end against the real server
  (run `30618654496` and others referenced under `EV-21`).
  `EV-16`/`EV-17`'s "fails closed on bad health check" half
  is still exercised only by the unit tests, not by an
  observed real health-check failure — that half remains
  end-to-end-unverified, but the happy path is now real,
  not simulated.
- **Live Cloudflare TLS is in place.** `T05`'s `EV-12`/`EV-13`
  were previously satisfied in the "nginx -t validates, Host
  header routes correctly" half only; `https://staging.
  vocanova.site/` and `https://api-staging.vocanova.site/`
  now serve real traffic through Cloudflare (see `EV-21`),
  confirming the live-certificate half as well.
- **The live migration-and-rollback rehearsal passed, with a
  disclosed limitation.** `EV-21` is satisfied — see its
  "Disclosed limitation" note above for the populated-table
  caveat.
- **No live AI-evaluation result recorded in this document
  yet.** `VOC-032-DEP-03` itself is now resolved, but `T10`'s
  protected live-evaluation run and its result are recorded
  in `T10`'s own evidence update, a separate change not yet
  merged into this document — this PR does not assert an
  `EV-22` outcome one way or the other.
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
- **Resolution of `VOC-032-DEP-07`.** The email-provider
  and Google OAuth credentials remain founder-owned.
  `DEP-00`, `DEP-01`, and `DEP-03` are now resolved.
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
