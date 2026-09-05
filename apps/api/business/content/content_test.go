package content

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleMemoryRepo() (*MemoryRepository, *MemorySavedStateReader) {
	sitID := MustParseUUID("00000000-0000-0000-0000-000000000001")
	wordID := MustParseUUID("00000000-0000-0000-0000-000000000002")
	meaningID := MustParseUUID("00000000-0000-0000-0000-000000000003")

	repo := NewMemoryRepository(MemoryRepositoryData{
		Situations: []Situation{
			{
				ID:               sitID,
				Slug:             "airport",
				Title:            "Airport",
				ShortDescription: "Airport words.",
				LevelBand:        "a1_a2",
				Category:         "travel",
				Status:           "active",
				DisplayOrder:     1,
			},
			{
				ID:               MustParseUUID("00000000-0000-0000-0000-000000000004"),
				Slug:             "restaurant",
				Title:            "Restaurant",
				ShortDescription: "Restaurant words.",
				Category:         "daily_life",
				Status:           "active",
				DisplayOrder:     2,
			},
			{
				ID:               MustParseUUID("00000000-0000-0000-0000-000000000005"),
				Slug:             "draft",
				Title:            "Draft",
				ShortDescription: "Draft.",
				Category:         "daily_life",
				Status:           "draft",
				DisplayOrder:     3,
			},
		},
		Words: []SeedWord{
			{
				ID:             wordID,
				Text:           "boarding pass",
				NormalizedText: "boarding pass",
				WordType:       "phrase",
				LanguageCode:   "en",
				Status:         "active",
			},
		},
		Meanings: []SeedMeaning{
			{
				ID:              meaningID,
				WordID:          wordID,
				PartOfSpeech:    "noun",
				ShortDefinition: "A document that lets you get on your flight.",
				MeaningOrder:    1,
				Status:          "active",
			},
		},
		Examples: []SeedExample{
			{
				ID:           MustParseUUID("00000000-0000-0000-0000-000000000006"),
				MeaningID:    meaningID,
				ExampleText:  "Please have your boarding pass ready.",
				ExampleOrder: 1,
				Status:       "active",
			},
		},
		Notes: []SeedNote{
			{
				ID:        MustParseUUID("00000000-0000-0000-0000-000000000007"),
				MeaningID: meaningID,
				NoteType:  "collocation",
				NoteText:  "show a boarding pass",
				NoteOrder: 1,
				Status:    "active",
			},
		},
		JourneyWords: []SeedJourneyWord{
			{
				ID:                 MustParseUUID("00000000-0000-0000-0000-000000000008"),
				JourneySituationID: sitID,
				MeaningID:          meaningID,
				RelevanceScore:     80,
				DisplayOrder:       intPtr(1),
				IsCore:             true,
			},
		},
	})

	userID := MustParseUUID("00000000-0000-0000-0000-000000000009")
	reader := NewMemorySavedStateReader(map[uuid.UUID]map[uuid.UUID]bool{
		userID: {meaningID: true},
	})
	return repo, reader
}

func TestServiceListSituationsOnlyActive(t *testing.T) {
	repo, _ := sampleMemoryRepo()
	svc := NewService(repo, nil)

	resp, err := svc.ListSituations(context.Background(), ListSituationsRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Items, 2)
	assert.Equal(t, "airport", resp.Items[0].Slug)
	assert.Equal(t, "restaurant", resp.Items[1].Slug)
}

func TestServiceListSituationsPagination(t *testing.T) {
	repo, _ := sampleMemoryRepo()
	svc := NewService(repo, nil)

	resp, err := svc.ListSituations(context.Background(), ListSituationsRequest{Limit: 1})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "airport", resp.Items[0].Slug)
	assert.NotEmpty(t, resp.NextCursor)

	resp2, err := svc.ListSituations(context.Background(), ListSituationsRequest{AfterCursor: resp.NextCursor, Limit: 1})
	require.NoError(t, err)
	require.Len(t, resp2.Items, 1)
	assert.Equal(t, "restaurant", resp2.Items[0].Slug)
	assert.Empty(t, resp2.NextCursor)
}

func TestServiceListSituationsTiedOrderAndExhaustedCursor(t *testing.T) {
	repo, _ := sampleMemoryRepo()
	repo.situations = append(repo.situations, Situation{
		ID: MustParseUUID("00000000-0000-0000-0000-000000000006"), Slug: "hotel", Title: "Hotel", Status: "active", DisplayOrder: 1,
	})
	svc := NewService(repo, nil)

	first, err := svc.ListSituations(t.Context(), ListSituationsRequest{Limit: 1})
	require.NoError(t, err)
	second, err := svc.ListSituations(t.Context(), ListSituationsRequest{AfterCursor: first.NextCursor, Limit: 1})
	require.NoError(t, err)
	require.Len(t, second.Items, 1)
	assert.Equal(t, "hotel", second.Items[0].Slug)

	exhausted := encodeSituationCursor(situationCursor{DisplayOrder: 99, ID: MustParseUUID("00000000-0000-0000-0000-000000000001")})
	empty, err := svc.ListSituations(t.Context(), ListSituationsRequest{AfterCursor: exhausted})
	require.NoError(t, err)
	assert.Empty(t, empty.Items)
	assert.Empty(t, empty.NextCursor)
}

func TestServiceListSituationsPrioritizesCategoryOnFirstPageOnly(t *testing.T) {
	repo, _ := sampleMemoryRepo()
	svc := NewService(repo, nil)

	// "restaurant" (daily_life, display_order 2) is boosted ahead of
	// "airport" (travel, display_order 1) when the requester's goal
	// category matches it (VOC-1183).
	resp, err := svc.ListSituations(context.Background(), ListSituationsRequest{PriorityCategory: "daily_life"})
	require.NoError(t, err)
	require.Len(t, resp.Items, 2)
	assert.Equal(t, "restaurant", resp.Items[0].Slug)
	assert.Equal(t, "airport", resp.Items[1].Slug)

	// No matching category (or no hint at all) leaves plain
	// display_order untouched.
	resp, err = svc.ListSituations(context.Background(), ListSituationsRequest{PriorityCategory: "work"})
	require.NoError(t, err)
	assert.Equal(t, "airport", resp.Items[0].Slug)
	assert.Equal(t, "restaurant", resp.Items[1].Slug)

	resp, err = svc.ListSituations(context.Background(), ListSituationsRequest{})
	require.NoError(t, err)
	assert.Equal(t, "airport", resp.Items[0].Slug)
	assert.Equal(t, "restaurant", resp.Items[1].Slug)
}

func TestServiceListSituationsPriorityOnlyAppliesToFirstPage(t *testing.T) {
	repo, _ := sampleMemoryRepo()
	svc := NewService(repo, nil)

	// Page 1: limit 1, priority "daily_life" — still returns "airport"
	// first because a single-item page has nothing to reorder, but the
	// cursor must still point past "airport" (display_order 1), not be
	// disturbed by prioritization.
	resp, err := svc.ListSituations(context.Background(), ListSituationsRequest{Limit: 1, PriorityCategory: "daily_life"})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "airport", resp.Items[0].Slug)
	require.NotEmpty(t, resp.NextCursor)

	// Page 2 (AfterCursor set): PriorityCategory is ignored past the
	// first page, so ordering is plain display_order and pagination is
	// unaffected by the earlier prioritization.
	resp2, err := svc.ListSituations(context.Background(), ListSituationsRequest{AfterCursor: resp.NextCursor, Limit: 1, PriorityCategory: "daily_life"})
	require.NoError(t, err)
	require.Len(t, resp2.Items, 1)
	assert.Equal(t, "restaurant", resp2.Items[0].Slug)
	assert.Empty(t, resp2.NextCursor)
}

func TestServiceGetSituationNotFound(t *testing.T) {
	repo, _ := sampleMemoryRepo()
	svc := NewService(repo, nil)

	_, err := svc.GetSituation(context.Background(), uuid.Nil, "unknown")
	assert.ErrorIs(t, err, ErrSituationNotFound)
}

func TestServiceGetSituationReturnsMeanings(t *testing.T) {
	repo, reader := sampleMemoryRepo()
	svc := NewService(repo, reader)

	userID := MustParseUUID("00000000-0000-0000-0000-000000000009")
	detail, err := svc.GetSituation(context.Background(), userID, "airport")
	require.NoError(t, err)
	assert.Equal(t, "airport", detail.Situation.Slug)
	require.Len(t, detail.Meanings, 1)
	assert.Equal(t, "boarding-pass", detail.Meanings[0].WordSlug)
	assert.Equal(t, "boarding pass", detail.Meanings[0].WordText)
	assert.True(t, detail.Meanings[0].Saved)
}

func TestServiceGetSituationOrdersCoreMeaningsBeforeDisplayOrderAndRelevance(t *testing.T) {
	situationID := MustParseUUID("00000000-0000-0000-0000-000000000101")
	wordID := MustParseUUID("00000000-0000-0000-0000-000000000102")
	firstID := MustParseUUID("00000000-0000-0000-0000-000000000103")
	coreID := MustParseUUID("00000000-0000-0000-0000-000000000104")
	moreRelevantID := MustParseUUID("00000000-0000-0000-0000-000000000105")
	undisplayedID := MustParseUUID("00000000-0000-0000-0000-000000000106")
	displayOrder := 1

	repo := NewMemoryRepository(MemoryRepositoryData{
		Situations: []Situation{{ID: situationID, Slug: "airport", Status: "active"}},
		Words:      []SeedWord{{ID: wordID, Text: "word", NormalizedText: "word", Status: "active"}},
		Meanings: []SeedMeaning{
			{ID: firstID, WordID: wordID, Status: "active"},
			{ID: coreID, WordID: wordID, Status: "active"},
			{ID: moreRelevantID, WordID: wordID, Status: "active"},
			{ID: undisplayedID, WordID: wordID, Status: "active"},
		},
		JourneyWords: []SeedJourneyWord{
			{JourneySituationID: situationID, MeaningID: firstID, DisplayOrder: &displayOrder, RelevanceScore: 20},
			{JourneySituationID: situationID, MeaningID: coreID, DisplayOrder: &displayOrder, RelevanceScore: 1, IsCore: true},
			{JourneySituationID: situationID, MeaningID: moreRelevantID, DisplayOrder: &displayOrder, RelevanceScore: 90},
			{JourneySituationID: situationID, MeaningID: undisplayedID, RelevanceScore: 100},
		},
	})

	detail, err := NewService(repo, nil).GetSituation(t.Context(), uuid.Nil, "airport")
	require.NoError(t, err)
	require.Len(t, detail.Meanings, 4)
	assert.Equal(t, []uuid.UUID{coreID, moreRelevantID, firstID, undisplayedID}, []uuid.UUID{
		detail.Meanings[0].MeaningID,
		detail.Meanings[1].MeaningID,
		detail.Meanings[2].MeaningID,
		detail.Meanings[3].MeaningID,
	})
}

func TestServiceGetSituationNoSavedReader(t *testing.T) {
	repo, _ := sampleMemoryRepo()
	svc := NewService(repo, nil)

	detail, err := svc.GetSituation(context.Background(), uuid.Nil, "airport")
	require.NoError(t, err)
	require.Len(t, detail.Meanings, 1)
	assert.False(t, detail.Meanings[0].Saved)
}

func TestServiceGetWordDetailNotFound(t *testing.T) {
	repo, _ := sampleMemoryRepo()
	svc := NewService(repo, nil)

	_, err := svc.GetWordDetail(context.Background(), uuid.Nil, "unknown")
	assert.ErrorIs(t, err, ErrWordNotFound)
}

func TestServiceGetWordDetailReturnsMeanings(t *testing.T) {
	repo, reader := sampleMemoryRepo()
	svc := NewService(repo, reader)

	userID := MustParseUUID("00000000-0000-0000-0000-000000000009")
	word, err := svc.GetWordDetail(context.Background(), userID, "boarding-pass")
	require.NoError(t, err)
	assert.Equal(t, "boarding pass", word.Text)
	assert.Equal(t, "boarding-pass", word.Slug)
	require.Len(t, word.Meanings, 1)
	assert.True(t, word.Meanings[0].Saved)
	require.Len(t, word.Meanings[0].Examples, 1)
	assert.Equal(t, "Please have your boarding pass ready.", word.Meanings[0].Examples[0].ExampleText)
	require.Len(t, word.Meanings[0].UsageNotes, 1)
	assert.Equal(t, "collocation", word.Meanings[0].UsageNotes[0].NoteType)
}

func TestServiceGetWordDetailCrossUserSavedState(t *testing.T) {
	repo, reader := sampleMemoryRepo()
	svc := NewService(repo, reader)

	otherUser := MustParseUUID("00000000-0000-0000-0000-00000000000a")
	word, err := svc.GetWordDetail(context.Background(), otherUser, "boarding-pass")
	require.NoError(t, err)
	require.Len(t, word.Meanings, 1)
	assert.False(t, word.Meanings[0].Saved)
}

func intPtr(i int) *int { return &i }
