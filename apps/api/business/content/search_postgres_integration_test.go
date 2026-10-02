package content

import (
	"database/sql"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSearchMeaningsPostgreSQLFiltersCountsAndCursors(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	db.SetMaxOpenConns(1)
	_, err = db.ExecContext(t.Context(), `CREATE TEMP TABLE user_word_knowledge(user_id uuid,meaning_id uuid,self_reported_known boolean);
 CREATE TEMP TABLE user_words(user_id uuid,meaning_id uuid,deleted_at timestamptz);
 CREATE TEMP TABLE canonical_words(id uuid PRIMARY KEY,text text NOT NULL,normalized_text text NOT NULL,status text NOT NULL,difficulty_level text);
	CREATE TEMP TABLE word_meanings(id uuid PRIMARY KEY,word_id uuid REFERENCES canonical_words(id),part_of_speech text,short_definition text,learner_definition text,status text,difficulty_level text);
	CREATE TEMP TABLE journey_situations(id uuid PRIMARY KEY,status text,category text);
	CREATE TEMP TABLE journey_words(journey_situation_id uuid REFERENCES journey_situations(id),meaning_id uuid REFERENCES word_meanings(id));`)
	require.NoError(t, err)
	data := searchFixtureData()
	for _, word := range data.Words {
		_, err = db.ExecContext(t.Context(), `INSERT INTO canonical_words VALUES($1,$2,$3,$4,NULLIF($5,''))`, word.ID, word.Text, word.NormalizedText, word.Status, word.DifficultyLevel)
		require.NoError(t, err)
	}
	for _, meaning := range data.Meanings {
		_, err = db.ExecContext(t.Context(), `INSERT INTO word_meanings VALUES($1,$2,$3,$4,$5,$6,NULLIF($7,''))`, meaning.ID, meaning.WordID, meaning.PartOfSpeech, meaning.ShortDefinition, meaning.LearnerDefinition, meaning.Status, meaning.DifficultyLevel)
		require.NoError(t, err)
	}
	for _, situation := range data.Situations {
		_, err = db.ExecContext(t.Context(), `INSERT INTO journey_situations VALUES($1,$2,$3)`, situation.ID, situation.Status, situation.Category)
		require.NoError(t, err)
	}
	for _, link := range data.JourneyWords {
		_, err = db.ExecContext(t.Context(), `INSERT INTO journey_words VALUES($1,$2)`, link.JourneySituationID, link.MeaningID)
		require.NoError(t, err)
	}
	repo := NewPostgreSQLRepository(db)
	memory := NewMemoryRepository(data)
	for _, req := range []SearchRequest{{}, {Query: "DOCUMENT"}, {Query: "ticket-like"}, {Query: "%"}, {Query: "_"}, {Category: "travel"}, {Category: "social"}, {Level: "a2"}, {Level: "b1"}, {Level: "unknown"}, {Query: "missing"}} {
		expected, err := memory.SearchMeanings(t.Context(), req)
		require.NoError(t, err)
		actual, err := repo.SearchMeanings(t.Context(), req)
		require.NoError(t, err)
		require.Equal(t, expected, actual)
	}
	owner, other := uuid.New(), uuid.New()
	_, err = db.ExecContext(t.Context(), `INSERT INTO user_word_knowledge VALUES($1,$3,true),($1,$4,false),($2,$4,true)`, owner, other, data.Meanings[0].ID, data.Meanings[1].ID)
	require.NoError(t, err)
	_, err = db.ExecContext(t.Context(), `INSERT INTO user_words VALUES($1,$2,NULL),($1,$3,NULL)`, owner, data.Meanings[0].ID, data.Meanings[2].ID)
	require.NoError(t, err)
	for _, tc := range []struct {
		user   uuid.UUID
		filter string
		count  int
	}{{owner, "known", 1}, {owner, "saved", 2}, {owner, "unexplored", 1}, {other, "known", 1}, {other, "saved", 0}, {other, "unexplored", 2}} {
		page, e := repo.SearchMeanings(t.Context(), SearchRequest{UserID: tc.user, Knowledge: tc.filter, Limit: 1})
		require.NoError(t, e)
		require.Equal(t, tc.count, page.TotalCount)
		if tc.count > 1 {
			require.NotEmpty(t, page.NextCursor)
			_, e = repo.SearchMeanings(t.Context(), SearchRequest{UserID: tc.user, Knowledge: "known", AfterCursor: page.NextCursor})
			require.ErrorIs(t, e, ErrInvalidCursor)
			_, e = repo.SearchMeanings(t.Context(), SearchRequest{UserID: uuid.New(), Knowledge: tc.filter, AfterCursor: page.NextCursor})
			require.ErrorIs(t, e, ErrInvalidCursor)
			next, e := repo.SearchMeanings(t.Context(), SearchRequest{UserID: tc.user, Knowledge: tc.filter, AfterCursor: page.NextCursor})
			require.NoError(t, e)
			require.Equal(t, tc.count, next.TotalCount)
			require.Len(t, next.Items, tc.count-1)
		}
	}
	first, err := repo.SearchMeanings(t.Context(), SearchRequest{Limit: 1})
	require.NoError(t, err)
	require.Equal(t, 3, first.TotalCount)
	require.Len(t, first.Items, 1)
	// Deleting the cursor boundary must not restart or duplicate the result set.
	_, err = db.ExecContext(t.Context(), `DELETE FROM journey_words WHERE meaning_id=$1`, first.Items[0].MeaningID)
	require.NoError(t, err)
	_, err = db.ExecContext(t.Context(), `DELETE FROM word_meanings WHERE id=$1`, first.Items[0].MeaningID)
	require.NoError(t, err)
	second, err := repo.SearchMeanings(t.Context(), SearchRequest{Limit: 1, AfterCursor: first.NextCursor})
	require.NoError(t, err)
	require.Equal(t, 2, second.TotalCount)
	require.Equal(t, data.Meanings[1].ID, second.Items[0].MeaningID)
	last, err := repo.SearchMeanings(t.Context(), SearchRequest{AfterCursor: second.NextCursor, Limit: 1})
	require.NoError(t, err)
	require.Empty(t, last.NextCursor)
	require.Equal(t, data.Meanings[2].ID, last.Items[0].MeaningID)
	endCursor := searchNextCursor(SearchRequest{}, last.Items[0])
	end, err := repo.SearchMeanings(t.Context(), SearchRequest{AfterCursor: endCursor})
	require.NoError(t, err)
	require.Equal(t, 2, end.TotalCount)
	require.Empty(t, end.Items)
	_, err = repo.SearchMeanings(t.Context(), SearchRequest{Category: "travel", AfterCursor: endCursor})
	require.ErrorIs(t, err, ErrInvalidCursor)
}
