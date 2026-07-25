package aifeedback

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SubmitSentenceFeedbackRequest is the service input for a sentence-feedback
// submission. It contains only the fields the frontend is permitted to send.
type SubmitSentenceFeedbackRequest struct {
	UserID         uuid.UUID
	SentenceText   string
	Source         string
	AttemptID      uuid.UUID
	IdempotencyKey string
}

// StoredFeedbackAttempt is the persisted row from ai_feedback_attempts.
type StoredFeedbackAttempt struct {
	ID                uuid.UUID
	LearnerSentenceID uuid.UUID
	Status            string
	Provider          string
	Model             string
	PromptVersion     string
	RequestHash       string
	FeedbackJSON      map[string]any
	FeedbackText      string
	ErrorCode         string
	ErrorMessage      string
}

// PendingAttempt holds the IDs created by CreatePendingAttempt.
type PendingAttempt struct {
	SentenceID uuid.UUID
	AttemptID  uuid.UUID
}

// Repository is the persistence boundary for the AI feedback domain.
type Repository interface {
	// LoadTarget loads the authoritative target word/phrase after verifying the
	// attempt is owned by the authenticated learner. If the attempt is missing,
	// inaccessible, or ineligible, it returns ErrTargetNotFound.
	LoadTarget(ctx context.Context, req LoadTargetRequest) (*Target, error)

	// GetFeedbackAttemptByRequestHash returns an existing attempt by its unique
	// deduplication hash. If no row exists, it returns nil, nil.
	GetFeedbackAttemptByRequestHash(ctx context.Context, requestHash string) (*StoredFeedbackAttempt, error)

	// CreatePendingAttempt inserts the learner_sentences row and the pending
	// ai_feedback_attempts row inside a single transaction. It does not call the
	// provider. provider and model are recorded on the pending attempt row.
	CreatePendingAttempt(ctx context.Context, req SubmitSentenceFeedbackRequest, target *Target, normalized string, requestHash string, provider string, model string, now time.Time) (*PendingAttempt, error)

	// CompleteFeedbackAttempt updates the attempt and sentence statuses after the
	// provider call. A non-nil feedback indicates success; otherwise failureCode
	// and failureMessage describe the failure.
	CompleteFeedbackAttempt(ctx context.Context, pending PendingAttempt, feedback *ProviderFeedback, failureCode, failureMessage string, now time.Time) error
}
