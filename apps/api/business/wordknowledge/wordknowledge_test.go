package wordknowledge

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"strings"
	"sync"
	"testing"
)

func TestAssessmentFingerprintRetainsExistingReceipts(t *testing.T) {
	m := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	req := WriteRequest{MeaningID: m, SelfReportedKnown: true, Note: "memo"}
	legacy := sha256.Sum256([]byte(`{"MeaningID":"00000000-0000-0000-0000-000000000001","Known":true,"Note":"memo","Delete":false}`))
	require.Equal(t, hex.EncodeToString(legacy[:]), fingerprint(req))
	req.Note = ""
	req.SelfReportedKnown = false
	req.Delete = true
	legacy = sha256.Sum256([]byte(`{"MeaningID":"00000000-0000-0000-0000-000000000001","Known":false,"Note":"","Delete":true}`))
	require.Equal(t, hex.EncodeToString(legacy[:]), fingerprint(req))
}

func TestPrivateKnowledgeIsolationReplayAndClear(t *testing.T) {
	m, u, other := uuid.New(), uuid.New(), uuid.New()
	s := NewService(NewMemoryRepository([]uuid.UUID{m}), nil)
	st, err := s.Get(t.Context(), u, m)
	require.NoError(t, err)
	require.False(t, st.SelfReportedKnown)
	require.Nil(t, st.UpdatedAt)
	req := WriteRequest{UserID: u, MeaningID: m, SelfReportedKnown: true, Note: "  My private memory  ", IdempotencyKey: "same-key"}
	st, err = s.Write(t.Context(), req)
	require.NoError(t, err)
	require.Equal(t, "My private memory", st.Note)
	foreign, err := s.Get(t.Context(), other, m)
	require.NoError(t, err)
	require.Empty(t, foreign.Note)
	require.False(t, foreign.SelfReportedKnown)
	reqOther := req
	reqOther.UserID = other
	reqOther.Note = "Other note"
	_, err = s.Write(t.Context(), reqOther)
	require.NoError(t, err)
	changed := req
	changed.Note = "different"
	_, err = s.Write(t.Context(), changed)
	require.ErrorIs(t, err, ErrConflict)
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() { defer wg.Done(); _, e := s.Write(t.Context(), req); require.NoError(t, e) }()
	}
	wg.Wait()
	count, err := s.CountKnown(t.Context(), u)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	_, err = s.Write(t.Context(), WriteRequest{UserID: u, MeaningID: m, Delete: true, IdempotencyKey: "clear"})
	require.NoError(t, err)
	st, err = s.Write(t.Context(), req)
	require.NoError(t, err)
	require.Empty(t, st.Note)
	require.False(t, st.SelfReportedKnown)
	require.Nil(t, st.UpdatedAt)
	count, err = s.CountKnown(t.Context(), u)
	require.NoError(t, err)
	require.Zero(t, count)
	foreign, err = s.Get(t.Context(), other, m)
	require.NoError(t, err)
	require.Equal(t, "Other note", foreign.Note)
	for _, note := range []string{strings.Repeat("é", 2001), "bad\x00note", string([]byte{0xff})} {
		req.Note = note
		req.IdempotencyKey = "invalid"
		_, err = s.Write(t.Context(), req)
		require.ErrorIs(t, err, ErrInvalid)
	}
	_, err = s.Get(t.Context(), u, uuid.New())
	require.ErrorIs(t, err, ErrNotFound)
}

func TestAssessmentOnlyPreservesNotesAndNewerState(t *testing.T) {
	m, u := uuid.New(), uuid.New()
	s := NewService(NewMemoryRepository([]uuid.UUID{m}), nil)
	_, err := s.Write(t.Context(), WriteRequest{UserID: u, MeaningID: m, Note: "My own memory", IdempotencyKey: "note"})
	require.NoError(t, err)
	assessment := WriteRequest{UserID: u, MeaningID: m, SelfReportedKnown: true, AssessmentOnly: true, IdempotencyKey: "assessment"}
	st, err := s.Write(t.Context(), assessment)
	require.NoError(t, err)
	require.True(t, st.SelfReportedKnown)
	require.Equal(t, "My own memory", st.Note)
	// PUT and PATCH are different requests even with an empty note.
	replacement := assessment
	replacement.AssessmentOnly = false
	_, err = s.Write(t.Context(), replacement)
	require.ErrorIs(t, err, ErrConflict)
	_, err = s.Write(t.Context(), WriteRequest{UserID: u, MeaningID: m, Note: "Newer note", SelfReportedKnown: false, IdempotencyKey: "newer"})
	require.NoError(t, err)
	st, err = s.Write(t.Context(), assessment)
	require.NoError(t, err)
	require.False(t, st.SelfReportedKnown)
	require.Equal(t, "Newer note", st.Note)
	changed := assessment
	changed.SelfReportedKnown = false
	_, err = s.Write(t.Context(), changed)
	require.ErrorIs(t, err, ErrConflict)
	changed.IdempotencyKey = "unmark"
	st, err = s.Write(t.Context(), changed)
	require.NoError(t, err)
	require.False(t, st.SelfReportedKnown)
	require.Equal(t, "Newer note", st.Note)
	_, err = s.Write(t.Context(), WriteRequest{UserID: u, MeaningID: m, Delete: true, IdempotencyKey: "clear"})
	require.NoError(t, err)
	st, err = s.Write(t.Context(), assessment)
	require.NoError(t, err)
	require.False(t, st.SelfReportedKnown)
	require.Empty(t, st.Note)
	require.Nil(t, st.UpdatedAt)
	assessment.IdempotencyKey = "fresh"
	st, err = s.Write(t.Context(), assessment)
	require.NoError(t, err)
	require.True(t, st.SelfReportedKnown)
	require.Empty(t, st.Note)
	for _, invalid := range []WriteRequest{
		{UserID: u, MeaningID: m, AssessmentOnly: true, Delete: true, IdempotencyKey: "bad-delete"},
		{UserID: u, MeaningID: m, AssessmentOnly: true, Note: "Must not replace", IdempotencyKey: "bad-note"},
	} {
		_, err = s.Write(t.Context(), invalid)
		require.ErrorIs(t, err, ErrInvalid)
	}
}
