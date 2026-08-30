---
id: DOC-04
title: VocaNova Technical Architecture
version: 1.0
document_type: technical-architecture
status: approved
owner: founder
canonical_path: docs/engineering/04-technical-architecture.md
approved_at: 2026-07-21
last_reviewed_at: 2026-07-21
review_cycle: quarterly
supersedes: null
related_documents:
  - DOC-05
  - DOC-06
  - DOC-07
  - DOC-08
  - DOC-09
  - DOC-10
  - DOC-11
  - DOC-17
related_decisions: []
adoption_change: VOC-008
source_files:
  - path: 04-technical-architecture.md
    sha256: 50ba0901ee5e877e98e7071c6930f809b0ebc6074858fd20e1ac7deae12403dc
---
# 04 — VocaNova Technical Architecture

## 1. Purpose and goals

Defines the technical architecture for Vocanova MVP: simple enough for a small team, secure by
default, scalable without premature complexity, ready for future mobile expansion.

## 2. Product technical direction

MVP platform: responsive, mobile-first web application, no native mobile app initially. Core loop:
discover → save → review (spaced repetition) → write sentence → receive AI feedback → build daily
habit.

## 3. Architecture principles

Modular monolith first; frontend/backend separated; business logic lives in the Go backend;
PostgreSQL is the source of truth; prefer simple stable technologies; avoid premature microservices;
design for future mobile clients; explicit domain boundaries; testable behavior; ADRs for important
decisions.

## 4. High-level architecture

```text
Browser
   |
Next.js Web Application
   |  HTTPS REST API
   v
Go Backend API
   |
   +-- PostgreSQL
   +-- Google OAuth + email magic-link delivery, with PostgreSQL-backed sessions
   +-- AI Provider
   +-- Feature flags
   +-- Observability
```

Future mobile: Next.js Web + Expo Mobile both call the same Go API / PostgreSQL.

## 5. Repository architecture

Single monorepo, `vocanova-platform` (see
[the migration notes](../archive/README-migration-notes.md#5-repository-name-conflict) for why this
name, not `vocanova`):

```text
vocanova-platform/
apps/
  web/          # Next.js
  api/          # Go backend
  mobile/       # Future Expo app
packages/
  api-client/
  design-tokens/
  eslint-config/
  typescript-config/
docs/
infra/
scripts/
.github/
```

## 6. Frontend architecture

Next.js (App Router) + React + TypeScript + Tailwind + shadcn/ui + TanStack Query + React Hook Form
+ Zod + Vitest + React Testing Library + Playwright + pnpm. Next.js never accesses PostgreSQL
directly; the Go backend remains the business authority; server state via TanStack Query.

## 7. Backend architecture

Go + chi + Huma v2 + Uber Fx + Ent ORM + PostgreSQL + `log/slog` + OpenTelemetry. Modular monolith
with business modules: `auth`, `user`, `settings`, `vocabulary`, `journey`, `review`, `sentence`,
`aifeedback`, `mission`, `progress`, `streak`. Business logic must not depend on HTTP, Huma, chi,
auth SDKs, or AI SDKs directly.

## 8. API architecture

REST, JSON, versioned under `/api/v1`. Huma generates OpenAPI 3.1. The generated artifact is
committed at `apps/api/openapi/vocanova.openapi.json` for TypeScript code generation and drift
detection; [06](06-backend-design.md) §5 and [07](07-api-contract-and-dto-design.md) define the
contract workflow. chi handles routing and middleware.

## 9. Authentication

Google OAuth + email magic link, no password login in MVP. Internal identity tables
(`users`, `user_identities`/`external_identities`) decouple business data from the external identity
provider — business tables reference Vocanova user IDs, never provider IDs directly.

## 10. Database architecture

PostgreSQL + Ent ORM. One database, clear domain ownership, reviewed migrations, UTC timestamps in
storage, IANA timezone strings for user-facing daily logic. Full schema in [05](05-database-design.md).

## 11. Spaced repetition

Deterministic stage-based scheduling (not FSRS in MVP). Rating scale, exact step mechanics, and
reset rule are canonical in [05](05-database-design.md) §9 — see
[the migration notes](../archive/README-migration-notes.md#2-review-rating-and-scheduling-conflict) for how the various
draft rating scales across documents were reconciled into one. Future algorithms can replace the
scheduler behind a stable interface.

## 12. Daily mission

A stable daily snapshot (user, local date, review target, selected items, completion status, policy
version). Settings changes apply from the next local day, not retroactively.

## 13. AI feedback architecture

AI purpose: help learners use vocabulary correctly. Canonical statuses are `correct` /
`needs_improvement` / `incorrect` (see
[the migration notes](../archive/README-migration-notes.md#1-ai-feedback-label-conflict) — this document originally used different
example labels; the authoritative model lives in [09](09-ai-features.md), not here). Architecture:
Business Service → Feedback Provider Interface → AI Provider. Rules: save the sentence before the AI
call, validate structured output, retry safely (bounded), store feedback history, control cost.

## 14. Progress and gamification

Backend owns Confidence Points, streaks, mission completion, and progress summaries. Points use an
event-based ledger with idempotency keys and transactional updates (see [05](05-database-design.md)
§12). Streak advances only after mission completion, uses local timezone, has gentle reset
behavior (grace days).

## 15. Background jobs

No Kafka, no complex queue in MVP. Simple synchronous workflows; lightweight cleanup jobs only
(expired sessions, expired magic links, old idempotency keys). Future: Temporal for long workflows,
transactional outbox for reliable events, if actually needed.

## 16. Security baseline

HTTPS only, strict CORS, security headers, input validation, authorization checks on every
learner-owned resource, rate limiting, secret management, secure database roles, privacy-aware
logging (no learner sentence text, no tokens/secrets in logs).

## 17. Observability

Structured logging, request IDs, OpenTelemetry, metrics, error tracking. Monitor API latency,
errors, database health, AI usage, job failures.

## 18. Testing strategy

Backend: unit, PostgreSQL integration, migration, API tests. Frontend: Vitest, React Testing
Library, Playwright. AI: fake-provider tests, evaluation dataset, controlled live tests (never a
paid provider in normal CI — see [09](09-ai-features.md) §23 Phase 1). Coverage is risk-based, not a flat
percentage target.

## 19. GitHub workflow

The repository uses `develop` and `main` as permanent branches with short-lived working
branches and governed pull requests. Exact merge, approval, and release authority is defined only by
[DOC-16](../governance/16-autonomous-development-operating-model.md),
[A-002](../governance/amendments/A-002-governed-autonomous-releases.md),
[A-003](../governance/amendments/A-003-governed-autonomous-engineering-authority.md),
[A-004](../governance/amendments/A-004-remove-founder-approval-gates-from-autonomous-engineering-workflows.md), and the
[approval matrix](../governance/approval-matrix.md). Governance permission does not imply that
automation is technically active; current activation is recorded in
[`repository-settings.md`](../governance/repository-settings.md) and the
[A-004 transition state](../governance/a004-transition-state.yaml).

## 20. CI/CD

Backend/frontend tests, type checks, security checks, generated-code checks in CI. Deploy to
development/staging/production. See [10](../operations/10-development-workflow.md) for the full pipeline and
the [canonical governance index](../governance/README.md) for merge/deploy authority, with
[DOC-19](../operations/19-governance-reconciliation-notes.md) available as orientation.

## 21. Scalability strategy

Start with modular monolith; extract services only when scaling, ownership, or reliability actually
require it and boundaries are proven. No microservices for MVP.

## 22. Future mobile architecture

React Native + Expo + Expo EAS, same API, when mobile work actually starts. Offline support is
postponed until then.

## Final technology stack

Frontend: Next.js, TypeScript, Tailwind, shadcn/ui. Backend: Go, chi, Huma, Uber Fx, Ent, PostgreSQL.
Auth: Google OAuth + email magic link. AI: provider abstraction, one provider operated at a time.
Infrastructure: see [10](../operations/10-development-workflow.md) (Cloudflare + Render, not a generic "managed
containers" placeholder — that document has the concrete, later infrastructure decision). Future:
Expo, Temporal.
