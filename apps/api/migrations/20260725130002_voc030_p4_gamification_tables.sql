-- atlas:txmode file
-- VOC-030-T00: confidence_point_ledger, streak_states, grace_day_ledger
-- (DOC-05 §12). All append-only or per-user-unique. The two ledgers have
-- append-only semantics (no cascading deletes) per DOC-05 §16. D02 adds
-- `word_added` reason / `user_word` source_type for the Add word reward.

CREATE TABLE confidence_point_ledger (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  amount integer NOT NULL,
  balance_after integer NOT NULL,
  reason text NOT NULL,
  source_type text NOT NULL,
  source_id uuid,
  idempotency_key text,
  metadata jsonb,
  occurred_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT confidence_point_ledger_amount_nonzero CHECK (amount <> 0),
  CONSTRAINT confidence_point_ledger_reason_valid CHECK (reason IN ('word_added', 'review_correct', 'daily_mission_completed', 'sentence_submitted', 'ai_feedback_received', 'streak_bonus', 'admin_adjustment')),
  CONSTRAINT confidence_point_ledger_source_type_valid CHECK (source_type IN ('user_word', 'review_attempt', 'daily_mission', 'learner_sentence', 'ai_feedback_attempt', 'streak', 'admin'))
);

CREATE UNIQUE INDEX confidence_point_ledger_user_id_idempotency_key_key
  ON confidence_point_ledger (user_id, idempotency_key)
  WHERE idempotency_key IS NOT NULL;
CREATE INDEX confidence_point_ledger_user_id_occurred_at_idx
  ON confidence_point_ledger (user_id, occurred_at);

CREATE TABLE streak_states (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE RESTRICT,
  current_streak_count integer NOT NULL DEFAULT 0,
  longest_streak_count integer NOT NULL DEFAULT 0,
  last_completed_local_date date,
  last_activity_local_date date,
  timezone text NOT NULL,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT streak_states_longest_ge_current CHECK (longest_streak_count >= current_streak_count),
  CONSTRAINT streak_states_counts_nonnegative CHECK (current_streak_count >= 0 AND longest_streak_count >= 0),
  CONSTRAINT streak_states_status_valid CHECK (status IN ('active', 'at_risk', 'broken'))
);

CREATE TABLE grace_day_ledger (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  amount integer NOT NULL,
  balance_after integer NOT NULL,
  reason text NOT NULL,
  source_type text NOT NULL,
  source_id uuid,
  applied_to_local_date date NOT NULL,
  timezone text NOT NULL,
  idempotency_key text,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT grace_day_ledger_amount_nonzero CHECK (amount <> 0),
  CONSTRAINT grace_day_ledger_reason_valid CHECK (reason IN ('earned_by_streak', 'manual_grant', 'used_for_missed_day', 'expired', 'admin_adjustment')),
  CONSTRAINT grace_day_ledger_source_type_valid CHECK (source_type IN ('daily_mission', 'streak', 'admin'))
);

CREATE UNIQUE INDEX grace_day_ledger_user_id_idempotency_key_key
  ON grace_day_ledger (user_id, idempotency_key)
  WHERE idempotency_key IS NOT NULL;
CREATE INDEX grace_day_ledger_user_id_applied_to_local_date_idx
  ON grace_day_ledger (user_id, applied_to_local_date);
