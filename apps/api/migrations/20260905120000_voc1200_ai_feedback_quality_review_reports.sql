-- atlas:txmode file
-- VOC-1200: learner reports against AI feedback results.
CREATE TABLE ai_feedback_quality_review_reports (
  id uuid PRIMARY KEY,
  ai_feedback_attempt_id uuid NOT NULL UNIQUE REFERENCES ai_feedback_attempts(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  reason text NOT NULL CHECK (reason IN ('already_correct', 'correction_changed_meaning', 'explanation_unclear', 'inappropriate', 'something_else')),
  state text NOT NULL DEFAULT 'open' CHECK (state IN ('open', 'reviewing', 'confirmed_issue', 'no_issue_found', 'duplicate', 'resolved')),
  classification text CHECK (classification IS NULL OR classification IN ('incorrect_judgment', 'unnecessary_correction', 'meaning_changed', 'unclear_explanation', 'inappropriate_tone', 'unsafe_response', 'regional_variant_error', 'provider_failure', 'other')),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);
CREATE INDEX ai_feedback_quality_review_reports_state_created_at_idx ON ai_feedback_quality_review_reports (state, created_at);
CREATE INDEX ai_feedback_quality_review_reports_user_id_idx ON ai_feedback_quality_review_reports (user_id);
