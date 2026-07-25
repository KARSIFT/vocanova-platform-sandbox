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

func TestSentenceContainsTargetPossessive(t *testing.T) {
	target := &Target{
		NormalizedWord: "work",
		WordType:       "word",
		PartOfSpeech:   "noun",
		AcceptedForms:  BuildAcceptedForms("work", "word", "noun"),
	}
	assert.True(t, SentenceContainsTarget("My work's quality is high.", target))
}
