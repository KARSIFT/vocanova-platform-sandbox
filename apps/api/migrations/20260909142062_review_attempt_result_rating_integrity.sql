-- atlas:txmode file
-- DOC-05 §9: result is objective correctness while rating is the scheduling
-- choice. Keep this NOT VALID so legacy rows do not block rollout; PostgreSQL
-- enforces the documented state machine for every new or updated row.
ALTER TABLE review_attempts
  ADD CONSTRAINT review_attempts_result_rating_valid
  CHECK (
    (result = 'skipped' AND rating IS NULL) OR
    (result = 'incorrect' AND rating IS NOT NULL AND rating = 'again') OR
    (result = 'correct' AND rating IS NOT NULL AND rating IN ('hard', 'good', 'easy'))
  ) NOT VALID;
