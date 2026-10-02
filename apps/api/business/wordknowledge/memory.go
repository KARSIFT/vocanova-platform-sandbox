package wordknowledge

import (
	"context"
	"github.com/google/uuid"
	"sync"
	"time"
)

type MemoryRepository struct {
	mu       sync.Mutex
	meanings map[uuid.UUID]bool
	states   map[[2]uuid.UUID]State
	receipts map[uuid.UUID]map[string]string
}

func NewMemoryRepository(meanings []uuid.UUID) *MemoryRepository {
	r := &MemoryRepository{meanings: map[uuid.UUID]bool{}, states: map[[2]uuid.UUID]State{}, receipts: map[uuid.UUID]map[string]string{}}
	for _, id := range meanings {
		r.meanings[id] = true
	}
	return r
}
func (r *MemoryRepository) get(u, m uuid.UUID) (*State, error) {
	if !r.meanings[m] {
		return nil, ErrNotFound
	}
	st, ok := r.states[[2]uuid.UUID{u, m}]
	if !ok {
		st = State{MeaningID: m}
	}
	if st.UpdatedAt != nil {
		t := *st.UpdatedAt
		st.UpdatedAt = &t
	}
	return &st, nil
}
func (r *MemoryRepository) Get(ctx context.Context, u, m uuid.UUID) (*State, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.get(u, m)
}
func (r *MemoryRepository) Write(ctx context.Context, req WriteRequest, now time.Time) (*State, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.meanings[req.MeaningID] {
		return nil, ErrNotFound
	}
	fp := fingerprint(req)
	if prior, ok := r.receipts[req.UserID][req.IdempotencyKey]; ok {
		if fp != prior {
			return nil, ErrConflict
		}
		return r.get(req.UserID, req.MeaningID)
	}
	key := [2]uuid.UUID{req.UserID, req.MeaningID}
	if req.Delete {
		delete(r.states, key)
	} else {
		note := req.Note
		if req.AssessmentOnly {
			note = r.states[key].Note
		}
		r.states[key] = State{MeaningID: req.MeaningID, SelfReportedKnown: req.SelfReportedKnown, Note: note, UpdatedAt: &now}
	}
	if r.receipts[req.UserID] == nil {
		r.receipts[req.UserID] = map[string]string{}
	}
	r.receipts[req.UserID][req.IdempotencyKey] = fp
	return r.get(req.UserID, req.MeaningID)
}
func (r *MemoryRepository) KnownStates(ctx context.Context, u uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := map[uuid.UUID]bool{}
	for _, id := range ids {
		out[id] = r.states[[2]uuid.UUID{u, id}].SelfReportedKnown
	}
	return out, nil
}
func (r *MemoryRepository) CountKnown(ctx context.Context, u uuid.UUID) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for key, st := range r.states {
		if key[0] == u && st.SelfReportedKnown {
			n++
		}
	}
	return n, nil
}
