-- atlas:txmode file
-- Issue #1431: platform-owned ordering positions are one-based. NOT VALID
-- preserves legacy content rows while protecting new inserts and updates.

ALTER TABLE word_meanings
  ADD CONSTRAINT word_meanings_meaning_order_positive
    CHECK (meaning_order > 0) NOT VALID;

ALTER TABLE word_examples
  ADD CONSTRAINT word_examples_example_order_positive
    CHECK (example_order > 0) NOT VALID;

ALTER TABLE usage_notes
  ADD CONSTRAINT usage_notes_note_order_positive
    CHECK (note_order > 0) NOT VALID;

ALTER TABLE journey_situations
  ADD CONSTRAINT journey_situations_display_order_positive
    CHECK (display_order > 0) NOT VALID;

ALTER TABLE journey_words
  ADD CONSTRAINT journey_words_display_order_positive
    CHECK (display_order IS NULL OR display_order > 0) NOT VALID;
