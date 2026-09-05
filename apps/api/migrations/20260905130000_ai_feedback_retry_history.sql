-- atlas:txmode file
-- Preserve failed provider generations while allowing one retryable pending
-- generation for the same logical feedback request.

DROP INDEX ai_feedback_attempts_request_hash_key;

CREATE UNIQUE INDEX ai_feedback_attempts_request_hash_active_key
  ON ai_feedback_attempts (request_hash)
  WHERE status IN ('pending', 'succeeded');
