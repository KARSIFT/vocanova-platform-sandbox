package content

import (
	"context"
	"github.com/google/uuid"
)

// KnowledgeReader exposes only assessments, never private note text.
type KnowledgeReader interface {
	KnownStates(context.Context, uuid.UUID, []uuid.UUID) (map[uuid.UUID]bool, error)
}

func (s *Service) SetKnowledgeReader(reader KnowledgeReader) {
	s.knowledge = reader
	if memory, ok := s.repo.(*MemoryRepository); ok {
		memory.knowledge = reader
	}
}
func (s *Service) knownStates(ctx context.Context, u uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]bool, error) {
	if s.knowledge == nil || u == uuid.Nil || len(ids) == 0 {
		return map[uuid.UUID]bool{}, nil
	}
	return s.knowledge.KnownStates(ctx, u, ids)
}
