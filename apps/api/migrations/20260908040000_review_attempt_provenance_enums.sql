-- atlas:txmode file
-- Issue #1410: DOC-05 §9 defines the closed domains for immutable review
-- history provenance. The original P2 table only required non-empty text,
-- which allowed direct/import writes that no documented review flow recognizes.

ALTER TABLE review_attempts
  ADD CONSTRAINT review_attempts_attempt_type_documented
    CHECK (attempt_type IN ('review', 'practice', 'placement', 'mission')) NOT VALID,
  ADD CONSTRAINT review_attempts_source_documented
    CHECK (source IN ('daily_review', 'word_detail', 'journey_practice', 'manual_practice')) NOT VALID;
