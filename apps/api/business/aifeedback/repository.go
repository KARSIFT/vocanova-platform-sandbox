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
	Reported          bool
}

// PendingAttempt holds the IDs created by CreatePendingAttempt.
type PendingAttempt struct {
	SentenceID uuid.UUID
	AttemptID  uuid.UUID
}

// RetryAttempt records whether a retry created a new provider generation or
// lost a concurrent race to one that is already active.
type RetryAttempt struct {
	Pending  *PendingAttempt
	Existing *StoredFeedbackAttempt
}

// QualityReviewReport is the internal record created when a learner reports
// feedback. Classification is intentionally unset until a later triage flow.
type QualityReviewReport struct {
	ID             uuid.UUID
	AttemptID      uuid.UUID
	UserID         uuid.UUID
	Reason         string
	State          string
	Classification *string
	CreatedAt      time.Time
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

	// CreateRetryAttempt appends a pending generation for a previously failed
	// logical submission, preserving the failed attempt for observability.
	CreateRetryAttempt(ctx context.Context, failed *StoredFeedbackAttempt, provider string, model string, now time.Time) (*RetryAttempt, error)

	// CompleteFeedbackAttempt updates the attempt and sentence statuses after the
	// provider call. A non-nil feedback indicates success; otherwise failureCode
	// and failureMessage describe the failure.
	CompleteFeedbackAttempt(ctx context.Context, pending PendingAttempt, feedback *ProviderFeedback, failureCode, failureMessage string, now time.Time) error

	// GetFeedbackAttemptOwner returns the learner user_id that owns the sentence
	// associated with the given ai_feedback_attempts row. If the attempt does not
	// exist, it returns ErrTargetNotFound so the caller can surface a 404.
	GetFeedbackAttemptOwner(ctx context.Context, attemptID uuid.UUID) (uuid.UUID, error)

	// CreateQualityReviewReport atomically claims the user-scoped reporting key
	// and creates at most one report for an attempt.
	// It returns false when that report already exists.
	CreateQualityReviewReport(ctx context.Context, report QualityReviewReport, idempotencyKey string) (bool, error)
}
