-- atlas:txmode file
-- Issue #1440: a positive streak must have a completion anchor, and every
-- completion is activity on that date or later. NOT VALID preserves legacy
-- contradictions while protecting all new inserts and updates.

ALTER TABLE streak_states
  ADD CONSTRAINT streak_states_positive_streak_has_date
    CHECK (current_streak_count = 0 OR last_completed_local_date IS NOT NULL)
    NOT VALID,
  ADD CONSTRAINT streak_states_completion_is_activity
    CHECK (
      last_completed_local_date IS NULL
      OR (
        last_activity_local_date IS NOT NULL
        AND last_completed_local_date <= last_activity_local_date
      )
    )
    NOT VALID;
