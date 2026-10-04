package lessons

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPostgreSQLMixedTypedReplayConcurrencyAndLegacySnapshot(t *testing.T) {
	db := lessonDB(t)
	ctx := t.Context()
	u := uuid.New()
	_, e := db.ExecContext(ctx, `INSERT INTO users(id)VALUES($1)`, u)
	require.NoError(t, e)
	svc := NewService(NewPostgreSQLRepository(db), nil)
	current, e := svc.Start(ctx, u, "conversation-start", "start-mixed")
	require.NoError(t, e)
	require.Equal(t, ExerciseVersion, current.ExerciseVersion)
	id := uuid.MustParse(current.ID)
	act := func(a Action, key string) {
		a.ExpectedRevision = current.Revision
		a.StepID = current.CurrentStep.ID
		a.ClientActionID = key
		var err error
		current, err = svc.Act(ctx, u, id, a, key)
		require.NoError(t, err)
	}
	for current.CurrentStep.Kind == "teach" {
		act(Action{Action: "continue"}, uuid.NewString())
	}
	act(Action{Action: "answer", ChoiceID: current.CurrentStep.Word.MeaningID}, "recall")
	act(Action{Action: "continue"}, "after-recall")
	require.Equal(t, "typed_recall", current.CurrentStep.Kind)
	require.Empty(t, current.CurrentStep.Choices)
	wrong := Action{StepID: current.CurrentStep.ID, ExpectedRevision: current.Revision, ClientActionID: "typed-wrong", Action: "answer", TypedAnswer: "light conversation"}
	current, e = svc.Act(ctx, u, id, wrong, "typed-wrong")
	require.NoError(t, e)
	require.False(t, current.Feedback.Correct)
	require.False(t, current.CanContinue)
	require.Equal(t, "small talk", current.Feedback.Answer)
	persisted, e := NewService(NewPostgreSQLRepository(db), nil).Get(ctx, u, id)
	require.NoError(t, e)
	require.Equal(t, current, persisted)
	replay, e := svc.Act(ctx, u, id, wrong, "typed-wrong")
	require.NoError(t, e)
	require.Equal(t, current, replay)
	conflicting := wrong
	conflicting.TypedAnswer = "small talk"
	_, e = svc.Act(ctx, u, id, conflicting, "typed-wrong")
	require.ErrorIs(t, e, ErrConflict)
	stale := wrong
	stale.ClientActionID = "stale"
	_, e = svc.Act(ctx, u, id, stale, "stale")
	require.ErrorIs(t, e, ErrConflict)
	_, e = svc.Act(ctx, u, id, Action{StepID: current.CurrentStep.ID, ExpectedRevision: current.Revision, ClientActionID: "skip", Action: "continue"}, "skip")
	require.ErrorIs(t, e, ErrConflict)
	good := Action{StepID: current.CurrentStep.ID, ExpectedRevision: current.Revision, ClientActionID: "typed-correct", Action: "answer", TypedAnswer: " SMALL   TALK "}
	var wg sync.WaitGroup
	states := make([]*Session, 2)
	errs := make([]error, 2)
	for i := range states {
		wg.Add(1)
		go func(i int) { defer wg.Done(); states[i], errs[i] = svc.Act(ctx, u, id, good, "typed-correct") }(i)
	}
	wg.Wait()
	for _, e := range errs {
		require.NoError(t, e)
	}
	require.Equal(t, states[0], states[1])
	current = states[0]
	require.True(t, current.Feedback.Correct)
	require.Equal(t, 1, current.FirstAnswersCorrect)
	require.Equal(t, 2, current.QuestionsAnswered)
	_, e = db.ExecContext(ctx, `UPDATE word_meanings SET short_definition='Changed after the session started' WHERE id=$1`, current.CurrentStep.Word.MeaningID)
	require.NoError(t, e)
	persisted, e = svc.Get(ctx, u, id)
	require.NoError(t, e)
	require.Equal(t, current, persisted)
	act(Action{Action: "continue"}, "after-typed")
	require.Equal(t, "listening_choice", current.CurrentStep.Kind)
	require.Equal(t, "casual", current.CurrentStep.SpeechText)
	require.Len(t, current.CurrentStep.Choices, 3)
	// A late retry after advancing must return current listening state, never old feedback.
	replay, e = svc.Act(ctx, u, id, good, "typed-correct")
	require.NoError(t, e)
	require.Equal(t, current, replay)
	var count int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM lesson_actions WHERE session_id=$1 AND client_action_id='typed-correct'`, id).Scan(&count))
	require.Equal(t, 1, count)
	// Insert an original unversioned snapshot exactly as a previous release stored it.
	d, ok := definition("airport")
	require.True(t, ok)
	legacy, e := buildSnapshot(d, seedWords(t, d), "legacy")
	require.NoError(t, e)
	raw, e := json.Marshal(legacy)
	require.NoError(t, e)
	require.NotContains(t, string(raw), "exerciseVersion")
	legacyID := uuid.New()
	_, e = db.ExecContext(ctx, `INSERT INTO lesson_sessions(id,user_id,lesson_key,lesson_version,snapshot,current_step,total_steps,revision,created_at,updated_at)VALUES($1,$2,$3,$4,$5,3,9,3,$6,$6)`, legacyID, u, d.Key, d.Version, string(raw), time.Now())
	require.NoError(t, e)
	old, e := svc.Start(ctx, u, "airport", "resume-legacy")
	require.NoError(t, e)
	require.Empty(t, old.ExerciseVersion)
	require.Equal(t, "recall", old.CurrentStep.Kind)
	require.Equal(t, "recall-1", old.CurrentStep.ID)
	old, e = svc.Act(ctx, u, legacyID, Action{StepID: old.CurrentStep.ID, ExpectedRevision: old.Revision, ClientActionID: "legacy-answer", Action: "answer", ChoiceID: old.CurrentStep.Word.MeaningID}, "legacy-answer")
	require.NoError(t, e)
	require.True(t, old.Feedback.Correct)
	require.Empty(t, old.Feedback.Answer)
	_, e = db.ExecContext(ctx, `UPDATE users SET status='disabled' WHERE id=$1`, u)
	require.NoError(t, e)
	_, e = svc.Get(ctx, u, id)
	require.ErrorIs(t, e, ErrNotFound)
	// Client-visible JSON never serializes private accepted answers.
	visible, e := json.Marshal(current)
	require.NoError(t, e)
	require.False(t, strings.Contains(string(visible), "accepted"))
}

func TestPostgreSQLMixedLessonPrivacyExport(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	db, e := sql.Open("postgres", dsn)
	require.NoError(t, e)
	t.Cleanup(func() { db.Close() })
	ctx := t.Context()
	u := uuid.New()
	_, e = db.ExecContext(ctx, `INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, u, u.String()+"@mixed-export.invalid")
	require.NoError(t, e)
	t.Cleanup(func() {
		for _, table := range []string{"lesson_actions", "lesson_sessions", "users"} {
			column := "user_id"
			if table == "users" {
				column = "id"
			}
			_, e := db.ExecContext(context.Background(), "DELETE FROM "+table+" WHERE "+column+"=$1", u)
			require.NoError(t, e)
		}
	})
	svc := NewService(NewPostgreSQLRepository(db), nil)
	current, e := svc.Start(ctx, u, "conversation-start", "private-start-key")
	require.NoError(t, e)
	id := uuid.MustParse(current.ID)
	for current.CurrentStep.Kind != "typed_recall" {
		a := Action{StepID: current.CurrentStep.ID, ExpectedRevision: current.Revision, ClientActionID: uuid.NewString(), Action: "continue"}
		if !current.CanContinue {
			a.Action = "answer"
			a.ChoiceID = current.CurrentStep.Word.MeaningID
		}
		current, e = svc.Act(ctx, u, id, a, a.ClientActionID)
		require.NoError(t, e)
	}
	_, e = svc.Act(ctx, u, id, Action{StepID: current.CurrentStep.ID, ExpectedRevision: current.Revision, ClientActionID: "private-typed-client-id", Action: "answer", TypedAnswer: "my original typed wording"}, "private-typed-request-key")
	require.NoError(t, e)
	d, ok := definition("airport")
	require.True(t, ok)
	legacy, e := buildSnapshot(d, seedWords(t, d), "legacy-export")
	require.NoError(t, e)
	raw, e := json.Marshal(legacy)
	require.NoError(t, e)
	_, e = db.ExecContext(ctx, `INSERT INTO lesson_sessions(id,user_id,lesson_key,lesson_version,snapshot,total_steps,created_at,updated_at)VALUES($1,$2,$3,$4,$5,9,$6,$6)`, uuid.New(), u, d.Key, d.Version, string(raw), time.Now())
	require.NoError(t, e)
	exported, e := accounts.NewPostgreSQLRepository(db).ExportPersonalData(ctx, u)
	require.NoError(t, e)
	require.Contains(t, string(exported), "my original typed wording")
	require.Contains(t, string(exported), ExerciseVersion)
	for _, hidden := range []string{"private-start-key", "private-typed-client-id", "private-typed-request-key", `"accepted"`, `"explanations"`, `"snapshot"`, `"fingerprint"`} {
		require.NotContains(t, string(exported), hidden)
	}
	var data struct {
		GuidedLessons []map[string]any `json:"guidedLessons"`
	}
	require.NoError(t, json.Unmarshal(exported, &data))
	require.Len(t, data.GuidedLessons, 2)
	for _, session := range data.GuidedLessons {
		if session["lessonKey"] == "airport" {
			require.Nil(t, session["exerciseVersion"])
		} else {
			require.Equal(t, ExerciseVersion, session["exerciseVersion"])
		}
	}
}
