// Package learning implements the learner-owned user_words save/unsave service.
// It writes no P2/P4 tables and exposes a read-only SavedStateReader boundary for
// the content module's saved-state overlay.
package learning

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/content"
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
	// SavedUserWordIDs returns a map from meaning_id to the owning user_word_id for
	// meanings that are currently saved by the requester. Missing or unsaved IDs
	// are omitted from the map.
	SavedUserWordIDs(ctx context.Context, userID uuid.UUID, meaningIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error)
	SavedWordStates(ctx context.Context, userID uuid.UUID, meaningIDs []uuid.UUID) (map[uuid.UUID]content.SavedWordState, error)
}

// IdempotencyStore scopes idempotency keys by user and operation.
type IdempotencyStore interface {
	// Check looks up a stored idempotency key. The returned status indicates
	// whether the key is absent, matches the fingerprint, or conflicts with it.
	Check(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) (IdempotencyStatus, error)
	// Record stores an idempotency key with its fingerprint.
	Record(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) error
}

// saveIdempotencyClaimer is implemented by the in-memory test store. Production
// saves deliberately do not use it: PostgreSQLRepository performs the claim in
// its transaction with the word and reward effects.
type saveIdempotencyClaimer interface {
	ClaimSave(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) (IdempotencyStatus, error)
	CompleteSave(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string)
	ReleaseSave(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string)
}

// atomicSaveRepository owns the database transaction for a save idempotency
// claim and its effects.
type atomicSaveRepository interface {
	SaveUserWordAtomically(ctx context.Context, req SaveUserWordRequest, fingerprint string, now time.Time) (*SavedMeaning, IdempotencyStatus, error)
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
	now := s.clock.Now().UTC()
	if repo, ok := s.repo.(atomicSaveRepository); ok {
		m, status, err := repo.SaveUserWordAtomically(ctx, req, fingerprint, now)
		if err != nil {
			return nil, err
		}
		if status == IdempotencyConflict {
			return nil, ErrIdempotencyConflict
		}
		return m, nil
	}
	if claimer, ok := s.idem.(saveIdempotencyClaimer); ok {
		status, err := claimer.ClaimSave(ctx, req.UserID, operationSaveUserWord, req.IdempotencyKey, fingerprint)
		if err != nil {
			return nil, fmt.Errorf("idempotency claim: %w", err)
		}
		if status == IdempotencyConflict || status == IdempotencyMatch {
			if status == IdempotencyConflict {
				return nil, ErrIdempotencyConflict
			}
			return s.repo.GetSavedMeaning(ctx, req.UserID, req.MeaningID)
		}
		m, err := s.repo.SaveUserWord(ctx, req, now)
		if err != nil {
			claimer.ReleaseSave(ctx, req.UserID, operationSaveUserWord, req.IdempotencyKey, fingerprint)
			return nil, err
		}
		claimer.CompleteSave(ctx, req.UserID, operationSaveUserWord, req.IdempotencyKey, fingerprint)
		return m, nil
	}
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

// SavedUserWordIDs implements the content.SavedStateReader boundary.
func (s *Service) SavedUserWordIDs(ctx context.Context, userID uuid.UUID, meaningIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	return s.repo.SavedUserWordIDs(ctx, userID, meaningIDs)
}

// SavedWordStates implements the content.SavedStateReader boundary.
func (s *Service) SavedWordStates(ctx context.Context, userID uuid.UUID, meaningIDs []uuid.UUID) (map[uuid.UUID]content.SavedWordState, error) {
	return s.repo.SavedWordStates(ctx, userID, meaningIDs)
}

func idempotencyFingerprint(meaningID uuid.UUID, source string) string {
	return fmt.Sprintf("%s|%s", meaningID.String(), source)
}
