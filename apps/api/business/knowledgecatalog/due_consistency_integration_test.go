package knowledgecatalog_test

import (
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/dailylesson"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/knowledgecatalog"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/reviews"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/wordknowledge"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Historical assessment-only writes leave saved review history active. Every
// due projection must nevertheless respect the independent Known assessment.
func TestPostgreSQLKnownSavedWordDueConsistencyAcrossLearnerSurfaces(t *testing.T) {
	db := knowledgecatalog.IsolatedCatalogDBForTest(t)
	knowledgecatalog.PreloadAuthoredCoreForTest(t, db)
	ctx := t.Context()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	require.NoError(t, knowledgecatalog.Seed(ctx, tx))
	require.NoError(t, tx.Commit())
	user, saved := uuid.New(), uuid.New()
	var meaning uuid.UUID
	require.NoError(t, db.QueryRowContext(ctx, `SELECT meaning_id FROM knowledge_catalog_entries WHERE rank=1`).Scan(&meaning))
	_, err = db.ExecContext(ctx, `INSERT INTO users(id,email,status,created_at,updated_at) VALUES($1,$2,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, user, user.String()+"@due-consistency.invalid")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO user_words(id,user_id,meaning_id,status,source,review_step,total_review_count,correct_review_count,next_review_at,added_at,created_at,updated_at) VALUES($1,$2,$3,'reviewing','manual',3,5,4,$4,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, saved, user, meaning, time.Now().Add(-time.Hour))
	require.NoError(t, err)
	learn := learning.NewPostgreSQLRepository(db)
	review := reviews.NewPostgreSQLRepository(db, nil)
	catalog := knowledgecatalog.NewService(knowledgecatalog.NewPostgreSQLRepository(db), nil)
	daily := dailylesson.NewService(db, review, nil)
	check := func(wantDue int) {
		t.Helper()
		summary, e := learn.GetKnowledgeSummary(ctx, user)
		require.NoError(t, e)
		require.Equal(t, 1, summary.Saved)
		require.Equal(t, wantDue, summary.Due)
		all, e := learn.ListSavedWords(ctx, learning.ListSavedWordsRequest{UserID: user})
		require.NoError(t, e)
		require.Len(t, all.Items, 1)
		require.Equal(t, wantDue == 1, all.Items[0].Due)
		due, e := learn.ListSavedWords(ctx, learning.ListSavedWordsRequest{UserID: user, DueOnly: true})
		require.NoError(t, e)
		require.Equal(t, wantDue, due.TotalCount)
		require.Len(t, due.Items, wantDue)
		detail, e := learn.SavedWordStates(ctx, user, []uuid.UUID{meaning})
		require.NoError(t, e)
		require.Equal(t, wantDue == 1, detail[meaning].Due)
		require.Equal(t, "reviewing", detail[meaning].Status)
		queue, e := review.ListDueWords(ctx, reviews.ListDueWordsRequest{UserID: user})
		require.NoError(t, e)
		require.Equal(t, wantDue, queue.TotalCount)
		m, e := catalog.Read(ctx, user, 1)
		require.NoError(t, e)
		require.Equal(t, wantDue == 1, m.CurrentBand.Items[0].Due)
		current, e := daily.Current(ctx, user)
		require.NoError(t, e)
		require.Equal(t, wantDue, current.DueCount)
	}
	check(1)
	assess := wordknowledge.NewService(wordknowledge.NewPostgreSQLRepository(db), nil)
	_, err = assess.Write(ctx, wordknowledge.WriteRequest{UserID: user, MeaningID: meaning, SelfReportedKnown: true, Note: "Keep my note and history", IdempotencyKey: "legacy-assess-known"})
	require.NoError(t, err)
	check(0)
	_, err = assess.Write(ctx, wordknowledge.WriteRequest{UserID: user, MeaningID: meaning, SelfReportedKnown: false, Note: "Keep my note and history", IdempotencyKey: "legacy-assess-learning"})
	require.NoError(t, err)
	check(1)
	var status, note string
	var step, total, correct int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT uw.status,uw.review_step,uw.total_review_count,uw.correct_review_count,k.note FROM user_words uw JOIN user_word_knowledge k USING(user_id,meaning_id) WHERE uw.id=$1`, saved).Scan(&status, &step, &total, &correct, &note))
	require.Equal(t, "reviewing", status)
	require.Equal(t, 3, step)
	require.Equal(t, 5, total)
	require.Equal(t, 4, correct)
	require.Equal(t, "Keep my note and history", note)
}
