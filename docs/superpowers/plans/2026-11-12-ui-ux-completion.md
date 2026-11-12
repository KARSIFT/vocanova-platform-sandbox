# Remaining UI/UX findings implementation plan

> **For agentic workers:** Follow repository AGENTS.md and the task ownership below. Root integrates and verifies; independent agents implement bounded areas and review the combined result.

**Goal:** Complete the actionable findings from the 12 November audit and prepare an honest, usable A2–B1 learner study.

**Architecture:** Extend existing recovery and draft mechanisms. Use native disclosures and the current token system. Preserve canonical vocabulary identities and expose reviewed learner explanations without breaking media bindings.

**Tech stack:** Existing Next.js/React, Go/PostgreSQL, pinned pnpm/Playwright. No new dependencies or authentication provider.

**Spec:** `docs/product/ui-ux-audit-2026-11-12.md`, remaining findings; owner instruction “ok do all of them”.

## Global constraints

- One delivery branch and one consolidated draft PR with hold. Existing checkout, no new repository/worktree.
- Three primary destinations, existing tokens, server-confirmed progress and idempotent mutations remain.
- Drafts are validated, same-user, tab-scoped and expire after two hours. Never persist passwords, confirmation tokens or cookies.
- Canonical definition revisions bind approved pictures/audio; preserve those identities. Learner copy must retain the exact sense and part of speech, with no invented CEFR certification.
- Tests use 3100/8180; never reserve 3000/8080 or stop Vergecloud. Browser audio and speech are muted.
- Public email responses must not reveal account existence; existing cooldowns and disabled capabilities remain enforced.
- Real learner observations are required for usability conclusions; the owner will arrange participants. No synthetic usability score.

## Review focus

- A stale successful response must not erase a newer draft, and changing users/logging out must remove private drafts.
- Resend failures keep recovery available, parallel clicks issue one request, and a different address clears transient secrets.
- Supplementary Unicode names use one code-point rule in signup, Profile, Settings and API-facing validation.
- Collapsed management errors must be visible/focused; pending and failed writes must not imply confirmed deletion/saving.
- Learner explanations must not cross senses or lose mappings; empty topic success must differ from an API error.

## Task 1 — Account recovery and consistent names

**Owner:** account_recovery. **Files:** password forms, Profile/email-change components and user-ID props, preferences-draft helper/cleanup callers, name validation and dedicated unit/browser tests.

- [ ] Add regression tests for resend/cooldown/parallel locking/change-address focus and observe failures on the baseline bundle.
- [ ] Extend the established magic-link recovery pattern to password signup/reset success states; keep generic messaging and passwords only in React memory.
- [ ] Add Profile/email drafts to the existing two-hour tab storage with user isolation, cleanup and snapshot-based confirmed-save clearing.
- [ ] Unify name code-point validation and test 80/81 supplementary Unicode characters.
- [ ] Verify unit suites and browser recovery, expired authentication, account boundaries, keyboard, 200% text, both themes and supported widths.

## Task 2 — Learning hierarchy and topic empty state

**Owner:** learning_hierarchy. **Files:** list collection/detail, story library and writing pages; own browser tests and existing affected assertions; API zero-situations test if needed.

- [ ] Reproduce missing empty state and competing primary actions with browser regressions.
- [ ] Keep learning content first; place create/rename/delete controls in accessible native disclosures, opening necessary controls for empty/error states.
- [ ] Give each story one primary continuation/start/repeat action; retain an optional details destination for in-progress stories.
- [ ] Render a successful zero-topic response with useful saved-word and Practice alternatives, distinct from loading/error states.
- [ ] Verify story states, list confirmed writes/recovery, empty-topic API contract and responsive keyboard/theme behavior.

## Task 3 — Learner explanations and catalog audit

**Owner:** content_readability; root independently reviews authored copy. Exact files/interfaces recorded in the task report before implementation.

- [ ] Audit all ranked canonical definitions for explicit readability/polysemy indicators; distinguish heuristic flags from learner comprehension.
- [ ] Trace the existing learner_definition field and establish revision-safe delivery across affected surfaces.
- [ ] Add failing identity/fallback/sense and representative hard-definition tests.
- [ ] Implement reviewed learner explanations using existing editorial fields/overlay infrastructure; retain source definitions and media identity.
- [ ] Verify catalog invariants, API/schema contracts, selected-sense navigation and readable definitions on learner-facing surfaces. Record exact reviewed coverage and any editorial work that cannot honestly be called complete.

## Task 4 — Real learner study and integrated acceptance

**Owner:** root. **Files:** product study guide and observation template, audit/current-state/release-readiness updates, private handoff.

- [ ] Prepare recruitment/consent guidance, neutral tasks, moderator prompts, observation sheet and decision criteria for 3–5 A2–B1 learners. Owner arranges sessions.
- [ ] Cover second meaning, story continuation, interrupted saving, progress comprehension and revised list/topic/recovery flows.
- [ ] Run integrated unit/API checks and targeted cross-engine browser matrix; inspect visual screenshots and accessibility of affected states.
- [ ] Obtain independent source/auth/data and editorial review; resolve consequential findings and verify combined diff.
- [ ] Open one draft/hold PR, request repository review and inspect actual CI. Keep real-user/provider/database evidence distinct from mocked browser results.
- [ ] After required CI and review pass, use the normal merge queue and strict automatic staging acceptance, then the separately dispatched production workflow, following the owner's existing merge/deploy instruction for this UI/UX work. Verify exact release identity and actual smoke outcomes.
- [ ] Update the handoff with changed files, exact checks, remaining evidence and next action. Never claim learner study execution before sessions happen.
