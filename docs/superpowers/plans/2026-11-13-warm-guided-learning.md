# Warm guided learning implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver the approved Warm guided coach experience across the existing first learning journey, with clearer next steps and exact-meaning writing access.

**Architecture:** Reuse current server-selected activities, confirmed mutations and saved-word routes. Introduce small presentational learning components and token-based styles, then apply them to Home, Discovery, real review players and sentence writing. No combined session, API contract, dependency, database, auth-policy, analytics or scheduling change.

**Tech Stack:** Next.js 16 / React 19, existing generated tokens/Tailwind, TypeScript and pinned Playwright. Local web3100/API8180; existing Node24.18/pnpm11.14 tooling.

**Spec:** `docs/superpowers/specs/2026-11-13-first-five-minutes-design.md` — owner approved Warm guided coach and this proposal on13 November2026.

## Global Constraints

- Preserve Home/Journey/Progress and existing blue identity, serif headwords and two themes.
- Home's primary action remains backend-selected; targets do not manufacture due words.
- Only confirmed exact-meaning outcomes can say Saved. Preserve pending identity, replay, revision and auth recovery.
- Review uses the real queue, current four ratings, grading and schedule. No same-day repeat after assessment or invented mastery/rewards.
- Sentence writing uses the exact saved `userWordId` and sense, existing drafts, provider feedback and safe retries.
- No automatic audio or compulsory animation. No user account writes during local verification.
- One delivery branch and one consolidated draft/hold PR, then normal queue/strict staging/manual production under the existing authorization.
- Learner sessions remain0 until observations actually occur; prototype measurements are not learner evidence.

## Review Focus

1. Returning learner with an active review and no new words: the active review stays primary.
2. Lost save receipt followed by retry or a range change: no invented confirmation, duplicate save or mismatched writing link.
3. A secondary meaning with an existing saved record: its writing destination opens that exact meaning and draft.
4. Provider delay, failure or session expiry while editing: current draft and request identity survive; no fake success or cross-account recovery.
5. Incorrect review, completion and a no-due return: truthful next-step copy preserves the actual schedule and never promises mastery or a fixed next-day workload.

## Shared interfaces and verification coordination

Task1 supplies `apps/web/src/ui/learning-guide.tsx`:

```ts
LearningIntro({title, eyebrow?, description?, headingLevel?, id?, className?}: {
  title: ReactNode; eyebrow?: string; description?: ReactNode;
  headingLevel?: 1 | 2 | 3; id?: string; className?: string;
}): ReactNode
LearningNote({children, className?}: {children: ReactNode; className?: string}): ReactNode
```

Both are presentational, usable by existing server/client components without new
client state. Root owns integration, builds, server lifecycle, whole-branch tests,
reviews, GitHub operations and releases. Workers own only listed source/tests,
never shared build output. Each worker proves new behavior RED against the current
baseline bundle; root rebuilds once integrated source is ready, then runs GREEN.
Independent task and whole-branch reviews inspect actual diffs and evidence.

### Task 1: Shared warm guidance and Home

**Files:** create `apps/web/src/ui/learning-guide.tsx`; modify `apps/web/src/app/globals.css`, `apps/web/src/app/(app)/home/page.tsx`; create `apps/web/tests/e2e/warm-guide-home.spec.ts`.

**Consumes:** unchanged current daily lesson, mission and saved-word responses.
**Produces:** the shared components above; token-based `.learning-*` presentation; coherent Home states.

- [ ] Add meaningful browser regressions for a warm reading surface, one filled next action, concise new-word guidance, active-review priority and no-due honesty. Assert actual href/start behavior and computed foreground/background, not a class-name mirror.
- [ ] Run against baseline at360/430/1280 in both themes; record the intended failures.
- [ ] Implement shared components and style the primary Home panel as a calm surface with blue action and clear reading order. Preserve all server reads, Suspense isolation, mission numbers and returning actions. Show only genuine saved meaning or explicitly labelled public example; no fabricated personal progress.
- [ ] Root rebuilds; run the new tests and existing Home/navigation/hierarchy suites.
- [ ] Review and commit only owned files after passing checks.

### Task 2: Discovery with a truthful bridge to using a saved meaning

**Files:** modify `apps/web/src/app/(app)/discovery/discovery.tsx` and narrowly relevant Discovery presentation/helpers; create `apps/web/tests/e2e/warm-guide-discovery.spec.ts`. Do not edit the shared knowledge-map stylesheet without root coordination.

**Consumes:** Task1 components; existing confirmed decision/fresh-read flow.
**Produces:** clearer entry/teaching hierarchy and a confirmed, exact-meaning next-step affordance that leaves current navigation intact.

- [ ] Add browser regressions for confirmed Learn, pending/lost receipt, Known and changed cursor/range. A writing affordance must target the confirmed meaning, never a guessed saved record or current next word.
- [ ] Run RED on baseline; inspect actual DTOs and existing writing deep-link helpers before selecting a destination.
- [ ] Apply shared guidance. Add an optional bridge only when the existing response/read proves the save; preserve advance, cursor, request keys, revisions, reconciliation and lock behavior. Use an existing exact-meaning route/anchor, avoiding new API requests/contracts where possible. Do not force writing or stop discovery.
- [ ] Root rebuilds; run new cases plus Discovery navigation/cursor/concurrency/word-detail tests.
- [ ] Review and commit only owned files after passing checks.

### Task 3: Reachable sentence writing and readable real feedback

**Files:** modify `apps/web/src/app/(app)/_components/word-detail-content.tsx`, `sentence-feedback.tsx`, existing saved-word presentation where necessary; create `apps/web/tests/e2e/warm-guide-writing.spec.ts`.

**Consumes:** Task1 components; current exact saved meaning, draft and feedback interfaces.
**Produces:** direct visible writing entry for a saved meaning, coherent input/feedback actions and concise next-step guidance.

- [ ] Add regressions for a secondary saved meaning, existing `#sentence-practice` deep link, retained draft after revision/failure, and real fixture success versus error. No arbitrary text may receive prototype sample feedback.
- [ ] Run RED against baseline and inspect disclosure/focus/native navigation behavior.
- [ ] Promote writing within each saved meaning's hierarchy while keeping personal organization optional. Preserve exact IDs, labels, privacy guidance, accepted edits, request keys, session handling, feedback source and report controls. Shared primary/secondary actions must not create competing filled buttons on the same focused task.
- [ ] Root rebuilds; run new cases and word-detail/writing/feedback/draft/session recovery suites.
- [ ] Review and commit only owned files after passing checks.

### Task 4: Real review and completion continuity

**Files:** modify `apps/web/src/app/(app)/reviews/_components/daily-lesson.tsx`, `review-session.tsx` and presentation-only completion views as needed; create `apps/web/tests/e2e/warm-guide-review.spec.ts`.

**Consumes:** Task1 components; unchanged daily/legacy review state machines and returned schedules.
**Produces:** readable review instructions, consistent controls, honest feedback/completion and existing exact-meaning sentence entry.

- [ ] Add regressions for ready/active/empty/completed states, incorrect outcome and actual next-review display. Verify original mutation bodies, original result identity and next action.
- [ ] Run RED against baseline; inspect both current daily player and compatibility review surfaces.
- [ ] Apply the approved presentation without changing reducer/effects, grading, same-day rules, action locks or accuracy. No fixed tomorrow/reward/mastery promise. Keep any target-word sentence link exact.
- [ ] Root rebuilds; run new and existing daily/legacy review/session/scheduling/hierarchy tests.
- [ ] Review and commit only owned files after passing checks.

### Task 5: Integrated evidence, study handoff and release

**Files:** update product/design/readiness docs and private handoff; retain protocol/blank sheet; no private screenshots or recordings in public commits.

- [ ] Build final source with local3100/8180 configuration; run types, lint, units and affected browser suites, then full hosted checks on the exact PR head.
- [ ] Check mobile/desktop, keyboard, both themes, 200% text, empty/error/pending/completed states. Capture and independently inspect real screenshots against the selected prototype. Keep prototype and production evidence labelled separately.
- [ ] Independently review exact mutation/auth/data/schedule boundaries and actual diffs, then disposition requested Claude review and all findings.
- [ ] Open one consolidated draft/hold PR against main once integration is ready; qualify any baseline/tooling limitations. Release only after applicable CI and normal merge queue pass.
- [ ] Strict staging must execute core/maturity/provider/export/CDN gates. Promote exact accepted commit with the manual production workflow; verify actual smoke and all four live version/two health endpoints.
- [ ] Update handoff and report delivered behavior, screenshots, exact release and remaining learner-study limits. Do not claim the whole product is finished.
