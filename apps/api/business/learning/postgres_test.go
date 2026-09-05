package learning

import (
	"errors"
	"regexp"
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
	mock.ExpectQuery("SELECT wm.id FROM word_meanings wm JOIN canonical_words cw").
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
	mock.ExpectQuery("SELECT wm.id FROM word_meanings wm JOIN canonical_words cw").
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

func TestPostgreSQLRepositorySaveUserWordRejectsInactiveCanonicalWord(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT wm.id FROM word_meanings wm JOIN canonical_words cw").
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
	mock.ExpectQuery("SELECT wm.id FROM word_meanings wm JOIN canonical_words cw").
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
		WithArgs(userID, sqlmock.AnyArg(), sqlmock.AnyArg(), 2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "meaning_id", "word_id", "text", "normalized_text", "part_of_speech", "short_definition", "status", "source", "added_at"}).
			AddRow("00000000-0000-0000-0000-000000000003", "00000000-0000-0000-0000-000000000002", "00000000-0000-0000-0000-000000000004", "boarding pass", "boarding pass", "noun", "A document.", "new", "journey", now))

	resp, err := repo.ListSavedWords(t.Context(), ListSavedWordsRequest{UserID: userID, Limit: 1})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "boarding-pass", resp.Items[0].WordSlug)
	assert.Empty(t, resp.NextCursor)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryListSavedWordsUsesLookahead(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	now := time.Now()
	mock.ExpectQuery("ORDER BY uw.added_at DESC, uw.id DESC").
		WithArgs(userID, sqlmock.AnyArg(), sqlmock.AnyArg(), 2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "meaning_id", "word_id", "text", "normalized_text", "part_of_speech", "short_definition", "status", "source", "added_at"}).
			AddRow("00000000-0000-0000-0000-000000000003", "00000000-0000-0000-0000-000000000002", "00000000-0000-0000-0000-000000000004", "first", "first", "noun", "First.", "new", "journey", now).
			AddRow("00000000-0000-0000-0000-000000000005", "00000000-0000-0000-0000-000000000006", "00000000-0000-0000-0000-000000000007", "second", "second", "noun", "Second.", "new", "journey", now))

	resp, err := repo.ListSavedWords(t.Context(), ListSavedWordsRequest{UserID: userID, Limit: 1})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.NotEmpty(t, resp.NextCursor)
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

func TestPostgreSQLRepositorySavedWordStatesUsesDatabaseTime(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	userID := MustParseUUID("00000000-0000-0000-0000-000000000001")
	meaningID := MustParseUUID("00000000-0000-0000-0000-000000000002")
	userWordID := MustParseUUID("00000000-0000-0000-0000-000000000003")

	mock.ExpectQuery("SELECT meaning_id, id, CASE WHEN status").
		WithArgs(userID, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"meaning_id", "id", "status", "due"}).
			AddRow(meaningID, userWordID, "learning", true))

	states, err := repo.SavedWordStates(t.Context(), userID, []uuid.UUID{meaningID})
	require.NoError(t, err)
	require.Contains(t, states, meaningID)
	assert.Equal(t, userWordID, states[meaningID].UserWordID)
	assert.Equal(t, "learning", states[meaningID].Status)
	assert.True(t, states[meaningID].Due)
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

// TestPostgreSQLRepositorySaveUserWordWithGamificationNil tests that word-addition
// without gamification works (gamification is optional and nil by default).
// This verifies that the pre-existing P1 behavior is byte-for-byte unchanged
// when gamification is disabled (VOC-030-TEST-11).
func TestPostgreSQLRepositorySaveUserWordWithGamificationNil(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Repository WITHOUT gamification service (nil) - the default
	repo := NewPostgreSQLRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	newID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT wm.id FROM word_meanings wm JOIN canonical_words cw").
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
	assert.Equal(t, newID, m.UserWordID)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgreSQLRepositorySaveUserWordFailedInsertNoReward tests that a
// failed user_words insert does not create any gamification reward (VOC-030-TEST-10).
func TestPostgreSQLRepositorySaveUserWordFailedInsertNoReward(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT wm.id FROM word_meanings wm JOIN canonical_words cw").
		WithArgs(meaningID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(meaningID))
	mock.ExpectQuery("SELECT id, deleted_at FROM user_words").
		WithArgs(userID, meaningID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "deleted_at"}))
	// Simulate a failed INSERT
	mock.ExpectQuery("INSERT INTO user_words").
		WithArgs(sqlmock.AnyArg(), userID, meaningID, "journey", now).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	_, err = repo.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID:    userID,
		MeaningID: meaningID,
		Source:    "journey",
	}, now)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgreSQLRepositorySaveUserWordRestoreDeletedNoNewReward tests that
// restoring a previously deleted word-add (idempotent operation) does not
// create a new gamification reward since the word row already exists
// (VOC-030-TEST-09 scenario - idempotent behavior).
func TestPostgreSQLRepositorySaveUserWordRestoreDeletedNoNewReward(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	existingID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT wm.id FROM word_meanings wm JOIN canonical_words cw").
		WithArgs(meaningID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(meaningID))
	mock.ExpectQuery("SELECT id, deleted_at FROM user_words").
		WithArgs(userID, meaningID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "deleted_at"}).
			AddRow(existingID, time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC))) // Previously deleted
	// Restore the deleted row as a fresh saved word instead of inserting a
	// new one. Every scheduling/history field must match a newly saved word.
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE user_words
			 SET deleted_at = NULL, status = 'new', source = $1, review_step = 0,
			     next_review_at = NULL, last_reviewed_at = NULL,
			     last_result = NULL, last_rating = NULL,
			     consecutive_correct_count = 0, consecutive_incorrect_count = 0,
			     total_review_count = 0, correct_review_count = 0,
			     mastered_at = NULL, ignored_at = NULL,
			     added_at = $2, updated_at = $2
			 WHERE id = $3`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), existingID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(savedMeaningQuery()).
		WithArgs(userID, meaningID).
		WillReturnRows(savedMeaningRow(existingID, meaningID, wordID, now))

	m, err := repo.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID:    userID,
		MeaningID: meaningID,
		Source:    "journey",
	}, now)
	require.NoError(t, err)
	assert.Equal(t, existingID, m.UserWordID)
	// When restoring a previously deleted word, the transaction does an UPDATE,
	// not an INSERT. The code path does not call gamification.GrantPoint when
	// restoring deleted words (this is correct idempotent behavior).
	require.NoError(t, mock.ExpectationsWereMet())
}
