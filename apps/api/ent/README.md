# Ent schemas

VOC-025-T00 defines the `users`, `external_identities`, `sessions`, and
`magic_links` identity schemas.

VOC-026-T00 adds the P1 canonical content and learner-owned schemas:
`canonical_words`, `word_meanings`, `word_examples`, `usage_notes`,
`journey_situations`, `journey_words`, and `user_words`.

Versioned Atlas SQL remains the migration authority; schema creation is never run
by API startup.
