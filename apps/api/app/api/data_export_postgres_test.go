package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/lessons"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/practice"
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
			for _, table := range []string{"practice_actions", "practice_sessions", "lesson_actions", "lesson_sessions", "story_actions", "story_sessions", "word_list_actions", "user_word_lists", "user_word_knowledge", "user_learning_preferences", "users"} {
				column := "user_id"
				if table == "users" {
					column = "id"
				}
				_, e := db.ExecContext(context.Background(), "DELETE FROM "+table+" WHERE "+column+"=$1", u)
				require.NoError(t, e)
			}
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
	require.Equal(t, "1.5", download["schemaVersion"])
	for _, field := range []string{"wordLists", "storySessions", "learningPreferences", "practiceSessions", "wordKnowledge", "guidedLessons"} {
		require.Equal(t, expected[field], download[field], "actual HTTP projection %s", field)
	}
	require.Contains(t, w.Body.String(), "Original guided answer")
	require.Contains(t, w.Body.String(), "Original practice answer")
	for _, hidden := range []string{"private-list-create", "private-list-member", "private-practice-start", "private-practice-client", "private-practice-request", "private-story-start", "private-story-client", "private-story-request", "private-lesson-start", "private-lesson-client", "private-lesson-request", "private-download-key", foreign.ID, `"snapshot"`, `"accepted"`, `"explanations"`, `"fingerprint"`, `"clientActionId"`} {
		require.NotContains(t, w.Body.String(), hidden)
	}
}
