package wordlists

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/practice"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPostgreSQLListsOwnershipReplayRevisionAndFrozenPractice(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	db, e := sql.Open("postgres", dsn)
	require.NoError(t, e)
	t.Cleanup(func() { db.Close() })
	ctx := t.Context()
	u, other, id := uuid.New(), uuid.New(), uuid.New()
	meaning := uuid.MustParse("ac93d068-7c3e-55c9-9d85-3fd3518592dc")
	exec := func(q string, args ...any) { _, e := db.ExecContext(ctx, q, args...); require.NoError(t, e) }
	for _, user := range []uuid.UUID{u, other} {
		exec(`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, user, user.String()+"@lists.invalid")
	}
	t.Cleanup(func() {
		for _, user := range []uuid.UUID{u, other} {
			for _, table := range []string{"practice_mistake_resolutions", "practice_actions", "practice_sessions", "word_list_actions", "user_word_lists", "users"} {
				_, e := db.ExecContext(context.Background(), "DELETE FROM "+table+" WHERE "+map[bool]string{true: "id", false: "user_id"}[table == "users"]+"=$1", user)
				require.NoError(t, e)
			}
		}
	})
	svc := NewService(NewPostgreSQLRepository(db), nil)
	req := WriteRequest{UserID: u, ListID: id, Operation: "put", Name: "  Travel words  ", IdempotencyKey: "create"}
	st, e := svc.Write(ctx, req)
	require.NoError(t, e)
	require.Equal(t, "Travel words", st.Name)
	require.Equal(t, 1, st.Revision)
	require.Empty(t, st.Members)
	_, e = svc.Get(ctx, other, id)
	require.ErrorIs(t, e, ErrNotFound)
	attack := req
	attack.UserID = other
	_, e = svc.Write(ctx, attack)
	require.ErrorIs(t, e, ErrConflict)
	change := req
	change.ExpectedRevision = 1
	change.Name = "Renamed"
	change.IdempotencyKey = "rename"
	st, e = svc.Write(ctx, change)
	require.NoError(t, e)
	require.Equal(t, 2, st.Revision)
	st, e = svc.Write(ctx, req)
	require.NoError(t, e)
	require.Equal(t, "Renamed", st.Name)
	require.Equal(t, 2, st.Revision)
	bad := req
	bad.Name = "different"
	_, e = svc.Write(ctx, bad)
	require.ErrorIs(t, e, ErrConflict)
	_, e = svc.Write(ctx, change)
	require.NoError(t, e)
	stale := change
	stale.IdempotencyKey = "stale"
	_, e = svc.Write(ctx, stale)
	require.ErrorIs(t, e, ErrConflict)
	add := WriteRequest{UserID: u, ListID: id, MeaningID: meaning, Operation: "add", ExpectedRevision: 2, IdempotencyKey: "add"}
	st, e = svc.Write(ctx, add)
	require.NoError(t, e)
	require.Equal(t, 3, st.Revision)
	require.Equal(t, 1, st.MemberCount)
	require.Equal(t, 1, st.UsableMemberCount)
	require.Equal(t, "greeting", st.Members[0].WordSlug)
	// Reload through a new repository to prove state is database-authoritative.
	reloaded, e := NewService(NewPostgreSQLRepository(db), nil).Get(ctx, u, id)
	require.NoError(t, e)
	require.Equal(t, st, reloaded)
	psvc := practice.NewService(practice.NewPostgreSQLRepository(db), nil)
	revision := 3
	start := practice.StartRequest{Mode: "typed_recall", ListID: id.String(), ListRevision: &revision}
	_, e = psvc.Start(ctx, other, start, "foreign-list-practice")
	require.ErrorIs(t, e, practice.ErrNotFound)
	session, e := psvc.Start(ctx, u, start, "list-practice")
	require.NoError(t, e)
	require.Equal(t, 1, session.TotalSteps)
	require.Equal(t, "Renamed", session.ListName)
	require.Equal(t, 3, *session.ListRevision)
	listening, e := psvc.Start(ctx, u, practice.StartRequest{Mode: "listening_choice", ListID: id.String(), ListRevision: &revision}, "list-listening")
	require.NoError(t, e)
	require.Len(t, listening.CurrentStep.Choices, 3)
	remove := add
	remove.Operation = "remove"
	remove.ExpectedRevision = 3
	remove.IdempotencyKey = "remove"
	st, e = svc.Write(ctx, remove)
	require.NoError(t, e)
	require.Zero(t, st.MemberCount)
	st, e = svc.Write(ctx, add)
	require.NoError(t, e)
	require.Zero(t, st.MemberCount)
	require.Equal(t, 4, st.Revision)
	revision = 4
	_, e = psvc.Start(ctx, u, practice.StartRequest{Mode: "typed_recall", ListID: id.String(), ListRevision: &revision}, "empty")
	require.ErrorIs(t, e, practice.ErrListEmpty)
	revision = 3
	_, e = psvc.Start(ctx, u, start, "stale-list")
	require.ErrorIs(t, e, practice.ErrConflict)
	// Concurrent fresh writes with one expected revision: exactly one wins.
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := add
			r.ExpectedRevision = 4
			r.IdempotencyKey = uuid.NewString()
			_, e := svc.Write(ctx, r)
			errs <- e
		}()
	}
	wg.Wait()
	close(errs)
	wins, conflicts := 0, 0
	for e := range errs {
		if e == nil {
			wins++
		} else {
			require.ErrorIs(t, e, ErrConflict)
			conflicts++
		}
	}
	require.Equal(t, 1, wins)
	require.Equal(t, 1, conflicts)
	// Transaction failure must not persist membership, revision or receipt.
	exec(`ALTER TABLE word_list_actions ADD CONSTRAINT list_test_reject CHECK(idempotency_key <> 'force-list-rollback') NOT VALID`)
	rollback := remove
	rollback.ExpectedRevision = 5
	rollback.IdempotencyKey = "force-list-rollback"
	_, e = svc.Write(ctx, rollback)
	require.Error(t, e)
	exec(`ALTER TABLE word_list_actions DROP CONSTRAINT list_test_reject`)
	st, e = svc.Get(ctx, u, id)
	require.NoError(t, e)
	require.Equal(t, 5, st.Revision)
	require.Equal(t, 1, st.MemberCount)
	del := WriteRequest{UserID: u, ListID: id, Operation: "delete", ExpectedRevision: 5, IdempotencyKey: "delete"}
	_, e = svc.Write(ctx, del)
	require.NoError(t, e)
	_, e = svc.Write(ctx, del)
	require.NoError(t, e)
	_, e = svc.Get(ctx, u, id)
	require.ErrorIs(t, e, ErrNotFound)
	_, e = svc.Write(ctx, req)
	require.ErrorIs(t, e, ErrNotFound)
	recreate := req
	recreate.IdempotencyKey = "recreate"
	_, e = svc.Write(ctx, recreate)
	require.ErrorIs(t, e, ErrNotFound)
	frozen, e := psvc.Get(ctx, u, uuid.MustParse(session.ID))
	require.NoError(t, e)
	require.Equal(t, session.CurrentStep, frozen.CurrentStep)
	require.Equal(t, "Renamed", frozen.ListName)
	replay, e := psvc.Start(ctx, u, start, "list-practice")
	require.NoError(t, e)
	require.Equal(t, session.ID, replay.ID)
	ar := accounts.NewPostgreSQLRepository(db)
	exported, e := ar.ExportPersonalData(ctx, u)
	require.NoError(t, e)
	require.Contains(t, string(exported), "wordLists")
	require.Contains(t, string(exported), "Renamed")
	require.Contains(t, string(exported), id.String())
	require.NotContains(t, string(exported), "list-practice")
	require.NotContains(t, string(exported), "fingerprint")
	for _, table := range []string{"user_words", "review_attempts", "daily_mission_snapshots", "confidence_point_ledger", "user_word_knowledge"} {
		var n int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE user_id=$1", u).Scan(&n))
		require.Zero(t, n)
	}
	exec(`UPDATE users SET status='deleted',deleted_at=$2 WHERE id=$1`, u, time.Now())
	_, e = svc.List(ctx, u)
	require.ErrorIs(t, e, ErrNotFound)
	_, e = svc.Write(ctx, del)
	require.ErrorIs(t, e, ErrNotFound)
	_, e = psvc.Get(ctx, u, uuid.MustParse(session.ID))
	require.ErrorIs(t, e, practice.ErrNotFound)
	counts, e := ar.AnonymizeUserData(ctx, u)
	require.NoError(t, e)
	require.Equal(t, int64(1), counts.WordLists)
	require.Positive(t, counts.WordListActions)
	require.Equal(t, int64(2), counts.PracticeSessions)
	var receipts int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM word_list_actions WHERE user_id=$1`, u).Scan(&receipts))
	require.Zero(t, receipts)
}

func TestPostgreSQLListLimitsAndUnsupportedMembership(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	db, e := sql.Open("postgres", dsn)
	require.NoError(t, e)
	t.Cleanup(func() { db.Close() })
	ctx := t.Context()
	u, id, word := uuid.New(), uuid.New(), uuid.New()
	exec := func(q string, args ...any) { _, e := db.ExecContext(ctx, q, args...); require.NoError(t, e) }
	exec(`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, u, u.String()+"@lists-limits.invalid")
	exec(`INSERT INTO canonical_words(id,text,normalized_text,status,created_at,updated_at)VALUES($1,$2,$2,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, word, "list-fixture-"+word.String())
	t.Cleanup(func() {
		for _, q := range []string{`DELETE FROM word_list_actions WHERE user_id=$1`, `DELETE FROM user_word_lists WHERE user_id=$1`, `DELETE FROM users WHERE id=$1`} {
			_, e := db.ExecContext(context.Background(), q, u)
			require.NoError(t, e)
		}
		_, e := db.ExecContext(context.Background(), `DELETE FROM word_meanings WHERE word_id=$1`, word)
		require.NoError(t, e)
		_, e = db.ExecContext(context.Background(), `DELETE FROM canonical_words WHERE id=$1`, word)
		require.NoError(t, e)
	})
	exec(`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at) SELECT gen_random_uuid(),$1,'noun','An uncurated test meaning',n,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP FROM generate_series(1,500) n`, word)
	var meaning uuid.UUID
	require.NoError(t, db.QueryRowContext(ctx, `SELECT id FROM word_meanings WHERE word_id=$1 ORDER BY meaning_order LIMIT 1`, word).Scan(&meaning))
	svc := NewService(NewPostgreSQLRepository(db), nil)
	_, e = svc.Write(ctx, WriteRequest{UserID: u, ListID: id, Operation: "put", Name: "Unsupported", IdempotencyKey: "create"})
	require.NoError(t, e)
	st, e := svc.Write(ctx, WriteRequest{UserID: u, ListID: id, MeaningID: meaning, Operation: "add", ExpectedRevision: 1, IdempotencyKey: "unsupported"})
	require.NoError(t, e)
	require.Equal(t, 1, st.MemberCount)
	require.Zero(t, st.UsableMemberCount)
	require.False(t, st.Members[0].PracticeAvailable)
	revision := 2
	psvc := practice.NewService(practice.NewPostgreSQLRepository(db), nil)
	_, e = psvc.Start(ctx, u, practice.StartRequest{Mode: "typed_recall", ListID: id.String(), ListRevision: &revision}, "unsupported-practice")
	require.ErrorIs(t, e, practice.ErrListEmpty)
	exec(`INSERT INTO user_word_list_members(list_id,meaning_id,created_at)SELECT $1,id,CURRENT_TIMESTAMP FROM word_meanings WHERE word_id=$2 ON CONFLICT DO NOTHING`, id, word)
	st, e = svc.Get(ctx, u, id)
	require.NoError(t, e)
	require.Equal(t, 500, st.MemberCount)
	require.Zero(t, st.UsableMemberCount)
	_, e = svc.Write(ctx, WriteRequest{UserID: u, ListID: id, MeaningID: uuid.MustParse("ac93d068-7c3e-55c9-9d85-3fd3518592dc"), Operation: "add", ExpectedRevision: 2, IdempotencyKey: "over-limit"})
	require.ErrorIs(t, e, ErrLimit)
	// Adding an existing meaning at the limit stays a successful idempotent membership.
	st, e = svc.Write(ctx, WriteRequest{UserID: u, ListID: id, MeaningID: meaning, Operation: "add", ExpectedRevision: 2, IdempotencyKey: "existing"})
	require.NoError(t, e)
	require.Equal(t, 500, st.MemberCount)
	exec(`INSERT INTO user_word_lists(id,user_id,name,revision,created_at,updated_at)SELECT gen_random_uuid(),$1,'Private list '||n,1,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP FROM generate_series(1,49) n`, u)
	lists, e := svc.List(ctx, u)
	require.NoError(t, e)
	require.Len(t, lists.Items, 50)
	_, e = svc.Write(ctx, WriteRequest{UserID: u, ListID: uuid.New(), Operation: "put", Name: "One more", IdempotencyKey: "over-list-limit"})
	require.ErrorIs(t, e, ErrLimit)
}

func TestPostgreSQLMixedActiveAndArchivedListPractice(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	admin, e := sql.Open("postgres", dsn)
	require.NoError(t, e)
	schema := "list_availability_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, e = admin.ExecContext(t.Context(), "CREATE SCHEMA "+pq.QuoteIdentifier(schema))
	require.NoError(t, e)
	scoped := dsn
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		uri, e := url.Parse(dsn)
		require.NoError(t, e)
		q := uri.Query()
		q.Set("search_path", schema)
		uri.RawQuery = q.Encode()
		scoped = uri.String()
	} else {
		scoped += " search_path=" + schema
	}
	db, e := sql.Open("postgres", scoped)
	require.NoError(t, e)
	t.Cleanup(func() {
		db.Close()
		_, _ = admin.ExecContext(context.Background(), "DROP SCHEMA "+pq.QuoteIdentifier(schema)+" CASCADE")
		admin.Close()
	})
	ctx := t.Context()
	_, e = db.ExecContext(ctx, `CREATE TABLE users(id uuid PRIMARY KEY,status text NOT NULL,deleted_at timestamptz)`)
	require.NoError(t, e)
	content, e := os.ReadFile("../../migrations/20260725100000_voc026_p1_content_tables.sql")
	require.NoError(t, e)
	_, e = db.ExecContext(ctx, strings.SplitN(string(content), "CREATE TABLE user_words", 2)[0])
	require.NoError(t, e)
	for _, migration := range []string{"20261004100000_word_lists.sql", "20261002230000_practice_sessions.sql"} {
		raw, e := os.ReadFile("../../migrations/" + migration)
		require.NoError(t, e)
		_, e = db.ExecContext(ctx, string(raw))
		require.NoError(t, e)
	}
	raw, e := os.ReadFile("../../cmd/seed/voc026-p1.json")
	require.NoError(t, e)
	var seed map[string][]map[string]any
	require.NoError(t, json.Unmarshal(raw, &seed))
	for _, table := range []string{"canonical_words", "word_meanings"} {
		for _, row := range seed[table] {
			row["created_at"] = "2026-10-04T12:00:00Z"
			row["updated_at"] = "2026-10-04T12:00:00Z"
			value, e := json.Marshal(row)
			require.NoError(t, e)
			_, e = db.ExecContext(ctx, "INSERT INTO "+table+" SELECT * FROM jsonb_populate_record(NULL::"+table+",$1::jsonb)", string(value))
			require.NoError(t, e)
		}
	}
	u, id := uuid.New(), uuid.New()
	_, e = db.ExecContext(ctx, `INSERT INTO users(id,status)VALUES($1,'active')`, u)
	require.NoError(t, e)
	svc := NewService(NewPostgreSQLRepository(db), nil)
	_, e = svc.Write(ctx, WriteRequest{UserID: u, ListID: id, Operation: "put", Name: "Mixed available words", IdempotencyKey: "create"})
	require.NoError(t, e)
	first := uuid.MustParse("ac93d068-7c3e-55c9-9d85-3fd3518592dc")
	second := uuid.MustParse("3265c02c-5751-5465-b94b-e767aa871d8d")
	for i, meaning := range []uuid.UUID{first, second} {
		_, e = svc.Write(ctx, WriteRequest{UserID: u, ListID: id, MeaningID: meaning, Operation: "add", ExpectedRevision: i + 1, IdempotencyKey: meaning.String()})
		require.NoError(t, e)
	}
	_, e = db.ExecContext(ctx, `UPDATE word_meanings SET status='archived' WHERE id=$1`, second)
	require.NoError(t, e)
	detail, e := svc.Get(ctx, u, id)
	require.NoError(t, e)
	require.Equal(t, 2, detail.MemberCount)
	require.Equal(t, 1, detail.UsableMemberCount)
	revision := 3
	psvc := practice.NewService(practice.NewPostgreSQLRepository(db), nil)
	session, e := psvc.Start(ctx, u, practice.StartRequest{Mode: "typed_recall", ListID: id.String(), ListRevision: &revision}, "mixed-active")
	require.NoError(t, e)
	require.Equal(t, 1, session.TotalSteps)
	session, e = psvc.Act(ctx, u, uuid.MustParse(session.ID), practice.Action{StepID: session.CurrentStep.ID, ExpectedRevision: session.Revision, ClientActionID: "answer", Action: "answer", TypedAnswer: "greeting"}, "answer")
	require.NoError(t, e)
	require.Equal(t, first.String(), session.Feedback.MeaningID)
	// Listening still explicitly requires reviewed active distractors for this meaning.
	_, e = psvc.Start(ctx, u, practice.StartRequest{Mode: "listening_choice", ListID: id.String(), ListRevision: &revision}, "missing-distractor")
	require.ErrorIs(t, e, practice.ErrContentUnavailable)
	_, e = db.ExecContext(ctx, `UPDATE word_meanings SET status='archived' WHERE id=$1`, first)
	require.NoError(t, e)
	_, e = psvc.Start(ctx, u, practice.StartRequest{Mode: "typed_recall", ListID: id.String(), ListRevision: &revision}, "all-archived")
	require.ErrorIs(t, e, practice.ErrListEmpty)
}
