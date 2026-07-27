package users

// SchemaDailyReviewTargetDefault mirrors the user_settings schema default for
// daily_review_target (DOC-05 §6: `daily_review_target integer NOT NULL DEFAULT
// 20`). The D04 seed rule treats exactly this value as "the row has not been
// customized yet" and overwrites it; any other value is preserved as the
// learner's already-explicit choice.
const SchemaDailyReviewTargetDefault = 20

// StoredUserSettings is the minimal view the D04 seed-eligibility decision
// needs about the learner's user_settings row. The user module's caller is
// responsible for fetching this from user_settings; `Stored` reports whether
// a row exists, and `DailyReviewTarget` is the current value (only meaningful
// when Stored is true).
type StoredUserSettings struct {
	Stored            bool
	DailyReviewTarget int
}

// SeedAction describes the operation the caller should take on the
// learner's user_settings row in response to a successful onboarding
// submission, per VOC-031-D04.
type SeedAction int

const (
	// SeedCreateRow: no existing user_settings row. The caller must
	// create a user_settings row whose daily_review_target equals the
	// onboarding answer. This is the only branch where a row insert is
	// required.
	SeedCreateRow SeedAction = iota
	// SeedOverwriteDefault: an existing user_settings row whose
	// daily_review_target still equals SchemaDailyReviewTargetDefault
	// (20) — never customized away from the schema default. The
	// caller must overwrite the existing row's daily_review_target
	// with the onboarding answer.
	SeedOverwriteDefault
	// SeedPreserveExisting: an existing user_settings row whose
	// daily_review_target has been customized away from the schema
	// default. The caller must leave the existing row untouched
	// (Decision.Value holds the value to preserve, identical to the
	// input DailyReviewTarget).
	SeedPreserveExisting
)

// SeedDecision is the result of ResolveDailyReviewTargetSeed. The
// transaction layer is responsible for materializing it into a real
// user_settings write inside the caller's existing *sql.Tx.
type SeedDecision struct {
	// Action is the operation the caller should take.
	Action SeedAction
	// Value is the daily_review_target the caller should write. It
	// equals the onboarding answer for SeedCreateRow and
	// SeedOverwriteDefault; it equals the existing stored value
	// (unchanged) for SeedPreserveExisting.
	Value int
}

// String makes SeedAction printable for logs and test failure output.
func (a SeedAction) String() string {
	switch a {
	case SeedCreateRow:
		return "create_row"
	case SeedOverwriteDefault:
		return "overwrite_default"
	case SeedPreserveExisting:
		return "preserve_existing"
	default:
		return "unknown"
	}
}

// ResolveDailyReviewTargetSeed implements the VOC-031-D04
// seed-eligibility rule exactly. It is a pure function (no IO) and
// matches the founder-supplied wording of D04 verbatim:
//
//   - No existing user_settings row (Stored == false) → create one
//     with daily_review_target = onboarding.
//   - Existing row with daily_review_target == 20 (the schema default)
//     → overwrite that column with onboarding.
//   - Existing row with any other daily_review_target → never
//     overwrite; the existing value is preserved exactly.
//
// Validation of the onboarding value (the 5..100 bound) is the
// caller's job — this function does not enforce it, because the
// user_onboarding_profiles.daily_review_target check constraint and
// the user_settings.daily_review_target check constraint both do at
// the DB layer. The T01 caller is expected to have rejected
// out-of-range submissions before reaching this function.
func ResolveDailyReviewTargetSeed(stored StoredUserSettings, onboarding int) SeedDecision {
	if !stored.Stored {
		return SeedDecision{Action: SeedCreateRow, Value: onboarding}
	}
	if stored.DailyReviewTarget == SchemaDailyReviewTargetDefault {
		return SeedDecision{Action: SeedOverwriteDefault, Value: onboarding}
	}
	return SeedDecision{Action: SeedPreserveExisting, Value: stored.DailyReviewTarget}
}
