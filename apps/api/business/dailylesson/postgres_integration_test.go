package dailylesson

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/knowledgecatalog"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/missions"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/reviews"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

type fixture struct {
	db     *sql.DB
	svc    *Service
	review *reviews.PostgreSQLRepository
	u      uuid.UUID
	now    time.Time
	words  []uuid.UUID
}

func setup(t *testing.T, n int) *fixture {
	t.Helper()
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		var err error
		dsn, err = pq.ParseURL(dsn)
		require.NoError(t, err)
	}
	admin, e := sql.Open("postgres", dsn)
	require.NoError(t, e)
	schema := "daily_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, e = admin.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, e)
	t.Cleanup(func() {
		_, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE")
		require.NoError(t, err)
		admin.Close()
	})
	db, e := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, e)
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	paths, e := filepath.Glob("../../migrations/*.sql")
	require.NoError(t, e)
	for _, p := range paths {
		if strings.HasSuffix(p, ".down.sql") {
			continue
		}
		raw, err := os.ReadFile(p)
		require.NoError(t, err)
		_, err = db.Exec(string(raw))
		require.NoError(t, err, p)
	}
	f := &fixture{db: db, u: uuid.New(), now: time.Now().UTC().Truncate(time.Microsecond)}
	exec := func(q string, args ...any) { t.Helper(); _, e := db.Exec(q, args...); require.NoError(t, e) }
	exec("INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',$3,$3)", f.u, f.u.String()+"@daily.invalid", f.now)
	exec("INSERT INTO user_settings(id,user_id,timezone,daily_review_target,created_at,updated_at)VALUES(gen_random_uuid(),$1,'Asia/Tehran',5,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)", f.u)
	situation := uuid.New()
	exec("INSERT INTO journey_situations(id,slug,title,short_description,category,status,display_order,created_at,updated_at)VALUES($1,$2,'Synthetic daily','Synthetic authored alternatives','daily_life','active',1,$3,$3)", situation, "daily-"+situation.String(), f.now)
	for i := 0; i < n+5; i++ {
		w, m, uw := uuid.New(), uuid.New(), uuid.New()
		text := fmt.Sprintf("daily-%s-%02d", f.u.String()[:8], i)
		exec("INSERT INTO canonical_words(id,text,normalized_text,language_code,status,created_at,updated_at)VALUES($1,$2,$2,'en','active',$3,$3)", w, text, f.now)
		exec("INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at)VALUES($1,$2,'noun',$3,1,'active',$4,$4)", m, w, fmt.Sprintf("An independent definition %s %d", f.u, i), f.now)
		exec("INSERT INTO journey_words(id,journey_situation_id,meaning_id,created_at,updated_at)VALUES(gen_random_uuid(),$1,$2,$3,$3)", situation, m, f.now)
		if i < n {
			exec("INSERT INTO user_words(id,user_id,meaning_id,status,source,review_step,added_at,created_at,updated_at)VALUES($1,$2,$3,'new','search',0,$4,$4,$4)", uw, f.u, m, f.now)
			f.words = append(f.words, uw)
		}
	}
	gam := gamification.NewService(gamification.NewRepository(db))
	mis := missions.NewService(missions.NewRepository(db), gam)
	f.review = reviews.NewPostgreSQLRepository(db, clock.Fixed{T: f.now}, reviews.WithGamificationService(gam), reviews.WithMissionsService(mis))
	f.svc = NewService(db, f.review, clock.Fixed{T: f.now})
	return f
}
func (f *fixture) state(t *testing.T, s *Session) *state {
	t.Helper()
	st, e := scan(f.db.QueryRow("SELECT state FROM daily_lesson_sessions WHERE id=$1", s.ID))
	require.NoError(t, e)
	return st
}
func (f *fixture) answer(t *testing.T, s *Session, correct bool) *Session {
	t.Helper()
	st := f.state(t, s)
	w := st.Words[st.Queue[0]]
	choice := w.AnswerID
	if !correct {
		for _, c := range w.Public.Choices {
			if c.ID != choice {
				choice = c.ID
				break
			}
		}
	}
	a := Action{StepID: s.CurrentStep.ID, ExpectedRevision: s.Revision, ClientActionID: uuid.NewString(), Action: "answer", ChoiceID: choice}
	key := uuid.NewString()
	out, e := f.svc.Act(t.Context(), f.u, st.ID, a, key)
	require.NoError(t, e)
	if out.Feedback.Reason == "" {
		require.Equal(t, choice, out.Feedback.SelectedChoiceID)
	} else {
		require.Empty(t, out.Feedback.SelectedChoiceID)
	}
	reloaded, e := f.svc.Get(t.Context(), f.u, st.ID)
	require.NoError(t, e)
	require.Equal(t, out.Feedback, reloaded.Feedback)
	replayed, e := f.svc.Act(t.Context(), f.u, st.ID, a, key)
	require.NoError(t, e)
	require.Equal(t, out.Feedback, replayed.Feedback)
	return out
}
func (f *fixture) next(t *testing.T, s *Session) *Session {
	t.Helper()
	out, e := f.svc.Act(t.Context(), f.u, uuid.MustParse(s.ID), Action{StepID: s.CurrentStep.ID, ExpectedRevision: s.Revision, ClientActionID: uuid.NewString(), Action: "continue"}, uuid.NewString())
	require.NoError(t, e)
	return out
}
func TestPostgreSQLDailyLessonAuthorityRetryCreditResumeAndExport(t *testing.T) {
	f := setup(t, 2)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	s, e := f.svc.Start(ctx, f.u, "start")
	require.NoError(t, e)
	require.Equal(t, 2, s.TotalWords)
	require.Len(t, s.CurrentStep.Choices, 4)
	raw, e := json.Marshal(s)
	require.NoError(t, e)
	require.NotContains(t, string(raw), "AnswerID")
	require.NotContains(t, string(raw), "ReviewCount")
	resumed, e := f.svc.Start(ctx, f.u, "resume")
	require.NoError(t, e)
	require.Equal(t, s.ID, resumed.ID)
	_, e = f.svc.Get(ctx, uuid.New(), uuid.MustParse(s.ID))
	require.ErrorIs(t, e, ErrNotFound)
	s = f.answer(t, s, false)
	require.False(t, s.Feedback.Correct)
	require.Zero(t, s.CreditedToday)
	var mission int
	require.NoError(t, f.db.QueryRow("SELECT reviews_completed FROM daily_mission_snapshots WHERE user_id=$1", f.u).Scan(&mission))
	require.Zero(t, mission)
	s = f.next(t, s)
	second := s.CurrentStep.ID
	s = f.answer(t, s, true)
	require.True(t, s.Feedback.Correct)
	require.Equal(t, 1, s.CreditedToday)
	s = f.next(t, s)
	require.NotEqual(t, second, s.CurrentStep.ID)
	s = f.answer(t, s, true)
	require.Equal(t, 2, s.CreditedToday)
	require.Equal(t, 1, s.FirstAnswersCorrect)
	require.Equal(t, 2, s.QuestionsAnswered)
	s = f.next(t, s)
	require.Equal(t, "completed", s.Status)
	require.Nil(t, s.CurrentStep)
	current, e := f.svc.Current(ctx, f.u)
	require.NoError(t, e)
	require.Nil(t, current.Session)
	require.Zero(t, current.DueCount)
	var due int
	require.NoError(t, f.db.QueryRow("SELECT count(*) FROM user_words WHERE user_id=$1 AND next_review_at<$2", f.u, f.now.Add(24*time.Hour)).Scan(&due))
	require.Zero(t, due)
	require.NoError(t, f.db.QueryRow("SELECT reviews_completed FROM daily_mission_snapshots WHERE user_id=$1", f.u).Scan(&mission))
	require.Equal(t, 2, mission)
	export, e := accounts.NewPostgreSQLRepository(f.db).ExportPersonalData(ctx, f.u)
	require.NoError(t, e)
	require.Contains(t, string(export), "\"dailyLessons\"")
	require.Contains(t, string(export), "\"dailyReviewSuccesses\"")
	require.NotContains(t, string(export), "AnswerID")
	require.NotContains(t, string(export), "idempotency_key")
}
func TestPostgreSQLDailyLessonConcurrentReplayStaleAndRemoved(t *testing.T) {
	f := setup(t, 1)
	s, e := f.svc.Start(t.Context(), f.u, "start")
	require.NoError(t, e)
	st := f.state(t, s)
	w := st.Words[0]
	a := Action{StepID: s.CurrentStep.ID, ExpectedRevision: 0, ClientActionID: "one-answer", Action: "answer", ChoiceID: w.AnswerID}
	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, e := f.svc.Act(t.Context(), f.u, st.ID, a, "answer-key")
			require.NoError(t, e)
			if got != nil {
				require.Equal(t, 1, got.CreditedToday)
			}
		}()
	}
	wg.Wait()
	var n int
	require.NoError(t, f.db.QueryRow("SELECT count(*) FROM review_attempts WHERE user_id=$1", f.u).Scan(&n))
	require.Equal(t, 1, n)
	_, e = f.svc.Act(t.Context(), f.u, st.ID, Action{StepID: a.StepID, ExpectedRevision: 0, ClientActionID: "different", Action: "answer", ChoiceID: a.ChoiceID}, "different")
	require.ErrorIs(t, e, ErrConflict)
	a.ChoiceID = "wrong"
	_, e = f.svc.Act(t.Context(), f.u, st.ID, a, "answer-key")
	require.ErrorIs(t, e, ErrConflict)
	// Existing review endpoint cannot create a second scheduled review of the now-not-due word.
	_, e = f.review.SubmitReview(t.Context(), reviews.SubmitReviewRequest{UserID: f.u, UserWordID: uuid.MustParse(w.UserWordID), MeaningID: uuid.MustParse(w.MeaningID), Source: "review", AttemptType: "review", PromptType: "self_check", Result: "correct", Rating: "good", AnsweredAt: f.now, ClientAttemptID: "fresh-legacy", IdempotencyKey: "fresh-legacy"})
	require.ErrorIs(t, e, reviews.ErrNotDue)
}
func TestPostgreSQLDailyLessonStaleWordAndChunkBoundary(t *testing.T) {
	f := setup(t, 12)
	s, e := f.svc.Start(t.Context(), f.u, "start")
	require.NoError(t, e)
	require.Equal(t, 10, s.TotalWords)
	st := f.state(t, s)
	w := st.Words[0]
	_, e = f.db.Exec("UPDATE user_words SET deleted_at=$2 WHERE id=$1", w.UserWordID, f.now)
	require.NoError(t, e)
	s = f.answer(t, s, true)
	require.True(t, s.Feedback.Skipped)
	require.False(t, s.Feedback.Correct)
	require.Equal(t, "removed", s.Feedback.Reason)
	require.Zero(t, s.CreditedToday)
	s = f.next(t, s)
	for s.Status != "completed" {
		s = f.answer(t, s, true)
		s = f.next(t, s)
	}
	current, e := f.svc.Current(t.Context(), f.u)
	require.NoError(t, e)
	require.Equal(t, 2, current.DueCount)
	chunk, e := f.svc.Start(t.Context(), f.u, "next-chunk")
	require.NoError(t, e)
	require.Equal(t, 2, chunk.TotalWords)
	require.NotEqual(t, s.ID, chunk.ID)
}
func TestPostgreSQLDailyLessonRollbackAndKnown(t *testing.T) {
	f := setup(t, 1)
	s, e := f.svc.Start(t.Context(), f.u, "start")
	require.NoError(t, e)
	st := f.state(t, s)
	w := st.Words[0]
	_, e = f.db.Exec("CREATE OR REPLACE FUNCTION reject_daily_action() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.idempotency_key='force-rollback' THEN RAISE EXCEPTION 'forced test rollback'; END IF; RETURN NEW; END $$")
	require.NoError(t, e)
	_, e = f.db.Exec("CREATE TRIGGER daily_test_rollback BEFORE INSERT ON daily_lesson_actions FOR EACH ROW EXECUTE FUNCTION reject_daily_action()")
	require.NoError(t, e)
	t.Cleanup(func() {
		_, _ = f.db.Exec("DROP TRIGGER IF EXISTS daily_test_rollback ON daily_lesson_actions")
		_, _ = f.db.Exec("DROP FUNCTION IF EXISTS reject_daily_action()")
	})
	a := Action{StepID: s.CurrentStep.ID, ExpectedRevision: 0, ClientActionID: "rollback", Action: "answer", ChoiceID: w.AnswerID}
	_, e = f.svc.Act(t.Context(), f.u, st.ID, a, "force-rollback")
	require.Error(t, e)
	require.True(t, strings.Contains(e.Error(), "forced test rollback"))
	for _, table := range []string{"review_attempts", "daily_review_successes", "daily_mission_snapshots", "confidence_point_ledger"} {
		var n int
		require.NoError(t, f.db.QueryRow("SELECT count(*) FROM "+table+" WHERE user_id=$1", f.u).Scan(&n))
		require.Zero(t, n, table)
	}
	_, e = f.db.Exec("INSERT INTO user_word_knowledge(user_id,meaning_id,self_reported_known,note)VALUES($1,$2,true,'keep my note')", f.u, w.MeaningID)
	require.NoError(t, e)
	got, e := f.svc.Act(t.Context(), f.u, st.ID, a, "known-answer")
	require.NoError(t, e)
	require.True(t, got.Feedback.Skipped)
	require.Equal(t, "already_known", got.Feedback.Reason)
}

func TestPostgreSQLDailyLessonDailyReceiptLocalDateHintAndPurge(t *testing.T) {
	f := setup(t, 1)
	// Just after Tehran midnight, while UTC is still on the previous date.
	f.now = time.Date(2026, 10, 17, 20, 31, 0, 0, time.UTC)
	gam := gamification.NewService(gamification.NewRepository(f.db))
	mis := missions.NewService(missions.NewRepository(f.db), gam)
	f.review = reviews.NewPostgreSQLRepository(f.db, clock.Fixed{T: f.now}, reviews.WithGamificationService(gam), reviews.WithMissionsService(mis))
	f.svc = NewService(f.db, f.review, clock.Fixed{T: f.now})
	s, e := f.svc.Start(t.Context(), f.u, "daily")
	require.NoError(t, e)
	st := f.state(t, s)
	w := st.Words[0]
	hinted, e := f.svc.Act(t.Context(), f.u, st.ID, Action{StepID: s.CurrentStep.ID, ExpectedRevision: 0, ClientActionID: "hint", Action: "answer", ChoiceID: w.AnswerID, WasHintUsed: true}, "hint")
	require.NoError(t, e)
	require.False(t, hinted.Feedback.Correct)
	require.False(t, hinted.Feedback.Credited)
	s = f.next(t, hinted)
	s = f.answer(t, s, true)
	require.Equal(t, 1, s.CreditedToday)
	var date string
	require.NoError(t, f.db.QueryRow("SELECT local_date::text FROM daily_review_successes WHERE user_id=$1", f.u).Scan(&date))
	require.Equal(t, "2026-10-18", date)
	s = f.next(t, s)
	require.Equal(t, "completed", s.Status)
	// Restoring a due timestamp cannot make a verified meaning eligible again today.
	_, e = f.db.Exec("UPDATE user_words SET next_review_at=NULL WHERE id=$1", w.UserWordID)
	require.NoError(t, e)
	_, e = f.review.SubmitReview(t.Context(), reviews.SubmitReviewRequest{UserID: f.u, UserWordID: uuid.MustParse(w.UserWordID), MeaningID: uuid.MustParse(w.MeaningID), Source: "review", AttemptType: "review", PromptType: "self_check", Result: "correct", Rating: "good", AnsweredAt: f.now, ClientAttemptID: "second-eligible", IdempotencyKey: "second-eligible"})
	require.ErrorIs(t, e, reviews.ErrNotDue)
	var n int
	require.NoError(t, f.db.QueryRow("SELECT count(*) FROM review_attempts WHERE user_id=$1", f.u).Scan(&n))
	require.Equal(t, 2, n, "hinted attempt plus later correct answer; denied legacy attempt writes nothing")
	require.NoError(t, f.db.QueryRow("SELECT count(*) FROM daily_review_successes WHERE user_id=$1", f.u).Scan(&n))
	require.Equal(t, 1, n)
	require.NoError(t, f.db.QueryRow("SELECT reviews_completed FROM daily_mission_snapshots WHERE user_id=$1", f.u).Scan(&n))
	require.Equal(t, 1, n)
	// Schema binds the receipt's user and meaning to its actual attempt.
	_, e = f.db.Exec("INSERT INTO daily_review_successes(user_id,meaning_id,local_date,review_attempt_id,created_at,daily_session_id) SELECT user_id,gen_random_uuid(),local_date+1,review_attempt_id,created_at,daily_session_id FROM daily_review_successes WHERE user_id=$1", f.u)
	require.Error(t, e)
	f.svc.clock = clock.Fixed{T: f.now.Add(24 * time.Hour)}
	overnight, e := f.svc.Get(t.Context(), f.u, uuid.MustParse(s.ID))
	require.NoError(t, e)
	require.Zero(t, overnight.CreditedToday, "yesterday's credits are not shown as today's")
	counters, e := accounts.NewPostgreSQLRepository(f.db).AnonymizeUserData(t.Context(), f.u)
	require.NoError(t, e)
	require.EqualValues(t, 1, counters.DailyLessonSessions)
	require.Positive(t, counters.DailyLessonActions)
	require.EqualValues(t, 1, counters.DailyReviewSuccesses)
	for _, table := range []string{"daily_lesson_sessions", "daily_lesson_actions", "daily_review_successes", "review_attempts", "user_words"} {
		require.NoError(t, f.db.QueryRow("SELECT count(*) FROM "+table+" WHERE user_id=$1", f.u).Scan(&n))
		require.Zero(t, n)
	}
}
func TestPostgreSQLDailyLessonLiveGenerationAndAccountClosure(t *testing.T) {
	f := setup(t, 1)
	s, e := f.svc.Start(t.Context(), f.u, "start")
	require.NoError(t, e)
	st := f.state(t, s)
	w := st.Words[0]
	// A different review owns the new generation. The open daily session cannot apply it again.
	_, e = f.review.SubmitReview(t.Context(), reviews.SubmitReviewRequest{UserID: f.u, UserWordID: uuid.MustParse(w.UserWordID), MeaningID: uuid.MustParse(w.MeaningID), Source: "review", AttemptType: "review", PromptType: "self_check", Result: "correct", Rating: "good", AnsweredAt: f.now, ClientAttemptID: "external", IdempotencyKey: "external"})
	require.NoError(t, e)
	s = f.answer(t, s, true)
	require.True(t, s.Feedback.Skipped)
	require.Equal(t, "already_reviewed", s.Feedback.Reason)
	require.Zero(t, s.CreditedToday)
	_, e = f.db.Exec("UPDATE users SET deleted_at=CURRENT_TIMESTAMP,updated_at=CURRENT_TIMESTAMP WHERE id=$1", f.u)
	require.NoError(t, e)
	_, e = f.svc.Get(t.Context(), f.u, st.ID)
	require.ErrorIs(t, e, ErrNotFound)
	_, e = f.svc.Current(t.Context(), f.u)
	require.ErrorIs(t, e, ErrNotFound)
	_, e = f.svc.Start(t.Context(), f.u, "closed")
	require.ErrorIs(t, e, ErrNotFound)
}

func TestPostgreSQLDailyReviewAndLegacySaveShareUserFirstLockOrder(t *testing.T) {
	f := setup(t, 1)
	f.db.SetMaxOpenConns(4)
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	s, e := f.svc.Start(ctx, f.u, "start")
	require.NoError(t, e)
	w := f.state(t, s).Words[0]
	tx, e := f.db.BeginTx(ctx, nil)
	require.NoError(t, e)
	defer tx.Rollback()
	require.NoError(t, lockUser(ctx, tx, f.u))
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		close(started)
		_, e := learning.NewPostgreSQLRepository(f.db).SaveUserWord(ctx, learning.SaveUserWordRequest{UserID: f.u, MeaningID: uuid.MustParse(w.MeaningID), Source: "search"}, f.now)
		done <- e
	}()
	<-started
	select {
	case e := <-done:
		t.Fatalf("save escaped active user transaction: %v", e)
	case <-time.After(50 * time.Millisecond):
	}
	_, e = f.review.SubmitReviewTx(ctx, tx, reviews.SubmitReviewRequest{UserID: f.u, UserWordID: uuid.MustParse(w.UserWordID), MeaningID: uuid.MustParse(w.MeaningID), Source: "review_session", AttemptType: "review", PromptType: "self_check", Result: "correct", Rating: "good", AnsweredAt: f.now, ClientAttemptID: "lock-order", IdempotencyKey: "lock-order"})
	require.NoError(t, e)
	require.NoError(t, tx.Commit())
	require.NoError(t, <-done)
	var n int
	require.NoError(t, f.db.QueryRowContext(ctx, "SELECT total_review_count FROM user_words WHERE id=$1", w.UserWordID).Scan(&n))
	require.Equal(t, 1, n)
}

func TestPostgreSQLDailyLessonRejectsSameSynsetDistractors(t *testing.T) {
	f := setup(t, 1)
	rows, e := f.db.Query("SELECT wm.id,cw.id FROM word_meanings wm JOIN canonical_words cw ON cw.id=wm.word_id ORDER BY cw.normalized_text")
	require.NoError(t, e)
	type reference struct{ meaning, word uuid.UUID }
	refs := []reference{}
	for rows.Next() {
		var r reference
		require.NoError(t, rows.Scan(&r.meaning, &r.word))
		refs = append(refs, r)
	}
	require.NoError(t, rows.Close())
	require.Len(t, refs, 6)
	for i, r := range refs {
		// All alternatives are deliberately a different gloss of the SAME source sense.
		_, e = f.db.Exec("INSERT INTO knowledge_catalog_entries(rank,word_id,meaning_id,source_rank,source_count,source_sense_key,source_name,catalog_version)VALUES($1,$2,$3,$1,1,'shared-test-synset','synthetic','english-10000-v1')", i+1, r.word, r.meaning)
		require.NoError(t, e)
	}
	_, e = f.svc.Start(t.Context(), f.u, "ambiguous")
	require.ErrorIs(t, e, ErrContentUnavailable)
	var n int
	require.NoError(t, f.db.QueryRow("SELECT count(*) FROM daily_lesson_sessions WHERE user_id=$1", f.u).Scan(&n))
	require.Zero(t, n)
}

func TestPostgreSQLSingleConnectionLegacySaveWithMissionSettings(t *testing.T) {
	f := setup(t, 1)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	var meaning uuid.UUID
	require.NoError(t, f.db.QueryRowContext(ctx, "SELECT wm.id FROM word_meanings wm WHERE NOT EXISTS(SELECT 1 FROM user_words uw WHERE uw.meaning_id=wm.id AND uw.user_id=$1) LIMIT 1", f.u).Scan(&meaning))
	gam := gamification.NewService(gamification.NewRepository(f.db))
	mis := missions.NewService(missions.NewRepository(f.db), gam)
	repo := learning.NewPostgreSQLRepository(f.db, gam, mis)
	saved, e := repo.SaveUserWord(ctx, learning.SaveUserWordRequest{UserID: f.u, MeaningID: meaning, Source: "search"}, f.now)
	require.NoError(t, e)
	require.True(t, saved.Saved)
	var reviewsCompleted, wordsAdded int
	require.NoError(t, f.db.QueryRowContext(ctx, "SELECT reviews_completed FROM daily_mission_snapshots WHERE user_id=$1", f.u).Scan(&reviewsCompleted))
	require.Zero(t, reviewsCompleted)
	require.NoError(t, f.db.QueryRowContext(ctx, "SELECT words_added FROM daily_activity_summaries WHERE user_id=$1", f.u).Scan(&wordsAdded))
	require.Equal(t, 1, wordsAdded)
}

func TestPostgreSQLForgedReviewMetadataCannotCreateVerifiedDailyCredit(t *testing.T) {
	f := setup(t, 1)
	s, e := f.svc.Start(t.Context(), f.u, "start")
	require.NoError(t, e)
	// Add another due meaning AFTER the snapshot. It must never count in this lesson.
	var meaning uuid.UUID
	require.NoError(t, f.db.QueryRow("SELECT id FROM word_meanings WHERE id NOT IN(SELECT meaning_id FROM user_words WHERE user_id=$1) LIMIT 1", f.u).Scan(&meaning))
	uw := uuid.New()
	_, e = f.db.Exec("INSERT INTO user_words(id,user_id,meaning_id,status,source,added_at,created_at,updated_at)VALUES($1,$2,$3,'new','search',$4,$4,$4)", uw, f.u, meaning, f.now)
	require.NoError(t, e)
	req := reviews.SubmitReviewRequest{UserID: f.u, UserWordID: uw, MeaningID: meaning, Source: "review", AttemptType: "review", PromptType: "self_check", Result: "correct", Rating: "good", AnsweredAt: f.now, ClientAttemptID: "forged-client", IdempotencyKey: "forged-key", Metadata: map[string]any{"dailySessionId": s.ID}}
	public := reviews.NewService(f.review, nil, clock.Fixed{T: f.now})
	_, e = public.SubmitReview(t.Context(), req)
	require.ErrorIs(t, e, reviews.ErrReservedMetadata)
	// Even bypassing HTTP/service metadata rejection cannot forge the private origin.
	_, e = f.review.SubmitReview(t.Context(), req)
	require.NoError(t, e)
	got, e := f.svc.Get(t.Context(), f.u, uuid.MustParse(s.ID))
	require.NoError(t, e)
	require.Zero(t, got.CreditedToday)
	var n int
	require.NoError(t, f.db.QueryRow("SELECT count(*) FROM daily_review_successes WHERE user_id=$1", f.u).Scan(&n))
	require.Zero(t, n)
	require.NoError(t, f.db.QueryRow("SELECT reviews_completed FROM daily_mission_snapshots WHERE user_id=$1", f.u).Scan(&n))
	require.Zero(t, n)
	// The typed internal path also refuses an owned meaning outside the frozen session.
	tx, e := f.db.BeginTx(t.Context(), nil)
	require.NoError(t, e)
	defer tx.Rollback()
	require.NoError(t, lockUser(t.Context(), tx, f.u))
	_, e = f.review.SubmitDailyReviewTx(t.Context(), tx, req, uuid.MustParse(s.ID))
	require.ErrorIs(t, e, reviews.ErrUserWordNotFound)
}

func TestPostgreSQLDailyLessonEditorialSynonymsCannotBecomeWrongOptions(t *testing.T) {
	for _, pair := range [][2]string{{"baggage", "luggage"}, {"answer", "reply"}, {"reservation", "booking"}, {"bag", "luggage"}, {"bags", "luggage"}} {
		t.Run(pair[0], func(t *testing.T) {
			f := setup(t, 1)
			rows, e := f.db.Query("SELECT wm.id,cw.id,wm.short_definition FROM word_meanings wm JOIN canonical_words cw ON cw.id=wm.word_id ORDER BY cw.normalized_text")
			require.NoError(t, e)
			type ref struct {
				m, w       uuid.UUID
				definition string
			}
			refs := []ref{}
			for rows.Next() {
				var r ref
				require.NoError(t, rows.Scan(&r.m, &r.w, &r.definition))
				refs = append(refs, r)
			}
			require.NoError(t, rows.Close())
			require.Len(t, refs, 6)
			_, e = f.db.Exec("UPDATE canonical_words SET text=$2,normalized_text=$2 WHERE id=$1", refs[0].w, pair[0])
			require.NoError(t, e)
			_, e = f.db.Exec("UPDATE canonical_words SET text=$2,normalized_text=$2 WHERE id=$1", refs[1].w, pair[1])
			require.NoError(t, e)
			// Three distinct editorial definitions refer to the same potential synonym.
			for i := 2; i <= 3; i++ {
				_, e = f.db.Exec("UPDATE word_meanings SET word_id=$2,meaning_order=$3 WHERE id=$1", refs[i].m, refs[1].w, i)
				require.NoError(t, e)
			}
			s, e := f.svc.Start(t.Context(), f.u, "aliases")
			require.NoError(t, e)
			require.Len(t, s.CurrentStep.Choices, 3, "only the correct meaning and two unrelated alternatives remain")
			for _, choice := range s.CurrentStep.Choices {
				for i := 1; i <= 3; i++ {
					require.NotEqual(t, refs[i].definition, choice.Text)
				}
			}
		})
	}
}

func TestPostgreSQLDailyWrongRetriesCannotFarmRatingPoints(t *testing.T) {
	f := setup(t, 1)
	s, e := f.svc.Start(t.Context(), f.u, "xp")
	require.NoError(t, e)
	for range 2 {
		s = f.answer(t, s, false)
		s = f.next(t, s)
	}
	var count, balance int
	require.NoError(t, f.db.QueryRow("SELECT count(*),COALESCE(sum(amount),0) FROM confidence_point_ledger WHERE user_id=$1", f.u).Scan(&count, &balance))
	require.Zero(t, count)
	require.Zero(t, balance)
	st := f.state(t, s)
	w := st.Words[0]
	a := Action{StepID: s.CurrentStep.ID, ExpectedRevision: s.Revision, ClientActionID: "correct-xp", Action: "answer", ChoiceID: w.AnswerID}
	correct, e := f.svc.Act(t.Context(), f.u, st.ID, a, "correct-xp")
	require.NoError(t, e)
	require.True(t, correct.Feedback.Credited)
	replay, e := f.svc.Act(t.Context(), f.u, st.ID, a, "correct-xp")
	require.NoError(t, e)
	require.Equal(t, correct, replay)
	require.NoError(t, f.db.QueryRow("SELECT count(*),COALESCE(sum(amount),0) FROM confidence_point_ledger WHERE user_id=$1", f.u).Scan(&count, &balance))
	require.Equal(t, 1, count)
	require.Equal(t, gamification.RewardReviewGood, balance)
	// A second internally graded correct result with no fresh local-day credit
	// still cannot generate a second rating award.
	_, e = f.db.Exec("UPDATE user_words SET next_review_at=NULL WHERE id=$1", w.UserWordID)
	require.NoError(t, e)
	tx, e := f.db.BeginTx(t.Context(), nil)
	require.NoError(t, e)
	require.NoError(t, lockUser(t.Context(), tx, f.u))
	req := reviews.SubmitReviewRequest{UserID: f.u, UserWordID: uuid.MustParse(w.UserWordID), MeaningID: uuid.MustParse(w.MeaningID), Source: "review_session", AttemptType: "review", PromptType: "self_check", Result: "correct", Rating: "good", AnsweredAt: f.now, ClientAttemptID: "duplicate-xp", IdempotencyKey: "duplicate-xp"}
	_, e = f.review.SubmitDailyReviewTx(t.Context(), tx, req, st.ID)
	require.NoError(t, e)
	require.NoError(t, tx.Commit())
	require.NoError(t, f.db.QueryRow("SELECT count(*),COALESCE(sum(amount),0) FROM confidence_point_ledger WHERE user_id=$1", f.u).Scan(&count, &balance))
	require.Equal(t, 1, count)
	require.Equal(t, gamification.RewardReviewGood, balance)
	// Negative control: a genuinely due standalone meaning without a verified
	// success today keeps the existing legacy Again rating award.
	legacyWord := uuid.New()
	var legacyMeaning uuid.UUID
	require.NoError(t, f.db.QueryRow(`SELECT id FROM word_meanings WHERE id<>$1 ORDER BY id LIMIT 1`, w.MeaningID).Scan(&legacyMeaning))
	_, e = f.db.Exec(`INSERT INTO user_words(id,user_id,meaning_id,status,source,review_step,added_at,created_at,updated_at)VALUES($1,$2,$3,'new','manual',0,$4,$4,$4)`, legacyWord, f.u, legacyMeaning, f.now)
	require.NoError(t, e)
	req.UserWordID = legacyWord
	req.MeaningID = legacyMeaning
	req.Source = "review"
	req.Result = "incorrect"
	req.Rating = "again"
	req.ClientAttemptID = "legacy-again"
	req.IdempotencyKey = "legacy-again"
	_, e = f.review.SubmitReview(t.Context(), req)
	require.NoError(t, e)
	require.NoError(t, f.db.QueryRow("SELECT count(*),COALESCE(sum(amount),0) FROM confidence_point_ledger WHERE user_id=$1", f.u).Scan(&count, &balance))
	require.Equal(t, 2, count)
	require.Equal(t, gamification.RewardReviewGood+gamification.RewardReviewAgain, balance)
}

// A verified success is authoritative for today's availability even when a
// legacy restore resets saved-word fields. Catalog Learn also retains a future
// schedule rather than undoing the successful review.
func TestPostgreSQLVerifiedTodayDueConsistencyAcrossRestorePaths(t *testing.T) {
	for _, restore := range []string{"known-learn", "catalog-soft-restore", "legacy-soft-restore"} {
		t.Run(restore, func(t *testing.T) {
			f := setup(t, 1)
			ctx := t.Context()
			// Make today's learner date differ from UTC so every projection must
			// use the saved timezone rather than a server-date shortcut.
			zone := "Etc/GMT+12"
			if f.now.Hour() >= 12 {
				zone = "Pacific/Kiritimati"
			}
			_, e := f.db.Exec(`UPDATE user_settings SET timezone=$2 WHERE user_id=$1`, f.u, zone)
			require.NoError(t, e)
			var meaning, word uuid.UUID
			require.NoError(t, f.db.QueryRow(`SELECT uw.meaning_id,wm.word_id FROM user_words uw JOIN word_meanings wm ON wm.id=uw.meaning_id WHERE uw.id=$1`, f.words[0]).Scan(&meaning, &word))
			_, e = f.db.Exec(`INSERT INTO knowledge_catalog_entries(rank,word_id,meaning_id,source_rank,source_count,source_sense_key,source_name,catalog_version)VALUES(1,$1,$2,1,1,'synthetic-due','synthetic','english-10000-v1')`, word, meaning)
			require.NoError(t, e)
			// Map reads deliberately require all 100 complete bands. Pad only this
			// isolated schema with synthetic content; authored choice pool stays six.
			_, e = f.db.Exec(`INSERT INTO canonical_words(id,text,normalized_text,status,created_at,updated_at)
 SELECT md5('due-word-'||n)::uuid,'due-padding-'||n,'due-padding-'||n,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP FROM generate_series(2,10000) n;
 INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at)
 SELECT md5('due-meaning-'||n)::uuid,md5('due-word-'||n)::uuid,'noun','Synthetic map padding',1,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP FROM generate_series(2,10000) n;
 INSERT INTO knowledge_catalog_entries(rank,word_id,meaning_id,source_rank,source_count,source_sense_key,source_name,catalog_version)
 SELECT n,md5('due-word-'||n)::uuid,md5('due-meaning-'||n)::uuid,n,1,'synthetic-'||n,'synthetic','english-10000-v1' FROM generate_series(2,10000) n`)
			require.NoError(t, e)
			learn := learning.NewPostgreSQLRepository(f.db)
			catalog := knowledgecatalog.NewService(knowledgecatalog.NewPostgreSQLRepository(f.db), clock.Fixed{T: f.now})
			item := func() knowledgecatalog.Item {
				t.Helper()
				m, err := catalog.Read(ctx, f.u, 1)
				require.NoError(t, err)
				return m.CurrentBand.Items[0]
			}
			check := func(want int) {
				t.Helper()
				summary, err := learn.GetKnowledgeSummary(ctx, f.u)
				require.NoError(t, err)
				require.Equal(t, want, summary.Due)
				all, err := learn.ListSavedWords(ctx, learning.ListSavedWordsRequest{UserID: f.u})
				require.NoError(t, err)
				require.Len(t, all.Items, 1)
				require.Equal(t, want == 1, all.Items[0].Due)
				due, err := learn.ListSavedWords(ctx, learning.ListSavedWordsRequest{UserID: f.u, DueOnly: true})
				require.NoError(t, err)
				require.Equal(t, want, due.TotalCount)
				require.Len(t, due.Items, want)
				states, err := learn.SavedWordStates(ctx, f.u, []uuid.UUID{meaning})
				require.NoError(t, err)
				require.Equal(t, want == 1, states[meaning].Due)
				queue, err := f.review.ListDueWords(ctx, reviews.ListDueWordsRequest{UserID: f.u})
				require.NoError(t, err)
				require.Equal(t, want, queue.TotalCount)
				require.Len(t, queue.Items, want)
				require.Equal(t, want == 1, item().Due)
				current, err := f.svc.Current(ctx, f.u)
				require.NoError(t, err)
				require.Equal(t, want, current.DueCount)
			}
			check(1)
			s, e := f.svc.Start(ctx, f.u, "start")
			require.NoError(t, e)
			s = f.answer(t, s, true)
			require.True(t, s.Feedback.Credited)
			f.next(t, s)
			check(0)
			var future time.Time
			require.NoError(t, f.db.QueryRow(`SELECT next_review_at FROM user_words WHERE id=$1`, f.words[0]).Scan(&future))
			require.True(t, future.After(f.now))
			if restore == "known-learn" {
				_, e = catalog.Decide(ctx, knowledgecatalog.DecisionRequest{UserID: f.u, MeaningID: meaning, Decision: "known", ExpectedRevision: item().Revision, IdempotencyKey: "known"})
				require.NoError(t, e)
			} else {
				require.NoError(t, learn.UnsaveUserWord(ctx, f.u, meaning, f.now))
			}
			if restore == "legacy-soft-restore" {
				_, _, e = learn.SaveUserWordAtomically(ctx, learning.SaveUserWordRequest{UserID: f.u, MeaningID: meaning, Source: "manual", IdempotencyKey: "restore"}, "restore", f.now)
				require.NoError(t, e)
			} else {
				_, e = catalog.Decide(ctx, knowledgecatalog.DecisionRequest{UserID: f.u, MeaningID: meaning, Decision: "learn", ExpectedRevision: item().Revision, IdempotencyKey: "learn"})
				require.NoError(t, e)
				var retained time.Time
				require.NoError(t, f.db.QueryRow(`SELECT next_review_at FROM user_words WHERE id=$1`, f.words[0]).Scan(&retained))
				require.True(t, future.Equal(retained), "Learn must retain the successful future schedule")
			}
			check(0)
			// Receipt exclusion is independent of potentially reset legacy fields.
			_, e = f.db.Exec(`UPDATE user_words SET next_review_at=NULL WHERE id=$1`, f.words[0])
			require.NoError(t, e)
			check(0)
			_, e = f.review.SubmitReview(ctx, reviews.SubmitReviewRequest{UserID: f.u, UserWordID: f.words[0], MeaningID: meaning, AttemptType: reviews.AttemptTypeReview, PromptType: reviews.PromptTypeSelfCheck, Result: reviews.ResultCorrect, Rating: reviews.RatingGood, AnsweredAt: f.now, Source: reviews.SourceReview, ClientAttemptID: "legacy-after-success", IdempotencyKey: "legacy-after-success"})
			require.ErrorIs(t, e, reviews.ErrNotDue)
			var attempts int
			require.NoError(t, f.db.QueryRow(`SELECT count(*) FROM review_attempts WHERE user_id=$1`, f.u).Scan(&attempts))
			require.Equal(t, 1, attempts)
			// Yesterday's receipt must not suppress today's genuinely due word.
			_, e = f.db.Exec(`UPDATE daily_review_successes SET local_date=local_date-1 WHERE user_id=$1`, f.u)
			require.NoError(t, e)
			check(1)
		})
	}
}
func TestPostgreSQLFreshSaveKeysLockUserBeforeForeignKeyClaims(t *testing.T) {
	f := setup(t, 0)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	var schema string
	require.NoError(t, f.db.QueryRow(`SHOW search_path`).Scan(&schema))
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		var e error
		dsn, e = pq.ParseURL(dsn)
		require.NoError(t, e)
	}
	application := "daily_save_lock_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	pool, e := sql.Open("postgres", dsn+" search_path="+schema+" application_name="+application)
	require.NoError(t, e)
	defer pool.Close()
	pool.SetMaxOpenConns(4)
	_, e = f.db.Exec(`CREATE SEQUENCE save_claim_observed;
 CREATE FUNCTION observe_save_claim() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM nextval('save_claim_observed'); RETURN NEW; END $$;
 CREATE TRIGGER observe_save_claim BEFORE INSERT ON idempotency_keys FOR EACH ROW EXECUTE FUNCTION observe_save_claim()`)
	require.NoError(t, e)
	var meaning uuid.UUID
	require.NoError(t, f.db.QueryRow(`SELECT id FROM word_meanings ORDER BY id LIMIT 1`).Scan(&meaning))
	blocker, e := f.db.BeginTx(ctx, nil)
	require.NoError(t, e)
	defer blocker.Rollback()
	_, e = blocker.ExecContext(ctx, `SELECT id FROM users WHERE id=$1 FOR UPDATE`, f.u)
	require.NoError(t, e)
	repo := learning.NewPostgreSQLRepository(pool)
	errs := make(chan error, 2)
	for _, key := range []string{"fresh-a", "fresh-b"} {
		go func(key string) {
			_, _, err := repo.SaveUserWordAtomically(ctx, learning.SaveUserWordRequest{UserID: f.u, MeaningID: meaning, Source: "manual", IdempotencyKey: key}, key, f.now)
			errs <- err
		}(key)
	}
	require.Eventually(t, func() bool {
		var waiting int
		err := pool.QueryRowContext(ctx, `SELECT count(*) FROM pg_stat_activity WHERE application_name=$1 AND wait_event_type='Lock'`, application).Scan(&waiting)
		return err == nil && waiting == 2
	}, 5*time.Second, 20*time.Millisecond, "both fresh saves must wait at the active-user boundary")
	// Sequences are not rolled back: this detects even uncommitted attempts to
	// INSERT an idempotency claim before acquiring the learner lock.
	var claimed bool
	require.NoError(t, pool.QueryRowContext(ctx, `SELECT is_called FROM save_claim_observed`).Scan(&claimed))
	require.False(t, claimed, "no FK-backed claim may precede the user lock")
	require.NoError(t, blocker.Commit())
	for range 2 {
		require.NoError(t, <-errs)
	}
	var words, keys int
	require.NoError(t, pool.QueryRowContext(ctx, `SELECT count(*) FROM user_words WHERE user_id=$1`, f.u).Scan(&words))
	require.NoError(t, pool.QueryRowContext(ctx, `SELECT count(*) FROM idempotency_keys WHERE user_id=$1`, f.u).Scan(&keys))
	require.Equal(t, 1, words)
	require.Equal(t, 2, keys)
}

func TestPostgreSQLDailyReviewNextDateAndKnowledgeAreIndependentOfGrade(t *testing.T) {
	for _, tc := range []struct {
		name, next string
		correct    bool
	}{
		{"correct later", "7d", true}, {"wrong later", "3d", false}, {"wrong known", "known", false}, {"correct known", "known", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := setup(t, 2)
			// Exercise PostgreSQL's microsecond precision against a real-clock-shaped value.
			f.svc.clock = clock.Fixed{T: f.now.Add(321 * time.Nanosecond)}
			s, err := f.svc.Start(t.Context(), f.u, "start")
			require.NoError(t, err)
			w := f.state(t, s).Words[0]
			_, err = f.db.Exec("INSERT INTO user_word_knowledge(user_id,meaning_id,self_reported_known,note,updated_at)VALUES($1,$2,false,'keep this note',$3)", f.u, w.MeaningID, f.now)
			require.NoError(t, err)
			s = f.answer(t, s, tc.correct)
			require.NotNil(t, s.Feedback.NextReviewAt)
			require.NotNil(t, s.Feedback.ReviewStep)
			var beforePoints int
			require.NoError(t, f.db.QueryRow("SELECT COALESCE(sum(amount),0) FROM confidence_point_ledger WHERE user_id=$1", f.u).Scan(&beforePoints))
			request := Action{StepID: s.CurrentStep.ID, ExpectedRevision: s.Revision, ClientActionID: uuid.NewString(), Action: "continue", NextReview: tc.next}
			key := uuid.NewString()
			out, err := f.svc.Act(t.Context(), f.u, uuid.MustParse(s.ID), request, key)
			require.NoError(t, err)
			require.Equal(t, 1, out.CompletedWords)
			require.NotEqual(t, s.CurrentStep.ID, out.CurrentStep.ID)
			var total, successes, points, step int
			var due time.Time
			require.NoError(t, f.db.QueryRow("SELECT total_review_count,review_step,next_review_at FROM user_words WHERE id=$1", w.UserWordID).Scan(&total, &step, &due))
			require.Equal(t, 1, total)
			require.Equal(t, *s.Feedback.ReviewStep, step)
			require.NoError(t, f.db.QueryRow("SELECT count(*) FROM daily_review_successes WHERE user_id=$1", f.u).Scan(&successes))
			if tc.correct {
				require.Equal(t, 1, successes)
			} else {
				require.Zero(t, successes)
			}
			require.NoError(t, f.db.QueryRow("SELECT COALESCE(sum(amount),0) FROM confidence_point_ledger WHERE user_id=$1", f.u).Scan(&points))
			require.Equal(t, beforePoints, points)
			var known bool
			var note string
			require.NoError(t, f.db.QueryRow("SELECT self_reported_known,note FROM user_word_knowledge WHERE user_id=$1 AND meaning_id=$2", f.u, w.MeaningID).Scan(&known, &note))
			require.Equal(t, "keep this note", note)
			require.Equal(t, tc.next == "known", known)
			if tc.next != "known" {
				days := 3
				if tc.next == "7d" {
					days = 7
				}
				require.WithinDuration(t, f.svc.clock.Now().Add(time.Duration(days)*24*time.Hour), due, time.Microsecond)
			}
			replay, err := f.svc.Act(t.Context(), f.u, uuid.MustParse(s.ID), request, key)
			require.NoError(t, err)
			require.Equal(t, out, replay)
			request.NextReview = "120d"
			_, err = f.svc.Act(t.Context(), f.u, uuid.MustParse(s.ID), request, key)
			require.ErrorIs(t, err, ErrConflict)
		})
	}
}

func TestPostgreSQLDailyReviewPreferenceCannotOverwriteAnotherTab(t *testing.T) {
	for _, change := range []string{"review_count", "schedule", "knowledge_note", "known", "removed", "status"} {
		t.Run(change, func(t *testing.T) {
			f := setup(t, 1)
			s, err := f.svc.Start(t.Context(), f.u, "start")
			require.NoError(t, err)
			w := f.state(t, s).Words[0]
			s = f.answer(t, s, true)
			switch change {
			case "review_count":
				_, err = f.db.Exec("UPDATE user_words SET total_review_count=total_review_count+1 WHERE id=$1", w.UserWordID)
			case "schedule":
				_, err = f.db.Exec("UPDATE user_words SET next_review_at=next_review_at+interval '1 hour' WHERE id=$1", w.UserWordID)
			case "knowledge_note":
				_, err = f.db.Exec("INSERT INTO user_word_knowledge(user_id,meaning_id,self_reported_known,note,updated_at)VALUES($1,$2,false,'new note',$3)", f.u, w.MeaningID, f.now.Add(time.Second))
			case "known":
				_, err = f.db.Exec("INSERT INTO user_word_knowledge(user_id,meaning_id,self_reported_known,note,updated_at)VALUES($1,$2,true,'new note',$3)", f.u, w.MeaningID, f.now.Add(time.Second))
			case "removed":
				_, err = f.db.Exec("UPDATE user_words SET deleted_at=$2 WHERE id=$1", w.UserWordID, f.now)
			case "status":
				_, err = f.db.Exec("UPDATE user_words SET status='archived' WHERE id=$1", w.UserWordID)
			}
			require.NoError(t, err)
			_, err = f.svc.Act(t.Context(), f.u, uuid.MustParse(s.ID), Action{StepID: s.CurrentStep.ID, ExpectedRevision: s.Revision, ClientActionID: uuid.NewString(), Action: "continue", NextReview: "7d"}, "schedule")
			require.ErrorIs(t, err, ErrConflict)
			reloaded, err := f.svc.Get(t.Context(), f.u, uuid.MustParse(s.ID))
			require.NoError(t, err)
			require.Equal(t, s.Revision, reloaded.Revision)
			require.Equal(t, s.CurrentStep, reloaded.CurrentStep)
			// Automatic Continue remains an escape from stale preference metadata.
			completed := f.next(t, reloaded)
			require.Equal(t, "completed", completed.Status)
		})
	}
}

func TestPostgreSQLDailyReviewLegacyFeedbackAndPreferenceValidation(t *testing.T) {
	f := setup(t, 1)
	s, err := f.svc.Start(t.Context(), f.u, "start")
	require.NoError(t, err)
	_, err = f.svc.Act(t.Context(), f.u, uuid.MustParse(s.ID), Action{StepID: s.CurrentStep.ID, ExpectedRevision: s.Revision, ClientActionID: uuid.NewString(), Action: "answer", ChoiceID: "missing", NextReview: "known"}, "invalid-answer")
	require.ErrorIs(t, err, ErrInvalid)
	s = f.answer(t, s, true)
	st := f.state(t, s)
	st.Feedback.NextReviewAt = nil
	st.Feedback.ReviewStep = nil
	data, err := json.Marshal(st)
	require.NoError(t, err)
	_, err = f.db.Exec("UPDATE daily_lesson_sessions SET state=$2 WHERE id=$1", s.ID, string(data))
	require.NoError(t, err)
	for _, value := range []string{"7d", "invalid"} {
		_, err = f.svc.Act(t.Context(), f.u, uuid.MustParse(s.ID), Action{StepID: s.CurrentStep.ID, ExpectedRevision: s.Revision, ClientActionID: uuid.NewString(), Action: "continue", NextReview: value}, "invalid-"+value)
		require.ErrorIs(t, err, ErrInvalid)
	}
	out := f.next(t, s)
	require.Equal(t, "completed", out.Status)
}

func TestDailyPronunciationProjectionUsesFrozenQueuedSense(t *testing.T) {
	noun, verb := uuid.NewString(), uuid.NewString()
	nounDefinition := "Information kept for later use."
	verbDefinition := strings.Repeat("To store information for later use. ", 24)
	st := &state{ID: uuid.New(), Words: []word{
		{MeaningID: noun, Definition: nounDefinition, Public: Step{ID: "noun", WordText: "record", PartOfSpeech: "noun", PronunciationMeaningID: "stale-public-identity", PronunciationRevision: "stale-public-revision"}},
		{MeaningID: verb, Definition: verbDefinition, Public: Step{ID: "verb", WordText: "record", PartOfSpeech: "verb", PronunciationMeaningID: "stale-public-identity", PronunciationRevision: "stale-public-revision"}},
	}, Queue: []int{1, 0}}
	first := project(st)
	require.Equal(t, verb, first.CurrentStep.PronunciationMeaningID, "queue identity, not first matching spelling")
	require.Equal(t, "verb", first.CurrentStep.PartOfSpeech)
	require.Equal(t, meaningidentity.Revision(verb, "record", "verb", verbDefinition), first.CurrentStep.PronunciationRevision)
	require.NotEqual(t, meaningidentity.Revision(verb, "record", "verb", verbDefinition[:500]), first.CurrentStep.PronunciationRevision)
	raw, err := json.Marshal(st)
	require.NoError(t, err)
	var resumed state
	require.NoError(t, json.Unmarshal(raw, &resumed))
	require.Equal(t, first.CurrentStep, project(&resumed).CurrentStep)
	require.Equal(t, "stale-public-revision", st.Words[1].Public.PronunciationRevision, "projection cannot rewrite persisted snapshots")
	resumed.Queue = []int{0}
	second := project(&resumed)
	require.Equal(t, noun, second.CurrentStep.PronunciationMeaningID)
	require.Equal(t, meaningidentity.Revision(noun, "record", "noun", nounDefinition), second.CurrentStep.PronunciationRevision)
	require.NotEqual(t, first.CurrentStep.PronunciationRevision, second.CurrentStep.PronunciationRevision)
}

func TestDailyPronunciationProjectionHistoricalIncompleteMetadata(t *testing.T) {
	for _, field := range []string{"meaning", "spelling", "pos", "definition"} {
		t.Run(field, func(t *testing.T) {
			w := word{MeaningID: uuid.NewString(), Definition: "A saved definition.", Public: Step{ID: "old", WordText: "record", PartOfSpeech: "verb", PronunciationRevision: "must-not-trust"}}
			switch field {
			case "meaning":
				w.MeaningID = ""
			case "spelling":
				w.Public.WordText = ""
			case "pos":
				w.Public.PartOfSpeech = ""
			case "definition":
				w.Definition = ""
			}
			st := &state{Words: []word{w}, Queue: []int{0}}
			current := project(st).CurrentStep
			require.NotNil(t, current)
			require.Empty(t, current.PronunciationRevision, "historical incomplete metadata must not invent a matching sense")
			require.Equal(t, w.MeaningID, current.PronunciationMeaningID)
		})
	}
	completed := time.Now()
	require.Nil(t, project(&state{CompletedAt: &completed}).CurrentStep)
}
