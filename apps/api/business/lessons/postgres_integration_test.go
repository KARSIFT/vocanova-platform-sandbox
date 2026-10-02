package lessons

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func lessonDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	schema := "lesson_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = admin.ExecContext(t.Context(), "CREATE SCHEMA "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	scoped := dsn
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		u, e := url.Parse(dsn)
		require.NoError(t, e)
		q := u.Query()
		q.Set("search_path", schema)
		u.RawQuery = q.Encode()
		scoped = u.String()
	} else {
		scoped += " search_path=" + schema
	}
	db, err := sql.Open("postgres", scoped)
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
		_, _ = admin.ExecContext(context.Background(), "DROP SCHEMA "+pq.QuoteIdentifier(schema)+" CASCADE")
		admin.Close()
	})
	_, err = db.ExecContext(t.Context(), `CREATE TABLE users(id uuid PRIMARY KEY,status text NOT NULL DEFAULT 'active',deleted_at timestamptz)`)
	require.NoError(t, err)
	content, err := os.ReadFile("../../migrations/20260725100000_voc026_p1_content_tables.sql")
	require.NoError(t, err)
	parts := strings.SplitN(string(content), "CREATE TABLE user_words", 2)
	require.Len(t, parts, 2)
	_, err = db.ExecContext(t.Context(), parts[0])
	require.NoError(t, err)
	migration, err := os.ReadFile("../../migrations/20261002210000_guided_lessons.sql")
	require.NoError(t, err)
	_, err = db.ExecContext(t.Context(), string(migration))
	require.NoError(t, err)
	raw, err := os.ReadFile("../../cmd/seed/voc026-p1.json")
	require.NoError(t, err)
	var seed map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &seed))
	for _, table := range []string{"journey_situations", "canonical_words", "word_meanings", "word_examples", "usage_notes", "journey_words"} {
		// Populate exact canonical seed values; test clock supplies required audit timestamps.
		var rows []map[string]any
		require.NoError(t, json.Unmarshal(seed[table], &rows))
		for _, row := range rows {
			row["created_at"] = "2026-10-02T12:00:00Z"
			row["updated_at"] = "2026-10-02T12:00:00Z"
			b, e := json.Marshal(row)
			require.NoError(t, e)
			_, e = db.ExecContext(t.Context(), "INSERT INTO "+pq.QuoteIdentifier(table)+" SELECT * FROM jsonb_populate_record(NULL::"+pq.QuoteIdentifier(table)+",$1::jsonb)", string(b))
			require.NoError(t, e)
		}
	}
	return db
}
func TestPostgreSQLGuidedLessonsResumeReplayConcurrencyAndRollback(t *testing.T) {
	db := lessonDB(t)
	ctx := t.Context()
	u, other := uuid.New(), uuid.New()
	_, err := db.ExecContext(ctx, `INSERT INTO users(id) VALUES($1),($2)`, u, other)
	require.NoError(t, err)
	now := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	svc := NewService(NewPostgreSQLRepository(db), clock.Fixed{T: now})
	var wg sync.WaitGroup
	sessions := make([]*Session, 2)
	errs := make([]error, 2)
	for i := range sessions {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sessions[i], errs[i] = svc.Start(ctx, u, "airport", "start-"+string(rune('a'+i)))
		}(i)
	}
	wg.Wait()
	for _, e := range errs {
		require.NoError(t, e)
	}
	require.Equal(t, sessions[0].ID, sessions[1].ID)
	current := sessions[0]
	id := uuid.MustParse(current.ID)
	_, err = svc.Start(ctx, u, "restaurant", "start-a")
	require.ErrorIs(t, err, ErrConflict)
	_, err = svc.Get(ctx, other, id)
	require.ErrorIs(t, err, ErrNotFound)
	_, err = svc.Act(ctx, other, id, Action{StepID: "teach-1", Action: "continue", ClientActionID: "foreign"}, "foreign")
	require.ErrorIs(t, err, ErrNotFound)
	original := Action{StepID: "teach-1", ExpectedRevision: 0, Action: "continue", ClientActionID: "first"}
	for i := range sessions {
		wg.Add(1)
		go func(i int) { defer wg.Done(); sessions[i], errs[i] = svc.Act(ctx, u, id, original, "first") }(i)
	}
	wg.Wait()
	for _, e := range errs {
		require.NoError(t, e)
	}
	require.Equal(t, 1, sessions[0].Revision)
	require.Equal(t, 1, sessions[1].Revision)
	conflicting := original
	conflicting.ClientActionID = "changed"
	_, err = svc.Act(ctx, u, id, conflicting, "first")
	require.ErrorIs(t, err, ErrConflict)
	current, err = svc.Get(ctx, u, id)
	require.NoError(t, err)
	// A second service/process sees the saved step, not a browser-local draft.
	resumed, err := NewService(NewPostgreSQLRepository(db), nil).Get(ctx, u, id)
	require.NoError(t, err)
	require.Equal(t, current, resumed)
	// Force failure after state UPDATE but before receipt INSERT: both roll back.
	_, err = db.ExecContext(ctx, `CREATE FUNCTION reject_test_receipt() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.idempotency_key='force-rollback' THEN RAISE EXCEPTION 'forced test failure'; END IF; RETURN NEW; END $$; CREATE TRIGGER reject_test_receipt BEFORE INSERT ON lesson_actions FOR EACH ROW EXECUTE FUNCTION reject_test_receipt()`)
	require.NoError(t, err)
	next := Action{StepID: current.CurrentStep.ID, ExpectedRevision: current.Revision, Action: "continue", ClientActionID: "force-rollback"}
	_, err = svc.Act(ctx, u, id, next, "force-rollback")
	require.Error(t, err)
	unchanged, err := svc.Get(ctx, u, id)
	require.NoError(t, err)
	require.Equal(t, current, unchanged)
	_, err = db.ExecContext(ctx, `DROP TRIGGER reject_test_receipt ON lesson_actions`)
	require.NoError(t, err)
	// Snapshot survives canonical edits; a newly started lesson validates live refs.
	_, err = db.ExecContext(ctx, `UPDATE word_meanings SET short_definition='Changed after lesson start' WHERE id=$1`, current.Words[0].MeaningID)
	require.NoError(t, err)
	resumed, err = svc.Get(ctx, u, id)
	require.NoError(t, err)
	require.Equal(t, current.Words, resumed.Words)
	wrongOnce := false
	for current.CompletedSteps < 8 || !current.CanContinue {
		a := Action{StepID: current.CurrentStep.ID, ExpectedRevision: current.Revision, Action: "continue", ClientActionID: uuid.NewString()}
		if !current.CanContinue {
			a.Action = "answer"
			a.ChoiceID = current.CurrentStep.Word.MeaningID
			if !wrongOnce {
				for _, c := range current.CurrentStep.Choices {
					if c.ID != a.ChoiceID {
						a.ChoiceID = c.ID
						break
					}
				}
				wrongOnce = true
			}
		}
		current, err = svc.Act(ctx, u, id, a, a.ClientActionID)
		require.NoError(t, err)
	}
	final := Action{StepID: current.CurrentStep.ID, ExpectedRevision: current.Revision, Action: "continue", ClientActionID: "finish"}
	for i := range sessions {
		wg.Add(1)
		go func(i int) { defer wg.Done(); sessions[i], errs[i] = svc.Act(ctx, u, id, final, "finish") }(i)
	}
	wg.Wait()
	for _, e := range errs {
		require.NoError(t, e)
	}
	require.Equal(t, "completed", sessions[0].Status)
	require.Equal(t, sessions[0], sessions[1])
	require.Equal(t, 5, sessions[0].FirstAnswersCorrect)
	require.Equal(t, 6, sessions[0].QuestionsAnswered)
	replay, err := svc.Act(ctx, u, id, original, "first")
	require.NoError(t, err)
	require.Equal(t, sessions[0], replay, "old replay must return latest state, never rewind")
	again, err := svc.Start(ctx, u, "airport", "after-completion")
	require.NoError(t, err)
	require.Equal(t, sessions[0], again)
	var total int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM lesson_actions WHERE session_id=$1 AND client_action_id='finish'`, id).Scan(&total))
	require.Equal(t, 1, total)
	var invalid int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM lesson_actions WHERE idempotency_key='force-rollback'`).Scan(&invalid))
	require.Zero(t, invalid)
	all, err := svc.List(ctx, u)
	require.NoError(t, err)
	require.Len(t, all, 30)
	completed := 0
	for _, item := range all {
		if item.Status == "completed" {
			require.Equal(t, "airport", item.Key)
			completed++
		}
	}
	require.Equal(t, 1, completed)
	for _, d := range catalog {
		if d.Key == "airport" {
			continue
		}
		started, e := svc.Start(ctx, u, d.Key, uuid.NewString())
		require.NoError(t, e)
		require.Len(t, started.Words, 3)
		require.Equal(t, 9, started.TotalSteps)
	}
	_, err = db.ExecContext(ctx, `UPDATE users SET status='disabled' WHERE id=$1`, u)
	require.NoError(t, err)
	_, err = svc.Start(ctx, u, "airport", "disabled")
	require.ErrorIs(t, err, ErrNotFound)
}
