// Package reviews implements the spaced-repetition review domain and persistence
// boundaries. It owns the scheduling domain and the due-queue read side; the
// submission transaction is added by later tasks.
package reviews

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
)

// DueWord is a saved meaning that is due for review now.
type DueWord struct {
	UserWordID      uuid.UUID
	MeaningID       uuid.UUID
	WordID          uuid.UUID
	WordText        string
	WordSlug        string
	PartOfSpeech    string
	ShortDefinition string
	Status          string
	ReviewStep      int
	NextReviewAt    *time.Time
	Source          string
}

// ListDueWordsRequest is a paginated query for the authenticated user's due words.
type ListDueWordsRequest struct {
	UserID      uuid.UUID
	AfterCursor string
	Limit       int
}

// ListDueWordsResponse is a paginated list of due words with a total count.
type ListDueWordsResponse struct {
	Items      []DueWord
	NextCursor string
	TotalCount int
}

// ReviewAttempt is one immutable learner response recorded in review_attempts.
type ReviewAttempt struct {
	ID                      uuid.UUID
	UserID                  uuid.UUID
	UserWordID              uuid.UUID
	MeaningID               uuid.UUID
	AttemptType             string
	PromptType              string
	Result                  string
	Rating                  string
	ReviewStepBefore        int
	ReviewStepAfter         int
	AnsweredAt              time.Time
	ResponseTimeMs          int
	SelectedOptionMeaningID *uuid.UUID
	TypedAnswer             *string
	WasHintUsed             bool
	Source                  string
	ClientAttemptID         string
	Metadata                map[string]any
	NextReviewAt            time.Time
}

// SubmitReviewRequest is a request to record a review attempt and update the
// schedule for the authenticated user.
type SubmitReviewRequest struct {
	UserID                  uuid.UUID
	UserWordID              uuid.UUID
	MeaningID               uuid.UUID
	AttemptType             string
	PromptType              string
	Result                  string
	Rating                  string
	AnsweredAt              time.Time
	ResponseTimeMs          int
	SelectedOptionMeaningID *uuid.UUID
	TypedAnswer             *string
	WasHintUsed             bool
	Source                  string
	ClientAttemptID         string
	Metadata                map[string]any
	IdempotencyKey          string
}

// Prompt types supported by the P2 review submission API.
const (
	PromptTypeMultipleChoice = "multiple_choice"
	PromptTypeSelfCheck      = "self_check"
)

// Attempt type for P2 review submissions.
const (
	AttemptTypeReview = "review"
)

// Source values for P2 review submissions.
const (
	SourceReview        = "review"
	SourceReviewSession = "review_session"
)

// Repository is the persistence boundary for the review domain.
type Repository interface {
	// ListDueWords returns the requester's due words according to the DOC-05 §9
	// due-word rule. Implementations are requester-scoped and must never expose
	// another learner's rows.
	ListDueWords(ctx context.Context, req ListDueWordsRequest) (*ListDueWordsResponse, error)
	// SubmitReview records a review attempt and updates the user_words schedule
	// inside one transaction.
	SubmitReview(ctx context.Context, req SubmitReviewRequest) (*ReviewAttempt, error)
	// GetReviewAttemptByClientAttemptID returns an existing attempt by its
	// client-provided idempotency identifier for the authenticated user.
	GetReviewAttemptByClientAttemptID(ctx context.Context, userID uuid.UUID, clientAttemptID string) (*ReviewAttempt, error)
}

// Public service errors.
var (
	ErrInvalidCursor           = errors.New("invalid cursor")
	ErrUserWordNotFound        = errors.New("saved word not found")
	ErrInvalidPromptType       = errors.New("invalid prompt type")
	ErrInvalidRatingForResult  = errors.New("rating is not permitted for the result")
	ErrInvalidAttemptType      = errors.New("invalid attempt type")
	ErrInvalidSource           = errors.New("invalid source")
	ErrClientAttemptIDRequired = errors.New("client attempt id required")
	ErrInvalidAnsweredAt       = errors.New("answered at is required")
	ErrInvalidResponseTimeMs   = errors.New("response time must be non-negative")
	ErrIdempotencyKeyRequired  = errors.New("idempotency key required")
	ErrIdempotencyConflict     = errors.New("idempotency key conflict")
	ErrReviewAttemptNotFound   = errors.New("review attempt not found")
)

const operationSubmitReview = "reviews:submit"

// Service implements the review-domain read side and submission write side.
type Service struct {
	repo  Repository
	idem  learning.IdempotencyStore
	clock clock.Clock
}

// NewService creates a reviews service.
func NewService(repo Repository, idem learning.IdempotencyStore, c clock.Clock) *Service {
	if idem == nil {
		idem = learning.NewMemoryIdempotencyStore()
	}
	if c == nil {
		c = clock.Real{}
	}
	return &Service{repo: repo, idem: idem, clock: c}
}

// ListDueWords returns the requester's due words.
func (s *Service) ListDueWords(ctx context.Context, req ListDueWordsRequest) (*ListDueWordsResponse, error) {
	if req.UserID == uuid.Nil {
		return nil, errors.New("user id required")
	}
	return s.repo.ListDueWords(ctx, req)
}

// SubmitReview records a review attempt, applies the scheduling rule exactly once,
// and returns the immutable attempt. It is idempotent by (user_id,
// client_attempt_id) and by the user-scoped Idempotency-Key.
func (s *Service) SubmitReview(ctx context.Context, req SubmitReviewRequest) (*ReviewAttempt, error) {
	if req.UserID == uuid.Nil {
		return nil, errors.New("user id required")
	}
	if req.UserWordID == uuid.Nil {
		return nil, ErrUserWordNotFound
	}
	if req.MeaningID == uuid.Nil {
		return nil, ErrUserWordNotFound
	}
	if req.AttemptType == "" {
		req.AttemptType = AttemptTypeReview
	}
	if req.Source == "" {
		req.Source = SourceReview
	}
	if req.PromptType != PromptTypeMultipleChoice && req.PromptType != PromptTypeSelfCheck {
		return nil, fmt.Errorf("%w: %q", ErrInvalidPromptType, req.PromptType)
	}
	if req.Result != ResultCorrect && req.Result != ResultIncorrect && req.Result != ResultSkipped {
		return nil, fmt.Errorf("%w: %q", ErrInvalidResult, req.Result)
	}
	if req.Rating != "" && req.Rating != RatingAgain && req.Rating != RatingHard && req.Rating != RatingGood && req.Rating != RatingEasy {
		return nil, fmt.Errorf("%w: %q", ErrInvalidRating, req.Rating)
	}
	if err := validateRatingForResult(req.Result, req.Rating); err != nil {
		return nil, err
	}
	if req.AttemptType != AttemptTypeReview {
		return nil, fmt.Errorf("%w: %q", ErrInvalidAttemptType, req.AttemptType)
	}
	if req.Source != SourceReview && req.Source != SourceReviewSession {
		return nil, fmt.Errorf("%w: %q", ErrInvalidSource, req.Source)
	}
	if req.ClientAttemptID == "" {
		return nil, ErrClientAttemptIDRequired
	}
	if req.AnsweredAt.IsZero() {
		return nil, ErrInvalidAnsweredAt
	}
	if req.ResponseTimeMs < 0 {
		return nil, ErrInvalidResponseTimeMs
	}
	if req.IdempotencyKey == "" {
		return nil, ErrIdempotencyKeyRequired
	}

	req.AnsweredAt = req.AnsweredAt.UTC()

	fingerprint := submitReviewFingerprint(req)
	status, err := s.idem.Check(ctx, req.UserID, operationSubmitReview, req.IdempotencyKey, fingerprint)
	if err != nil {
		return nil, fmt.Errorf("idempotency check: %w", err)
	}
	if status == learning.IdempotencyConflict {
		return nil, ErrIdempotencyConflict
	}
	if status == learning.IdempotencyMatch {
		attempt, err := s.repo.GetReviewAttemptByClientAttemptID(ctx, req.UserID, req.ClientAttemptID)
		if err != nil {
			return nil, fmt.Errorf("fetch idempotent attempt: %w", err)
		}
		return attempt, nil
	}

	attempt, err := s.repo.SubmitReview(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := s.idem.Record(ctx, req.UserID, operationSubmitReview, req.IdempotencyKey, fingerprint); err != nil {
		return nil, fmt.Errorf("record idempotency: %w", err)
	}
	return attempt, nil
}

func validateRatingForResult(result, rating string) error {
	switch result {
	case ResultSkipped:
		if rating != "" {
			return fmt.Errorf("%w: skipped attempts cannot have a rating", ErrInvalidRatingForResult)
		}
	case ResultIncorrect:
		if rating != RatingAgain {
			return fmt.Errorf("%w: incorrect attempts must be rated again", ErrInvalidRatingForResult)
		}
	case ResultCorrect:
		if rating == "" || rating == RatingAgain {
			return fmt.Errorf("%w: correct attempts must be rated hard, good, or easy", ErrInvalidRatingForResult)
		}
	}
	return nil
}

func submitReviewFingerprint(req SubmitReviewRequest) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%s|%s|%s|%s|%s|%s|%t|%d|%s|",
		req.UserWordID.String(),
		req.MeaningID.String(),
		req.AttemptType,
		req.PromptType,
		req.Result,
		req.Rating,
		req.Source,
		req.ClientAttemptID,
		req.WasHintUsed,
		req.ResponseTimeMs,
		req.AnsweredAt.Format(time.RFC3339Nano),
	)
	if req.SelectedOptionMeaningID != nil {
		fmt.Fprintf(h, "%s|", req.SelectedOptionMeaningID.String())
	}
	if req.TypedAnswer != nil {
		fmt.Fprintf(h, "%s|", *req.TypedAnswer)
	}
	return hex.EncodeToString(h.Sum(nil))
}
