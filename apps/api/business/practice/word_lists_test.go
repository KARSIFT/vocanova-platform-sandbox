package practice

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestListPracticeSelectsOnlyMembershipWithoutFallback(t *testing.T) {
	words := seedWords(t)
	revision := 2
	req := StartRequest{Mode: "typed_recall", ListID: uuid.NewString(), ListRevision: &revision}
	_, e := build(req, words, map[string]Source{}, "list")
	require.ErrorIs(t, e, ErrListEmpty)
	_, e = build(req, words, map[string]Source{uuid.NewString(): {}}, "unsupported")
	require.ErrorIs(t, e, ErrListEmpty)
	selected := words[0]
	snap, e := build(req, words, map[string]Source{selected.MeaningID: {}}, "list")
	require.NoError(t, e)
	require.Len(t, snap.Steps, 1)
	require.Equal(t, selected.MeaningID, snap.Steps[0].Word.MeaningID)
	req.Mode = "listening_choice"
	snap, e = build(req, words, map[string]Source{selected.MeaningID: {}}, "list")
	require.NoError(t, e)
	require.Len(t, snap.Steps, 1)
	require.Len(t, snap.Steps[0].Public.Choices, 3)
}
func TestListSelectionValidationAndLegacyFingerprint(t *testing.T) {
	svc := NewService(nil, nil)
	id := uuid.NewString()
	revision := 2
	for _, req := range []StartRequest{{Mode: "typed_recall", ListID: "invalid", ListRevision: &revision}, {Mode: "typed_recall", ListID: id}, {Mode: "typed_recall", ListRevision: &revision}, {Mode: "mistakes", ListID: id, ListRevision: &revision}, {Mode: "typed_recall", ListID: id, ListRevision: &revision, LessonKey: "airport"}} {
		_, e := svc.Start(t.Context(), uuid.New(), req, "request")
		require.ErrorIs(t, e, ErrInvalid)
	}
	legacy := struct {
		Mode      string `json:"mode"`
		LessonKey string `json:"lessonKey,omitempty"`
	}{"typed_recall", "airport"}
	require.Equal(t, fingerprint(legacy), fingerprint(StartRequest{Mode: "typed_recall", LessonKey: "airport"}))
}
