---
id: DOC-03
title: VocaNova UI/UX Design
version: 1.0
document_type: ui-ux-design
status: approved
owner: founder
canonical_path: docs/design/03-ui-ux-design.md
approved_at: 2026-07-21
last_reviewed_at: 2026-07-21
review_cycle: quarterly
supersedes: null
related_documents:
  - DOC-00
  - DOC-01
  - DOC-08
  - DOC-09
related_decisions: []
adoption_change: VOC-008
source_files:
  - path: 03-ui-ux-design.md
    sha256: f3f37beea86bc29a5230f66731730ab28a07635546d60084e49f954e53b30ed4
---

# 03 — VocaNova UI/UX Design

## 1. UX purpose and principles

Vocanova's UX exists to make one thing effortless: turning a short daily session into visible
progress toward practically usable English. Guiding principles carried through every screen:

- **One clear action per screen.** Avoid competing calls to action; the learner should always know
  the single next thing to do.
- **Practical over academic.** Situations and real usage over abstract grammar taxonomy.
- **Encouraging, not gamified for its own sake.** Confidence Points and streaks support the habit;
  they are not the point of the product.
- **Mobile-first, responsive, not native.** Design for 360–430px viewports first; desktop is a
  wider layout of the same experience, not a separate design.
- **Backend is authoritative.** The frontend never invents progress, mission completion, or
  scheduling state — it only reflects what the backend confirms (this recurs throughout
  [07](../engineering/07-api-contract-and-dto-design.md) and [09](../engineering/09-ai-features.md) as a hard rule, not just a UX
  preference).

## 2. Information architecture and navigation

Three primary destinations, presented in bottom navigation on mobile and adapted
to the desktop workspace (see [the visual direction](learning-workspace.md)):

1. **Home** — Today's Mission, streak, due-review entry, quick sentence-practice entry, discovery
   teaser.
2. **Journey** — situation-based discovery and saved-word management.
3. **Progress** — Confidence Points, streak history, completion summary.

Sentence practice is a **reusable component**, not a fourth tab or standalone route — it is invoked
from Home, Word Detail, and Review Completion. This is a deliberate MVP scoping decision: it keeps
the tab bar simple and treats "write a sentence" as something the learner does _in the middle of_
another activity, not as its own destination.

The [maturity delivery](../product/mature-learning-and-account-experience.md) adds
Sentence History under Progress at `/progress/sentences`. It presents the retained
`learner_sentences` / `ai_feedback_attempts` data (see
[05](../engineering/05-database-design.md) §11), including processing failures and
pending feedback without presenting them as learning outcomes. The original MVP
deferred this screen; its [migration rationale](../archive/README-migration-notes.md#4-sentence-history-screen-conflict)
remains historical context. History is a secondary destination, not a fourth tab.

## 3. Onboarding flow

Short, low-friction sequence: sign-in (Google OAuth, verified email/password, or configured email magic link) → a handful of
onboarding questions (English level, native language, learning goal, main use case, daily review
target) → straight into the first Today's Mission. Onboarding answers populate
`user_onboarding_profiles` (see [05](../engineering/05-database-design.md) §6) and are used to seed sensible defaults,
not to gate access behind a long survey.

## 4. Daily session flow (the core loop, screen by screen)

1. **Home** shows the day's mission (review target, progress toward it), current streak, and a
   route into Journey/Discover if the learner wants new words.
2. **Review** is a focused, single-item-at-a-time session: show the prompt, learner responds
   (multiple-choice, self-check, or typed depending on prompt type — see
   [05](../engineering/05-database-design.md) §9), reveal correctness, then record a rating. Objective
   incorrect answers record `Again`; objective correct answers allow Hard/Good/Easy. Self-check
   prompts derive correctness from the learner's selected rating. Then advance. No competing UI
   during an active review item.
3. **Review completion** surfaces a summary and offers sentence practice as an optional next step
   using one of the words just reviewed.
4. **Sentence practice** (reusable component): shows the target word/meaning, an input box, a short
   instruction ("Write one sentence using '{targetWord}'"), and a privacy reminder not to include
   personal information. Submitting shows a calm pending state (input preserved, no duplicate
   submission possible), then the AI result: status, corrected sentence when needed, short
   explanation, one improvement tip, and a visible "AI can make mistakes" disclaimer plus a
   report-feedback action. See [09](../engineering/09-ai-features.md) for exact field-level behavior — this is the UI
   shell around that contract, not a re-specification of it.
5. **Progress** shows Confidence Points total, streak, and a simple day-by-day or week-by-week
   completion view — motivational, not analytical.

## 5. Journey / Discover UX

Discovery is organized by real-life situation (Airport, Restaurant, Hotel Check-in, Job Interview,
Daily Conversation, Work Meeting, University Class, etc. — see [05](../engineering/05-database-design.md) §8 for the
full `journey_situations` model), not by grammar topic or difficulty tier alone. Within a situation,
words are shown one at a time or as a short scannable list; the backend controls ordering
(core words first, then display order, then relevance — see [05](../engineering/05-database-design.md) §8). A word
already in the learner's saved list is visually marked and excluded from "new" recommendations.
Saving must succeed against the backend before the UI reflects it as saved — no optimistic-only
save state that could desync from the backend.

## 6. Word Detail UX

Shows the canonical word/phrase, its meaning(s), part of speech, example sentences, and usage notes
(collocation, register, common mistakes — see [05](../engineering/05-database-design.md) §7). Includes a save/unsave
control and an entry point into sentence practice for that specific word (see
[05](../engineering/05-database-design.md) §7). If the learner has
already saved the word, shows their current review state (e.g. "due today," "learning," "mastered")
rather than treating Word Detail as purely a content-browsing page.

## 7. Review UX detail

Ratings surface as **Again / Hard / Good / Easy** (see
[the migration notes](../archive/README-migration-notes.md#2-review-rating-and-scheduling-conflict) for how this was reconciled
against three other rating-scale drafts found in earlier documents). The review screen must remain
usable one-handed on a phone: large touch targets (44px minimum per [08](08-web-app-design.md)), no
required typing unless the prompt type is explicitly typing/sentence-usage.

## 8. Progress, Streak, Confidence Points, and celebration moments

Progress screen is intentionally simple: Confidence Point total, current/longest streak, and a
completion history view. Celebration moments (small positive feedback on mission completion,
streak milestones) should be lightweight — a brief animation or message, not a full-screen
interruption — consistent with "encouraging, not gamified for its own sake" from §1.

## 9. Empty, loading, and error states

Required for every screen with dynamic content:

- **Empty** — first-time/no-history states must explain what will appear here and how to get there
  (e.g. empty Progress screen invites the learner to complete their first mission), not just show a
  blank area.
- **Loading** — calm, non-jarring; sentence-feedback pending state specifically must preserve the
  learner's typed input and disable duplicate submission (see [09](../engineering/09-ai-features.md) §5).
  **Error/retry** — every network-dependent screen needs a safe retry path that doesn't lose learner
  input or falsely imply something completed. AI feedback failures specifically must never claim
  mission completion (see [09](../engineering/09-ai-features.md) §5 and §8).

## 10. Accessibility

Target: WCAG 2.2 AA. Keyboard operability, visible focus states, screen-reader-friendly labels on
forms and icon-only controls, sufficient color contrast, and no information conveyed by color alone
(e.g. correct/incorrect review feedback must also use an icon or text label, not just a color
change). Full testing requirements are in [10](../operations/10-development-workflow.md).

## 11. Visual design direction

Clean, calm, encouraging visual tone — not exam-like, not childish. Tailwind CSS with shadcn/ui-style
components (see [08](08-web-app-design.md) for the exact frontend stack). Avoid visual patterns that
read as "grading" (red X marks, harsh error colors) in favor of supportive framing, consistent with
the AI tone rules in [09](../engineering/09-ai-features.md) §13 ("Great use of this word," "Almost right," never
"Your English is bad").

The product maturity pass uses one primary-blue/secondary-violet/neutral token palette across
public and authenticated pages. Page content is bounded on desktop and padded for phones;
white rounded surfaces, consistent focus indicators, and 44px actions recur across the app.
Small text uses a minimum 12px token (14px for supporting copy). Mission actions remain visible
without scrolling on a 360×640 phone. Brand presentation and form treatments should remain
consistent through sign-in, onboarding, learning, and account management.

## 12. UX risks and mitigations

- **Risk: AI feedback UI reads as a test/grade, discouraging learners.** Mitigation: encouraging
  copy rules (§11 above and [09](../engineering/09-ai-features.md) §13), always show the original sentence, avoid
  inventing corrections for already-correct sentences.
- **Risk: Review sessions feel like a chore.** Mitigation: keep sessions short by default (backend
  daily target, not an open-ended queue), show visible progress within the session.
- **Risk: Mobile one-handed usability failures.** Mitigation: 44px minimum touch targets, bottom
  navigation reachable by thumb, no required multi-step gestures.
- **Risk: Learner distrust of AI feedback correctness.** Mitigation: visible "AI can make mistakes"
  disclaimer and a report-feedback action on every result (see [09](../engineering/09-ai-features.md) §16).

## 13. Final MVP UX summary

The MVP UX is deliberately narrow: three tabs, one focused daily mission, one reusable sentence-
practice component reachable from three entry points, and a simple progress view. Nothing in this
document introduces a screen or flow beyond what [DOC-01](../product/01-mvp-prd.md) §2 and §3 already
scope as MVP-complete.
