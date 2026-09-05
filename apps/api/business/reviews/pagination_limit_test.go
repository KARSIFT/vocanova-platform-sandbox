package reviews

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPaginationHonorsDocumentedMaximum(t *testing.T) {
	userID, wordID := uuid.New(), uuid.New()
	data := MemoryRepositoryData{Words: []MemoryWord{{ID: wordID, Text: "work", NormalizedText: "work", Status: "active"}}}
	for i := range 61 {
		meaningID := uuid.New()
		data.Meanings = append(data.Meanings, MemoryMeaning{ID: meaningID, WordID: wordID, Status: "active"})
		data.UserWords = append(data.UserWords, MemoryUserWord{ID: uuid.New(), UserID: userID, MeaningID: meaningID, Status: "new", AddedAt: time.Unix(int64(i), 0)})
	}
	repo := NewMemoryRepository(data)
	page, err := repo.ListDueWords(t.Context(), ListDueWordsRequest{UserID: userID, Limit: 100})
	require.NoError(t, err)
	require.Equal(t, 50, len(page.Items))
	require.NotEmpty(t, page.NextCursor)
	last, err := repo.ListDueWords(t.Context(), ListDueWordsRequest{UserID: userID, Limit: 100, AfterCursor: page.NextCursor})
	require.NoError(t, err)
	require.Equal(t, 11, len(last.Items))
	require.Empty(t, last.NextCursor)
	require.Equal(t, 61, page.TotalCount)
	require.Equal(t, 61, last.TotalCount)
	seen := map[uuid.UUID]bool{}
	for _, item := range append(page.Items, last.Items...) {
		require.False(t, seen[item.UserWordID], "cursor must not repeat a saved word")
		seen[item.UserWordID] = true
	}
	defaultPage, err := repo.ListDueWords(t.Context(), ListDueWordsRequest{UserID: userID})
	require.NoError(t, err)
	require.Equal(t, 20, len(defaultPage.Items))
}

func TestPostgreSQLPaginationCapsLookaheadAt51(t *testing.T) {
	for _, requested := range []int{51, 100, 1000} {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		repo := NewPostgreSQLRepository(db, nil)
		userID := uuid.New()
		mock.ExpectQuery("SELECT \\* FROM").WithArgs(userID, sqlmock.AnyArg(), sqlmock.AnyArg(), 51).WillReturnRows(sqlmock.NewRows([]string{"unused"}))
		_, err = repo.ListDueWords(t.Context(), ListDueWordsRequest{UserID: userID, Limit: requested})
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
		mock.ExpectClose()
		require.NoError(t, db.Close())
	}
}
