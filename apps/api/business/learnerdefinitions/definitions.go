// Package learnerdefinitions supplies original plain-English explanations for
// exact canonical meanings. These display additions never change grading,
// catalog selection, canonical revisions, or existing media bindings.
package learnerdefinitions

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
)

//go:embed definitions.json
var originalDefinitions []byte

type entry struct {
	MeaningID         string `json:"meaningId"`
	Revision          string `json:"revision"`
	Word              string `json:"word"`
	PartOfSpeech      string `json:"partOfSpeech"`
	ShortDefinition   string `json:"shortDefinition"`
	LearnerDefinition string `json:"learnerDefinition"`
}

var entries, byRevision = load()

func load() ([]entry, map[string]string) {
	return parseDefinitions(originalDefinitions)
}

func parseDefinitions(data []byte) ([]entry, map[string]string) {
	var rows []entry
	if err := json.Unmarshal(data, &rows); err != nil {
		panic(fmt.Errorf("invalid learner definitions: %w", err))
	}
	lookup := make(map[string]string, len(rows))
	for _, row := range rows {
		revision := meaningidentity.Revision(row.MeaningID, row.Word, row.PartOfSpeech, row.ShortDefinition)
		if revision == "" || revision != row.Revision || strings.TrimSpace(row.LearnerDefinition) == "" || lookup[revision] != "" {
			panic("invalid or duplicate learner definition: " + row.Word)
		}
		lookup[revision] = row.LearnerDefinition
	}
	return rows, lookup
}

// Resolve prefers an existing authored explanation. Otherwise it returns an
// original explanation only when all canonical identity fields match exactly.
// An unknown or changed meaning has no overlay and keeps its source definition.
func Resolve(id, word, pos, definition, authored string) string {
	if strings.TrimSpace(authored) != "" {
		return authored
	}
	return byRevision[meaningidentity.Revision(id, word, pos, definition)]
}
