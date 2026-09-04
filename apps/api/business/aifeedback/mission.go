package aifeedback

import (
	"context"

	"github.com/google/uuid"
)

// MissionUpdater is the seam for daily-mission / streak / point updates
// triggered by a completed sentence-feedback attempt (VOC-028-D01). The real
// implementation is missions.MissionUpdater (apps/api/business/missions/
// service.go), wired in by the production composition root
// (apps/api/app/api/production.go) as of issue #1177. NewService falls back
// to StubMissionUpdater below only when no MissionUpdater is supplied (e.g.
// tests, or the OpenAPI-generation-only wiring in openapi.go), so a caller
// that omits it gets an honest no-op instead of a panic.
type MissionUpdater interface {
	Update(ctx context.Context, userID, sentenceID uuid.UUID) (bool, error)
}

// StubMissionUpdater returns a backend-decided false result and writes
// nothing. It exists as a safe default for callers that don't have a real
// MissionUpdater available (see the package doc comment above); it must not
// be used in production request handling.
type StubMissionUpdater struct{}

// NewStubMissionUpdater creates the no-op mission-completion stub.
func NewStubMissionUpdater() *StubMissionUpdater {
	return &StubMissionUpdater{}
}

// Update always returns false, nil: this stub never writes.
func (s *StubMissionUpdater) Update(ctx context.Context, userID, sentenceID uuid.UUID) (bool, error) {
	return false, nil
}
