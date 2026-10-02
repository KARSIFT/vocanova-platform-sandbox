// Package achievements projects milestones from durable qualified learning
// history. It writes no badges, points, SRS state or mission progress.
package achievements

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"time"
)

var ErrNotFound = errors.New("active learner not found")

const CatalogVersion = "1"

type Achievement struct {
	ID          string     `json:"id"`
	Label       string     `json:"label"`
	Description string     `json:"description"`
	Category    string     `json:"category" enum:"lessons,practice,reviews,writing"`
	Criterion   string     `json:"criterion" enum:"participation,unaided_recall"`
	Current     int        `json:"current"`
	Target      int        `json:"target"`
	Earned      bool       `json:"earned"`
	EarnedAt    *time.Time `json:"earnedAt,omitempty"`
}
type AchievementList struct {
	CatalogVersion string        `json:"catalogVersion"`
	Items          []Achievement `json:"items"`
}
type Metric struct {
	Count      int
	Thresholds map[int]time.Time
}
type Repository interface {
	Metrics(context.Context, uuid.UUID) (map[string]Metric, error)
}
type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo} }

type definition struct {
	id, label, description, category, criterion, metric string
	target                                              int
}

var catalog = []definition{
	{"guided-first", "First lesson", "Complete one guided lesson. This celebrates participation, not mastery.", "lessons", "participation", "lessons", 1},
	{"guided-three", "Three lessons", "Complete three different guided lessons.", "lessons", "participation", "lessons", 3},
	{"guided-seven", "Seven lessons", "Complete seven different guided lessons.", "lessons", "participation", "lessons", 7},
	{"recall-independent-first", "Recall on your own", "Complete a typed recall or mistake practice with every answer correct on the first try, without showing answers.", "practice", "unaided_recall", "independent", 1},
	{"practice-five", "Keep practising", "Complete five practice sessions. Practice with help counts; this celebrates participation.", "practice", "participation", "practice", 5},
	{"reviews-ten", "Ten reviews", "Answer ten scheduled review prompts. Correct answers and mistakes both count; skipped prompts do not.", "reviews", "participation", "reviews", 10},
	{"reviews-fifty", "Fifty reviews", "Answer fifty scheduled review prompts. This celebrates returning to practise, not language proficiency.", "reviews", "participation", "reviews", 50},
	{"writing-first", "First sentence checked", "Write a sentence and receive saved feedback. Any completed feedback verdict counts, including advice to improve.", "writing", "participation", "writing", 1},
}

func (s *Service) List(ctx context.Context, u uuid.UUID) (*AchievementList, error) {
	if u == uuid.Nil {
		return nil, ErrNotFound
	}
	metrics, err := s.repo.Metrics(ctx, u)
	if err != nil {
		return nil, err
	}
	out := &AchievementList{CatalogVersion: CatalogVersion, Items: make([]Achievement, 0, len(catalog))}
	for _, d := range catalog {
		m := metrics[d.metric]
		n := m.Count
		if n < 0 {
			n = 0
		}
		if n > d.target {
			n = d.target
		}
		a := Achievement{ID: d.id, Label: d.label, Description: d.description, Category: d.category, Criterion: d.criterion, Current: n, Target: d.target}
		if at, ok := m.Thresholds[d.target]; ok && m.Count >= d.target && !at.IsZero() {
			a.Earned = true
			t := at.UTC()
			a.EarnedAt = &t
		}
		out.Items = append(out.Items, a)
	}
	return out, nil
}
