package learning

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func savedMeaningQuery() string {
	return "SELECT uw.id, uw.meaning_id, cw.id, cw.text, cw.normalized_text, wm.part_of_speech, wm.short_definition, uw.status, uw.source, uw.added_at"
}

func savedMeaningColumns() []string {
	return []string{"id", "meaning_id", "word_id", "text", "normalized_text", "part_of_speech", "short_definition", "status", "source", "added_at"}
}

func savedMeaningRow(id, meaningID, wordID uuid.UUID, now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows(savedMeaningColumns()).
		AddRow(id.String(), meaningID.String(), wordID.String(), "boarding pass", "boarding pass", "noun", "A document.", "new", "journey", now)
}

func TestPostgreSQLRepositorySaveUserWord(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	newID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM word_meanings").
		WithArgs(meaningID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(meaningID))
	mock.ExpectQuery("SELECT id, deleted_at FROM user_words").
		WithArgs(userID, meaningID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "deleted_at"}))
	mock.ExpectQuery("INSERT INTO user_words").
		WithArgs(sqlmock.AnyArg(), userID, meaningID, "journey", now).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(newID))
	mock.ExpectCommit()
	mock.ExpectQuery(savedMeaningQuery()).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(savedMeaningRow(newID, meaningID, wordID, now))

	m, err := repo.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID:    userID,
		MeaningID: meaningID,
		Source:    "journey",
	}, now)
	require.NoError(t, err)
	assert.Equal(t, "new", m.Status)
	assert.Equal(t, "journey", m.Source)
	assert.Equal(t, "boarding-pass", m.WordSlug)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositorySaveUserWordMeaningNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM word_meanings").
		WithArgs(meaningID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectRollback()

	_, err = repo.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID:    userID,
		MeaningID: meaningID,
		Source:    "journey",
	}, now)
	assert.ErrorIs(t, err, ErrMeaningNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositorySaveUserWordAlreadySaved(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	now := time.Now()
	existingID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM word_meanings").
		WithArgs(meaningID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(meaningID))
	mock.ExpectQuery("SELECT id, deleted_at FROM user_words").
		WithArgs(userID, meaningID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "deleted_at"}).
			AddRow(existingID, nil))
	mock.ExpectCommit()
	mock.ExpectQuery(savedMeaningQuery()).
		WithArgs(userID, meaningID).
		WillReturnRows(savedMeaningRow(existingID, meaningID, wordID, now))

	m, err := repo.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID:    userID,
		MeaningID: meaningID,
		Source:    "manual",
	}, now)
	require.NoError(t, err)
	assert.Equal(t, "journey", m.Source)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryUnsaveUserWord(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	now := time.Now()

	mock.ExpectExec("UPDATE user_words").
		WithArgs(sqlmock.AnyArg(), userID, meaningID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.UnsaveUserWord(t.Context(), userID, meaningID, now))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryUnsaveUserWordNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	now := time.Now()

	mock.ExpectExec("UPDATE user_words").
		WithArgs(sqlmock.AnyArg(), userID, meaningID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	assert.ErrorIs(t, repo.UnsaveUserWord(t.Context(), userID, meaningID, now), ErrUserWordNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryListSavedWords(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	now := time.Now()

	mock.ExpectQuery("SELECT uw.id, uw.meaning_id, cw.id, cw.text, cw.normalized_text").
		WithArgs(userID, sqlmock.AnyArg(), sqlmock.AnyArg(), 20).
		WillReturnRows(sqlmock.NewRows([]string{"id", "meaning_id", "word_id", "text", "normalized_text", "part_of_speech", "short_definition", "status", "source", "added_at"}).
			AddRow("00000000-0000-0000-0000-000000000003", "00000000-0000-0000-0000-000000000002", "00000000-0000-0000-0000-000000000004", "boarding pass", "boarding pass", "noun", "A document.", "new", "journey", now))

	resp, err := repo.ListSavedWords(t.Context(), ListSavedWordsRequest{UserID: userID})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "boarding-pass", resp.Items[0].WordSlug)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryIsSaved(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	mock.ExpectQuery("SELECT meaning_id FROM user_words").
		WithArgs(userID, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"meaning_id"}).AddRow(meaningID))

	states, err := repo.IsSaved(t.Context(), userID, []uuid.UUID{meaningID})
	require.NoError(t, err)
	assert.True(t, states[meaningID])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryGetSavedMeaning(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	now := time.Now()
	rowID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	mock.ExpectQuery(savedMeaningQuery()).
		WithArgs(userID, meaningID).
		WillReturnRows(savedMeaningRow(rowID, meaningID, wordID, now))

	m, err := repo.GetSavedMeaning(t.Context(), userID, meaningID)
	require.NoError(t, err)
	assert.Equal(t, rowID, m.UserWordID)
	require.NoError(t, mock.ExpectationsWereMet())
}
