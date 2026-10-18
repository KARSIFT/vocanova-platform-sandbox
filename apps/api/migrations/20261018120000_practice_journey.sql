-- atlas:txmode file
-- Personal mixed practice reuses durable sessions and action receipts. It never
-- writes review schedules, daily review successes, missions or rewards.
ALTER TABLE practice_sessions DROP CONSTRAINT practice_sessions_mode_check;
ALTER TABLE practice_sessions ADD CONSTRAINT practice_sessions_mode_check CHECK(mode IN ('typed_recall','listening_choice','mistakes','journey'));
ALTER TABLE practice_sessions DROP CONSTRAINT practice_sessions_total_steps_check;
ALTER TABLE practice_sessions ADD CONSTRAINT practice_sessions_total_steps_check CHECK(total_steps BETWEEN 1 AND 12);
CREATE UNIQUE INDEX practice_journey_active_user_idx ON practice_sessions(user_id) WHERE mode='journey' AND status='in_progress';
CREATE INDEX practice_actions_recent_meaning_idx ON practice_actions(user_id,meaning_id,created_at DESC) WHERE meaning_id IS NOT NULL;
