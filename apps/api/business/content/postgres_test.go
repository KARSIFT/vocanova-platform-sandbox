package content

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgreSQLRepositoryListSituations(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)

	mock.ExpectQuery("SELECT id, slug, title, short_description, level_band, category, status, display_order, created_at, updated_at").
		WithArgs(0, sqlmock.AnyArg(), 20).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "title", "short_description", "level_band", "category", "status", "display_order", "created_at", "updated_at"}).
			AddRow("00000000-0000-0000-0000-000000000001", "airport", "Airport", "Airport words.", "a1_a2", "travel", "active", 1, time.Now(), time.Now()))

	resp, err := repo.ListSituations(t.Context(), ListSituationsRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "airport", resp.Items[0].Slug)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryGetSituationBySlug(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)

	mock.ExpectQuery("SELECT id, slug, title, short_description, level_band, category, status, display_order, created_at, updated_at").
		WithArgs("airport").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "title", "short_description", "level_band", "category", "status", "display_order", "created_at", "updated_at"}).
			AddRow("00000000-0000-0000-0000-000000000001", "airport", "Airport", "Airport words.", "a1_a2", "travel", "active", 1, time.Now(), time.Now()))

	sit, err := repo.GetSituationBySlug(t.Context(), "airport")
	require.NoError(t, err)
	assert.Equal(t, "airport", sit.Slug)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryGetSituationBySlugNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)

	mock.ExpectQuery("SELECT id, slug, title").
		WithArgs("unknown").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "title", "short_description", "level_band", "category", "status", "display_order", "created_at", "updated_at"}))

	_, err = repo.GetSituationBySlug(t.Context(), "unknown")
	assert.ErrorIs(t, err, ErrSituationNotFound)
}

func TestPostgreSQLRepositoryGetMeaningsBySituation(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)

	mock.ExpectQuery("SELECT m.id, m.word_id, cw.text, cw.normalized_text").
		WithArgs(MustParseUUID("00000000-0000-0000-0000-000000000001")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word_id", "text", "normalized_text", "part_of_speech", "short_definition", "display_order"}).
			AddRow("00000000-0000-0000-0000-000000000002", "00000000-0000-0000-0000-000000000003", "boarding pass", "boarding pass", "noun", "A document that lets you get on your flight.", 1))

	meanings, err := repo.GetMeaningsBySituation(t.Context(), MustParseUUID("00000000-0000-0000-0000-000000000001"))
	require.NoError(t, err)
	require.Len(t, meanings, 1)
	assert.Equal(t, "boarding-pass", meanings[0].WordSlug)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryGetWordBySlug(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	wordID := MustParseUUID("00000000-0000-0000-0000-000000000002")
	meaningID := MustParseUUID("00000000-0000-0000-0000-000000000003")

	mock.ExpectQuery("SELECT id, text, normalized_text, word_type, difficulty_level").
		WithArgs("boarding-pass").
		WillReturnRows(sqlmock.NewRows([]string{"id", "text", "normalized_text", "word_type", "difficulty_level"}).
			AddRow(wordID.String(), "boarding pass", "boarding pass", "phrase", nil))

	mock.ExpectQuery("SELECT id, part_of_speech, short_definition, learner_definition, meaning_order").
		WithArgs(wordID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "part_of_speech", "short_definition", "learner_definition", "meaning_order"}).
			AddRow(meaningID.String(), "noun", "A document that lets you get on your flight.", nil, 1))

	mock.ExpectQuery("SELECT id, meaning_id, example_text, situation_label").
		WillReturnRows(sqlmock.NewRows([]string{"id", "meaning_id", "example_text", "situation_label"}).
			AddRow("00000000-0000-0000-0000-000000000004", meaningID.String(), "Please have your boarding pass ready.", "Airport"))

	mock.ExpectQuery("SELECT id, meaning_id, note_type, note_text").
		WillReturnRows(sqlmock.NewRows([]string{"id", "meaning_id", "note_type", "note_text"}).
			AddRow("00000000-0000-0000-0000-000000000005", meaningID.String(), "collocation", "show a boarding pass"))

	word, err := repo.GetWordBySlug(t.Context(), "boarding-pass")
	require.NoError(t, err)
	assert.Equal(t, "boarding-pass", word.Slug)
	require.Len(t, word.Meanings, 1)
	require.Len(t, word.Meanings[0].Examples, 1)
	require.Len(t, word.Meanings[0].UsageNotes, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryGetWordBySlugNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)

	mock.ExpectQuery("SELECT id, text, normalized_text, word_type, difficulty_level").
		WithArgs("unknown").
		WillReturnRows(sqlmock.NewRows([]string{"id", "text", "normalized_text", "word_type", "difficulty_level"}))

	_, err = repo.GetWordBySlug(t.Context(), "unknown")
	assert.ErrorIs(t, err, ErrWordNotFound)
}

func TestWordSlug(t *testing.T) {
	assert.Equal(t, "boarding-pass", wordSlug("boarding pass"))
	assert.Equal(t, "check-out", wordSlug("check-out"))
}

func TestSituationCursorRoundTrip(t *testing.T) {
	c := situationCursor{DisplayOrder: 3, ID: uuid.MustParse("00000000-0000-0000-0000-000000000001")}
	s := encodeSituationCursor(c)
	decoded, err := decodeSituationCursor(s)
	require.NoError(t, err)
	assert.Equal(t, c, decoded)
}

func TestSituationCursorInvalid(t *testing.T) {
	_, err := decodeSituationCursor("not-valid")
	require.Error(t, err)
}
