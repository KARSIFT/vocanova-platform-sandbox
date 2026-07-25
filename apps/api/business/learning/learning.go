// Package learning implements the learner-owned user_words save/unsave service.
// It writes no P2/P4 tables and exposes a read-only SavedStateReader boundary for
// the content module's saved-state overlay.
package learning

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
)

// UserWord is a learner-owned saved meaning record.
type UserWord struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	MeaningID  uuid.UUID
	Status     string
	Source     string
	ReviewStep int
	AddedAt    time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

// SavedMeaning is a saved meaning returned in lists, with the canonical word
// details needed by the Home/Progress screens without an extra round-trip.
type SavedMeaning struct {
	UserWordID      uuid.UUID
	MeaningID       uuid.UUID
	WordID          uuid.UUID
	WordText        string
	WordSlug        string
	PartOfSpeech    string
	ShortDefinition string
	Status          string
	Source          string
	Saved           bool
	AddedAt         time.Time
}

// SaveUserWordRequest is a request to save a meaning for the authenticated user.
type SaveUserWordRequest struct {
	UserID         uuid.UUID
	MeaningID      uuid.UUID
	Source         string
	IdempotencyKey string
}

// ListSavedWordsRequest is a paginated query for the authenticated user's saved meanings.
type ListSavedWordsRequest struct {
	UserID      uuid.UUID
	AfterCursor string
	Limit       int
}

// ListSavedWordsResponse is a paginated list of saved meanings.
type ListSavedWordsResponse struct {
	Items      []SavedMeaning
	NextCursor string
}

// Repository is the persistence boundary for learner-owned user_words.
type Repository interface {
	// SaveUserWord inserts or restores a user_words row for the requester.
	// It returns the existing active row without error when the meaning is already saved.
	SaveUserWord(ctx context.Context, req SaveUserWordRequest, now time.Time) (*SavedMeaning, error)
	// UnsaveUserWord soft-deletes the requester's saved row for the meaning.
	UnsaveUserWord(ctx context.Context, userID, meaningID uuid.UUID, now time.Time) error
	// GetSavedMeaning returns the saved meaning with canonical details for the requester.
	GetSavedMeaning(ctx context.Context, userID, meaningID uuid.UUID) (*SavedMeaning, error)
	// ListSavedWords returns the requester's saved meanings ordered by most recent first.
	ListSavedWords(ctx context.Context, req ListSavedWordsRequest) (*ListSavedWordsResponse, error)
	// IsSaved returns requester-scoped saved states for the given meaning IDs.
	IsSaved(ctx context.Context, userID uuid.UUID, meaningIDs []uuid.UUID) (map[uuid.UUID]bool, error)
}

// IdempotencyStore scopes idempotency keys by user and operation.
type IdempotencyStore interface {
	// Check looks up a stored idempotency key. The returned status indicates
	// whether the key is absent, matches the fingerprint, or conflicts with it.
	Check(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) (IdempotencyStatus, error)
	// Record stores an idempotency key with its fingerprint.
	Record(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) error
}

// IdempotencyStatus describes the result of an idempotency key lookup.
type IdempotencyStatus int

const (
	IdempotencyAbsent IdempotencyStatus = iota
	IdempotencyMatch
	IdempotencyConflict
)

// Public service errors.
var (
	ErrMeaningNotFound        = errors.New("meaning not found")
	ErrUserWordNotFound       = errors.New("user word not found")
	ErrIdempotencyConflict    = errors.New("idempotency key conflict")
	ErrInvalidCursor          = errors.New("invalid cursor")
	ErrIdempotencyKeyRequired = errors.New("idempotency key required")
)

const operationSaveUserWord = "user_words:save"

// Service implements save/unsave/list for user_words.
type Service struct {
	repo  Repository
	idem  IdempotencyStore
	clock clock.Clock
}

// NewService creates a learning service.
func NewService(repo Repository, idem IdempotencyStore, c clock.Clock) *Service {
	if c == nil {
		c = clock.Real{}
	}
	return &Service{repo: repo, idem: idem, clock: c}
}

// SaveUserWord saves a meaning for the authenticated requester.
// It is idempotent by meaning and by idempotency key.
func (s *Service) SaveUserWord(ctx context.Context, req SaveUserWordRequest) (*SavedMeaning, error) {
	if req.UserID == uuid.Nil {
		return nil, errors.New("user id required")
	}
	if req.IdempotencyKey == "" {
		return nil, ErrIdempotencyKeyRequired
	}

	fingerprint := idempotencyFingerprint(req.MeaningID, req.Source)
	status, err := s.idem.Check(ctx, req.UserID, operationSaveUserWord, req.IdempotencyKey, fingerprint)
	if err != nil {
		return nil, fmt.Errorf("idempotency check: %w", err)
	}
	if status == IdempotencyConflict {
		return nil, ErrIdempotencyConflict
	}
	if status == IdempotencyMatch {
		return s.repo.GetSavedMeaning(ctx, req.UserID, req.MeaningID)
	}

	now := s.clock.Now().UTC()
	m, err := s.repo.SaveUserWord(ctx, req, now)
	if err != nil {
		return nil, err
	}

	if err := s.idem.Record(ctx, req.UserID, operationSaveUserWord, req.IdempotencyKey, fingerprint); err != nil {
		return nil, fmt.Errorf("record idempotency: %w", err)
	}
	return m, nil
}

// UnsaveUserWord soft-deletes the requester's saved meaning.
func (s *Service) UnsaveUserWord(ctx context.Context, userID, meaningID uuid.UUID) error {
	if userID == uuid.Nil {
		return errors.New("user id required")
	}
	now := s.clock.Now().UTC()
	return s.repo.UnsaveUserWord(ctx, userID, meaningID, now)
}

// ListSavedWords returns the requester's saved meanings.
func (s *Service) ListSavedWords(ctx context.Context, req ListSavedWordsRequest) (*ListSavedWordsResponse, error) {
	if req.UserID == uuid.Nil {
		return nil, errors.New("user id required")
	}
	return s.repo.ListSavedWords(ctx, req)
}

// IsSaved implements the content.SavedStateReader boundary.
func (s *Service) IsSaved(ctx context.Context, userID uuid.UUID, meaningIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	return s.repo.IsSaved(ctx, userID, meaningIDs)
}

func idempotencyFingerprint(meaningID uuid.UUID, source string) string {
	return fmt.Sprintf("%s|%s", meaningID.String(), source)
}
