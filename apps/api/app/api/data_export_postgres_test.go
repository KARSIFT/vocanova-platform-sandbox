package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/dailylesson"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/lessons"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/missions"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/practice"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/reviews"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/stories"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/wordlists"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestPersonalDataExportHTTPPostgreSQLLearningHistory(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	db, e := sql.Open("postgres", dsn)
	require.NoError(t, e)
	t.Cleanup(func() { db.Close() })
	ctx := t.Context()
	owner, other := uuid.New(), uuid.New()
	for _, u := range []uuid.UUID{owner, other} {
		_, e = db.ExecContext(ctx, `INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, u, u.String()+"@http-export.invalid")
		require.NoError(t, e)
	}
	t.Cleanup(func() {
		for _, u := range []uuid.UUID{owner, other} {
			_, e := accounts.NewPostgreSQLRepository(db).AnonymizeUserData(context.Background(), u)
			require.NoError(t, e)
			_, e = db.ExecContext(context.Background(), `DELETE FROM users WHERE id=$1`, u)
			require.NoError(t, e)
		}
	})
	meaning := uuid.MustParse("329c4ec9-6562-568e-9e71-9134da2f0d8d")
	lists := wordlists.NewService(wordlists.NewPostgreSQLRepository(db), nil)
	list, e := lists.Write(ctx, wordlists.WriteRequest{UserID: owner, ListID: uuid.New(), Operation: "put", Name: "My exported list", IdempotencyKey: "private-list-create"})
	require.NoError(t, e)
	list, e = lists.Write(ctx, wordlists.WriteRequest{UserID: owner, ListID: uuid.MustParse(list.ID), MeaningID: meaning, Operation: "add", ExpectedRevision: list.Revision, IdempotencyKey: "private-list-member"})
	require.NoError(t, e)
	psvc := practice.NewService(practice.NewPostgreSQLRepository(db), nil)
	p, e := psvc.Start(ctx, owner, practice.StartRequest{Mode: "typed_recall", ListID: list.ID, ListRevision: &list.Revision}, "private-practice-start")
	require.NoError(t, e)
	_, e = psvc.Act(ctx, owner, uuid.MustParse(p.ID), practice.Action{StepID: p.CurrentStep.ID, Action: "answer", TypedAnswer: "Original practice answer", ExpectedRevision: p.Revision, ClientActionID: "private-practice-client"}, "private-practice-request")
	require.NoError(t, e)
	ssvc := stories.NewService(stories.NewPostgreSQLRepository(db), nil)
	story, e := ssvc.Start(ctx, owner, stories.StoryStartRequest{StoryKey: "a-quiet-lunch"}, "private-story-start")
	require.NoError(t, e)
	_, e = ssvc.Act(ctx, owner, uuid.MustParse(story.ID), stories.StoryAction{StepID: story.CurrentStep.ID, Action: "continue", ExpectedRevision: story.Revision, ClientActionID: "private-story-client"}, "private-story-request")
	require.NoError(t, e)
	foreign, e := ssvc.Start(ctx, other, stories.StoryStartRequest{StoryKey: "a-quiet-lunch"}, "private-foreign-start")
	require.NoError(t, e)
	lsvc := lessons.NewService(lessons.NewPostgreSQLRepository(db), nil)
	lesson, e := lsvc.Start(ctx, owner, "conversation-start", "private-lesson-start")
	require.NoError(t, e)
	for lesson.CurrentStep.Kind != "typed_recall" {
		action := lessons.Action{StepID: lesson.CurrentStep.ID, ExpectedRevision: lesson.Revision, ClientActionID: uuid.NewString(), Action: "continue"}
		if !lesson.CanContinue {
			action.Action = "answer"
			action.ChoiceID = lesson.CurrentStep.Word.MeaningID
		}
		lesson, e = lsvc.Act(ctx, owner, uuid.MustParse(lesson.ID), action, action.ClientActionID)
		require.NoError(t, e)
	}
	_, e = lsvc.Act(ctx, owner, uuid.MustParse(lesson.ID), lessons.Action{StepID: lesson.CurrentStep.ID, Action: "answer", TypedAnswer: "Original guided answer", ExpectedRevision: lesson.Revision, ClientActionID: "private-lesson-client"}, "private-lesson-request")
	require.NoError(t, e)
	_, e = db.ExecContext(ctx, `INSERT INTO user_word_knowledge(user_id,meaning_id,self_reported_known,note,updated_at)VALUES($1,$2,true,'My private learner note',CURRENT_TIMESTAMP)`, owner, meaning)
	require.NoError(t, e)
	_, e = db.ExecContext(ctx, `INSERT INTO user_learning_preferences(id,user_id,learning_goal,main_use_case,revision,created_at,updated_at)VALUES($2,$1,'travel','travel',1,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, owner, uuid.New())
	require.NoError(t, e)
	// Exercise the real personal journey, including checked sentence token
	// ordering and an explicit assisted listening action. The private answer
	// snapshot is used only by this test; it must never reach the HTTP export.
	journey, e := psvc.Start(ctx, owner, practice.StartRequest{Mode: "journey"}, "private-journey-start")
	require.NoError(t, e)
	require.Equal(t, 1, journey.JourneyLesson)
	require.Zero(t, journey.NewWords)
	sawSentenceBuild := false
	for journey.Status != "completed" {
		step := journey.CurrentStep
		action := practice.Action{StepID: step.ID, ExpectedRevision: journey.Revision, ClientActionID: "private-journey-" + uuid.NewString(), Action: "answer"}
		if step.Kind == "listening_choice" {
			action.Action = "reveal"
		} else {
			var correct string
			var tokenJSON []byte
			e = db.QueryRowContext(ctx, `SELECT step->>'CorrectChoice',step->'CorrectTokens' FROM practice_sessions s CROSS JOIN LATERAL jsonb_array_elements(s.snapshot->'Steps') step WHERE s.id=$1 AND step->'Public'->>'id'=$2`, journey.ID, step.ID).Scan(&correct, &tokenJSON)
			require.NoError(t, e)
			if step.Kind == "sentence_build" {
				sawSentenceBuild = true
				require.NoError(t, json.Unmarshal(tokenJSON, &action.TokenIDs))
				require.NotEmpty(t, action.TokenIDs)
			} else {
				action.ChoiceID = correct
			}
		}
		journey, e = psvc.Act(ctx, owner, uuid.MustParse(journey.ID), action, action.ClientActionID)
		require.NoError(t, e)
		require.True(t, journey.CanContinue)
		action = practice.Action{StepID: step.ID, ExpectedRevision: journey.Revision, ClientActionID: "private-journey-" + uuid.NewString(), Action: "continue"}
		journey, e = psvc.Act(ctx, owner, uuid.MustParse(journey.ID), action, action.ClientActionID)
		require.NoError(t, e)
	}
	require.True(t, sawSentenceBuild, "the HTTP export must cover actual sentence-build action receipts")
	// A separate due word earns verified daily credit, then a Continue action
	// chooses seven days. This preference is history, not another graded review.
	reviewNow := time.Now().UTC().Truncate(time.Microsecond)
	userWord, reviewMeaning := uuid.New(), uuid.MustParse("ac93d068-7c3e-55c9-9d85-3fd3518592dc")
	_, e = db.ExecContext(ctx, `INSERT INTO user_words(id,user_id,meaning_id,status,source,review_step,added_at,created_at,updated_at)VALUES($1,$2,$3,'new','manual',0,$4,$4,$4)`, userWord, owner, reviewMeaning, reviewNow)
	require.NoError(t, e)
	gam := gamification.NewService(gamification.NewRepository(db))
	mis := missions.NewService(missions.NewRepository(db), gam)
	reviewRepo := reviews.NewPostgreSQLRepository(db, clock.Fixed{T: reviewNow}, reviews.WithGamificationService(gam), reviews.WithMissionsService(mis))
	dsvc := dailylesson.NewService(db, reviewRepo, clock.Fixed{T: reviewNow})
	daily, e := dsvc.Start(ctx, owner, "private-daily-start")
	require.NoError(t, e)
	require.Equal(t, 1, daily.TotalWords)
	var answer string
	require.NoError(t, db.QueryRowContext(ctx, `SELECT state->'Words'->0->>'AnswerID' FROM daily_lesson_sessions WHERE id=$1`, daily.ID).Scan(&answer))
	da := dailylesson.Action{StepID: daily.CurrentStep.ID, ExpectedRevision: daily.Revision, ClientActionID: "private-daily-answer", Action: "answer", ChoiceID: answer}
	daily, e = dsvc.Act(ctx, owner, uuid.MustParse(daily.ID), da, "private-daily-answer-request")
	require.NoError(t, e)
	require.True(t, daily.Feedback.Correct)
	require.True(t, daily.Feedback.Credited)
	require.Equal(t, 1, daily.CreditedToday)
	da = dailylesson.Action{StepID: daily.CurrentStep.ID, ExpectedRevision: daily.Revision, ClientActionID: "private-daily-next", Action: "continue", NextReview: "7d"}
	daily, e = dsvc.Act(ctx, owner, uuid.MustParse(daily.ID), da, "private-daily-next-request")
	require.NoError(t, e)
	require.Equal(t, "completed", daily.Status)
	var scheduled time.Time
	require.NoError(t, db.QueryRowContext(ctx, `SELECT next_review_at FROM user_words WHERE id=$1`, userWord).Scan(&scheduled))
	require.Equal(t, reviewNow.Add(7*24*time.Hour), scheduled)
	repo := accounts.NewPostgreSQLRepository(db)
	projected, e := repo.ExportPersonalData(ctx, owner)
	require.NoError(t, e)
	var expected map[string]any
	require.NoError(t, json.Unmarshal(projected, &expected))
	c := &clock.Fixed{T: testNow()}
	ar := auth.NewMemoryRepository()
	ar.UpsertUser(&auth.User{ID: owner, Email: owner.String() + "@http-export.invalid", Status: "active"})
	as := auth.NewService(ar, nil, nil, c, auth.NewFixedWindowRateLimiter(c, time.Hour, 100), auth.Config{Environment: "test", Cookie: auth.CookieConfig{Name: "session", CSRName: "csrf"}})
	svc := accounts.NewService(repo, ar, nil, accounts.NewMemoryIdempotencyStore(), c, auth.NewFixedWindowRateLimiter(c, time.Hour, 100), accounts.Config{})
	a := humachi.New(chi.NewMux(), huma.DefaultConfig("http-export-pg", "test"))
	a.UseMiddleware(withHumaContext)
	a.UseMiddleware(AuthMiddleware(as))
	RegisterPersonalDataExports(a, svc, as)
	csrf, cookie := as.IssueCSRFCookie()
	req := exportRequest(t, owner, "private-download-key", true)
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrf)
	w := httptest.NewRecorder()
	a.Adapter().ServeHTTP(w, req)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	var download map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &download))
	require.Equal(t, "1.7", download["schemaVersion"])
	for _, field := range []string{"wordLists", "storySessions", "learningPreferences", "practiceSessions", "dailyLessons", "dailyReviewSuccesses", "wordKnowledge", "guidedLessons"} {
		require.Equal(t, expected[field], download[field], "actual HTTP projection %s", field)
	}
	require.Contains(t, w.Body.String(), "Original guided answer")
	require.Contains(t, w.Body.String(), "Original practice answer")
	require.Len(t, download["dailyLessons"], 1)
	require.Len(t, download["dailyReviewSuccesses"], 1)
	require.Contains(t, w.Body.String(), `"nextReview":"7d"`)
	require.Contains(t, w.Body.String(), `"journeyLesson":1`)
	require.Contains(t, w.Body.String(), `"tokenIds":[`)
	require.Contains(t, w.Body.String(), `"selectedTokenIds":[`)
	require.Contains(t, w.Body.String(), `"correctTokenIds":[`)
	require.Contains(t, w.Body.String(), `"revealed":true`)
	for _, hidden := range []string{"private-list-create", "private-list-member", "private-practice-start", "private-practice-client", "private-practice-request", "private-story-start", "private-story-client", "private-story-request", "private-lesson-start", "private-lesson-client", "private-lesson-request", "private-download-key", "private-journey-", "private-daily-", foreign.ID, `"snapshot"`, `"accepted"`, `"explanations"`, `"fingerprint"`, `"clientActionId"`, `"CorrectTokens"`, `"CorrectChoice"`, `"AnswerID"`, `"state"`, `"idempotencyKey"`} {
		require.NotContains(t, w.Body.String(), hidden)
	}
}
