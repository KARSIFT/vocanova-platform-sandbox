-- atlas:txmode transaction
-- VOC-027-T00: review_attempts immutable-history table for spaced-repetition
-- review submissions. Applied explicitly by migration tooling; the API process
-- does not execute migrations at startup.

CREATE TABLE review_attempts (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  user_word_id uuid NOT NULL REFERENCES user_words(id) ON DELETE RESTRICT,
  meaning_id uuid NOT NULL REFERENCES word_meanings(id) ON DELETE RESTRICT,
  attempt_type text NOT NULL CHECK (attempt_type <> ''),
  prompt_type text NOT NULL CHECK (prompt_type IN ('multiple_choice', 'self_check')),
  result text NOT NULL CHECK (result IN ('correct', 'incorrect', 'skipped')),
  rating text CHECK (rating IS NULL OR rating IN ('again', 'hard', 'good', 'easy')),
  review_step_before integer NOT NULL CHECK (review_step_before >= 0 AND review_step_before <= 7),
  review_step_after integer NOT NULL CHECK (review_step_after >= 0 AND review_step_after <= 7),
  answered_at timestamptz NOT NULL,
  response_time_ms integer NOT NULL DEFAULT 0 CHECK (response_time_ms >= 0),
  selected_option_meaning_id uuid REFERENCES word_meanings(id) ON DELETE RESTRICT,
  typed_answer text,
  was_hint_used boolean NOT NULL DEFAULT false,
  source text NOT NULL CHECK (source <> ''),
  client_attempt_id text,
  metadata jsonb,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT client_attempt_id_non_empty CHECK (client_attempt_id IS NULL OR client_attempt_id <> '')
);

CREATE UNIQUE INDEX review_attempts_user_id_client_attempt_id_key
  ON review_attempts (user_id, client_attempt_id)
  WHERE client_attempt_id IS NOT NULL;

CREATE INDEX review_attempts_user_id_answered_at_idx
  ON review_attempts (user_id, answered_at);
CREATE INDEX review_attempts_user_word_id_answered_at_idx
  ON review_attempts (user_word_id, answered_at);
CREATE INDEX review_attempts_meaning_id_answered_at_idx
  ON review_attempts (meaning_id, answered_at);
