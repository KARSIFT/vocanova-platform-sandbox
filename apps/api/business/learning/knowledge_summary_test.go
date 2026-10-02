package learning

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeSummaryCountsFullSetAndDueEligibility(t *testing.T) {
	now := time.Now().UTC()
	past, future := now.Add(-time.Hour), now.Add(time.Hour)
	owner, other, wordID, meaningID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	rows := []MemoryUserWord{}
	for i := 0; i < 55; i++ {
		rows = append(rows, MemoryUserWord{ID: uuid.New(), UserID: owner, MeaningID: meaningID, Status: "new"})
	}
	rows = append(rows,
		MemoryUserWord{ID: uuid.New(), UserID: owner, MeaningID: meaningID, Status: "new", TotalReviewCount: 1, NextReviewAt: &future},
		MemoryUserWord{ID: uuid.New(), UserID: owner, MeaningID: meaningID, Status: "learning", NextReviewAt: &now},
		MemoryUserWord{ID: uuid.New(), UserID: owner, MeaningID: meaningID, Status: "reviewing", NextReviewAt: &past},
		MemoryUserWord{ID: uuid.New(), UserID: owner, MeaningID: meaningID, Status: "reviewing", NextReviewAt: &future},
		MemoryUserWord{ID: uuid.New(), UserID: owner, MeaningID: meaningID, Status: "mastered", NextReviewAt: &past},
		MemoryUserWord{ID: uuid.New(), UserID: owner, MeaningID: meaningID, Status: "ignored"},
		MemoryUserWord{ID: uuid.New(), UserID: owner, MeaningID: meaningID, Status: "archived"},
		MemoryUserWord{ID: uuid.New(), UserID: owner, MeaningID: meaningID, Status: "learning", DeletedAt: &now},
		MemoryUserWord{ID: uuid.New(), UserID: other, MeaningID: meaningID, Status: "new"},
	)
	meanings := make([]MemoryMeaning, len(rows))
	for i := range rows {
		rows[i].MeaningID = uuid.New()
		meanings[i] = MemoryMeaning{ID: rows[i].MeaningID, WordID: wordID, Status: "archived"}
	}
	repo := NewMemoryRepository(MemoryRepositoryData{UserWords: rows, Words: []MemoryWord{{ID: wordID, Status: "archived"}}, Meanings: meanings})
	svc := NewService(repo, nil, nil)
	summary, err := svc.GetKnowledgeSummary(t.Context(), owner)
	require.NoError(t, err)
	assert.Equal(t, &KnowledgeSummary{Saved: 62, New: 55, Learning: 2, Reviewing: 2, Mastered: 1, Ignored: 1, Archived: 1, Due: 57}, summary)
	assert.Equal(t, summary.Saved, summary.New+summary.Learning+summary.Reviewing+summary.Mastered+summary.Ignored+summary.Archived)
	empty, err := svc.GetKnowledgeSummary(t.Context(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, &KnowledgeSummary{}, empty)
}
