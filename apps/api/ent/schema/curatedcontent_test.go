package schema

import "testing"

func TestValidateCuratedContentNonblank(t *testing.T) {
	for _, value := range []string{"", " \t\r\n", "\u00a0\u2007\u202f\u3000"} {
		if err := validateCuratedContentNonblank(value); err == nil {
			t.Fatalf("validateCuratedContentNonblank(%q) succeeded", value)
		}
	}
	for _, value := range []string{"word", "\u00a0word\u3000"} {
		if err := validateCuratedContentNonblank(value); err != nil {
			t.Fatalf("validateCuratedContentNonblank(%q): %v", value, err)
		}
	}
}
