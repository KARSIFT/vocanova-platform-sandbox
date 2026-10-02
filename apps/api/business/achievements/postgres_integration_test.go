package achievements

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestAchievementsPostgreSQLQualifiedEventsThresholdsAndUnsave(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	ctx := t.Context()
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	u, other, w, m, saved := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	exec := func(q string, args ...any) {
		t.Helper()
		_, e := db.ExecContext(ctx, q, args...)
		require.NoError(t, e)
	}
	for _, id := range []uuid.UUID{u, other} {
		exec(`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',$3,$3)`, id, id.String()+"@achievements.invalid", base)
	}
	exec(`INSERT INTO canonical_words(id,text,normalized_text,status,created_at,updated_at)VALUES($1,$2,$2,'active',$3,$3)`, w, "achievement-"+w.String(), base)
	exec(`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at)VALUES($1,$2,'noun','Synthetic meaning',1,'active',$3,$3)`, m, w, base)
	exec(`INSERT INTO user_words(id,user_id,meaning_id,status,source,added_at,created_at,updated_at)VALUES($1,$2,$3,'learning','manual',$4,$4,$4)`, saved, u, m, base)
	t.Cleanup(func() {
		c := context.Background()
		for _, id := range []uuid.UUID{u, other} {
			_, _ = db.ExecContext(c, `DELETE FROM ai_feedback_attempts WHERE learner_sentence_id IN(SELECT id FROM learner_sentences WHERE user_id=$1)`, id)
			for _, table := range []string{"learner_sentences", "practice_mistake_resolutions", "practice_actions", "practice_sessions", "lesson_actions", "lesson_sessions", "review_attempts", "user_words"} {
				_, _ = db.ExecContext(c, "DELETE FROM "+table+" WHERE user_id=$1", id)
			}
			_, _ = db.ExecContext(c, `DELETE FROM users WHERE id=$1`, id)
		}
		_, _ = db.ExecContext(c, `DELETE FROM word_meanings WHERE id=$1`, m)
		_, _ = db.ExecContext(c, `DELETE FROM canonical_words WHERE id=$1`, w)
	})
	lesson := func(user uuid.UUID, key string, at time.Time, completed, receipt bool) {
		id := uuid.New()
		status := "in_progress"
		current := 0
		var done any
		if completed {
			status = "completed"
			current = 9
			done = at
		}
		exec(`INSERT INTO lesson_sessions(id,user_id,lesson_key,lesson_version,snapshot,current_step,total_steps,status,completed_at,created_at,updated_at)VALUES($1,$2,$3,'1','{}',$4,9,$5,$6,$7,$7)`, id, user, key, current, status, done, at)
		if receipt {
			for n := 0; n < 2; n++ {
				exec(`INSERT INTO lesson_actions(id,session_id,user_id,operation,idempotency_key,client_action_id,fingerprint,action,result,created_at)VALUES($1,$2,$3,'action',$4,$4,repeat('a',64),'{"action":"continue"}','{"status":"completed"}',$5)`, uuid.New(), id, user, uuid.NewString(), at.Add(time.Duration(n)*time.Second))
			}
		}
	}
	for i, key := range []string{"airport", "restaurant", "hotel-check-in", "job-interview", "daily-conversation", "work-meeting", "university-class"} {
		lesson(u, key, base.Add(time.Duration(i)*time.Minute), true, true)
	}
	lesson(u, "unconfirmed", base.Add(-time.Hour), true, false)
	lesson(u, "unfinished", base.Add(-time.Hour), false, true)
	lesson(other, "foreign", base.Add(-time.Hour), true, true)
	practice := func(user uuid.UUID, mode string, at time.Time, first int, completed, receipt bool) {
		id := uuid.New()
		status := "in_progress"
		current := 0
		var done any
		if completed {
			status = "completed"
			current = 3
			done = at
		}
		exec(`INSERT INTO practice_sessions(id,user_id,mode,snapshot,current_step,total_steps,first_answers_correct,questions_answered,status,completed_at,created_at,updated_at)VALUES($1,$2,$3,'{}',$4,3,$5,3,$6,$7,$8,$8)`, id, user, mode, current, first, status, done, at)
		if receipt {
			exec(`INSERT INTO practice_actions(id,session_id,user_id,operation,idempotency_key,client_action_id,fingerprint,action,result,created_at)VALUES($1,$2,$3,'action',$4,$4,repeat('a',64),'{"action":"continue"}','{"status":"completed"}',$5)`, uuid.New(), id, user, uuid.NewString(), at)
		}
	}
	for i := 0; i < 6; i++ {
		mode := "typed_recall"
		correct := 0
		if i == 2 {
			mode = "listening_choice"
			correct = 3
		}
		if i == 5 {
			correct = 3
		}
		practice(u, mode, base.Add(time.Duration(20+i)*time.Minute), correct, true, true)
	}
	practice(u, "typed_recall", base.Add(-time.Hour), 3, true, false)
	practice(u, "typed_recall", base.Add(-time.Hour), 3, false, true)
	practice(other, "typed_recall", base.Add(-time.Hour), 3, true, true)
	for i := 0; i < 52; i++ {
		result := "incorrect"
		var rating any = "again"
		attempt := "review"
		if i%2 == 0 {
			result = "correct"
			rating = "good"
		}
		if i == 50 {
			result = "skipped"
			rating = nil
		}
		if i == 51 {
			attempt = "unrelated"
		}
		exec(`INSERT INTO review_attempts(id,user_id,user_word_id,meaning_id,attempt_type,prompt_type,result,rating,review_step_before,review_step_after,answered_at,source,created_at,updated_at)VALUES($1,$2,$3,$4,$5,'self_check',$6,$7,0,0,$8,'review',$9,$9)`, uuid.New(), u, saved, m, attempt, result, rating, base.AddDate(20, 0, 0), base.Add(time.Duration(100+i)*time.Minute))
	}
	writing := func(user uuid.UUID, at time.Time, status, verdict string, deleted bool) {
		id, aid := uuid.New(), uuid.New()
		var completed, feedback, errorCode, deletedAt any
		sentenceStatus := "submitted"
		if status == "succeeded" {
			completed = at
			feedback = fmt.Sprintf(`{"status":%q}`, verdict)
			sentenceStatus = "feedback_ready"
		}
		if status == "failed" {
			errorCode = "synthetic_failure"
			sentenceStatus = "feedback_failed"
		}
		if deleted {
			deletedAt = at
		}
		exec(`INSERT INTO learner_sentences(id,user_id,sentence_text,normalized_sentence_text,source,status,submitted_at,deleted_at,created_at,updated_at)VALUES($1,$2,'Synthetic sentence.','synthetic sentence.','free_practice',$3,$4,$5,$4,$4)`, id, user, sentenceStatus, at, deletedAt)
		exec(`INSERT INTO ai_feedback_attempts(id,learner_sentence_id,status,provider,model,prompt_version,request_hash,feedback_json,error_code,completed_at,created_at,updated_at)VALUES($1,$2,$3,'synthetic','synthetic','v1',$4,$5,$6,$7,$8,$8)`, aid, id, status, aid.String(), feedback, errorCode, completed, at)
	}
	writing(u, base.Add(-time.Hour), "succeeded", "correct", true)
	writing(u, base.Add(-time.Hour), "failed", "", false)
	writing(u, base.Add(-time.Hour), "pending", "", false)
	writing(u, base.Add(-time.Hour), "succeeded", "not_a_feedback_verdict", false)
	writing(u, base.Add(200*time.Minute), "succeeded", "incorrect", false)
	writing(other, base.Add(-time.Hour), "succeeded", "correct", false)
	svc := NewService(NewPostgreSQLRepository(db))
	// Run the actual projection on a connection that refuses data writes.
	db.SetMaxOpenConns(1)
	exec(`SET default_transaction_read_only=on`)
	out, err := svc.List(ctx, u)
	exec(`SET default_transaction_read_only=off`)
	require.NoError(t, err)
	require.Len(t, out.Items, 8)
	expected := map[string]time.Time{"guided-first": base, "guided-three": base.Add(2 * time.Minute), "guided-seven": base.Add(6 * time.Minute), "recall-independent-first": base.Add(25 * time.Minute), "practice-five": base.Add(24 * time.Minute), "reviews-ten": base.Add(109 * time.Minute), "reviews-fifty": base.Add(149 * time.Minute), "writing-first": base.Add(200 * time.Minute)}
	for _, a := range out.Items {
		require.True(t, a.Earned, a.ID)
		require.Equal(t, a.Target, a.Current, a.ID)
		require.Equal(t, expected[a.ID], *a.EarnedAt, a.ID)
	}
	before, err := json.Marshal(out)
	require.NoError(t, err)
	exec(`UPDATE user_words SET deleted_at=CURRENT_TIMESTAMP WHERE id=$1`, saved)
	exec(`UPDATE learner_sentences SET status='archived' WHERE user_id=$1 AND deleted_at IS NULL AND status='feedback_ready'`, u)
	after, err := svc.List(ctx, u)
	require.NoError(t, err)
	raw, err := json.Marshal(after)
	require.NoError(t, err)
	require.JSONEq(t, string(before), string(raw), "unsaving cannot revoke history badges")
	metrics, err := NewPostgreSQLRepository(db).Metrics(ctx, u)
	require.NoError(t, err)
	require.Equal(t, 7, metrics["lessons"].Count)
	require.Equal(t, 6, metrics["practice"].Count)
	require.Equal(t, 1, metrics["independent"].Count)
	require.Equal(t, 50, metrics["reviews"].Count)
	require.Equal(t, 1, metrics["writing"].Count)
	otherResult, err := svc.List(ctx, other)
	require.NoError(t, err)
	for _, a := range otherResult.Items {
		if a.ID == "reviews-ten" || a.ID == "reviews-fifty" || a.ID == "guided-three" {
			require.False(t, a.Earned)
		}
	}
	for _, table := range []string{"daily_mission_snapshots", "confidence_point_ledger"} {
		var n int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE user_id=$1", u).Scan(&n))
		require.Zero(t, n, "GET must not create rewards")
	}
	exec(`UPDATE users SET status='deleted',deleted_at=CURRENT_TIMESTAMP WHERE id=$1`, u)
	_, err = svc.List(ctx, u)
	require.ErrorIs(t, err, ErrNotFound)
}
