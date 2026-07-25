package aifeedback

import (
	"context"

	"github.com/google/uuid"
)

// MissionUpdater is the P4-flagged seam for daily-mission / streak / point
// updates (VOC-028-D01). The real write is not implemented in P3; the stub
// honestly surfaces that mission completion is not yet wired.
type MissionUpdater interface {
	Update(ctx context.Context, userID, sentenceID uuid.UUID) (bool, error)
}

// StubMissionUpdater returns a backend-decided false result and writes nothing.
type StubMissionUpdater struct{}

// NewStubMissionUpdater creates the P3 mission-completion stub.
func NewStubMissionUpdater() *StubMissionUpdater {
	return &StubMissionUpdater{}
}

// Update always returns false, nil: the real mission write is owned by P4.
func (s *StubMissionUpdater) Update(ctx context.Context, userID, sentenceID uuid.UUID) (bool, error) {
	return false, nil
}
