-- atlas:txmode file
-- DOC-05 §§2,10,12: earned and spent are distinct non-negative daily
-- aggregates. The immutable ledger stays signed; this cached projection must
-- preserve the direction in its documented counter.

ALTER TABLE daily_activity_summaries
  ADD CONSTRAINT daily_activity_summaries_confidence_point_counters_nonnegative
    CHECK (
      confidence_points_earned >= 0
      AND confidence_points_spent >= 0
    ) NOT VALID;
