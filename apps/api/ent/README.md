# Ent schemas

VOC-025-T00 defines the `users`, `external_identities`, `sessions`, and
`magic_links` identity schemas.

VOC-026-T00 adds the P1 canonical content and learner-owned schemas:
`canonical_words`, `word_meanings`, `word_examples`, `usage_notes`,
`journey_situations`, `journey_words`, and `user_words`.

VOC-027-T00 adds the P2 `review_attempts` immutable-history table.

VOC-028-T00 adds the P3 AI-feedback tables: `learner_sentences` (learner-
generated original sentences, soft-deleted) and `ai_feedback_attempts`
(immutable feedback-generation history).

Versioned Atlas SQL remains the migration authority; schema creation is never run
by API startup.
