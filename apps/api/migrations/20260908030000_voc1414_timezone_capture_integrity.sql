-- atlas:txmode file
-- DOC-05 §4 requires a timezone capture alongside historical local-date
-- records. Keep existing rows rollout-safe while rejecting blank values for
-- every new or updated record.

ALTER TABLE daily_mission_snapshots
  ADD CONSTRAINT daily_mission_snapshots_timezone_nonblank
  CHECK (timezone !~ '^[[:space:]]*$') NOT VALID;

ALTER TABLE daily_activity_summaries
  ADD CONSTRAINT daily_activity_summaries_timezone_nonblank
  CHECK (timezone !~ '^[[:space:]]*$') NOT VALID;

ALTER TABLE streak_states
  ADD CONSTRAINT streak_states_timezone_nonblank
  CHECK (timezone !~ '^[[:space:]]*$') NOT VALID;

ALTER TABLE grace_day_ledger
  ADD CONSTRAINT grace_day_ledger_timezone_nonblank
  CHECK (timezone !~ '^[[:space:]]*$') NOT VALID;
