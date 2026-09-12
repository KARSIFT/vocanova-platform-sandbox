package aifeedback

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildAcceptedFormsVerb(t *testing.T) {
	forms := BuildAcceptedForms("work", "word", "verb")
	assert.Contains(t, forms, "work")
	assert.Contains(t, forms, "works")
	assert.Contains(t, forms, "worked")
	assert.Contains(t, forms, "working")
}

func TestBuildAcceptedFormsNoun(t *testing.T) {
	forms := BuildAcceptedForms("box", "word", "noun")
	assert.Contains(t, forms, "box")
	assert.Contains(t, forms, "boxes")
}

func TestBuildAcceptedFormsPhraseOnlyExact(t *testing.T) {
	forms := BuildAcceptedForms("give up", "phrasal_verb", "verb")
	assert.Equal(t, []string{"give up"}, forms)
}

func TestBuildAcceptedFormsNounPhrasePluralizesFinalNoun(t *testing.T) {
	tests := []struct {
		name     string
		word     string
		wordType string
		want     []string
		notWant  string
	}{
		{"regular s ending", "security check", "phrase", []string{"security check", "security checks"}, "security checkes"},
		{"s and ch ending", "boarding pass", "phrase", []string{"boarding pass", "boarding passes"}, "boarding passs"},
		{"consonant y ending", "capital city", "phrase", []string{"capital cities", "capital city"}, "capital citys"},
		{"vowel y ending", "public holiday", "phrase", []string{"public holiday", "public holidays"}, "public holidayses"},
		{"collocation", "lunch box", "collocation", []string{"lunch box", "lunch boxes"}, "lunch boxs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			forms := BuildAcceptedForms(tt.word, tt.wordType, "noun")
			assert.Equal(t, tt.want, forms)
			assert.NotContains(t, forms, tt.notWant)
		})
	}
}

func TestBuildAcceptedFormsNounIdiomRemainsExact(t *testing.T) {
	forms := BuildAcceptedForms("red herring", "idiom", "noun")
	assert.Equal(t, []string{"red herring"}, forms)
}

func TestSentenceContainsTargetInflection(t *testing.T) {
	target := &Target{
		NormalizedWord: "work",
		WordType:       "word",
		PartOfSpeech:   "verb",
		AcceptedForms:  BuildAcceptedForms("work", "word", "verb"),
	}
	assert.True(t, SentenceContainsTarget("I worked yesterday.", target))
	assert.True(t, SentenceContainsTarget("She works hard.", target))
	assert.True(t, SentenceContainsTarget("He is working now.", target))
	assert.False(t, SentenceContainsTarget("I play every day.", target))
}

func TestSentenceContainsTargetCapitalization(t *testing.T) {
	target := &Target{
		NormalizedWord: "work",
		WordType:       "word",
		PartOfSpeech:   "verb",
		AcceptedForms:  BuildAcceptedForms("work", "word", "verb"),
	}
	assert.True(t, SentenceContainsTarget("Work makes me happy.", target))
}

func TestSentenceContainsTargetPhrase(t *testing.T) {
	target := &Target{
		NormalizedWord: "give up",
		WordType:       "phrasal_verb",
		PartOfSpeech:   "verb",
		AcceptedForms:  BuildAcceptedForms("give up", "phrasal_verb", "verb"),
	}
	assert.True(t, SentenceContainsTarget("Never give up!", target))
	assert.False(t, SentenceContainsTarget("I give you a book.", target))
}

func TestSentenceContainsTargetConfiguredPhraseVariant(t *testing.T) {
	target := &Target{
		NormalizedWord: "give up",
		WordType:       "phrasal_verb",
		PartOfSpeech:   "verb",
		AcceptedForms:  []string{"give up", "gave up"},
	}

	assert.True(t, SentenceContainsTarget("I gave up yesterday.", target))
}

func TestSentenceContainsTargetNounPhraseRequiresExactTokenSequence(t *testing.T) {
	target := &Target{
		NormalizedWord: "security check",
		WordType:       "phrase",
		PartOfSpeech:   "noun",
		AcceptedForms:  BuildAcceptedForms("security check", "phrase", "noun"),
	}

	assert.True(t, SentenceContainsTarget("Security checks improve safety.", target))
	assert.False(t, SentenceContainsTarget("A security checkpoint improves safety.", target))
	assert.False(t, SentenceContainsTarget("Security checklists improve safety.", target))
	assert.False(t, SentenceContainsTarget("Security thorough checks improve safety.", target))
}

func TestSentenceContainsTargetAcceptsConfiguredIrregularNounPhraseVariant(t *testing.T) {
	target := &Target{
		NormalizedWord: "attorney general",
		WordType:       "phrase",
		PartOfSpeech:   "noun",
		AcceptedForms:  []string{"attorney general", "attorneys general"},
	}

	assert.True(t, SentenceContainsTarget("The attorneys general met today.", target))
}

func TestSentenceContainsTargetPossessive(t *testing.T) {
	target := &Target{
		NormalizedWord: "work",
		WordType:       "word",
		PartOfSpeech:   "noun",
		AcceptedForms:  BuildAcceptedForms("work", "word", "noun"),
	}
	assert.True(t, SentenceContainsTarget("My work's quality is high.", target))
}
