-- atlas:txmode file
-- VOC-1398: a completed AI feedback attempt must contain exactly the payload
-- for its documented terminal outcome. These checks protect read/replay paths
-- from direct writes or future migrations that would otherwise persist a
-- "successful" attempt without a structured learner result. The constraints
-- are NOT VALID so deployment remains safe if historical records predate these
-- invariants; PostgreSQL still enforces them for every new or changed row.

ALTER TABLE ai_feedback_attempts
  ADD CONSTRAINT ai_feedback_attempts_feedback_json_required_on_success
    CHECK (status <> 'succeeded' OR (feedback_json IS NOT NULL AND jsonb_typeof(feedback_json) = 'object')) NOT VALID,
  ADD CONSTRAINT ai_feedback_attempts_feedback_json_only_on_success
    CHECK (status = 'succeeded' OR feedback_json IS NULL) NOT VALID,
  ADD CONSTRAINT ai_feedback_attempts_feedback_text_only_on_success
    CHECK (status = 'succeeded' OR feedback_text IS NULL) NOT VALID,
  ADD CONSTRAINT ai_feedback_attempts_error_code_only_on_failure
    CHECK (status = 'failed' OR error_code IS NULL) NOT VALID,
  ADD CONSTRAINT ai_feedback_attempts_error_message_only_on_failure
    CHECK (status = 'failed' OR error_message IS NULL) NOT VALID;
