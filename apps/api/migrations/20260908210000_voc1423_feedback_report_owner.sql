-- atlas:txmode file
-- DOC-05 §§11/14/16 and DOC-09 §16: a learner report belongs to the
-- learner-owned sentence feedback it reviews. PostgreSQL cannot express that
-- ownership chain with a CHECK or ordinary foreign key, so protect new report
-- ownership writes with a trigger. Existing retained records are deliberately
-- not scanned or rewritten by this additive rollout.

CREATE FUNCTION enforce_ai_feedback_quality_review_report_owner()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM ai_feedback_attempts AS attempt
    JOIN learner_sentences AS sentence ON sentence.id = attempt.learner_sentence_id
    WHERE attempt.id = NEW.ai_feedback_attempt_id
      AND sentence.user_id = NEW.user_id
  ) THEN
    RAISE EXCEPTION
      USING ERRCODE = '23503',
            MESSAGE = 'ai feedback quality review report user must own its feedback attempt';
  END IF;

  RETURN NEW;
END;
$$;

CREATE TRIGGER ai_feedback_quality_review_reports_owner_matches_attempt
  BEFORE INSERT OR UPDATE OF ai_feedback_attempt_id, user_id
  ON ai_feedback_quality_review_reports
  FOR EACH ROW
  EXECUTE FUNCTION enforce_ai_feedback_quality_review_report_owner();

CREATE FUNCTION enforce_ai_feedback_attempt_report_owner()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM ai_feedback_quality_review_reports AS report
    JOIN learner_sentences AS sentence ON sentence.id = NEW.learner_sentence_id
    WHERE report.ai_feedback_attempt_id = NEW.id
      AND report.user_id <> sentence.user_id
  ) THEN
    RAISE EXCEPTION
      USING ERRCODE = '23503',
            MESSAGE = 'ai feedback attempt cannot move away from its report owner';
  END IF;

  RETURN NEW;
END;
$$;

CREATE TRIGGER ai_feedback_attempts_owner_matches_reports
  BEFORE UPDATE OF learner_sentence_id
  ON ai_feedback_attempts
  FOR EACH ROW
  EXECUTE FUNCTION enforce_ai_feedback_attempt_report_owner();

CREATE FUNCTION enforce_learner_sentence_report_owner()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM ai_feedback_quality_review_reports AS report
    JOIN ai_feedback_attempts AS attempt ON attempt.id = report.ai_feedback_attempt_id
    WHERE attempt.learner_sentence_id = NEW.id
      AND report.user_id <> NEW.user_id
  ) THEN
    RAISE EXCEPTION
      USING ERRCODE = '23503',
            MESSAGE = 'learner sentence cannot move away from its feedback report owner';
  END IF;

  RETURN NEW;
END;
$$;

CREATE TRIGGER learner_sentences_owner_matches_feedback_reports
  BEFORE UPDATE OF user_id
  ON learner_sentences
  FOR EACH ROW
  EXECUTE FUNCTION enforce_learner_sentence_report_owner();
