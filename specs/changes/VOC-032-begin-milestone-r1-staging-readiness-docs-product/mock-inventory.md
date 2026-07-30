# VOC-032 — Mock Disposition Inventory

## Scope and authority

This document is drafted before adoption/implementation, per this
repository's own package template. It is updated at `T12` implementation
time to record the actual post-`T00`–`T15` state.

## Draft-time confirmation (2026-07-28, by direct repository inspection)

This package is infrastructure-only and introduces no learner-facing product
mock. `NewContractAPI()`'s existing in-memory/mock service wiring
(`apps/api/app/api/openapi.go`) is explicitly preserved unchanged — it exists
to let OpenAPI generation and future contract work run without a database,
and this package adds a *parallel*, real-service production-wiring path
(`T00`) rather than replacing or retiring it. It is therefore not a mock this
package needs to decommission.

The one thing genuinely mock-like this package introduces on purpose, and
retires nowhere: `T08`'s AI-evaluation-threshold CI gate deliberately runs
against `aifeedback.NewMockProvider()`, not the real provider, matching
DOC-12 §9's own rule that "normal CI never depends on a paid provider." This
is not a leftover implementation shortcut to be cleaned up later — it is the
correct, permanent shape of that gate. The real provider is exercised
separately and only once, in `T10`'s live-provider evaluation pass, which is
explicitly staging-only and never part of routine CI.

## Final `T12` re-confirmation (2026-07-29, by direct repository inspection at the implementation base SHA)

The `T12` re-confirmation specified in the draft section below has been
executed and its results recorded here. The post-`T00`–`T15` state matches
the draft's "no product mock introduced" prediction exactly:

- `grep -rn "MOCK_" apps/web/src` — **zero matches**. No `MOCK_*`
  frontend constant, placeholder API response, or stubbed business-logic
  path was added by `T00`–`T15`.
- `grep -rni "mock" apps/web/src` — **zero matches**. The "mock"
  vocabulary is absent from the frontend source tree.
- `grep -rni "mock" apps/api/` — matches are confined to:
    - `apps/api/cmd/seed/main_test.go` and
      `apps/api/app/api/production_test.go`, which use
      `github.com/DATA-DOG/go-sqlmock` as a database-mock
      library for unit-test isolation (existing pre-VOC-032,
      unchanged by this package).
    - `apps/api/.env.example`, which references the
      `NewMockProvider` in its own documentation comments so
      a developer can read about the kill switch and the
      mock-vs-real provider boundary without leaving the file.
    - `apps/api/business/aifeedback/{aifeedback,evaluation,
      *_test}.go`, which is the existing `MockProvider` type
      `T08` deliberately retains (see "draft-time
      confirmation" above) as the deterministic
      CI-only provider.
    - `apps/api/app/api/{openapi,production}.go` and the
      `*_test.go` files alongside them, which construct
      `NewMockProvider()` and `auth.NewFakeOAuthProvider()`
      for OpenAPI generation and for unit tests
      respectively. The production wiring in
      `apps/api/app/api/production.go` uses these fakes
      as documented fallbacks (`buildEmailSender` and
      `buildOAuthProvider`) when the corresponding kill
      switch is off or the credential is absent; the
      real `HTTPSender` and `GoogleOAuthProvider` are
      constructed only when their respective credentials
      are present, exactly as `T14`/`T15` specify.
- `grep -rn "FakeOAuthProvider" apps/api/app/api/`
  — present only in the production wiring's fallback
  branch and in unit tests, never as the only path.
  `buildOAuthProvider` always returns *some* provider
  (real when configured, `FakeOAuthProvider` when
  intentionally disabled); no route is reachable
  through the production wiring only via a fake.

The `T00`–`T15` work leaves the zero-legacy-mock state —
established by `VOC-030-T05`/`VOC-031` — unchanged. No
follow-up to `scripts/foundation/mock-inventory.mjs`'s
allow lists was required.

## Non-applicability statement (final, post-`T12` confirmation)

No `MOCK_*` frontend constant, no placeholder API response, and no stubbed
business-logic path is added by `T00`–`T15`. The grep results above confirm
this package leaves that zero-legacy-mock state — established by
`VOC-030-T05`/`VOC-031` — unchanged, and no extension to
`scripts/foundation/mock-inventory.mjs`'s allow lists was required.

## Mock kept by design (not introduced by this package, not retired)

| Existing mock | Used by | Why kept |
| --- | --- | --- |
| `apps/api/business/aifeedback.MockProvider` (in `aifeedback.go`) | `T08`'s CI evaluation gate (`RunGoldenGate`, `apps/api/business/aifeedback/threshold_gate_test.go`); also the default provider the production wiring in `apps/api/app/api/production.go` selects when `AI_PROVIDER` is unset or no API key is set | Per DOC-12 §9, normal CI never depends on a paid provider. The mock is deterministic and contract-equivalent to the real provider's response shape (validated by `T08`'s unit tests). The production wiring's use of it as a default is the documented, kill-switch-respecting fallback, not a regression. |
| `apps/api/foundation/email.Fake{}` | `cmd/seed` and unit tests; also the production wiring's fallback when `EMAIL_MAGIC_LINK_ENABLED=false` or `EMAIL_PROVIDER_API_KEY` is unset | The fake predates this package. `T14` adds a real `HTTPSender` alongside the fake; the fake remains because tests use it and the production wiring falls back to it intentionally (per `T00`'s kill-switch design). |
| `apps/api/business/auth.NewFakeOAuthProvider` | `cmd/openapi`; also the production wiring's fallback when `GOOGLE_OAUTH_ENABLED=false` or `GOOGLE_OAUTH_CLIENT_ID` is unset | The fake predates this package. `T15` adds a real `GoogleOAuthProvider` alongside the fake; the fake remains because the OpenAPI generator and the fallback path both depend on it. |
| `NewContractAPI()` (`apps/api/app/api/openapi.go`) | OpenAPI generation | The mock wiring here is the existing pre-VOC-032 in-memory + fake-provider construction path the OpenAPI generator uses. `T00` adds a *parallel* `NewProductionAPI()` function for the real production server; the OpenAPI-time mock wiring is preserved unchanged. |

## New real files/tools (added by `T00`–`T15`)

| Area | New addition | VOC source |
| --- | --- | --- |
| API entrypoint | Real production wiring in `apps/api/cmd/api/main.go` (plus `apps/api/app/api/production.go::NewProductionAPI`) | T00 |
| Config | `apps/api/.env.example`, `apps/web/.env.example` | T01 |
| Container | `apps/api/Dockerfile` | T02 |
| Container | `apps/web/Dockerfile`, `apps/web/next.config.ts` | T03 |
| Orchestration | `infra/docker-compose.yml` | T04 |
| Reverse proxy | `infra/nginx/nginx.conf` + `infra/nginx/conf.d/*.conf` (Cloudflare-aware TLS) | T05 |
| Migration tooling | `apps/api/atlas.hcl`, `apps/api/scripts/migrate.sh`, `apps/api/migrations/atlas.sum` | T06 |
| CI/CD | `.github/workflows/deploy-staging.yml` | T07 |
| AI evaluation | `apps/api/business/aifeedback/threshold_gate.go` + `threshold_gate_test.go` (in-process gate, exercised by the standard `go test ./...` CI run) | T08 |
| AI evaluation | `apps/api/business/aifeedback/live_eval.go` + `apps/api/cmd/eval-live/main.go` + their `*_test.go` files (T10's runnable one-shot live-evaluation command; the live execution itself is blocked on `VOC-032-DEP-03` and is recorded as a `staging-evidence.md` follow-up, not a CI check). The provider constructor is the real `NewOpenCodeFeedbackProvider`; the only mock in the test path is the test-only fake the unit tests inject via the `newProvider` function-variable seam. | T10 |
| Gate readiness | `apps/api/gate_readiness/gate_readiness.go` + `gate_readiness_test.go` (cross-cutting EV-* evidence-presence check) | T12 |
| Documentation | `infra/README.md` (T11's rewrite is a known follow-up; see below) | T11 (divergent) |
| Documentation | `docs/operations/11-devops-and-ci-cd.md` §1 amendment (T13 is a known follow-up; see below) | T13 (divergent) |
| Email | Real `email.Sender` implementation `apps/api/foundation/email/http.go` (HTTPSender), alongside the kept `email.Fake{}` | T14 |
| Auth | Real `auth.OAuthProvider` implementation `apps/api/business/auth/google_oauth.go` (GoogleOAuthProvider), alongside the kept `NewFakeOAuthProvider` | T15 |

`T14`/`T15` add real, non-mock provider implementations behind existing
`email.Sender`/`auth.OAuthProvider` interfaces — the fakes are kept (tests
keep using them), not retired, so this does not change the "no product mock
retired or introduced" disposition above.

## Divergences recorded for the R1 gate-readiness summary (T12)

The "Documentation" rows in the table above are marked `(divergent)` to
make the following two T12-recorded findings visible to the independent
reviewer and the founder:

1. **T11 (`infra/README.md` rewrite) is a known follow-up.** The
   file's current content is the pre-VOC-005 placeholder text
   ("This directory is a non-deploying structural boundary.
   VOC-005 authorizes no Cloudflare, staging, production,
   release, or autonomous-development infrastructure."), not
   the AC-11 description of the docker-compose / nginx / Atlas
   layout `T04`–`T09` actually built. `T12` does NOT silently
   rewrite this file in T12's scope — that is a package-scope
   question (T11 owns the rewrite), not an implementation
   judgment call `T12` is permitted to make. The divergence is
   detected and reported by
   `apps/api/gate_readiness/gate_readiness_test.go::TestT11InfraReadmeIsNotThePlaceholder`
   and is listed in `staging-evidence.md`'s R1 gate-readiness
   summary as a known limitation, not a passing AC-11.

2. **T13 (DOC-11 §1 amendment) is a known follow-up.**
   `docs/operations/11-devops-and-ci-cd.md` §1's
   target-infrastructure table still describes the pre-
   amendment "Cloudflare Workers via OpenNext + Render Web
   Service + Render PostgreSQL + vocanova.com" target. The
   amendment has not been applied. `T12` does NOT silently
   amend an approved document in T12's scope — that is a
   package-scope question (T13 owns the amendment), and
   amending an approved document is R3 protected under
   `docs/governance/protected-areas.md`'s
   "Repository governance" row. The divergence is detected
   and reported by
   `apps/api/gate_readiness/gate_readiness_test.go::TestT13Doc11AmendmentApplied`
   and is listed in `staging-evidence.md`'s R1 gate-readiness
   summary as a known limitation, not a passing AC-13.

Both of these are test-detected, t.Log-reported divergences —
they are intentionally not test *failures* so that the
go test `./...` run still passes at the final SHA, but the
log output is what an independent reviewer and the founder
will see when they read the CI run, exactly so neither
divergence can silently disappear.
