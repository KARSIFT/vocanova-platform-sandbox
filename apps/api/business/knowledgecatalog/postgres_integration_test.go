package knowledgecatalog

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/wordknowledge"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestPostgreSQLDiscoveryAtomicChoicesIsolationHistoryAndReplay(t *testing.T) {
	db := isolatedCatalogDB(t)
	ctx := t.Context()
	now := time.Now().UTC()
	u, other, w, m := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	rank := 10000
	exec := func(q string, args ...any) {
		t.Helper()
		_, e := db.ExecContext(ctx, q, args...)
		require.NoError(t, e)
	}
	// Prefer seeded public content. Otherwise create one isolated public fixture.
	fixture := false
	err := db.QueryRowContext(ctx, `SELECT word_id,meaning_id FROM knowledge_catalog_entries ORDER BY rank LIMIT 1`).Scan(&w, &m)
	if err == sql.ErrNoRows {
		fixture = true
		exec(`INSERT INTO canonical_words(id,text,normalized_text,status,created_at,updated_at)VALUES($1,$2,$2,'active',$3,$3)`, w, "catalog-test-"+w.String(), now)
		exec(`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at)VALUES($1,$2,'noun','A synthetic test meaning',1,'active',$3,$3)`, m, w, now)
		exec(`INSERT INTO knowledge_catalog_entries(rank,word_id,meaning_id,source_rank,source_count,source_sense_key,source_name,catalog_version)VALUES($1,$2,$3,1,1,'synthetic-test','synthetic-test','english-10000-v1')`, rank, w, m)
	} else {
		require.NoError(t, err)
	}
	for _, id := range []uuid.UUID{u, other} {
		exec(`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',$3,$3)`, id, id.String()+"@catalog.invalid", now)
	}
	t.Cleanup(func() {
		for _, id := range []uuid.UUID{u, other} {
			for _, table := range []string{"word_knowledge_actions", "user_word_knowledge", "user_words", "users"} {
				q := "DELETE FROM " + table + " WHERE user_id=$1"
				if table == "users" {
					q = "DELETE FROM users WHERE id=$1"
				}
				db.ExecContext(context.Background(), q, id)
			}
		}
		if fixture {
			db.ExecContext(context.Background(), `DELETE FROM knowledge_catalog_entries WHERE meaning_id=$1`, m)
			db.ExecContext(context.Background(), `DELETE FROM word_meanings WHERE id=$1`, m)
			db.ExecContext(context.Background(), `DELETE FROM canonical_words WHERE id=$1`, w)
		}
	})
	repo := NewPostgreSQLRepository(db)
	svc := NewService(repo, nil)
	read := func(user uuid.UUID) Item {
		t.Helper()
		tx, e := db.BeginTx(ctx, nil)
		require.NoError(t, e)
		defer tx.Rollback()
		x, e := readItem(ctx, tx, user, m, time.Now().UTC())
		require.NoError(t, e)
		return x
	}
	fresh := read(u)
	require.False(t, fresh.SelfReportedKnown)
	require.False(t, fresh.Saved)
	assess := wordknowledge.NewService(wordknowledge.NewPostgreSQLRepository(db), nil)
	_, err = assess.Write(ctx, wordknowledge.WriteRequest{UserID: u, MeaningID: m, Note: "Keep my private note", IdempotencyKey: "note"})
	require.NoError(t, err)
	_, err = svc.Decide(ctx, DecisionRequest{UserID: u, MeaningID: m, Decision: "learn", ExpectedRevision: fresh.Revision, IdempotencyKey: "stale-note"})
	require.ErrorIs(t, err, ErrStale)
	req := DecisionRequest{UserID: u, MeaningID: m, Decision: "learn", ExpectedRevision: read(u).Revision, IdempotencyKey: "learn-one"}
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for range 8 {
		wg.Add(1)
		go func() { defer wg.Done(); _, e := svc.Decide(ctx, req); errs <- e }()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		require.NoError(t, e)
	}
	learned := read(u)
	require.True(t, learned.Saved)
	require.True(t, learned.Due)
	require.False(t, learned.SelfReportedKnown)
	require.False(t, read(other).Saved)
	var count int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM word_knowledge_actions WHERE user_id=$1 AND idempotency_key='learn-one'`, u).Scan(&count))
	require.Equal(t, 1, count)
	// A word with genuine past review history is paused without resetting history.
	exec(`UPDATE user_words SET status='reviewing',review_step=3,total_review_count=5,correct_review_count=4,consecutive_correct_count=2,next_review_at=$3,updated_at=$4 WHERE user_id=$1 AND meaning_id=$2`, u, m, now.Add(time.Hour), now.Add(time.Second))
	knownReq := DecisionRequest{UserID: u, MeaningID: m, Decision: "known", ExpectedRevision: read(u).Revision, IdempotencyKey: "known-one"}
	st, err := svc.Decide(ctx, knownReq)
	require.NoError(t, err)
	require.True(t, st.SelfReportedKnown)
	require.False(t, st.Saved)
	require.False(t, st.Due)
	var status, note string
	var reviews, step int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT uw.status,uw.total_review_count,uw.review_step,k.note FROM user_words uw JOIN user_word_knowledge k USING(user_id,meaning_id) WHERE uw.user_id=$1 AND uw.meaning_id=$2`, u, m).Scan(&status, &reviews, &step, &note))
	require.Equal(t, "ignored", status)
	require.Equal(t, 5, reviews)
	require.Equal(t, 3, step)
	require.Equal(t, "Keep my private note", note)
	// Same receipt returns current state, never re-applies an older learning choice.
	replay, err := svc.Decide(ctx, req)
	require.NoError(t, err)
	require.True(t, replay.SelfReportedKnown)
	require.False(t, replay.Saved)
	changed := req
	changed.Decision = "known"
	_, err = svc.Decide(ctx, changed)
	require.ErrorIs(t, err, ErrConflict)
	_, err = svc.Decide(ctx, DecisionRequest{UserID: u, MeaningID: m, Decision: "learn", ExpectedRevision: learned.Revision, IdempotencyKey: "stale"})
	require.ErrorIs(t, err, ErrStale)
	resumed, err := svc.Decide(ctx, DecisionRequest{UserID: u, MeaningID: m, Decision: "learn", ExpectedRevision: read(u).Revision, IdempotencyKey: "resume"})
	require.NoError(t, err)
	require.True(t, resumed.Saved)
	require.False(t, resumed.Due, "resuming preserves the existing future review schedule")
	require.False(t, resumed.SelfReportedKnown)
	require.NoError(t, db.QueryRowContext(ctx, `SELECT total_review_count,review_step FROM user_words WHERE user_id=$1 AND meaning_id=$2`, u, m).Scan(&reviews, &step))
	require.Equal(t, 5, reviews)
	require.Equal(t, 3, step)

	// A failing receipt insert rolls back the assessment AND saved-word pause.
	suffix := strings.ReplaceAll(uuid.NewString(), "-", "")
	fn, trigger := "catalog_test_reject_"+suffix, "catalog_test_receipt_"+suffix
	exec(`CREATE FUNCTION ` + fn + `() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.user_id = '` + u.String() + `'::uuid AND NEW.idempotency_key='force-rollback' THEN RAISE EXCEPTION 'synthetic receipt failure'; END IF; RETURN NEW; END $$`)
	exec(`CREATE TRIGGER ` + trigger + ` BEFORE INSERT ON word_knowledge_actions FOR EACH ROW EXECUTE FUNCTION ` + fn + `()`)
	t.Cleanup(func() {
		db.Exec(`DROP TRIGGER IF EXISTS ` + trigger + ` ON word_knowledge_actions`)
		db.Exec(`DROP FUNCTION IF EXISTS ` + fn + `()`)
	})
	beforeFailure := read(u)
	_, err = svc.Decide(ctx, DecisionRequest{UserID: u, MeaningID: m, Decision: "known", ExpectedRevision: beforeFailure.Revision, IdempotencyKey: "force-rollback"})
	require.Error(t, err)
	afterFailure := read(u)
	require.Equal(t, beforeFailure, afterFailure)
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM word_knowledge_actions WHERE user_id=$1 AND idempotency_key='force-rollback'`, u).Scan(&count))
	require.Zero(t, count)
	for _, table := range []string{"review_attempts", "confidence_point_ledger", "daily_mission_snapshots"} {
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE user_id=$1", u).Scan(&count))
		require.Zero(t, count)
	}
}
