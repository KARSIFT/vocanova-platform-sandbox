-- atlas:txmode file
-- VOC-1427: DOC-05 §16 keeps AI feedback generation history immutable during
-- the active-account lifecycle. The only supported mutation is the service's
-- pending-to-succeeded or pending-to-failed finalization; a retry appends a
-- separate pending row. Account deletion is the sole delete path and shares
-- VOC-1402's transaction-local vocanova.ledger_purge gate.

CREATE FUNCTION vocanova_reject_ai_feedback_attempt_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP = 'DELETE' THEN
    IF current_setting('vocanova.ledger_purge', true) = 'on' THEN
      RETURN OLD;
    END IF;

    RAISE EXCEPTION 'ai_feedback_attempts rows are immutable'
      USING ERRCODE = '55000';
  END IF;

  -- CompleteFeedbackAttempt and CompleteSuccessfulFeedbackAttempt only settle
  -- a pending attempt. Request identity/provenance stays fixed; terminal
  -- payload, completion time, and updated_at are written atomically here.
  IF OLD.status = 'pending'
     AND NEW.status IN ('succeeded', 'failed')
     AND NEW.id = OLD.id
     AND NEW.learner_sentence_id = OLD.learner_sentence_id
     AND NEW.provider = OLD.provider
     AND NEW.model = OLD.model
     AND NEW.prompt_version = OLD.prompt_version
     AND NEW.request_hash = OLD.request_hash
     AND NEW.started_at IS NOT DISTINCT FROM OLD.started_at
     AND NEW.created_at = OLD.created_at THEN
    RETURN NEW;
  END IF;

  RAISE EXCEPTION 'ai_feedback_attempts rows may only finalize pending attempts'
    USING ERRCODE = '55000';
END;
$$;

CREATE TRIGGER ai_feedback_attempts_immutable_history
  BEFORE UPDATE OR DELETE ON ai_feedback_attempts
  FOR EACH ROW EXECUTE FUNCTION vocanova_reject_ai_feedback_attempt_mutation();
