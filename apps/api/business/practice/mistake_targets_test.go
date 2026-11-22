package practice

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func rankedWord() Word {
	return Word{MeaningID: uuid.NewString(), WordText: "harbor", WordSlug: "harbor", PartOfSpeech: "noun", Definition: "a sheltered area of water where ships can stay safely"}
}

func TestMistakesRankedTargetUsesFrozenRecognitionFallback(t *testing.T) {
	pool := seedWords(t)
	ranked := rankedWord()
	starter := pool[0]
	sources := map[string]Source{
		starter.MeaningID: {Kind: "lesson", ID: uuid.NewString()},
		ranked.MeaningID:  {Kind: "review", ID: uuid.NewString()},
	}
	targets := &rankedTargets{words: []Word{ranked}, recognition: map[string]bool{ranked.MeaningID: true}, pool: pool}
	words := append(append([]Word{}, pool...), ranked)
	snap, err := buildWith(StartRequest{Mode: "mistakes"}, words, sources, "ranked-seed", targets)
	require.NoError(t, err)
	require.Equal(t, MistakesContentVersion, snap.ContentVersion)
	require.Len(t, snap.Steps, 2)
	kinds := map[string]privateStep{}
	for _, step := range snap.Steps {
		kinds[step.Word.MeaningID] = step
		require.NotNil(t, step.Source, "every mistakes step keeps its exact source")
		require.Equal(t, sources[step.Word.MeaningID], *step.Source)
	}
	require.Equal(t, "typed_recall", kinds[starter.MeaningID].Public.Kind, "starter mistakes keep typed recall")
	recognition := kinds[ranked.MeaningID]
	require.Equal(t, "meaning_choice", recognition.Public.Kind)
	require.Equal(t, "harbor", recognition.Public.WordText)
	require.Equal(t, "harbor", recognition.Public.SpeechText, "recognition never claims missing audio")
	require.GreaterOrEqual(t, len(recognition.Public.Choices), 3)
	require.NotEmpty(t, recognition.CorrectChoice)

	// Public projection hides the answer key and the frozen source.
	for i := range snap.Steps {
		st := State{ID: uuid.New(), Snapshot: snap, Index: i}
		raw, err := json.Marshal(project(st))
		require.NoError(t, err)
		for _, secret := range []string{"CorrectChoice", "Accepted", "Source", sources[ranked.MeaningID].ID} {
			require.NotContains(t, string(raw), secret)
		}
	}

	// A wrong first answer never resolves; a correct unassisted one does.
	wrong := ""
	for _, c := range recognition.Public.Choices {
		if c.ID != recognition.CorrectChoice {
			wrong = c.ID
			break
		}
	}
	st := State{Snapshot: Snapshot{Mode: "mistakes", Steps: []privateStep{recognition}}}
	resolved, err := apply(&st, Action{StepID: recognition.Public.ID, Action: "answer", ChoiceID: wrong}, time.Now())
	require.NoError(t, err)
	require.False(t, resolved)
	require.False(t, st.Feedback.Correct)
	resolved, err = apply(&st, Action{StepID: recognition.Public.ID, ExpectedRevision: 1, Action: "answer", ChoiceID: recognition.CorrectChoice}, time.Now())
	require.NoError(t, err)
	require.False(t, resolved, "a retry after a miss is assisted")
	require.True(t, st.Feedback.Assisted)

	st = State{Snapshot: Snapshot{Mode: "mistakes", Steps: []privateStep{recognition}}}
	resolved, err = apply(&st, Action{StepID: recognition.Public.ID, Action: "answer", ChoiceID: recognition.CorrectChoice}, time.Now())
	require.NoError(t, err)
	require.True(t, resolved)
	require.Equal(t, recognition.CorrectChoice, project(st).Feedback.CorrectChoiceID)

	st = State{Snapshot: Snapshot{Mode: "mistakes", Steps: []privateStep{recognition}}}
	resolved, err = apply(&st, Action{StepID: recognition.Public.ID, Action: "reveal"}, time.Now())
	require.NoError(t, err)
	require.False(t, resolved, "revealing never resolves a mistake")
}

func TestMistakesStarterOnlyKeepsOriginalContentVersion(t *testing.T) {
	pool := seedWords(t)
	sources := map[string]Source{pool[3].MeaningID: {Kind: "review", ID: uuid.NewString()}}
	snap, err := buildWith(StartRequest{Mode: "mistakes"}, pool, sources, "starter-seed", &rankedTargets{recognition: map[string]bool{}})
	require.NoError(t, err)
	require.Equal(t, ContentVersion, snap.ContentVersion)
	require.Len(t, snap.Steps, 1)
	require.Equal(t, "typed_recall", snap.Steps[0].Public.Kind)

	// A ranked target filtered by lesson key cannot enter a lesson session.
	ranked := rankedWord()
	sources[ranked.MeaningID] = Source{Kind: "review", ID: uuid.NewString()}
	targets := &rankedTargets{words: []Word{ranked}, recognition: map[string]bool{ranked.MeaningID: true}, pool: pool}
	snap, err = buildWith(StartRequest{Mode: "mistakes", LessonKey: pool[3].LessonKey}, append(append([]Word{}, pool...), ranked), sources, "lesson-seed", targets)
	require.NoError(t, err)
	require.Equal(t, ContentVersion, snap.ContentVersion)
	require.Len(t, snap.Steps, 1)
}

func TestMistakesTypedCueSafetyAndSupportedAvailability(t *testing.T) {
	require.True(t, cueHidesAnswer(rankedWord().Definition, rankedWord().WordText))
	require.False(t, cueHidesAnswer("a light that you switch on", "light"))
	require.False(t, cueHidesAnswer("  ", "light"))
	require.False(t, cueHidesAnswer("Water that flows toward another river.", "river"))
	require.False(t, cueHidesAnswer("The driver’s “seat-belt” can protect them.", "seat belt"))
	require.True(t, cueHidesAnswer("The body part for hearing.", "ear"))

	pool := seedWords(t)
	ranked := rankedWord()
	unsupported := uuid.NewString()
	sources := map[string]Source{
		pool[0].MeaningID: {Kind: "practice", ID: uuid.NewString()},
		ranked.MeaningID:  {Kind: "review", ID: uuid.NewString()},
		unsupported:       {Kind: "review", ID: uuid.NewString()},
	}
	got := supportedMistakes(sources, &rankedTargets{words: []Word{ranked}, recognition: map[string]bool{ranked.MeaningID: true}})
	require.Len(t, got, 2, "availability counts only targets a session can include")
	require.Contains(t, got, pool[0].MeaningID)
	require.Contains(t, got, ranked.MeaningID)
	require.NotContains(t, got, unsupported)
}

func TestMistakesContentVersionLoadsOnlyForMistakesSnapshots(t *testing.T) {
	pool := seedWords(t)
	ranked := rankedWord()
	sources := map[string]Source{ranked.MeaningID: {Kind: "review", ID: uuid.NewString()}}
	snap, err := buildWith(StartRequest{Mode: "mistakes"}, append(append([]Word{}, pool...), ranked), sources, "load", &rankedTargets{words: []Word{ranked}, recognition: map[string]bool{ranked.MeaningID: true}, pool: pool})
	require.NoError(t, err)
	raw, err := json.Marshal(snap)
	require.NoError(t, err)
	now := time.Now()
	row := fakeRow{values: []any{uuid.New(), uuid.New(), raw, 0, 0, []byte(nil), 0, 0, now, now, (*time.Time)(nil)}}
	st, err := scan(row)
	require.NoError(t, err)
	require.Equal(t, MistakesContentVersion, st.Snapshot.ContentVersion)

	snap.Mode = "typed_recall"
	raw, err = json.Marshal(snap)
	require.NoError(t, err)
	row.values[2] = raw
	_, err = scan(row)
	require.ErrorIs(t, err, ErrContentUnavailable)
}

type fakeRow struct{ values []any }

func (r fakeRow) Scan(dest ...any) error {
	for i, d := range dest {
		switch target := d.(type) {
		case *uuid.UUID:
			*target = r.values[i].(uuid.UUID)
		case *[]byte:
			*target = r.values[i].([]byte)
		case *int:
			*target = r.values[i].(int)
		case *time.Time:
			*target = r.values[i].(time.Time)
		case **time.Time:
			*target = r.values[i].(*time.Time)
		}
	}
	return nil
}
