---
id: DOC-00
title: VocaNova Product Bible
version: 1.0
document_type: product-bible
status: approved
owner: founder
canonical_path: docs/product/00-product-bible.md
approved_at: 2026-07-21
last_reviewed_at: 2026-07-21
review_cycle: semiannual
supersedes: null
related_documents:
  - DOC-01
  - DOC-02
  - DOC-03
  - DOC-05
  - DOC-09
  - DOC-12
related_decisions: []
adoption_change: VOC-008
source_files:
  - path: 01-product-bible-and-prd.md
    sha256: ffafedf6bb6e1ff6c7e04f8ce67c23478592dd099a543a648d970bf5733f8009
---
# 00 — VocaNova Product Bible

## 1. What Vocanova is

Vocanova is an AI-powered practical English vocabulary learning platform for global A2–B1 learners,
delivered as a responsive, mobile-first web application (not a native app in MVP).

Core learning loop:

1. Discover practical words and phrases, organized by real-life situation (airport, restaurant,
   job interview, work meeting, daily conversation, etc.), not by abstract grammar topic.
2. Save words into personal vocabulary.
3. Review saved words on a spaced-repetition schedule.
4. Write an original sentence using a saved word.
5. Receive concise, encouraging, accurate AI feedback on that sentence.
6. Build a daily habit through a "Today's Mission," Confidence Points, and a gentle streak.

The product's differentiator versus Duolingo-style apps: the learner **produces** language (writes
their own sentence) instead of only selecting or translating a predefined answer, and feedback is
narrowly focused on correct use of the one target word, not general essay grading.

## 2. Target learner

- CEFR A2–B1, global, English as a second language.
- Motivated by practical, immediate use (travel, work, daily life, conversation) rather than exam
  prep or academic study, though "exam" and "study" remain supported onboarding goals.
- Mobile-first usage pattern: short daily sessions, not long study blocks.

## 3. Gamification model

- **Confidence Points** — the product's point/reward currency, earned for reviews, word additions,
  daily-mission completion, and sentence submissions. Source of truth is an append-only ledger
  (`confidence_point_ledger`), not a mutable balance field — see [05](../engineering/05-database-design.md) §12.
- **Streak** — advances only after a full daily mission is completed, uses the learner's local
  timezone, and has a gentle reset (grace days) rather than a hard break. A grace day is earned
  every 7 completed days, capped at a balance of 2.
- **Daily Mission ("Today's Mission")** — a stable, timezone-aware daily snapshot of a review target,
  optional new-word target, and optional sentence-practice bonus. Settings changes (e.g. changing
  daily review target) apply from the next local day, not retroactively.

## 4. Spaced repetition (high level)

MVP uses deterministic, step-based scheduling (steps 0–7), not a probabilistic algorithm like FSRS —
that's an explicit future swap point behind a stable scheduling interface, not an MVP requirement.
Learner-facing rating scale, exact schedule mechanics, and the reset rule are defined precisely in
[05](../engineering/05-database-design.md) §9, and [08](../design/08-web-app-design.md). See
[the migration notes](../archive/README-migration-notes.md#2-review-rating-and-scheduling-conflict) for how the four
different rating-scale drafts across the source docs were reconciled — the two-word summary is:
**"Again / Hard / Good / Easy" ratings, `review_step` 0–7, two consecutive incorrect answers reset to
step 0.**

## 5. AI sentence feedback (high level)

One learner-facing AI capability in MVP: evaluate one learner-written sentence using a selected
target word/phrase and return a status of `correct`, `needs_improvement`, or `incorrect`, plus a
short explanation and (when needed) a corrected sentence. Full behavior, safety rules, prompt
architecture, and evaluation thresholds are in [09](../engineering/09-ai-features.md) — that document is the single
source of truth for AI behavior; don't restate feedback-label wording elsewhere (see
[the migration notes](../archive/README-migration-notes.md#1-ai-feedback-label-conflict)).

Explicit MVP non-goals: open-ended AI chat, general AI tutor, essay correction, pronunciation
scoring, speech recognition, roleplay, AI-generated vocabulary as the authoritative content source,
user-selectable models. Full non-goal list in [09](../engineering/09-ai-features.md) §4.

## 6. Product authority boundary

The delivery participants are the founder/product owner, ChatGPT as planning and architecture
advisor, Codex as implementation worker, Claude Code as independent verifier, GitHub as repository
system of record, and GitHub Actions as deterministic automation.

The product owner sets product scope. Repository delivery follows [AGENTS.md](../../AGENTS.md):
one PR against `main`, required CI checks, and automated review. The former multi-role
governance pipeline is retired. Historical change packages explain earlier implementation
decisions but do not require recreating that process.
