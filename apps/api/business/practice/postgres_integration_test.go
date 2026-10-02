package practice

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"os"
	"sync"
	"testing"
	"time"
)

func practiceDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	raw, err := os.ReadFile("../../cmd/seed/voc026-p1.json")
	require.NoError(t, err)
	var seed map[string][]map[string]any
	require.NoError(t, json.Unmarshal(raw, &seed))
	for _, table := range []string{"canonical_words", "word_meanings"} {
		for _, row := range seed[table] {
			row["created_at"] = "2026-10-02T12:00:00Z"
			row["updated_at"] = "2026-10-02T12:00:00Z"
			raw, e := json.Marshal(row)
			require.NoError(t, e)
			_, e = db.ExecContext(t.Context(), "INSERT INTO "+table+" SELECT * FROM jsonb_populate_record(NULL::"+table+",$1::jsonb) ON CONFLICT(id) DO NOTHING", string(raw))
			require.NoError(t, e)
		}
	}
	return db
}
func TestPracticePostgreSQLAuthorityReplayMistakesExportPurge(t *testing.T) {
	db := practiceDB(t)
	ctx := t.Context()
	u, other := uuid.New(), uuid.New()
	exec := func(q string, args ...any) {
		t.Helper()
		_, err := db.ExecContext(ctx, q, args...)
		require.NoError(t, err)
	}
	for _, id := range []uuid.UUID{u, other} {
		exec(`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, id, id.String()+"@practice.invalid")
	}
	t.Cleanup(func() {
		for _, id := range []uuid.UUID{u, other} {
			for _, table := range []string{"practice_mistake_resolutions", "practice_actions", "practice_sessions", "lesson_actions", "lesson_sessions", "review_attempts", "user_words"} {
				_, _ = db.ExecContext(context.Background(), "DELETE FROM "+table+" WHERE user_id=$1", id)
			}
			_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE id=$1", id)
		}
	})
	repo := NewPostgreSQLRepository(db)
	svc := NewService(repo, nil)
	req := StartRequest{Mode: "typed_recall", LessonKey: "airport"}
	var wg sync.WaitGroup
	results := make(chan *Session, 8)
	errs := make(chan error, 8)
	for range 8 {
		wg.Add(1)
		go func() { defer wg.Done(); s, e := svc.Start(ctx, u, req, "same-start"); results <- s; errs <- e }()
	}
	wg.Wait()
	close(results)
	close(errs)
	for e := range errs {
		require.NoError(t, e)
	}
	var session *Session
	for s := range results {
		if session != nil {
			require.Equal(t, session.ID, s.ID)
		}
		session = s
	}
	id := uuid.MustParse(session.ID)
	_, err := svc.Get(ctx, other, id)
	require.ErrorIs(t, err, ErrNotFound)
	_, err = svc.Start(ctx, u, StartRequest{Mode: "listening_choice"}, "same-start")
	require.ErrorIs(t, err, ErrConflict)
	raw, err := json.Marshal(session)
	require.NoError(t, err)
	require.NotContains(t, string(raw), `"wordText"`)
	require.NotContains(t, string(raw), "CorrectChoice")
	state, err := repo.Get(ctx, u, id)
	require.NoError(t, err)
	first := state.Snapshot.Steps[0]
	wrong := Action{StepID: first.Public.ID, Action: "answer", TypedAnswer: "private wrong answer", ClientActionID: "wrong"}
	actionResults := make(chan *Session, 8)
	actionErrors := make(chan error, 8)
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, e := svc.Act(ctx, u, id, wrong, "wrong")
			actionResults <- out
			actionErrors <- e
		}()
	}
	wg.Wait()
	close(actionResults)
	close(actionErrors)
	for e := range actionErrors {
		require.NoError(t, e)
	}
	for out := range actionResults {
		require.Equal(t, 1, out.Revision)
		session = out
	}
	var receiptCount int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM practice_actions WHERE user_id=$1 AND operation='action'`, u).Scan(&receiptCount))
	require.Equal(t, 1, receiptCount)
	_, err = svc.Act(ctx, other, id, wrong, "foreign-action")
	require.ErrorIs(t, err, ErrNotFound)
	stale := wrong
	stale.ClientActionID = "stale"
	_, err = svc.Act(ctx, u, id, stale, "stale")
	require.ErrorIs(t, err, ErrConflict)
	require.False(t, session.Feedback.Correct)
	replayed, err := svc.Act(ctx, u, id, wrong, "wrong")
	require.NoError(t, err)
	require.Equal(t, session.Revision, replayed.Revision)
	list, err := svc.List(ctx, u)
	require.NoError(t, err)
	require.Equal(t, 1, list.AvailableMistakes)
	mistake, err := svc.Start(ctx, u, StartRequest{Mode: "mistakes"}, "mistake-old")
	require.NoError(t, err)
	require.Equal(t, 1, mistake.TotalSteps)
	// Create a newer error after the mistake session captured its exact source.
	newer, err := svc.Start(ctx, u, req, "newer-session")
	require.NoError(t, err)
	newID := uuid.MustParse(newer.ID)
	newState, err := repo.Get(ctx, u, newID)
	require.NoError(t, err)
	for newer.CurrentStep != nil {
		step := newState.Snapshot.Steps[newer.CompletedSteps]
		answer := step.Word.WordText
		if step.Word.MeaningID == first.Word.MeaningID {
			answer = "new wrong answer"
		}
		a := Action{StepID: step.Public.ID, ExpectedRevision: newer.Revision, ClientActionID: uuid.NewString(), Action: "answer", TypedAnswer: answer}
		newer, err = svc.Act(ctx, u, newID, a, a.ClientActionID)
		require.NoError(t, err)
		a = Action{StepID: step.Public.ID, ExpectedRevision: newer.Revision, ClientActionID: uuid.NewString(), Action: "continue"}
		newer, err = svc.Act(ctx, u, newID, a, a.ClientActionID)
		require.NoError(t, err)
	}
	require.Equal(t, "completed", newer.Status)
	mid := uuid.MustParse(mistake.ID)
	a := Action{StepID: mistake.CurrentStep.ID, ExpectedRevision: mistake.Revision, ClientActionID: "resolve-old", Action: "answer", TypedAnswer: first.Word.WordText}
	_, err = svc.Act(ctx, u, mid, a, "resolve-old")
	require.NoError(t, err)
	list, err = svc.List(ctx, u)
	require.NoError(t, err)
	require.Equal(t, 1, list.AvailableMistakes, "old session must not resolve newer error")
	latest, err := svc.Start(ctx, u, StartRequest{Mode: "mistakes"}, "mistake-latest")
	require.NoError(t, err)
	a = Action{StepID: latest.CurrentStep.ID, ClientActionID: "resolve-latest", Action: "answer", TypedAnswer: first.Word.WordText}
	_, err = svc.Act(ctx, u, uuid.MustParse(latest.ID), a, "resolve-latest")
	require.NoError(t, err)
	list, err = svc.List(ctx, u)
	require.NoError(t, err)
	require.Zero(t, list.AvailableMistakes)
	_, err = svc.Start(ctx, u, StartRequest{Mode: "mistakes"}, "empty")
	require.ErrorIs(t, err, ErrNoMistakes)
	again, err := svc.Start(ctx, u, req, "repeat-after-completion")
	require.NoError(t, err)
	require.NotEqual(t, newer.ID, again.ID)
	// Snapshot survives canonical edits; a failed receipt cannot advance progress.
	originalDefinition := first.Word.Definition
	exec(`UPDATE word_meanings SET short_definition='Changed after snapshot' WHERE id=$1`, first.Word.MeaningID)
	resumed, err := NewService(repo, nil).Get(ctx, u, id)
	exec(`UPDATE word_meanings SET short_definition=$2 WHERE id=$1`, first.Word.MeaningID, originalDefinition)
	require.NoError(t, err)
	require.Equal(t, session.CurrentStep.Prompt, resumed.CurrentStep.Prompt)
	exec(`ALTER TABLE practice_actions ADD CONSTRAINT practice_test_reject CHECK(idempotency_key <> 'force-rollback') NOT VALID`)
	a = Action{StepID: again.CurrentStep.ID, ClientActionID: "force-rollback", Action: "reveal"}
	_, err = svc.Act(ctx, u, uuid.MustParse(again.ID), a, "force-rollback")
	exec(`ALTER TABLE practice_actions DROP CONSTRAINT practice_test_reject`)
	require.Error(t, err)
	resumed, err = svc.Get(ctx, u, uuid.MustParse(again.ID))
	require.NoError(t, err)
	require.Zero(t, resumed.Revision)
	require.Nil(t, resumed.Feedback)
	foreign, err := svc.Start(ctx, other, StartRequest{Mode: "listening_choice", LessonKey: "airport"}, "other")
	require.NoError(t, err)
	require.NotEmpty(t, foreign.CurrentStep.SpeechText)
	for _, table := range []string{"user_words", "review_attempts", "daily_mission_snapshots", "confidence_point_ledger"} {
		var n int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE user_id=$1", u).Scan(&n))
		require.Zero(t, n)
	}
	ar := accounts.NewPostgreSQLRepository(db)
	data, err := ar.ExportPersonalData(ctx, u)
	require.NoError(t, err)
	require.Contains(t, string(data), "private wrong answer")
	require.NotContains(t, string(data), "CorrectChoice")
	require.NotContains(t, string(data), "same-start")
	require.NotContains(t, string(data), foreign.ID)
	exec(`UPDATE users SET status='deleted',deleted_at=CURRENT_TIMESTAMP WHERE id=$1`, u)
	_, err = svc.Act(ctx, u, id, wrong, "wrong")
	require.ErrorIs(t, err, ErrNotFound)
	counts, err := ar.AnonymizeUserData(ctx, u)
	require.NoError(t, err)
	require.Positive(t, counts.PracticeSessions)
	require.Positive(t, counts.PracticeActions)
	require.Equal(t, int64(2), counts.PracticeMistakeResolutions)
	for _, table := range []string{"practice_sessions", "practice_actions", "practice_mistake_resolutions"} {
		var n int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE user_id=$1", u).Scan(&n))
		require.Zero(t, n)
	}
	_, err = svc.Get(ctx, other, uuid.MustParse(foreign.ID))
	require.NoError(t, err)
}
func TestPracticePostgreSQLHistoricalMistakeSources(t *testing.T) {
	db := practiceDB(t)
	ctx := t.Context()
	u, other := uuid.New(), uuid.New()
	now := time.Now().UTC()
	exec := func(q string, args ...any) {
		t.Helper()
		_, e := db.ExecContext(ctx, q, args...)
		require.NoError(t, e)
	}
	for _, id := range []uuid.UUID{u, other} {
		exec(`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',$3,$3)`, id, id.String()+"@practice.invalid", now)
	}
	t.Cleanup(func() {
		for _, id := range []uuid.UUID{u, other} {
			for _, table := range []string{"practice_mistake_resolutions", "practice_actions", "practice_sessions", "lesson_actions", "lesson_sessions", "review_attempts", "user_words"} {
				_, _ = db.ExecContext(context.Background(), "DELETE FROM "+table+" WHERE user_id=$1", id)
			}
			_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE id=$1", id)
		}
	})
	sid := uuid.New()
	exec(`INSERT INTO lesson_sessions(id,user_id,lesson_key,lesson_version,snapshot,total_steps,created_at,updated_at)VALUES($1,$2,'airport','1','{}',9,$3,$3)`, sid, u, now)
	// These historical fixtures belong to Airport, independently of course order.
	boardingPass := "8a6ec801-3292-5fc2-bfb3-242ac405d891"
	feedback, _ := json.Marshal(map[string]any{"feedback": map[string]any{"correct": false, "correctChoiceId": boardingPass}})
	exec(`INSERT INTO lesson_actions(id,session_id,user_id,operation,idempotency_key,client_action_id,fingerprint,action,result,created_at)VALUES($1,$2,$3,'action','wrong','wrong',repeat('a',64),'{"action":"answer"}',$4,$5)`, uuid.New(), sid, u, string(feedback), now)
	for i, prompt := range []string{"multiple_choice", "self_check", "multiple_choice"} {
		user := u
		if i == 2 {
			user = other
		}
		word := uuid.New()
		meaning := []string{"6a61d337-80e7-5d50-bc61-8b45b6a054a1", "bee2408e-5991-5805-923c-6fba33ab7e7a", "329c4ec9-6562-568e-9e71-9134da2f0d8d"}[i]
		exec(`INSERT INTO user_words(id,user_id,meaning_id,status,source,added_at,created_at,updated_at)VALUES($1,$2,$3,'learning','manual',$4,$4,$4)`, word, user, meaning, now)
		exec(`INSERT INTO review_attempts(id,user_id,user_word_id,meaning_id,attempt_type,prompt_type,result,rating,review_step_before,review_step_after,answered_at,source,created_at,updated_at)VALUES($1,$2,$3,$4,'scheduled',$5,'incorrect','again',0,0,$6,'review',$6,$6)`, uuid.New(), user, word, meaning, prompt, now)
	}
	svc := NewService(NewPostgreSQLRepository(db), nil)
	list, err := svc.List(ctx, u)
	require.NoError(t, err)
	require.Equal(t, 2, list.AvailableMistakes, "only own objectively graded errors")
	session, err := svc.Start(ctx, u, StartRequest{Mode: "mistakes", LessonKey: "airport"}, "historical")
	require.NoError(t, err)
	require.Equal(t, 2, session.TotalSteps)
	require.Equal(t, "typed_recall", session.CurrentStep.Kind)
}
