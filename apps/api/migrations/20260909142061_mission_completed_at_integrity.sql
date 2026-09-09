-- atlas:txmode file
-- Issue #1400: completed_at records when a mission first reached completed.
-- The original table check enforces completed -> timestamp; this complementary
-- check enforces timestamp -> completed. NOT VALID keeps rollout additive over
-- possible legacy contradictions while protecting all future writes.

ALTER TABLE daily_mission_snapshots
  ADD CONSTRAINT daily_mission_snapshots_completed_at_only_on_done
  CHECK (status = 'completed' OR completed_at IS NULL)
  NOT VALID;
