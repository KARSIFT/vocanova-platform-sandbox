package content

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func rankedSearchFixture() MemoryRepositoryData {
	data := MemoryRepositoryData{}
	for i, text := range []string{"accelerator", "make", "maker", "zebra"} {
		id := uuid.New()
		data.Words = append(data.Words, SeedWord{ID: id, Text: text, NormalizedText: text, Status: "active", DifficultyLevel: "a2"})
		data.Meanings = append(data.Meanings, SeedMeaning{ID: uuid.New(), WordID: id, PartOfSpeech: "verb", ShortDefinition: "make something happen", Status: "active"})
		if i == 1 {
			data.Meanings = append(data.Meanings, SeedMeaning{ID: uuid.New(), WordID: id, PartOfSpeech: "noun", ShortDefinition: "a recognizable kind", Status: "active", DifficultyLevel: "b1"})
		}
	}
	return data
}

func TestSearchExactHeadwordBeforeDefinitionMatches(t *testing.T) {
	data := rankedSearchFixture()
	repo := NewMemoryRepository(data)
	for _, limit := range []int{1, 2, 20} {
		req := SearchRequest{Query: " MAKE ", Limit: limit}
		var words []string
		seen := map[uuid.UUID]bool{}
		for {
			page, err := repo.SearchMeanings(t.Context(), req)
			require.NoError(t, err)
			require.Equal(t, 5, page.TotalCount)
			for _, item := range page.Items {
				require.False(t, seen[item.MeaningID])
				seen[item.MeaningID] = true
				words = append(words, item.WordText)
			}
			if page.NextCursor == "" {
				break
			}
			req.AfterCursor = page.NextCursor
		}
		require.Equal(t, []string{"make", "make", "accelerator", "maker", "zebra"}, words)
	}
	page, err := repo.SearchMeanings(t.Context(), SearchRequest{Query: "make", Level: "b1"})
	require.NoError(t, err)
	require.Equal(t, 1, page.TotalCount)
	require.Equal(t, "noun", page.Items[0].PartOfSpeech)
}

func TestSearchLegacyCursorKeepsAlphabeticalContinuation(t *testing.T) {
	data := rankedSearchFixture()
	repo := NewMemoryRepository(data)
	req := SearchRequest{Query: "make", Limit: 1}
	// A pre-deployment first page ended at accelerator.
	raw, err := json.Marshal(searchCursor{Version: 1, Filter: searchFilter(req), Text: "accelerator", ID: data.Meanings[0].ID})
	require.NoError(t, err)
	req.AfterCursor = base64.RawURLEncoding.EncodeToString(raw)
	var words []string
	for {
		page, err := repo.SearchMeanings(t.Context(), req)
		require.NoError(t, err)
		for _, item := range page.Items {
			words = append(words, item.WordText)
		}
		if page.NextCursor == "" {
			break
		}
		req.AfterCursor = page.NextCursor
	}
	require.Equal(t, []string{"make", "make", "maker", "zebra"}, words)
}

func TestSearchRejectsOutOfRangeCursorRank(t *testing.T) {
	req := SearchRequest{Query: "make"}
	for _, rank := range []int{-1, 2} {
		raw, err := json.Marshal(searchCursor{Version: 2, Filter: searchFilter(req), Text: "make", ID: uuid.New(), Rank: rank})
		require.NoError(t, err)
		req.AfterCursor = base64.RawURLEncoding.EncodeToString(raw)
		_, _, err = prepareSearch(req)
		require.ErrorIs(t, err, ErrInvalidCursor)
	}
}

func TestSearchPostgreSQLExactRankAndLegacyPagination(t *testing.T) {
	data := rankedSearchFixture()
	db := searchPostgresFixture(t, data)
	repo, memory := NewPostgreSQLRepository(db), NewMemoryRepository(data)
	for _, req := range []SearchRequest{{Query: " MAKE ", Limit: 1}, {Query: "make", Limit: 2}, {Query: "make", Level: "b1", Limit: 1}, {Query: "make", Category: "travel"}, {Query: "missing"}, {}} {
		for {
			want, err := memory.SearchMeanings(t.Context(), req)
			require.NoError(t, err)
			got, err := repo.SearchMeanings(t.Context(), req)
			require.NoError(t, err)
			require.Equal(t, want, got)
			if got.NextCursor == "" {
				break
			}
			req.AfterCursor = got.NextCursor
		}
	}
	req := SearchRequest{Query: "make", Limit: 1}
	raw, err := json.Marshal(searchCursor{Version: 1, Filter: searchFilter(req), Text: "accelerator", ID: data.Meanings[0].ID})
	require.NoError(t, err)
	req.AfterCursor = base64.RawURLEncoding.EncodeToString(raw)
	for {
		want, err := memory.SearchMeanings(t.Context(), req)
		require.NoError(t, err)
		got, err := repo.SearchMeanings(t.Context(), req)
		require.NoError(t, err)
		require.Equal(t, want, got)
		if got.NextCursor == "" {
			break
		}
		req.AfterCursor = got.NextCursor
	}
	// Cursor rank is self-contained: deleting the boundary must still allow the
	// transition from exact matches to alphabetically earlier definition matches.
	page, err := repo.SearchMeanings(t.Context(), SearchRequest{Query: "make", Limit: 2})
	require.NoError(t, err)
	_, err = db.Exec(`DELETE FROM word_meanings WHERE id=$1`, page.Items[1].MeaningID)
	require.NoError(t, err)
	next, err := repo.SearchMeanings(t.Context(), SearchRequest{Query: "make", Limit: 1, AfterCursor: page.NextCursor})
	require.NoError(t, err)
	require.Equal(t, "accelerator", next.Items[0].WordText)
	require.Equal(t, 4, next.TotalCount)
}
