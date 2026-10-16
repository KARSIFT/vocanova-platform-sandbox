package knowledgecatalog

import (
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCatalogHasRealStableEntriesAndReviewedCore(t *testing.T) {
	data, err := loadCatalog()
	require.NoError(t, err)
	require.Len(t, data.Items, 10000)
	require.Equal(t, wordNetNotice, data.WordNetLicense)
	byWord := map[string]CatalogEntry{}
	suspicious := regexp.MustCompile(`(?i)ethnic slur|racial slur|offensive term|obscene|vulgar|derogatory|sexual intercourse|genital|penis|prostitute|archaic|obsolete|oriental|savage|retardation|intelligence quotient|\d`)
	for i, e := range data.Items {
		require.Equal(t, i+1, e.Rank)
		require.NotEmpty(t, e.Definition)
		require.NotEqual(t, uuid.Nil, e.WordID)
		require.NotEqual(t, uuid.Nil, e.MeaningID)
		require.NotContains(t, e.Word, " ")
		_, duplicate := byWord[e.Word]
		require.False(t, duplicate)
		byWord[e.Word] = e
		if e.Source != "Vocanova authored" {
			namespace := uuid.NewSHA1(uuid.NameSpaceURL, []byte("https://vocanova.site/catalog/wordnet-3.0"))
			require.Equal(t, uuid.NewSHA1(namespace, []byte("primary-meaning:en:"+e.Word)), e.MeaningID)
		}
		if e.Source == "WordNet 3.0" {
			require.Empty(t, e.Example, e.Word)
			require.NotContains(t, e.Definition, `; "`, e.Word)
			require.NotNil(t, e.TagCount, e.Word)
			require.Positive(t, *e.TagCount, e.Word)
			require.False(t, suspicious.MatchString(e.Definition), e.Word+": "+e.Definition)
		}
	}
	require.Equal(t, "pronoun", byWord["i"].PartOfSpeech)
	require.Contains(t, byWord["i"].Definition, "speaking")
	require.Equal(t, "pronoun", byWord["us"].PartOfSpeech)
	require.Equal(t, "determiner", byWord["a"].PartOfSpeech)
	require.Equal(t, "verb", byWord["can"].PartOfSpeech)
	require.Contains(t, byWord["can"].Definition, "able")
	require.Contains(t, byWord["planet"].Definition, "star")
	require.NotContains(t, byWord["planet"].Definition, "nine")
	require.Equal(t, "determiner", byWord["more"].PartOfSpeech)
	require.NotContains(t, byWord["queen"].Definition, "insect")
	for _, word := range []string{"damn", "goddamn", "goddamned", "idiot", "harry", "jimmy", "nigger", "fucking", "gage", "coon", "nuthouse", "spic", "nymphomaniac"} {
		_, present := byWord[word]
		require.False(t, present, word)
	}
}
func TestPostgreSQLCompleteCatalogSeedAndConsistentBandCounts(t *testing.T) {
	db := isolatedCatalogDB(t)
	original := preloadAuthoredCore(t, db)
	snapshot := func(table, key string) string {
		t.Helper()
		var result string
		err := db.QueryRow(`SELECT jsonb_agg(to_jsonb(t) ORDER BY t.id)::text FROM `+table+` t WHERE id IN (SELECT (x->>'id')::uuid FROM jsonb_array_elements($1::jsonb->'`+key+`') x)`, string(original)).Scan(&result)
		require.NoError(t, err)
		return result
	}
	wordsBefore, meaningsBefore, examplesBefore := snapshot("canonical_words", "canonical_words"), snapshot("word_meanings", "word_meanings"), snapshot("word_examples", "word_examples")
	tx, err := db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	require.NoError(t, Seed(t.Context(), tx))
	require.NoError(t, tx.Commit())
	catalogSnapshot := func() string {
		t.Helper()
		var state string
		err := db.QueryRow(`SELECT jsonb_agg(to_jsonb(wm) ORDER BY wm.id)::text FROM word_meanings wm JOIN knowledge_catalog_entries kc ON kc.meaning_id=wm.id`).Scan(&state)
		require.NoError(t, err)
		return state
	}
	catalogBefore := catalogSnapshot()
	// Reruns preserve public identity and do not duplicate examples or learner rows.
	tx, err = db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	require.NoError(t, Seed(t.Context(), tx))
	require.NoError(t, tx.Commit())
	require.Equal(t, catalogBefore, catalogSnapshot())
	require.Equal(t, wordsBefore, snapshot("canonical_words", "canonical_words"))
	require.Equal(t, meaningsBefore, snapshot("word_meanings", "word_meanings"))
	require.Equal(t, examplesBefore, snapshot("word_examples", "word_examples"))
	user := uuid.New()
	_, err = db.ExecContext(t.Context(), `INSERT INTO users(id,email,status,created_at,updated_at) VALUES($1,$2,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, user, user.String()+"@catalog-band.invalid")
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM word_knowledge_actions WHERE user_id=$1`, user)
		db.Exec(`DELETE FROM user_word_knowledge WHERE user_id=$1`, user)
		db.Exec(`DELETE FROM user_words WHERE user_id=$1`, user)
		db.Exec(`DELETE FROM users WHERE id=$1`, user)
	})
	svc := NewService(NewPostgreSQLRepository(db), nil)
	first, err := svc.Read(t.Context(), user, 1)
	require.NoError(t, err)
	require.Equal(t, 10000, first.TotalWords)
	require.Len(t, first.Bands, 100)
	require.Len(t, first.CurrentBand.Items, 100)
	require.Equal(t, 1, first.CurrentBand.Items[0].Rank)
	require.Equal(t, "I", first.CurrentBand.Items[1].WordText)
	require.Equal(t, wordNetNotice, first.Attribution.WordNetLicense)
	for _, b := range first.Bands {
		require.Equal(t, 100, b.TotalWords)
		require.Equal(t, 100, b.UnseenCount)
		require.Zero(t, b.KnownCount)
		require.Zero(t, b.LearningCount)
	}
	last, err := svc.Read(t.Context(), user, 100)
	require.NoError(t, err)
	require.Len(t, last.CurrentBand.Items, 100)
	require.Equal(t, 9901, last.CurrentBand.Items[0].Rank)
	require.Equal(t, 10000, last.CurrentBand.Items[99].Rank)
	target := first.CurrentBand.Items[0]
	_, err = svc.Decide(t.Context(), DecisionRequest{UserID: user, MeaningID: target.MeaningID, Decision: "known", ExpectedRevision: target.Revision, IdempotencyKey: "band-known"})
	require.NoError(t, err)
	after, err := svc.Read(t.Context(), user, 1)
	require.NoError(t, err)
	require.Equal(t, 1, after.Bands[0].KnownCount)
	require.Equal(t, 99, after.Bands[0].UnseenCount)
	target = after.CurrentBand.Items[0]
	_, err = svc.Decide(t.Context(), DecisionRequest{UserID: user, MeaningID: target.MeaningID, Decision: "learn", ExpectedRevision: target.Revision, IdempotencyKey: "band-learn"})
	require.NoError(t, err)
	after, err = svc.Read(t.Context(), user, 1)
	require.NoError(t, err)
	require.Zero(t, after.Bands[0].KnownCount)
	require.Equal(t, 1, after.Bands[0].LearningCount)
	require.Equal(t, 99, after.Bands[0].UnseenCount)
}
