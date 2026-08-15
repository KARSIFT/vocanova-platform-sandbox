package gamification

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func atLocalDate(t *testing.T, d string) time.Time {
	t.Helper()
	v, err := time.Parse("2006-01-02", d)
	require.NoError(t, err)
	return v
}

func TestReconcileStreakNoOpWhenTodayAlreadyCompleted(t *testing.T) {
	// Rule (a): today is already completed → no-op, no state change.
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	today := atLocalDate(t, "2026-07-26")
	state := StreakState{
		CurrentStreakCount:     5,
		LongestStreakCount:     7,
		LastCompletedLocalDate: &today,
		Timezone:               "UTC",
		Status:                 StreakStatusActive,
	}
	snapshots := []StreakSnapshot{
		{LocalDate: today, Status: MissionStatusCompleted},
	}
	rec, err := ReconcileStreak("user-1", now, state, GraceBalance{Balance: 1}, snapshots, false)
	require.NoError(t, err)
	assert.Equal(t, 5, rec.NewState.CurrentStreakCount)
	assert.Equal(t, 7, rec.NewState.LongestStreakCount)
	assert.Equal(t, StreakStatusActive, rec.NewState.Status)
	assert.Nil(t, rec.GraceDayEarned)
	assert.Nil(t, rec.GraceDayUsed)
}

func TestReconcileStreakFirstEverCompletionStartsStreak(t *testing.T) {
	// A brand-new user completes their first mission today: streak = 1,
	// longest = 1, no grace day to apply.
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	state := StreakState{Timezone: "UTC"}
	rec, err := ReconcileStreak("user-1", now, state, GraceBalance{Balance: 0}, nil, true)
	require.NoError(t, err)
	assert.Equal(t, 1, rec.NewState.CurrentStreakCount)
	assert.Equal(t, 1, rec.NewState.LongestStreakCount)
	assert.Equal(t, StreakStatusActive, rec.NewState.Status)
	assert.Nil(t, rec.GraceDayEarned)
}

func TestReconcileStreakAdvancesOnConsecutiveCompletion(t *testing.T) {
	// Rule (b): yesterday completed, today completed → streak advances.
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	yesterday := atLocalDate(t, "2026-07-25")
	state := StreakState{
		CurrentStreakCount:     3,
		LongestStreakCount:     3,
		LastCompletedLocalDate: &yesterday,
		Timezone:               "UTC",
		Status:                 StreakStatusActive,
	}
	snapshots := []StreakSnapshot{
		{LocalDate: yesterday, Status: MissionStatusCompleted},
	}
	rec, err := ReconcileStreak("user-1", now, state, GraceBalance{Balance: 0}, snapshots, true)
	require.NoError(t, err)
	assert.Equal(t, 4, rec.NewState.CurrentStreakCount)
	assert.Equal(t, 4, rec.NewState.LongestStreakCount)
	assert.Equal(t, StreakStatusActive, rec.NewState.Status)
}

func TestReconcileStreakAdvancesWhenTodayAlreadyCompletedInSnapshots(t *testing.T) {
	// Regression (VOC-082): after MarkSnapshotCompleted in the same call,
	// the fetched snapshot list includes today as completed while
	// currentCompletion=true. This must advance the streak using
	// yesterday as the anchor, not return ErrInvalidStreakSnapshot.
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	today := atLocalDate(t, "2026-07-26")
	yesterday := atLocalDate(t, "2026-07-25")
	state := StreakState{
		CurrentStreakCount:     3,
		LongestStreakCount:     3,
		LastCompletedLocalDate: &yesterday,
		Timezone:               "UTC",
		Status:                 StreakStatusActive,
	}
	snapshots := []StreakSnapshot{
		{LocalDate: yesterday, Status: MissionStatusCompleted},
		{LocalDate: today, Status: MissionStatusCompleted},
	}
	rec, err := ReconcileStreak("user-1", now, state, GraceBalance{Balance: 0}, snapshots, true)
	require.NoError(t, err)
	assert.Equal(t, 4, rec.NewState.CurrentStreakCount)
	assert.Equal(t, 4, rec.NewState.LongestStreakCount)
	assert.Equal(t, StreakStatusActive, rec.NewState.Status)
}

func TestReconcileStreakFirstCompletionWhenTodayInSnapshots(t *testing.T) {
	// Regression (VOC-082): first-ever completion with today already
	// marked completed in the fetched list starts a streak of 1.
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	today := atLocalDate(t, "2026-07-26")
	state := StreakState{Timezone: "UTC"}
	snapshots := []StreakSnapshot{
		{LocalDate: today, Status: MissionStatusCompleted},
	}
	rec, err := ReconcileStreak("user-1", now, state, GraceBalance{Balance: 0}, snapshots, true)
	require.NoError(t, err)
	assert.Equal(t, 1, rec.NewState.CurrentStreakCount)
	assert.Equal(t, 1, rec.NewState.LongestStreakCount)
	assert.Equal(t, StreakStatusActive, rec.NewState.Status)
}

func TestReconcileStreakRejectsFutureLastGood(t *testing.T) {
	// VOC-082-D01: genuinely future completed snapshots remain rejected.
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	tomorrow := atLocalDate(t, "2026-07-27")
	state := StreakState{Timezone: "UTC", Status: StreakStatusActive}
	snapshots := []StreakSnapshot{
		{LocalDate: tomorrow, Status: MissionStatusCompleted},
	}
	_, err := ReconcileStreak("user-1", now, state, GraceBalance{Balance: 0}, snapshots, true)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidStreakSnapshot)
}

func TestReconcileStreakEarnsGraceDayAt7(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	yesterday := atLocalDate(t, "2026-07-25")
	state := StreakState{
		CurrentStreakCount:     6,
		LongestStreakCount:     6,
		LastCompletedLocalDate: &yesterday,
		Timezone:               "UTC",
		Status:                 StreakStatusActive,
	}
	snapshots := []StreakSnapshot{{LocalDate: yesterday, Status: MissionStatusCompleted}}
	rec, err := ReconcileStreak("user-1", now, state, GraceBalance{Balance: 0}, snapshots, true)
	require.NoError(t, err)
	assert.Equal(t, 7, rec.NewState.CurrentStreakCount)
	require.NotNil(t, rec.GraceDayEarned)
	assert.Equal(t, 1, rec.GraceDayEarned.Amount)
	assert.Equal(t, 1, rec.GraceDayEarned.BalanceAfter)
	assert.Equal(t, GraceReasonEarnedByStreak, rec.GraceDayEarned.Reason)
}

func TestReconcileStreakDoesNotEarnGraceDayWhenCapped(t *testing.T) {
	// Cap is 2: if balance is already at 2, do not earn even when streak
	// hits 7.
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	yesterday := atLocalDate(t, "2026-07-25")
	state := StreakState{
		CurrentStreakCount:     6,
		LongestStreakCount:     6,
		LastCompletedLocalDate: &yesterday,
		Timezone:               "UTC",
		Status:                 StreakStatusActive,
	}
	snapshots := []StreakSnapshot{{LocalDate: yesterday, Status: MissionStatusCompleted}}
	rec, err := ReconcileStreak("user-1", now, state, GraceBalance{Balance: 2}, snapshots, true)
	require.NoError(t, err)
	assert.Equal(t, 7, rec.NewState.CurrentStreakCount)
	assert.Nil(t, rec.GraceDayEarned)
}

func TestReconcileStreakAtRiskWhenYesterdayDoneButTodayNotYet(t *testing.T) {
	// Read-only reconciliation: yesterday was good, today not yet
	// completed, no current completion signaled. Streak is at risk.
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	yesterday := atLocalDate(t, "2026-07-25")
	state := StreakState{
		CurrentStreakCount:     3,
		LongestStreakCount:     3,
		LastCompletedLocalDate: &yesterday,
		Timezone:               "UTC",
		Status:                 StreakStatusActive,
	}
	snapshots := []StreakSnapshot{{LocalDate: yesterday, Status: MissionStatusCompleted}}
	rec, err := ReconcileStreak("user-1", now, state, GraceBalance{Balance: 0}, snapshots, false)
	require.NoError(t, err)
	assert.Equal(t, 3, rec.NewState.CurrentStreakCount)
	assert.Equal(t, StreakStatusAtRisk, rec.NewState.Status)
}

func TestReconcileStreakConsumesGraceWhenYesterdayMissed(t *testing.T) {
	// Rule (c): yesterday missed, grace day available → consume and
	// advance streak. The state knows the last good day was 2 days ago
	// (we have a 4-day streak); yesterday was missed, today is being
	// completed with grace available.
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	yesterday := atLocalDate(t, "2026-07-25")
	twoDaysAgo := atLocalDate(t, "2026-07-24")
	state := StreakState{
		CurrentStreakCount:     4,
		LongestStreakCount:     4,
		LastCompletedLocalDate: &twoDaysAgo,
		Timezone:               "UTC",
		Status:                 StreakStatusAtRisk,
	}
	snapshots := []StreakSnapshot{{LocalDate: yesterday, Status: MissionStatusMissed}}
	rec, err := ReconcileStreak("user-1", now, state, GraceBalance{Balance: 1}, snapshots, true)
	require.NoError(t, err)
	assert.Equal(t, 5, rec.NewState.CurrentStreakCount, "streak continues through grace-protected day")
	assert.Equal(t, StreakStatusActive, rec.NewState.Status)
	require.NotNil(t, rec.GraceDayUsed)
	assert.Equal(t, -1, rec.GraceDayUsed.Amount)
	assert.Equal(t, 0, rec.GraceDayUsed.BalanceAfter)
	assert.Equal(t, GraceReasonUsedForMissedDay, rec.GraceDayUsed.Reason)
	assert.True(t, rec.YesterdayWasMissed)
	require.NotNil(t, rec.YesterdayProtectedLocalDate)
	assert.Equal(t, yesterday, *rec.YesterdayProtectedLocalDate)
}

func TestReconcileStreakBreaksWhenNoGraceAndYesterdayMissed(t *testing.T) {
	// Rule (d): yesterday missed and no grace → break streak. Today's
	// completion starts a new streak of 1.
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	yesterday := atLocalDate(t, "2026-07-25")
	twoDaysAgo := atLocalDate(t, "2026-07-24")
	state := StreakState{
		CurrentStreakCount:     4,
		LongestStreakCount:     10,
		LastCompletedLocalDate: &twoDaysAgo,
		Timezone:               "UTC",
		Status:                 StreakStatusAtRisk,
	}
	snapshots := []StreakSnapshot{{LocalDate: yesterday, Status: MissionStatusMissed}}
	rec, err := ReconcileStreak("user-1", now, state, GraceBalance{Balance: 0}, snapshots, true)
	require.NoError(t, err)
	assert.Equal(t, 1, rec.NewState.CurrentStreakCount, "broken streak, new completion starts at 1")
	assert.Equal(t, 10, rec.NewState.LongestStreakCount, "longest preserved across the break")
	assert.Equal(t, StreakStatusActive, rec.NewState.Status)
}

func TestReconcileStreakBreaksOnMultiDayGap(t *testing.T) {
	// Rule (d): 3+ days missed (last good is 3 days ago) → break even
	// with grace available, because a single grace day protects at most
	// one missed day.
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	threeDaysAgo := atLocalDate(t, "2026-07-23")
	state := StreakState{
		CurrentStreakCount: 4,
		LongestStreakCount: 10,
		Timezone:           "UTC",
		Status:             StreakStatusAtRisk,
	}
	snapshots := []StreakSnapshot{
		{LocalDate: threeDaysAgo, Status: MissionStatusCompleted},
		{LocalDate: atLocalDate(t, "2026-07-24"), Status: MissionStatusMissed},
		{LocalDate: atLocalDate(t, "2026-07-25"), Status: MissionStatusMissed},
	}
	rec, err := ReconcileStreak("user-1", now, state, GraceBalance{Balance: 2}, snapshots, true)
	require.NoError(t, err)
	assert.Equal(t, 1, rec.NewState.CurrentStreakCount, "broken, new completion starts at 1")
	assert.Equal(t, 10, rec.NewState.LongestStreakCount)
	assert.Nil(t, rec.GraceDayUsed, "multi-day gap: no grace consumed")
}

func TestReconcileStreakReadOnlyRespectsAtRisk(t *testing.T) {
	// Read-only: yesterday missed, grace available, no current completion
	// → still flagged at_risk, no consumed grace yet, but yesterday
	// protected intent is signaled.
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	yesterday := atLocalDate(t, "2026-07-25")
	twoDaysAgo := atLocalDate(t, "2026-07-24")
	state := StreakState{
		CurrentStreakCount:     4,
		LongestStreakCount:     4,
		LastCompletedLocalDate: &twoDaysAgo,
		Timezone:               "UTC",
		Status:                 StreakStatusAtRisk,
	}
	snapshots := []StreakSnapshot{{LocalDate: yesterday, Status: MissionStatusMissed}}
	rec, err := ReconcileStreak("user-1", now, state, GraceBalance{Balance: 1}, snapshots, false)
	require.NoError(t, err)
	assert.Equal(t, 4, rec.NewState.CurrentStreakCount)
	assert.Equal(t, StreakStatusAtRisk, rec.NewState.Status)
	assert.Nil(t, rec.GraceDayUsed, "no grace consumed without a completion")
}

func TestReconcileStreakProtectedSnapshotCountsAsGood(t *testing.T) {
	// A previously protected (grace-applied) day is treated as good for
	// the purposes of "last good day" detection.
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	yesterday := atLocalDate(t, "2026-07-25")
	state := StreakState{
		CurrentStreakCount:     4,
		LongestStreakCount:     4,
		LastCompletedLocalDate: &yesterday,
		Timezone:               "UTC",
		Status:                 StreakStatusActive,
	}
	snapshots := []StreakSnapshot{{LocalDate: yesterday, Status: MissionStatusProtected}}
	rec, err := ReconcileStreak("user-1", now, state, GraceBalance{Balance: 0}, snapshots, true)
	require.NoError(t, err)
	assert.Equal(t, 5, rec.NewState.CurrentStreakCount, "protected day counts as completed for streak")
	assert.Equal(t, StreakStatusActive, rec.NewState.Status)
}

func TestReconcileStreakIdempotencyKeysContainUserAndDate(t *testing.T) {
	// The grace-day ledger entries' idempotency keys must include user id
	// and local date so they are unique per (user, day, kind).
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	yesterday := atLocalDate(t, "2026-07-25")
	state := StreakState{
		CurrentStreakCount: 6,
		LongestStreakCount: 6,
		Timezone:           "UTC",
		Status:             StreakStatusActive,
	}
	snapshots := []StreakSnapshot{{LocalDate: yesterday, Status: MissionStatusCompleted}}
	rec, err := ReconcileStreak("user-abc", now, state, GraceBalance{Balance: 0}, snapshots, true)
	require.NoError(t, err)
	require.NotNil(t, rec.GraceDayEarned)
	assert.Equal(t, "streak:user-abc:2026-07-26:grace_day_earned", rec.GraceDayEarned.IdempotencyKey.String())
}
