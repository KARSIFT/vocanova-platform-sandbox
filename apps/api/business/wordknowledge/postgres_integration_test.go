package wordknowledge_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/wordknowledge"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// Uses the actual, fully migrated disposable database. No schema substitutes.
func TestWordKnowledgePostgreSQLPrivacyReplayExportAndPurge(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	u, other, w, m, saved := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()
	exec := func(q string, args ...any) {
		t.Helper()
		_, e := db.ExecContext(t.Context(), q, args...)
		require.NoError(t, e)
	}
	for _, id := range []uuid.UUID{u, other} {
		exec(`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',$3,$3)`, id, id.String()+"@knowledge.invalid", now)
	}
	exec(`INSERT INTO canonical_words(id,text,normalized_text,status,created_at,updated_at)VALUES($1,$2,$2,'active',$3,$3)`, w, "knowledge-test-"+w.String(), now)
	exec(`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at)VALUES($1,$2,'noun','A synthetic test meaning',1,'active',$3,$3)`, m, w, now)
	exec(`INSERT INTO user_words(id,user_id,meaning_id,status,source,review_step,next_review_at,added_at,created_at,updated_at)VALUES($1,$2,$3,'reviewing','manual',3,$4,$5,$5,$5)`, saved, u, m, now.Add(time.Hour), now)
	t.Cleanup(func() {
		ctx := context.Background()
		for _, id := range []uuid.UUID{u, other} {
			for _, table := range []string{"word_knowledge_actions", "user_word_knowledge", "user_words"} {
				_, _ = db.ExecContext(ctx, "DELETE FROM "+table+" WHERE user_id=$1", id)
			}
			_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id=$1", id)
		}
		_, _ = db.ExecContext(ctx, "DELETE FROM word_meanings WHERE id=$1", m)
		_, _ = db.ExecContext(ctx, "DELETE FROM canonical_words WHERE id=$1", w)
	})
	var savedBefore string
	require.NoError(t, db.QueryRowContext(t.Context(), `SELECT row_to_json(uw)::text FROM user_words uw WHERE id=$1`, saved).Scan(&savedBefore))
	svc := wordknowledge.NewService(wordknowledge.NewPostgreSQLRepository(db), nil)
	req := wordknowledge.WriteRequest{UserID: u, MeaningID: m, SelfReportedKnown: true, Note: "Private owner note", IdempotencyKey: "one"}
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for range 8 {
		wg.Add(1)
		go func() { defer wg.Done(); _, e := svc.Write(t.Context(), req); errs <- e }()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		require.NoError(t, e)
	}
	var count int
	require.NoError(t, db.QueryRowContext(t.Context(), `SELECT count(*) FROM word_knowledge_actions WHERE user_id=$1`, u).Scan(&count))
	require.Equal(t, 1, count)
	st, err := svc.Get(t.Context(), other, m)
	require.NoError(t, err)
	require.Empty(t, st.Note)
	require.False(t, st.SelfReportedKnown)
	known, err := svc.KnownStates(t.Context(), other, []uuid.UUID{m})
	require.NoError(t, err)
	require.False(t, known[m])
	count, err = svc.CountKnown(t.Context(), u)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	foreign := req
	foreign.UserID = other
	foreign.Note = "Private other note"
	_, err = svc.Write(t.Context(), foreign)
	require.NoError(t, err)
	bad := req
	bad.Note = "new text"
	_, err = svc.Write(t.Context(), bad)
	require.ErrorIs(t, err, wordknowledge.ErrConflict)
	_, err = svc.Write(t.Context(), wordknowledge.WriteRequest{UserID: other, MeaningID: m, Delete: true, IdempotencyKey: "clear"})
	require.NoError(t, err)
	st, err = svc.Get(t.Context(), u, m)
	require.NoError(t, err)
	require.Equal(t, req.Note, st.Note)
	_, err = svc.Write(t.Context(), wordknowledge.WriteRequest{UserID: u, MeaningID: m, Delete: true, IdempotencyKey: "clear"})
	require.NoError(t, err)
	st, err = svc.Write(t.Context(), req)
	require.NoError(t, err)
	require.Empty(t, st.Note)
	require.False(t, st.SelfReportedKnown)
	// A receipt failure must roll back the state write in the same transaction.
	exec(`ALTER TABLE word_knowledge_actions ADD CONSTRAINT knowledge_test_reject CHECK(idempotency_key <> 'force-rollback') NOT VALID`)
	rollback := req
	rollback.IdempotencyKey = "force-rollback"
	_, err = svc.Write(t.Context(), rollback)
	exec(`ALTER TABLE word_knowledge_actions DROP CONSTRAINT knowledge_test_reject`)
	require.Error(t, err)
	st, err = svc.Get(t.Context(), u, m)
	require.NoError(t, err)
	require.Empty(t, st.Note)
	req.IdempotencyKey = "restore"
	_, err = svc.Write(t.Context(), req)
	require.NoError(t, err)
	foreign.IdempotencyKey = "restore"
	_, err = svc.Write(t.Context(), foreign)
	require.NoError(t, err)
	var savedAfter string
	require.NoError(t, db.QueryRowContext(t.Context(), `SELECT row_to_json(uw)::text FROM user_words uw WHERE id=$1`, saved).Scan(&savedAfter))
	require.Equal(t, savedBefore, savedAfter)
	for _, table := range []string{"review_attempts", "daily_mission_snapshots", "confidence_point_ledger"} {
		require.NoError(t, db.QueryRowContext(t.Context(), "SELECT count(*) FROM "+table+" WHERE user_id=$1", u).Scan(&count))
		require.Zero(t, count)
	}
	ar := accounts.NewPostgreSQLRepository(db)
	raw, err := ar.ExportPersonalData(t.Context(), u)
	require.NoError(t, err)
	require.Contains(t, string(raw), "Private owner note")
	require.NotContains(t, string(raw), "Private other note")
	require.NotContains(t, string(raw), "fingerprint")
	var exported map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &exported))
	var entries []map[string]any
	require.NoError(t, json.Unmarshal(exported["wordKnowledge"], &entries))
	require.Len(t, entries, 1)
	require.Equal(t, true, entries[0]["selfReportedKnown"])
	exec(`UPDATE users SET status='deleted',deleted_at=CURRENT_TIMESTAMP WHERE id=$1`, u)
	_, err = svc.Write(t.Context(), req)
	require.ErrorIs(t, err, wordknowledge.ErrNotFound)
	_, err = svc.Get(t.Context(), u, m)
	require.ErrorIs(t, err, wordknowledge.ErrNotFound)
	counters, err := ar.AnonymizeUserData(t.Context(), u)
	require.NoError(t, err)
	require.Equal(t, int64(1), counters.WordKnowledge)
	require.Equal(t, int64(3), counters.WordKnowledgeActions)
	for _, table := range []string{"user_word_knowledge", "word_knowledge_actions"} {
		require.NoError(t, db.QueryRowContext(t.Context(), "SELECT count(*) FROM "+table+" WHERE user_id=$1", u).Scan(&count))
		require.Zero(t, count)
	}
	st, err = svc.Get(t.Context(), other, m)
	require.NoError(t, err)
	require.Equal(t, "Private other note", st.Note)
	count, err = svc.CountKnown(t.Context(), other)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestWordKnowledgePostgreSQLPartialAssessmentIsAtomic(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	u, w, m := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()
	exec := func(q string, args ...any) {
		t.Helper()
		_, e := db.ExecContext(t.Context(), q, args...)
		require.NoError(t, e)
	}
	exec(`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',$3,$3)`, u, u.String()+"@assessment.invalid", now)
	exec(`INSERT INTO canonical_words(id,text,normalized_text,status,created_at,updated_at)VALUES($1,$2,$2,'active',$3,$3)`, w, "assessment-"+w.String(), now)
	exec(`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at)VALUES($1,$2,'noun','A synthetic test meaning',1,'active',$3,$3)`, m, w, now)
	t.Cleanup(func() {
		ctx := context.Background()
		for _, table := range []string{"word_knowledge_actions", "user_word_knowledge"} {
			_, _ = db.ExecContext(ctx, "DELETE FROM "+table+" WHERE user_id=$1", u)
		}
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id=$1", u)
		_, _ = db.ExecContext(ctx, "DELETE FROM word_meanings WHERE id=$1", m)
		_, _ = db.ExecContext(ctx, "DELETE FROM canonical_words WHERE id=$1", w)
	})
	svc := wordknowledge.NewService(wordknowledge.NewPostgreSQLRepository(db), nil)
	_, err = svc.Write(t.Context(), wordknowledge.WriteRequest{UserID: u, MeaningID: m, Note: "Preserve my latest note", IdempotencyKey: "note"})
	require.NoError(t, err)
	req := wordknowledge.WriteRequest{UserID: u, MeaningID: m, SelfReportedKnown: true, AssessmentOnly: true, IdempotencyKey: "assessment"}
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for range 8 {
		wg.Add(1)
		go func() { defer wg.Done(); _, e := svc.Write(t.Context(), req); errs <- e }()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		require.NoError(t, e)
	}
	st, err := svc.Get(t.Context(), u, m)
	require.NoError(t, err)
	require.True(t, st.SelfReportedKnown)
	require.Equal(t, "Preserve my latest note", st.Note)
	var count int
	require.NoError(t, db.QueryRowContext(t.Context(), `SELECT count(*) FROM word_knowledge_actions WHERE user_id=$1 AND idempotency_key='assessment'`, u).Scan(&count))
	require.Equal(t, 1, count)
	newer := wordknowledge.WriteRequest{UserID: u, MeaningID: m, SelfReportedKnown: false, Note: "Edited after assessment", IdempotencyKey: "newer"}
	_, err = svc.Write(t.Context(), newer)
	require.NoError(t, err)
	st, err = svc.Write(t.Context(), req)
	require.NoError(t, err)
	require.False(t, st.SelfReportedKnown)
	require.Equal(t, newer.Note, st.Note)
	// A failed receipt rolls back even the partial state change.
	exec(`ALTER TABLE word_knowledge_actions ADD CONSTRAINT assessment_test_reject CHECK(idempotency_key <> 'assessment-rollback') NOT VALID`)
	rollback := req
	rollback.IdempotencyKey = "assessment-rollback"
	_, err = svc.Write(t.Context(), rollback)
	exec(`ALTER TABLE word_knowledge_actions DROP CONSTRAINT assessment_test_reject`)
	require.Error(t, err)
	st, err = svc.Get(t.Context(), u, m)
	require.NoError(t, err)
	require.False(t, st.SelfReportedKnown)
	require.Equal(t, newer.Note, st.Note)
	// The same assessment may be changed explicitly without touching the note.
	req.IdempotencyKey = "assessment-again"
	_, err = svc.Write(t.Context(), req)
	require.NoError(t, err)
	req.IdempotencyKey = "unmark"
	req.SelfReportedKnown = false
	st, err = svc.Write(t.Context(), req)
	require.NoError(t, err)
	require.False(t, st.SelfReportedKnown)
	require.Equal(t, newer.Note, st.Note)
	for _, table := range []string{"user_words", "review_attempts", "daily_mission_snapshots", "confidence_point_ledger"} {
		require.NoError(t, db.QueryRowContext(t.Context(), "SELECT count(*) FROM "+table+" WHERE user_id=$1", u).Scan(&count))
		require.Zero(t, count)
	}
	// Concurrent full-note replacement and assessment preserve the replacement
	// note in either serial order. The shared user lock protects both operations.
	full := wordknowledge.WriteRequest{UserID: u, MeaningID: m, Note: "Concurrent full update", SelfReportedKnown: true, IdempotencyKey: "full-concurrent"}
	partial := wordknowledge.WriteRequest{UserID: u, MeaningID: m, SelfReportedKnown: false, AssessmentOnly: true, IdempotencyKey: "partial-concurrent"}
	concurrentErrors := make(chan error, 2)
	for _, write := range []wordknowledge.WriteRequest{full, partial} {
		go func() { _, e := svc.Write(t.Context(), write); concurrentErrors <- e }()
	}
	for range 2 {
		require.NoError(t, <-concurrentErrors)
	}
	st, err = svc.Get(t.Context(), u, m)
	require.NoError(t, err)
	require.Equal(t, full.Note, st.Note)
	knownAfterBoth := st.SelfReportedKnown
	for _, write := range []wordknowledge.WriteRequest{partial, full} {
		st, err = svc.Write(t.Context(), write)
		require.NoError(t, err)
		require.Equal(t, full.Note, st.Note)
		require.Equal(t, knownAfterBoth, st.SelfReportedKnown)
	}
	partial.SelfReportedKnown = true
	_, err = svc.Write(t.Context(), partial)
	require.ErrorIs(t, err, wordknowledge.ErrConflict)

	// Hold the deactivation transaction until a separate writer is demonstrably
	// waiting on its row lock. After commit it must reject the now-inactive user.
	writerDB, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer writerDB.Close()
	writerDB.SetMaxOpenConns(1)
	var writerPID int
	require.NoError(t, writerDB.QueryRowContext(t.Context(), "SELECT pg_backend_pid()").Scan(&writerPID))
	tx, err := db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	defer tx.Rollback()
	_, err = tx.ExecContext(t.Context(), `UPDATE users SET status='deleted',deleted_at=CURRENT_TIMESTAMP WHERE id=$1`, u)
	require.NoError(t, err)
	writer := wordknowledge.NewService(wordknowledge.NewPostgreSQLRepository(writerDB), nil)
	partial.IdempotencyKey = "after-deactivation"
	writeResult := make(chan error, 1)
	go func() { _, e := writer.Write(t.Context(), partial); writeResult <- e }()
	require.Eventually(t, func() bool {
		var blocked bool
		e := db.QueryRowContext(t.Context(), `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE pid=$1 AND wait_event_type='Lock')`, writerPID).Scan(&blocked)
		return e == nil && blocked
	}, 5*time.Second, 10*time.Millisecond)
	require.NoError(t, tx.Commit())
	select {
	case err = <-writeResult:
		require.ErrorIs(t, err, wordknowledge.ErrNotFound)
	case <-time.After(5 * time.Second):
		t.Fatal("assessment did not finish after deactivation committed")
	}
	require.NoError(t, db.QueryRowContext(t.Context(), `SELECT count(*) FROM word_knowledge_actions WHERE user_id=$1 AND idempotency_key='after-deactivation'`, u).Scan(&count))
	require.Zero(t, count)
}
