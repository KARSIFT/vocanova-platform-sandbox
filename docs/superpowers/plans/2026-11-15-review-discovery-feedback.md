# Review and Discovery Feedback Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Goal:** Deliver all ten owner annotations, effort-based daily goals and mixed review questions in one PR and release.

**Architecture:** Extend durable daily review rather than replacing it with practice. The API owns canonical-word effort credit, goal-sized sessions, frozen question modes and account discovery position. React consumes confirmed progress and adds post-answer teaching plus compact accessible navigation.

**Tech Stack:** Go 1.26.5, PostgreSQL, Huma, Next.js 16, React 19, TypeScript, pnpm 11.14.0, pinned Playwright 1.62.1.

**Spec:** `docs/superpowers/specs/2026-11-15-review-discovery-feedback-design.md`

## Global Constraints

- Existing checkout, branch `codex/review-discovery-feedback-20261115`; owner explicitly declines new directories/worktrees.
- One consolidated draft/hold PR, normal CI/review/merge queue, strict staging then manual production; user authorizes merge/deployment.
- Ordinary controls reuse tokens/components; learner progress remains server-authoritative and mutations idempotent.
- Vocanova ports 3100/8180, preserve Vergecloud 3000/8080 and prototype3102, keep Mac muted.
- No secrets, production configuration, live generated questions, fabricated coverage or historical streak rewriting.
- Root owns Git operations, generated contracts/API client, shared mock fixture, builds and servers; agents have explicit disjoint ownership.

## Review Focus

- A late response or retry after learner-local midnight must preserve action identity and avoid double credit (Task1).
- Repeated wrong answers/sibling meanings/parallel tabs cannot farm goal or completion rewards (Task1).
- Broken picture delivery and ambiguous gap options must leave a valid accessible question (Tasks1/2).
- Account switching, reordered position writes and explicit links must not resume another account or overwrite newer intent (Task3).
- An expired cookie or upstream outage must not redirect root/login in a loop or cache identity (Task4).

### Task 1: Authoritative daily review and mixed question API

**Owner/files:** backend implementer: `apps/api/business/dailylesson/**`, `apps/api/business/reviews/**`, relevant `apps/api/business/missions/**`, new `apps/api/migrations/20261115100000_daily_review_effort.sql`, `apps/api/app/api/daily_lessons.go` and its tests. No API-client/generated files, discovery code or web edits.

**Interfaces:** Add `DailyReviewGoal` with JSON `localDate`, `target`, `reviewed`, `remaining`, `complete`, optional `nextReviewAt`; expose `goal` on current/session responses. Keep legacy fields. Add `additional` boolean query parameter to POST daily-lessons; service `StartAdditional` or compatible optional argument supports deliberate extra review without breaking existing Start callers. Step kind union becomes `meaning_choice,picture_choice,gap_choice`; choices retain `id,text` with optional `image` matching existing MeaningImage DTO; step gains optional `context` for sentence gap. Option IDs and private answer remain frozen. Publish exact interfaces immediately to root/frontend.

- [ ] Add PostgreSQL regressions demonstrating goal15 starts15, wrong answers earn distinct-word effort but no successful-answer XP, skip earns no credit, concurrent/replayed/sibling answers count once; run baseline and record expected RED.
- [ ] Implement durable unique effort credit with atomic mission/reward/streak integration; reconcile only current local day from verified existing attempts, never rewrite completed past days. Preserve success records.
- [ ] Project one goal across Current/Get/Start/Act; use remaining allowance for new sessions, preserve old snapshots/replay responses, handle overnight sessions and explicit additional review.
- [ ] Add frozen picture/gap snapshots and real eligible approved catalog cases; test all4options, ambiguous/absent content fallback, revision mismatch, reload and answer grading before implementation. Reuse existing reviewed examples and approved images; no provider generation.
- [ ] Run focused Go and PostgreSQL tests with no relevant skips; include due<target, target changes mid-day, optional additional review, actual next due timestamp, historical session compatibility and midnight transitions. Root commits accepted changes after review.

### Task 2: Review/Home UI, post-answer teaching and relative scheduling

**Owner/files:** web implementer: `apps/web/src/app/(app)/reviews/**`, `apps/web/src/app/(app)/home/page.tsx`, new review-specific CSS/helpers/components and tests under `apps/web/tests/e2e/review-owner-feedback.spec.ts`, helper unit tests. May narrowly adjust `MeaningPicture` contract documentation to allow explicit balanced picture options; do not edit shared mock/API-client/globals without root coordination.

**Consumes:** Task1 goal/step types; root updates API client. `startDailyLesson(key, init?)` accepts additional via query through a dedicated optional argument agreed with root; existing callers remain source-compatible.

- [ ] Add regressions for goal15 count, wrong-answer completion, completed-session truthiness, all-available vs goal-complete, optional extra review; record RED before production edits (root coordinates browser build/run).
- [ ] Consume confirmed goal and session status on Home/player/completion, display tick and move daily-complete primary action to discovery/practice. Do not calculate effort from client answers or force due dates tomorrow.
- [ ] Render all3question kinds accessibly: four pictures with all-text fallback, four text definitions, gap with four word choices; no target reveal for gaps, stable Check/Continue and recovery.
- [ ] Remove exercise attribution, load canonical other meanings and exact-revision picture only after checked correct/wrong answer; test optional-read failure, stale responses, single meaning and reload. Keep Continue available and one feedback announcement.
- [ ] Add timezone-aware relative schedule helper and tests covering tomorrow/weeks/months/DST; retain exact date as accessible detail and existing persisted interval choices.
- [ ] Verify lint/types/helper tests and coordinated browser matrix360/430/desktop, light/dark/keyboard/200%text. Root commits accepted changes after review.

### Task 3: Discovery navigation, range previews and account resume

**Owner/files:** discovery implementer: `apps/web/src/app/(app)/discovery/**`, reusable map mini-grid extracted from `knowledge-map/map-view.tsx`, new `apps/api/business/discoveryposition/**`, `apps/api/app/api/discovery_position.go` plus tests, migration `20261115110000_discovery_position.sql`, relevant export/deletion integration additions. New browser test `discovery-owner-feedback.spec.ts`. No production/openapi registration/API-client/shared mock edits; send root exact wiring patch.

**Interfaces:** GET/PUT `/api/v1/discovery-position`, requester scoped and no-store; PUT CSRF protected, body `{catalogVersion,band,rank,active,expectedRevision}`; response `{position: null | {catalogVersion,band,rank,active,revision}}`; stale revision409. Position is optional preference and never learning credit. Root adds client methods and route registration.

- [ ] Add RED regressions for4330→Home→genericdiscovery, explicit-link precedence, account isolation, known-word exact revisit and stale preference writes.
- [ ] Move Previous word below card, keep range/word selection above; preserve locks and cross-band behavior.
- [ ] Extract self-contained miniature100word visualization, add readable range counts, one selectable row; test legacy fallback and state agreement.
- [ ] Implement validated account position storage/read/CAS update and export/deletion coverage; preserve explicit deep links, fresh server state, optional persistence failure and Back/Forward.
- [ ] Run focused Go tests and helper/type/lint checks; coordinate browser evidence with root rather than start conflicting servers. Root commits accepted changes after review.

### Task 4: Integration, authenticated root and shared contracts

**Owner/files:** root: `apps/web/src/middleware.ts`, its tests, `packages/api-client/src/index.ts`/tests, `apps/api/app/api/{production,openapi}.go`, generated API contracts, shared mock fixture and combined browser/staging tests, Help source link as needed.

- [ ] Reproduce authenticated-root marketing behavior with middleware test; implement root identity routing preserving anonymous page, onboarding, expired cookies and recoverable upstream outage. Run middleware regressions.
- [ ] Apply agent-reported interfaces to API client/routes/generated OpenAPI; extend tests for additional flag/position/goal/choice contract and backwards compatibility.
- [ ] Update shared synthetic fixtures for new server semantics without weakening actual assertions; add stricter real staging checks for effort-based target and mixed-mode availability where actual approved content exists.
- [ ] Coordinate frozen install/build, disposable PostgreSQL full API suite, client/web unit tests, lint/typecheck/e2e types and relevant whole browser matrix, screenshot/axe/keyboard review. Record exact pass/fail/skip/retry counts and limitations.

### Task 5: Independent review, one PR and release

**Owner/files:** root/reviewers: full branch diff, product/design acceptance docs, private handoff; workflow configuration remains unchanged unless evidenced necessary.

- [ ] Independently review each task's spec and code; fix findings with targeted RED/GREEN evidence. Review full branch for auth/account isolation, goal/streak/rewards, question content/answer leaks and media failure modes.
- [ ] Update spec status and product implementation/acceptance map with actual results. Commit local changes in logical slices on one branch; push one draft/hold PR targetingmain and request short Claude review.
- [ ] Inspect exact-head applicable CI/review results, including actual browser/Lighthouse reports and retries. Ready/unhold only after accepted results; normal merge queue, verify merged tree equals accepted head.
- [ ] Strict staging all actual4cases with no failures/skips/retries, exact merged version/SHA/health; then dispatch canonical production workflow once with origin-cutover=skip.
- [ ] Verify actual production smoke30PASS0FAIL0SKIP separately from existing owner-waivedINS11findings; independently verify4version endpoints/2APIhealth/pairedtimestamps. Update PR/handoff, stop owned servers and report deployed links/version with limitations.
