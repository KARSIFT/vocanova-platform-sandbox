// Package knowledgecatalog presents a stable, sourced word inventory and atomic
// discovery choices. Self-assessment never awards review credit or mastery.
package knowledgecatalog

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
)

const CatalogVersion = "english-10000-v1"
const BandSize = 100
const TotalWords = 10000

var (
	ErrInvalid  = errors.New("invalid knowledge map request")
	ErrNotFound = errors.New("catalog meaning or active learner not found")
	ErrConflict = errors.New("idempotency key already used for another request")
	ErrStale    = errors.New("word changed; refresh before choosing again")
)

type Ordering struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	SourceURL   string `json:"sourceUrl"`
}

func ordering() Ordering {
	return Ordering{"Common words first", "Ordered by English subtitle frequency among selected words. Source definitions and a reviewed grammar core; not a personalized or CEFR ranking.", "https://github.com/hermitdave/FrequencyWords"}
}

type Band struct {
	Number        int `json:"number"`
	StartRank     int `json:"startRank"`
	EndRank       int `json:"endRank"`
	TotalWords    int `json:"totalWords"`
	KnownCount    int `json:"knownCount"`
	LearningCount int `json:"learningCount"`
	UnseenCount   int `json:"unseenCount"`
}
type Item struct {
	Rank              int       `json:"rank"`
	MeaningID         uuid.UUID `json:"meaningId"`
	WordID            uuid.UUID `json:"wordId"`
	WordText          string    `json:"wordText"`
	WordSlug          string    `json:"wordSlug"`
	PartOfSpeech      string    `json:"partOfSpeech"`
	ShortDefinition   string    `json:"shortDefinition"`
	ExampleSentence   string    `json:"exampleSentence,omitempty"`
	SelfReportedKnown bool      `json:"selfReportedKnown"`
	Saved             bool      `json:"saved"`
	Due               bool      `json:"due"`
	Revision          string    `json:"revision"`
}
type CurrentBand struct {
	Number int    `json:"number"`
	Items  []Item `json:"items"`
}
type Map struct {
	Attribution    Attribution `json:"attribution"`
	CatalogVersion string      `json:"catalogVersion"`
	TotalWords     int         `json:"totalWords"`
	BandSize       int         `json:"bandSize"`
	Ordering       Ordering    `json:"ordering"`
	Bands          []Band      `json:"bands"`
	CurrentBand    CurrentBand `json:"currentBand"`
}
type DecisionRequest struct {
	UserID, MeaningID                          uuid.UUID
	Decision, ExpectedRevision, IdempotencyKey string
}
type DecisionResult struct {
	MeaningID         uuid.UUID `json:"meaningId"`
	Decision          string    `json:"decision"`
	Revision          string    `json:"revision"`
	SelfReportedKnown bool      `json:"selfReportedKnown"`
	Saved             bool      `json:"saved"`
	Due               bool      `json:"due"`
}
type Repository interface {
	Read(context.Context, uuid.UUID, int, time.Time) (*Map, error)
	Decide(context.Context, DecisionRequest, time.Time) (*DecisionResult, error)
}
type Service struct {
	repo  Repository
	clock clock.Clock
}

func NewService(repo Repository, clk clock.Clock) *Service {
	if clk == nil {
		clk = clock.Real{}
	}
	return &Service{repo, clk}
}
func (s *Service) Read(ctx context.Context, user uuid.UUID, band int) (*Map, error) {
	if user == uuid.Nil || band < 1 || band > TotalWords/BandSize {
		return nil, ErrInvalid
	}
	return s.repo.Read(ctx, user, band, s.clock.Now())
}
func (s *Service) Decide(ctx context.Context, req DecisionRequest) (*DecisionResult, error) {
	if req.UserID == uuid.Nil || req.MeaningID == uuid.Nil || (req.Decision != "learn" && req.Decision != "known") || len(req.ExpectedRevision) != 64 || strings.TrimSpace(req.IdempotencyKey) == "" || len(req.IdempotencyKey) > 128 || strings.ContainsRune(req.IdempotencyKey, 0) {
		return nil, ErrInvalid
	}
	if _, err := hex.DecodeString(req.ExpectedRevision); err != nil {
		return nil, ErrInvalid
	}
	return s.repo.Decide(ctx, req, s.clock.Now())
}
func fingerprint(req DecisionRequest) string {
	raw, _ := json.Marshal([]string{"knowledge-map-v1", req.MeaningID.String(), req.Decision, req.ExpectedRevision})
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
