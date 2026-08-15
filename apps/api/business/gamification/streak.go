package gamification

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// StreakSnapshot is the minimal view of a user's daily-mission history the
// reconciliation algorithm needs. The repository layer reads this from
// daily_mission_snapshots and passes it in; the function itself is pure.
type StreakSnapshot struct {
	LocalDate    time.Time // date-only; UTC midnight
	Status       string    // open / completed / missed / protected
	CompletedAt  *time.Time
	GraceApplied bool
	GraceDayID   *string
}

// StreakState is the minimal view of a user's streak_states row. The
// repository layer reads this; reconciliation returns the new state to
// apply.
type StreakState struct {
	CurrentStreakCount     int
	LongestStreakCount     int
	LastCompletedLocalDate *time.Time
	LastActivityLocalDate  *time.Time
	Timezone               string
	Status                 string // active / at_risk / broken
}

// GraceBalance is the user's current available grace-day balance. The
// repository layer reads this from the latest grace_day_ledger row (or
// 0 if no rows exist).
type GraceBalance struct {
	Balance int
}

// StreakReconciliation is the deterministic output of reconciling the user's
// streak at a moment in time. The transaction layer writes back the new
// streak_states row and, when a transition triggers one, a grace-day ledger
// entry (EarnedByStreak or UsedForMissedDay) with the supplied idempotency
// key.
type StreakReconciliation struct {
	// NewState is the new streak state to persist.
	NewState StreakState
	// GraceDayEarned is non-nil when this reconciliation earns a new grace
	// day and a grace_day_ledger row should be inserted.
	GraceDayEarned *GraceLedgerEntry
	// GraceDayUsed is non-nil when this reconciliation consumed an
	// available grace day to protect a missed day and a grace_day_ledger
	// row should be inserted (with a negative amount).
	GraceDayUsed *GraceLedgerEntry
	// YesterdayProtectedLocalDate is set when the reconciliation decided
	// to apply a grace day to yesterday (so the caller can mark yesterday's
	// daily_mission_snapshot.status='protected' and grace_applied=true).
	YesterdayProtectedLocalDate *time.Time
	// YesterdaySnapshotID is the snapshot id of yesterday that should be
	// marked protected. Nil when no protection happened.
	YesterdaySnapshotID *string
	// YesterdayWasMissed is true when yesterday's snapshot was in
	// status='missed' (i.e. the day was genuinely missed). Used by the
	// caller to decide whether to mark it protected.
	YesterdayWasMissed bool
}

// GraceLedgerEntry is the per-streak-write grace_day_ledger insert. The
// transaction layer turns it into a real row.
type GraceLedgerEntry struct {
	UserID             string
	Amount             int
	BalanceAfter       int
	Reason             string
	SourceType         string
	SourceID           *uuid.UUID
	AppliedToLocalDate time.Time
	Timezone           string
	IdempotencyKey     GraceIdempotencyKey
}

// Streak errors.
var (
	ErrInvalidStreakSnapshot = errors.New("invalid streak snapshot")
)

// ReconcileStreak is the pure streak-reconciliation algorithm. It takes the
// current streak state, the user's available grace-day balance, and a window
// of daily-mission snapshots ending at today (in the resolved timezone), and
// returns the deterministic new state. Per DOC-05 §12, transitions are
// computed from daily_mission_snapshots, lazily at read/write time (no
// queue/cron — DOC-06 §15). Rules:
//
//   - (a) today is already completed → no-op
//   - (b) yesterday completed → today's completion (signaled by currentCompletion)
//     advances the streak by one
//   - (c) yesterday missed and a grace day is available and unused → consume
//     one grace day, mark yesterday's snapshot status='protected'
//     grace_applied=true (signaled by currentCompletion); the streak
//     continues uninterrupted
//   - (d) yesterday missed with no grace day, or two or more days missed →
//     current_streak_count resets to 0, status='broken'
//
// currentCompletion indicates the user just completed today's mission in the
// same call. Without it, the function is a read-only reconciliation (still
// triggers (c) if applicable, but does not advance the streak or earn a grace
// day).
func ReconcileStreak(
	userID string,
	now time.Time,
	state StreakState,
	grace GraceBalance,
	snapshots []StreakSnapshot,
	currentCompletion bool,
) (StreakReconciliation, error) {
	today, err := LocalDate(now, state.Timezone)
	if err != nil {
		return StreakReconciliation{}, err
	}
	yesterday := today.AddDate(0, 0, -1)

	idx := buildSnapshotIndex(snapshots)
	todaySnap, hasToday := idx[dateKey(today)]
	yesterdaySnap, hasYesterday := idx[dateKey(yesterday)]

	if hasToday && todaySnap.Status == MissionStatusCompleted && !currentCompletion {
		// (a) no-op; today's mission already complete, streak stays.
		unchanged := state
		unchanged.LastActivityLocalDate = &today
		if unchanged.LastCompletedLocalDate == nil {
			unchanged.LastCompletedLocalDate = &today
		}
		unchanged.Status = StreakStatusActive
		return StreakReconciliation{NewState: unchanged}, nil
	}

	// Two-or-more-days-missed detection: find the most recent completed or
	// protected snapshot. If its local date is more than 1 day before
	// today, the streak is broken (rule d) regardless of grace day
	// availability — a single grace day protects one missed day only.
	lastGood := mostRecentGood(snapshots)
	if lastGood == nil && state.LastCompletedLocalDate != nil {
		// The state row is the only source of the last good day. This
		// is the normal case for a learner who has been using the
		// product long enough that the daily-mission history we
		// fetched in this call is incomplete (e.g. the read window
		// is bounded, or older snapshots were archived).
		d := *state.LastCompletedLocalDate
		lastGood = &d
	}
	if lastGood == nil {
		// No prior completion ever. Without a state-stored
		// LastCompletedLocalDate or any in-window snapshot to anchor
		// against, the safest read-only behavior is to preserve the
		// existing state (including its status) and just stamp
		// last_activity_local_date; on a currentCompletion this is
		// the first-ever completion which starts a streak of 1.
		if !currentCompletion {
			unchanged := state
			unchanged.LastActivityLocalDate = &today
			return StreakReconciliation{NewState: unchanged}, nil
		}
		// First ever completion today: streak becomes 1, no grace day
		// to apply, no grace day earned yet (you earn after seven
		// completed days, not on day one).
		newState := state
		newState.CurrentStreakCount = 1
		newState.LongestStreakCount = maxInt(state.LongestStreakCount, 1)
		newState.LastCompletedLocalDate = &today
		newState.LastActivityLocalDate = &today
		newState.Status = StreakStatusActive
		return StreakReconciliation{NewState: newState}, nil
	}

	gap := daysBetween(today, *lastGood)
	if gap <= 0 {
		if currentCompletion && dateKey(*lastGood) == dateKey(today) {
			// The caller just marked today completed in this same call
			// (applyP4ReviewWiring: MarkSnapshotCompleted then fetch snaps).
			// Today is a valid current completion, not an invalid anchor;
			// reconcile using the most recent good day before today.
			if prior := mostRecentGoodBefore(snapshots, today); prior != nil {
				lastGood = prior
			} else if state.LastCompletedLocalDate != nil && state.LastCompletedLocalDate.Before(today) {
				d := *state.LastCompletedLocalDate
				lastGood = &d
			} else {
				// First-ever completion today with no prior anchor.
				newState := state
				newState.CurrentStreakCount = 1
				newState.LongestStreakCount = maxInt(state.LongestStreakCount, 1)
				newState.LastCompletedLocalDate = &today
				newState.LastActivityLocalDate = &today
				newState.Status = StreakStatusActive
				return StreakReconciliation{NewState: newState}, nil
			}
			gap = daysBetween(today, *lastGood)
		}
		if gap <= 0 {
			// Defensive: a future "lastGood" date is impossible under the
			// unique index, but be safe.
			return StreakReconciliation{}, fmt.Errorf("%w: last good date %s is not before today %s", ErrInvalidStreakSnapshot, lastGood.Format("2006-01-02"), today.Format("2006-01-02"))
		}
	}

	newState := state
	newState.LastActivityLocalDate = &today
	if newState.Timezone == "" {
		newState.Timezone = state.Timezone
	}

	rec := StreakReconciliation{}

	if gap == 1 {
		// Yesterday (or one day ago) is the last good day. Today is the
		// next day; streak continues if today is completed.
		if !currentCompletion {
			// Read-only: the streak is at risk if last good is
			// yesterday (we haven't done today yet) but otherwise
			// remains as-is.
			if hasYesterday {
				newState.Status = StreakStatusAtRisk
			} else {
				newState.Status = StreakStatusActive
			}
			return StreakReconciliation{NewState: newState}, nil
		}
		newState.CurrentStreakCount = state.CurrentStreakCount + 1
		newState.LongestStreakCount = maxInt(state.LongestStreakCount, newState.CurrentStreakCount)
		newState.LastCompletedLocalDate = &today
		newState.Status = StreakStatusActive
		// Grace day earned every GraceDayEarnEveryCompletedDays days.
		if newState.CurrentStreakCount > 0 && newState.CurrentStreakCount%GraceDayEarnEveryCompletedDays == 0 && grace.Balance < GraceDayMaxBalance {
			newBalance := grace.Balance + 1
			if newBalance > GraceDayMaxBalance {
				newBalance = GraceDayMaxBalance
			}
			rec.GraceDayEarned = &GraceLedgerEntry{
				UserID:             userID,
				Amount:             1,
				BalanceAfter:       newBalance,
				Reason:             GraceReasonEarnedByStreak,
				SourceType:         GraceSourceStreak,
				AppliedToLocalDate: today,
				Timezone:           state.Timezone,
				IdempotencyKey:     StreakGraceDayEarnedKey(userID, dateKey(today)),
			}
		}
		return StreakReconciliation{NewState: newState, GraceDayEarned: rec.GraceDayEarned}, nil
	}

	// gap >= 2: yesterday was missed (or worse). If gap == 2 and yesterday
	// is in the missed status, we have one missed day. If grace is
	// available, consume it; otherwise break the streak.
	if gap == 2 && hasYesterday && yesterdaySnap.Status == MissionStatusMissed && grace.Balance > 0 {
		if !currentCompletion {
			// No completion yet today; the grace day is held in
			// reserve until today's completion lands. Snapshot
			// status of yesterday is updated lazily by the caller
			// when the read happens — here we just signal intent.
			rec.YesterdayProtectedLocalDate = &yesterday
			rec.YesterdayWasMissed = true
			if yesterdaySnap.GraceDayID != nil {
				id := *yesterdaySnap.GraceDayID
				rec.YesterdaySnapshotID = &id
			}
			newState.Status = StreakStatusAtRisk
			return StreakReconciliation{NewState: newState, YesterdayProtectedLocalDate: &yesterday, YesterdayWasMissed: true, YesterdaySnapshotID: rec.YesterdaySnapshotID}, nil
		}
		// Today completed and grace available: consume one grace day
		// and mark yesterday protected, streak advances.
		newBalance := grace.Balance - 1
		used := GraceLedgerEntry{
			UserID:             userID,
			Amount:             -1,
			BalanceAfter:       newBalance,
			Reason:             GraceReasonUsedForMissedDay,
			SourceType:         GraceSourceStreak,
			AppliedToLocalDate: yesterday,
			Timezone:           state.Timezone,
			IdempotencyKey:     StreakGraceDayUsedKey(userID, dateKey(yesterday)),
		}
		rec.GraceDayUsed = &used
		rec.YesterdayProtectedLocalDate = &yesterday
		rec.YesterdayWasMissed = true
		if yesterdaySnap.GraceDayID != nil {
			id := *yesterdaySnap.GraceDayID
			rec.YesterdaySnapshotID = &id
		}
		newState.CurrentStreakCount = state.CurrentStreakCount + 1
		newState.LongestStreakCount = maxInt(state.LongestStreakCount, newState.CurrentStreakCount)
		newState.LastCompletedLocalDate = &today
		newState.Status = StreakStatusActive
		// Grace day earned this completion as well (the rule is
		// "every 7 completed days", and the protected day counts as
		// a completed day for streak-counting purposes).
		if newState.CurrentStreakCount > 0 && newState.CurrentStreakCount%GraceDayEarnEveryCompletedDays == 0 && newBalance < GraceDayMaxBalance {
			earnedBalance := newBalance + 1
			if earnedBalance > GraceDayMaxBalance {
				earnedBalance = GraceDayMaxBalance
			}
			rec.GraceDayEarned = &GraceLedgerEntry{
				UserID:             userID,
				Amount:             1,
				BalanceAfter:       earnedBalance,
				Reason:             GraceReasonEarnedByStreak,
				SourceType:         GraceSourceStreak,
				AppliedToLocalDate: today,
				Timezone:           state.Timezone,
				IdempotencyKey:     StreakGraceDayEarnedKey(userID, dateKey(today)),
			}
		}
		return StreakReconciliation{NewState: newState, GraceDayEarned: rec.GraceDayEarned, GraceDayUsed: rec.GraceDayUsed, YesterdayProtectedLocalDate: &yesterday, YesterdayWasMissed: true, YesterdaySnapshotID: rec.YesterdaySnapshotID}, nil
	}

	// gap >= 2 with no applicable grace, or gap > 2: streak breaks.
	newState.CurrentStreakCount = 0
	newState.Status = StreakStatusBroken
	if currentCompletion {
		// Today completed after a multi-day gap: this starts a new
		// streak of length 1, but the multi-day gap is already
		// reflected in the reset. Update last completed.
		newState.CurrentStreakCount = 1
		newState.LongestStreakCount = maxInt(state.LongestStreakCount, 1)
		newState.LastCompletedLocalDate = &today
		newState.Status = StreakStatusActive
	}
	return StreakReconciliation{NewState: newState}, nil
}

func stateUserID(_ StreakState) string {
	// The user_id is propagated by the caller via the ReconcileStreak
	// userID parameter; this helper is retained for tests that don't
	// exercise the userID parameter directly.
	return ""
}

func dateKey(d time.Time) string {
	return d.Format("2006-01-02")
}

func buildSnapshotIndex(snaps []StreakSnapshot) map[string]StreakSnapshot {
	out := make(map[string]StreakSnapshot, len(snaps))
	for _, s := range snaps {
		out[dateKey(s.LocalDate)] = s
	}
	return out
}

func mostRecentGood(snaps []StreakSnapshot) *time.Time {
	var best *time.Time
	for _, s := range snaps {
		if s.Status != MissionStatusCompleted && s.Status != MissionStatusProtected {
			continue
		}
		if best == nil || s.LocalDate.After(*best) {
			d := s.LocalDate
			best = &d
		}
	}
	return best
}

func mostRecentGoodBefore(snaps []StreakSnapshot, before time.Time) *time.Time {
	var best *time.Time
	for _, s := range snaps {
		if s.Status != MissionStatusCompleted && s.Status != MissionStatusProtected {
			continue
		}
		if !s.LocalDate.Before(before) {
			continue
		}
		if best == nil || s.LocalDate.After(*best) {
			d := s.LocalDate
			best = &d
		}
	}
	return best
}

func daysBetween(later, earlier time.Time) int {
	if !later.After(earlier) {
		return 0
	}
	diff := later.Sub(earlier)
	return int(diff.Hours() / 24)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
