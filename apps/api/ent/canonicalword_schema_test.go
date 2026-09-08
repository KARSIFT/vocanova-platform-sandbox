package ent_test

import (
	"testing"

	"github.com/KARSIFT/vocanova-platform/apps/api/ent/schema"
)

func TestCanonicalWordFrequencyRankIsOptionalAndPositive(t *testing.T) {
	var descriptorFound bool
	for _, field := range (schema.CanonicalWord{}).Fields() {
		descriptor := field.Descriptor()
		if descriptor.Name != "frequency_rank" {
			continue
		}
		descriptorFound = true
		if !descriptor.Optional || !descriptor.Nillable {
			t.Fatalf("frequency_rank optional/nillable = %t/%t, want true/true", descriptor.Optional, descriptor.Nillable)
		}
		if len(descriptor.Validators) == 0 {
			t.Fatal("frequency_rank has no positive-value validator")
		}
		for _, value := range []int{-1, 0} {
			for _, validator := range descriptor.Validators {
				validate, ok := validator.(func(int) error)
				if !ok {
					t.Fatalf("frequency_rank validator type = %T, want func(int) error", validator)
				}
				if err := validate(value); err == nil {
					t.Fatalf("frequency_rank validator accepted %d", value)
				}
			}
		}
		for _, validator := range descriptor.Validators {
			validate := validator.(func(int) error)
			if err := validate(1); err != nil {
				t.Fatalf("frequency_rank validator rejected boundary 1: %v", err)
			}
		}
	}
	if !descriptorFound {
		t.Fatal("frequency_rank field not found")
	}
}
