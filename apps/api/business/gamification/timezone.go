package gamification

import (
	"errors"
	"fmt"
	"time"
)

// DefaultTimezone is the IANA fallback used when no per-user row exists and no
// request-time client timezone was supplied. It matches the user_settings
// schema default in DOC-05 §6.
const DefaultTimezone = "UTC"

// DefaultDailyReviewTarget is the user_settings schema default for
// daily_review_target (DOC-05 §6). All P4 review-target reads fall back to it
// when no per-user stored value is set.
const DefaultDailyReviewTarget = 20

// MinDailyReviewTarget and MaxDailyReviewTarget mirror the user_settings and
// daily_mission_snapshots check constraints (5..100). They are exported so the
// API/Huma layer can echo the same bounds.
const (
	MinDailyReviewTarget = 5
	MaxDailyReviewTarget = 100
)

// ResolvedSettings is the result of resolving a per-user timezone and daily
// review target. It is the input the missions module uses to lazily create
// daily_mission_snapshots and what the public reads return.
type ResolvedSettings struct {
	Timezone          string
	DailyReviewTarget int
}

// UserSettingsSource is the minimal per-user settings view that the
// timezone/target resolver needs. The missions module fetches this from the
// user_settings table (lazily creating the row on first use).
type UserSettingsSource struct {
	Stored            bool
	Timezone          string
	DailyReviewTarget int
}

// ErrInvalidTimezone is returned by ResolveSettings when the request-time
// client-supplied IANA timezone fails validation. Per D01, an unrecognized
// client value is rejected, not silently defaulted.
var ErrInvalidTimezone = errors.New("invalid IANA timezone")

// ResolveSettings implements the D01 timezone/target resolution chain:
//
//  1. If the per-user user_settings row has a stored non-default value, use it.
//  2. Else, if a request-time client-supplied IANA timezone is provided, validate
//     it against the IANA timezone database and use it (with the user_settings
//     schema default daily_review_target).
//  3. Else, fall back to UTC and the user_settings default of 20 reviews.
//
// This function is pure (no IO) and deterministic. The caller's repository
// layer is responsible for reading the user_settings row and calling
// LoadLocation on the request-time client timezone before invoking this.
func ResolveSettings(stored UserSettingsSource, clientTimezone string) (ResolvedSettings, error) {
	if stored.Stored && stored.Timezone != "" && stored.Timezone != DefaultTimezone {
		if _, err := time.LoadLocation(stored.Timezone); err != nil {
			return ResolvedSettings{}, fmt.Errorf("%w: stored %q: %v", ErrInvalidTimezone, stored.Timezone, err)
		}
		loc := stored.Timezone
		target := stored.DailyReviewTarget
		if target <= 0 {
			target = DefaultDailyReviewTarget
		}
		if target < MinDailyReviewTarget || target > MaxDailyReviewTarget {
			return ResolvedSettings{}, fmt.Errorf("stored daily review target %d out of range [%d,%d]", target, MinDailyReviewTarget, MaxDailyReviewTarget)
		}
		return ResolvedSettings{Timezone: loc, DailyReviewTarget: target}, nil
	}
	if clientTimezone != "" {
		if _, err := time.LoadLocation(clientTimezone); err != nil {
			return ResolvedSettings{}, fmt.Errorf("%w: client %q: %v", ErrInvalidTimezone, clientTimezone, err)
		}
		return ResolvedSettings{Timezone: clientTimezone, DailyReviewTarget: DefaultDailyReviewTarget}, nil
	}
	return ResolvedSettings{Timezone: DefaultTimezone, DailyReviewTarget: DefaultDailyReviewTarget}, nil
}

// LocalDate returns the calendar date for a given instant in the given IANA
// timezone. The caller must have already validated the timezone. The
// returned time.Time is a midnight UTC representation of the local date so it
// can be stored as a `date` column reliably.
func LocalDate(now time.Time, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %q: %v", ErrInvalidTimezone, timezone, err)
	}
	local := now.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC), nil
}

// LocalDateYesterday is a convenience used by streak reconciliation to compute
// the prior local date.
func LocalDateYesterday(now time.Time, timezone string) (time.Time, error) {
	today, err := LocalDate(now, timezone)
	if err != nil {
		return time.Time{}, err
	}
	return today.AddDate(0, 0, -1), nil
}

// IsValidIANATimezone returns true if the IANA database recognizes the given
// name. This is the boundary the API layer uses to reject a malformed
// client-supplied timezone before it ever reaches the daily-date math.
func IsValidIANATimezone(name string) bool {
	if name == "" {
		return false
	}
	_, err := time.LoadLocation(name)
	return err == nil
}
