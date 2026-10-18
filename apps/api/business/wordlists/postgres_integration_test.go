package wordlists

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/knowledgecatalog"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaninguses"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/practice"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"net/url"
	"os"
	"path/filepath"
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

func TestPostgreSQLSupplementalListAvailabilityMatchesFrozenPractice(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	schema := "list_supplement_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = admin.ExecContext(t.Context(), "CREATE SCHEMA "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		dsn, err = pq.ParseURL(dsn)
		require.NoError(t, err)
	}
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
		_, e := admin.ExecContext(context.Background(), "DROP SCHEMA "+pq.QuoteIdentifier(schema)+" CASCADE")
		require.NoError(t, e)
		admin.Close()
	})
	ctx := t.Context()
	migrations, err := filepath.Glob("../../migrations/*.sql")
	require.NoError(t, err)
	for _, path := range migrations {
		raw, e := os.ReadFile(path)
		require.NoError(t, e)
		_, e = db.ExecContext(ctx, string(raw))
		require.NoError(t, e, path)
	}
	raw, err := os.ReadFile("../../cmd/seed/voc026-p1.json")
	require.NoError(t, err)
	var seed map[string][]map[string]any
	require.NoError(t, json.Unmarshal(raw, &seed))
	for _, table := range []string{"canonical_words", "word_meanings", "word_examples"} {
		for _, row := range seed[table] {
			row["created_at"] = "2026-10-18T12:00:00Z"
			row["updated_at"] = "2026-10-18T12:00:00Z"
			value, e := json.Marshal(row)
			require.NoError(t, e)
			_, e = db.ExecContext(ctx, "INSERT INTO "+table+" SELECT * FROM jsonb_populate_record(NULL::"+table+",$1::jsonb)", string(value))
			require.NoError(t, e)
		}
	}
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	require.NoError(t, knowledgecatalog.Seed(ctx, tx))
	require.NoError(t, meaninguses.Seed(ctx, tx))
	require.NoError(t, tx.Commit())
	u, other := uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{u, other} {
		_, err = db.ExecContext(ctx, `INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, id, id.String()+"@supplement-list.invalid")
		require.NoError(t, err)
	}
	lists := NewService(NewPostgreSQLRepository(db), nil)
	psvc := practice.NewService(practice.NewPostgreSQLRepository(db), clock.Fixed{T: time.Now().UTC().Truncate(time.Microsecond)})
	entries := meaninguses.Entries()
	create := func(t *testing.T, ids []string) *WordListDetail {
		t.Helper()
		id := uuid.New()
		_, e := lists.Write(ctx, WriteRequest{UserID: u, ListID: id, Operation: "put", Name: "Selected uses", IdempotencyKey: uuid.NewString()})
		require.NoError(t, e)
		revision := 1
		for _, meaning := range ids {
			_, e = lists.Write(ctx, WriteRequest{UserID: u, ListID: id, MeaningID: uuid.MustParse(meaning), Operation: "add", ExpectedRevision: revision, IdempotencyKey: uuid.NewString()})
			require.NoError(t, e)
			revision++
		}
		detail, e := lists.Get(ctx, u, id)
		require.NoError(t, e)
		return detail
	}
	check := func(t *testing.T, mode string, ids []string) {
		t.Helper()
		detail := create(t, ids)
		require.Equal(t, len(ids), detail.UsableMemberCount)
		for _, member := range detail.Members {
			require.True(t, member.PracticeAvailable)
		}
		revision := detail.Revision
		req := practice.StartRequest{Mode: mode, ListID: detail.ID, ListRevision: &revision}
		key := uuid.NewString()
		session, e := psvc.Start(ctx, u, req, key)
		require.NoError(t, e, "every advertised available use must start")
		require.Equal(t, detail.UsableMemberCount, session.TotalSteps, "list practice cannot silently omit available senses")
		var snapshot practice.Snapshot
		var data []byte
		require.NoError(t, db.QueryRowContext(ctx, `SELECT snapshot FROM practice_sessions WHERE id=$1`, session.ID).Scan(&data))
		require.NoError(t, json.Unmarshal(data, &snapshot))
		want := map[string]bool{}
		for _, id := range ids {
			want[id] = true
		}
		actual := map[string]bool{}
		for _, saved := range snapshot.Steps {
			require.True(t, want[saved.Word.MeaningID], "distractors must not become selected targets")
			require.False(t, actual[saved.Word.MeaningID])
			actual[saved.Word.MeaningID] = true
			require.NotEmpty(t, saved.Word.PartOfSpeech)
			if meaninguses.SupportsMeaning(saved.Word.MeaningID, saved.Word.WordText, saved.Word.Definition) {
				context, ok := meaninguses.ReviewedContext(saved.Word.MeaningID, saved.Word.WordText, saved.Word.Definition)
				require.True(t, ok)
				require.Equal(t, context, saved.Word.Example)
			}
			if mode == "listening_choice" {
				require.GreaterOrEqual(t, len(saved.Public.Choices), 3)
				require.NotEmpty(t, saved.Public.SpeechText)
			}
		}
		require.Equal(t, want, actual)
		if mode == "listening_choice" {
			first := snapshot.Steps[0].Word
			require.Equal(t, first.MeaningID, session.CurrentStep.PronunciationMeaningID)
			require.Equal(t, meaningidentity.Revision(first.MeaningID, first.WordText, first.PartOfSpeech, first.Definition), session.CurrentStep.PronunciationRevision)
		}
		replay, e := psvc.Start(ctx, u, req, key)
		require.NoError(t, e)
		require.Equal(t, session, replay)
		_, e = psvc.Start(ctx, other, req, uuid.NewString())
		require.ErrorIs(t, e, practice.ErrNotFound)
		firstSaved := snapshot.Steps[0]
		answer := practice.Action{StepID: session.CurrentStep.ID, ExpectedRevision: session.Revision, ClientActionID: uuid.NewString(), Action: "answer"}
		if mode == "listening_choice" {
			answer.ChoiceID = firstSaved.CorrectChoice
		} else {
			answer.TypedAnswer = firstSaved.Word.WordText
		}
		answerKey := uuid.NewString()
		session, e = psvc.Act(ctx, u, uuid.MustParse(session.ID), answer, answerKey)
		require.NoError(t, e)
		require.True(t, session.Feedback.Correct)
		require.Equal(t, firstSaved.Word.MeaningID, session.Feedback.MeaningID)
		require.Equal(t, firstSaved.Word.PartOfSpeech, session.Feedback.PartOfSpeech)
		require.Equal(t, 1, session.FirstAnswersCorrect)
		ack, e := psvc.Act(ctx, u, uuid.MustParse(session.ID), answer, answerKey)
		require.NoError(t, e)
		require.Equal(t, session, ack)

		// Frozen target selection and metadata survive source edits and list deletion.
		first := snapshot.Steps[0].Word
		_, e = db.ExecContext(ctx, `UPDATE word_meanings SET short_definition='Edited after the session started.' WHERE id=$1`, first.MeaningID)
		require.NoError(t, e)
		reloaded, e := psvc.Get(ctx, u, uuid.MustParse(session.ID))
		require.NoError(t, e)
		require.Equal(t, session.CurrentStep, reloaded.CurrentStep)
		_, e = db.ExecContext(ctx, `UPDATE word_meanings SET short_definition=$2 WHERE id=$1`, first.MeaningID, first.Definition)
		require.NoError(t, e)
		_, e = lists.Write(ctx, WriteRequest{UserID: u, ListID: uuid.MustParse(detail.ID), Operation: "delete", ExpectedRevision: detail.Revision, IdempotencyKey: uuid.NewString()})
		require.NoError(t, e)
		reloaded, e = psvc.Get(ctx, u, uuid.MustParse(session.ID))
		require.NoError(t, e)
		require.Equal(t, session.CurrentStep, reloaded.CurrentStep)
		_, e = psvc.Start(ctx, u, req, uuid.NewString())
		require.ErrorIs(t, e, practice.ErrNotFound)

	}
	for _, entry := range entries {
		for _, mode := range []string{"typed_recall", "listening_choice"} {
			t.Run(entry.SourceSenseKey+"/"+mode, func(t *testing.T) { check(t, mode, []string{entry.MeaningID.String()}) })
		}
	}
	for _, mode := range []string{"typed_recall", "listening_choice"} {
		t.Run("mixed/"+mode, func(t *testing.T) {
			check(t, mode, []string{"ac93d068-7c3e-55c9-9d85-3fd3518592dc", entries[0].MeaningID.String(), entries[3].MeaningID.String()})
		})
	}
	// Both list availability and newly created practice exclude retired or edited uses.
	valid, archived, changed := entries[0], entries[1], entries[2]
	detail := create(t, []string{valid.MeaningID.String(), archived.MeaningID.String(), changed.MeaningID.String()})
	_, err = db.ExecContext(ctx, `UPDATE word_meanings SET status='archived' WHERE id=$1`, archived.MeaningID)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `UPDATE word_meanings SET short_definition='Not the reviewed definition.' WHERE id=$1`, changed.MeaningID)
	require.NoError(t, err)
	detail, err = lists.Get(ctx, u, uuid.MustParse(detail.ID))
	require.NoError(t, err)
	require.Equal(t, 1, detail.UsableMemberCount)
	for _, mode := range []string{"typed_recall", "listening_choice"} {
		revision := detail.Revision
		session, e := psvc.Start(ctx, u, practice.StartRequest{Mode: mode, ListID: detail.ID, ListRevision: &revision}, uuid.NewString())
		require.NoError(t, e)
		require.Equal(t, 1, session.TotalSteps)
		var data []byte
		require.NoError(t, db.QueryRowContext(ctx, `SELECT snapshot FROM practice_sessions WHERE id=$1`, session.ID).Scan(&data))
		var snapshot practice.Snapshot
		require.NoError(t, json.Unmarshal(data, &snapshot))
		require.Equal(t, valid.MeaningID.String(), snapshot.Steps[0].Word.MeaningID)
	}

	// A reused supplemental ID with a different POS is not advertised or loaded.
	_, err = db.ExecContext(ctx, `UPDATE word_meanings SET short_definition=$2,part_of_speech='noun' WHERE id=$1`, changed.MeaningID, changed.Definition)
	require.NoError(t, err)
	detail, err = lists.Get(ctx, u, uuid.MustParse(detail.ID))
	require.NoError(t, err)
	require.Equal(t, 1, detail.UsableMemberCount)
	for _, member := range detail.Members {
		if member.MeaningID == changed.MeaningID.String() {
			require.False(t, member.PracticeAvailable)
		}
	}
	for _, mode := range []string{"typed_recall", "listening_choice"} {
		revision := detail.Revision
		session, e := psvc.Start(ctx, u, practice.StartRequest{Mode: mode, ListID: detail.ID, ListRevision: &revision}, uuid.NewString())
		require.NoError(t, e)
		require.Equal(t, 1, session.TotalSteps)
	}
	// A foreign-language canonical row cannot reuse these English identities.
	_, err = db.ExecContext(ctx, `UPDATE canonical_words SET language_code='fr' WHERE id=$1`, valid.WordID)
	require.NoError(t, err)
	detail, err = lists.Get(ctx, u, uuid.MustParse(detail.ID))
	require.NoError(t, err)
	require.Zero(t, detail.UsableMemberCount)
	for _, mode := range []string{"typed_recall", "listening_choice"} {
		revision := detail.Revision
		_, e := psvc.Start(ctx, u, practice.StartRequest{Mode: mode, ListID: detail.ID, ListRevision: &revision}, uuid.NewString())
		require.ErrorIs(t, e, practice.ErrListEmpty)
	}
	for _, table := range []string{"user_word_knowledge", "user_words", "review_attempts", "confidence_point_ledger"} {
		var count int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE user_id=$1", u).Scan(&count))
		require.Zero(t, count, table)
	}
}
