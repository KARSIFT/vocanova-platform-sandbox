package knowledgecatalog_test

import (
	"os"
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

// A server-audited answer or explicit Skip retires every use of that canonical
// word for the local day, even if remove/restore or stale state resets its date.
func TestPostgreSQLWrongAndSkipRestoreDueConsistencyAcrossLearnerSurfaces(t *testing.T) {
	db := knowledgecatalog.IsolatedCatalogDBForTest(t)
	knowledgecatalog.PreloadAuthoredCoreForTest(t, db)
	ctx := t.Context()
	raw, err := os.ReadFile("../../cmd/seed/voc026-p1.json")
	require.NoError(t, err)
	for _, query := range []string{
		`INSERT INTO journey_situations(id,slug,title,short_description,level_band,category,status,display_order,created_at,updated_at) SELECT (x->>'id')::uuid,x->>'slug',x->>'title',x->>'short_description',x->>'level_band',x->>'category',x->>'status',(x->>'display_order')::integer,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP FROM jsonb_array_elements($1::jsonb->'journey_situations') x`,
		`INSERT INTO journey_words(id,journey_situation_id,meaning_id,relevance_score,display_order,is_core,created_at,updated_at) SELECT (x->>'id')::uuid,(x->>'journey_situation_id')::uuid,(x->>'meaning_id')::uuid,(x->>'relevance_score')::numeric,(x->>'display_order')::integer,(x->>'is_core')::boolean,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP FROM jsonb_array_elements($1::jsonb->'journey_words') x`,
	} {
		_, err = db.ExecContext(ctx, query, string(raw))
		require.NoError(t, err)
	}
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	require.NoError(t, knowledgecatalog.Seed(ctx, tx))
	require.NoError(t, tx.Commit())
	var meaning, wordID uuid.UUID
	var definition string
	require.NoError(t, db.QueryRowContext(ctx, `SELECT kc.meaning_id,kc.word_id,wm.short_definition FROM knowledge_catalog_entries kc JOIN word_meanings wm ON wm.id=kc.meaning_id WHERE kc.rank=1`).Scan(&meaning, &wordID, &definition))
	sibling := uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at) VALUES($1,$2,'noun','A separate synthetic use of this canonical word.',99,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, sibling, wordID)
	require.NoError(t, err)
	for _, action := range []string{"wrong", "skip"} {
		t.Run(action, func(t *testing.T) {
			user := uuid.New()
			now := time.Now().UTC()
			_, err := db.ExecContext(ctx, `INSERT INTO users(id,email,status,created_at,updated_at) VALUES($1,$2,'active',$3,$3)`, user, user.String()+"@review-audit.invalid", now)
			require.NoError(t, err)
			_, err = db.ExecContext(ctx, `INSERT INTO user_settings(id,user_id,timezone,created_at,updated_at) VALUES(gen_random_uuid(),$1,'Pacific/Kiritimati',$2,$2)`, user, now)
			require.NoError(t, err)
			// The primary use is deterministically first. The sibling is equally
			// eligible except for its later due time, and cannot create a second card.
			for i, id := range []uuid.UUID{meaning, sibling} {
				_, err = db.ExecContext(ctx, `INSERT INTO user_words(id,user_id,meaning_id,status,source,review_step,next_review_at,added_at,created_at,updated_at) VALUES($1,$2,$3,'reviewing','manual',2,$4,$5,$5,$5)`, uuid.New(), user, id, now.Add(time.Duration(i-2)*time.Hour), now)
				require.NoError(t, err)
			}
			learn := learning.NewPostgreSQLRepository(db)
			review := reviews.NewPostgreSQLRepository(db, nil)
			catalog := knowledgecatalog.NewService(knowledgecatalog.NewPostgreSQLRepository(db), nil)
			daily := dailylesson.NewService(db, review, nil)
			check := func(wantDue int) knowledgecatalog.Item {
				t.Helper()
				summary, e := learn.GetKnowledgeSummary(ctx, user)
				require.NoError(t, e)
				require.Equal(t, 2, summary.Saved, "saved totals remain meaning counts")
				require.Equal(t, wantDue, summary.Due)
				all, e := learn.ListSavedWords(ctx, learning.ListSavedWordsRequest{UserID: user})
				require.NoError(t, e)
				require.Len(t, all.Items, 2)
				flagged := 0
				for _, item := range all.Items {
					if item.Due {
						flagged++
					}
				}
				require.Equal(t, wantDue, flagged)
				due, e := learn.ListSavedWords(ctx, learning.ListSavedWordsRequest{UserID: user, DueOnly: true})
				require.NoError(t, e)
				require.Equal(t, wantDue, due.TotalCount)
				require.Len(t, due.Items, wantDue)
				states, e := learn.SavedWordStates(ctx, user, []uuid.UUID{meaning, sibling})
				require.NoError(t, e)
				require.Equal(t, wantDue == 1, states[meaning].Due)
				require.False(t, states[sibling].Due)
				queue, e := review.ListDueWords(ctx, reviews.ListDueWordsRequest{UserID: user})
				require.NoError(t, e)
				require.Equal(t, wantDue, queue.TotalCount)
				require.Len(t, queue.Items, wantDue)
				m, e := catalog.Read(ctx, user, 1)
				require.NoError(t, e)
				require.Equal(t, wantDue == 1, m.CurrentBand.Items[0].Due, "map follows canonical review audit after either outcome")
				current, e := daily.Current(ctx, user)
				require.NoError(t, e)
				require.Equal(t, wantDue, current.DueCount)
				return m.CurrentBand.Items[0]
			}
			check(1)
			session, err := daily.Start(ctx, user, "start")
			require.NoError(t, err)
			require.Equal(t, 1, session.TotalWords)
			request := dailylesson.Action{StepID: session.CurrentStep.ID, ExpectedRevision: session.Revision, ClientActionID: "answer", Action: "skip"}
			if action == "wrong" {
				request.Action = "answer"
				for _, choice := range session.CurrentStep.Choices {
					if choice.Text != definition {
						request.ChoiceID = choice.ID
						break
					}
				}
				require.NotEmpty(t, request.ChoiceID)
			}
			answered, err := daily.Act(ctx, user, uuid.MustParse(session.ID), request, "answer")
			require.NoError(t, err)
			require.False(t, answered.Feedback.Correct)
			require.False(t, answered.Feedback.Credited)
			_, err = daily.Act(ctx, user, uuid.MustParse(session.ID), dailylesson.Action{StepID: answered.CurrentStep.ID, ExpectedRevision: answered.Revision, ClientActionID: "continue", Action: "continue"}, "continue")
			require.NoError(t, err)
			check(0)
			for i, id := range []uuid.UUID{meaning, sibling} {
				require.NoError(t, learn.UnsaveUserWord(ctx, user, id, time.Now().UTC()))
				_, err = learn.SaveUserWord(ctx, learning.SaveUserWordRequest{UserID: user, MeaningID: id, Source: "manual"}, time.Now().UTC())
				require.NoError(t, err)
				// Model a legacy restore that also erased schedule/last-review
				// fields. Only server-created audit can keep this out of today.
				_, err = db.ExecContext(ctx, `UPDATE user_words SET next_review_at=$3,last_reviewed_at=NULL WHERE user_id=$1 AND meaning_id=$2 AND deleted_at IS NULL`, user, id, now.Add(time.Duration(i-2)*time.Hour))
				require.NoError(t, err)
			}
			item := check(0)
			_, err = catalog.Decide(ctx, knowledgecatalog.DecisionRequest{UserID: user, MeaningID: meaning, Decision: "known", ExpectedRevision: item.Revision, IdempotencyKey: "known"})
			require.NoError(t, err)
			item = check(0)
			result, err := catalog.Decide(ctx, knowledgecatalog.DecisionRequest{UserID: user, MeaningID: meaning, Decision: "learn", ExpectedRevision: item.Revision, IdempotencyKey: "learn"})
			require.NoError(t, err)
			require.False(t, result.Due, "decision response uses the same audit eligibility as the map")
			check(0)
			_, err = daily.Start(ctx, user, "same-day-restart")
			require.ErrorIs(t, err, dailylesson.ErrNoDue)
			var attempts, successes int
			require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM review_attempts WHERE user_id=$1`, user).Scan(&attempts))
			require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM daily_review_successes WHERE user_id=$1`, user).Scan(&successes))
			require.Equal(t, 1, attempts)
			require.Zero(t, successes)
		})
	}
}
