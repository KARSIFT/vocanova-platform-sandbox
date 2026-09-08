-- atlas:txmode file
-- DOC-05 §§2,10: daily review aggregates are counters. Preserve the
-- immutable review_attempts source of truth by rejecting impossible cached
-- projections rather than allowing a corrupted summary to reach Home/streak
-- and progress reads.

ALTER TABLE daily_activity_summaries
  ADD CONSTRAINT daily_activity_summaries_review_counters_nonnegative
    CHECK (
      reviews_attempted >= 0
      AND reviews_correct >= 0
      AND reviews_skipped >= 0
    ),
  ADD CONSTRAINT daily_activity_summaries_review_counters_classified_within_attempted
    CHECK (reviews_correct <= reviews_attempted - reviews_skipped);
