package practice

import (
	"context"
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/lessons"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestPracticePostgreSQLMixedTypedLessonMistakeTargetAndResolution(t *testing.T) {
	db := practiceDB(t)
	ctx := t.Context()
	raw, e := os.ReadFile("../../cmd/seed/voc026-p1.json")
	require.NoError(t, e)
	var seed map[string][]map[string]any
	require.NoError(t, json.Unmarshal(raw, &seed))
	// New lessons also require their authored examples and active situation links.
	for _, table := range []string{"journey_situations", "word_examples", "usage_notes", "journey_words"} {
		for _, row := range seed[table] {
			row["created_at"] = "2026-10-04T12:00:00Z"
			row["updated_at"] = "2026-10-04T12:00:00Z"
			value, e := json.Marshal(row)
			require.NoError(t, e)
			_, e = db.ExecContext(ctx, "INSERT INTO "+table+" SELECT * FROM jsonb_populate_record(NULL::"+table+",$1::jsonb) ON CONFLICT(id) DO NOTHING", string(value))
			require.NoError(t, e)
		}
	}
	u, other := uuid.New(), uuid.New()
	for _, user := range []uuid.UUID{u, other} {
		_, e = db.ExecContext(ctx, `INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, user, user.String()+"@typed-mistake.invalid")
		require.NoError(t, e)
	}
	t.Cleanup(func() {
		for _, user := range []uuid.UUID{u, other} {
			for _, table := range []string{"practice_mistake_resolutions", "practice_actions", "practice_sessions", "lesson_actions", "lesson_sessions", "users"} {
				column := "user_id"
				if table == "users" {
					column = "id"
				}
				_, e := db.ExecContext(context.Background(), "DELETE FROM "+table+" WHERE "+column+"=$1", user)
				require.NoError(t, e)
			}
		}
	})
	lsvc := lessons.NewService(lessons.NewPostgreSQLRepository(db), nil)
	current, e := lsvc.Start(ctx, u, "conversation-start", "lesson-start")
	require.NoError(t, e)
	id := uuid.MustParse(current.ID)
	for current.CurrentStep.Kind != "typed_recall" {
		a := lessons.Action{StepID: current.CurrentStep.ID, ExpectedRevision: current.Revision, ClientActionID: uuid.NewString(), Action: "continue"}
		if !current.CanContinue {
			a.Action = "answer"
			a.ChoiceID = current.CurrentStep.Word.MeaningID
		}
		current, e = lsvc.Act(ctx, u, id, a, a.ClientActionID)
		require.NoError(t, e)
	}
	target := current.CurrentStep.Word.MeaningID
	require.Equal(t, "small talk", current.CurrentStep.Word.WordText)
	wrong := lessons.Action{StepID: current.CurrentStep.ID, ExpectedRevision: current.Revision, ClientActionID: "typed-wrong", Action: "answer", TypedAnswer: "casual"}
	current, e = lsvc.Act(ctx, u, id, wrong, "typed-wrong")
	require.NoError(t, e)
	require.False(t, current.Feedback.Correct)
	require.Empty(t, current.Feedback.CorrectChoiceID)
	psvc := NewService(NewPostgreSQLRepository(db), nil)
	list, e := psvc.List(ctx, u)
	require.NoError(t, e)
	require.Equal(t, 1, list.AvailableMistakes, "wrong typed lesson target must be available for later practice")
	foreign, e := psvc.List(ctx, other)
	require.NoError(t, e)
	require.Zero(t, foreign.AvailableMistakes)
	corrected := wrong
	corrected.ExpectedRevision = current.Revision
	corrected.ClientActionID = "typed-correct"
	corrected.TypedAnswer = "small talk"
	_, e = lsvc.Act(ctx, u, id, corrected, "typed-correct")
	require.NoError(t, e)
	list, e = psvc.List(ctx, u)
	require.NoError(t, e)
	require.Equal(t, 1, list.AvailableMistakes, "a guided correct retry does not erase the first mistake history")
	practice, e := psvc.Start(ctx, u, StartRequest{Mode: "mistakes"}, "mistake-practice")
	require.NoError(t, e)
	require.Equal(t, 1, practice.TotalSteps)
	answer := Action{StepID: practice.CurrentStep.ID, ExpectedRevision: practice.Revision, ClientActionID: "resolve", Action: "answer", TypedAnswer: "small talk"}
	resolved, e := psvc.Act(ctx, u, uuid.MustParse(practice.ID), answer, "resolve")
	require.NoError(t, e)
	require.True(t, resolved.Feedback.Correct)
	require.Equal(t, target, resolved.Feedback.MeaningID)
	list, e = psvc.List(ctx, u)
	require.NoError(t, e)
	require.Zero(t, list.AvailableMistakes)
	_, e = lsvc.Act(ctx, u, id, wrong, "typed-wrong")
	require.NoError(t, e)
	list, e = psvc.List(ctx, u)
	require.NoError(t, e)
	require.Zero(t, list.AvailableMistakes, "exact historical replay cannot resurrect a resolved mistake")
}
