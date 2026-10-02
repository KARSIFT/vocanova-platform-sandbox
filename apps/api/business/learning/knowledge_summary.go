package learning

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// KnowledgeSummary counts saved meanings, not distinct words or proficiency.
// Saved includes every undeleted status; Due overlaps new/learning/reviewing
// and uses the same eligibility condition as the review queue, without a daily cap.
type KnowledgeSummary struct {
	SelfReportedKnown int
	Saved             int
	New               int
	Learning          int
	Reviewing         int
	Mastered          int
	Ignored           int
	Archived          int
	Due               int
}

func (s *Service) GetKnowledgeSummary(ctx context.Context, userID uuid.UUID) (*KnowledgeSummary, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user id required")
	}
	summary, err := s.repo.GetKnowledgeSummary(ctx, userID)
	if err != nil {
		return nil, err
	}
	if s.knowledge != nil {
		summary.SelfReportedKnown, err = s.knowledge.CountKnown(ctx, userID)
		if err != nil {
			return nil, err
		}
	}
	return summary, nil
}

func (r *PostgreSQLRepository) GetKnowledgeSummary(ctx context.Context, userID uuid.UUID) (*KnowledgeSummary, error) {
	var summary KnowledgeSummary
	err := r.db.QueryRowContext(ctx, `SELECT count(*),
	 count(*) FILTER (WHERE uw.status='new' AND uw.total_review_count=0),
	 count(*) FILTER (WHERE uw.status='learning' OR (uw.status='new' AND uw.total_review_count>0)),
	 count(*) FILTER (WHERE uw.status='reviewing'),
	 count(*) FILTER (WHERE uw.status='mastered'),
	 count(*) FILTER (WHERE uw.status='ignored'),
	 count(*) FILTER (WHERE uw.status='archived'),
	 count(*) FILTER (WHERE uw.status IN ('new','learning','reviewing') AND (uw.next_review_at IS NULL OR uw.next_review_at <= CURRENT_TIMESTAMP))
	 FROM user_words uw JOIN word_meanings wm ON wm.id=uw.meaning_id JOIN canonical_words cw ON cw.id=wm.word_id
	 WHERE uw.user_id=$1 AND uw.deleted_at IS NULL`, userID).Scan(&summary.Saved, &summary.New, &summary.Learning, &summary.Reviewing, &summary.Mastered, &summary.Ignored, &summary.Archived, &summary.Due)
	if err != nil {
		return nil, fmt.Errorf("get knowledge summary: %w", err)
	}
	return &summary, nil
}

func (r *MemoryRepository) GetKnowledgeSummary(ctx context.Context, userID uuid.UUID) (*KnowledgeSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	var summary KnowledgeSummary
	for _, word := range r.userWords {
		if word.UserID != userID || word.DeletedAt != nil {
			continue
		}
		// Match the canonical inner joins in the saved list and due queue, without
		// hiding an existing saved state when canonical content is later archived.
		found := false
		for _, meaning := range r.meanings {
			if meaning.ID == word.MeaningID {
				for _, canonical := range r.words {
					if canonical.ID == meaning.WordID {
						found = true
					}
				}
			}
		}
		if !found {
			continue
		}
		summary.Saved++
		switch reviewState(word) {
		case "new":
			summary.New++
		case "learning":
			summary.Learning++
		case "reviewing":
			summary.Reviewing++
		case "mastered":
			summary.Mastered++
		case "ignored":
			summary.Ignored++
		case "archived":
			summary.Archived++
		}
		if (word.Status == "new" || word.Status == "learning" || word.Status == "reviewing") && (word.NextReviewAt == nil || !word.NextReviewAt.After(now)) {
			summary.Due++
		}
	}
	return &summary, nil
}

// KnowledgeReader counts explicit self-assessments independently of saved stages.
type KnowledgeReader interface {
	CountKnown(context.Context, uuid.UUID) (int, error)
}

func (s *Service) SetKnowledgeReader(reader KnowledgeReader) { s.knowledge = reader }
