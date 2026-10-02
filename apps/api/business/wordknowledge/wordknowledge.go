// Package wordknowledge owns private self-assessments and notes. These records
// never change saved words, scheduling, review credit, missions or proficiency.
package wordknowledge

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
	ErrNotFound = errors.New("meaning or active learner not found")
	ErrInvalid  = errors.New("invalid word knowledge request")
	ErrConflict = errors.New("idempotency key already used for a different request")
)

type State struct {
	MeaningID         uuid.UUID
	SelfReportedKnown bool
	Note              string
	UpdatedAt         *time.Time
}
type WriteRequest struct {
	UserID, MeaningID    uuid.UUID
	SelfReportedKnown    bool
	Note, IdempotencyKey string
	Delete               bool
	// AssessmentOnly changes known status while preserving the current private note.
	AssessmentOnly bool
}
type Repository interface {
	Get(context.Context, uuid.UUID, uuid.UUID) (*State, error)
	Write(context.Context, WriteRequest, time.Time) (*State, error)
	KnownStates(context.Context, uuid.UUID, []uuid.UUID) (map[uuid.UUID]bool, error)
	CountKnown(context.Context, uuid.UUID) (int, error)
}
type Service struct {
	repo  Repository
	clock clock.Clock
}

func NewService(repo Repository, clk clock.Clock) *Service {
	if clk == nil {
		clk = clock.Real{}
	}
	return &Service{repo: repo, clock: clk}
}
func (s *Service) Get(ctx context.Context, u, m uuid.UUID) (*State, error) {
	if u == uuid.Nil || m == uuid.Nil {
		return nil, ErrInvalid
	}
	return s.repo.Get(ctx, u, m)
}
func (s *Service) Write(ctx context.Context, req WriteRequest) (*State, error) {
	req.Note = strings.TrimSpace(req.Note)
	if req.AssessmentOnly && (req.Delete || req.Note != "") {
		return nil, ErrInvalid
	}
	if req.UserID == uuid.Nil || req.MeaningID == uuid.Nil || strings.TrimSpace(req.IdempotencyKey) == "" || len(req.IdempotencyKey) > 128 || strings.ContainsRune(req.IdempotencyKey, 0) || !utf8.ValidString(req.Note) || strings.ContainsRune(req.Note, 0) || utf8.RuneCountInString(req.Note) > 2000 {
		return nil, ErrInvalid
	}
	if req.Delete {
		req.Note = ""
		req.SelfReportedKnown = false
	}
	return s.repo.Write(ctx, req, s.clock.Now())
}
func (s *Service) KnownStates(ctx context.Context, u uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]bool, error) {
	return s.repo.KnownStates(ctx, u, ids)
}
func (s *Service) CountKnown(ctx context.Context, u uuid.UUID) (int, error) {
	return s.repo.CountKnown(ctx, u)
}
func fingerprint(req WriteRequest) string {
	raw, _ := json.Marshal(struct {
		MeaningID uuid.UUID
		Known     bool
		Note      string
		Delete    bool
		// Omit false to retain existing PUT/DELETE receipt fingerprints.
		AssessmentOnly bool `json:",omitempty"`
	}{req.MeaningID, req.SelfReportedKnown, req.Note, req.Delete, req.AssessmentOnly})
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
