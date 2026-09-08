-- atlas:txmode file
-- DOC-05 §16 requires explicit, verified disposal of learner-linked records.
-- Reports are deleted by the account-deletion transaction before their
-- parents, so automatic cascades would hide an accidental data loss.
ALTER TABLE ai_feedback_quality_review_reports
  DROP CONSTRAINT ai_feedback_quality_review_reports_ai_feedback_attempt_id_fkey,
  DROP CONSTRAINT ai_feedback_quality_review_reports_user_id_fkey;

ALTER TABLE ai_feedback_quality_review_reports
  ADD CONSTRAINT ai_feedback_quality_review_reports_ai_feedback_attempt_id_fkey
    FOREIGN KEY (ai_feedback_attempt_id) REFERENCES ai_feedback_attempts(id) ON DELETE RESTRICT,
  ADD CONSTRAINT ai_feedback_quality_review_reports_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
