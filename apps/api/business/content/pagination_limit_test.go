package content

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSituationPaginationHonorsDocumentedMaximum(t *testing.T) {
	var situations []Situation
	for i := range 61 {
		situations = append(situations, Situation{ID: uuid.New(), Status: "active", DisplayOrder: i})
	}
	repo := NewMemoryRepository(MemoryRepositoryData{Situations: situations})
	page, err := repo.ListSituations(t.Context(), ListSituationsRequest{Limit: 100})
	require.NoError(t, err)
	require.Equal(t, 50, len(page.Items), "DOC-07 caps every cursor page at 50 items")
	require.NotEmpty(t, page.NextCursor)
	last, err := repo.ListSituations(t.Context(), ListSituationsRequest{Limit: 100, AfterCursor: page.NextCursor})
	require.NoError(t, err)
	require.Len(t, last.Items, 11)
	require.Empty(t, last.NextCursor)
	seen := map[uuid.UUID]bool{}
	for _, item := range append(page.Items, last.Items...) {
		require.False(t, seen[item.ID])
		seen[item.ID] = true
	}
	defaultPage, err := repo.ListSituations(t.Context(), ListSituationsRequest{})
	require.NoError(t, err)
	require.Len(t, defaultPage.Items, 20)
}

func TestPostgreSQLPaginationCapsLookaheadAt51(t *testing.T) {
	for _, requested := range []int{51, 100, 1000} {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		mock.ExpectQuery("SELECT id, slug").WithArgs(nil, sqlmock.AnyArg(), 51).WillReturnRows(sqlmock.NewRows([]string{"unused"}))
		_, err = NewPostgreSQLRepository(db).ListSituations(t.Context(), ListSituationsRequest{Limit: requested})
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
		mock.ExpectClose()
		require.NoError(t, db.Close())
	}
}
