package discoveryposition

import (
	"database/sql"
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/knowledgecatalog"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestPositionValidation(t *testing.T) {
	good := Update{ExpectedAccountID: uuid.New(), CatalogVersion: knowledgecatalog.CatalogVersion, Band: 44, Rank: 4330, Active: true}
	require.NoError(t, good.Validate())
	for _, invalid := range []Update{
		{CatalogVersion: "old", Band: 44, Rank: 4330},
		{CatalogVersion: knowledgecatalog.CatalogVersion, Band: 43, Rank: 4330},
		{CatalogVersion: knowledgecatalog.CatalogVersion, Band: 1, Rank: 0},
		{CatalogVersion: knowledgecatalog.CatalogVersion, Band: 101, Rank: 10001},
		{CatalogVersion: knowledgecatalog.CatalogVersion, Band: 44, Rank: 4330, ExpectedRevision: -1},
	} {
		invalid.ExpectedAccountID = good.ExpectedAccountID
		require.ErrorIs(t, invalid.Validate(), ErrInvalid)
	}
}

func TestPostgreSQLPositionIsolationCASAndDeletedAccount(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	schema := "position_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = admin.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE")
		require.NoError(t, err)
		admin.Close()
	})
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	paths, err := filepath.Glob("../../migrations/*.sql")
	require.NoError(t, err)
	for _, path := range paths {
		if strings.HasSuffix(path, ".down.sql") {
			continue
		}
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = db.Exec(string(raw))
		require.NoError(t, err, path)
	}
	for rank := 4330; rank <= 4334; rank++ {
		w, m := uuid.New(), uuid.New()
		_, err = db.Exec(`INSERT INTO canonical_words(id,text,normalized_text,language_code,status,created_at,updated_at)VALUES($1,$2,$2,'en','active',now(),now())`, w, w.String())
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at)VALUES($1,$2,'noun','Synthetic position fixture',1,'active',now(),now())`, m, w)
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO knowledge_catalog_entries(rank,word_id,meaning_id,source_rank,source_count,source_sense_key,source_name,catalog_version)VALUES($1,$2,$3,$1,1,'synthetic','synthetic','english-10000-v1')`, rank, w, m)
		require.NoError(t, err)
	}
	ctx := t.Context()
	uid, other := uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{uid, other} {
		_, err = db.ExecContext(ctx, `INSERT INTO users(id,email,status,onboarding_status,created_at,updated_at) VALUES($1,$2,'active','completed',now(),now())`, id, id.String()+"@position.invalid")
		require.NoError(t, err)
	}
	t.Cleanup(func() {
		for _, id := range []uuid.UUID{uid, other} {
			db.Exec(`DELETE FROM user_discovery_positions WHERE user_id=$1`, id)
			db.Exec(`DELETE FROM users WHERE id=$1`, id)
		}
	})
	repo := NewPostgreSQLRepository(db)
	svc := NewService(repo)
	p, err := svc.Get(ctx, uid)
	require.NoError(t, err)
	require.Nil(t, p)
	accountRepo := accounts.NewPostgreSQLRepository(db)
	export := func(id uuid.UUID) map[string]any {
		t.Helper()
		raw, e := accountRepo.ExportPersonalData(ctx, id)
		require.NoError(t, e)
		var value map[string]any
		require.NoError(t, json.Unmarshal(raw, &value))
		delete(value, "exportedAt")
		return value
	}
	before := export(uid)
	update := Update{ExpectedAccountID: uid, CatalogVersion: knowledgecatalog.CatalogVersion, Band: 44, Rank: 4330, Active: true}
	p, err = svc.Put(ctx, uid, update)
	require.NoError(t, err)
	require.EqualValues(t, 1, p.Revision)
	require.Equal(t, 4330, p.Rank)
	after := export(uid)
	require.EqualValues(t, 4330, after["discoveryPosition"].(map[string]any)["rank"])
	require.Nil(t, export(other)["discoveryPosition"], "preference export must be requester-scoped")
	delete(before, "discoveryPosition")
	delete(after, "discoveryPosition")
	require.Equal(t, before, after, "position changes no exported learning progress or settings")

	// A stale page retains A as its owner even when the authenticated request is B.
	// Both have revision 1, so a revision-only fence would accept this overwrite.
	otherPosition, err := svc.Put(ctx, other, Update{ExpectedAccountID: other, CatalogVersion: knowledgecatalog.CatalogVersion, Band: 44, Rank: 4334})
	require.NoError(t, err)
	require.EqualValues(t, 1, otherPosition.Revision)
	stale := Update{ExpectedAccountID: uid, CatalogVersion: knowledgecatalog.CatalogVersion, Band: 44, Rank: 4331, ExpectedRevision: 1}
	for range 2 {
		_, err = svc.Put(ctx, other, stale)
		require.ErrorIs(t, err, ErrAccountChanged)
		_, err = repo.Put(ctx, other, stale)
		require.ErrorIs(t, err, ErrAccountChanged)
	}
	unchanged, err := svc.Get(ctx, other)
	require.NoError(t, err)
	require.Equal(t, otherPosition, unchanged)
	_, err = db.ExecContext(ctx, `DELETE FROM user_discovery_positions WHERE user_id=$1`, other)
	require.NoError(t, err)
	// Lost-response replay succeeds without increment, but other stale intent fails.
	replay, err := svc.Put(ctx, uid, update)
	require.NoError(t, err)
	require.Equal(t, p, replay)
	update.Rank = 4331
	_, err = svc.Put(ctx, uid, update)
	require.ErrorIs(t, err, ErrConflict)
	foreign, err := svc.Get(ctx, other)
	require.NoError(t, err)
	require.Nil(t, foreign)
	// Two competing writes against one revision can confirm only one destination.
	var wg sync.WaitGroup
	errors := make(chan error, 2)
	for _, rank := range []int{4332, 4333} {
		wg.Add(1)
		go func(rank int) {
			defer wg.Done()
			_, e := svc.Put(ctx, uid, Update{ExpectedAccountID: uid, CatalogVersion: knowledgecatalog.CatalogVersion, Band: 44, Rank: rank, Active: true, ExpectedRevision: 1})
			errors <- e
		}(rank)
	}
	wg.Wait()
	close(errors)
	successes, conflicts := 0, 0
	for e := range errors {
		if e == nil {
			successes++
		} else {
			require.ErrorIs(t, e, ErrConflict)
			conflicts++
		}
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, conflicts)
	p, err = svc.Get(ctx, uid)
	require.NoError(t, err)
	require.EqualValues(t, 2, p.Revision)
	_, err = db.ExecContext(ctx, `UPDATE users SET deleted_at=now() WHERE id=$1`, uid)
	require.NoError(t, err)
	_, err = svc.Get(ctx, uid)
	require.ErrorIs(t, err, ErrNotFound)
	_, err = svc.Put(ctx, uid, Update{ExpectedAccountID: uid, CatalogVersion: knowledgecatalog.CatalogVersion, Band: 44, Rank: 4334, ExpectedRevision: 2})
	require.ErrorIs(t, err, ErrNotFound)
	_, err = svc.Put(ctx, other, Update{ExpectedAccountID: other, CatalogVersion: knowledgecatalog.CatalogVersion, Band: 44, Rank: 4330})
	require.NoError(t, err)
	counters, err := accountRepo.AnonymizeUserData(ctx, uid)
	require.NoError(t, err)
	require.EqualValues(t, 1, counters.DiscoveryPositions)
	var remaining int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM user_discovery_positions WHERE user_id=$1`, uid).Scan(&remaining))
	require.Zero(t, remaining)
	foreign, err = svc.Get(ctx, other)
	require.NoError(t, err)
	require.Equal(t, 4330, foreign.Rank)

}
