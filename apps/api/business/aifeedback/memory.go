package aifeedback

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryRepository is a deterministic, in-process repository for service and
// route tests. It is not safe for concurrent cross-process requests.
type MemoryRepository struct {
	mu             sync.Mutex
	userWords      []MemoryUserWord
	reviewAttempts []MemoryReviewAttempt
	meanings       []MemoryMeaning
	words          []MemoryWord
	sentences      []MemoryLearnerSentence
	attempts       []MemoryAIFeedbackAttempt
	reports        []QualityReviewReport
}

// MemoryUserWord mirrors the user_words row.
type MemoryUserWord struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	MeaningID uuid.UUID
	Status    string
	DeletedAt *time.Time
}

// MemoryReviewAttempt mirrors the review_attempts row.
type MemoryReviewAttempt struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	UserWordID uuid.UUID
	MeaningID  uuid.UUID
}

// MemoryMeaning mirrors the word_meanings row.
type MemoryMeaning struct {
	ID              uuid.UUID
	WordID          uuid.UUID
	PartOfSpeech    string
	ShortDefinition string
	Status          string
}

// MemoryWord mirrors the canonical_words row.
type MemoryWord struct {
	ID              uuid.UUID
	Text            string
	NormalizedText  string
	WordType        string
	DifficultyLevel string
	Status          string
}

// MemoryLearnerSentence mirrors the learner_sentences row.
type MemoryLearnerSentence struct {
	ID                     uuid.UUID
	UserID                 uuid.UUID
	MeaningID              uuid.UUID
	UserWordID             uuid.UUID
	SentenceText           string
	NormalizedSentenceText string
	Source                 string
	Status                 string
	SubmittedAt            time.Time
}

// MemoryAIFeedbackAttempt mirrors the ai_feedback_attempts row.
type MemoryAIFeedbackAttempt struct {
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

// MemoryRepositoryData holds seed data for the memory repository.
type MemoryRepositoryData struct {
	UserWords      []MemoryUserWord
	ReviewAttempts []MemoryReviewAttempt
	Meanings       []MemoryMeaning
	Words          []MemoryWord
}

// NewMemoryRepository initializes an in-memory repository from seed data.
func NewMemoryRepository(data MemoryRepositoryData) *MemoryRepository {
	return &MemoryRepository{
		userWords:      data.UserWords,
		reviewAttempts: data.ReviewAttempts,
		meanings:       data.Meanings,
		words:          data.Words,
	}
}

// LoadTarget implements Repository.
func (r *MemoryRepository) LoadTarget(ctx context.Context, req LoadTargetRequest) (*Target, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch req.Source {
	case SourceWordDetail:
		return r.loadTargetFromUserWord(req.UserID, req.AttemptID)
	case SourceReview:
		return r.loadTargetFromReviewAttempt(req.UserID, req.AttemptID)
	case SourceDailyMission, SourceFreePractice:
		return nil, ErrTargetNotFound
	default:
		return nil, ErrTargetNotFound
	}
}

func (r *MemoryRepository) loadTargetFromUserWord(userID, userWordID uuid.UUID) (*Target, error) {
	for _, uw := range r.userWords {
		if uw.ID == userWordID && uw.UserID == userID && uw.DeletedAt == nil {
			return r.buildTarget(uw.MeaningID, userWordID, nil)
		}
	}
	return nil, ErrTargetNotFound
}

func (r *MemoryRepository) loadTargetFromReviewAttempt(userID, reviewAttemptID uuid.UUID) (*Target, error) {
	for _, ra := range r.reviewAttempts {
		if ra.ID == reviewAttemptID && ra.UserID == userID {
			return r.buildTarget(ra.MeaningID, ra.UserWordID, &reviewAttemptID)
		}
	}
	return nil, ErrTargetNotFound
}

func (r *MemoryRepository) buildTarget(meaningID, userWordID uuid.UUID, reviewAttemptID *uuid.UUID) (*Target, error) {
	var m *MemoryMeaning
	for i := range r.meanings {
		if r.meanings[i].ID == meaningID {
			m = &r.meanings[i]
			break
		}
	}
	if m == nil {
		return nil, ErrTargetNotFound
	}
	var w *MemoryWord
	for i := range r.words {
		if r.words[i].ID == m.WordID {
			w = &r.words[i]
			break
		}
	}
	if w == nil {
		return nil, ErrTargetNotFound
	}

	return &Target{
		WordID:          w.ID,
		MeaningID:       m.ID,
		UserWordID:      userWordID,
		ReviewAttemptID: reviewAttemptID,
		WordText:        w.Text,
		NormalizedWord:  w.NormalizedText,
		WordType:        w.WordType,
		PartOfSpeech:    m.PartOfSpeech,
		ShortDefinition: m.ShortDefinition,
		LearnerLevel:    learnerLevel(w.DifficultyLevel),
		AcceptedForms:   BuildAcceptedForms(w.NormalizedText, w.WordType, m.PartOfSpeech),
	}, nil
}

// GetFeedbackAttemptByRequestHash implements Repository.
func (r *MemoryRepository) GetFeedbackAttemptByRequestHash(ctx context.Context, requestHash string) (*StoredFeedbackAttempt, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, a := range r.attempts {
		if a.RequestHash == requestHash && a.Status == AttemptStatusSucceeded {
			return r.toStoredAttempt(a), nil
		}
	}
	for _, a := range r.attempts {
		if a.RequestHash == requestHash && a.Status == AttemptStatusPending {
			return r.toStoredAttempt(a), nil
		}
	}
	for i := len(r.attempts) - 1; i >= 0; i-- {
		if r.attempts[i].RequestHash == requestHash {
			return r.toStoredAttempt(r.attempts[i]), nil
		}
	}
	return nil, nil
}

// CreatePendingAttempt implements Repository.
func (r *MemoryRepository) CreatePendingAttempt(ctx context.Context, req SubmitSentenceFeedbackRequest, target *Target, normalized string, requestHash string, provider string, model string, now time.Time) (*PendingAttempt, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	sentenceID := uuid.New()
	attemptID := uuid.New()

	r.sentences = append(r.sentences, MemoryLearnerSentence{
		ID:                     sentenceID,
		UserID:                 req.UserID,
		MeaningID:              target.MeaningID,
		UserWordID:             target.UserWordID,
		SentenceText:           req.SentenceText,
		NormalizedSentenceText: normalized,
		Source:                 req.Source,
		Status:                 SentenceStatusSubmitted,
		SubmittedAt:            now,
	})

	r.attempts = append(r.attempts, MemoryAIFeedbackAttempt{
		ID:                attemptID,
		LearnerSentenceID: sentenceID,
		Status:            AttemptStatusPending,
		Provider:          provider,
		Model:             model,
		PromptVersion:     PromptVersionSentenceFeedbackV1,
		RequestHash:       requestHash,
	})

	return &PendingAttempt{SentenceID: sentenceID, AttemptID: attemptID}, nil
}

func (r *MemoryRepository) CreateRetryAttempt(ctx context.Context, failed *StoredFeedbackAttempt, provider string, model string, now time.Time) (*RetryAttempt, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Keep the deterministic repository subject to the same active-generation
	// invariant as PostgreSQL's partial unique index. A concurrent fresh key
	// returns the extant pending/succeeded result rather than adding a second
	// generation that could call the provider.
	for _, attempt := range r.attempts {
		if attempt.RequestHash == failed.RequestHash && (attempt.Status == AttemptStatusPending || attempt.Status == AttemptStatusSucceeded) {
			return &RetryAttempt{Existing: r.toStoredAttempt(attempt)}, nil
		}
	}

	attemptID := uuid.New()
	r.attempts = append(r.attempts, MemoryAIFeedbackAttempt{
		ID:                attemptID,
		LearnerSentenceID: failed.LearnerSentenceID,
		Status:            AttemptStatusPending,
		Provider:          provider,
		Model:             model,
		PromptVersion:     PromptVersionSentenceFeedbackV1,
		RequestHash:       failed.RequestHash,
	})
	for i := range r.sentences {
		if r.sentences[i].ID == failed.LearnerSentenceID {
			r.sentences[i].Status = SentenceStatusSubmitted
			break
		}
	}
	return &RetryAttempt{Pending: &PendingAttempt{SentenceID: failed.LearnerSentenceID, AttemptID: attemptID}}, nil
}

// GetFeedbackAttemptOwner implements Repository.
func (r *MemoryRepository) GetFeedbackAttemptOwner(ctx context.Context, attemptID uuid.UUID) (uuid.UUID, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, a := range r.attempts {
		if a.ID == attemptID {
			for _, s := range r.sentences {
				if s.ID == a.LearnerSentenceID {
					return s.UserID, nil
				}
			}
		}
	}
	return uuid.Nil, ErrTargetNotFound
}

// CreateQualityReviewReport implements Repository.
func (r *MemoryRepository) CreateQualityReviewReport(ctx context.Context, report QualityReviewReport) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.reports {
		if existing.AttemptID == report.AttemptID {
			if existing.Reason != report.Reason {
				return false, ErrReportIdempotencyConflict
			}
			return false, nil
		}
	}
	r.reports = append(r.reports, report)
	return true, nil
}

// QualityReviewReports returns a copy for deterministic tests.
func (r *MemoryRepository) QualityReviewReports() []QualityReviewReport {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]QualityReviewReport(nil), r.reports...)
}

// CompleteFeedbackAttempt implements Repository.
func (r *MemoryRepository) CompleteFeedbackAttempt(ctx context.Context, pending PendingAttempt, feedback *ProviderFeedback, failureCode, failureMessage string, now time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	updated := false
	for i := range r.attempts {
		if r.attempts[i].ID == pending.AttemptID && r.attempts[i].Status == AttemptStatusPending {
			updated = true
			if feedback != nil {
				r.attempts[i].Status = AttemptStatusSucceeded
				r.attempts[i].FeedbackJSON = feedback.RawJSON
				r.attempts[i].FeedbackText = feedback.Explanation
			} else {
				code := failureCode
				if code == "" {
					code = ErrorCodeTemporaryFailure
				}
				r.attempts[i].Status = AttemptStatusFailed
				r.attempts[i].ErrorCode = code
				r.attempts[i].ErrorMessage = failureMessage
			}
		}
	}

	if !updated {
		return nil
	}
	for i := range r.sentences {
		if r.sentences[i].ID == pending.SentenceID {
			if feedback != nil {
				r.sentences[i].Status = SentenceStatusFeedbackReady
			} else {
				r.sentences[i].Status = SentenceStatusFeedbackFailed
			}
		}
	}

	return nil
}

func (r *MemoryRepository) toStoredAttempt(a MemoryAIFeedbackAttempt) *StoredFeedbackAttempt {
	reported := false
	for _, report := range r.reports {
		if report.AttemptID == a.ID {
			reported = true
			break
		}
	}
	return &StoredFeedbackAttempt{
		Reported:          reported,
		ID:                a.ID,
		LearnerSentenceID: a.LearnerSentenceID,
		Status:            a.Status,
		Provider:          a.Provider,
		Model:             a.Model,
		PromptVersion:     a.PromptVersion,
		RequestHash:       a.RequestHash,
		FeedbackJSON:      a.FeedbackJSON,
		FeedbackText:      a.FeedbackText,
		ErrorCode:         a.ErrorCode,
		ErrorMessage:      a.ErrorMessage,
	}
}

// compile-time interface check.
var _ Repository = (*MemoryRepository)(nil)
