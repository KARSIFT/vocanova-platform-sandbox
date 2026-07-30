-- atlas:txmode file
-- VOC-028-T00: ai_feedback_attempts immutable-history table for P3 AI feedback
-- generations. Applied explicitly by migration tooling; the API process does
-- not execute migrations at startup.

CREATE TABLE ai_feedback_attempts (
  id uuid PRIMARY KEY,
  learner_sentence_id uuid NOT NULL REFERENCES learner_sentences(id) ON DELETE RESTRICT,
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'succeeded', 'failed', 'cancelled')),
  provider text NOT NULL CHECK (provider <> ''),
  model text NOT NULL CHECK (model <> ''),
  prompt_version text NOT NULL CHECK (prompt_version <> ''),
  request_hash text NOT NULL CHECK (request_hash <> ''),
  feedback_json jsonb,
  feedback_text text,
  error_code text,
  error_message text,
  started_at timestamptz,
  completed_at timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT ai_feedback_attempts_completed_at_required_on_success CHECK (status <> 'succeeded' OR completed_at IS NOT NULL),
  CONSTRAINT ai_feedback_attempts_error_code_required_on_failure CHECK (status <> 'failed' OR error_code IS NOT NULL)
);

CREATE UNIQUE INDEX ai_feedback_attempts_request_hash_key
  ON ai_feedback_attempts (request_hash);
CREATE INDEX ai_feedback_attempts_learner_sentence_id_started_at_idx
  ON ai_feedback_attempts (learner_sentence_id, started_at);
CREATE INDEX ai_feedback_attempts_status_started_at_idx
  ON ai_feedback_attempts (status, started_at);
