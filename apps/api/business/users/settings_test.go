package users

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSettingsUpdateValidateAcceptsEmpty verifies that a fully-empty
// payload (a PATCH with no fields) passes validation. The service
// layer treats an empty update as a no-op read, so the absence of
// any field-level value must not be reported as an error.
func TestSettingsUpdateValidateAcceptsEmpty(t *testing.T) {
	require.NoError(t, SettingsUpdate{}.Validate())
}

// TestSettingsUpdateIsEmpty pins the helper the service layer uses
// to short-circuit an empty PATCH.
func TestSettingsUpdateIsEmpty(t *testing.T) {
	assert.True(t, SettingsUpdate{}.IsEmpty())
	target := 30
	assert.False(t, SettingsUpdate{DailyReviewTarget: &target}.IsEmpty())
}

// TestSettingsUpdateValidateAcceptsBoundaryDailyReviewTarget covers
// the exact 5..100 inclusive bounds the user_settings check
// constraint enforces. Both boundary values are accepted.
func TestSettingsUpdateValidateAcceptsBoundaryDailyReviewTarget(t *testing.T) {
	for _, value := range []int{MinDailyReviewTarget, MaxDailyReviewTarget} {
		v := value
		require.NoErrorf(t, SettingsUpdate{DailyReviewTarget: &v}.Validate(), "value %d", v)
	}
}

// TestSettingsUpdateValidateRejectsOutOfRangeDailyReviewTarget
// covers VOC-031-TEST-08: every value outside the 5..100 bound is
// rejected with a clear error message.
func TestSettingsUpdateValidateRejectsOutOfRangeDailyReviewTarget(t *testing.T) {
	for _, value := range []int{0, 1, MinDailyReviewTarget - 1, MaxDailyReviewTarget + 1, 1000, -5} {
		v := value
		err := SettingsUpdate{DailyReviewTarget: &v}.Validate()
		require.Errorf(t, err, "value %d must be rejected", v)
		assert.Contains(t, err.Error(), "daily review target")
	}
}

// TestSettingsUpdateValidateAcceptsEveryReviewIntervalPreset pins
// the enum accepted by the API layer. The set must match the
// user_settings.review_interval_preset check constraint exactly.
func TestSettingsUpdateValidateAcceptsEveryReviewIntervalPreset(t *testing.T) {
	for _, value := range []string{ReviewIntervalPresetVocabDefault, ReviewIntervalPresetWordUpLike, ReviewIntervalPresetCustom} {
		v := value
		require.NoErrorf(t, SettingsUpdate{ReviewIntervalPreset: &v}.Validate(), "value %s", v)
	}
}

// TestSettingsUpdateValidateRejectsUnknownReviewIntervalPreset
// covers VOC-031-TEST-08: any value outside the
// vocanova_default|wordup_like|custom enum is rejected.
func TestSettingsUpdateValidateRejectsUnknownReviewIntervalPreset(t *testing.T) {
	for _, value := range []string{"", "anki", "default", "VocabDefault"} {
		v := value
		err := SettingsUpdate{ReviewIntervalPreset: &v}.Validate()
		require.Errorf(t, err, "value %q must be rejected", v)
		assert.Contains(t, err.Error(), "review interval preset")
	}
}

// TestSettingsUpdateValidateAcceptsSupportedAppLanguage pins
// VOC-031-D06: only "en" is accepted at launch. The function does
// not allow a multi-language picker.
func TestSettingsUpdateValidateAcceptsSupportedAppLanguage(t *testing.T) {
	for _, value := range SupportedAppLanguages {
		v := value
		require.NoErrorf(t, SettingsUpdate{AppLanguage: &v}.Validate(), "value %s", v)
	}
}

// TestSettingsUpdateValidateRejectsUnsupportedAppLanguage covers
// VOC-031-TEST-08 and the D06 boundary: every non-en value
// currently in the schema's regex (^[A-Za-z]{2,8}$) is rejected
// at the service layer so the API never claims a capability the
// product does not have.
func TestSettingsUpdateValidateRejectsUnsupportedAppLanguage(t *testing.T) {
	for _, value := range []string{"es", "de", "fr", "EN", "english", "1234"} {
		v := value
		err := SettingsUpdate{AppLanguage: &v}.Validate()
		require.Errorf(t, err, "value %q must be rejected", v)
		assert.Contains(t, err.Error(), "app language")
	}
}

// TestSettingsUpdateValidateAcceptsBooleanFields verifies that
// every bool value (true and false) passes validation, since the
// HTTP wire format cannot omit a non-nullable bool and a caller
// must be able to write either value.
func TestSettingsUpdateValidateAcceptsBooleanFields(t *testing.T) {
	for _, value := range []bool{true, false} {
		v := value
		require.NoError(t, SettingsUpdate{NotificationsEnabled: &v}.Validate())
		require.NoError(t, SettingsUpdate{MarketingEmailsEnabled: &v}.Validate())
	}
}

// TestSettingsUpdateValidateRejectsOverlongDisplayName covers
// VOC-031-TEST-08: displayName beyond MaxDisplayNameLength is
// rejected with a clear error.
func TestSettingsUpdateValidateRejectsOverlongDisplayName(t *testing.T) {
	overlong := make([]byte, MaxDisplayNameLength+1)
	for i := range overlong {
		overlong[i] = 'a'
	}
	name := string(overlong)
	err := SettingsUpdate{DisplayName: &name}.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "display name")
}

// TestSettingsUpdateValidateAcceptsMaxLengthDisplayName verifies
// the inclusive length boundary: a display name of exactly
// MaxDisplayNameLength characters is accepted.
func TestSettingsUpdateValidateAcceptsMaxLengthDisplayName(t *testing.T) {
	exactlyMax := make([]byte, MaxDisplayNameLength)
	for i := range exactlyMax {
		exactlyMax[i] = 'a'
	}
	name := string(exactlyMax)
	require.NoError(t, SettingsUpdate{DisplayName: &name}.Validate())
}

// TestSettingsUpdateValidateAcceptsEmptyDisplayName pins that a
// learner can clear their display name (write an empty string).
// The field is learner-controlled; "remove my display name" is a
// legitimate value.
func TestSettingsUpdateValidateAcceptsEmptyDisplayName(t *testing.T) {
	name := ""
	require.NoError(t, SettingsUpdate{DisplayName: &name}.Validate())
}

// TestSettingsUpdateValidateRejectsUnknownFieldValues combines
// every validation rule in one test so a regression that drops
// any one of them is visible at a glance.
func TestSettingsUpdateValidateRejectsUnknownFieldValues(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*SettingsUpdate)
		wantSub string
	}{
		{"daily-review-low", func(u *SettingsUpdate) {
			v := MinDailyReviewTarget - 1
			u.DailyReviewTarget = &v
		}, "daily review target"},
		{"daily-review-high", func(u *SettingsUpdate) {
			v := MaxDailyReviewTarget + 1
			u.DailyReviewTarget = &v
		}, "daily review target"},
		{"review-interval", func(u *SettingsUpdate) {
			v := "anki"
			u.ReviewIntervalPreset = &v
		}, "review interval preset"},
		{"app-language", func(u *SettingsUpdate) {
			v := "es"
			u.AppLanguage = &v
		}, "app language"},
		{"display-name-too-long", func(u *SettingsUpdate) {
			overlong := make([]byte, MaxDisplayNameLength+1)
			for i := range overlong {
				overlong[i] = 'a'
			}
			v := string(overlong)
			u.DisplayName = &v
		}, "display name"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := SettingsUpdate{}
			tc.mutate(&u)
			err := u.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantSub)
		})
	}
}
