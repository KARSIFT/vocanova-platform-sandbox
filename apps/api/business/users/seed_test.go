package users

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestResolveDailyReviewTargetSeedNoExistingRow covers VOC-031-TEST-01:
// the "no existing user_settings row" branch of VOC-031-D04. The caller
// must create a row whose daily_review_target equals the onboarding
// answer.
func TestResolveDailyReviewTargetSeedNoExistingRow(t *testing.T) {
	dec := ResolveDailyReviewTargetSeed(StoredUserSettings{Stored: false}, 30)
	assert.Equal(t, SeedCreateRow, dec.Action)
	assert.Equal(t, 30, dec.Value, "no existing row: caller writes the onboarding answer")
}

// TestResolveDailyReviewTargetSeedExistingDefault covers VOC-031-TEST-02:
// the "existing row at the schema default" branch of VOC-031-D04. The
// caller must overwrite the row's daily_review_target with the
// onboarding answer.
func TestResolveDailyReviewTargetSeedExistingDefault(t *testing.T) {
	dec := ResolveDailyReviewTargetSeed(StoredUserSettings{
		Stored:            true,
		DailyReviewTarget: SchemaDailyReviewTargetDefault,
	}, 35)
	assert.Equal(t, SeedOverwriteDefault, dec.Action)
	assert.Equal(t, 35, dec.Value, "schema-default existing row: caller overwrites with the onboarding answer")
}

// TestResolveDailyReviewTargetSeedCustomizedPreserved covers
// VOC-031-TEST-03: the "existing row with a customized
// daily_review_target" branch of VOC-031-D04. The caller must never
// overwrite the existing value, regardless of what the onboarding
// answer says.
func TestResolveDailyReviewTargetSeedCustomizedPreserved(t *testing.T) {
	dec := ResolveDailyReviewTargetSeed(StoredUserSettings{
		Stored:            true,
		DailyReviewTarget: 35,
	}, 5)
	assert.Equal(t, SeedPreserveExisting, dec.Action)
	assert.Equal(t, 35, dec.Value, "customized existing row: caller never overwrites; existing value preserved exactly")

	dec = ResolveDailyReviewTargetSeed(StoredUserSettings{
		Stored:            true,
		DailyReviewTarget: 100,
	}, 7)
	assert.Equal(t, SeedPreserveExisting, dec.Action)
	assert.Equal(t, 100, dec.Value, "customized existing row at the upper bound: still preserved")

	dec = ResolveDailyReviewTargetSeed(StoredUserSettings{
		Stored:            true,
		DailyReviewTarget: 5,
	}, 99)
	assert.Equal(t, SeedPreserveExisting, dec.Action)
	assert.Equal(t, 5, dec.Value, "customized existing row at the lower bound: still preserved")
}

// TestResolveDailyReviewTargetSeedExistingDefaultExactlyTwentysAreOverwritten
// pins the rule that SchemaDailyReviewTargetDefault (20) is the
// *exact* value the function treats as "untouched schema default".
// Any other stored value — even a value the row could plausibly
// have been seeded to by a prior onboarding submission — falls
// through to SeedPreserveExisting.
func TestResolveDailyReviewTargetSeedExistingDefaultExactlyTwentysAreOverwritten(t *testing.T) {
	for _, v := range []int{5, 19, 21, 35, 100} {
		t.Run("", func(t *testing.T) {
			dec := ResolveDailyReviewTargetSeed(StoredUserSettings{
				Stored:            true,
				DailyReviewTarget: v,
			}, 30)
			if v == SchemaDailyReviewTargetDefault {
				assert.Equal(t, SeedOverwriteDefault, dec.Action, "stored == schema default: overwrite with onboarding answer")
				assert.Equal(t, 30, dec.Value)
			} else {
				assert.Equal(t, SeedPreserveExisting, dec.Action, "stored != schema default: preserve existing")
				assert.Equal(t, v, dec.Value)
			}
		})
	}
}

// TestSeedActionString ensures the printable form is stable enough
// for logs/test output; the test is intentionally minimal because
// no production code switches on the string, but a future caller
// may.
func TestSeedActionString(t *testing.T) {
	assert.Equal(t, "create_row", SeedCreateRow.String())
	assert.Equal(t, "overwrite_default", SeedOverwriteDefault.String())
	assert.Equal(t, "preserve_existing", SeedPreserveExisting.String())
	assert.Equal(t, "unknown", SeedAction(99).String())
}
