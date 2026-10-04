// Package stories provides original self-study stories, without SRS or reward credit.
package stories

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
)

var (
	ErrNotFound           = errors.New("story not found")
	ErrInvalid            = errors.New("invalid story request")
	ErrConflict           = errors.New("story changed or replay conflict")
	ErrContentUnavailable = errors.New("story content unavailable")
)

const ContentVersion = "original-dialogues-v1"
const GradingVersion = "curated-choice-v1"

type StoryLine struct {
	ID      string `json:"id"`
	Speaker string `json:"speaker"`
	Text    string `json:"text"`
}
type StoryVocabulary struct {
	WordText   string `json:"wordText"`
	WordSlug   string `json:"wordSlug"`
	MeaningID  string `json:"meaningId"`
	Definition string `json:"definition"`
}
type StoryChoice struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
type StoryStep struct {
	ID      string        `json:"id"`
	Kind    string        `json:"kind" enum:"line,comprehension,phrase_completion"`
	Line    *StoryLine    `json:"line,omitempty"`
	Prompt  string        `json:"prompt,omitempty"`
	Choices []StoryChoice `json:"choices,omitempty"`
}
type StoryFeedback struct {
	StepID      string `json:"stepId"`
	Correct     bool   `json:"correct"`
	Answer      string `json:"answer"`
	Explanation string `json:"explanation"`
}
type StorySummary struct {
	ID                  string     `json:"id"`
	StoryKey            string     `json:"storyKey"`
	Title               string     `json:"title"`
	Situation           string     `json:"situation"`
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
type StorySession struct {
	StorySummary
	CurrentStep  *StoryStep        `json:"currentStep"`
	Feedback     *StoryFeedback    `json:"feedback"`
	CanContinue  bool              `json:"canContinue"`
	VisibleLines []StoryLine       `json:"visibleLines"`
	Vocabulary   []StoryVocabulary `json:"vocabulary"`
}
type StoryCatalogItem struct {
	Key            string            `json:"key"`
	Title          string            `json:"title"`
	Description    string            `json:"description"`
	Situation      string            `json:"situation"`
	Level          string            `json:"level"`
	ContentVersion string            `json:"contentVersion"`
	LineCount      int               `json:"lineCount"`
	QuestionCount  int               `json:"questionCount"`
	Vocabulary     []StoryVocabulary `json:"vocabulary"`
	LatestSession  *StorySummary     `json:"latestSession,omitempty"`
}
type StoryLibrary struct {
	Items []StoryCatalogItem `json:"items"`
}
type StoryReading struct {
	StoryCatalogItem
	Lines []StoryLine `json:"lines"`
}
type StoryStartRequest struct {
	StoryKey string `json:"storyKey" minLength:"1" maxLength:"100"`
}
type StoryAction struct {
	StepID           string `json:"stepId" minLength:"1" maxLength:"100"`
	ExpectedRevision int    `json:"expectedRevision" minimum:"0"`
	ClientActionID   string `json:"clientActionId" minLength:"1" maxLength:"128"`
	Action           string `json:"action" enum:"answer,continue"`
	ChoiceID         string `json:"choiceId,omitempty" maxLength:"100"`
}

// These private snapshot types never leave the repository. Historical sessions
// use their own frozen steps rather than rebuilding from the current catalog.
type privateStep struct {
	Public        StoryStep
	CorrectChoice string
	Explanation   string
}
type Snapshot struct {
	Story          StoryCatalogItem
	ContentVersion string
	GradingVersion string
	Steps          []privateStep
}
type State struct {
	ID, UserID                                              uuid.UUID
	Snapshot                                                Snapshot
	Index, Revision, FirstAnswersCorrect, QuestionsAnswered int
	Feedback                                                *StoryFeedback
	CreatedAt, UpdatedAt                                    time.Time
	CompletedAt                                             *time.Time
}
type Repository interface {
	Active(context.Context, uuid.UUID) error
	List(context.Context, uuid.UUID) ([]State, error)
	Start(context.Context, uuid.UUID, StoryStartRequest, string, time.Time) (*State, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (*State, error)
	Act(context.Context, uuid.UUID, uuid.UUID, StoryAction, string, time.Time) (*State, error)
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
func (s *Service) List(ctx context.Context, u uuid.UUID) (*StoryLibrary, error) {
	if u == uuid.Nil {
		return nil, ErrNotFound
	}
	states, err := s.repo.List(ctx, u)
	if err != nil {
		return nil, err
	}
	out := &StoryLibrary{Items: []StoryCatalogItem{}}
	for _, story := range catalog {
		item := story.Story
		for _, st := range states {
			if st.Snapshot.Story.Key == item.Key {
				v := project(st).StorySummary
				item.LatestSession = &v
				break
			}
		}
		out.Items = append(out.Items, item)
	}
	return out, nil
}
func (s *Service) Read(ctx context.Context, u uuid.UUID, key string) (*StoryReading, error) {
	if u == uuid.Nil {
		return nil, ErrNotFound
	}
	if err := s.repo.Active(ctx, u); err != nil {
		return nil, err
	}
	snap, err := build(key)
	if err != nil {
		return nil, err
	}
	out := &StoryReading{StoryCatalogItem: snap.Story, Lines: []StoryLine{}}
	for _, step := range snap.Steps {
		if step.Public.Line != nil {
			out.Lines = append(out.Lines, *step.Public.Line)
		}
	}
	return out, nil
}
func validKey(k string) bool {
	return strings.TrimSpace(k) != "" && len(k) <= 128 && utf8.ValidString(k) && !strings.ContainsRune(k, 0)
}
func (s *Service) Start(ctx context.Context, u uuid.UUID, req StoryStartRequest, key string) (*StorySession, error) {
	if u == uuid.Nil {
		return nil, ErrNotFound
	}
	if !validKey(key) || !validKey(req.StoryKey) || len(req.StoryKey) > 100 {
		return nil, ErrInvalid
	}
	st, err := s.repo.Start(ctx, u, req, key, s.clock.Now())
	if err != nil {
		return nil, err
	}
	v := project(*st)
	return &v, nil
}
func (s *Service) Get(ctx context.Context, u, id uuid.UUID) (*StorySession, error) {
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
func (s *Service) Act(ctx context.Context, u, id uuid.UUID, a StoryAction, key string) (*StorySession, error) {
	if u == uuid.Nil || id == uuid.Nil {
		return nil, ErrNotFound
	}
	if !validKey(key) || !validKey(a.ClientActionID) || a.ExpectedRevision < 0 || a.StepID == "" || len(a.StepID) > 100 || len(a.ChoiceID) > 100 || (a.Action != "answer" && a.Action != "continue") {
		return nil, ErrInvalid
	}
	st, err := s.repo.Act(ctx, u, id, a, key, s.clock.Now())
	if err != nil {
		return nil, err
	}
	v := project(*st)
	return &v, nil
}
func project(st State) StorySession {
	status := "in_progress"
	if st.CompletedAt != nil {
		status = "completed"
	}
	out := StorySession{StorySummary: StorySummary{ID: st.ID.String(), StoryKey: st.Snapshot.Story.Key, Title: st.Snapshot.Story.Title, Situation: st.Snapshot.Story.Situation, ContentVersion: st.Snapshot.ContentVersion, GradingVersion: st.Snapshot.GradingVersion, Status: status, Revision: st.Revision, CompletedSteps: st.Index, TotalSteps: len(st.Snapshot.Steps), FirstAnswersCorrect: st.FirstAnswersCorrect, QuestionsAnswered: st.QuestionsAnswered, CreatedAt: st.CreatedAt, UpdatedAt: st.UpdatedAt, CompletedAt: st.CompletedAt}, VisibleLines: []StoryLine{}, Vocabulary: st.Snapshot.Story.Vocabulary}
	for i, step := range st.Snapshot.Steps {
		if i <= st.Index && step.Public.Line != nil {
			out.VisibleLines = append(out.VisibleLines, *step.Public.Line)
		}
	}
	if st.CompletedAt == nil && st.Index < len(st.Snapshot.Steps) {
		step := st.Snapshot.Steps[st.Index].Public
		out.CurrentStep = &step
		out.Feedback = st.Feedback
		out.CanContinue = step.Kind == "line" || (st.Feedback != nil && st.Feedback.Correct)
	}
	return out
}
func fingerprint(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func apply(st *State, a StoryAction, now time.Time) error {
	if st.CompletedAt != nil || a.ExpectedRevision != st.Revision || st.Index >= len(st.Snapshot.Steps) {
		return ErrConflict
	}
	step := st.Snapshot.Steps[st.Index]
	if a.StepID != step.Public.ID {
		return ErrConflict
	}
	switch a.Action {
	case "continue":
		if a.ChoiceID != "" {
			return ErrInvalid
		}
		if step.Public.Kind != "line" && (st.Feedback == nil || !st.Feedback.Correct) {
			return ErrConflict
		}
		st.Index++
		st.Feedback = nil
		if st.Index == len(st.Snapshot.Steps) {
			st.CompletedAt = &now
		}
	case "answer":
		if step.Public.Kind == "line" || a.ChoiceID == "" {
			return ErrInvalid
		}
		if st.Feedback != nil && st.Feedback.Correct {
			return ErrConflict
		}
		valid := false
		answer := ""
		for _, c := range step.Public.Choices {
			if c.ID == a.ChoiceID {
				valid = true
			}
			if c.ID == step.CorrectChoice {
				answer = c.Text
			}
		}
		if !valid {
			return ErrInvalid
		}
		correct := a.ChoiceID == step.CorrectChoice
		if st.Feedback == nil {
			st.QuestionsAnswered++
			if correct {
				st.FirstAnswersCorrect++
			}
		}
		st.Feedback = &StoryFeedback{StepID: step.Public.ID, Correct: correct, Answer: answer, Explanation: step.Explanation}
	default:
		return ErrInvalid
	}
	st.Revision++
	st.UpdatedAt = now
	return nil
}
