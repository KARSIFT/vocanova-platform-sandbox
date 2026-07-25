package aifeedback

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// ValidationResult is the outcome of deterministic sentence validation.
type ValidationResult struct {
	Valid      bool
	Code       string
	Normalized string
	WordCount  int
}

// Default validation limits (DOC-09 §6).
const (
	DefaultMaxSentenceLength = 300
	DefaultMinSentenceWords  = 3
)

// ValidateSentence checks the learner sentence against the deterministic input
// rules. It never calls an external model. Validation codes are stable and are
// surfaced to the caller so the frontend can show a safe retryable message.
func ValidateSentence(input string, target *Target) ValidationResult {
	normalized := normalizeSentence(input)

	if normalized == "" {
		return ValidationResult{Code: ValidationCodeInvalidInput, Normalized: normalized}
	}

	if hasInvalidCharacters(normalized) {
		return ValidationResult{Code: ValidationCodeInvalidInput, Normalized: normalized}
	}

	if !isPrimarilyEnglish(normalized) {
		return ValidationResult{Code: ValidationCodeUnsupportedLanguage, Normalized: normalized}
	}

	tokens := meaningfulTokens(normalized)
	if len(tokens) < DefaultMinSentenceWords {
		return ValidationResult{Code: ValidationCodeTooShort, Normalized: normalized, WordCount: len(tokens)}
	}

	if len([]rune(normalized)) > DefaultMaxSentenceLength {
		return ValidationResult{Code: ValidationCodeTooLong, Normalized: normalized, WordCount: len(tokens)}
	}

	if target == nil || !SentenceContainsTarget(normalized, target) {
		return ValidationResult{Code: ValidationCodeMissingTarget, Normalized: normalized, WordCount: len(tokens)}
	}

	return ValidationResult{Valid: true, Normalized: normalized, WordCount: len(tokens)}
}

func normalizeSentence(s string) string {
	s = strings.TrimSpace(s)
	s = norm.NFKC.String(s)
	fields := strings.Fields(s)
	s = strings.Join(fields, " ")
	return strings.ToLower(s)
}

func meaningfulTokens(s string) []string {
	var out []string
	for _, tok := range strings.Fields(s) {
		tok = stripPunctuation(tok)
		tok = stripPossessive(tok)
		if hasLetters(tok) {
			out = append(out, tok)
		}
	}
	return out
}

func hasLetters(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func isPrimarilyEnglish(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.Is(unicode.Latin, r) {
			return false
		}
	}
	return true
}

func hasInvalidCharacters(s string) bool {
	for _, r := range s {
		if r == '\x00' {
			return true
		}
		if unicode.IsControl(r) && !unicode.IsSpace(r) {
			return true
		}
	}
	return false
}
