package practice

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// Daily review records graded answers as self-check attempts. Only a miss with
// its exact owned effort receipt may become an optional practice mistake.
func TestPracticePostgreSQLDailyMissProvenanceRankedFallbackAndResolution(t *testing.T) {
	db := practiceDB(t)
	ctx := t.Context()
	u, other := uuid.New(), uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	exec := func(q string, args ...any) {
		t.Helper()
		_, err := db.ExecContext(ctx, q, args...)
		require.NoError(t, err)
	}
	for _, id := range []uuid.UUID{u, other} {
		exec(`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',$3,$3)`, id, id.String()+"@practice.invalid", now)
	}
	t.Cleanup(func() {
		for _, id := range []uuid.UUID{u, other} {
			for _, table := range []string{"practice_mistake_resolutions", "practice_actions", "practice_sessions", "daily_review_effort", "review_attempts", "daily_lesson_sessions", "user_words"} {
				_, _ = db.ExecContext(context.Background(), "DELETE FROM "+table+" WHERE user_id=$1", id)
			}
			_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE id=$1", id)
		}
	})
	type target struct{ word, meaning, userWord uuid.UUID }
	ranked := func(text, definition string) target {
		tg := target{uuid.New(), uuid.New(), uuid.New()}
		exec(`INSERT INTO canonical_words(id,text,normalized_text,language_code,status,created_at,updated_at)VALUES($1,$2,$2,'en','active',$3,$3)`, tg.word, text, now)
		exec(`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at)VALUES($1,$2,'noun',$3,1,'active',$4,$4)`, tg.meaning, tg.word, definition, now)
		return tg
	}
	starter := func(meaning string) target {
		tg := target{meaning: uuid.MustParse(meaning), userWord: uuid.New()}
		require.NoError(t, db.QueryRowContext(ctx, `SELECT word_id FROM word_meanings WHERE id=$1`, tg.meaning).Scan(&tg.word))
		return tg
	}
	harbor := ranked("harborx", "a sheltered area of water where ships can stay safely")
	gate := starter("6a61d337-80e7-5d50-bc61-8b45b6a054a1")
	luggage := starter("bee2408e-5991-5805-923c-6fba33ab7e7a")
	hinted := ranked("quayx", "a stone platform beside water for loading boats")
	skipped := ranked("ferryx", "a boat that carries people across water")
	unfrozen := ranked("anchorx", "a heavy object that keeps a boat in place")
	selfSpelling := ranked("tidex", "the tidex that rises and falls each day")
	all := []target{harbor, gate, luggage, hinted, skipped, unfrozen, selfSpelling}
	for _, tg := range all {
		exec(`INSERT INTO user_words(id,user_id,meaning_id,status,source,added_at,created_at,updated_at)VALUES($1,$2,$3,'learning','search',$4,$4,$4)`, tg.userWord, u, tg.meaning, now)
	}
	frozen := func(targets ...target) string {
		words := []map[string]string{}
		for _, tg := range targets {
			words = append(words, map[string]string{"MeaningID": tg.meaning.String(), "UserWordID": tg.userWord.String()})
		}
		raw, err := json.Marshal(map[string]any{"Words": words})
		require.NoError(t, err)
		return string(raw)
	}
	session, otherSession := uuid.New(), uuid.New()
	exec(`INSERT INTO daily_lesson_sessions(id,user_id,state,status,revision,created_at,updated_at,completed_at)VALUES($1,$2,$3,'completed',4,$4,$4,$4)`, session, u, frozen(harbor, gate, luggage, hinted, skipped, selfSpelling), now)
	exec(`INSERT INTO daily_lesson_sessions(id,user_id,state,status,revision,created_at,updated_at,completed_at)VALUES($1,$2,$3,'completed',1,$4,$4,$4)`, otherSession, u, frozen(harbor), now)
	day := now.Format("2006-01-02")
	attempt := func(tg target, result string, hint bool, at time.Time) uuid.UUID {
		id := uuid.New()
		rating := any("again")
		if result == "skipped" {
			rating = nil
		}
		exec(`INSERT INTO review_attempts(id,user_id,user_word_id,meaning_id,attempt_type,prompt_type,result,rating,review_step_before,review_step_after,answered_at,was_hint_used,source,created_at,updated_at)VALUES($1,$2,$3,$4,'review','self_check',$5,$6,0,0,$7,$8,'review_session',$7,$7)`, id, u, tg.userWord, tg.meaning, result, rating, at, hint)
		return id
	}
	effort := func(tg target, attemptID, sessionID uuid.UUID, localDate string, at time.Time) {
		exec(`INSERT INTO daily_review_effort(user_id,word_id,meaning_id,local_date,review_attempt_id,daily_session_id,created_at)VALUES($1,$2,$3,$4,$5,$6,$7)`, u, tg.word, tg.meaning, localDate, attemptID, sessionID, at)
	}
	harborMiss := attempt(harbor, "incorrect", false, now.Add(-5*time.Minute))
	effort(harbor, harborMiss, session, day, now.Add(-5*time.Minute))
	gateMiss := attempt(gate, "incorrect", false, now.Add(-4*time.Minute))
	effort(gate, gateMiss, session, day, now.Add(-4*time.Minute))
	// A standalone self-rating has the same prompt type but no effort receipt.
	attempt(luggage, "incorrect", false, now.Add(-3*time.Minute))
	// Defense in depth: assisted and skipped answers never count, even if a
	// receipt were present.
	hintedMiss := attempt(hinted, "incorrect", true, now.Add(-3*time.Minute))
	effort(hinted, hintedMiss, session, day, now.Add(-3*time.Minute))
	skip := attempt(skipped, "skipped", false, now.Add(-2*time.Minute))
	effort(skipped, skip, session, day, now.Add(-2*time.Minute))
	// The receipt's session must have frozen this exact user word.
	unfrozenMiss := attempt(unfrozen, "incorrect", false, now.Add(-2*time.Minute))
	effort(unfrozen, unfrozenMiss, otherSession, day, now.Add(-2*time.Minute))
	// A definition that spells its own headword is not a fair recognition target.
	selfMiss := attempt(selfSpelling, "incorrect", false, now.Add(-time.Minute))
	effort(selfSpelling, selfMiss, session, day, now.Add(-time.Minute))

	repo := NewPostgreSQLRepository(db)
	svc := NewService(repo, nil)
	list, err := svc.List(ctx, u)
	require.NoError(t, err)
	require.Equal(t, 2, list.AvailableMistakes, "only the provenance-verified, practisable daily misses")
	foreign, err := svc.List(ctx, other)
	require.NoError(t, err)
	require.Zero(t, foreign.AvailableMistakes)

	started, err := svc.Start(ctx, u, StartRequest{Mode: "mistakes"}, "daily-mistakes")
	require.NoError(t, err)
	require.Equal(t, 2, started.TotalSteps)
	require.Equal(t, MistakesContentVersion, started.ContentVersion)
	replayed, err := svc.Start(ctx, u, StartRequest{Mode: "mistakes"}, "daily-mistakes")
	require.NoError(t, err)
	require.Equal(t, started.ID, replayed.ID)
	id := uuid.MustParse(started.ID)
	state, err := repo.Get(ctx, u, id)
	require.NoError(t, err)
	kinds := map[string]string{}
	for _, step := range state.Snapshot.Steps {
		require.NotNil(t, step.Source)
		require.Equal(t, "review", step.Source.Kind)
		kinds[step.Word.MeaningID] = step.Public.Kind
	}
	require.Equal(t, "meaning_choice", kinds[harbor.meaning.String()])
	require.Equal(t, "typed_recall", kinds[gate.meaning.String()])

	session0 := started
	for session0.CurrentStep != nil {
		step := state.Snapshot.Steps[session0.CompletedSteps]
		a := Action{StepID: step.Public.ID, ExpectedRevision: session0.Revision, ClientActionID: uuid.NewString(), Action: "answer"}
		if step.Public.Kind == "meaning_choice" {
			a.ChoiceID = step.CorrectChoice
		} else {
			a.TypedAnswer = step.Word.WordText
		}
		session0, err = svc.Act(ctx, u, id, a, a.ClientActionID)
		require.NoError(t, err)
		require.True(t, session0.Feedback.Correct)
		a = Action{StepID: step.Public.ID, ExpectedRevision: session0.Revision, ClientActionID: uuid.NewString(), Action: "continue"}
		session0, err = svc.Act(ctx, u, id, a, a.ClientActionID)
		require.NoError(t, err)
	}
	var resolved int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM practice_mistake_resolutions WHERE user_id=$1 AND source_kind='review' AND source_id=ANY($2::uuid[])`, u, pq.Array([]string{harborMiss.String(), gateMiss.String()})).Scan(&resolved))
	require.Equal(t, 2, resolved)
	list, err = svc.List(ctx, u)
	require.NoError(t, err)
	require.Zero(t, list.AvailableMistakes)

	// A newer daily miss outranks the resolved one and makes the word return.
	tomorrow := now.Add(24 * time.Hour)
	newer := attempt(harbor, "incorrect", false, tomorrow)
	effort(harbor, newer, session, tomorrow.Format("2006-01-02"), tomorrow)
	list, err = svc.List(ctx, u)
	require.NoError(t, err)
	require.Equal(t, 1, list.AvailableMistakes)

	t.Run("older success preserves a newer ranked practice miss", func(t *testing.T) {
		future := &clock.Fixed{T: tomorrow.Add(time.Minute)}
		svc := NewService(repo, future)
		start := func(key string) (*Session, privateStep) {
			t.Helper()
			session, err := svc.Start(ctx, u, StartRequest{Mode: "mistakes"}, key)
			require.NoError(t, err)
			require.Equal(t, 1, session.TotalSteps)
			state, err := repo.Get(ctx, u, uuid.MustParse(session.ID))
			require.NoError(t, err)
			return session, state.Snapshot.Steps[0]
		}
		answer := func(session *Session, choice string) *Session {
			t.Helper()
			future.Advance(time.Minute)
			action := Action{StepID: session.CurrentStep.ID, ExpectedRevision: session.Revision, ClientActionID: uuid.NewString(), Action: "answer", ChoiceID: choice}
			answered, err := svc.Act(ctx, u, uuid.MustParse(session.ID), action, action.ClientActionID)
			require.NoError(t, err)
			return answered
		}
		older, oldStep := start("older-snapshot")
		parallel, parallelStep := start("parallel-snapshot")
		require.Equal(t, oldStep.Source, parallelStep.Source)
		wrong := ""
		for _, option := range parallelStep.Public.Choices {
			if option.ID != parallelStep.CorrectChoice {
				wrong = option.ID
				break
			}
		}
		require.NotEmpty(t, wrong)
		missed := answer(parallel, wrong)
		require.False(t, missed.Feedback.Correct)
		require.False(t, missed.Feedback.Assisted)
		require.True(t, answer(older, oldStep.CorrectChoice).Feedback.Correct)
		list, err := svc.List(ctx, u)
		require.NoError(t, err)
		require.Equal(t, 1, list.AvailableMistakes, "the newer practice miss remains available")
		fresh, freshStep := start("fresh-snapshot")
		require.Equal(t, "practice", freshStep.Source.Kind)
		// An assisted retry belongs to the same practice opportunity.
		retry := answer(missed, wrong)
		require.True(t, retry.Feedback.Assisted)
		require.True(t, answer(fresh, freshStep.CorrectChoice).Feedback.Correct)
		list, err = svc.List(ctx, u)
		require.NoError(t, err)
		require.Zero(t, list.AvailableMistakes)
	})

	// Saved progress survives a restart of the service.
	resumed, err := NewService(repo, nil).Get(ctx, u, id)
	require.NoError(t, err)
	require.Equal(t, "completed", resumed.Status)
}
