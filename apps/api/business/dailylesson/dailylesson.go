// Package dailylesson owns durable server-graded lessons selected from actual
// due saved meanings. Historical catalog lessons and practice stay independent.
package dailylesson

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/knowledgecatalog"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/reviews"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/revieweligibility"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrNotFound           = errors.New("daily lesson not found")
	ErrInvalid            = errors.New("invalid daily lesson action")
	ErrConflict           = errors.New("daily lesson changed or idempotency conflict")
	ErrNoDue              = errors.New("no words due")
	ErrContentUnavailable = errors.New("meaning choices unavailable")
)

type DailyLessonChoice struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
type DailyLessonStep struct {
	PronunciationMeaningID string   `json:"pronunciationMeaningId,omitempty" doc:"Selected meaning used only to resolve pronunciation; no answer key."`
	PronunciationRevision  string   `json:"pronunciationRevision,omitempty"`
	ID                     string   `json:"id"`
	Kind                   string   `json:"kind" enum:"meaning_choice"`
	WordText               string   `json:"wordText"`
	WordSlug               string   `json:"wordSlug"`
	PartOfSpeech           string   `json:"partOfSpeech"`
	Choices                []Choice `json:"choices"`
}
type DailyLessonFeedback struct {
	StepID           string     `json:"stepId"`
	Correct          bool       `json:"correct"`
	Answer           string     `json:"answer"`
	Explanation      string     `json:"explanation"`
	MeaningID        string     `json:"meaningId"`
	WordText         string     `json:"wordText"`
	Credited         bool       `json:"credited"`
	Skipped          bool       `json:"skipped"`
	Reason           string     `json:"reason,omitempty"`
	CorrectChoiceID  string     `json:"correctChoiceId,omitempty"`
	SelectedChoiceID string     `json:"selectedChoiceId,omitempty"`
	NextReviewAt     *time.Time `json:"nextReviewAt,omitempty"`
	ReviewStep       *int       `json:"reviewStep,omitempty"`
}
type DailyLessonSession struct {
	Attribution         knowledgecatalog.Attribution `json:"attribution"`
	ID                  string                       `json:"id"`
	Status              string                       `json:"status" enum:"in_progress,completed"`
	Revision            int                          `json:"revision"`
	CompletedWords      int                          `json:"completedWords"`
	TotalWords          int                          `json:"totalWords"`
	FirstAnswersCorrect int                          `json:"firstAnswersCorrect"`
	QuestionsAnswered   int                          `json:"questionsAnswered"`
	CreditedToday       int                          `json:"creditedToday"`
	CurrentStep         *Step                        `json:"currentStep"`
	Feedback            *Feedback                    `json:"feedback"`
	CanContinue         bool                         `json:"canContinue"`
	CreatedAt           time.Time                    `json:"createdAt"`
	UpdatedAt           time.Time                    `json:"updatedAt"`
	CompletedAt         *time.Time                   `json:"completedAt,omitempty"`
}
type DailyLessonCurrent struct {
	Session  *Session `json:"session"`
	DueCount int      `json:"dueCount"`
}
type DailyLessonAction struct {
	StepID           string `json:"stepId" minLength:"1" maxLength:"100"`
	ExpectedRevision int    `json:"expectedRevision" minimum:"0"`
	ClientActionID   string `json:"clientActionId" minLength:"1" maxLength:"128"`
	Action           string `json:"action" enum:"answer,continue,skip"`
	ChoiceID         string `json:"choiceId,omitempty" maxLength:"100"`
	WasHintUsed      bool   `json:"wasHintUsed,omitempty"`
	NextReview       string `json:"nextReview,omitempty" enum:"automatic,1d,3d,7d,14d,30d,60d,120d,known"`
}

// Domain aliases keep callers concise while schema names remain unique.
type Choice = DailyLessonChoice
type Step = DailyLessonStep
type Feedback = DailyLessonFeedback
type Session = DailyLessonSession
type Current = DailyLessonCurrent
type Action = DailyLessonAction

type word struct {
	UserWordID, MeaningID, WordID string
	ReviewCount                   int
	Public                        Step
	AnswerID, Definition          string
	Answered, Resolved            bool
	KnowledgeUpdatedAt            *time.Time
}
type state struct {
	ID, UserID                                                                      uuid.UUID
	Words                                                                           []word
	Queue                                                                           []int
	CompletedWords, FirstAnswersCorrect, QuestionsAnswered, CreditedToday, Revision int
	Feedback                                                                        *Feedback
	CreatedAt, UpdatedAt                                                            time.Time
	CompletedAt                                                                     *time.Time
}
type Service struct {
	db      *sql.DB
	reviews *reviews.PostgreSQLRepository
	clock   clock.Clock
}

func NewService(db *sql.DB, r *reviews.PostgreSQLRepository, c clock.Clock) *Service {
	if c == nil {
		c = clock.Real{}
	}
	return &Service{db: db, reviews: r, clock: c}
}
func validKey(s string) bool {
	return strings.TrimSpace(s) != "" && len(s) <= 128 && utf8.ValidString(s) && !strings.ContainsRune(s, 0)
}
func fingerprint(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// project preserves the original frozen snapshot denominator, including completed
// historical sessions that allowed multiple uses of a canonical word. New session
// selection already contains one use per canonical word.
func project(st *state) *Session {
	out := &Session{Attribution: knowledgecatalog.SourceAttribution(), ID: st.ID.String(), Status: "in_progress", Revision: st.Revision, CompletedWords: st.CompletedWords, TotalWords: len(st.Words), FirstAnswersCorrect: st.FirstAnswersCorrect, QuestionsAnswered: st.QuestionsAnswered, CreditedToday: st.CreditedToday, Feedback: st.Feedback, CanContinue: st.Feedback != nil, CreatedAt: st.CreatedAt, UpdatedAt: st.UpdatedAt, CompletedAt: st.CompletedAt}
	if st.CompletedAt != nil {
		out.Status = "completed"
	} else if len(st.Queue) > 0 {
		step := st.Words[st.Queue[0]].Public
		step.PronunciationMeaningID = st.Words[st.Queue[0]].MeaningID
		step.PronunciationRevision = meaningidentity.Revision(step.PronunciationMeaningID, step.WordText, step.PartOfSpeech, st.Words[st.Queue[0]].Definition)
		out.CurrentStep = &step
	}
	return out
}
func scan(row interface{ Scan(...any) error }) (*state, error) {
	var raw []byte
	if e := row.Scan(&raw); errors.Is(e, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if e != nil {
		return nil, e
	}
	var st state
	if e := json.Unmarshal(raw, &st); e != nil {
		return nil, e
	}
	return &st, nil
}
func lockUser(ctx context.Context, tx *sql.Tx, u uuid.UUID) error {
	var id uuid.UUID
	e := tx.QueryRowContext(ctx, "SELECT id FROM users WHERE id=$1 AND status='active' AND deleted_at IS NULL FOR UPDATE", u).Scan(&id)
	if errors.Is(e, sql.ErrNoRows) {
		return ErrNotFound
	}
	return e
}
func (s *Service) Get(ctx context.Context, u, id uuid.UUID) (*Session, error) {
	tx, e := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	st, e := scan(tx.QueryRowContext(ctx, "SELECT s.state FROM daily_lesson_sessions s JOIN users u ON u.id=s.user_id WHERE s.user_id=$1 AND s.id=$2 AND u.status='active' AND u.deleted_at IS NULL", u, id))
	if e != nil {
		return nil, e
	}
	if e = pruneQueue(ctx, tx, st, s.clock.Now().UTC()); e != nil {
		return nil, e
	}
	if e = refreshCredits(ctx, tx, st, s.clock.Now().UTC()); e != nil {
		return nil, e
	}
	return project(st), tx.Commit()
}
func dueSQL() string {
	return ` FROM user_words uw JOIN word_meanings wm ON wm.id=uw.meaning_id JOIN canonical_words cw ON cw.id=wm.word_id WHERE uw.user_id=$1 AND ` + revieweligibility.Due("uw", "$2", true)
}

func wordKey(w word) string {
	if w.WordID != "" {
		return w.WordID
	}
	return w.MeaningID
}

// pruneQueue removes legacy delayed retries and sibling uses without changing a
// frozen question or grading it. The current answer feedback stays until Continue.
func pruneQueue(ctx context.Context, tx *sql.Tx, st *state, now time.Time) error {
	if st.CompletedAt != nil {
		return nil
	}
	ids := []string{}
	for _, w := range st.Words {
		if _, err := uuid.Parse(w.WordID); err == nil {
			ids = append(ids, w.WordID)
		}
	}
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT wm.word_id::text FROM word_meanings wm WHERE wm.word_id=ANY($3::uuid[]) AND `+revieweligibility.ReviewedToday("$1", "wm.word_id", "$2"), st.UserID, now, pq.Array(ids))
	if err != nil {
		return err
	}
	reviewed := map[string]bool{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		reviewed[id] = true
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	queue := []int{}
	seen := map[string]bool{}
	for position, index := range st.Queue {
		w := &st.Words[index]
		key := wordKey(*w)
		if position == 0 && st.Feedback != nil {
			queue = append(queue, index)
			seen[key] = true
			continue
		}
		if seen[key] || w.Resolved || w.Answered || reviewed[key] {
			continue
		}
		queue = append(queue, index)
		seen[key] = true
	}
	st.Queue = queue
	pending := map[string]bool{}
	for _, index := range queue {
		if !st.Words[index].Resolved {
			pending[wordKey(st.Words[index])] = true
		}
	}
	// Keep the original frozen snapshot denominator and earned answer history.
	// Legacy sibling cards removed from the queue are completed as passive skips;
	// they must not shrink historical accuracy below already earned answers.
	st.CompletedWords = len(st.Words) - len(pending)
	if len(queue) == 0 {
		st.CompletedAt = &now
	}
	return nil
}

func (s *Service) Current(ctx context.Context, u uuid.UUID) (*Current, error) {
	tx, e := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	var active bool
	if e = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND status='active' AND deleted_at IS NULL)", u).Scan(&active); e != nil {
		return nil, e
	}
	if !active {
		return nil, ErrNotFound
	}
	out := &Current{}
	st, e := scan(tx.QueryRowContext(ctx, "SELECT state FROM daily_lesson_sessions WHERE user_id=$1 AND status='in_progress'", u))
	if e == nil {
		if e = pruneQueue(ctx, tx, st, s.clock.Now().UTC()); e != nil {
			return nil, e
		}
		if e = refreshCredits(ctx, tx, st, s.clock.Now().UTC()); e != nil {
			return nil, e
		}
		out.Session = project(st)
	} else if !errors.Is(e, ErrNotFound) {
		return nil, e
	}
	if e = tx.QueryRowContext(ctx, "SELECT count(*)"+dueSQL(), u, s.clock.Now().UTC()).Scan(&out.DueCount); e != nil {
		return nil, e
	}
	return out, tx.Commit()
}
func save(ctx context.Context, tx *sql.Tx, st *state) error {
	raw, e := json.Marshal(st)
	if e != nil {
		return e
	}
	status := "in_progress"
	if st.CompletedAt != nil {
		status = "completed"
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO daily_lesson_sessions(id,user_id,state,status,revision,created_at,updated_at,completed_at)VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT(id) DO UPDATE SET state=EXCLUDED.state,status=EXCLUDED.status,revision=EXCLUDED.revision,updated_at=EXCLUDED.updated_at,completed_at=EXCLUDED.completed_at`, st.ID, st.UserID, string(raw), status, st.Revision, st.CreatedAt, st.UpdatedAt, st.CompletedAt)
	return e
}
func receipt(ctx context.Context, tx *sql.Tx, u uuid.UUID, op, key, client, fp string, sid uuid.UUID) (*Session, bool, error) {
	rows, e := tx.QueryContext(ctx, `SELECT session_id,idempotency_key,client_action_id,fingerprint,result FROM daily_lesson_actions WHERE user_id=$1 AND operation=$2 AND(idempotency_key=$3 OR(session_id=$4 AND client_action_id=$5))`, u, op, key, sid, client)
	if e != nil {
		return nil, false, e
	}
	defer rows.Close()
	var out *Session
	for rows.Next() {
		var id uuid.UUID
		var k, c, p string
		var raw []byte
		if e = rows.Scan(&id, &k, &c, &p, &raw); e != nil {
			return nil, false, e
		}
		if (sid != uuid.Nil && id != sid) || k != key || c != client || p != fp {
			return nil, false, ErrConflict
		}
		var got Session
		if e = json.Unmarshal(raw, &got); e != nil {
			return nil, false, e
		}
		out = &got
	}
	if e = rows.Err(); e != nil {
		return nil, false, e
	}
	return out, out != nil, nil
}
func record(ctx context.Context, tx *sql.Tx, st *state, op, key, client, fp string, a any) error {
	action, e := json.Marshal(a)
	if e != nil {
		return e
	}
	out, e := json.Marshal(project(st))
	if e != nil {
		return e
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO daily_lesson_actions(id,session_id,user_id,operation,idempotency_key,client_action_id,fingerprint,action,result,created_at)VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, uuid.New(), st.ID, st.UserID, op, key, client, fp, string(action), string(out), st.UpdatedAt)
	return e
}
func (s *Service) Start(ctx context.Context, u uuid.UUID, key string) (*Session, error) {
	if u == uuid.Nil || !validKey(key) {
		return nil, ErrInvalid
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	if e = lockUser(ctx, tx, u); e != nil {
		return nil, e
	}
	fp := fingerprint(struct{}{})
	if out, ok, e := receipt(ctx, tx, u, "start", key, key, fp, uuid.Nil); e != nil {
		return nil, e
	} else if ok {
		return out, tx.Commit()
	}
	st, e := scan(tx.QueryRowContext(ctx, "SELECT state FROM daily_lesson_sessions WHERE user_id=$1 AND status='in_progress' FOR UPDATE", u))
	if e == nil {
		if e = pruneQueue(ctx, tx, st, s.clock.Now().UTC()); e != nil {
			return nil, e
		}
		if st.CompletedAt != nil {
			previous := st
			if e = save(ctx, tx, previous); e != nil {
				return nil, e
			}
			st, e = s.build(ctx, tx, u, s.clock.Now().UTC())
			if errors.Is(e, ErrNoDue) {
				st, e = previous, nil
			}
		}
	}
	if errors.Is(e, ErrNotFound) {
		st, e = s.build(ctx, tx, u, s.clock.Now().UTC())
	}
	if e != nil {
		return nil, e
	}
	if e = pruneQueue(ctx, tx, st, s.clock.Now().UTC()); e != nil {
		return nil, e
	}
	if e = refreshCredits(ctx, tx, st, s.clock.Now().UTC()); e != nil {
		return nil, e
	}
	if e = save(ctx, tx, st); e != nil {
		return nil, e
	}
	if e = record(ctx, tx, st, "start", key, key, fp, struct{}{}); e != nil {
		return nil, e
	}
	return project(st), tx.Commit()
}
func (s *Service) build(ctx context.Context, tx *sql.Tx, u uuid.UUID, now time.Time) (*state, error) {
	rows, e := tx.QueryContext(ctx, `SELECT uw.id,uw.meaning_id,cw.id,cw.text,cw.normalized_text,wm.part_of_speech,wm.short_definition,uw.total_review_count`+dueSQL()+` ORDER BY COALESCE(uw.next_review_at,'-infinity'::timestamptz),uw.id LIMIT 10`, u, now)
	if e != nil {
		return nil, e
	}
	st := &state{ID: uuid.New(), UserID: u, CreatedAt: now, UpdatedAt: now, Words: []word{}, Queue: []int{}}
	for rows.Next() {
		var w word
		var text, slug, pos string
		if e = rows.Scan(&w.UserWordID, &w.MeaningID, &w.WordID, &text, &slug, &pos, &w.Definition, &w.ReviewCount); e != nil {
			rows.Close()
			return nil, e
		}
		w.Public = Step{ID: uuid.NewString(), Kind: "meaning_choice", WordText: text, WordSlug: strings.ReplaceAll(slug, " ", "-"), PartOfSpeech: pos, Choices: []Choice{}}
		st.Words = append(st.Words, w)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	if len(st.Words) == 0 {
		return nil, ErrNoDue
	}
	for i := range st.Words {
		w := &st.Words[i]
		w.AnswerID = uuid.NewString()
		w.Public.Choices = append(w.Public.Choices, Choice{ID: w.AnswerID, Text: w.Definition})
		candidates, e := tx.QueryContext(ctx, candidateDefinitionsSQL, w.WordID, w.Definition, w.Public.WordText, w.MeaningID)
		if e != nil {
			return nil, e
		}
		definitions := []string{}
		for candidates.Next() {
			var d, candidateWord string
			if e = candidates.Scan(&d, &candidateWord); e != nil {
				candidates.Close()
				return nil, e
			}
			aliases, err := knowledgecatalog.SemanticAliases(w.Public.WordText, candidateWord)
			if err != nil {
				candidates.Close()
				return nil, ErrContentUnavailable
			}
			if aliases {
				continue
			}
			definitions = append(definitions, d)
		}
		e = candidates.Err()
		candidates.Close()
		if e != nil {
			return nil, e
		}
		// Stable pseudorandom choice placement for this frozen session, independent of queue size.
		for n := 0; n < 3 && len(definitions) > 0; n++ {
			h := sha256.Sum256([]byte(st.ID.String() + w.MeaningID + string(rune(n))))
			j := int(h[0]) % len(definitions)
			w.Public.Choices = append(w.Public.Choices, Choice{ID: uuid.NewString(), Text: definitions[j]})
			definitions = append(definitions[:j], definitions[j+1:]...)
		}
		if len(w.Public.Choices) < 3 {
			return nil, ErrContentUnavailable
		}
		h := sha256.Sum256([]byte(w.Public.ID))
		j := int(h[0]) % len(w.Public.Choices)
		w.Public.Choices[0], w.Public.Choices[j] = w.Public.Choices[j], w.Public.Choices[0]
		st.Queue = append(st.Queue, i)
	}
	return st, nil
}
func (s *Service) Act(ctx context.Context, u, id uuid.UUID, a Action, key string) (*Session, error) {
	if u == uuid.Nil || id == uuid.Nil || !validKey(key) || !validKey(a.ClientActionID) || !validKey(a.StepID) || a.ExpectedRevision < 0 || len(a.ChoiceID) > 100 || (a.Action != "answer" && a.Action != "continue" && a.Action != "skip") {
		return nil, ErrInvalid
	}
	if a.NextReview != "" && (a.Action != "continue" || !validNextReview(a.NextReview)) {
		return nil, ErrInvalid
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	if e = lockUser(ctx, tx, u); e != nil {
		return nil, e
	}
	fp := fingerprint(struct {
		ID     uuid.UUID
		Action Action
	}{id, a})
	if out, ok, e := receipt(ctx, tx, u, "action", key, a.ClientActionID, fp, id); e != nil {
		return nil, e
	} else if ok {
		return out, tx.Commit()
	}
	st, e := scan(tx.QueryRowContext(ctx, "SELECT state FROM daily_lesson_sessions WHERE user_id=$1 AND id=$2 FOR UPDATE", u, id))
	if e != nil {
		return nil, e
	}
	now := s.clock.Now().UTC()
	if e = pruneQueue(ctx, tx, st, now); e != nil {
		return nil, e
	}
	if st.CompletedAt != nil || st.Revision != a.ExpectedRevision || len(st.Queue) == 0 {
		return nil, ErrConflict
	}
	w := &st.Words[st.Queue[0]]
	if a.StepID != w.Public.ID {
		return nil, ErrConflict
	}
	if a.Action == "continue" {
		if st.Feedback == nil || a.ChoiceID != "" || a.WasHintUsed {
			return nil, ErrInvalid
		}
		if a.NextReview != "" && a.NextReview != "automatic" {
			if e = reschedule(ctx, tx, u, w, st.Feedback, a.NextReview, now); e != nil {
				return nil, e
			}
			// A later date or self-assessment changes future scheduling only.
			// It never turns an incorrect answer into a credit.
			if !w.Resolved {
				w.Resolved = true
				st.CompletedWords++
			}
		}
		if !w.Resolved {
			w.Resolved = true
			st.CompletedWords++
		}
		st.Queue = st.Queue[1:]
		st.Feedback = nil
		if len(st.Queue) == 0 {
			st.CompletedAt = &now
		}
	} else {
		if st.Feedback != nil {
			return nil, ErrConflict
		}
		if a.Action == "skip" && (a.ChoiceID != "" || a.WasHintUsed) {
			return nil, ErrInvalid
		}
		correct := false
		if a.Action == "answer" {
			found := false
			for _, c := range w.Public.Choices {
				if c.ID == a.ChoiceID {
					found = true
					correct = c.ID == w.AnswerID
				}
			}
			if !found {
				return nil, ErrInvalid
			}
			if a.WasHintUsed {
				correct = false
			}
		}
		reason := ""
		var count int
		var status string
		var deleted sql.NullTime
		var due sql.NullTime
		var known bool
		var knowledgeUpdated sql.NullTime
		e = tx.QueryRowContext(ctx, `SELECT total_review_count,status,deleted_at,next_review_at,EXISTS(SELECT 1 FROM user_word_knowledge k WHERE k.user_id=$2 AND k.meaning_id=uw.meaning_id AND k.self_reported_known),(SELECT k.updated_at FROM user_word_knowledge k WHERE k.user_id=$2 AND k.meaning_id=uw.meaning_id) FROM user_words uw WHERE id=$1 AND user_id=$2 FOR UPDATE`, w.UserWordID, u).Scan(&count, &status, &deleted, &due, &known, &knowledgeUpdated)
		if errors.Is(e, sql.ErrNoRows) {
			reason = "removed"
		} else if e != nil {
			return nil, e
		} else if deleted.Valid || status == "ignored" || status == "archived" || status == "mastered" {
			reason = "removed"
		} else if known {
			reason = "already_known"
		} else if count != w.ReviewCount {
			reason = "already_reviewed"
		} else if !w.Answered && due.Valid && due.Time.After(now) {
			reason = "already_reviewed"
		}
		if reason == "" {
			var reviewed bool
			if e = tx.QueryRowContext(ctx, `SELECT `+revieweligibility.ReviewedToday("$1", "(SELECT word_id FROM word_meanings WHERE id=$2)", "$3"), u, w.MeaningID, now).Scan(&reviewed); e != nil {
				return nil, e
			}
			if reviewed {
				reason = "already_reviewed"
			}
		}
		skipped := a.Action == "skip" || reason != ""
		fb := &Feedback{StepID: w.Public.ID, Correct: correct && !skipped, Answer: w.Definition, Explanation: "", MeaningID: w.MeaningID, WordText: w.Public.WordText, Skipped: skipped, Reason: reason, CorrectChoiceID: w.AnswerID}
		if a.Action == "answer" && reason == "" {
			fb.SelectedChoiceID = a.ChoiceID
		}
		if skipped && (a.Action != "skip" || reason != "") {
			fb.Explanation = "This word was left for later or changed elsewhere."
			w.Resolved = true
			st.CompletedWords++
		} else {
			if !w.Answered && a.Action != "skip" {
				st.QuestionsAnswered++
				if correct {
					st.FirstAnswersCorrect++
				}
			}
			result, rating := reviews.ResultIncorrect, reviews.RatingAgain
			if correct {
				result, rating = reviews.ResultCorrect, reviews.RatingGood
			}
			if a.Action == "skip" {
				result, rating = reviews.ResultSkipped, ""
			}
			req := reviews.SubmitReviewRequest{UserID: u, UserWordID: uuid.MustParse(w.UserWordID), MeaningID: uuid.MustParse(w.MeaningID), AttemptType: reviews.AttemptTypeReview, PromptType: reviews.PromptTypeSelfCheck, Result: result, Rating: rating, AnsweredAt: now, SelectedOptionMeaningID: nil, WasHintUsed: a.WasHintUsed, Source: reviews.SourceReviewSession, ClientAttemptID: "daily:" + fingerprint(a.ClientActionID), IdempotencyKey: "daily:" + fingerprint(key), ExpectedReviewCount: &w.ReviewCount}
			attempt, e := s.reviews.SubmitDailyReviewTx(ctx, tx, req, st.ID)
			if e != nil {
				return nil, e
			}
			w.ReviewCount++
			w.Answered = true
			if knowledgeUpdated.Valid {
				w.KnowledgeUpdatedAt = &knowledgeUpdated.Time
			} else {
				w.KnowledgeUpdatedAt = nil
			}
			var confirmedDue time.Time
			var confirmedStep int
			if e = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM daily_review_successes WHERE review_attempt_id=$1),next_review_at,review_step FROM user_words WHERE id=$2 AND user_id=$3", attempt.ID, w.UserWordID, u).Scan(&fb.Credited, &confirmedDue, &confirmedStep); e != nil {
				return nil, e
			}
			fb.NextReviewAt = &confirmedDue
			fb.ReviewStep = &confirmedStep

			w.Resolved = true
			st.CompletedWords++
			if a.Action == "skip" {
				fb.Explanation = "This word was left for another day."
			} else if correct {
				fb.Explanation = "Correct. This word is reviewed for today."
			} else {
				fb.Explanation = "Review the meaning. This word will return on another day."
			}
		}
		st.Feedback = fb
	}
	st.Revision++
	st.UpdatedAt = now
	if e = pruneQueue(ctx, tx, st, s.clock.Now().UTC()); e != nil {
		return nil, e
	}
	if e = refreshCredits(ctx, tx, st, s.clock.Now().UTC()); e != nil {
		return nil, e
	}
	if e = save(ctx, tx, st); e != nil {
		return nil, e
	}
	if e = record(ctx, tx, st, "action", key, a.ClientActionID, fp, a); e != nil {
		return nil, e
	}
	return project(st), tx.Commit()
}

func validNextReview(value string) bool {
	switch value {
	case "automatic", "1d", "3d", "7d", "14d", "30d", "60d", "120d", "known":
		return true
	default:
		return false
	}
}

// reschedule is a separate learner preference, not a second graded review.
// The parent transaction already owns the user/session locks. Compare the
// schedule and assessment seen by Answer so Continue cannot overwrite newer
// reviews, removals, notes or knowledge decisions from another tab.
func reschedule(ctx context.Context, tx *sql.Tx, u uuid.UUID, w *word, fb *Feedback, value string, now time.Time) error {
	if fb.Skipped || !w.Answered || fb.NextReviewAt == nil || fb.ReviewStep == nil {
		return ErrInvalid
	}
	var count, step int
	var status string
	var deleted, due, knowledgeUpdated sql.NullTime
	var known bool
	err := tx.QueryRowContext(ctx, `SELECT uw.total_review_count,uw.review_step,uw.status,uw.deleted_at,uw.next_review_at,COALESCE(k.self_reported_known,false),k.updated_at
 FROM user_words uw LEFT JOIN user_word_knowledge k ON k.user_id=uw.user_id AND k.meaning_id=uw.meaning_id
 WHERE uw.id=$1 AND uw.user_id=$2 AND uw.meaning_id=$3 FOR UPDATE OF uw`, w.UserWordID, u, w.MeaningID).Scan(&count, &step, &status, &deleted, &due, &known, &knowledgeUpdated)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrConflict
	}
	if err != nil {
		return err
	}
	if deleted.Valid || known || (status != "new" && status != "learning" && status != "reviewing") || count != w.ReviewCount || step != *fb.ReviewStep || !due.Valid || !due.Time.Equal(*fb.NextReviewAt) || knowledgeUpdated.Valid != (w.KnowledgeUpdatedAt != nil) || (knowledgeUpdated.Valid && !knowledgeUpdated.Time.Equal(*w.KnowledgeUpdatedAt)) {
		return ErrConflict
	}
	if value == "known" {
		_, err = tx.ExecContext(ctx, `INSERT INTO user_word_knowledge(user_id,meaning_id,self_reported_known,note,updated_at) VALUES($1,$2,true,'',$3)
 ON CONFLICT(user_id,meaning_id) DO UPDATE SET self_reported_known=true,updated_at=EXCLUDED.updated_at`, u, w.MeaningID, now)
		return err
	}
	days := map[string]int{"1d": 1, "3d": 3, "7d": 7, "14d": 14, "30d": 30, "60d": 60, "120d": 120}[value]
	if days == 0 {
		return ErrInvalid
	}
	_, err = tx.ExecContext(ctx, `UPDATE user_words SET next_review_at=$3,updated_at=$4 WHERE id=$1 AND user_id=$2`, w.UserWordID, u, now.Add(time.Duration(days)*24*time.Hour), now)
	return err
}

// Only this lesson's successful receipts on the current learner-local date count
// as creditedToday. Reading an overnight session must not reuse yesterday's sum.
func refreshCredits(ctx context.Context, tx *sql.Tx, st *state, now time.Time) error {
	return tx.QueryRowContext(ctx, `SELECT count(*) FROM daily_review_successes d JOIN review_attempts a ON a.id=d.review_attempt_id
 JOIN daily_lesson_sessions s ON s.id=d.daily_session_id AND s.user_id=d.user_id
 WHERE d.user_id=$1 AND d.daily_session_id=$2
 AND EXISTS(SELECT 1 FROM jsonb_array_elements(s.state->'Words') w WHERE w->>'MeaningID'=d.meaning_id::text AND w->>'UserWordID'=a.user_word_id::text)
 AND d.local_date=($3 AT TIME ZONE COALESCE((SELECT timezone FROM user_settings WHERE user_id=$1),'UTC'))::date`, st.UserID, st.ID.String(), now).Scan(&st.CreditedToday)
}
