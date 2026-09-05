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
	attempts  []MemoryReviewAttempt
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
	TotalReviewCount          int
	CorrectReviewCount        int
	ConsecutiveCorrectCount   int
	ConsecutiveIncorrectCount int
	LastReviewedAt            *time.Time
	LastResult                string
	LastRating                string
	AddedAt                   time.Time
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

// MemoryReviewAttempt is an in-memory review_attempts row for tests.
type MemoryReviewAttempt struct {
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
	if limit > 50 {
		limit = 50
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
		start = len(items)
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

func (r *MemoryRepository) GetReviewAttemptByClientAttemptID(ctx context.Context, userID uuid.UUID, clientAttemptID string) (*ReviewAttempt, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, a := range r.attempts {
		if a.UserID == userID && a.ClientAttemptID == clientAttemptID {
			return r.toReviewAttempt(a), nil
		}
	}
	return nil, ErrReviewAttemptNotFound
}

func (r *MemoryRepository) SubmitReview(ctx context.Context, req SubmitReviewRequest) (*ReviewAttempt, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for an existing attempt by (user_id, client_attempt_id).
	for i, a := range r.attempts {
		if a.UserID == req.UserID && a.ClientAttemptID == req.ClientAttemptID {
			if reviewAttemptsEqual(a, req) {
				return r.toReviewAttempt(r.attempts[i]), nil
			}
			return nil, ErrIdempotencyConflict
		}
	}

	idx, ok := r.findActiveUserWord(req.UserWordID)
	if !ok || r.userWords[idx].UserID != req.UserID {
		return nil, ErrUserWordNotFound
	}
	if r.userWords[idx].MeaningID != req.MeaningID {
		return nil, ErrUserWordNotFound
	}

	uw := &r.userWords[idx]
	prior := ReviewState{
		ReviewStep: uw.ReviewStep,
		Counters: ReviewCounters{
			TotalReviewCount:          uw.TotalReviewCount,
			CorrectReviewCount:        uw.CorrectReviewCount,
			ConsecutiveCorrectCount:   uw.ConsecutiveCorrectCount,
			ConsecutiveIncorrectCount: uw.ConsecutiveIncorrectCount,
		},
	}
	scheduledAt := req.scheduledAt
	if scheduledAt.IsZero() {
		scheduledAt = time.Now().UTC()
	}
	sched, err := ApplyReview(prior, ApplyReviewRequest{Result: req.Result, Rating: req.Rating}, scheduledAt)
	if err != nil {
		return nil, err
	}

	attempt := MemoryReviewAttempt{
		ID:                      uuid.New(),
		UserID:                  req.UserID,
		UserWordID:              req.UserWordID,
		MeaningID:               req.MeaningID,
		AttemptType:             req.AttemptType,
		PromptType:              req.PromptType,
		Result:                  req.Result,
		Rating:                  req.Rating,
		ReviewStepBefore:        prior.ReviewStep,
		ReviewStepAfter:         sched.ReviewStep,
		AnsweredAt:              req.AnsweredAt,
		ResponseTimeMs:          req.ResponseTimeMs,
		SelectedOptionMeaningID: req.SelectedOptionMeaningID,
		TypedAnswer:             req.TypedAnswer,
		WasHintUsed:             req.WasHintUsed,
		Source:                  req.Source,
		ClientAttemptID:         req.ClientAttemptID,
		Metadata:                req.Metadata,
	}
	r.attempts = append(r.attempts, attempt)

	uw.ReviewStep = sched.ReviewStep
	uw.NextReviewAt = &sched.NextReviewAt
	uw.LastReviewedAt = &sched.LastReviewedAt
	uw.LastResult = sched.LastResult
	uw.LastRating = sched.LastRating
	uw.TotalReviewCount = sched.Counters.TotalReviewCount
	uw.CorrectReviewCount = sched.Counters.CorrectReviewCount
	uw.ConsecutiveCorrectCount = sched.Counters.ConsecutiveCorrectCount
	uw.ConsecutiveIncorrectCount = sched.Counters.ConsecutiveIncorrectCount

	return r.toReviewAttempt(attempt), nil
}

func (r *MemoryRepository) findActiveUserWord(id uuid.UUID) (int, bool) {
	for i, uw := range r.userWords {
		if uw.ID == id && uw.DeletedAt == nil {
			return i, true
		}
	}
	return 0, false
}

func (r *MemoryRepository) toReviewAttempt(a MemoryReviewAttempt) *ReviewAttempt {
	var nextReviewAt time.Time
	if idx, ok := r.findActiveUserWord(a.UserWordID); ok {
		if r.userWords[idx].NextReviewAt != nil {
			nextReviewAt = *r.userWords[idx].NextReviewAt
		}
	}
	return &ReviewAttempt{
		ID:                      a.ID,
		UserID:                  a.UserID,
		UserWordID:              a.UserWordID,
		MeaningID:               a.MeaningID,
		AttemptType:             a.AttemptType,
		PromptType:              a.PromptType,
		Result:                  a.Result,
		Rating:                  a.Rating,
		ReviewStepBefore:        a.ReviewStepBefore,
		ReviewStepAfter:         a.ReviewStepAfter,
		AnsweredAt:              a.AnsweredAt,
		ResponseTimeMs:          a.ResponseTimeMs,
		SelectedOptionMeaningID: a.SelectedOptionMeaningID,
		TypedAnswer:             a.TypedAnswer,
		WasHintUsed:             a.WasHintUsed,
		Source:                  a.Source,
		ClientAttemptID:         a.ClientAttemptID,
		Metadata:                a.Metadata,
		NextReviewAt:            nextReviewAt,
	}
}

func reviewAttemptsEqual(a MemoryReviewAttempt, req SubmitReviewRequest) bool {
	if a.UserWordID != req.UserWordID || a.MeaningID != req.MeaningID {
		return false
	}
	if a.AttemptType != req.AttemptType || a.PromptType != req.PromptType || a.Result != req.Result || a.Rating != req.Rating {
		return false
	}
	if a.ResponseTimeMs != req.ResponseTimeMs || a.WasHintUsed != req.WasHintUsed || a.Source != req.Source {
		return false
	}
	if !a.AnsweredAt.Equal(req.AnsweredAt) {
		return false
	}
	if !ptrUUIDEqual(a.SelectedOptionMeaningID, req.SelectedOptionMeaningID) {
		return false
	}
	if !ptrStringEqual(a.TypedAnswer, req.TypedAnswer) {
		return false
	}
	return true
}

func ptrUUIDEqual(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func ptrStringEqual(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func wordSlug(normalized string) string {
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

// compile-time interface check.
var _ Repository = (*MemoryRepository)(nil)
