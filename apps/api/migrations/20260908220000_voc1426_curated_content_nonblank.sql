-- atlas:txmode file
-- Issue #1426: required platform-owned learning content must contain at least
-- one non-whitespace character. The explicit Unicode White_Space set avoids
-- the collation-dependent semantics of PostgreSQL's [:space:] regex class.
-- NOT VALID preserves any legacy draft/content rows while protecting all new
-- inserts and updates.

ALTER TABLE canonical_words
  ADD CONSTRAINT canonical_words_text_nonblank
    CHECK (text ~ U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]') NOT VALID,
  ADD CONSTRAINT canonical_words_normalized_text_nonblank
    CHECK (normalized_text ~ U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]') NOT VALID;

ALTER TABLE word_meanings
  ADD CONSTRAINT word_meanings_short_definition_nonblank
    CHECK (short_definition ~ U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]') NOT VALID;

ALTER TABLE word_examples
  ADD CONSTRAINT word_examples_example_text_nonblank
    CHECK (example_text ~ U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]') NOT VALID;

ALTER TABLE usage_notes
  ADD CONSTRAINT usage_notes_note_text_nonblank
    CHECK (note_text ~ U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]') NOT VALID;

ALTER TABLE journey_situations
  ADD CONSTRAINT journey_situations_slug_nonblank
    CHECK (slug ~ U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]') NOT VALID,
  ADD CONSTRAINT journey_situations_title_nonblank
    CHECK (title ~ U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]') NOT VALID,
  ADD CONSTRAINT journey_situations_short_description_nonblank
    CHECK (short_description ~ U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]') NOT VALID;
