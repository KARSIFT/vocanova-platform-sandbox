package learning

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryRepository is a deterministic in-memory repository for service and route
// tests. It is not concurrency-safe for racing cross-process requests.
type MemoryRepository struct {
	mu        sync.Mutex
	userWords []MemoryUserWord
	meanings  []MemoryMeaning
	words     []MemoryWord
}

// MemoryUserWord is an in-memory user_words row for tests.
type MemoryUserWord struct {
	ID                        uuid.UUID
	UserID                    uuid.UUID
	MeaningID                 uuid.UUID
	Status                    string
	Source                    string
	ReviewStep                int
	NextReviewAt              *time.Time
	LastReviewedAt            *time.Time
	LastResult                string
	LastRating                string
	ConsecutiveCorrectCount   int
	ConsecutiveIncorrectCount int
	TotalReviewCount          int
	CorrectReviewCount        int
	MasteredAt                *time.Time
	IgnoredAt                 *time.Time
	AddedAt                   time.Time
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	DeletedAt                 *time.Time
}

// MemoryMeaning is an in-memory canonical meaning for tests.
type MemoryMeaning struct {
	ID              uuid.UUID
	WordID          uuid.UUID
	PartOfSpeech    string
	ShortDefinition string
	Status          string
}

// MemoryWord is an in-memory canonical word for tests.
type MemoryWord struct {
	ID             uuid.UUID
	Text           string
	NormalizedText string
	Status         string
}

// MemoryRepositoryData holds seed data for the memory repository.
type MemoryRepositoryData struct {
	UserWords []MemoryUserWord
	Meanings  []MemoryMeaning
	Words     []MemoryWord
}

// NewMemoryRepository initializes an in-memory repository from data.
func NewMemoryRepository(data MemoryRepositoryData) *MemoryRepository {
	return &MemoryRepository{
		userWords: data.UserWords,
		meanings:  data.Meanings,
		words:     data.Words,
	}
}

func (r *MemoryRepository) SaveUserWord(ctx context.Context, req SaveUserWordRequest, now time.Time) (*SavedMeaning, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.meaningActive(req.MeaningID) {
		return nil, ErrMeaningNotFound
	}

	for i := range r.userWords {
		uw := &r.userWords[i]
		if uw.UserID == req.UserID && uw.MeaningID == req.MeaningID {
			if uw.DeletedAt == nil {
				return r.savedMeaningFromUserWord(*uw), nil
			}
			uw.DeletedAt = nil
			uw.Status = "new"
			uw.Source = req.Source
			uw.ReviewStep = 0
			uw.NextReviewAt = nil
			uw.LastReviewedAt = nil
			uw.LastResult = ""
			uw.LastRating = ""
			uw.ConsecutiveCorrectCount = 0
			uw.ConsecutiveIncorrectCount = 0
			uw.TotalReviewCount = 0
			uw.CorrectReviewCount = 0
			uw.MasteredAt = nil
			uw.IgnoredAt = nil
			uw.AddedAt = now
			uw.UpdatedAt = now
			return r.savedMeaningFromUserWord(*uw), nil
		}
	}

	uw := MemoryUserWord{
		ID:        uuid.New(),
		UserID:    req.UserID,
		MeaningID: req.MeaningID,
		Status:    "new",
		Source:    req.Source,
		AddedAt:   now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.userWords = append(r.userWords, uw)
	return r.savedMeaningFromUserWord(uw), nil
}

func (r *MemoryRepository) UnsaveUserWord(ctx context.Context, userID, meaningID uuid.UUID, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.userWords {
		uw := &r.userWords[i]
		if uw.UserID == userID && uw.MeaningID == meaningID && uw.DeletedAt == nil {
			uw.DeletedAt = &now
			uw.UpdatedAt = now
			return nil
		}
	}
	return ErrUserWordNotFound
}

func (r *MemoryRepository) GetSavedMeaning(ctx context.Context, userID, meaningID uuid.UUID) (*SavedMeaning, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, uw := range r.userWords {
		if uw.UserID == userID && uw.MeaningID == meaningID && uw.DeletedAt == nil {
			return r.savedMeaningFromUserWord(uw), nil
		}
	}
	return nil, ErrUserWordNotFound
}

func (r *MemoryRepository) ListSavedWords(ctx context.Context, req ListSavedWordsRequest) (*ListSavedWordsResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	items := r.activeSavedMeanings(req.UserID)
	sort.Slice(items, func(i, j int) bool {
		if items[i].AddedAt.Equal(items[j].AddedAt) {
			return items[i].UserWordID.String() > items[j].UserWordID.String()
		}
		return items[i].AddedAt.After(items[j].AddedAt)
	})

	start := 0
	if req.AfterCursor != "" {
		c, err := decodeSavedCursor(req.AfterCursor)
		if err != nil {
			return nil, ErrInvalidCursor
		}
		for i, it := range items {
			if it.AddedAt.Before(c.AddedAt) || (it.AddedAt.Equal(c.AddedAt) && it.UserWordID.String() <= c.ID.String()) {
				start = i + 1
				break
			}
		}
	}
	if start > len(items) {
		start = len(items)
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}
	page := items[start:end]
	resp := &ListSavedWordsResponse{Items: page}
	if len(page) == limit && end < len(items) {
		last := page[len(page)-1]
		resp.NextCursor = encodeSavedCursor(savedCursor{AddedAt: last.AddedAt, ID: last.UserWordID})
	}
	return resp, nil
}

// SavedUserWordIDs implements SavedStateReader.
func (r *MemoryRepository) SavedUserWordIDs(ctx context.Context, userID uuid.UUID, meaningIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make(map[uuid.UUID]uuid.UUID, len(meaningIDs))
	for _, id := range meaningIDs {
		for _, uw := range r.userWords {
			if uw.UserID == userID && uw.MeaningID == id && uw.DeletedAt == nil {
				out[id] = uw.ID
				break
			}
		}
	}
	return out, nil
}

// IsSaved implements SavedStateReader.
func (r *MemoryRepository) IsSaved(ctx context.Context, userID uuid.UUID, meaningIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make(map[uuid.UUID]bool, len(meaningIDs))
	for _, id := range meaningIDs {
		out[id] = false
		for _, uw := range r.userWords {
			if uw.UserID == userID && uw.MeaningID == id && uw.DeletedAt == nil {
				out[id] = true
				break
			}
		}
	}
	return out, nil
}

func (r *MemoryRepository) meaningActive(id uuid.UUID) bool {
	for _, m := range r.meanings {
		if m.ID == id && m.Status == "active" {
			return true
		}
	}
	return false
}

func (r *MemoryRepository) activeSavedMeanings(userID uuid.UUID) []SavedMeaning {
	var out []SavedMeaning
	for _, uw := range r.userWords {
		if uw.UserID != userID || uw.DeletedAt != nil {
			continue
		}
		m, ok := r.findMeaning(uw.MeaningID)
		if !ok {
			continue
		}
		w, ok := r.findWord(m.WordID)
		if !ok {
			continue
		}
		out = append(out, SavedMeaning{
			UserWordID:      uw.ID,
			MeaningID:       m.ID,
			WordID:          w.ID,
			WordText:        w.Text,
			WordSlug:        wordSlug(w.NormalizedText),
			PartOfSpeech:    m.PartOfSpeech,
			ShortDefinition: m.ShortDefinition,
			Status:          uw.Status,
			Source:          uw.Source,
			Saved:           true,
			AddedAt:         uw.AddedAt,
		})
	}
	return out
}

func (r *MemoryRepository) findMeaning(id uuid.UUID) (MemoryMeaning, bool) {
	for _, m := range r.meanings {
		if m.ID == id {
			return m, true
		}
	}
	return MemoryMeaning{}, false
}

func (r *MemoryRepository) findWord(id uuid.UUID) (MemoryWord, bool) {
	for _, w := range r.words {
		if w.ID == id {
			return w, true
		}
	}
	return MemoryWord{}, false
}

func (r *MemoryRepository) savedMeaningFromUserWord(uw MemoryUserWord) *SavedMeaning {
	m, ok := r.findMeaning(uw.MeaningID)
	if !ok {
		return nil
	}
	w, ok := r.findWord(m.WordID)
	if !ok {
		return nil
	}
	return &SavedMeaning{
		UserWordID:      uw.ID,
		MeaningID:       m.ID,
		WordID:          w.ID,
		WordText:        w.Text,
		WordSlug:        wordSlug(w.NormalizedText),
		PartOfSpeech:    m.PartOfSpeech,
		ShortDefinition: m.ShortDefinition,
		Status:          uw.Status,
		Source:          uw.Source,
		Saved:           uw.DeletedAt == nil,
		AddedAt:         uw.AddedAt,
	}
}

func wordSlug(normalized string) string {
	// Mirrors the canonical slug derivation in the content module.
	out := ""
	for _, r := range normalized {
		if r == ' ' {
			out += "-"
		} else {
			out += string(r)
		}
	}
	return out
}
