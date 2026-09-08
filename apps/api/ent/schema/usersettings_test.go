package schema

import (
	"os"
	"strings"
	"testing"

	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
)

// TestUserSettingsAppLanguageMatchesVOC1445Migration keeps Ent's declarative
// schema aligned with the forward migration that is authoritative at runtime.
func TestUserSettingsAppLanguageMatchesVOC1445Migration(t *testing.T) {
	contents, err := os.ReadFile("../../migrations/20260908300000_voc1445_en_only_app_language.sql")
	if err != nil {
		t.Fatal(err)
	}
	migration := string(contents)
	if !strings.Contains(migration, "CHECK (app_language = 'en') NOT VALID") {
		t.Fatal("VOC-1445 migration no longer declares en-only app language")
	}
	settings := UserSettings{}
	annotations := settings.Annotations()
	if len(annotations) != 1 {
		t.Fatalf("UserSettings annotations = %d, want one SQL annotation", len(annotations))
	}
	annotation, ok := annotations[0].(entsql.Annotation)
	if !ok {
		t.Fatalf("UserSettings annotation type = %T, want entsql.Annotation", annotations[0])
	}
	if got := annotation.Checks["app_language_valid"]; got != "app_language = 'en'" {
		t.Fatalf("Ent app-language check = %q, want en-only predicate", got)
	}
	for _, schemaField := range settings.Fields() {
		descriptor := schemaField.Descriptor()
		if descriptor.Name != "app_language" {
			continue
		}
		if descriptor.Info.Type != field.TypeEnum || descriptor.Default != "en" || len(descriptor.Enums) != 1 || descriptor.Enums[0].V != "en" {
			t.Fatalf("Ent app-language field = %#v, want en-only enum with en default", descriptor)
		}
		return
	}
	t.Fatal("UserSettings is missing app_language")
}
