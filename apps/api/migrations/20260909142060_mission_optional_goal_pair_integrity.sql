-- atlas:txmode file
-- Issue #1397: optional daily bonus goals are target/completed pairs. Make
-- both columns absent together, or require a bounded progress value whenever
-- the target is present. NOT VALID preserves rollout over legacy rows while
-- enforcing the invariant for all new writes.

ALTER TABLE daily_mission_snapshots
  ADD CONSTRAINT daily_mission_snapshots_new_word_goal_pair_consistent
  CHECK (
    (new_word_target IS NULL AND new_words_completed IS NULL)
    OR (
      new_word_target IS NOT NULL
      AND new_words_completed IS NOT NULL
      AND new_words_completed >= 0
      AND new_words_completed <= new_word_target
    )
  ) NOT VALID;

ALTER TABLE daily_mission_snapshots
  ADD CONSTRAINT daily_mission_snapshots_sentence_goal_pair_consistent
  CHECK (
    (sentence_practice_target IS NULL AND sentence_practices_completed IS NULL)
    OR (
      sentence_practice_target IS NOT NULL
      AND sentence_practices_completed IS NOT NULL
      AND sentence_practices_completed >= 0
      AND sentence_practices_completed <= sentence_practice_target
    )
  ) NOT VALID;
