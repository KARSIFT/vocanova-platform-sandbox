package schema

import "testing"

func TestReviewAttemptEntFieldsAreImmutable(t *testing.T) {
	fields := append(ReviewAttempt{}.Fields(), ImmutableTimeMixin{}.Fields()...)
	for _, field := range fields {
		descriptor := field.Descriptor()
		if !descriptor.Immutable {
			t.Errorf("review attempt field %q must be create-only in Ent", descriptor.Name)
		}
		if descriptor.UpdateDefault != nil {
			t.Errorf("review attempt field %q must not have an update default", descriptor.Name)
		}
	}
}
