package learnerdefinitions

import (
	"compress/gzip"
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestResolveRequiresExactMeaningAndPreservesAuthoredExplanation(t *testing.T) {
	id, word, pos, definition := "6d7b77f2-7914-5bcb-99f3-73ce1ea4619f", "abandon", "verb", "forsake, leave behind"
	require.Equal(t, "To leave someone or something behind.", Resolve(id, word, pos, definition, ""))
	require.Equal(t, "An existing authored explanation.", Resolve(id, word, pos, definition, "An existing authored explanation."))
	for _, input := range [][4]string{{"other", word, pos, definition}, {id, "abandoned", pos, definition}, {id, word, "noun", definition}, {id, word, pos, "Another use."}} {
		require.Empty(t, Resolve(input[0], input[1], input[2], input[3], ""))
	}
	require.Equal(t, "To leave someone or something behind.", Resolve(id, word, pos, definition, " \n "))
}

func TestReviewedDefinitionsBindToFrozenCanonicalSource(t *testing.T) {
	f, err := os.Open("../knowledgecatalog/catalog.json.gz")
	require.NoError(t, err)
	defer f.Close()
	r, err := gzip.NewReader(f)
	require.NoError(t, err)
	defer r.Close()
	var source struct {
		Items []struct {
			MeaningID  string `json:"meaning_id"`
			Word       string `json:"word"`
			POS        string `json:"part_of_speech"`
			Definition string `json:"definition"`
		}
	}
	require.NoError(t, json.NewDecoder(r).Decode(&source))
	canonical := map[string]string{}
	for _, item := range source.Items {
		word := item.Word
		if word == "i" {
			word = "I"
		}
		canonical[item.MeaningID] = meaningidentity.Revision(item.MeaningID, word, item.POS, item.Definition)
	}
	require.NotEmpty(t, entries)
	seen := map[string]bool{}
	for _, entry := range entries {
		require.False(t, seen[entry.MeaningID], entry.MeaningID)
		seen[entry.MeaningID] = true
		require.Equal(t, canonical[entry.MeaningID], entry.Revision, entry.Word)
		require.NotEmpty(t, entry.LearnerDefinition, entry.Word)
		require.NotEqual(t, entry.ShortDefinition, entry.LearnerDefinition, entry.Word)
	}
}

func TestEmbeddedDefinitionGuardsRejectInvalidContent(t *testing.T) {
	valid := entries[0]
	encode := func(rows ...entry) []byte {
		data, err := json.Marshal(rows)
		require.NoError(t, err)
		return data
	}
	empty := valid
	empty.LearnerDefinition = " \n "
	mismatched := valid
	mismatched.Revision = "not-the-canonical-revision"
	cases := []struct {
		name string
		data []byte
	}{
		{"malformed JSON", []byte(`[{`)},
		{"empty explanation", encode(empty)},
		{"mismatched revision", encode(mismatched)},
		{"duplicate revision", encode(valid, valid)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { require.Panics(t, func() { parseDefinitions(tc.data) }) })
	}
	t.Run("valid data keeps canonical binding", func(t *testing.T) {
		rows, index := parseDefinitions(encode(valid))
		require.Equal(t, []entry{valid}, rows)
		require.Equal(t, map[string]string{valid.Revision: valid.LearnerDefinition}, index)
	})
}
