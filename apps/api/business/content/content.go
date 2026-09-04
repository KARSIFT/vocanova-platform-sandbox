// Package content implements the discovery and canonical-word read service. It owns
// no learner state; saved-state overlay is provided through a read-only
// SavedStateReader boundary.
package content

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Situation is a public journey situation projection.
type Situation struct {
	ID               uuid.UUID
	Slug             string
	Title            string
	ShortDescription string
	LevelBand        string
	Category         string
	Status           string
	DisplayOrder     int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// MeaningSummary is a meaning entry as shown in a situation drill-down.
type MeaningSummary struct {
	MeaningID       uuid.UUID
	WordID          uuid.UUID
	WordSlug        string
	WordText        string
	PartOfSpeech    string
	ShortDefinition string
	Saved           bool
}

// SituationDetail is a situation with its meanings.
type SituationDetail struct {
	Situation Situation
	Meanings  []MeaningSummary
}

// WordExample is a canonical example sentence for a meaning.
type WordExample struct {
	ID             uuid.UUID
	MeaningID      uuid.UUID
	ExampleText    string
	SituationLabel string
}

// WordUsageNote is a usage note for a meaning.
type WordUsageNote struct {
	ID        uuid.UUID
	MeaningID uuid.UUID
	NoteType  string
	NoteText  string
}

// WordMeaning is a meaning with its examples and usage notes.
type WordMeaning struct {
	ID                uuid.UUID
	PartOfSpeech      string
	ShortDefinition   string
	LearnerDefinition string
	MeaningOrder      int
	Examples          []WordExample
	UsageNotes        []WordUsageNote
	Saved             bool
	UserWordID        uuid.UUID
}

// WordDetail is a canonical word with all its meanings.
type WordDetail struct {
	ID              uuid.UUID
	Text            string
	NormalizedText  string
	Slug            string
	WordType        string
	DifficultyLevel string
	Meanings        []WordMeaning
}

// Repository is the persistence boundary for canonical content.
type Repository interface {
	ListSituations(ctx context.Context, req ListSituationsRequest) (*ListSituationsResponse, error)
	GetSituationBySlug(ctx context.Context, slug string) (*Situation, error)
	GetMeaningsBySituation(ctx context.Context, situationID uuid.UUID) ([]MeaningSummary, error)
	GetWordBySlug(ctx context.Context, wordSlug string) (*WordDetail, error)
}

// SavedStateReader returns requester-scoped saved states for meaning IDs.
// Implementations are read-only and must never expose another learner's rows.
type SavedStateReader interface {
	IsSaved(ctx context.Context, userID uuid.UUID, meaningIDs []uuid.UUID) (map[uuid.UUID]bool, error)
	// SavedUserWordIDs returns a map from meaning_id to the owning user_word_id for
	// meanings that are currently saved by the requester. Missing or unsaved IDs
	// are omitted from the map.
	SavedUserWordIDs(ctx context.Context, userID uuid.UUID, meaningIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error)
}

// ListSituationsRequest is a paginated query.
//
// PriorityCategory is an optional hint that boosts situations whose
// Category matches it to the front of the first page (VOC-1183): a new
// learner's onboarding goal (main use case) should shape what Discover
// surfaces first, rather than being collected and never read. It is
// only applied to the first page (AfterCursor == "") — display_order
// remains the sole ordering for every page after that, and it never
// changes the pagination cursor, so paging behavior is unaffected.
type ListSituationsRequest struct {
	AfterCursor      string
	Limit            int
	PriorityCategory string
}

// ListSituationsResponse is a paginated list of situations.
type ListSituationsResponse struct {
	Items      []Situation
	NextCursor string
}

// Public service errors.
var (
	ErrSituationNotFound = errors.New("situation not found")
	ErrWordNotFound      = errors.New("word not found")
	ErrInvalidCursor     = errors.New("invalid cursor")
)

// Service implements the discovery and canonical word read operations.
type Service struct {
	repo   Repository
	reader SavedStateReader
}

// NewService creates a content service.
func NewService(repo Repository, reader SavedStateReader) *Service {
	return &Service{repo: repo, reader: reader}
}

// ListSituations returns active situations in display order, with the
// first page reordered (stable partition) so situations matching
// req.PriorityCategory surface first — see ListSituationsRequest.
func (s *Service) ListSituations(ctx context.Context, req ListSituationsRequest) (*ListSituationsResponse, error) {
	resp, err := s.repo.ListSituations(ctx, req)
	if err != nil {
		return nil, err
	}
	if req.PriorityCategory != "" && req.AfterCursor == "" {
		prioritizeByCategory(resp.Items, req.PriorityCategory)
	}
	return resp, nil
}

// prioritizeByCategory stable-partitions items in place so items whose
// Category matches priority come first, preserving their relative
// display_order within each partition. NextCursor is computed by the
// repository from the pre-partition ordering, so this reordering never
// affects pagination correctness.
func prioritizeByCategory(items []Situation, priority string) {
	matched := make([]Situation, 0, len(items))
	rest := make([]Situation, 0, len(items))
	for _, item := range items {
		if item.Category == priority {
			matched = append(matched, item)
		} else {
			rest = append(rest, item)
		}
	}
	copy(items, matched)
	copy(items[len(matched):], rest)
}

// GetSituation returns a situation with the requester's saved-state overlay.
func (s *Service) GetSituation(ctx context.Context, userID uuid.UUID, slug string) (*SituationDetail, error) {
	situation, err := s.repo.GetSituationBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, ErrSituationNotFound) {
			return nil, err
		}
		return nil, err
	}
	meanings, err := s.repo.GetMeaningsBySituation(ctx, situation.ID)
	if err != nil {
		return nil, err
	}
	if err := s.applySaved(ctx, userID, &meanings); err != nil {
		return nil, err
	}
	return &SituationDetail{Situation: *situation, Meanings: meanings}, nil
}

// GetWordDetail returns a canonical word with its meanings, examples, usage
// notes, and the requester's saved-state overlay.
func (s *Service) GetWordDetail(ctx context.Context, userID uuid.UUID, wordSlug string) (*WordDetail, error) {
	word, err := s.repo.GetWordBySlug(ctx, wordSlug)
	if err != nil {
		return nil, err
	}
	if err := s.applySavedWord(ctx, userID, word); err != nil {
		return nil, err
	}
	return word, nil
}

func (s *Service) applySaved(ctx context.Context, userID uuid.UUID, meanings *[]MeaningSummary) error {
	if s.reader == nil || userID == uuid.Nil || len(*meanings) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, len(*meanings))
	for i, m := range *meanings {
		ids[i] = m.MeaningID
	}
	states, err := s.reader.IsSaved(ctx, userID, ids)
	if err != nil {
		return err
	}
	for i := range *meanings {
		(*meanings)[i].Saved = states[(*meanings)[i].MeaningID]
	}
	return nil
}

func (s *Service) applySavedWord(ctx context.Context, userID uuid.UUID, word *WordDetail) error {
	if s.reader == nil || userID == uuid.Nil || len(word.Meanings) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, len(word.Meanings))
	for i, m := range word.Meanings {
		ids[i] = m.ID
	}
	states, err := s.reader.IsSaved(ctx, userID, ids)
	if err != nil {
		return err
	}
	userWordIDs, err := s.reader.SavedUserWordIDs(ctx, userID, ids)
	if err != nil {
		return err
	}
	for i := range word.Meanings {
		word.Meanings[i].Saved = states[word.Meanings[i].ID]
		if id, ok := userWordIDs[word.Meanings[i].ID]; ok {
			word.Meanings[i].UserWordID = id
		}
	}
	return nil
}
