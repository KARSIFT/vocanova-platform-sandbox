package schema

import (
	"errors"
	"strings"
)

const curatedContentNonblankPattern = "U&'[^\\0009-\\000D\\0020\\0085\\00A0\\1680\\2000-\\200A\\2028\\2029\\202F\\205F\\3000]'"

func curatedContentNonblankCheck(column string) string {
	return column + " ~ " + curatedContentNonblankPattern
}

// validateCuratedContentNonblank mirrors the migration's Unicode White_Space
// check. Keep Ent's early validation aligned with PostgreSQL's final guard.
func validateCuratedContentNonblank(value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("must contain a non-whitespace character")
	}
	return nil
}
