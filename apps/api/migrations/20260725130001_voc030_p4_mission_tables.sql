-- atlas:txmode file
-- VOC-030-T00: daily_mission_snapshots and daily_activity_summaries tables
-- (DOC-05 §10). Both are owned by the missions module; records of truth remain
-- review_attempts, learner_sentences, ai_feedback_attempts, and
-- confidence_point_ledger.

CREATE TABLE daily_mission_snapshots (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  local_date date NOT NULL,
  timezone text NOT NULL,
  review_target integer NOT NULL,
  reviews_completed integer NOT NULL DEFAULT 0,
  new_word_target integer,
  new_words_completed integer,
  sentence_practice_target integer,
  sentence_practices_completed integer,
  policy_version text NOT NULL,
  status text NOT NULL DEFAULT 'open',
  completed_at timestamptz,
  grace_applied boolean NOT NULL DEFAULT false,
  grace_day_id uuid,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT daily_mission_snapshots_review_target_in_range CHECK (review_target >= 5 AND review_target <= 100),
  CONSTRAINT daily_mission_snapshots_reviews_completed_in_range CHECK (reviews_completed >= 0 AND reviews_completed <= review_target),
  CONSTRAINT daily_mission_snapshots_new_word_target_in_range CHECK (new_word_target IS NULL OR (new_word_target >= 1 AND new_word_target <= 100)),
  CONSTRAINT daily_mission_snapshots_new_words_completed_in_range CHECK (new_word_target IS NULL OR (new_words_completed >= 0 AND new_words_completed <= new_word_target)),
  CONSTRAINT daily_mission_snapshots_sentence_target_in_range CHECK (sentence_practice_target IS NULL OR (sentence_practice_target >= 1 AND sentence_practice_target <= 100)),
  CONSTRAINT daily_mission_snapshots_sentence_completed_in_range CHECK (sentence_practice_target IS NULL OR (sentence_practices_completed >= 0 AND sentence_practices_completed <= sentence_practice_target)),
  CONSTRAINT daily_mission_snapshots_status_valid CHECK (status IN ('open', 'completed', 'missed', 'protected')),
  CONSTRAINT daily_mission_snapshots_completed_at_required_on_done CHECK (status <> 'completed' OR completed_at IS NOT NULL)
);

CREATE UNIQUE INDEX daily_mission_snapshots_user_id_local_date_key
  ON daily_mission_snapshots (user_id, local_date);
CREATE INDEX daily_mission_snapshots_user_id_status_idx
  ON daily_mission_snapshots (user_id, status);

CREATE TABLE daily_activity_summaries (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  local_date date NOT NULL,
  timezone text NOT NULL,
  reviews_attempted integer NOT NULL DEFAULT 0,
  reviews_correct integer NOT NULL DEFAULT 0,
  reviews_skipped integer NOT NULL DEFAULT 0,
  words_discovered integer NOT NULL DEFAULT 0,
  words_added integer NOT NULL DEFAULT 0,
  sentences_submitted integer NOT NULL DEFAULT 0,
  ai_feedback_received integer NOT NULL DEFAULT 0,
  confidence_points_earned integer NOT NULL DEFAULT 0,
  confidence_points_spent integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX daily_activity_summaries_user_id_local_date_key
  ON daily_activity_summaries (user_id, local_date);
