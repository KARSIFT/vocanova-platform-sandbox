package content

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func searchFixtureData() MemoryRepositoryData {
	w1, w2, w3 := uuid.MustParse("10000000-0000-0000-0000-000000000001"), uuid.MustParse("10000000-0000-0000-0000-000000000002"), uuid.MustParse("10000000-0000-0000-0000-000000000003")
	m1, m2, m3, m4, m5 := uuid.MustParse("20000000-0000-0000-0000-000000000001"), uuid.MustParse("20000000-0000-0000-0000-000000000002"), uuid.MustParse("20000000-0000-0000-0000-000000000003"), uuid.MustParse("20000000-0000-0000-0000-000000000004"), uuid.MustParse("20000000-0000-0000-0000-000000000005")
	s1, s2, s3 := uuid.New(), uuid.New(), uuid.New()
	return MemoryRepositoryData{
		Words:        []SeedWord{{ID: w1, Text: "Boarding pass", NormalizedText: "boarding pass", Status: "active", DifficultyLevel: "a2"}, {ID: w2, Text: "Gate", NormalizedText: "gate", Status: "active"}, {ID: w3, Text: "Archived", NormalizedText: "archived", Status: "archived"}},
		Meanings:     []SeedMeaning{{ID: m1, WordID: w1, PartOfSpeech: "noun", ShortDefinition: "A travel document", LearnerDefinition: "A ticket-like document you show before boarding", Status: "active"}, {ID: m2, WordID: w1, PartOfSpeech: "noun", ShortDefinition: "A second selected sense", Status: "active", DifficultyLevel: "b1"}, {ID: m3, WordID: w2, PartOfSpeech: "noun", ShortDefinition: "A 100% synthetic definition", Status: "active"}, {ID: m4, WordID: w1, ShortDefinition: "Hidden draft", Status: "draft"}, {ID: m5, WordID: w3, ShortDefinition: "Hidden archived word", Status: "active"}},
		Situations:   []Situation{{ID: s1, Category: "travel", Status: "active"}, {ID: s2, Category: "travel", Status: "active"}, {ID: s3, Category: "social", Status: "archived"}},
		JourneyWords: []SeedJourneyWord{{JourneySituationID: s1, MeaningID: m1}, {JourneySituationID: s2, MeaningID: m1}, {JourneySituationID: s1, MeaningID: m2}, {JourneySituationID: s3, MeaningID: m3}},
	}
}

func TestSearchMeaningFiltersAndPagination(t *testing.T) {
	repo := NewMemoryRepository(searchFixtureData())
	for _, tc := range []struct {
		name  string
		req   SearchRequest
		count int
		level string
	}{
		{"browse active meanings", SearchRequest{}, 3, ""},
		{"trim and case", SearchRequest{Query: "  DOCUMENT  "}, 1, "a2"},
		{"full definition", SearchRequest{Query: "ticket-like"}, 1, "a2"},
		{"word", SearchRequest{Query: "BOARDING"}, 2, ""},
		{"literal percent", SearchRequest{Query: "%"}, 1, "unknown"},
		{"literal underscore", SearchRequest{Query: "_"}, 0, ""},
		{"category duplicate links", SearchRequest{Category: "travel"}, 2, ""},
		{"archived situation excluded", SearchRequest{Category: "social"}, 0, ""},
		{"meaning level overrides word", SearchRequest{Level: "b1"}, 1, "b1"},
		{"word level fallback", SearchRequest{Level: "a2"}, 1, "a2"},
		{"unknown fallback", SearchRequest{Level: "unknown"}, 1, "unknown"},
		{"combined filters", SearchRequest{Query: "document", Category: "travel", Level: "b1"}, 0, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := repo.SearchMeanings(t.Context(), tc.req)
			require.NoError(t, err)
			assert.Len(t, resp.Items, tc.count)
			assert.Equal(t, tc.count, resp.TotalCount)
			if tc.level != "" {
				assert.Equal(t, tc.level, resp.Items[0].DifficultyLevel)
			}
		})
	}
	page, err := repo.SearchMeanings(t.Context(), SearchRequest{Limit: 1})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.NotEmpty(t, page.NextCursor)
	assert.Equal(t, 3, page.TotalCount)
	seen := []uuid.UUID{page.Items[0].MeaningID}
	for page.NextCursor != "" {
		page, err = repo.SearchMeanings(t.Context(), SearchRequest{Limit: 1, AfterCursor: page.NextCursor})
		require.NoError(t, err)
		assert.Equal(t, 3, page.TotalCount)
		require.Len(t, page.Items, 1)
		seen = append(seen, page.Items[0].MeaningID)
	}
	assert.Equal(t, []uuid.UUID{repo.meanings[0].ID, repo.meanings[1].ID, repo.meanings[2].ID}, seen)
	// Exhausted cursors still report the full filtered total.
	exhausted := searchNextCursor(SearchRequest{}, page.Items[0])
	page, err = repo.SearchMeanings(t.Context(), SearchRequest{AfterCursor: exhausted})
	require.NoError(t, err)
	assert.Empty(t, page.Items)
	assert.Equal(t, 3, page.TotalCount)
}

func TestSearchRejectsInvalidFiltersAndChangedCursorScope(t *testing.T) {
	repo := NewMemoryRepository(searchFixtureData())
	for _, req := range []SearchRequest{{Query: strings.Repeat("é", 101)}, {Category: "invalid"}, {Level: "A2"}, {Level: "c2"}} {
		_, err := repo.SearchMeanings(t.Context(), req)
		assert.ErrorIs(t, err, ErrInvalidSearch)
	}
	_, err := repo.SearchMeanings(t.Context(), SearchRequest{Query: strings.Repeat("é", 100)})
	require.NoError(t, err)
	page, err := repo.SearchMeanings(t.Context(), SearchRequest{Query: "BOARDING", Limit: 1})
	require.NoError(t, err)
	require.NotEmpty(t, page.NextCursor)
	_, err = repo.SearchMeanings(t.Context(), SearchRequest{Query: " boarding ", Limit: 1, AfterCursor: page.NextCursor})
	require.NoError(t, err)
	for _, req := range []SearchRequest{{Query: "gate", AfterCursor: page.NextCursor}, {Query: "boarding", Category: "travel", AfterCursor: page.NextCursor}, {Query: "boarding", Level: "a2", AfterCursor: page.NextCursor}, {AfterCursor: "bad"}} {
		_, err := repo.SearchMeanings(t.Context(), req)
		assert.ErrorIs(t, err, ErrInvalidCursor)
	}
}

func TestSearchSavedStateIsRequesterScopedAndNotMutated(t *testing.T) {
	data := searchFixtureData()
	owner, other, savedID := uuid.New(), uuid.New(), uuid.New()
	reader := NewMemorySavedStateReaderWithStates(map[uuid.UUID]map[uuid.UUID]SavedWordState{owner: {data.Meanings[0].ID: {UserWordID: savedID, Status: "learning", Due: true}}})
	svc := NewService(NewMemoryRepository(data), reader)
	owned, err := svc.Search(t.Context(), owner, SearchRequest{Query: "document"})
	require.NoError(t, err)
	require.Len(t, owned.Items, 1)
	assert.True(t, owned.Items[0].Saved)
	assert.Equal(t, savedID, owned.Items[0].UserWordID)
	assert.True(t, owned.Items[0].Due)
	foreign, err := svc.Search(t.Context(), other, SearchRequest{Query: "document"})
	require.NoError(t, err)
	assert.False(t, foreign.Items[0].Saved)
	assert.Equal(t, uuid.Nil, foreign.Items[0].UserWordID)
	assert.Empty(t, foreign.Items[0].ReviewState)
	assert.False(t, foreign.Items[0].Due)
}

func TestSearchCapsPagesWithoutLosingTotal(t *testing.T) {
	data := MemoryRepositoryData{}
	for i := 0; i < 55; i++ {
		wordID := uuid.New()
		data.Words = append(data.Words, SeedWord{ID: wordID, Text: fmt.Sprintf("word %02d", i), NormalizedText: fmt.Sprintf("word %02d", i), Status: "active"})
		data.Meanings = append(data.Meanings, SeedMeaning{ID: uuid.New(), WordID: wordID, ShortDefinition: "A synthetic meaning", Status: "active"})
	}
	repo := NewMemoryRepository(data)
	first, err := repo.SearchMeanings(t.Context(), SearchRequest{Limit: 500})
	require.NoError(t, err)
	require.Len(t, first.Items, 50)
	assert.Equal(t, 55, first.TotalCount)
	require.NotEmpty(t, first.NextCursor)
	second, err := repo.SearchMeanings(t.Context(), SearchRequest{Limit: 500, AfterCursor: first.NextCursor})
	require.NoError(t, err)
	require.Len(t, second.Items, 5)
	assert.Equal(t, 55, second.TotalCount)
	assert.Empty(t, second.NextCursor)
	defaultPage, err := repo.SearchMeanings(t.Context(), SearchRequest{})
	require.NoError(t, err)
	assert.Len(t, defaultPage.Items, 20)
}
