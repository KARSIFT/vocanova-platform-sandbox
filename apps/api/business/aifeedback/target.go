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
	addCuratedNaturalForms(forms, base, wordType, partOfSpeech)

	if isPhraseType(wordType) {
		if isRegularNounPhraseType(wordType) && partOfSpeech == "noun" {
			addNounPhraseForms(forms, base)
		}
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

// addCuratedNaturalForms approves only these canonical target/type/POS tuples.
// Phrase heads are not inflected generically: idioms and other phrase types may
// have different meanings or morphology. Presence still says nothing about
// grammatical correctness or whether the learner used the selected meaning.
func addCuratedNaturalForms(forms map[string]struct{}, base, wordType, partOfSpeech string) {
	var approved []string
	switch {
	case wordType == "word" && partOfSpeech == "verb":
		switch base {
		case "travel":
			approved = []string{"travelled", "travelling"}
		case "learn":
			approved = []string{"learnt"}
		case "organize":
			approved = []string{"organise", "organises", "organised", "organising"}
		case "cancel":
			approved = []string{"cancelled", "cancelling"}
		}
	case base == "catch up" && wordType == "phrase" && partOfSpeech == "verb":
		approved = []string{"catches up", "catching up", "caught up"}
	case base == "meet up" && wordType == "phrasal_verb" && partOfSpeech == "verb":
		approved = []string{"meets up", "meeting up", "met up"}
	case base == "try on" && wordType == "phrasal_verb" && partOfSpeech == "verb":
		approved = []string{"tries on", "trying on", "tried on"}
		// These authored clothing forms permit only the two object pronouns,
		// not arbitrary gaps or a generic separable-phrasal-verb algorithm.
		for _, head := range []string{"try", "tries", "trying", "tried"} {
			for _, object := range []string{"it", "them"} {
				approved = append(approved, head+" "+object+" on")
			}
		}
	case base == "call back" && wordType == "phrasal_verb" && partOfSpeech == "verb":
		approved = []string{"calls back", "calling back", "called back"}
		for _, head := range []string{"call", "calls", "calling", "called"} {
			for _, object := range []string{"me", "you", "him", "her", "us", "them"} {
				approved = append(approved, head+" "+object+" back")
			}
		}
	case base == "keep in touch" && wordType == "idiom" && partOfSpeech == "verb":
		approved = []string{"keeps in touch", "keeping in touch", "kept in touch"}
	case base == "sounds good" && wordType == "phrase" && partOfSpeech == "phrase":
		approved = []string{"sound good", "sounded good"}
	}
	for _, form := range approved {
		addForm(forms, form)
	}
}

func isPhraseType(wordType string) bool {
	switch wordType {
	case "phrase", "phrasal_verb", "idiom", "collocation":
		return true
	}
	return false
}

func isRegularNounPhraseType(wordType string) bool {
	return wordType == "phrase" || wordType == "collocation"
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
	// The canonical syllabus usage note explicitly approves both syllabuses
	// and syllabi (cmd/seed/voc026-p1.json). Keep this exception noun-only.
	if base == "syllabus" {
		addForm(forms, "syllabi")
	}
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

// addNounPhraseForms adds regular plural forms by inflecting only the final
// token. Irregular or non-final plurals must be configured explicitly.
func addNounPhraseForms(forms map[string]struct{}, base string) {
	parts := strings.Fields(base)
	if len(parts) < 2 {
		// Hyphenated noun compounds such as follow-up and check-out are one
		// whitespace token. They still take the regular suffix on the last part;
		// phrase matching already recognizes their hyphen-separated tokens.
		if len(parts) == 1 && strings.Contains(base, "-") {
			addRegularPluralNounForm(forms, base)
		}
		return
	}

	finalForms := map[string]struct{}{}
	addRegularPluralNounForm(finalForms, parts[len(parts)-1])
	prefix := strings.Join(parts[:len(parts)-1], " ")
	for final := range finalForms {
		addForm(forms, prefix+" "+final)
	}
}

func addRegularPluralNounForm(forms map[string]struct{}, base string) {
	switch {
	case strings.HasSuffix(base, "y") && !hasVowelBeforeSuffix(base, "y"):
		addForm(forms, base[:len(base)-1]+"ies")
	case strings.HasSuffix(base, "s"), strings.HasSuffix(base, "x"),
		strings.HasSuffix(base, "ch"), strings.HasSuffix(base, "sh"),
		strings.HasSuffix(base, "o"):
		addForm(forms, base+"es")
	default:
		addForm(forms, base+"s")
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
// target word/phrase or one of its accepted forms. Phrase targets and their
// configured variants are matched as sequences of tokens; single-word targets
// are matched against any token.
func SentenceContainsTarget(sentence string, target *Target) bool {
	sentence = strings.ToLower(strings.TrimSpace(sentence))
	tokens := sentenceTokens(sentence)

	canonicalPhraseTokens := sentenceTokens(strings.ToLower(strings.TrimSpace(target.NormalizedWord)))
	if isPhraseType(target.WordType) || len(canonicalPhraseTokens) > 1 {
		forms := append([]string{target.NormalizedWord}, target.AcceptedForms...)
		for _, form := range forms {
			phraseTokens := sentenceTokens(strings.ToLower(strings.TrimSpace(form)))
			if target.WordType == "phrasal_verb" && target.PartOfSpeech == "verb" &&
				(target.NormalizedWord == "try on" || target.NormalizedWord == "call back") && len(phraseTokens) == 3 {
				if containsCuratedObjectForm(sentence, phraseTokens) {
					return true
				}
				continue
			}
			if containsTokenSequence(tokens, phraseTokens) {
				return true
			}
		}
		return false
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

// Only the new curated three-word object-pronoun forms use this path. Keep
// punctuation between the words from joining separate clauses or sentences.
// Outer quotation marks and sentence punctuation do not hide a valid form.
func containsCuratedObjectForm(sentence string, form []string) bool {
	words := strings.Fields(sentence)
	for i := 0; i+2 < len(words); i++ {
		if strings.TrimLeftFunc(words[i], isPunctuationOrSymbol) == form[0] &&
			words[i+1] == form[1] && strings.TrimRightFunc(words[i+2], isPunctuationOrSymbol) == form[2] {
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
