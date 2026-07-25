package aifeedback

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSentenceAcceptsValidSentence(t *testing.T) {
	target := &Target{
		NormalizedWord: "work",
		WordType:       "word",
		PartOfSpeech:   "verb",
		AcceptedForms:  BuildAcceptedForms("work", "word", "verb"),
	}
	result := ValidateSentence("I work every day.", target)
	assert.True(t, result.Valid)
	assert.Equal(t, "i work every day.", result.Normalized)
	assert.Equal(t, 4, result.WordCount)
}

func TestValidateSentenceTooShort(t *testing.T) {
	target := &Target{
		NormalizedWord: "work",
		WordType:       "word",
		PartOfSpeech:   "verb",
		AcceptedForms:  BuildAcceptedForms("work", "word", "verb"),
	}
	result := ValidateSentence("I work", target)
	assert.False(t, result.Valid)
	assert.Equal(t, ValidationCodeTooShort, result.Code)
}

func TestValidateSentenceTooLong(t *testing.T) {
	target := &Target{
		NormalizedWord: "work",
		WordType:       "word",
		PartOfSpeech:   "verb",
		AcceptedForms:  BuildAcceptedForms("work", "word", "verb"),
	}
	long := "I work " + strings.Repeat("very ", 100) + "hard."
	result := ValidateSentence(long, target)
	assert.False(t, result.Valid)
	assert.Equal(t, ValidationCodeTooLong, result.Code)
}

func TestValidateSentenceMissingTarget(t *testing.T) {
	target := &Target{
		NormalizedWord: "work",
		WordType:       "word",
		PartOfSpeech:   "verb",
		AcceptedForms:  BuildAcceptedForms("work", "word", "verb"),
	}
	result := ValidateSentence("I play every day.", target)
	assert.False(t, result.Valid)
	assert.Equal(t, ValidationCodeMissingTarget, result.Code)
}

func TestValidateSentenceUnsupportedLanguage(t *testing.T) {
	target := &Target{
		NormalizedWord: "work",
		WordType:       "word",
		PartOfSpeech:   "verb",
		AcceptedForms:  BuildAcceptedForms("work", "word", "verb"),
	}
	result := ValidateSentence("我每天都工作。", target)
	assert.False(t, result.Valid)
	assert.Equal(t, ValidationCodeUnsupportedLanguage, result.Code)
}

func TestValidateSentenceInvalidInput(t *testing.T) {
	target := &Target{
		NormalizedWord: "work",
		WordType:       "word",
		PartOfSpeech:   "verb",
		AcceptedForms:  BuildAcceptedForms("work", "word", "verb"),
	}
	result := ValidateSentence("   \t\n   ", target)
	assert.False(t, result.Valid)
	assert.Equal(t, ValidationCodeInvalidInput, result.Code)
}

func TestValidateSentenceCollapsesWhitespace(t *testing.T) {
	target := &Target{
		NormalizedWord: "work",
		WordType:       "word",
		PartOfSpeech:   "verb",
		AcceptedForms:  BuildAcceptedForms("work", "word", "verb"),
	}
	result := ValidateSentence("  I    work   every   day  ", target)
	assert.True(t, result.Valid)
	assert.Equal(t, "i work every day", result.Normalized)
}
