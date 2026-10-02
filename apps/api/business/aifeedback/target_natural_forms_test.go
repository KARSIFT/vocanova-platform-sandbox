package aifeedback

import (
	"slices"
	"testing"
)

func TestCuratedNaturalVerbForms(t *testing.T) {
	for _, tc := range []struct {
		word, wordType string
		forms          []string
	}{
		{"travel", "word", []string{"travelled", "travelling", "traveled", "traveling"}},
		{"learn", "word", []string{"learnt", "learned"}},
		{"organize", "word", []string{"organise", "organises", "organised", "organising", "organizes", "organized", "organizing"}},
		{"cancel", "word", []string{"cancelled", "cancelling", "canceled", "canceling"}},
		{"catch up", "phrase", []string{"catch up", "catches up", "catching up", "caught up"}},
		{"meet up", "phrasal_verb", []string{"meet up", "meets up", "meeting up", "met up"}},
		{"keep in touch", "idiom", []string{"keep in touch", "keeps in touch", "keeping in touch", "kept in touch"}},
	} {
		t.Run(tc.word, func(t *testing.T) {
			target := &Target{NormalizedWord: tc.word, WordType: tc.wordType, PartOfSpeech: "verb", AcceptedForms: BuildAcceptedForms(tc.word, tc.wordType, "verb")}
			for _, form := range tc.forms {
				if !slices.Contains(target.AcceptedForms, form) {
					t.Errorf("missing curated form %q in %v", form, target.AcceptedForms)
				}
				// This checks lexical presence only; it does not certify grammar or meaning.
				if got := ValidateSentence("They "+form+" every week.", target); !got.Valid {
					t.Errorf("%q rejected: %s", form, got.Code)
				}
			}
		})
	}
}

func TestCuratedNaturalFormsStayScoped(t *testing.T) {
	for _, tc := range []struct{ word, wordType, pos, excluded string }{
		{"travel", "word", "noun", "travelled"},
		{"learn", "word", "noun", "learnt"},
		{"organize", "word", "noun", "organised"},
		{"cancel", "word", "adjective", "cancelled"},
		{"travel", "phrase", "verb", "travelled"},
		{"catch up", "phrase", "noun", "caught up"},
		{"catch up", "idiom", "verb", "caught up"},
		{"meet up", "phrasal_verb", "noun", "met up"},
		{"meet up", "phrase", "verb", "met up"},
		{"keep in touch", "idiom", "noun", "kept in touch"},
		{"keep in touch", "phrasal_verb", "verb", "kept in touch"},
		{"keep quiet", "idiom", "verb", "kept quiet"},
		{"sounds good", "phrase", "verb", "sounded good"},
		{"sounds good", "idiom", "phrase", "sound good"},
	} {
		if got := BuildAcceptedForms(tc.word, tc.wordType, tc.pos); slices.Contains(got, tc.excluded) {
			t.Errorf("%s/%s/%s unexpectedly expands to %q", tc.word, tc.wordType, tc.pos, tc.excluded)
		}
	}
}

func TestNaturalFormsPreserveContiguousWholeTokenMatching(t *testing.T) {
	for _, tc := range []struct{ word, wordType, sentence string }{
		{"travel", "word", "We discussed untravelled routes today."},
		{"learn", "word", "The unlearnt lesson was important."},
		{"organize", "word", "We reorganised all the notes."},
		{"cancel", "word", "The uncancelled booking is active."},
		{"catch up", "phrase", "We caught right up yesterday."},
		{"meet up", "phrasal_verb", "We met her up there."},
		{"keep in touch", "idiom", "We kept regularly in touch."},
		{"keep in touch", "idiom", "Please stay in touch with me."},
	} {
		target := &Target{NormalizedWord: tc.word, WordType: tc.wordType, PartOfSpeech: "verb", AcceptedForms: BuildAcceptedForms(tc.word, tc.wordType, "verb")}
		if got := ValidateSentence(tc.sentence, target); got.Valid || got.Code != ValidationCodeMissingTarget {
			t.Errorf("%q should miss target %q: %+v", tc.sentence, tc.word, got)
		}
	}
	target := &Target{NormalizedWord: "sounds good", WordType: "phrase", AcceptedForms: BuildAcceptedForms("sounds good", "phrase", "phrase")}
	if got := ValidateSentence("Sounds good!", target); got.Valid || got.Code != ValidationCodeTooShort {
		t.Errorf("short response must retain minimum-word validation: %+v", got)
	}
	if got := ValidateSentence("That sounds good.", target); !got.Valid {
		t.Errorf("three-word response rejected: %+v", got)
	}
}

func TestCuratedSoundsGoodForms(t *testing.T) {
	target := &Target{NormalizedWord: "sounds good", WordType: "phrase", PartOfSpeech: "phrase", AcceptedForms: BuildAcceptedForms("sounds good", "phrase", "phrase")}
	for _, sentence := range []string{"That sounds good.", "Those plans sound good.", "That sounded good."} {
		if got := ValidateSentence(sentence, target); !got.Valid {
			t.Errorf("%q rejected: %s", sentence, got.Code)
		}
	}
	for _, sentence := range []string{"That sounded very good.", "Those plans sound goodish.", "That seems good to me."} {
		if got := ValidateSentence(sentence, target); got.Valid || got.Code != ValidationCodeMissingTarget {
			t.Errorf("%q should miss target: %+v", sentence, got)
		}
	}
}
