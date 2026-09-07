package learning

import (
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleLearningRepo() (*MemoryRepository, *MemoryIdempotencyStore) {
	wordID := MustParseUUID("00000000-0000-0000-0000-000000000001")
	meaningID := MustParseUUID("00000000-0000-0000-0000-000000000002")
	repo := NewMemoryRepository(MemoryRepositoryData{
		Words: []MemoryWord{
			{ID: wordID, Text: "boarding pass", NormalizedText: "boarding pass", Status: "active"},
		},
		Meanings: []MemoryMeaning{
			{ID: meaningID, WordID: wordID, PartOfSpeech: "noun", ShortDefinition: "A document.", Status: "active"},
		},
	})
	return repo, NewMemoryIdempotencyStore()
}

func TestServiceSaveUserWordCreatesRow(t *testing.T) {
	repo, idem := sampleLearningRepo()
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	svc := NewService(repo, idem, &clock.Fixed{T: now})

	userID := MustParseUUID("00000000-0000-0000-0000-000000000003")
	meaningID := MustParseUUID("00000000-0000-0000-0000-000000000002")

	m, err := svc.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID:         userID,
		MeaningID:      meaningID,
		Source:         "journey",
		IdempotencyKey: "key-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "new", m.Status)
	assert.Equal(t, "journey", m.Source)
	assert.True(t, m.Saved)
	assert.True(t, m.AddedAt.Equal(now))

	saved, err := svc.IsSaved(t.Context(), userID, []uuid.UUID{meaningID})
	require.NoError(t, err)
	assert.True(t, saved[meaningID])
}

func TestMemorySavedWordStateLabelsReviewedLegacyNewWordAsLearning(t *testing.T) {
	repo, _ := sampleLearningRepo()
	userID := MustParseUUID("00000000-0000-0000-0000-000000000003")
	meaningID := MustParseUUID("00000000-0000-0000-0000-000000000002")
	nextReviewAt := time.Now().Add(time.Hour)
	repo.userWords = append(repo.userWords, MemoryUserWord{
		ID:               uuid.New(),
		UserID:           userID,
		MeaningID:        meaningID,
		Status:           "new",
		TotalReviewCount: 1,
		NextReviewAt:     &nextReviewAt,
	})

	states, err := repo.SavedWordStates(t.Context(), userID, []uuid.UUID{meaningID})
	require.NoError(t, err)
	assert.Equal(t, "learning", states[meaningID].Status)
	assert.False(t, states[meaningID].Due)
}

func TestServiceSaveUserWordRequiresIdempotencyKey(t *testing.T) {
	repo, idem := sampleLearningRepo()
	svc := NewService(repo, idem, clock.Real{})

	_, err := svc.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID:    MustParseUUID("00000000-0000-0000-0000-000000000003"),
		MeaningID: MustParseUUID("00000000-0000-0000-0000-000000000002"),
		Source:    "journey",
	})
	assert.ErrorIs(t, err, ErrIdempotencyKeyRequired)
}

func TestServiceSaveUserWordUnknownMeaning(t *testing.T) {
	repo, idem := sampleLearningRepo()
	svc := NewService(repo, idem, clock.Real{})

	_, err := svc.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID:         MustParseUUID("00000000-0000-0000-0000-000000000003"),
		MeaningID:      MustParseUUID("00000000-0000-0000-0000-000000000099"),
		Source:         "journey",
		IdempotencyKey: "key-1",
	})
	assert.ErrorIs(t, err, ErrMeaningNotFound)
}

func TestServiceSaveUserWordAlreadySavedIsIdempotent(t *testing.T) {
	repo, idem := sampleLearningRepo()
	svc := NewService(repo, idem, clock.Real{})

	userID := MustParseUUID("00000000-0000-0000-0000-000000000003")
	meaningID := MustParseUUID("00000000-0000-0000-0000-000000000002")

	first, err := svc.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID:         userID,
		MeaningID:      meaningID,
		Source:         "journey",
		IdempotencyKey: "key-1",
	})
	require.NoError(t, err)

	second, err := svc.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID:         userID,
		MeaningID:      meaningID,
		Source:         "manual",
		IdempotencyKey: "key-2",
	})
	require.NoError(t, err)
	assert.Equal(t, first.UserWordID, second.UserWordID)
}

func TestServiceSaveUserWordIdempotencyKeyReplay(t *testing.T) {
	repo, idem := sampleLearningRepo()
	svc := NewService(repo, idem, clock.Real{})

	userID := MustParseUUID("00000000-0000-0000-0000-000000000003")
	meaningID := MustParseUUID("00000000-0000-0000-0000-000000000002")
	key := "replay-key"

	first, err := svc.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID:         userID,
		MeaningID:      meaningID,
		Source:         "journey",
		IdempotencyKey: key,
	})
	require.NoError(t, err)

	second, err := svc.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID:         userID,
		MeaningID:      meaningID,
		Source:         "journey",
		IdempotencyKey: key,
	})
	require.NoError(t, err)
	assert.Equal(t, first.UserWordID, second.UserWordID)

	_, err = svc.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID:         userID,
		MeaningID:      meaningID,
		Source:         "manual",
		IdempotencyKey: key,
	})
	assert.ErrorIs(t, err, ErrIdempotencyConflict)
}

func TestServiceSaveUserWordIdempotencyKeyIsolatedByUser(t *testing.T) {
	repo, idem := sampleLearningRepo()
	svc := NewService(repo, idem, clock.Real{})

	meaningID := MustParseUUID("00000000-0000-0000-0000-000000000002")
	key := "shared-key"

	_, err := svc.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID:         MustParseUUID("00000000-0000-0000-0000-000000000003"),
		MeaningID:      meaningID,
		Source:         "journey",
		IdempotencyKey: key,
	})
	require.NoError(t, err)

	_, err = svc.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID:         MustParseUUID("00000000-0000-0000-0000-000000000004"),
		MeaningID:      meaningID,
		Source:         "manual",
		IdempotencyKey: key,
	})
	require.NoError(t, err)
}

func TestServiceUnsaveUserWord(t *testing.T) {
	repo, idem := sampleLearningRepo()
	svc := NewService(repo, idem, clock.Real{})

	userID := MustParseUUID("00000000-0000-0000-0000-000000000003")
	meaningID := MustParseUUID("00000000-0000-0000-0000-000000000002")

	_, err := svc.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID:         userID,
		MeaningID:      meaningID,
		Source:         "journey",
		IdempotencyKey: "key-1",
	})
	require.NoError(t, err)

	require.NoError(t, svc.UnsaveUserWord(t.Context(), userID, meaningID))

	saved, err := svc.IsSaved(t.Context(), userID, []uuid.UUID{meaningID})
	require.NoError(t, err)
	assert.False(t, saved[meaningID])
}

func TestServiceSaveUserWordRestoresDeletedWordWithFreshSchedule(t *testing.T) {
	repo, idem := sampleLearningRepo()
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	svc := NewService(repo, idem, &clock.Fixed{T: now})
	userID := MustParseUUID("00000000-0000-0000-0000-000000000003")
	meaningID := MustParseUUID("00000000-0000-0000-0000-000000000002")

	first, err := svc.SaveUserWord(t.Context(), SaveUserWordRequest{UserID: userID, MeaningID: meaningID, Source: "journey", IdempotencyKey: "first-save"})
	require.NoError(t, err)
	require.NoError(t, svc.UnsaveUserWord(t.Context(), userID, meaningID))

	future := now.Add(24 * time.Hour)
	repo.userWords[0].ReviewStep = 7
	repo.userWords[0].NextReviewAt = &future
	repo.userWords[0].LastReviewedAt = &now
	repo.userWords[0].LastResult = "correct"
	repo.userWords[0].LastRating = "easy"
	repo.userWords[0].ConsecutiveCorrectCount = 4
	repo.userWords[0].ConsecutiveIncorrectCount = 1
	repo.userWords[0].TotalReviewCount = 9
	repo.userWords[0].CorrectReviewCount = 8
	repo.userWords[0].MasteredAt = &now
	repo.userWords[0].IgnoredAt = &now

	restored, err := svc.SaveUserWord(t.Context(), SaveUserWordRequest{UserID: userID, MeaningID: meaningID, Source: "manual", IdempotencyKey: "restored-save"})
	require.NoError(t, err)
	assert.Equal(t, first.UserWordID, restored.UserWordID)

	got := repo.userWords[0]
	assert.Equal(t, "new", got.Status)
	assert.Equal(t, 0, got.ReviewStep)
	assert.Nil(t, got.NextReviewAt, "a restored word must be immediately due")
	assert.Nil(t, got.LastReviewedAt)
	assert.Empty(t, got.LastResult)
	assert.Empty(t, got.LastRating)
	assert.Zero(t, got.ConsecutiveCorrectCount)
	assert.Zero(t, got.ConsecutiveIncorrectCount)
	assert.Zero(t, got.TotalReviewCount)
	assert.Zero(t, got.CorrectReviewCount)
	assert.Nil(t, got.MasteredAt)
	assert.Nil(t, got.IgnoredAt)
}

func TestServiceUnsaveUserWordUnknownReturnsNotFound(t *testing.T) {
	repo, idem := sampleLearningRepo()
	svc := NewService(repo, idem, clock.Real{})

	err := svc.UnsaveUserWord(t.Context(),
		MustParseUUID("00000000-0000-0000-0000-000000000003"),
		MustParseUUID("00000000-0000-0000-0000-000000000002"),
	)
	assert.ErrorIs(t, err, ErrUserWordNotFound)
}

func TestServiceUnsaveUserWordCannotTouchOtherUser(t *testing.T) {
	repo, idem := sampleLearningRepo()
	svc := NewService(repo, idem, clock.Real{})

	owner := MustParseUUID("00000000-0000-0000-0000-000000000003")
	other := MustParseUUID("00000000-0000-0000-0000-000000000004")
	meaningID := MustParseUUID("00000000-0000-0000-0000-000000000002")

	_, err := svc.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID:         owner,
		MeaningID:      meaningID,
		Source:         "journey",
		IdempotencyKey: "key-1",
	})
	require.NoError(t, err)

	assert.ErrorIs(t, svc.UnsaveUserWord(t.Context(), other, meaningID), ErrUserWordNotFound)
}

func TestServiceListSavedWords(t *testing.T) {
	repo, idem := sampleLearningRepo()
	svc := NewService(repo, idem, clock.Real{})

	userID := MustParseUUID("00000000-0000-0000-0000-000000000003")
	meaningID := MustParseUUID("00000000-0000-0000-0000-000000000002")

	_, err := svc.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID:         userID,
		MeaningID:      meaningID,
		Source:         "journey",
		IdempotencyKey: "key-1",
	})
	require.NoError(t, err)

	resp, err := svc.ListSavedWords(t.Context(), ListSavedWordsRequest{UserID: userID})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "boarding-pass", resp.Items[0].WordSlug)
	assert.Equal(t, "A document.", resp.Items[0].ShortDefinition)
	assert.True(t, resp.Items[0].Saved)
	assert.Empty(t, resp.NextCursor)
}

func TestServiceListSavedWordsCursor(t *testing.T) {
	repo, idem := sampleLearningRepo()
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	c := &clock.Fixed{T: now}
	svc := NewService(repo, idem, c)
	userID := MustParseUUID("00000000-0000-0000-0000-000000000003")

	wordID := MustParseUUID("00000000-0000-0000-0000-000000000001")
	meaningID1 := MustParseUUID("00000000-0000-0000-0000-000000000002")
	meaningID2 := MustParseUUID("00000000-0000-0000-0000-000000000005")
	repo.meanings = append(repo.meanings, MemoryMeaning{
		ID: meaningID2, WordID: wordID, PartOfSpeech: "noun", ShortDefinition: "Second.", Status: "active",
	})

	_, err := svc.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID: userID, MeaningID: meaningID1, Source: "journey", IdempotencyKey: "k1",
	})
	require.NoError(t, err)
	c.Advance(time.Second)
	_, err = svc.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID: userID, MeaningID: meaningID2, Source: "journey", IdempotencyKey: "k2",
	})
	require.NoError(t, err)

	resp, err := svc.ListSavedWords(t.Context(), ListSavedWordsRequest{UserID: userID, Limit: 1})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, meaningID2, resp.Items[0].MeaningID)
	assert.NotEmpty(t, resp.NextCursor)

	resp2, err := svc.ListSavedWords(t.Context(), ListSavedWordsRequest{UserID: userID, AfterCursor: resp.NextCursor, Limit: 1})
	require.NoError(t, err)
	require.Len(t, resp2.Items, 1)
	assert.Equal(t, meaningID1, resp2.Items[0].MeaningID)
}

func TestServiceListSavedWordsInvalidCursor(t *testing.T) {
	repo, idem := sampleLearningRepo()
	svc := NewService(repo, idem, clock.Real{})

	_, err := svc.ListSavedWords(t.Context(), ListSavedWordsRequest{
		UserID:      MustParseUUID("00000000-0000-0000-0000-000000000003"),
		AfterCursor: "not-valid",
	})
	assert.ErrorIs(t, err, ErrInvalidCursor)
}

func TestServiceListSavedWordsDeletedCursorDoesNotSkipBoundary(t *testing.T) {
	repo, idem := sampleLearningRepo()
	svc := NewService(repo, idem, clock.Real{})
	userID := MustParseUUID("00000000-0000-0000-0000-000000000003")
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)

	for _, id := range []string{"00000000-0000-0000-0000-000000000010", "00000000-0000-0000-0000-000000000011"} {
		repo.meanings = append(repo.meanings, MemoryMeaning{ID: MustParseUUID(id), WordID: repo.words[0].ID, PartOfSpeech: "noun", ShortDefinition: id, Status: "active"})
	}
	repo.userWords = []MemoryUserWord{
		{ID: MustParseUUID("00000000-0000-0000-0000-000000000003"), UserID: userID, MeaningID: MustParseUUID("00000000-0000-0000-0000-000000000010"), AddedAt: now},
		{ID: MustParseUUID("00000000-0000-0000-0000-000000000002"), UserID: userID, MeaningID: MustParseUUID("00000000-0000-0000-0000-000000000011"), AddedAt: now},
		{ID: MustParseUUID("00000000-0000-0000-0000-000000000001"), UserID: userID, MeaningID: MustParseUUID("00000000-0000-0000-0000-000000000002"), AddedAt: now},
	}

	first, err := svc.ListSavedWords(t.Context(), ListSavedWordsRequest{UserID: userID, Limit: 1})
	require.NoError(t, err)
	require.Len(t, first.Items, 1)
	deletedAt := now.Add(time.Second)
	repo.userWords[0].DeletedAt = &deletedAt

	next, err := svc.ListSavedWords(t.Context(), ListSavedWordsRequest{UserID: userID, AfterCursor: first.NextCursor, Limit: 1})
	require.NoError(t, err)
	require.Len(t, next.Items, 1)
	assert.Equal(t, MustParseUUID("00000000-0000-0000-0000-000000000011"), next.Items[0].MeaningID)
}

func TestServiceListSavedWordsExhaustedCursorReturnsEmptyPage(t *testing.T) {
	repo, idem := sampleLearningRepo()
	svc := NewService(repo, idem, clock.Real{})
	repo.userWords = []MemoryUserWord{{
		ID: MustParseUUID("00000000-0000-0000-0000-000000000010"), UserID: MustParseUUID("00000000-0000-0000-0000-000000000003"),
		MeaningID: MustParseUUID("00000000-0000-0000-0000-000000000002"), AddedAt: time.Now(),
	}}
	cursor := encodeSavedCursor(savedCursor{AddedAt: time.Time{}, ID: MustParseUUID("00000000-0000-0000-0000-000000000001")})

	resp, err := svc.ListSavedWords(t.Context(), ListSavedWordsRequest{UserID: MustParseUUID("00000000-0000-0000-0000-000000000003"), AfterCursor: cursor})
	require.NoError(t, err)
	assert.Empty(t, resp.Items)
	assert.Empty(t, resp.NextCursor)
}

func TestServiceIsSavedCrossUserIsolation(t *testing.T) {
	repo, idem := sampleLearningRepo()
	svc := NewService(repo, idem, clock.Real{})

	owner := MustParseUUID("00000000-0000-0000-0000-000000000003")
	other := MustParseUUID("00000000-0000-0000-0000-000000000004")
	meaningID := MustParseUUID("00000000-0000-0000-0000-000000000002")

	_, err := svc.SaveUserWord(t.Context(), SaveUserWordRequest{
		UserID: owner, MeaningID: meaningID, Source: "journey", IdempotencyKey: "key-1",
	})
	require.NoError(t, err)

	states, err := svc.IsSaved(t.Context(), other, []uuid.UUID{meaningID})
	require.NoError(t, err)
	assert.False(t, states[meaningID])
}

func TestSavedCursorRoundTrip(t *testing.T) {
	c := savedCursor{AddedAt: time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC), ID: uuid.MustParse("00000000-0000-0000-0000-000000000001")}
	s := encodeSavedCursor(c)
	decoded, err := decodeSavedCursor(s)
	require.NoError(t, err)
	assert.Equal(t, c, decoded)
}

func TestSavedCursorInvalid(t *testing.T) {
	_, err := decodeSavedCursor("not-valid")
	require.Error(t, err)
}

func MustParseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}
