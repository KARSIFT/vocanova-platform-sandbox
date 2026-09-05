package users

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Settings is the public, requester-owned projection of every editable
// Settings field T02 exposes via GET /api/v1/settings. The fields
// are exactly the founder-directed set: six per-user values on the
// existing user_settings row plus users.display_name. Anything not
// listed here (notably timezone, until a future package asks for it)
// is not part of the public Settings surface and is intentionally
// omitted to keep the contract narrow.
type Settings struct {
	DailyReviewTarget      int
	ReviewIntervalPreset   string
	AppLanguage            string
	NotificationsEnabled   bool
	MarketingEmailsEnabled bool
	DisplayName            string
}

// SettingsUpdate is the partial-update payload PATCH /api/v1/settings
// accepts. Every field is a pointer: a nil pointer means "leave the
// stored value alone"; a non-nil pointer means "write this value,
// validating it first". The discipline matches the PATCH semantics
// in DOC-07 §3 and mirrors the partial-update pattern Huma's
// `optional` + pointer field combination supports.
type SettingsUpdate struct {
	DailyReviewTarget      *int
	ReviewIntervalPreset   *string
	AppLanguage            *string
	NotificationsEnabled   *bool
	MarketingEmailsEnabled *bool
	DisplayName            *string
}

// IsEmpty reports whether the update payload contains no field-level
// changes. The service uses this to skip the round-trip when the
// caller sends an empty PATCH body (DOC-07 §3: a no-op PATCH is a
// well-formed read).
func (u SettingsUpdate) IsEmpty() bool {
	return u.DailyReviewTarget == nil &&
		u.ReviewIntervalPreset == nil &&
		u.AppLanguage == nil &&
		u.NotificationsEnabled == nil &&
		u.MarketingEmailsEnabled == nil &&
		u.DisplayName == nil
}

// SettingsRepository is the persistence boundary for the Settings
// read/write pair. It is intentionally separate from the onboarding
// Repository so the D04 seed function's read-only view of
// user_settings and the Settings write path can evolve independently.
// The transactions are owned by the repository implementation; the
// service layer never sees a *sql.Tx.
type SettingsRepository interface {
	// GetSettings returns the learner's full Settings projection.
	// When no user_settings row exists yet, every field is filled
	// from the user_settings schema defaults plus an empty
	// display_name, so the API layer never has to fabricate
	// a 404-or-defaults signal — the response shape is stable.
	GetSettings(ctx context.Context, userID uuid.UUID) (Settings, error)
	// UpdateSettings atomically applies a partial update to the
	// requester's user_settings row and users.display_name. Any
	// field left nil in update is preserved. When no
	// user_settings row exists yet, the implementation must
	// upsert one with the supplied fields plus the schema
	// defaults for the rest (handling the concurrent first-read
	// race the gamification module's lazy-create path could
	// otherwise collide with — VOC-031-R05).
	UpdateSettings(ctx context.Context, userID uuid.UUID, update SettingsUpdate, now time.Time) (Settings, error)
}

// Review interval preset enum values (user_settings check
// constraint).
const (
	ReviewIntervalPresetVocabDefault = "vocanova_default"
	ReviewIntervalPresetWordUpLike   = "wordup_like"
	ReviewIntervalPresetCustom       = "custom"
)

// MaxDisplayNameLength caps users.display_name at a length that
// fits the existing A1 identity DTO's 120-char ceiling plus a small
// safety margin against future client-side rendering assumptions.
// It is intentionally narrower than the schema column to keep the
// API contract honest about what a learner-facing display name is.
const MaxDisplayNameLength = 80

// SupportedAppLanguages is the founder-directed set of accepted
// appLanguage values. Per VOC-031-D06, only "en" is accepted at
// launch; the i18n infrastructure this would otherwise unlock does
// not exist in this repository today, so accepting additional
// values would silently lie about a capability the product does
// not have. Expanding the set is future scope once real i18n
// exists.
var SupportedAppLanguages = []string{"en"}

// ErrInvalidSettings distinguishes learner input errors from persistence failures.
var ErrInvalidSettings = errors.New("invalid settings")

// Validate enforces the founder-directed field constraints for a
// PATCH /api/v1/settings payload. It returns a stable, exported
// error for every rejection so the API layer can map them to
// 400 without a string-match. The DB-level check constraints
// remain a second line of defense.
func (u SettingsUpdate) Validate() error {
	if u.DailyReviewTarget != nil {
		if *u.DailyReviewTarget < MinDailyReviewTarget || *u.DailyReviewTarget > MaxDailyReviewTarget {
			return fmt.Errorf("%w: daily review target %d out of range [%d,%d]", ErrInvalidSettings, *u.DailyReviewTarget, MinDailyReviewTarget, MaxDailyReviewTarget)
		}
	}
	if u.ReviewIntervalPreset != nil {
		switch *u.ReviewIntervalPreset {
		case ReviewIntervalPresetVocabDefault, ReviewIntervalPresetWordUpLike, ReviewIntervalPresetCustom:
		default:
			return fmt.Errorf("%w: invalid review interval preset %q", ErrInvalidSettings, *u.ReviewIntervalPreset)
		}
	}
	if u.AppLanguage != nil {
		if !isSupportedAppLanguage(*u.AppLanguage) {
			return fmt.Errorf("%w: app language %q is not supported", ErrInvalidSettings, *u.AppLanguage)
		}
	}
	if u.DisplayName != nil {
		// Display name is a learner-controlled string. Trim
		// nothing here; the API layer is expected to pass
		// the raw value and the length cap rejects values
		// that cannot render usefully.
		// Match the API schema's maxLength: Unicode code points, not UTF-8 bytes.
		if utf8.RuneCountInString(*u.DisplayName) > MaxDisplayNameLength {
			return fmt.Errorf("%w: display name longer than %d characters", ErrInvalidSettings, MaxDisplayNameLength)
		}
	}
	return nil
}

func isSupportedAppLanguage(value string) bool {
	for _, supported := range SupportedAppLanguages {
		if value == supported {
			return true
		}
	}
	return false
}

// ErrSettingsNotFound is returned by GetSettings when the
// learner identified by userID does not exist (or has been soft-
// deleted). It mirrors users.ErrUserNotFound so the API layer
// can map it to 404 with the same privacy-preserving posture.
var ErrSettingsNotFound = errors.New("settings not found")
