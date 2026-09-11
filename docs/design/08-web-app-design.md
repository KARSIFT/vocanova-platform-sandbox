---
id: DOC-08
title: VocaNova Web Application Design
version: 1.0
document_type: web-application-design
status: approved
owner: founder
canonical_path: docs/design/08-web-app-design.md
approved_at: 2026-07-21
last_reviewed_at: 2026-07-21
review_cycle: quarterly
supersedes: null
related_documents:
  - DOC-03
  - DOC-04
  - DOC-07
  - DOC-09
related_decisions: []
adoption_change: VOC-008
source_files:
  - path: 08-web-app-design.md
    sha256: da9154f1962e52f5046c712e581f5627122f48aec86684b24f69de1b9ee129d5
---

# 08 — VocaNova Web Application Design

## Summary

Responsive, mobile-first Next.js web application integrated with the Go backend API. Core loop:
discover useful words → save → review with spaced repetition → use words in learner sentences →
receive lightweight AI feedback → build daily habit.

## Frontend foundation

Next.js App Router + TypeScript, deployed in Docker alongside the Go API and PostgreSQL.
The current staging/production deployment workflows and [development guide](../development.md)
are authoritative; the earlier Cloudflare Workers + Render split is not the deployed topology.
Go `/api/v1` is the backend authority. Single repo,
frontend under `apps/web` per [04](../engineering/04-technical-architecture.md) §5. pnpm and Tailwind.
The current implementation uses server components for page data, a shared typed fetch client,
and controlled React forms for interactive learning. Shared presentation primitives live in
`src/ui/`; adding a state or form library is not itself a product completion requirement.
Node's test runner covers client helpers and middleware; Playwright covers browser behavior.

## Architecture

```text
src/app/            # routes, layouts, and route-local components
src/ui/             # shared visual primitives
src/lib/            # API integration, session, cookies, and helpers
tests/e2e/          # local browser and accessibility checks
tests/staging-e2e/  # deployed learning-flow verification
```

Feature areas: auth, onboarding, dashboard, discovery, words, reviews, sentences, progress,
settings.

The [learning workspace visual direction](learning-workspace.md) defines shared
branding, responsive navigation, reading widths, and the browser review workflow.

## Routing

Route groups: `(public)`, `(onboarding)`, `(app)`.

```text
/
/login
/magic-link
/onboarding
/home
/discover
/words
/words/[userWordId]
/review
/review/session
/progress
/progress/sentences
/settings
/settings/account
```

Sentence history is a secondary Progress route added by the
[maturity delivery](../product/mature-learning-and-account-experience.md).
Sentence practice remains a reusable component within learning activities.

## Core UX decisions

- **Home**: daily-mission focused; the next action is visible above the mobile navigation without
  scrolling at 360×640. Show backend-confirmed mission state, streak, and a compact saved-word
  preview. Sentence practice opens through a deliberate disclosure, not a form for every word.
- **Discovery**: one word at a time; backend controls content/sequencing; save must succeed before
  moving forward.
- **Review**: focused session, show-answer active recall, **ratings: Again, Hard, Good, Easy**.
  Result and rating remain distinct: objective incorrect answers record `Again`; objective correct
  answers allow Hard/Good/Easy; self-check prompts derive result from the chosen rating. See
  [05](../engineering/05-database-design.md) §9 and [06](../engineering/06-backend-design.md) §10;
  the backend controls scheduling and progress.
- **Sentence practice**: part of the MVP core loop, reusable component, accessible from Home, Word
  Detail, and Review Completion; AI feedback includes result, correction, explanation, improvement
  tip (see [09](../engineering/09-ai-features.md) for the full contract).
- **Progress**: simple, motivation-focused, backend-authoritative.
- **Settings**: learning preferences, account basics, account deletion with confirmation.

## API integration

Handwritten fetch wrapper against `/api/v1`, `credentials: "include"`, `X-CSRF-Token` on unsafe
methods, `Idempotency-Key` support, no frontend token storage. TypeScript types generated from Huma
OpenAPI via `openapi-typescript`; contract drift detected in CI.

## Quality standards

Mobile-first: target 360–430px, minimum 44px touch targets. Accessibility: WCAG 2.2 AA target,
keyboard support, focus management, screen-reader-friendly forms. Performance: lightweight
dependencies, route-level code splitting, Lighthouse targets Performance 85+ / Accessibility 95+ /
Best Practices 90+.

## Codex handoff order

Next.js setup → UI foundation → OpenAPI generation → API client → TanStack Query → routes/layouts →
authentication → onboarding → core learning loop → tests and accessibility.

## Claude Code review

Architecture boundaries, API contract usage, security, auth/session behavior, CSRF/idempotency,
accessibility, loading/error states, tests, overengineering.

## MVP completion criteria

Authentication, onboarding, home mission loop, discovery, saved words, review, sentence feedback,
progress, settings/account management all work; CI contract checks exist; critical flows are tested.
(Matches [DOC-01](../product/01-mvp-prd.md) §3 — restated here only as the web-app-specific checklist,
not a separate decision.)

## Account and appearance extension (0.2.0)

The [account/settings acceptance](../product/password-profile-and-theme.md)
extends the original MVP. Google and email/password are explicit configured
choices, with email verification and recovery states. Settings links to Profile
and Account security while retaining the three learning tabs. Profile edits the
existing display-name setting; email changes remain verified account operations.

Light, Dark and System are device preferences, defaulting to System. Apply the
resolved theme before hydration, follow OS changes only in System mode, and retain
readable error/success states and visible focus indicators in either palette.
Settings About exposes public release identity for support and deployment checks.
