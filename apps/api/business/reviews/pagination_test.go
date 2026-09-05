package reviews

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestDuePaginationAdvancesAndExhausts(t *testing.T) {
	for _, mixed := range []bool{false, true} {
		t.Run(map[bool]string{false: "unscheduled", true: "mixed_and_tied"}[mixed], func(t *testing.T) {
			userID, otherID, meaningID, wordID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
			past := time.Now().Add(-time.Hour)
			data := MemoryRepositoryData{
				Meanings: []MemoryMeaning{{ID: meaningID, WordID: wordID}},
				Words:    []MemoryWord{{ID: wordID, Text: "work", NormalizedText: "work"}},
			}
			for i := 1; i <= 6; i++ {
				id := uuid.UUID{15: byte(i)}
				word := MemoryUserWord{ID: id, UserID: userID, MeaningID: meaningID, Status: "new"}
				if mixed && i > 2 {
					word.NextReviewAt = &past
				}
				data.UserWords = append(data.UserWords, word)
			}
			data.UserWords = append(data.UserWords, MemoryUserWord{ID: uuid.New(), UserID: otherID, MeaningID: meaningID, Status: "new"})
			repo := NewMemoryRepository(data)
			cursor := ""
			var last DueWord
			for page := 0; page < 3; page++ {
				resp, err := repo.ListDueWords(context.Background(), ListDueWordsRequest{UserID: userID, Limit: 2, AfterCursor: cursor})
				if err != nil {
					t.Fatal(err)
				}
				if len(resp.Items) != 2 || resp.TotalCount != 6 {
					t.Fatalf("page %d: %+v", page, resp)
				}
				for i, item := range resp.Items {
					if item.UserWordID != (uuid.UUID{15: byte(page*2 + i + 1)}) {
						t.Fatalf("repeated or out-of-order item: %s", item.UserWordID)
					}
				}
				if (resp.NextCursor != "") != (page < 2) {
					t.Fatalf("wrong continuation on page %d", page)
				}
				cursor = resp.NextCursor
				last = resp.Items[1]
			}
			exhausted := dueCursor{ID: last.UserWordID}
			if last.NextReviewAt != nil {
				exhausted.NextReviewAt = *last.NextReviewAt
			}
			resp, err := repo.ListDueWords(context.Background(), ListDueWordsRequest{UserID: userID, Limit: 2, AfterCursor: encodeDueCursor(exhausted)})
			if err != nil {
				t.Fatal(err)
			}
			if len(resp.Items) != 0 || resp.TotalCount != 6 || resp.NextCursor != "" {
				t.Fatalf("exhausted cursor restarted: %+v", resp)
			}
		})
	}
}
