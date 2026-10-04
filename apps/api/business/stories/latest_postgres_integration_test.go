package stories

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestStoriesPostgreSQLLatestPerKeySkipsOlderSnapshots(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	db, e := sql.Open("postgres", dsn)
	require.NoError(t, e)
	t.Cleanup(func() { db.Close() })
	ctx := t.Context()
	owner, other, empty := uuid.New(), uuid.New(), uuid.New()
	for _, u := range []uuid.UUID{owner, other, empty} {
		_, e = db.ExecContext(ctx, `INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, u, u.String()+"@story-latest.invalid")
		require.NoError(t, e)
	}
	t.Cleanup(func() {
		for _, u := range []uuid.UUID{owner, other, empty} {
			for _, table := range []string{"story_sessions", "users"} {
				column := "user_id"
				if table == "users" {
					column = "id"
				}
				_, e := db.ExecContext(context.Background(), "DELETE FROM "+table+" WHERE "+column+"=$1", u)
				require.NoError(t, e)
			}
		}
	})
	repo := NewPostgreSQLRepository(db)
	states, e := repo.List(ctx, empty)
	require.NoError(t, e)
	require.Empty(t, states)
	now := time.Now().UTC()
	insert := func(u, id uuid.UUID, key string, snapshot string, updated time.Time) {
		t.Helper()
		_, e := db.ExecContext(ctx, `INSERT INTO story_sessions(id,user_id,story_key,content_version,grading_version,snapshot,total_steps,created_at,updated_at)VALUES($1,$2,$3,$4,$5,$6,10,$7,$7)`, id, u, key, ContentVersion, GradingVersion, snapshot, updated)
		require.NoError(t, e)
	}
	snap, e := build("a-quiet-lunch")
	require.NoError(t, e)
	snap.Story.Title = "A frozen previous title"
	raw, e := json.Marshal(snap)
	require.NoError(t, e)
	// Many attempts and even an older unsupported snapshot cannot make the
	// current library decode every historical record or fail on an unselected one.
	for i := 0; i < 100; i++ {
		insert(owner, uuid.New(), snap.Story.Key, string(raw), now.Add(-time.Duration(i+1)*time.Hour))
	}
	insert(owner, uuid.New(), snap.Story.Key, `{}`, now.Add(-200*time.Hour))
	lower := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	higher := uuid.MustParse("f0000000-0000-4000-8000-000000000001")
	insert(owner, lower, snap.Story.Key, string(raw), now)
	insert(owner, higher, snap.Story.Key, string(raw), now)
	second, e := build("the-right-platform")
	require.NoError(t, e)
	secondRaw, e := json.Marshal(second)
	require.NoError(t, e)
	secondID := uuid.New()
	insert(owner, secondID, second.Story.Key, string(secondRaw), now.Add(-time.Minute))
	// Foreign attempts must not hide or replace the owner's most recent record.
	insert(other, uuid.New(), snap.Story.Key, `{}`, now.Add(time.Hour))
	states, e = repo.List(ctx, owner)
	require.NoError(t, e)
	require.Len(t, states, 2)
	byKey := map[string]State{}
	for _, st := range states {
		byKey[st.Snapshot.Story.Key] = st
	}
	require.Equal(t, higher, byKey[snap.Story.Key].ID, "ties retain deterministic UUID order")
	require.Equal(t, "A frozen previous title", byKey[snap.Story.Key].Snapshot.Story.Title)
	require.Equal(t, secondID, byKey[second.Story.Key].ID)
	var count int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM story_sessions WHERE user_id=$1`, owner).Scan(&count))
	require.Equal(t, 104, count, "historical attempts remain persisted")
	_, e = db.ExecContext(ctx, `UPDATE users SET deleted_at=CURRENT_TIMESTAMP WHERE id=$1`, owner)
	require.NoError(t, e)
	_, e = repo.List(ctx, owner)
	require.ErrorIs(t, e, ErrNotFound)
}
