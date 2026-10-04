// Package wordlists owns private meaning collections independently of SRS and knowledge.
package wordlists

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrNotFound = errors.New("list, meaning or active learner not found")
	ErrInvalid  = errors.New("invalid list request")
	ErrConflict = errors.New("list revision or idempotency conflict")
	ErrLimit    = errors.New("list limit reached")
)

const MaxLists = 50
const MaxMembers = 500

type WordListSummary struct {
	ID                string    `json:"id" format:"uuid"`
	Name              string    `json:"name"`
	Revision          int       `json:"revision"`
	MemberCount       int       `json:"memberCount"`
	UsableMemberCount int       `json:"usableMemberCount" doc:"Active meanings supported by the reviewed practice catalog"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}
type WordListMember struct {
	MeaningID         string    `json:"meaningId" format:"uuid"`
	WordID            string    `json:"wordId" format:"uuid"`
	WordSlug          string    `json:"wordSlug"`
	WordText          string    `json:"wordText"`
	ShortDefinition   string    `json:"shortDefinition"`
	PartOfSpeech      string    `json:"partOfSpeech"`
	PracticeAvailable bool      `json:"practiceAvailable"`
	AddedAt           time.Time `json:"addedAt"`
}
type WordListDetail struct {
	WordListSummary
	Members []WordListMember `json:"members"`
}
type WordListsResponse struct {
	Items []WordListSummary `json:"items"`
}
type WriteRequest struct {
	UserID, ListID, MeaningID uuid.UUID
	Operation                 string
	Name                      string
	ExpectedRevision          int
	IdempotencyKey            string
}
type Repository interface {
	List(context.Context, uuid.UUID) (*WordListsResponse, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (*WordListDetail, error)
	Write(context.Context, WriteRequest, time.Time) (*WordListDetail, error)
}
type Service struct {
	repo  Repository
	clock clock.Clock
}

func NewService(r Repository, c clock.Clock) *Service {
	if c == nil {
		c = clock.Real{}
	}
	return &Service{r, c}
}
func (s *Service) List(ctx context.Context, u uuid.UUID) (*WordListsResponse, error) {
	if u == uuid.Nil {
		return nil, ErrNotFound
	}
	return s.repo.List(ctx, u)
}
func (s *Service) Get(ctx context.Context, u, id uuid.UUID) (*WordListDetail, error) {
	if u == uuid.Nil || id == uuid.Nil {
		return nil, ErrNotFound
	}
	return s.repo.Get(ctx, u, id)
}
func (s *Service) Write(ctx context.Context, r WriteRequest) (*WordListDetail, error) {
	r.Name = strings.TrimSpace(r.Name)
	if r.UserID == uuid.Nil || r.ListID == uuid.Nil || r.ExpectedRevision < 0 || strings.TrimSpace(r.IdempotencyKey) == "" || len(r.IdempotencyKey) > 128 || !utf8.ValidString(r.IdempotencyKey) || strings.ContainsRune(r.IdempotencyKey, 0) {
		return nil, ErrInvalid
	}
	switch r.Operation {
	case "put":
		if r.MeaningID != uuid.Nil || r.Name == "" || !utf8.ValidString(r.Name) || utf8.RuneCountInString(r.Name) > 80 || strings.ContainsAny(r.Name, "\x00\r\n") {
			return nil, ErrInvalid
		}
	case "delete":
		if r.MeaningID != uuid.Nil || r.Name != "" {
			return nil, ErrInvalid
		}
	case "add", "remove":
		if r.MeaningID == uuid.Nil || r.Name != "" {
			return nil, ErrInvalid
		}
	default:
		return nil, ErrInvalid
	}
	return s.repo.Write(ctx, r, s.clock.Now())
}
func fingerprint(r WriteRequest) string {
	raw, _ := json.Marshal(struct {
		ListID, MeaningID uuid.UUID
		Operation, Name   string
		Revision          int
	}{r.ListID, r.MeaningID, r.Operation, r.Name, r.ExpectedRevision})
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
