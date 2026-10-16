package knowledgecatalog

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
)

//go:embed semantic-aliases.json.gz
var aliasCompressed []byte
var aliasesOnce sync.Once
var aliasEntries map[string][]string
var aliasAncestors map[string][]string
var aliasLoadError error

func normalizeAlias(word string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(word, "_", " "), "-", " "))), " ")
}
func loadAliases() {
	aliasesOnce.Do(func() {
		reader, err := gzip.NewReader(bytes.NewReader(aliasCompressed))
		if err != nil {
			aliasLoadError = err
			return
		}
		defer reader.Close()
		raw, err := io.ReadAll(io.LimitReader(reader, 6*1024*1024+1))
		if err != nil {
			aliasLoadError = err
			return
		}
		if len(raw) > 6*1024*1024 {
			aliasLoadError = errors.New("semantic alias index exceeds bound")
			return
		}
		var data struct {
			Version   string              `json:"version"`
			License   string              `json:"wordnet_license"`
			Entries   map[string][]string `json:"entries"`
			Ancestors map[string][]string `json:"ancestors"`
		}
		if err = json.Unmarshal(raw, &data); err != nil {
			aliasLoadError = err
			return
		}
		if data.Version != "wordnet-3.0-all-senses-v2" || data.License != wordNetNotice || len(data.Entries) < 10000 || len(data.Ancestors) < 10000 {
			aliasLoadError = errors.New("invalid semantic alias index")
			return
		}
		aliasAncestors = data.Ancestors
		aliasEntries = map[string][]string{}
		for word, senses := range data.Entries {
			key := normalizeAlias(word)
			aliasEntries[key] = append(aliasEntries[key], senses...)
		}
	})
}

// Resolve regular noun plurals even when a form is outside the ranked catalog.
// Only existing source noun senses are added; no arbitrary word is invented.
func aliasSenses(word string) []string {
	out := append([]string(nil), aliasEntries[word]...)
	for _, rule := range [][2]string{{"s", ""}, {"ses", "s"}, {"xes", "x"}, {"zes", "z"}, {"ches", "ch"}, {"shes", "sh"}, {"men", "man"}, {"ies", "y"}} {
		if !strings.HasSuffix(word, rule[0]) {
			continue
		}
		base := strings.TrimSuffix(word, rule[0]) + rule[1]
		for _, sense := range aliasEntries[base] {
			if strings.HasPrefix(sense, "noun:") {
				out = append(out, sense)
			}
		}
	}
	return out
}

// SemanticAliases conservatively rejects potential synonyms from meaning-choice
// distractors. It compares ALL original WordNet senses, independently of the
// catalog's editorial teaching definition or source_sense_key. Direct hypernym
// ancestry within two steps is also rejected, but shared generic ancestors are
// not compared. Over-exclusion
// is intentional. Unknown words are still compared by normalized spelling.
// This heuristic cannot establish that all remaining distractors are different.
func SemanticAliases(left, right string) (bool, error) {
	left, right = normalizeAlias(left), normalizeAlias(right)
	if left == "" || right == "" {
		return false, nil
	}
	if left == right {
		return true, nil
	}
	loadAliases()
	if aliasLoadError != nil {
		return false, aliasLoadError
	}
	senses := map[string]bool{}
	for _, id := range aliasSenses(left) {
		senses[id] = true
	}
	rightSenses := map[string]bool{}
	for _, id := range aliasSenses(right) {
		rightSenses[id] = true
		if senses[id] {
			return true, nil
		}
		for _, ancestor := range aliasAncestors[id] {
			if senses[ancestor] {
				return true, nil
			}
		}
	}
	for _, id := range aliasSenses(left) {
		for _, ancestor := range aliasAncestors[id] {
			if rightSenses[ancestor] {
				return true, nil
			}
		}
	}
	return false, nil
}
