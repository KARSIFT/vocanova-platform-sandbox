// Package practice provides repeatable self-study without changing SRS, saved
// words, missions, mastery or rewards. Only the API grades and advances sessions.
package practice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrNotFound           = errors.New("practice not found")
	ErrInvalid            = errors.New("invalid practice action")
	ErrConflict           = errors.New("practice changed or idempotency conflict")
	ErrNoMistakes         = errors.New("no supported mistakes to practise")
	ErrListEmpty          = errors.New("list has no supported meanings to practise")
	ErrContentUnavailable = errors.New("practice content unavailable")
)

const ContentVersion = "starter-90-v2"
const GradingVersion = "exact-recall-v1"

type PracticeStartRequest struct {
	Mode         string `json:"mode" enum:"typed_recall,listening_choice,mistakes"`
	LessonKey    string `json:"lessonKey,omitempty" maxLength:"100"`
	ListID       string `json:"listId,omitempty" format:"uuid"`
	ListRevision *int   `json:"listRevision,omitempty" minimum:"1"`
}
type PracticeChoice struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
type PracticeStep struct {
	ID             string   `json:"id"`
	Kind           string   `json:"kind" enum:"typed_recall,listening_choice"`
	Prompt         string   `json:"prompt"`
	Choices        []Choice `json:"choices"`
	SpeechText     string   `json:"speechText,omitempty"`
	SpeechLanguage string   `json:"speechLanguage,omitempty"`
}
type PracticeFeedback struct {
	StepID      string `json:"stepId"`
	Correct     bool   `json:"correct"`
	Assisted    bool   `json:"assisted"`
	Answer      string `json:"answer"`
	Explanation string `json:"explanation"`
	WordText    string `json:"wordText"`
	WordSlug    string `json:"wordSlug"`
	MeaningID   string `json:"meaningId"`
}
type PracticeSummary struct {
	ListID              string     `json:"listId,omitempty"`
	ListName            string     `json:"listName,omitempty"`
	ListRevision        *int       `json:"listRevision,omitempty"`
	ID                  string     `json:"id"`
	Mode                string     `json:"mode"`
	LessonKey           string     `json:"lessonKey,omitempty"`
	ContentVersion      string     `json:"contentVersion"`
	GradingVersion      string     `json:"gradingVersion"`
	Status              string     `json:"status" enum:"in_progress,completed"`
	Revision            int        `json:"revision"`
	CompletedSteps      int        `json:"completedSteps"`
	TotalSteps          int        `json:"totalSteps"`
	FirstAnswersCorrect int        `json:"firstAnswersCorrect"`
	QuestionsAnswered   int        `json:"questionsAnswered"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
	CompletedAt         *time.Time `json:"completedAt,omitempty"`
}
type PracticeSession struct {
	Summary
	CurrentStep *Step     `json:"currentStep"`
	Feedback    *Feedback `json:"feedback"`
	CanContinue bool      `json:"canContinue"`
}
type PracticeList struct {
	Items             []Summary `json:"items"`
	AvailableMistakes int       `json:"availableMistakes"`
}
type PracticeAction struct {
	StepID           string `json:"stepId" minLength:"1" maxLength:"100"`
	ExpectedRevision int    `json:"expectedRevision" minimum:"0"`
	ClientActionID   string `json:"clientActionId" minLength:"1" maxLength:"128"`
	Action           string `json:"action" enum:"answer,continue,reveal"`
	TypedAnswer      string `json:"typedAnswer,omitempty" maxLength:"200"`
	ChoiceID         string `json:"choiceId,omitempty" maxLength:"100"`
}

// Private types are persisted only; public responses always pass through project.
type Word struct{ MeaningID, WordText, WordSlug, Definition, LessonKey string }
type Source struct {
	Kind, ID  string
	CreatedAt time.Time
}
type privateStep struct {
	Public        Step
	Word          Word
	Accepted      []string
	CorrectChoice string
	Source        *Source
}
type Snapshot struct {
	ListID, ListName                                string `json:",omitempty"`
	ListRevision                                    *int   `json:",omitempty"`
	Mode, LessonKey, ContentVersion, GradingVersion string
	Steps                                           []privateStep
}
type State struct {
	ID, UserID                                              uuid.UUID
	Snapshot                                                Snapshot
	Index, Revision, FirstAnswersCorrect, QuestionsAnswered int
	Feedback                                                *Feedback
	CreatedAt, UpdatedAt                                    time.Time
	CompletedAt                                             *time.Time
}
type Repository interface {
	List(context.Context, uuid.UUID) ([]State, int, error)
	Start(context.Context, uuid.UUID, StartRequest, string, time.Time) (*State, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (*State, error)
	Act(context.Context, uuid.UUID, uuid.UUID, Action, string, time.Time) (*State, error)
}
type Service struct {
	repo  Repository
	clock clock.Clock
}

func NewService(repo Repository, c clock.Clock) *Service {
	if c == nil {
		c = clock.Real{}
	}
	return &Service{repo, c}
}
func (s *Service) List(ctx context.Context, u uuid.UUID) (*List, error) {
	if u == uuid.Nil {
		return nil, ErrNotFound
	}
	states, n, err := s.repo.List(ctx, u)
	if err != nil {
		return nil, err
	}
	out := &List{Items: []Summary{}, AvailableMistakes: n}
	for _, st := range states {
		out.Items = append(out.Items, project(st).Summary)
	}
	return out, nil
}
func validKey(k string) bool {
	return strings.TrimSpace(k) != "" && len(k) <= 128 && !strings.ContainsRune(k, 0) && utf8.ValidString(k)
}
func (s *Service) Start(ctx context.Context, u uuid.UUID, req StartRequest, key string) (*Session, error) {
	if u == uuid.Nil {
		return nil, ErrNotFound
	}
	if !validKey(key) || !validLesson(req.LessonKey) || (req.Mode != "typed_recall" && req.Mode != "listening_choice" && req.Mode != "mistakes") {
		return nil, ErrInvalid
	}
	if req.ListID != "" {
		if id, err := uuid.Parse(req.ListID); err != nil || id == uuid.Nil || req.ListRevision == nil || *req.ListRevision < 1 || req.LessonKey != "" || req.Mode == "mistakes" {
			return nil, ErrInvalid
		}
	} else if req.ListRevision != nil {
		return nil, ErrInvalid
	}
	st, err := s.repo.Start(ctx, u, req, key, s.clock.Now())
	if err != nil {
		return nil, err
	}
	v := project(*st)
	return &v, nil
}
func (s *Service) Get(ctx context.Context, u, id uuid.UUID) (*Session, error) {
	if u == uuid.Nil || id == uuid.Nil {
		return nil, ErrNotFound
	}
	st, err := s.repo.Get(ctx, u, id)
	if err != nil {
		return nil, err
	}
	v := project(*st)
	return &v, nil
}
func (s *Service) Act(ctx context.Context, u, id uuid.UUID, a Action, key string) (*Session, error) {
	if u == uuid.Nil || id == uuid.Nil {
		return nil, ErrNotFound
	}
	if !validKey(key) || !validKey(a.ClientActionID) || a.ExpectedRevision < 0 || a.StepID == "" || len(a.StepID) > 100 || len(a.ChoiceID) > 100 || !utf8.ValidString(a.TypedAnswer) || strings.ContainsRune(a.TypedAnswer, 0) || utf8.RuneCountInString(a.TypedAnswer) > 200 || (a.Action != "answer" && a.Action != "continue" && a.Action != "reveal") {
		return nil, ErrInvalid
	}
	st, err := s.repo.Act(ctx, u, id, a, key, s.clock.Now())
	if err != nil {
		return nil, err
	}
	v := project(*st)
	return &v, nil
}
func project(st State) Session {
	status := "in_progress"
	if st.CompletedAt != nil {
		status = "completed"
	}
	p := Session{Summary: Summary{ListID: st.Snapshot.ListID, ListName: st.Snapshot.ListName, ListRevision: st.Snapshot.ListRevision, ID: st.ID.String(), Mode: st.Snapshot.Mode, LessonKey: st.Snapshot.LessonKey, ContentVersion: st.Snapshot.ContentVersion, GradingVersion: st.Snapshot.GradingVersion, Status: status, Revision: st.Revision, CompletedSteps: st.Index, TotalSteps: len(st.Snapshot.Steps), FirstAnswersCorrect: st.FirstAnswersCorrect, QuestionsAnswered: st.QuestionsAnswered, CreatedAt: st.CreatedAt, UpdatedAt: st.UpdatedAt, CompletedAt: st.CompletedAt}}
	if st.CompletedAt == nil && st.Index < len(st.Snapshot.Steps) {
		step := st.Snapshot.Steps[st.Index].Public
		p.CurrentStep = &step
		p.Feedback = st.Feedback
		p.CanContinue = st.Feedback != nil
	}
	return p
}
func normalize(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(norm.NFC.String(s))), " "))
}
func fingerprint(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// apply returns whether this fresh action independently resolved its source.
func apply(st *State, a Action, now time.Time) (bool, error) {
	if st.CompletedAt != nil || a.ExpectedRevision != st.Revision || st.Index >= len(st.Snapshot.Steps) {
		return false, ErrConflict
	}
	step := st.Snapshot.Steps[st.Index]
	if a.StepID != step.Public.ID {
		return false, ErrConflict
	}
	resolved := false
	if a.Action == "continue" {
		if a.TypedAnswer != "" || a.ChoiceID != "" {
			return false, ErrInvalid
		}
		if st.Feedback == nil {
			return false, ErrConflict
		}
		st.Index++
		st.Feedback = nil
		if st.Index == len(st.Snapshot.Steps) {
			st.CompletedAt = &now
		}
	}
	if a.Action == "answer" || a.Action == "reveal" {
		if st.Feedback != nil && (st.Feedback.Correct || a.Action == "reveal") {
			return false, ErrConflict
		}
		first := st.Feedback == nil
		assisted := !first
		correct := false
		if a.Action == "reveal" {
			if a.TypedAnswer != "" || a.ChoiceID != "" {
				return false, ErrInvalid
			}
			assisted = true
		} else if step.Public.Kind == "typed_recall" {
			if a.ChoiceID != "" || normalize(a.TypedAnswer) == "" {
				return false, ErrInvalid
			}
			for _, v := range step.Accepted {
				if normalize(a.TypedAnswer) == normalize(v) {
					correct = true
				}
			}
		} else {
			if a.TypedAnswer != "" || a.ChoiceID == "" {
				return false, ErrInvalid
			}
			valid := false
			for _, c := range step.Public.Choices {
				if c.ID == a.ChoiceID {
					valid = true
				}
			}
			if !valid {
				return false, ErrInvalid
			}
			correct = a.ChoiceID == step.CorrectChoice
		}
		if first {
			st.QuestionsAnswered++
			if correct && !assisted {
				st.FirstAnswersCorrect++
				resolved = step.Source != nil
			}
		}
		explanation := "The word from this lesson is “" + step.Word.WordText + "”. " + step.Word.Definition
		if assisted {
			explanation += " This answer was practised with help."
		}
		st.Feedback = &Feedback{StepID: a.StepID, Correct: correct, Assisted: assisted, Answer: step.Word.WordText, Explanation: explanation, WordText: step.Word.WordText, WordSlug: step.Word.WordSlug, MeaningID: step.Word.MeaningID}
	}
	st.Revision++
	st.UpdatedAt = now
	return resolved, nil
}

// Package-local aliases preserve concise domain code; public reflected names
// are globally unique in the shared Huma schema registry.
type StartRequest = PracticeStartRequest
type Choice = PracticeChoice
type Step = PracticeStep
type Feedback = PracticeFeedback
type Summary = PracticeSummary
type Session = PracticeSession
type List = PracticeList
type Action = PracticeAction
