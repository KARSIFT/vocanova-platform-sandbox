package aifeedback

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestAIFeedbackOutcomeIntegrityPostgreSQL exercises the state-dependent
// payload checks on PostgreSQL itself. Keeping this DDL close to the migration
// lets the test use a private schema while asserting the exact production
// invariants against the database engine, not a mock.
func TestAIFeedbackOutcomeIntegrityPostgreSQL(t *testing.T) {
	db := isolatedReportPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, `
		CREATE TABLE ai_feedback_attempts (
			id uuid PRIMARY KEY,
			status text NOT NULL CHECK (status IN ('pending', 'succeeded', 'failed', 'cancelled')),
			feedback_json jsonb,
			feedback_text text,
			error_code text,
			error_message text,
			completed_at timestamptz,
			CONSTRAINT completed_at_required_on_success CHECK (status <> 'succeeded' OR completed_at IS NOT NULL),
			CONSTRAINT error_code_required_on_failure CHECK (status <> 'failed' OR error_code IS NOT NULL)
		)`)
	require.NoError(t, err)

	now := time.Date(2026, 9, 8, 14, 0, 0, 0, time.UTC)
	// A historic row that predates the new invariant must not make deployment
	// fail. NOT VALID constraints still reject every new or changed row.
	legacyID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO ai_feedback_attempts (id, status, completed_at)
		VALUES ($1, 'succeeded', $2)`, legacyID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		ALTER TABLE ai_feedback_attempts
			ADD CONSTRAINT feedback_json_required_on_success CHECK (status <> 'succeeded' OR (feedback_json IS NOT NULL AND jsonb_typeof(feedback_json) = 'object')) NOT VALID,
			ADD CONSTRAINT feedback_json_only_on_success CHECK (status = 'succeeded' OR feedback_json IS NULL) NOT VALID,
			ADD CONSTRAINT feedback_text_only_on_success CHECK (status = 'succeeded' OR feedback_text IS NULL) NOT VALID,
			ADD CONSTRAINT error_code_only_on_failure CHECK (status = 'failed' OR error_code IS NULL) NOT VALID,
			ADD CONSTRAINT error_message_only_on_failure CHECK (status = 'failed' OR error_message IS NULL) NOT VALID`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `UPDATE ai_feedback_attempts SET feedback_text = 'changed' WHERE id = $1`, legacyID)
	require.Error(t, err, "changing a legacy-invalid row must enforce the new constraints")

	insert := func(status string, feedbackJSON, feedbackText, errorCode, errorMessage any, completedAt any) error {
		_, err := db.ExecContext(ctx, `
			INSERT INTO ai_feedback_attempts (id, status, feedback_json, feedback_text, error_code, error_message, completed_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.New(), status, feedbackJSON, feedbackText, errorCode, errorMessage, completedAt)
		return err
	}

	require.NoError(t, insert(AttemptStatusPending, nil, nil, nil, nil, nil))
	require.NoError(t, insert(AttemptStatusSucceeded, `{"status":"correct"}`, "Good use of the word.", nil, nil, now))
	require.NoError(t, insert(AttemptStatusFailed, nil, nil, ErrorCodeTemporaryFailure, "provider timeout", now))
	require.NoError(t, insert(AttemptStatusCancelled, nil, nil, nil, nil, now))

	for _, malformed := range []struct {
		name       string
		status     string
		json, text any
		code, msg  any
		completed  any
	}{
		{name: "success without structured output", status: AttemptStatusSucceeded, completed: now},
		{name: "success with JSON null", status: AttemptStatusSucceeded, json: `null`, completed: now},
		{name: "success with JSON scalar", status: AttemptStatusSucceeded, json: `"correct"`, completed: now},
		{name: "success with failure code", status: AttemptStatusSucceeded, json: `{"status":"correct"}`, code: ErrorCodeTemporaryFailure, completed: now},
		{name: "failed with success payload", status: AttemptStatusFailed, json: `{"status":"correct"}`, code: ErrorCodeTemporaryFailure, completed: now},
		{name: "pending with feedback text", status: AttemptStatusPending, text: "leaked terminal text"},
		{name: "cancelled with failure fields", status: AttemptStatusCancelled, code: ErrorCodeTemporaryFailure, msg: "cancelled"},
	} {
		t.Run(malformed.name, func(t *testing.T) {
			err := insert(malformed.status, malformed.json, malformed.text, malformed.code, malformed.msg, malformed.completed)
			require.Error(t, err)
		})
	}
}
