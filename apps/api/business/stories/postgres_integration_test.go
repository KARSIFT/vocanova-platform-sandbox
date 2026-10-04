package stories

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"os"
	"sync"
	"testing"
	"time"
)

func TestStoriesPostgreSQLOwnershipConcurrentReplayResumeHistoryAndActiveUser(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	ctx := t.Context()
	owner, other := uuid.New(), uuid.New()
	exec := func(q string, args ...any) {
		t.Helper()
		_, e := db.ExecContext(ctx, q, args...)
		require.NoError(t, e)
	}
	for _, u := range []uuid.UUID{owner, other} {
		exec("INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)", u, u.String()+"@stories.invalid")
	}
	t.Cleanup(func() {
		for _, u := range []uuid.UUID{owner, other} {
			_, _ = db.ExecContext(context.Background(), "DELETE FROM story_actions WHERE user_id=$1", u)
			_, _ = db.ExecContext(context.Background(), "DELETE FROM story_sessions WHERE user_id=$1", u)
			_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE id=$1", u)
		}
	})
	repo := NewPostgreSQLRepository(db)
	svc := NewService(repo, nil)
	var wg sync.WaitGroup
	results := make(chan *StorySession, 8)
	errs := make(chan error, 8)
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s, e := svc.Start(ctx, owner, StoryStartRequest{StoryKey: "a-quiet-lunch"}, "same-start")
			results <- s
			errs <- e
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	for e := range errs {
		require.NoError(t, e)
	}
	var session *StorySession
	for s := range results {
		if session != nil {
			require.Equal(t, session.ID, s.ID)
		}
		session = s
	}
	id := uuid.MustParse(session.ID)
	t.Run("retired catalog preserves exact start replay", func(t *testing.T) {
		saved := catalog
		catalog = catalog[1:]
		defer func() { catalog = saved }()
		replayed, e := svc.Start(ctx, owner, StoryStartRequest{StoryKey: "a-quiet-lunch"}, "same-start")
		require.NoError(t, e)
		require.Equal(t, id.String(), replayed.ID)
		require.Equal(t, "a-quiet-lunch", replayed.StoryKey)
		_, e = svc.Start(ctx, owner, StoryStartRequest{StoryKey: "a-quiet-lunch"}, "retired-fresh")
		require.ErrorIs(t, e, ErrNotFound)
		_, e = svc.Start(ctx, owner, StoryStartRequest{StoryKey: "the-right-platform"}, "same-start")
		require.ErrorIs(t, e, ErrConflict)
	})

	require.Zero(t, session.Revision)
	_, err = svc.Get(ctx, other, id)
	require.ErrorIs(t, err, ErrNotFound)
	_, err = svc.Act(ctx, other, id, StoryAction{StepID: session.CurrentStep.ID, Action: "continue", ClientActionID: "other"}, "other")
	require.ErrorIs(t, err, ErrNotFound)
	_, err = svc.Start(ctx, owner, StoryStartRequest{StoryKey: "the-right-platform"}, "same-start")
	require.ErrorIs(t, err, ErrConflict)
	action := StoryAction{StepID: session.CurrentStep.ID, Action: "continue", ClientActionID: "once"}
	actionResults := make(chan *StorySession, 8)
	actionErrors := make(chan error, 8)
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s, e := svc.Act(ctx, owner, id, action, "same-action")
			actionResults <- s
			actionErrors <- e
		}()
	}
	wg.Wait()
	close(actionResults)
	close(actionErrors)
	for e := range actionErrors {
		require.NoError(t, e)
	}
	for s := range actionResults {
		require.Equal(t, 1, s.Revision)
		session = s
	}
	changed := action
	changed.Action = "answer"
	changed.ChoiceID = "choice-1"
	_, err = svc.Act(ctx, owner, id, changed, "same-action")
	require.ErrorIs(t, err, ErrConflict)
	changed = action
	changed.ClientActionID = "other-key"
	_, err = svc.Act(ctx, owner, id, changed, "different-key")
	require.ErrorIs(t, err, ErrConflict)
	_, err = svc.Act(ctx, owner, id, action, "different-key")
	require.ErrorIs(t, err, ErrConflict)
	// A fresh stale action cannot overwrite the committed line progression.
	changed = action
	changed.ClientActionID = "stale"
	_, err = svc.Act(ctx, owner, id, changed, "stale")
	require.ErrorIs(t, err, ErrConflict)
	// Freeze a historical snapshot wording, then prove GET and resume use it.
	st, err := repo.Get(ctx, owner, id)
	require.NoError(t, err)
	st.Snapshot.Steps[0].Public.Line.Text = "This line belongs to the original saved edition."
	raw, err := json.Marshal(st.Snapshot)
	require.NoError(t, err)
	exec("UPDATE story_sessions SET snapshot=$2 WHERE id=$1", id, string(raw))
	session, err = svc.Get(ctx, owner, id)
	require.NoError(t, err)
	require.Equal(t, "This line belongs to the original saved edition.", session.VisibleLines[0].Text)
	// Complete an attempt: deliberately miss the first question then retry.
	failed := false
	for session.Status != "completed" {
		st, err = repo.Get(ctx, owner, id)
		require.NoError(t, err)
		step := st.Snapshot.Steps[st.Index]
		a := StoryAction{StepID: step.Public.ID, ExpectedRevision: session.Revision, ClientActionID: uuid.NewString(), Action: "continue"}
		if step.Public.Kind != "line" && !session.CanContinue {
			a.Action = "answer"
			a.ChoiceID = step.CorrectChoice
			if !failed {
				for _, c := range step.Public.Choices {
					if c.ID != step.CorrectChoice {
						a.ChoiceID = c.ID
						break
					}
				}
				failed = true
			}
		}
		session, err = svc.Act(ctx, owner, id, a, a.ClientActionID)
		require.NoError(t, err)
		// Reconstruct the repository and service, proving progress is in PostgreSQL.
		session, err = NewService(NewPostgreSQLRepository(db), nil).Get(ctx, owner, id)
		require.NoError(t, err)
	}
	require.Equal(t, 1, session.FirstAnswersCorrect)
	require.Equal(t, 2, session.QuestionsAnswered)
	require.Len(t, session.VisibleLines, 8)
	completed, err := svc.Get(ctx, owner, id)
	require.NoError(t, err)
	require.Equal(t, "completed", completed.Status)
	// Replaying an old acknowledged action returns current authoritative progress.
	replayed, err := svc.Act(ctx, owner, id, action, "same-action")
	require.NoError(t, err)
	require.Equal(t, session.Revision, replayed.Revision)
	fresh, err := svc.Start(ctx, owner, StoryStartRequest{StoryKey: "a-quiet-lunch"}, "fresh")
	require.NoError(t, err)
	require.NotEqual(t, id.String(), fresh.ID)
	require.Equal(t, ContentVersion, fresh.ContentVersion)
	require.NotEqual(t, completed.VisibleLines[0].Text, fresh.VisibleLines[0].Text)
	library, err := svc.List(ctx, owner)
	require.NoError(t, err)
	require.Equal(t, fresh.ID, library.Items[0].LatestSession.ID)
	var count int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM story_sessions WHERE user_id=$1", owner).Scan(&count))
	require.Equal(t, 2, count)
	exec("UPDATE users SET status='disabled' WHERE id=$1", owner)
	_, err = svc.Get(ctx, owner, id)
	require.ErrorIs(t, err, ErrNotFound)
	_, err = svc.Read(ctx, owner, "a-quiet-lunch")
	require.ErrorIs(t, err, ErrNotFound)
	_, err = svc.List(ctx, owner)
	require.ErrorIs(t, err, ErrNotFound)
	_, err = svc.Start(ctx, owner, StoryStartRequest{StoryKey: "a-quiet-lunch"}, "suspended")
	require.ErrorIs(t, err, ErrNotFound)
	_, err = svc.Act(ctx, owner, id, action, "same-action")
	require.ErrorIs(t, err, ErrNotFound)
	exec("UPDATE users SET status='active',deleted_at=$2 WHERE id=$1", owner, time.Now())
	_, err = svc.Get(ctx, owner, id)
	require.ErrorIs(t, err, ErrNotFound)
	foreign, err := svc.Start(ctx, other, StoryStartRequest{StoryKey: "a-room-for-tonight"}, "foreign-private-key")
	require.NoError(t, err)
	ar := accounts.NewPostgreSQLRepository(db)
	data, err := ar.ExportPersonalData(ctx, owner)
	require.NoError(t, err)
	require.Contains(t, string(data), "\"storySessions\"")
	require.Contains(t, string(data), "\"storyKey\": \"a-quiet-lunch\"")
	require.Contains(t, string(data), "\"correct\": false")
	for _, private := range []string{"Snapshot", "CorrectChoice", "same-start", "same-action", "clientActionId", "idempotency_key", "fingerprint", foreign.ID, "foreign-private-key"} {
		require.NotContains(t, string(data), private)
	}
	counts, err := ar.AnonymizeUserData(ctx, owner)
	require.NoError(t, err)
	require.Equal(t, int64(2), counts.StorySessions)
	require.Positive(t, counts.StoryActions)
	for _, table := range []string{"story_sessions", "story_actions"} {
		var n int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE user_id=$1", owner).Scan(&n))
		require.Zero(t, n)
	}
	_, err = svc.Get(ctx, other, uuid.MustParse(foreign.ID))
	require.NoError(t, err)

}
