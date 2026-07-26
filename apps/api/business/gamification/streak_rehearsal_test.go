package gamification

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReconcileStreakMultiDayGapBreak_RehearsalR06 covers the multi-day
// streak-break scenario explicitly called out in the spec as VOC-030-R06.
func TestReconcileStreakMultiDayGapBreak_RehearsalR06(t *testing.T) {
	// User had a 12-day streak (longest 12) but the last 4 days are
	// missed. Even with a full grace balance (2), the streak breaks and
	// a fresh streak of 1 begins today.
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	lastGood := atLocalDate(t, "2026-07-22")
	state := StreakState{
		CurrentStreakCount:     12,
		LongestStreakCount:     12,
		LastCompletedLocalDate: &lastGood,
		Timezone:               "UTC",
		Status:                 StreakStatusBroken,
	}
	snapshots := []StreakSnapshot{
		{LocalDate: lastGood, Status: MissionStatusCompleted},
		{LocalDate: atLocalDate(t, "2026-07-23"), Status: MissionStatusMissed},
		{LocalDate: atLocalDate(t, "2026-07-24"), Status: MissionStatusMissed},
		{LocalDate: atLocalDate(t, "2026-07-25"), Status: MissionStatusMissed},
	}
	rec, err := ReconcileStreak("user-r06", now, state, GraceBalance{Balance: 2}, snapshots, true)
	require.NoError(t, err)
	assert.Equal(t, 1, rec.NewState.CurrentStreakCount, "multi-day gap: streak resets to 1 (today's completion)")
	assert.Equal(t, 12, rec.NewState.LongestStreakCount, "longest preserved")
	assert.Equal(t, StreakStatusActive, rec.NewState.Status)
	assert.Nil(t, rec.GraceDayUsed, "multi-day gap: no grace consumed")
	assert.Nil(t, rec.GraceDayEarned, "streak of 1, not 7: no grace earned")
}
