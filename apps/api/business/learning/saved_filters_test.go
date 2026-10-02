package learning

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func savedFilterFixture() (MemoryRepositoryData, uuid.UUID, uuid.UUID) {
	owner, other := uuid.New(), uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	past, future := now.Add(-time.Hour), now.Add(time.Hour)
	data := MemoryRepositoryData{}
	add := func(text, status string, reviews int, next *time.Time, user uuid.UUID, deleted *time.Time) {
		word, meaning := uuid.New(), uuid.New()
		data.Words = append(data.Words, MemoryWord{ID: word, Text: text, NormalizedText: strings.ToLower(text), Status: "archived"})
		data.Meanings = append(data.Meanings, MemoryMeaning{ID: meaning, WordID: word, PartOfSpeech: "noun", ShortDefinition: "Money returned after a purchase.", Status: "archived"})
		data.UserWords = append(data.UserWords, MemoryUserWord{ID: uuid.New(), UserID: user, MeaningID: meaning, Status: status, TotalReviewCount: reviews, NextReviewAt: next, DeletedAt: deleted, Source: "manual", AddedAt: now.Add(-time.Duration(len(data.UserWords)) * time.Minute)})
	}
	for range 55 {
		add("refund", "new", 0, nil, owner, nil)
	}
	add("refund progress", "new", 1, &future, owner, nil)
	add("refund learning", "learning", 1, &past, owner, nil)
	add("refund review", "reviewing", 2, &past, owner, nil)
	add("refund later", "reviewing", 2, &future, owner, nil)
	add("refund complete", "mastered", 5, &past, owner, nil)
	add("refund ignored", "ignored", 0, nil, owner, nil)
	add("refund archived", "archived", 0, nil, owner, nil)
	add("rate_100%", "mastered", 5, nil, owner, nil)
	add("refund removed", "learning", 1, &past, owner, &now)
	add("refund foreign", "new", 0, nil, other, nil)
	return data, owner, other
}

func TestSavedCollectionFiltersFullSetStagesAndDue(t *testing.T) {
	data, owner, other := savedFilterFixture()
	repo := NewMemoryRepository(data)
	for _, query := range []struct {
		stage string
		due   bool
		count int
	}{
		{"", false, 63}, {"new", false, 55}, {"learning", false, 2}, {"reviewing", false, 2}, {"mastered", false, 2}, {"ignored", false, 1}, {"archived", false, 1},
		{"", true, 57}, {"learning", true, 1}, {"reviewing", true, 1}, {"mastered", true, 0}, {"ignored", true, 0}, {"archived", true, 0},
	} {
		response, err := repo.ListSavedWords(t.Context(), ListSavedWordsRequest{UserID: owner, Stage: query.stage, DueOnly: query.due, Limit: 1})
		require.NoError(t, err)
		require.Equal(t, query.count, response.TotalCount, "stage=%s due=%v", query.stage, query.due)
		for _, item := range response.Items {
			if query.stage != "" {
				require.Equal(t, query.stage, item.ReviewState)
			}
			if query.due {
				require.True(t, item.Due)
			}
		}
	}
	request := ListSavedWordsRequest{UserID: owner, Query: " REFUND ", Limit: 50}
	first, err := repo.ListSavedWords(t.Context(), request)
	require.NoError(t, err)
	require.Equal(t, 62, first.TotalCount)
	require.Len(t, first.Items, 50)
	request.AfterCursor = first.NextCursor
	second, err := repo.ListSavedWords(t.Context(), request)
	require.NoError(t, err)
	require.Equal(t, 62, second.TotalCount)
	require.Len(t, second.Items, 12)
	require.Empty(t, second.NextCursor)
	seen := map[uuid.UUID]bool{}
	for _, item := range append(first.Items, second.Items...) {
		require.False(t, seen[item.UserWordID])
		seen[item.UserWordID] = true
	}
	foreign, err := repo.ListSavedWords(t.Context(), ListSavedWordsRequest{UserID: other, Query: "refund"})
	require.NoError(t, err)
	require.Equal(t, 1, foreign.TotalCount)
}

func TestSavedCollectionFiltersLiteralSearchAndValidation(t *testing.T) {
	data, owner, _ := savedFilterFixture()
	repo := NewMemoryRepository(data)
	for _, q := range []string{"%", "_", "_100%"} {
		result, err := repo.ListSavedWords(t.Context(), ListSavedWordsRequest{UserID: owner, Query: q})
		require.NoError(t, err)
		require.Equal(t, 1, result.TotalCount)
		require.Equal(t, "rate_100%", result.Items[0].WordText)
	}
	result, err := repo.ListSavedWords(t.Context(), ListSavedWordsRequest{UserID: owner, Query: " RETURNED "})
	require.NoError(t, err)
	require.Equal(t, 63, result.TotalCount)
	_, err = repo.ListSavedWords(t.Context(), ListSavedWordsRequest{UserID: owner, Query: strings.Repeat("é", 100)})
	require.NoError(t, err, "limit is Unicode characters rather than UTF-8 bytes")
	for _, q := range []string{strings.Repeat("é", 101), "refund\x00", string([]byte{0xff})} {
		_, err = repo.ListSavedWords(t.Context(), ListSavedWordsRequest{UserID: owner, Query: q})
		require.ErrorIs(t, err, ErrInvalidSavedFilter)
	}
	for _, stage := range []string{"known", "Learning", "due", "new OR true"} {
		_, err = repo.ListSavedWords(t.Context(), ListSavedWordsRequest{UserID: owner, Stage: stage})
		require.ErrorIs(t, err, ErrInvalidSavedFilter)
	}
}

func TestSavedCollectionFiltersCursorBindingDeletionAndExhaustion(t *testing.T) {
	data, owner, other := savedFilterFixture()
	repo := NewMemoryRepository(data)
	request := ListSavedWordsRequest{UserID: owner, Query: "refund", Limit: 1}
	first, err := repo.ListSavedWords(t.Context(), request)
	require.NoError(t, err)
	for _, changed := range []ListSavedWordsRequest{
		{UserID: other, Query: "refund"}, {UserID: owner, Query: "returned"}, {UserID: owner, Query: "refund", Stage: "new"}, {UserID: owner, Query: "refund", DueOnly: true},
	} {
		changed.AfterCursor = first.NextCursor
		_, err = repo.ListSavedWords(t.Context(), changed)
		require.ErrorIs(t, err, ErrInvalidCursor)
	}
	nextRequest := request
	nextRequest.AfterCursor = first.NextCursor
	before, err := repo.ListSavedWords(t.Context(), nextRequest)
	require.NoError(t, err)
	now := time.Now()
	for i := range repo.userWords {
		if repo.userWords[i].ID == first.Items[0].UserWordID {
			repo.userWords[i].DeletedAt = &now
		}
	}
	after, err := repo.ListSavedWords(t.Context(), nextRequest)
	require.NoError(t, err)
	require.Equal(t, before.Items[0].UserWordID, after.Items[0].UserWordID)
	require.Equal(t, 61, after.TotalCount, "deleting cursor boundary changes count, not the next row")
	normalized, _, err := prepareSavedWords(request)
	require.NoError(t, err)
	exhaustedRequest := request
	exhaustedRequest.AfterCursor = nextSavedCursor(normalized, SavedMeaning{UserWordID: uuid.New(), AddedAt: time.Time{}})
	exhausted, err := repo.ListSavedWords(t.Context(), exhaustedRequest)
	require.NoError(t, err)
	require.Empty(t, exhausted.Items)
	require.Empty(t, exhausted.NextCursor)
	require.Equal(t, 61, exhausted.TotalCount)
	for _, cursor := range []string{"broken", strings.Repeat("x", 2049), encodeSavedCursor(savedCursor{ID: uuid.New(), AddedAt: time.Now()})} {
		invalid := request
		invalid.AfterCursor = cursor
		_, err = repo.ListSavedWords(t.Context(), invalid)
		require.ErrorIs(t, err, ErrInvalidCursor)
	}
}
