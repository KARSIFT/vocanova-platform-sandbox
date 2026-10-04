// Package lessons owns guided practice, independently of saved-word scheduling,
// daily review credit and mastery. Only server-validated actions advance a lesson.
package lessons

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"golang.org/x/text/unicode/norm"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
)

var (
	ErrNotFound           = errors.New("lesson not found")
	ErrConflict           = errors.New("lesson state changed; reload to continue")
	ErrInvalid            = errors.New("invalid lesson action")
	ErrContentUnavailable = errors.New("lesson content unavailable")
)

type Word struct {
	MeaningID    string `json:"meaningId"`
	WordText     string `json:"wordText"`
	WordSlug     string `json:"wordSlug"`
	PartOfSpeech string `json:"partOfSpeech"`
	Definition   string `json:"definition"`
	Example      string `json:"example"`
	UsageNote    string `json:"usageNote"`
}
type Choice struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
type Step struct {
	ID             string   `json:"id"`
	Kind           string   `json:"kind" enum:"teach,recall,context,typed_recall,listening_choice"`
	SpeechText     string   `json:"speechText,omitempty"`
	SpeechLanguage string   `json:"speechLanguage,omitempty"`
	Word           Word     `json:"word"`
	Prompt         string   `json:"prompt"`
	Context        string   `json:"context,omitempty"`
	Choices        []Choice `json:"choices"`
}
type Feedback struct {
	Answer          string `json:"answer,omitempty"`
	StepID          string `json:"stepId"`
	Correct         bool   `json:"correct"`
	Explanation     string `json:"explanation"`
	CorrectChoiceID string `json:"correctChoiceId"`
}
type Session struct {
	ExerciseVersion     string     `json:"exerciseVersion,omitempty"`
	ID                  string     `json:"id"`
	LessonKey           string     `json:"lessonKey"`
	LessonVersion       string     `json:"lessonVersion"`
	Title               string     `json:"title"`
	SituationSlug       string     `json:"situationSlug"`
	Status              string     `json:"status" enum:"in_progress,completed"`
	Revision            int        `json:"revision"`
	CompletedSteps      int        `json:"completedSteps"`
	TotalSteps          int        `json:"totalSteps"`
	Words               []Word     `json:"words"`
	CurrentStep         *Step      `json:"currentStep"`
	Feedback            *Feedback  `json:"feedback"`
	CanContinue         bool       `json:"canContinue"`
	FirstAnswersCorrect int        `json:"firstAnswersCorrect"`
	QuestionsAnswered   int        `json:"questionsAnswered"`
	CompletedAt         *time.Time `json:"completedAt,omitempty"`
}
type Summary struct {
	Key            string `json:"key"`
	Version        string `json:"version"`
	Title          string `json:"title"`
	SituationSlug  string `json:"situationSlug"`
	SituationTitle string `json:"situationTitle"`
	Description    string `json:"description"`
	WordCount      int    `json:"wordCount"`
	StepCount      int    `json:"stepCount"`
	Status         string `json:"status" enum:"not_started,in_progress,completed"`
	SessionID      string `json:"sessionId,omitempty"`
	CompletedSteps int    `json:"completedSteps"`
}
type Action struct {
	TypedAnswer      string `json:"typedAnswer,omitempty" maxLength:"200"`
	StepID           string `json:"stepId" minLength:"1" maxLength:"100"`
	ExpectedRevision int    `json:"expectedRevision" minimum:"0"`
	ClientActionID   string `json:"clientActionId" minLength:"1" maxLength:"128"`
	Action           string `json:"action" enum:"answer,continue"`
	ChoiceID         string `json:"choiceId,omitempty" maxLength:"100"`
}

// Snapshot includes private answer keys. Never serialize it as an API response.
type Snapshot struct {
	ExerciseVersion string        `json:"exerciseVersion,omitempty"`
	Definition      Definition    `json:"definition"`
	Words           []Word        `json:"words"`
	Steps           []privateStep `json:"steps"`
}
type privateStep struct {
	Accepted        []string          `json:"accepted,omitempty"`
	Step            Step              `json:"step"`
	CorrectChoiceID string            `json:"correctChoiceId"`
	Explanations    map[string]string `json:"explanations"`
}
type State struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	Snapshot            Snapshot
	Index               int
	Revision            int
	Feedback            *Feedback
	FirstAnswersCorrect int
	QuestionsAnswered   int
	CompletedAt         *time.Time
}

type Repository interface {
	List(context.Context, uuid.UUID) ([]State, error)
	Start(context.Context, uuid.UUID, Definition, string, time.Time) (*State, error)
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
	return &Service{repo: repo, clock: c}
}
func (s *Service) List(ctx context.Context, u uuid.UUID) ([]Summary, error) {
	if u == uuid.Nil {
		return nil, ErrNotFound
	}
	states, err := s.repo.List(ctx, u)
	if err != nil {
		return nil, err
	}
	out := make([]Summary, 0, len(catalog))
	for _, d := range catalog {
		item := Summary{Key: d.Key, Version: d.Version, Title: d.Title, SituationSlug: d.SituationSlug, SituationTitle: d.SituationTitle, Description: d.Description, WordCount: len(d.Words), StepCount: len(d.Words) * 3, Status: "not_started"}
		for _, st := range states {
			if st.Snapshot.Definition.Key == d.Key {
				p := project(st)
				item.Status = p.Status
				item.SessionID = p.ID
				item.CompletedSteps = p.CompletedSteps
				item.Version = p.LessonVersion
				item.Title = p.Title
				item.StepCount = p.TotalSteps
				item.WordCount = len(p.Words)
				break
			}
		}
		out = append(out, item)
	}
	return out, nil
}
func validKey(k string) bool {
	return strings.TrimSpace(k) != "" && len(k) <= 128 && validPlain(k, 128)
}
func validPlain(s string, max int) bool {
	if len(s) > max || !utf8.ValidString(s) {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}
func normalizeAnswer(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(norm.NFC.String(s)), " "))
}
func (s *Service) Start(ctx context.Context, u uuid.UUID, key, idem string) (*Session, error) {
	if u == uuid.Nil {
		return nil, ErrNotFound
	}
	if !validKey(idem) {
		return nil, ErrInvalid
	}
	d, ok := definition(key)
	if !ok {
		return nil, ErrNotFound
	}
	st, err := s.repo.Start(ctx, u, d, idem, s.clock.Now().UTC())
	if err != nil {
		return nil, err
	}
	p := project(*st)
	return &p, nil
}
func (s *Service) Get(ctx context.Context, u, id uuid.UUID) (*Session, error) {
	if u == uuid.Nil || id == uuid.Nil {
		return nil, ErrNotFound
	}
	st, err := s.repo.Get(ctx, u, id)
	if err != nil {
		return nil, err
	}
	p := project(*st)
	return &p, nil
}
func (s *Service) Act(ctx context.Context, u, id uuid.UUID, a Action, idem string) (*Session, error) {
	if u == uuid.Nil || id == uuid.Nil {
		return nil, ErrNotFound
	}
	if !validKey(idem) || !validKey(a.ClientActionID) || !validPlain(a.StepID, 100) || !validPlain(a.ChoiceID, 100) || !validPlain(a.TypedAnswer, 200) || a.StepID == "" || (a.TypedAnswer != "" && a.ChoiceID != "") || (a.Action == "continue" && (a.TypedAnswer != "" || a.ChoiceID != "")) || a.ExpectedRevision < 0 || (a.Action != "answer" && a.Action != "continue") {
		return nil, ErrInvalid
	}
	st, err := s.repo.Act(ctx, u, id, a, idem, s.clock.Now().UTC())
	if err != nil {
		return nil, err
	}
	p := project(*st)
	return &p, nil
}
func project(st State) Session {
	d := st.Snapshot.Definition
	p := Session{ExerciseVersion: st.Snapshot.ExerciseVersion, ID: st.ID.String(), LessonKey: d.Key, LessonVersion: d.Version, Title: d.Title, SituationSlug: d.SituationSlug, Status: "in_progress", Revision: st.Revision, CompletedSteps: st.Index, TotalSteps: len(st.Snapshot.Steps), Words: st.Snapshot.Words, Feedback: st.Feedback, FirstAnswersCorrect: st.FirstAnswersCorrect, QuestionsAnswered: st.QuestionsAnswered, CompletedAt: st.CompletedAt}
	if st.CompletedAt != nil {
		p.Status = "completed"
		p.Feedback = nil
		return p
	}
	if st.Index < len(st.Snapshot.Steps) {
		step := st.Snapshot.Steps[st.Index].Step
		p.CurrentStep = &step
		p.CanContinue = step.Kind == "teach" || (st.Feedback != nil && st.Feedback.StepID == step.ID && st.Feedback.Correct)
	}
	return p
}
func apply(st *State, a Action, now time.Time) error {
	if !validPlain(a.TypedAnswer, 200) || (a.TypedAnswer != "" && a.ChoiceID != "") || (a.Action != "continue" && a.Action != "answer") {
		return ErrInvalid
	}
	if st.CompletedAt != nil || a.ExpectedRevision != st.Revision || st.Index < 0 || st.Index >= len(st.Snapshot.Steps) {
		return ErrConflict
	}
	step := st.Snapshot.Steps[st.Index]
	if a.StepID != step.Step.ID {
		return ErrConflict
	}
	if a.Action == "continue" {
		if a.ChoiceID != "" || a.TypedAnswer != "" {
			return ErrInvalid
		}
		if step.Step.Kind != "teach" && (st.Feedback == nil || !st.Feedback.Correct || st.Feedback.StepID != a.StepID) {
			return ErrConflict
		}
		st.Index++
		st.Feedback = nil
		if st.Index == len(st.Snapshot.Steps) {
			st.CompletedAt = &now
		}
	} else {
		if step.Step.Kind == "teach" {
			return ErrInvalid
		}
		if st.Feedback != nil && st.Feedback.Correct {
			return ErrConflict
		}
		correct := false
		explanation := ""
		answer := ""
		if step.Step.Kind == "typed_recall" {
			if a.ChoiceID != "" || normalizeAnswer(a.TypedAnswer) == "" || len(step.Accepted) == 0 {
				return ErrInvalid
			}
			for _, accepted := range step.Accepted {
				if normalizeAnswer(a.TypedAnswer) == normalizeAnswer(accepted) {
					correct = true
				}
			}
			answer = step.Step.Word.WordText
			explanation = "The word or phrase is “" + answer + "”. " + step.Step.Word.Definition
		} else {
			if a.TypedAnswer != "" {
				return ErrInvalid
			}
			var ok bool
			explanation, ok = step.Explanations[a.ChoiceID]
			if !ok {
				return ErrInvalid
			}
			correct = a.ChoiceID == step.CorrectChoiceID
		}
		if st.Feedback == nil {
			st.QuestionsAnswered++
			if correct {
				st.FirstAnswersCorrect++
			}
		}
		st.Feedback = &Feedback{StepID: a.StepID, Correct: correct, Explanation: explanation, CorrectChoiceID: step.CorrectChoiceID, Answer: answer}

	}
	st.Revision++
	return nil
}
func fingerprint(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
