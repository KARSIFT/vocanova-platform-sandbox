package practice

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPostgreSQLScanKnownContentVersions(t *testing.T) {
	for _, tc := range []struct {
		content, grading string
		valid            bool
	}{{"starter-21-v1", "exact-recall-v1", true}, {"starter-90-v2", "exact-recall-v1", true}, {"starter-unknown", "exact-recall-v1", false}, {"starter-21-v1", "unknown-grading", false}} {
		t.Run(tc.content+"/"+tc.grading, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			u, id := uuid.New(), uuid.New()
			now := time.Now().UTC()
			snapshot := Snapshot{Mode: "typed_recall", LessonKey: "airport", ContentVersion: tc.content, GradingVersion: tc.grading, Steps: []privateStep{{Public: Step{ID: "persisted-v1-step", Kind: "typed_recall", Prompt: "Bags and suitcases you take on a trip.", Choices: []Choice{}}, Word: Word{WordText: "luggage"}, Accepted: []string{"luggage"}}}}
			raw, err := json.Marshal(snapshot)
			require.NoError(t, err)
			mock.ExpectQuery("SELECT.*FROM practice_sessions").WithArgs(id, u).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "snapshot", "current_step", "revision", "feedback", "first_answers_correct", "questions_answered", "created_at", "updated_at", "completed_at"}).AddRow(id, u, raw, 0, 0, nil, 0, 0, now, now, nil))
			st, err := NewPostgreSQLRepository(db).Get(t.Context(), u, id)
			if tc.valid {
				require.NoError(t, err)
				require.Equal(t, snapshot, st.Snapshot)
			} else {
				require.ErrorIs(t, err, ErrContentUnavailable)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPracticePostgreSQLResumesAndActsOnHistoricalContentSnapshot(t *testing.T) {
	db := practiceDB(t)
	ctx := t.Context()
	u := uuid.New()
	_, err := db.ExecContext(ctx, `INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, u, u.String()+"@practice-history.invalid")
	require.NoError(t, err)
	t.Cleanup(func() {
		for _, table := range []string{"practice_mistake_resolutions", "practice_actions", "practice_sessions"} {
			_, _ = db.ExecContext(context.Background(), "DELETE FROM "+table+" WHERE user_id=$1", u)
		}
		_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE id=$1", u)
	})
	repo := NewPostgreSQLRepository(db)
	// PostgreSQL stores microsecond precision; use an exact controlled timestamp
	// so equality checks verify replay content, not database timestamp rounding.
	svc := NewService(repo, clock.Fixed{T: time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)})
	initial, err := svc.Start(ctx, u, StartRequest{Mode: "typed_recall", LessonKey: "airport"}, "historical-start")
	require.NoError(t, err)
	id := uuid.MustParse(initial.ID)
	// Airport's original three meanings and snapshot representation are unchanged.
	// Persist the old version and a frozen historical prompt before going through
	// repository scan, GET, list, start replay, grading and action replay.
	_, err = db.ExecContext(ctx, `UPDATE practice_sessions SET snapshot=jsonb_set(jsonb_set(snapshot,'{ContentVersion}','"starter-21-v1"'),'{Steps,0,Public,prompt}','"The original saved airport question"') WHERE id=$1`, id)
	require.NoError(t, err)
	state, err := repo.Get(ctx, u, id)
	require.NoError(t, err)
	require.Equal(t, "starter-21-v1", state.Snapshot.ContentVersion)
	read, err := svc.Get(ctx, u, id)
	require.NoError(t, err)
	require.Equal(t, "The original saved airport question", read.CurrentStep.Prompt)
	listed, err := svc.List(ctx, u)
	require.NoError(t, err)
	require.Len(t, listed.Items, 1)
	require.Equal(t, "starter-21-v1", listed.Items[0].ContentVersion)
	startReplay, err := svc.Start(ctx, u, StartRequest{Mode: "typed_recall", LessonKey: "airport"}, "historical-start")
	require.NoError(t, err)
	require.Equal(t, read, startReplay)
	a := Action{StepID: read.CurrentStep.ID, ExpectedRevision: 0, ClientActionID: "historical-answer", Action: "answer", TypedAnswer: state.Snapshot.Steps[0].Word.WordText}
	answered, err := svc.Act(ctx, u, id, a, "historical-answer")
	require.NoError(t, err)
	require.True(t, answered.Feedback.Correct)
	require.Equal(t, "starter-21-v1", answered.ContentVersion)
	replayed, err := svc.Act(ctx, u, id, a, "historical-answer")
	require.NoError(t, err)
	require.Equal(t, answered, replayed)
	continued, err := svc.Act(ctx, u, id, Action{StepID: read.CurrentStep.ID, ExpectedRevision: 1, ClientActionID: "historical-continue", Action: "continue"}, "historical-continue")
	require.NoError(t, err)
	require.Equal(t, 1, continued.CompletedSteps)
	require.Equal(t, "starter-21-v1", continued.ContentVersion)
	stored, err := repo.Get(ctx, u, id)
	require.NoError(t, err)
	require.Equal(t, state.Snapshot, stored.Snapshot, "actions must not migrate or rebuild a historical snapshot")
}
