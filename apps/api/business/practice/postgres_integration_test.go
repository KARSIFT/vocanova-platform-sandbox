package practice

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
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

func TestPracticePostgreSQLPersonalJourneyAuthorityAndLocalDay(t *testing.T) {
	db := practiceDB(t)
	ctx := t.Context()
	u, other, broadWord, broadMeaning := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Date(2026, 10, 18, 20, 29, 0, 0, time.UTC)
	exec := func(q string, args ...any) {
		t.Helper()
		_, err := db.ExecContext(ctx, q, args...)
		require.NoError(t, err)
	}
	for _, id := range []uuid.UUID{u, other} {
		exec(`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',$3,$3)`, id, id.String()+"@personal-journey.invalid", now)
	}
	exec(`INSERT INTO user_settings(id,user_id,timezone,daily_review_target,created_at,updated_at)VALUES(gen_random_uuid(),$1,'Asia/Tehran',5,$2,$2)`, u, now)
	exec(`INSERT INTO canonical_words(id,text,normalized_text,language_code,status,created_at,updated_at)VALUES($1,'synthetic-broad-target','synthetic-broad-target','en','active',$2,$2)`, broadWord, now)
	exec(`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at)VALUES($1,$2,'noun','A distinct synthetic broad-catalog target meaning.',1,'active',$3,$3)`, broadMeaning, broadWord, now)
	exec(`INSERT INTO user_word_knowledge(user_id,meaning_id,self_reported_known,note,updated_at)VALUES($1,$2,true,'Preserve my private note',$3)`, u, broadMeaning, now)
	t.Cleanup(func() {
		for _, id := range []uuid.UUID{u, other} {
			for _, table := range []string{"practice_mistake_resolutions", "practice_actions", "practice_sessions", "user_word_knowledge", "user_settings"} {
				_, _ = db.ExecContext(context.Background(), "DELETE FROM "+table+" WHERE user_id=$1", id)
			}
			_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE id=$1", id)
		}
		_, _ = db.ExecContext(context.Background(), "DELETE FROM word_meanings WHERE id=$1", broadMeaning)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM canonical_words WHERE id=$1", broadWord)
	})
	repo := NewPostgreSQLRepository(db)
	svc := NewService(repo, clock.Fixed{T: now})
	before, err := svc.List(ctx, u)
	require.NoError(t, err)
	require.Equal(t, 1, before.Journey.TotalKnownWords)
	require.Equal(t, 3, before.Journey.DailyGoal)
	var wg sync.WaitGroup
	started := make(chan *Session, 4)
	errs := make(chan error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s, e := svc.Start(ctx, u, StartRequest{Mode: "journey"}, uuid.NewString())
			started <- s
			errs <- e
		}()
	}
	wg.Wait()
	close(started)
	close(errs)
	for e := range errs {
		require.NoError(t, e)
	}
	var current *Session
	for s := range started {
		if current == nil {
			current = s
		} else {
			require.Equal(t, current.ID, s.ID)
		}
	}
	require.Equal(t, 1, current.JourneyLesson)
	require.Zero(t, current.NewWords)
	require.Equal(t, 2, current.TotalSteps)
	sid := uuid.MustParse(current.ID)
	_, err = svc.Get(ctx, other, sid)
	require.ErrorIs(t, err, ErrNotFound)
	alias, err := svc.Start(ctx, u, StartRequest{Mode: "journey"}, "lost-start-response")
	require.NoError(t, err)
	require.Equal(t, sid.String(), alias.ID)
	for current.Status != "completed" {
		st, e := repo.Get(ctx, u, sid)
		require.NoError(t, e)
		step := st.Snapshot.Steps[st.Index]
		a := Action{StepID: step.Public.ID, ExpectedRevision: current.Revision, ClientActionID: uuid.NewString(), Action: "answer", ChoiceID: step.CorrectChoice}
		if step.Public.Kind == "teach" {
			a.Action = "continue"
			a.ChoiceID = ""
		}
		if step.Public.Kind == "sentence_build" {
			a.ChoiceID = ""
			a.TokenIDs = step.CorrectTokens
		}
		current, e = svc.Act(ctx, u, sid, a, a.ClientActionID)
		require.NoError(t, e)
		ack, e := svc.Act(ctx, u, sid, a, a.ClientActionID)
		require.NoError(t, e)
		require.Equal(t, current.Revision, ack.Revision)
		if a.Action == "continue" {
			continue
		}
		next := Action{StepID: step.Public.ID, ExpectedRevision: current.Revision, ClientActionID: uuid.NewString(), Action: "continue"}
		current, e = svc.Act(ctx, u, sid, next, next.ClientActionID)
		require.NoError(t, e)
	}
	list, err := svc.List(ctx, u)
	require.NoError(t, err)
	require.Equal(t, 1, list.Journey.CompletedLessons)
	require.Equal(t, 1, list.Journey.TodayCompletedLessons)
	require.Equal(t, 2, list.Journey.NextLesson)
	require.Nil(t, list.Journey.CurrentSession)
	// Tehran midnight, not UTC midnight, resets today's count. Total journey
	// completion and the saved completed session remain unchanged.
	svc.clock = clock.Fixed{T: now.Add(2 * time.Minute)}
	list, err = svc.List(ctx, u)
	require.NoError(t, err)
	require.Zero(t, list.Journey.TodayCompletedLessons)
	require.Equal(t, 1, list.Journey.CompletedLessons)
	replayed, err := svc.Start(ctx, u, StartRequest{Mode: "journey"}, "lost-start-response")
	require.NoError(t, err)
	require.Equal(t, "completed", replayed.Status)
	require.Equal(t, sid.String(), replayed.ID)
	for _, table := range []string{"user_words", "review_attempts", "daily_review_successes", "daily_mission_snapshots", "confidence_point_ledger"} {
		var n int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE user_id=$1", u).Scan(&n))
		require.Zero(t, n, table)
	}
	var note string
	require.NoError(t, db.QueryRowContext(ctx, `SELECT note FROM user_word_knowledge WHERE user_id=$1 AND meaning_id=$2`, u, broadMeaning).Scan(&note))
	require.Equal(t, "Preserve my private note", note)
	data, err := accounts.NewPostgreSQLRepository(db).ExportPersonalData(ctx, u)
	require.NoError(t, err)
	var exported map[string]any
	require.NoError(t, json.Unmarshal(data, &exported))
	var foundJourney bool
	for _, entry := range exported["practiceSessions"].([]any) {
		session := entry.(map[string]any)
		if session["id"] == sid.String() {
			require.Equal(t, float64(1), session["journeyLesson"])
			foundJourney = true
		}
	}
	require.True(t, foundJourney)
	require.NotContains(t, string(data), "CorrectTokens")
	require.NotContains(t, string(data), "lost-start-response")
	// A familiar authored word and a broad familiar word produce a genuinely
	// mixed lesson without displacing the learner's general vocabulary.
	greeting := catalog[0]
	exampleID := uuid.New()
	exec(`INSERT INTO user_word_knowledge(user_id,meaning_id,self_reported_known,note,updated_at)VALUES($1,$2,true,'',$3)`, u, greeting.MeaningID, now)
	exec(`INSERT INTO word_examples(id,meaning_id,example_text,example_order,status,created_at,updated_at)VALUES($1,$2,'A greeting can start a conversation.',0,'active',$3,$3)`, exampleID, greeting.MeaningID, now)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM word_examples WHERE id=$1`, exampleID)
	})
	current, err = svc.Start(ctx, u, StartRequest{Mode: "journey"}, "mixed-next")
	require.NoError(t, err)
	require.Equal(t, 2, current.JourneyLesson)
	require.Zero(t, current.NewWords)
	require.Equal(t, 5, current.TotalSteps)
	sid = uuid.MustParse(current.ID)
	sawBuild, sawGap := false, false
	for current.Status != "completed" {
		state, e := repo.Get(ctx, u, sid)
		require.NoError(t, e)
		step := state.Snapshot.Steps[state.Index]
		a := Action{StepID: step.Public.ID, ExpectedRevision: current.Revision, ClientActionID: uuid.NewString(), Action: "answer", ChoiceID: step.CorrectChoice}
		if step.Public.Kind == "gap_choice" {
			sawGap = true
			require.NotContains(t, step.Public.Context, "greeting")
		}
		if step.Public.Kind == "sentence_build" {
			sawBuild = true
			a.ChoiceID = ""
			a.TokenIDs = append([]string{}, step.CorrectTokens...)
			a.TokenIDs[0], a.TokenIDs[1] = a.TokenIDs[1], a.TokenIDs[0]
			current, e = svc.Act(ctx, u, sid, a, a.ClientActionID)
			require.NoError(t, e)
			require.False(t, current.Feedback.Correct)
			require.False(t, current.CanContinue)
			reloaded, e := svc.Get(ctx, u, sid)
			require.NoError(t, e)
			require.Equal(t, a.TokenIDs, reloaded.Feedback.SelectedTokenIDs)
			require.Equal(t, step.CorrectTokens, reloaded.Feedback.CorrectTokenIDs)
			a.ExpectedRevision = current.Revision
			a.ClientActionID = uuid.NewString()
			a.TokenIDs = step.CorrectTokens
		}
		current, e = svc.Act(ctx, u, sid, a, a.ClientActionID)
		require.NoError(t, e)
		require.True(t, current.Feedback.Correct)
		ack, e := svc.Act(ctx, u, sid, a, a.ClientActionID)
		require.NoError(t, e)
		require.Equal(t, current.Feedback, ack.Feedback)
		a = Action{StepID: step.Public.ID, ExpectedRevision: current.Revision, ClientActionID: uuid.NewString(), Action: "continue"}
		current, e = svc.Act(ctx, u, sid, a, a.ClientActionID)
		require.NoError(t, e)
	}
	require.True(t, sawBuild)
	require.True(t, sawGap)
	list, err = svc.List(ctx, u)
	require.NoError(t, err)
	require.Equal(t, 2, list.Journey.CompletedLessons)
	require.Equal(t, 1, list.Journey.TodayCompletedLessons)
	data, err = accounts.NewPostgreSQLRepository(db).ExportPersonalData(ctx, u)
	require.NoError(t, err)
	require.Contains(t, string(data), "tokenIds")
	require.NotContains(t, string(data), "CorrectTokens")
}

func practiceDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	schema := "practice_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = admin.ExecContext(t.Context(), "CREATE SCHEMA "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	db, err := sql.Open("postgres", practiceScopedDSN(t, dsn, schema))
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
		_, err := admin.ExecContext(context.Background(), "DROP SCHEMA "+pq.QuoteIdentifier(schema)+" CASCADE")
		require.NoError(t, err)
		admin.Close()
	})
	paths, err := filepath.Glob("../../migrations/*.sql")
	require.NoError(t, err)
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = db.ExecContext(t.Context(), string(raw))
		require.NoError(t, err, path)
	}
	seedPracticeCanonical(t, db)
	return db
}

func practiceScopedDSN(t *testing.T, dsn, schema string) string {
	t.Helper()
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		var err error
		dsn, err = pq.ParseURL(dsn)
		require.NoError(t, err)
	}
	return dsn + " search_path=" + schema
}

func seedPracticeCanonical(t *testing.T, db *sql.DB) {
	t.Helper()
	raw, err := os.ReadFile("../../cmd/seed/voc026-p1.json")
	require.NoError(t, err)
	var seed map[string][]map[string]any
	require.NoError(t, json.Unmarshal(raw, &seed))
	for _, table := range []string{"canonical_words", "word_meanings"} {
		for _, row := range seed[table] {
			row["created_at"] = "2026-10-02T12:00:00Z"
			row["updated_at"] = "2026-10-02T12:00:00Z"
			raw, e := json.Marshal(row)
			require.NoError(t, e)
			_, e = db.ExecContext(t.Context(), "INSERT INTO "+table+" SELECT * FROM jsonb_populate_record(NULL::"+table+",$1::jsonb) ON CONFLICT(id) DO NOTHING", string(raw))
			require.NoError(t, e)
		}
	}
}

func TestPracticePostgreSQLFixtureIsolatesExternalCanonicalChanges(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	externalSchema := "practice_external_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = admin.ExecContext(t.Context(), "CREATE SCHEMA "+pq.QuoteIdentifier(externalSchema))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.ExecContext(context.Background(), "DROP SCHEMA "+pq.QuoteIdentifier(externalSchema)+" CASCADE")
		require.NoError(t, err)
		admin.Close()
	})
	for _, table := range []string{"canonical_words", "word_meanings"} {
		_, err = admin.ExecContext(t.Context(), "CREATE TABLE "+pq.QuoteIdentifier(externalSchema)+"."+table+" (LIKE "+table+" INCLUDING ALL)")
		require.NoError(t, err)
	}
	externalDSN := practiceScopedDSN(t, dsn, externalSchema)
	external, err := sql.Open("postgres", externalDSN)
	require.NoError(t, err)
	t.Cleanup(func() { external.Close() })
	seedPracticeCanonical(t, external)
	// Reproduce another package's committed content-unavailability fixture and
	// keep it changed throughout this check, rather than hoping to win a race.
	_, err = external.ExecContext(t.Context(), `UPDATE word_meanings SET status='archived' WHERE id=$1`, catalog[0].MeaningID)
	require.NoError(t, err)
	t.Setenv("VOCANOVA_TEST_POSTGRES_DSN", externalDSN)
	fixture := practiceDB(t)
	tx, err := fixture.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	defer tx.Rollback()
	words, err := loadWords(t.Context(), tx)
	require.NoError(t, err, "practice fixtures must not read another package's temporary canonical changes")
	require.Len(t, words, len(catalog))
	var status string
	require.NoError(t, external.QueryRowContext(t.Context(), `SELECT status FROM word_meanings WHERE id=$1`, catalog[0].MeaningID).Scan(&status))
	require.Equal(t, "archived", status, "fixture setup must preserve external data")
}
func TestPracticePostgreSQLAuthorityReplayMistakesExportPurge(t *testing.T) {
	db := practiceDB(t)
	ctx := t.Context()
	u, other := uuid.New(), uuid.New()
	exec := func(q string, args ...any) {
		t.Helper()
		_, err := db.ExecContext(ctx, q, args...)
		require.NoError(t, err)
	}
	for _, id := range []uuid.UUID{u, other} {
		exec(`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, id, id.String()+"@practice.invalid")
	}
	t.Cleanup(func() {
		for _, id := range []uuid.UUID{u, other} {
			for _, table := range []string{"practice_mistake_resolutions", "practice_actions", "practice_sessions", "lesson_actions", "lesson_sessions", "review_attempts", "user_words"} {
				_, _ = db.ExecContext(context.Background(), "DELETE FROM "+table+" WHERE user_id=$1", id)
			}
			_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE id=$1", id)
		}
	})
	repo := NewPostgreSQLRepository(db)
	svc := NewService(repo, nil)
	req := StartRequest{Mode: "typed_recall", LessonKey: "airport"}
	var wg sync.WaitGroup
	results := make(chan *Session, 8)
	errs := make(chan error, 8)
	for range 8 {
		wg.Add(1)
		go func() { defer wg.Done(); s, e := svc.Start(ctx, u, req, "same-start"); results <- s; errs <- e }()
	}
	wg.Wait()
	close(results)
	close(errs)
	for e := range errs {
		require.NoError(t, e)
	}
	var session *Session
	for s := range results {
		if session != nil {
			require.Equal(t, session.ID, s.ID)
		}
		session = s
	}
	id := uuid.MustParse(session.ID)
	_, err := svc.Get(ctx, other, id)
	require.ErrorIs(t, err, ErrNotFound)
	_, err = svc.Start(ctx, u, StartRequest{Mode: "listening_choice"}, "same-start")
	require.ErrorIs(t, err, ErrConflict)
	raw, err := json.Marshal(session)
	require.NoError(t, err)
	require.NotContains(t, string(raw), `"wordText"`)
	require.NotContains(t, string(raw), "CorrectChoice")
	state, err := repo.Get(ctx, u, id)
	require.NoError(t, err)
	first := state.Snapshot.Steps[0]
	wrong := Action{StepID: first.Public.ID, Action: "answer", TypedAnswer: "private wrong answer", ClientActionID: "wrong"}
	actionResults := make(chan *Session, 8)
	actionErrors := make(chan error, 8)
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, e := svc.Act(ctx, u, id, wrong, "wrong")
			actionResults <- out
			actionErrors <- e
		}()
	}
	wg.Wait()
	close(actionResults)
	close(actionErrors)
	for e := range actionErrors {
		require.NoError(t, e)
	}
	for out := range actionResults {
		require.Equal(t, 1, out.Revision)
		session = out
	}
	var receiptCount int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM practice_actions WHERE user_id=$1 AND operation='action'`, u).Scan(&receiptCount))
	require.Equal(t, 1, receiptCount)
	_, err = svc.Act(ctx, other, id, wrong, "foreign-action")
	require.ErrorIs(t, err, ErrNotFound)
	stale := wrong
	stale.ClientActionID = "stale"
	_, err = svc.Act(ctx, u, id, stale, "stale")
	require.ErrorIs(t, err, ErrConflict)
	require.False(t, session.Feedback.Correct)
	replayed, err := svc.Act(ctx, u, id, wrong, "wrong")
	require.NoError(t, err)
	require.Equal(t, session.Revision, replayed.Revision)
	list, err := svc.List(ctx, u)
	require.NoError(t, err)
	require.Equal(t, 1, list.AvailableMistakes)
	mistake, err := svc.Start(ctx, u, StartRequest{Mode: "mistakes"}, "mistake-old")
	require.NoError(t, err)
	require.Equal(t, 1, mistake.TotalSteps)
	// Create a newer error after the mistake session captured its exact source.
	newer, err := svc.Start(ctx, u, req, "newer-session")
	require.NoError(t, err)
	newID := uuid.MustParse(newer.ID)
	newState, err := repo.Get(ctx, u, newID)
	require.NoError(t, err)
	for newer.CurrentStep != nil {
		step := newState.Snapshot.Steps[newer.CompletedSteps]
		answer := step.Word.WordText
		if step.Word.MeaningID == first.Word.MeaningID {
			answer = "new wrong answer"
		}
		a := Action{StepID: step.Public.ID, ExpectedRevision: newer.Revision, ClientActionID: uuid.NewString(), Action: "answer", TypedAnswer: answer}
		newer, err = svc.Act(ctx, u, newID, a, a.ClientActionID)
		require.NoError(t, err)
		a = Action{StepID: step.Public.ID, ExpectedRevision: newer.Revision, ClientActionID: uuid.NewString(), Action: "continue"}
		newer, err = svc.Act(ctx, u, newID, a, a.ClientActionID)
		require.NoError(t, err)
	}
	require.Equal(t, "completed", newer.Status)
	mid := uuid.MustParse(mistake.ID)
	a := Action{StepID: mistake.CurrentStep.ID, ExpectedRevision: mistake.Revision, ClientActionID: "resolve-old", Action: "answer", TypedAnswer: first.Word.WordText}
	_, err = svc.Act(ctx, u, mid, a, "resolve-old")
	require.NoError(t, err)
	list, err = svc.List(ctx, u)
	require.NoError(t, err)
	require.Equal(t, 1, list.AvailableMistakes, "old session must not resolve newer error")
	latest, err := svc.Start(ctx, u, StartRequest{Mode: "mistakes"}, "mistake-latest")
	require.NoError(t, err)
	a = Action{StepID: latest.CurrentStep.ID, ClientActionID: "resolve-latest", Action: "answer", TypedAnswer: first.Word.WordText}
	_, err = svc.Act(ctx, u, uuid.MustParse(latest.ID), a, "resolve-latest")
	require.NoError(t, err)
	list, err = svc.List(ctx, u)
	require.NoError(t, err)
	require.Zero(t, list.AvailableMistakes)
	_, err = svc.Start(ctx, u, StartRequest{Mode: "mistakes"}, "empty")
	require.ErrorIs(t, err, ErrNoMistakes)
	again, err := svc.Start(ctx, u, req, "repeat-after-completion")
	require.NoError(t, err)
	require.NotEqual(t, newer.ID, again.ID)
	// Snapshot survives canonical edits; a failed receipt cannot advance progress.
	originalDefinition := first.Word.Definition
	exec(`UPDATE word_meanings SET short_definition='Changed after snapshot' WHERE id=$1`, first.Word.MeaningID)
	resumed, err := NewService(repo, nil).Get(ctx, u, id)
	exec(`UPDATE word_meanings SET short_definition=$2 WHERE id=$1`, first.Word.MeaningID, originalDefinition)
	require.NoError(t, err)
	require.Equal(t, session.CurrentStep.Prompt, resumed.CurrentStep.Prompt)
	exec(`ALTER TABLE practice_actions ADD CONSTRAINT practice_test_reject CHECK(idempotency_key <> 'force-rollback') NOT VALID`)
	a = Action{StepID: again.CurrentStep.ID, ClientActionID: "force-rollback", Action: "reveal"}
	_, err = svc.Act(ctx, u, uuid.MustParse(again.ID), a, "force-rollback")
	exec(`ALTER TABLE practice_actions DROP CONSTRAINT practice_test_reject`)
	require.Error(t, err)
	resumed, err = svc.Get(ctx, u, uuid.MustParse(again.ID))
	require.NoError(t, err)
	require.Zero(t, resumed.Revision)
	require.Nil(t, resumed.Feedback)
	foreign, err := svc.Start(ctx, other, StartRequest{Mode: "listening_choice", LessonKey: "airport"}, "other")
	require.NoError(t, err)
	require.NotEmpty(t, foreign.CurrentStep.SpeechText)
	for _, table := range []string{"user_words", "review_attempts", "daily_mission_snapshots", "confidence_point_ledger"} {
		var n int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE user_id=$1", u).Scan(&n))
		require.Zero(t, n)
	}
	ar := accounts.NewPostgreSQLRepository(db)
	data, err := ar.ExportPersonalData(ctx, u)
	require.NoError(t, err)
	require.Contains(t, string(data), "private wrong answer")
	require.NotContains(t, string(data), "CorrectChoice")
	require.NotContains(t, string(data), "same-start")
	require.NotContains(t, string(data), foreign.ID)
	exec(`UPDATE users SET status='deleted',deleted_at=CURRENT_TIMESTAMP WHERE id=$1`, u)
	_, err = svc.Act(ctx, u, id, wrong, "wrong")
	require.ErrorIs(t, err, ErrNotFound)
	counts, err := ar.AnonymizeUserData(ctx, u)
	require.NoError(t, err)
	require.Positive(t, counts.PracticeSessions)
	require.Positive(t, counts.PracticeActions)
	require.Equal(t, int64(2), counts.PracticeMistakeResolutions)
	for _, table := range []string{"practice_sessions", "practice_actions", "practice_mistake_resolutions"} {
		var n int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE user_id=$1", u).Scan(&n))
		require.Zero(t, n)
	}
	_, err = svc.Get(ctx, other, uuid.MustParse(foreign.ID))
	require.NoError(t, err)
}
func TestPracticePostgreSQLHistoricalMistakeSources(t *testing.T) {
	db := practiceDB(t)
	ctx := t.Context()
	u, other := uuid.New(), uuid.New()
	now := time.Now().UTC()
	exec := func(q string, args ...any) {
		t.Helper()
		_, e := db.ExecContext(ctx, q, args...)
		require.NoError(t, e)
	}
	for _, id := range []uuid.UUID{u, other} {
		exec(`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',$3,$3)`, id, id.String()+"@practice.invalid", now)
	}
	t.Cleanup(func() {
		for _, id := range []uuid.UUID{u, other} {
			for _, table := range []string{"practice_mistake_resolutions", "practice_actions", "practice_sessions", "lesson_actions", "lesson_sessions", "review_attempts", "user_words"} {
				_, _ = db.ExecContext(context.Background(), "DELETE FROM "+table+" WHERE user_id=$1", id)
			}
			_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE id=$1", id)
		}
	})
	sid := uuid.New()
	exec(`INSERT INTO lesson_sessions(id,user_id,lesson_key,lesson_version,snapshot,total_steps,created_at,updated_at)VALUES($1,$2,'airport','1','{}',9,$3,$3)`, sid, u, now)
	// These historical fixtures belong to Airport, independently of course order.
	boardingPass := "8a6ec801-3292-5fc2-bfb3-242ac405d891"
	feedback, _ := json.Marshal(map[string]any{"feedback": map[string]any{"correct": false, "correctChoiceId": boardingPass}})
	exec(`INSERT INTO lesson_actions(id,session_id,user_id,operation,idempotency_key,client_action_id,fingerprint,action,result,created_at)VALUES($1,$2,$3,'action','wrong','wrong',repeat('a',64),'{"action":"answer"}',$4,$5)`, uuid.New(), sid, u, string(feedback), now)
	for i, prompt := range []string{"multiple_choice", "self_check", "multiple_choice"} {
		user := u
		if i == 2 {
			user = other
		}
		word := uuid.New()
		meaning := []string{"6a61d337-80e7-5d50-bc61-8b45b6a054a1", "bee2408e-5991-5805-923c-6fba33ab7e7a", "329c4ec9-6562-568e-9e71-9134da2f0d8d"}[i]
		exec(`INSERT INTO user_words(id,user_id,meaning_id,status,source,added_at,created_at,updated_at)VALUES($1,$2,$3,'learning','manual',$4,$4,$4)`, word, user, meaning, now)
		exec(`INSERT INTO review_attempts(id,user_id,user_word_id,meaning_id,attempt_type,prompt_type,result,rating,review_step_before,review_step_after,answered_at,source,created_at,updated_at)VALUES($1,$2,$3,$4,'scheduled',$5,'incorrect','again',0,0,$6,'review',$6,$6)`, uuid.New(), user, word, meaning, prompt, now)
	}
	svc := NewService(NewPostgreSQLRepository(db), nil)
	list, err := svc.List(ctx, u)
	require.NoError(t, err)
	require.Equal(t, 2, list.AvailableMistakes, "only own objectively graded errors")
	session, err := svc.Start(ctx, u, StartRequest{Mode: "mistakes", LessonKey: "airport"}, "historical")
	require.NoError(t, err)
	require.Equal(t, 2, session.TotalSteps)
	require.Equal(t, "typed_recall", session.CurrentStep.Kind)
}
