-- atlas:txmode file
-- VOC-1438: frequency ranks are one-based when catalogued. Keep this check
-- NOT VALID so an existing catalog can be migrated before any legacy bad rows
-- are corrected; PostgreSQL still enforces it for every subsequent insert and
-- update.

ALTER TABLE canonical_words
  ADD CONSTRAINT canonical_words_frequency_rank_positive
  CHECK (frequency_rank IS NULL OR frequency_rank > 0) NOT VALID;
