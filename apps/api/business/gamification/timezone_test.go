package gamification

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveSettingsStoredNonDefault(t *testing.T) {
	res, err := ResolveSettings(UserSettingsSource{
		Stored:            true,
		Timezone:          "Europe/Berlin",
		DailyReviewTarget: 30,
	}, "")
	require.NoError(t, err)
	assert.Equal(t, "Europe/Berlin", res.Timezone)
	assert.Equal(t, 30, res.DailyReviewTarget)
}

func TestResolveSettingsStoredDefaultFallsThroughToClientTimezone(t *testing.T) {
	res, err := ResolveSettings(UserSettingsSource{
		Stored:            true,
		Timezone:          DefaultTimezone,
		DailyReviewTarget: 20,
	}, "Europe/Berlin")
	require.NoError(t, err)
	assert.Equal(t, "Europe/Berlin", res.Timezone)
	assert.Equal(t, DefaultDailyReviewTarget, res.DailyReviewTarget)
}

func TestResolveSettingsNoStoredFallsThroughToClientTimezone(t *testing.T) {
	res, err := ResolveSettings(UserSettingsSource{Stored: false}, "America/New_York")
	require.NoError(t, err)
	assert.Equal(t, "America/New_York", res.Timezone)
	assert.Equal(t, DefaultDailyReviewTarget, res.DailyReviewTarget)
}

func TestResolveSettingsNoStoredNoClientFallsToDefaults(t *testing.T) {
	res, err := ResolveSettings(UserSettingsSource{Stored: false}, "")
	require.NoError(t, err)
	assert.Equal(t, DefaultTimezone, res.Timezone)
	assert.Equal(t, DefaultDailyReviewTarget, res.DailyReviewTarget)
}

func TestResolveSettingsRejectsInvalidClientTimezone(t *testing.T) {
	_, err := ResolveSettings(UserSettingsSource{Stored: false}, "Not/A/Real/Zone")
	require.ErrorIs(t, err, ErrInvalidTimezone)
}

func TestResolveSettingsRejectsInvalidStoredTimezone(t *testing.T) {
	_, err := ResolveSettings(UserSettingsSource{
		Stored:            true,
		Timezone:          "Not/A/Real/Zone",
		DailyReviewTarget: 20,
	}, "")
	require.ErrorIs(t, err, ErrInvalidTimezone)
}

func TestResolveSettingsRejectsStoredTargetOutOfRange(t *testing.T) {
	_, err := ResolveSettings(UserSettingsSource{
		Stored:            true,
		Timezone:          "Europe/Berlin",
		DailyReviewTarget: 4,
	}, "")
	require.Error(t, err)
	_, err = ResolveSettings(UserSettingsSource{
		Stored:            true,
		Timezone:          "Europe/Berlin",
		DailyReviewTarget: 101,
	}, "")
	require.Error(t, err)
}

func TestIsValidIANATimezone(t *testing.T) {
	assert.True(t, IsValidIANATimezone("UTC"))
	assert.True(t, IsValidIANATimezone("Europe/Berlin"))
	assert.True(t, IsValidIANATimezone("America/Los_Angeles"))
	assert.False(t, IsValidIANATimezone(""))
	assert.False(t, IsValidIANATimezone("Not/A/Real/Zone"))
}

func TestLocalDate(t *testing.T) {
	// 2026-07-26 23:30 UTC == 2026-07-27 01:30 in Europe/Berlin.
	now := time.Date(2026, 7, 26, 23, 30, 0, 0, time.UTC)
	d, err := LocalDate(now, "Europe/Berlin")
	require.NoError(t, err)
	assert.Equal(t, 2026, d.Year())
	assert.Equal(t, time.July, d.Month())
	assert.Equal(t, 27, d.Day())

	d2, err := LocalDate(now, "UTC")
	require.NoError(t, err)
	assert.Equal(t, 26, d2.Day())

	_, err = LocalDate(now, "Not/A/Real/Zone")
	require.ErrorIs(t, err, ErrInvalidTimezone)
}

func TestLocalDateYesterday(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	d, err := LocalDateYesterday(now, "UTC")
	require.NoError(t, err)
	assert.Equal(t, 2026, d.Year())
	assert.Equal(t, time.July, d.Month())
	assert.Equal(t, 25, d.Day())
}
