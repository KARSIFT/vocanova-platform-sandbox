// Package reviews implements the spaced-repetition review domain and persistence
// boundaries. It owns the scheduling domain and the due-queue read side; the
// submission transaction is added by later tasks.
package reviews

import (
	"context"
	"errors"
	"time"

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

// Repository is the persistence boundary for the review domain.
type Repository interface {
	// ListDueWords returns the requester's due words according to the DOC-05 §9
	// due-word rule. Implementations are requester-scoped and must never expose
	// another learner's rows.
	ListDueWords(ctx context.Context, req ListDueWordsRequest) (*ListDueWordsResponse, error)
}

// Public service errors.
var (
	ErrInvalidCursor = errors.New("invalid cursor")
)

// Service implements the review-domain read side.
type Service struct {
	repo Repository
}

// NewService creates a reviews service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ListDueWords returns the requester's due words.
func (s *Service) ListDueWords(ctx context.Context, req ListDueWordsRequest) (*ListDueWordsResponse, error) {
	if req.UserID == uuid.Nil {
		return nil, errors.New("user id required")
	}
	return s.repo.ListDueWords(ctx, req)
}
