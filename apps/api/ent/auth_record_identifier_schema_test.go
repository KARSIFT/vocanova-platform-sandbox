package ent_test

import (
	"testing"

	"entgo.io/ent"
	"github.com/KARSIFT/vocanova-platform/apps/api/ent/schema"
)

func TestAuthRecordIdentifierFieldsRejectExactlyUnicodeWhiteSpace(t *testing.T) {
	fieldsBySchema := map[string][]ent.Field{
		"EmailChangeLink":  (schema.EmailChangeLink{}).Fields(),
		"ExternalIdentity": (schema.ExternalIdentity{}).Fields(),
		"MagicLink":        (schema.MagicLink{}).Fields(),
	}
	targets := map[string]map[string]bool{
		"EmailChangeLink":  {"new_email": true, "environment": true},
		"ExternalIdentity": {"provider_subject": true},
		"MagicLink":        {"email": true, "environment": true},
	}
	allUnicodeWhiteSpace := "\u0009\u000A\u000B\u000C\u000D\u0020\u0085\u00A0\u1680\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200A\u2028\u2029\u202F\u205F\u3000"

	for schemaName, fields := range fieldsBySchema {
		for _, field := range fields {
			descriptor := field.Descriptor()
			if !targets[schemaName][descriptor.Name] {
				continue
			}
			validators := stringValidators(t, schemaName+"."+descriptor.Name, descriptor.Validators)
			for _, blank := range []string{"", " \t\r\n", allUnicodeWhiteSpace} {
				if !anyRejects(validators, blank) {
					t.Fatalf("%s accepted whitespace-only value %q", schemaName+"."+descriptor.Name, blank)
				}
			}
			for _, valid := range []string{"identifier", "\u00A0identifier\u3000", "\u200B"} {
				for _, validate := range validators {
					if err := validate(valid); err != nil {
						t.Fatalf("%s rejected valid value %q: %v", schemaName+"."+descriptor.Name, valid, err)
					}
				}
			}
			delete(targets[schemaName], descriptor.Name)
		}
		if len(targets[schemaName]) != 0 {
			t.Fatalf("%s missing validated fields %v", schemaName, targets[schemaName])
		}
	}
}

func stringValidators(t *testing.T, fieldName string, raw []any) []func(string) error {
	t.Helper()
	if len(raw) < 2 {
		t.Fatalf("%s validators = %d, want NotEmpty plus Unicode whitespace validation", fieldName, len(raw))
	}
	validators := make([]func(string) error, 0, len(raw))
	for _, validator := range raw {
		validate, ok := validator.(func(string) error)
		if !ok {
			t.Fatalf("%s validator type = %T, want func(string) error", fieldName, validator)
		}
		validators = append(validators, validate)
	}
	return validators
}

func anyRejects(validators []func(string) error, value string) bool {
	for _, validate := range validators {
		if validate(value) != nil {
			return true
		}
	}
	return false
}
