// Package discoveryposition stores optional navigation preferences, never learning credit.
package discoveryposition

import (
	"context"
	"errors"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/knowledgecatalog"
	"github.com/google/uuid"
)

var (
	ErrAccountChanged = errors.New("discovery account changed")
	ErrInvalid        = errors.New("invalid discovery position")
	ErrConflict       = errors.New("discovery position changed")
	ErrNotFound       = errors.New("active learner not found")
)

type Position struct {
	CatalogVersion string `json:"catalogVersion"`
	Band           int    `json:"band" minimum:"1" maximum:"100"`
	Rank           int    `json:"rank" minimum:"1" maximum:"10000"`
	Active         bool   `json:"active"`
	Revision       int64  `json:"revision" minimum:"1"`
}
type Update struct {
	ExpectedAccountID uuid.UUID `json:"expectedAccountId" format:"uuid"`
	CatalogVersion    string    `json:"catalogVersion" minLength:"1" maxLength:"64"`
	Band              int       `json:"band" minimum:"1" maximum:"100"`
	Rank              int       `json:"rank" minimum:"1" maximum:"10000"`
	Active            bool      `json:"active"`
	ExpectedRevision  int64     `json:"expectedRevision" minimum:"0"`
}

func (u Update) Validate() error {
	if u.ExpectedAccountID == uuid.Nil || u.CatalogVersion != knowledgecatalog.CatalogVersion || u.Rank < 1 || u.Rank > knowledgecatalog.TotalWords || u.Band != (u.Rank-1)/knowledgecatalog.BandSize+1 || u.ExpectedRevision < 0 {
		return ErrInvalid
	}
	return nil
}

type Repository interface {
	Get(context.Context, uuid.UUID) (*Position, error)
	Put(context.Context, uuid.UUID, Update) (*Position, error)
}
type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo} }
func (s *Service) Get(ctx context.Context, user uuid.UUID) (*Position, error) {
	if user == uuid.Nil {
		return nil, ErrNotFound
	}
	return s.repo.Get(ctx, user)
}
func (s *Service) Put(ctx context.Context, user uuid.UUID, u Update) (*Position, error) {
	if user == uuid.Nil {
		return nil, ErrNotFound
	}
	if err := u.Validate(); err != nil {
		return nil, err
	}
	// The request identity is immutable: this comparison and the write use the
	// same authenticated owner, without a separate identity preflight.
	if u.ExpectedAccountID != user {
		return nil, ErrAccountChanged
	}
	return s.repo.Put(ctx, user, u)
}
