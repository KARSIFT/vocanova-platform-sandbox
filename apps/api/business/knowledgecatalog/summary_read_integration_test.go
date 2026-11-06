package knowledgecatalog

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPostgreSQLMapSummaryPreservesHistoryAndUserBoundaries(t *testing.T) {
	db := isolatedCatalogDB(t)
	preloadAuthoredCore(t, db)
	ctx := t.Context()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	require.NoError(t, Seed(ctx, tx))
	require.NoError(t, tx.Commit())
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	exec := func(query string, args ...any) {
		t.Helper()
		_, e := db.ExecContext(ctx, query, args...)
		require.NoError(t, e)
	}
	user, other := uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{user, other} {
		exec(`INSERT INTO users(id,email,status,created_at,updated_at) VALUES($1,$2,'active',$3,$3)`, id, id.String()+"@summary.invalid", now)
	}
	meanings := map[int]uuid.UUID{}
	for _, rank := range []int{1, 2, 3, 4, 5, 6, 7, 100, 101, 200, 201} {
		var id uuid.UUID
		require.NoError(t, db.QueryRowContext(ctx, `SELECT meaning_id FROM knowledge_catalog_entries WHERE rank=$1`, rank).Scan(&id))
		meanings[rank] = id
	}
	known := func(owner uuid.UUID, rank int, value bool) {
		t.Helper()
		exec(`INSERT INTO user_word_knowledge(user_id,meaning_id,self_reported_known,note,updated_at) VALUES($1,$2,$3,'',$4)`, owner, meanings[rank], value, now)
	}
	saved := func(id, owner uuid.UUID, rank int, status string, deleted any) {
		t.Helper()
		exec(`INSERT INTO user_words(id,user_id,meaning_id,status,source,review_step,total_review_count,next_review_at,added_at,created_at,updated_at,deleted_at) VALUES($1,$2,$3,$4,'manual',2,4,$5,$5,$5,$5,$6)`, id, owner, meanings[rank], status, now.Add(-24*time.Hour), deleted)
	}
	known(user, 1, true) // Known wins even when an active saved row remains.
	saved(uuid.New(), user, 1, "reviewing", nil)
	saved(uuid.New(), user, 2, "learning", nil)
	saved(uuid.New(), user, 3, "ignored", nil)
	saved(uuid.New(), user, 4, "archived", nil)
	saved(uuid.New(), user, 5, "learning", now.Add(-time.Hour))
	saved(uuid.New(), user, 6, "learning", now.Add(-time.Hour))
	saved(uuid.New(), user, 6, "reviewing", nil)
	known(user, 7, false)
	saved(uuid.New(), user, 100, "mastered", nil)
	known(user, 101, true)
	saved(uuid.New(), user, 200, "new", nil)
	// Removed rows tie on deletion and update time. The smaller UUID is the
	// existing deterministic restore target, despite another later removal.
	winner := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	loser := uuid.MustParse("20000000-0000-4000-8000-000000000001")
	saved(winner, user, 201, "learning", now.Add(-2*time.Hour))
	saved(loser, user, 201, "reviewing", now.Add(-2*time.Hour))
	saved(uuid.New(), user, 201, "mastered", now.Add(-time.Hour))
	known(other, 2, true)
	saved(uuid.New(), other, 5, "learning", nil)
	saved(uuid.New(), other, 201, "reviewing", nil)

	repo := NewPostgreSQLRepository(db)
	read := func(owner uuid.UUID, band int) *Map {
		t.Helper()
		state, e := repo.Read(ctx, owner, band, now)
		require.NoError(t, e)
		require.Equal(t, TotalWords, state.TotalWords)
		require.Len(t, state.Bands, 100)
		require.Len(t, state.CurrentBand.Items, BandSize)
		return state
	}
	checkBands := func(state *Map, counts map[int][2]int) {
		t.Helper()
		for i, band := range state.Bands {
			want := counts[i+1]
			require.Equal(t, Band{Number: i + 1, StartRank: i*100 + 1, EndRank: (i + 1) * 100, TotalWords: 100, KnownCount: want[0], LearningCount: want[1], UnseenCount: 100 - want[0] - want[1]}, band)
		}
	}
	first := read(user, 1)
	checkBands(first, map[int][2]int{1: {1, 3}, 2: {1, 1}})
	for i, item := range first.CurrentBand.Items {
		require.Equal(t, i+1, item.Rank)
		require.Equal(t, i == 0, item.SelfReportedKnown)
		require.Equal(t, i == 0 || i == 1 || i == 5 || i == 99, item.Saved)
	}
	otherState := read(other, 3)
	checkBands(otherState, map[int][2]int{1: {1, 1}, 3: {0, 1}})
	removed := read(user, 3)
	checkBands(removed, map[int][2]int{1: {1, 3}, 2: {1, 1}})
	require.False(t, removed.CurrentBand.Items[0].Saved)
	require.False(t, removed.CurrentBand.Items[0].Due)
	before := removed.CurrentBand.Items[0].Revision
	// A nonselected historical row cannot invalidate the visible revision.
	exec(`UPDATE user_words SET review_step=3 WHERE id=$1`, loser)
	require.Equal(t, before, read(user, 3).CurrentBand.Items[0].Revision)
	// A change to the selected removed row must still invalidate stale tabs,
	// even though removed history contributes nothing to the band summary.
	exec(`UPDATE user_words SET review_step=3 WHERE id=$1`, winner)
	after := read(user, 3)
	require.NotEqual(t, before, after.CurrentBand.Items[0].Revision)
	require.Equal(t, removed.Bands, after.Bands)
	request := DecisionRequest{UserID: user, MeaningID: meanings[201], Decision: "learn", ExpectedRevision: before, IdempotencyKey: "restore-history"}
	_, err = repo.Decide(ctx, request, now)
	require.ErrorIs(t, err, ErrStale)
	request.ExpectedRevision = after.CurrentBand.Items[0].Revision
	restored, err := repo.Decide(ctx, request, now)
	require.NoError(t, err)
	require.True(t, restored.Saved)
	var restoredID uuid.UUID
	var reviews, step int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT id,total_review_count,review_step FROM user_words WHERE user_id=$1 AND meaning_id=$2 AND deleted_at IS NULL`, user, meanings[201]).Scan(&restoredID, &reviews, &step))
	require.Equal(t, winner, restoredID)
	require.Equal(t, 4, reviews)
	require.Equal(t, 3, step)
	checkBands(read(user, 3), map[int][2]int{1: {1, 3}, 2: {1, 1}, 3: {0, 1}})
	require.Equal(t, otherState, read(other, 3))
}
