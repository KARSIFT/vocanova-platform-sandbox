package aifeedback

import (
	"context"
	"sync"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
)

// GenerationGate checks whether AI generation is globally enabled before the
// request-specific rate limiter is consulted. It implements the emergency
// AI-disable switch and global request/cost ceilings from DOC-09 §19.
// A failing gate must not affect non-AI learning features.
type GenerationGate interface {
	Check(ctx context.Context, userID uuid.UUID) error
}

// GenerationGateConfig is the configurable starting set for the AI-disable and
// cost-ceiling seams. Values are intentionally defaults; activation values are
// founder-controlled at runtime.
type GenerationGateConfig struct {
	Enabled                  bool
	DailyRequestCeiling      int
	MonthlyCostWarningCents  int
	MonthlyCostHardStopCents int
	RequestCostCents         int
}

// DefaultGenerationGateConfig returns the approved starting defaults from
// DOC-09 §19. Cost/request is an estimate; real billing is reconciled against
// provider invoices. The defaults leave the gate enabled with loose ceilings so
// tests can exercise the seam without blocking normal mock traffic.
func DefaultGenerationGateConfig() GenerationGateConfig {
	return GenerationGateConfig{
		Enabled:                  true,
		DailyRequestCeiling:      0, // 0 = no global daily ceiling
		MonthlyCostWarningCents:  0, // 0 = no warning
		MonthlyCostHardStopCents: 0, // 0 = no hard stop
		RequestCostCents:         0, // 0 = no cost accounting
	}
}

// MemoryGenerationGate is a deterministic, in-process implementation of
// GenerationGate. It is not safe across processes.
type MemoryGenerationGate struct {
	mu      sync.Mutex
	config  GenerationGateConfig
	clock   clock.Clock
	daily   map[string]int
	monthly map[string]int
}

// NewMemoryGenerationGate creates an in-process generation gate.
func NewMemoryGenerationGate(config GenerationGateConfig, c clock.Clock) *MemoryGenerationGate {
	if c == nil {
		c = clock.Real{}
	}
	return &MemoryGenerationGate{
		config:  config,
		clock:   c,
		daily:   make(map[string]int),
		monthly: make(map[string]int),
	}
}

// Check returns ErrAIGenerationDisabled when the gate is disabled or the global
// hard stop is reached. It returns ErrRateLimited when the daily request ceiling
// is exhausted. On success it records the request against the running counters.
func (g *MemoryGenerationGate) Check(ctx context.Context, userID uuid.UUID) error {
	if !g.config.Enabled {
		return ErrAIGenerationDisabled
	}

	now := g.clock.Now().UTC()
	dayKey := now.Format("2006-01-02")
	monthKey := now.Format("2006-01")

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.config.DailyRequestCeiling > 0 && g.daily[dayKey] >= g.config.DailyRequestCeiling {
		return ErrRateLimited
	}

	if g.config.MonthlyCostHardStopCents > 0 && g.config.RequestCostCents > 0 {
		if g.monthly[monthKey]+g.config.RequestCostCents >= g.config.MonthlyCostHardStopCents {
			return ErrAIGenerationDisabled
		}
	}

	g.daily[dayKey]++
	if g.config.RequestCostCents > 0 {
		g.monthly[monthKey] += g.config.RequestCostCents
	}
	return nil
}

// AlwaysEnabledGate is a no-op gate that always permits generation.
type AlwaysEnabledGate struct{}

// NewAlwaysEnabledGate returns a gate that never blocks.
func NewAlwaysEnabledGate() *AlwaysEnabledGate {
	return &AlwaysEnabledGate{}
}

// Check implements GenerationGate.
func (AlwaysEnabledGate) Check(ctx context.Context, userID uuid.UUID) error { return nil }

// DisabledGate is a gate that always reports generation as disabled.
type DisabledGate struct{}

// NewDisabledGate returns a gate that always blocks AI generation.
func NewDisabledGate() *DisabledGate {
	return &DisabledGate{}
}

// Check implements GenerationGate.
func (DisabledGate) Check(ctx context.Context, userID uuid.UUID) error {
	return ErrAIGenerationDisabled
}

// CostWarningStatus returns whether the current monthly spend has crossed the
// configured warning threshold. It is a read-only status check and never
// blocks generation.
func (g *MemoryGenerationGate) CostWarningStatus() (currentCents, warningCents int, triggered bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	monthKey := g.clock.Now().UTC().Format("2006-01")
	currentCents = g.monthly[monthKey]
	warningCents = g.config.MonthlyCostWarningCents
	triggered = warningCents > 0 && currentCents >= warningCents
	return currentCents, warningCents, triggered
}
