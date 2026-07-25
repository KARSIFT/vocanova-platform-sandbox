package reviews

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
	ID           uuid.UUID
	UserID       uuid.UUID
	MeaningID    uuid.UUID
	Status       string
	Source       string
	ReviewStep   int
	NextReviewAt *time.Time
	AddedAt      time.Time
	DeletedAt    *time.Time
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

func (r *MemoryRepository) ListDueWords(ctx context.Context, req ListDueWordsRequest) (*ListDueWordsResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	items := r.activeDueWords(req.UserID)
	sort.Slice(items, func(i, j int) bool {
		leftNull := items[i].NextReviewAt == nil
		rightNull := items[j].NextReviewAt == nil
		if leftNull && !rightNull {
			return true
		}
		if !leftNull && rightNull {
			return false
		}
		if !leftNull && !rightNull {
			if !items[i].NextReviewAt.Equal(*items[j].NextReviewAt) {
				return items[i].NextReviewAt.Before(*items[j].NextReviewAt)
			}
		}
		return items[i].UserWordID.String() < items[j].UserWordID.String()
	})

	start := 0
	if req.AfterCursor != "" {
		c, err := decodeDueCursor(req.AfterCursor)
		if err != nil {
			return nil, ErrInvalidCursor
		}
		for i, it := range items {
			cursorTime := c.NextReviewAt
			itTimeZero := it.NextReviewAt == nil
			if itTimeZero && cursorTime.IsZero() {
				if it.UserWordID.String() > c.ID.String() {
					start = i
					break
				}
				continue
			}
			if !itTimeZero && !cursorTime.IsZero() {
				if it.NextReviewAt.After(cursorTime) ||
					(it.NextReviewAt.Equal(cursorTime) && it.UserWordID.String() > c.ID.String()) {
					start = i
					break
				}
			}
			if !itTimeZero && cursorTime.IsZero() {
				start = i
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
	resp := &ListDueWordsResponse{Items: page, TotalCount: len(items)}
	if len(page) == limit && end < len(items) {
		last := page[len(page)-1]
		cursor := dueCursor{ID: last.UserWordID}
		if last.NextReviewAt != nil {
			cursor.NextReviewAt = *last.NextReviewAt
		}
		resp.NextCursor = encodeDueCursor(cursor)
	}
	return resp, nil
}

func (r *MemoryRepository) activeDueWords(userID uuid.UUID) []DueWord {
	now := time.Now().UTC()
	var out []DueWord
	for _, uw := range r.userWords {
		if uw.UserID != userID || uw.DeletedAt != nil {
			continue
		}
		if !isDueStatus(uw.Status) {
			continue
		}
		if uw.NextReviewAt != nil && uw.NextReviewAt.After(now) {
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
		d := DueWord{
			UserWordID:      uw.ID,
			MeaningID:       m.ID,
			WordID:          w.ID,
			WordText:        w.Text,
			WordSlug:        wordSlug(w.NormalizedText),
			PartOfSpeech:    m.PartOfSpeech,
			ShortDefinition: m.ShortDefinition,
			Status:          uw.Status,
			ReviewStep:      uw.ReviewStep,
			Source:          uw.Source,
		}
		if uw.NextReviewAt != nil {
			cp := *uw.NextReviewAt
			d.NextReviewAt = &cp
		}
		out = append(out, d)
	}
	return out
}

func isDueStatus(status string) bool {
	switch status {
	case "new", "learning", "reviewing":
		return true
	}
	return false
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
