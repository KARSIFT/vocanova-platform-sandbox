package aifeedback

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Target is the authoritative word/phrase data loaded for a feedback request.
// It is built only after the service has confirmed the attempt is owned by the
// authenticated learner.
type Target struct {
	WordID          uuid.UUID
	MeaningID       uuid.UUID
	UserWordID      uuid.UUID
	ReviewAttemptID *uuid.UUID
	WordText        string
	NormalizedWord  string
	WordType        string
	PartOfSpeech    string
	ShortDefinition string
	LearnerLevel    string
	AcceptedForms   []string
}

// LoadTargetRequest identifies the learner-owned attempt a sentence belongs to.
type LoadTargetRequest struct {
	UserID    uuid.UUID
	Source    string
	AttemptID uuid.UUID
}

// BuildAcceptedForms returns deterministic accepted forms for a target word or
// phrase. It does not silently accept synonyms; only the canonical target and
// the configured inflection/variant forms are accepted.
func BuildAcceptedForms(word, wordType, partOfSpeech string) []string {
	base := strings.ToLower(strings.TrimSpace(word))
	forms := map[string]struct{}{base: {}}

	if isPhraseType(wordType) {
		return sortedForms(forms)
	}

	switch partOfSpeech {
	case "verb":
		addVerbForms(forms, base)
	case "noun":
		addNounForms(forms, base)
	case "adjective", "adverb":
		addAdjectiveForms(forms, base)
	default:
		addDefaultForms(forms, base)
	}
	return sortedForms(forms)
}

func isPhraseType(wordType string) bool {
	switch wordType {
	case "phrase", "phrasal_verb", "idiom", "collocation":
		return true
	}
	return false
}

func addVerbForms(forms map[string]struct{}, base string) {
	addForm(forms, base+"s")
	if shouldAddEs(base) {
		addForm(forms, base+"es")
	}
	if strings.HasSuffix(base, "y") && !hasVowelBeforeSuffix(base, "y") {
		n := len(base) - 1
		addForm(forms, base[:n]+"ies")
		addForm(forms, base[:n]+"ied")
	}

	if strings.HasSuffix(base, "e") {
		stem := base[:len(base)-1]
		addForm(forms, stem+"ed")
		addForm(forms, stem+"ing")
		addForm(forms, base+"d")
	} else {
		addForm(forms, base+"ed")
		addForm(forms, base+"ing")
	}
}

func shouldAddEs(base string) bool {
	return strings.HasSuffix(base, "s") || strings.HasSuffix(base, "x") ||
		strings.HasSuffix(base, "ch") || strings.HasSuffix(base, "sh") ||
		strings.HasSuffix(base, "o")
}

func addNounForms(forms map[string]struct{}, base string) {
	addForm(forms, base+"s")

	switch {
	case strings.HasSuffix(base, "y") && !hasVowelBeforeSuffix(base, "y"):
		addForm(forms, base[:len(base)-1]+"ies")
	case strings.HasSuffix(base, "s"), strings.HasSuffix(base, "x"),
		strings.HasSuffix(base, "ch"), strings.HasSuffix(base, "sh"),
		strings.HasSuffix(base, "o"):
		addForm(forms, base+"es")
	}
}

func addAdjectiveForms(forms map[string]struct{}, base string) {
	addForm(forms, base+"er")
	addForm(forms, base+"est")
}

func addDefaultForms(forms map[string]struct{}, base string) {
	addForm(forms, base+"s")
	addForm(forms, base+"ed")
	addForm(forms, base+"ing")
}

func addForm(forms map[string]struct{}, form string) {
	if form == "" {
		return
	}
	forms[form] = struct{}{}
}

func hasVowelBeforeSuffix(base, suffix string) bool {
	idx := strings.LastIndex(base, suffix)
	if idx <= 0 {
		return false
	}
	return isVowel(rune(base[idx-1]))
}

func isVowel(r rune) bool {
	switch r {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}

func sortedForms(forms map[string]struct{}) []string {
	out := make([]string, 0, len(forms))
	for k := range forms {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// SentenceContainsTarget checks whether the normalized sentence contains the
// target word/phrase or one of its accepted forms. Phrase targets are matched
// as a sequence of tokens; single-word targets are matched against any token.
func SentenceContainsTarget(sentence string, target *Target) bool {
	sentence = strings.ToLower(strings.TrimSpace(sentence))
	tokens := sentenceTokens(sentence)

	phraseTokens := sentenceTokens(strings.ToLower(strings.TrimSpace(target.NormalizedWord)))
	if isPhraseType(target.WordType) || len(phraseTokens) > 1 {
		return containsTokenSequence(tokens, phraseTokens)
	}

	forms := make(map[string]struct{}, len(target.AcceptedForms))
	for _, f := range target.AcceptedForms {
		forms[f] = struct{}{}
	}

	for _, tok := range tokens {
		tok = stripPunctuation(tok)
		tok = stripPossessive(tok)
		if _, ok := forms[tok]; ok {
			return true
		}
	}
	return false
}

func containsTokenSequence(tokens, phrase []string) bool {
	if len(phrase) == 0 {
		return false
	}
	for i := 0; i <= len(tokens)-len(phrase); i++ {
		match := true
		for j, pt := range phrase {
			if stripPunctuation(tokens[i+j]) != stripPunctuation(pt) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func sentenceTokens(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		if unicode.IsSpace(r) {
			return true
		}
		switch r {
		case '-', '–', '—', '/':
			return true
		}
		return false
	})
}

func stripPunctuation(s string) string {
	start := 0
	for start < len(s) {
		r, size := utf8.DecodeRuneInString(s[start:])
		if isPunctuationOrSymbol(r) {
			start += size
			continue
		}
		break
	}
	end := len(s)
	for end > start {
		r, size := utf8.DecodeLastRuneInString(s[:end])
		if isPunctuationOrSymbol(r) {
			end -= size
			continue
		}
		break
	}
	return s[start:end]
}

func stripPossessive(s string) string {
	if strings.HasSuffix(s, "'s") && len(s) > 2 {
		return s[:len(s)-2]
	}
	if strings.HasSuffix(s, "'") && len(s) > 1 {
		return s[:len(s)-1]
	}
	return s
}

func isPunctuationOrSymbol(r rune) bool {
	if r == '\'' {
		return false
	}
	return unicode.IsPunct(r) || unicode.IsSymbol(r)
}
