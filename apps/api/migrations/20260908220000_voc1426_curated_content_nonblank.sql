-- atlas:txmode file
-- Issue #1426: required platform-owned learning content must contain at least
-- one non-whitespace character. NOT VALID preserves any legacy draft/content
-- rows while protecting all new inserts and updates.

ALTER TABLE canonical_words
  ADD CONSTRAINT canonical_words_text_nonblank
    CHECK (text ~ '[^[:space:]]') NOT VALID,
  ADD CONSTRAINT canonical_words_normalized_text_nonblank
    CHECK (normalized_text ~ '[^[:space:]]') NOT VALID;

ALTER TABLE word_meanings
  ADD CONSTRAINT word_meanings_short_definition_nonblank
    CHECK (short_definition ~ '[^[:space:]]') NOT VALID;

ALTER TABLE word_examples
  ADD CONSTRAINT word_examples_example_text_nonblank
    CHECK (example_text ~ '[^[:space:]]') NOT VALID;

ALTER TABLE usage_notes
  ADD CONSTRAINT usage_notes_note_text_nonblank
    CHECK (note_text ~ '[^[:space:]]') NOT VALID;

ALTER TABLE journey_situations
  ADD CONSTRAINT journey_situations_slug_nonblank
    CHECK (slug ~ '[^[:space:]]') NOT VALID,
  ADD CONSTRAINT journey_situations_title_nonblank
    CHECK (title ~ '[^[:space:]]') NOT VALID,
  ADD CONSTRAINT journey_situations_short_description_nonblank
    CHECK (short_description ~ '[^[:space:]]') NOT VALID;
