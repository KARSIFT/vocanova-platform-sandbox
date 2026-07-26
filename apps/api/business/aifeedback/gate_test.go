package aifeedback

import (
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDefaultGenerationGateConfigIsEnabledWithLooseCeilings(t *testing.T) {
	cfg := DefaultGenerationGateConfig()
	assert.True(t, cfg.Enabled)
	assert.Equal(t, 0, cfg.DailyRequestCeiling)
	assert.Equal(t, 0, cfg.MonthlyCostWarningCents)
	assert.Equal(t, 0, cfg.MonthlyCostHardStopCents)
}

func TestMemoryGenerationGateAllowsRequestsWhenEnabled(t *testing.T) {
	c := clock.Fixed{T: time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)}
	gate := NewMemoryGenerationGate(DefaultGenerationGateConfig(), c)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	assert.NoError(t, gate.Check(t.Context(), userID))
	assert.NoError(t, gate.Check(t.Context(), userID))
}

func TestDisabledGateAlwaysReturnsDisabled(t *testing.T) {
	gate := NewDisabledGate()
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	assert.ErrorIs(t, gate.Check(t.Context(), userID), ErrAIGenerationDisabled)
}

func TestMemoryGenerationGateEnforcesDailyRequestCeiling(t *testing.T) {
	c := clock.Fixed{T: time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)}
	gate := NewMemoryGenerationGate(GenerationGateConfig{Enabled: true, DailyRequestCeiling: 2}, c)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	assert.NoError(t, gate.Check(t.Context(), userID))
	assert.NoError(t, gate.Check(t.Context(), userID))
	assert.ErrorIs(t, gate.Check(t.Context(), userID), ErrRateLimited)
}

func TestMemoryGenerationGateEnforcesMonthlyCostHardStop(t *testing.T) {
	c := clock.Fixed{T: time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)}
	gate := NewMemoryGenerationGate(GenerationGateConfig{
		Enabled:                  true,
		MonthlyCostHardStopCents: 100,
		RequestCostCents:         50,
	}, c)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	assert.NoError(t, gate.Check(t.Context(), userID))
	assert.ErrorIs(t, gate.Check(t.Context(), userID), ErrAIGenerationDisabled)
}

func TestMemoryGenerationGateCostWarning(t *testing.T) {
	c := clock.Fixed{T: time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)}
	gate := NewMemoryGenerationGate(GenerationGateConfig{
		Enabled:                 true,
		MonthlyCostWarningCents: 100,
		RequestCostCents:        60,
	}, c)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	assert.NoError(t, gate.Check(t.Context(), userID))
	current, warning, triggered := gate.CostWarningStatus()
	assert.Equal(t, 60, current)
	assert.Equal(t, 100, warning)
	assert.False(t, triggered)
}
