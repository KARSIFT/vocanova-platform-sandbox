package practice

import (
	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaninguses"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
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

func TestSupplementalListSupportChecksFullReviewedIdentity(t *testing.T) {
	for _, entry := range meaninguses.Entries() {
		fields := [6]string{entry.MeaningID.String(), entry.WordID.String(), entry.Word, entry.PartOfSpeech, entry.Definition, "en"}
		require.True(t, SupportsListMeaning(fields[0], fields[1], fields[2], fields[3], fields[4], fields[5]))
		for i := range fields {
			changed := fields
			changed[i] = "unreviewed-value"
			require.False(t, SupportsListMeaning(changed[0], changed[1], changed[2], changed[3], changed[4], changed[5]), entry.SourceSenseKey)
		}
	}
}

func TestSupplementalResumeVerbDoesNotAcceptCVNounAliases(t *testing.T) {
	var entry meaninguses.Entry
	for _, candidate := range meaninguses.Entries() {
		if candidate.Word == "resume" && candidate.PartOfSpeech == "verb" {
			entry = candidate
			break
		}
	}
	require.NotEqual(t, uuid.Nil, entry.MeaningID)
	w := Word{MeaningID: entry.MeaningID.String(), WordText: entry.Word, PartOfSpeech: entry.PartOfSpeech, Definition: entry.Definition}
	revision := 1
	snap, err := build(StartRequest{Mode: "typed_recall", ListID: uuid.NewString(), ListRevision: &revision}, []Word{w}, map[string]Source{w.MeaningID: {}}, "resume-verb")
	require.NoError(t, err)
	for _, wrong := range []string{"résumé", "resumé", "re\u0301sume\u0301", "resume\u0301"} {
		st := State{Snapshot: snap}
		_, e := apply(&st, Action{StepID: snap.Steps[0].Public.ID, Action: "answer", TypedAnswer: wrong}, time.Now())
		require.NoError(t, e)
		require.False(t, st.Feedback.Correct, "CV-noun spelling must not answer the verb clue")
		require.Zero(t, st.FirstAnswersCorrect)
	}
	st := State{Snapshot: snap}
	_, err = apply(&st, Action{StepID: snap.Steps[0].Public.ID, Action: "answer", TypedAnswer: "resume"}, time.Now())
	require.NoError(t, err)
	require.True(t, st.Feedback.Correct)
	require.Equal(t, 1, st.FirstAnswersCorrect)
	require.Contains(t, acceptedWordAnswers(Word{WordText: "resume", PartOfSpeech: "noun"}), "résumé")
	require.Contains(t, acceptedWordAnswers(Word{WordText: "resume"}), "résumé", "historical snapshots/builders retain aliases")
}
