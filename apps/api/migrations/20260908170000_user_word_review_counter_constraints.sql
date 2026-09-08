-- atlas:txmode file
-- DOC-05 §2 and §9: user_words is authoritative per-word learning state.
-- Counts cannot be negative even when a direct/future write bypasses service
-- validation. NOT VALID keeps this forward-only migration deployable if a
-- legacy cached row is already corrupt; PostgreSQL enforces it for all new
-- and updated rows immediately.

ALTER TABLE user_words
  ADD CONSTRAINT user_words_review_counts_nonnegative
  CHECK (
    consecutive_correct_count >= 0
    AND consecutive_incorrect_count >= 0
    AND total_review_count >= 0
    AND correct_review_count >= 0
  ) NOT VALID;
