-- atlas:txmode file
-- Issue #1406: user_words is the current review-scheduling state, so its last
-- result and rating must obey the same canonical pairings as review history.
-- NOT VALID preserves rollout over possible legacy contradictions while the
-- constraint still protects every newly inserted or updated row.

ALTER TABLE user_words
  ADD CONSTRAINT user_words_last_result_rating_consistent
  CHECK (
    (
      (last_result IS NULL AND last_rating IS NULL)
      OR (last_result = 'skipped' AND last_rating IS NULL)
      OR (
        last_result = 'incorrect'
        AND last_rating IS NOT NULL
        AND last_rating = 'again'
      )
      OR (
        last_result = 'correct'
        AND last_rating IS NOT NULL
        AND last_rating IN ('hard', 'good', 'easy')
      )
    ) IS TRUE
  )
  NOT VALID;
