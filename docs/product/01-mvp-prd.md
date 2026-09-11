---
id: DOC-01
title: VocaNova MVP PRD
version: 1.0
document_type: product-requirements
status: approved
owner: founder
canonical_path: docs/product/01-mvp-prd.md
approved_at: 2026-07-21
last_reviewed_at: 2026-07-21
review_cycle: monthly
supersedes: null
related_documents:
  - DOC-00
  - DOC-03
  - DOC-08
  - DOC-09
  - DOC-12
related_decisions: []
adoption_change: VOC-008
source_files:
  - path: 01-product-bible-and-prd.md
    sha256: ffafedf6bb6e1ff6c7e04f8ce67c23478592dd099a543a648d970bf5733f8009
---
# 01 — VocaNova MVP PRD

## 1. Product baseline

The product vision, target learner, learning loop, gamification, spaced repetition, and AI-feedback
principles are defined in [DOC-00](00-product-bible.md). This PRD defines the bounded MVP surface and
completion criteria.

## 2. MVP core screens (3-tab navigation)

- **Home** — Today's Mission, streak, due-review count, quick entry to discovery and sentence
  practice.
- **Journey** — situation-based discovery (Airport, Restaurant, Hotel Check-in, Job Interview, Daily
  Conversation, Work Meeting, University Class, etc.), word detail, save/unsave.
- **Progress** — Confidence Points, streak, completion history, motivation-focused summary.

Sentence practice is **not** a fourth tab — it's a reusable component surfaced from Home, Word
Detail, and Review Completion. See [03](../design/03-ui-ux-design.md).

The original MVP retained sentence history without a dedicated screen (see
[the migration notes](../archive/README-migration-notes.md#4-sentence-history-screen-conflict)).
The founder-authorized [maturity delivery](mature-learning-and-account-experience.md)
extends Progress with `/progress/sentences`: a paginated record of original sentences
and their feedback. Sentence practice remains contextual, and the app retains three
primary tabs. History records practice; it does not claim measured fluency or mastery.

## 3. MVP completion criteria

The MVP is done when an authenticated A2–B1 learner can, on a responsive mobile-first web app:

1. discover practical vocabulary by situation;
2. view useful word information (meaning, part of speech, examples, usage notes);
3. save and remove saved words;
4. review due words through the spaced-repetition loop;
5. write a sentence using a selected word and receive AI feedback;
6. complete a daily mission and see it reflected in progress/streak;
7. see a meaningful, backend-authoritative progress summary;
8. do all of the above through Google OAuth or email magic-link authentication, with no password
   login in MVP.

The subsequent [password, profile and appearance delivery](password-profile-and-theme.md)
extends authentication with verified email/password registration and recovery,
and adds editable profile details and Light/Dark/System preferences. The original
MVP exclusion above describes the initial release, not a permanent restriction.

## 4. Explicit MVP exclusions

Native mobile app (React Native/Expo — architected for, not built), leaderboards, badges, social
challenges, rewards store, subscriptions/monetization, teacher dashboards, multi-provider AI
routing, model fine-tuning, complex microservices, message queues without a proven need. Full list
in [DOC-12](12-mvp-implementation-plan.md) §10.
