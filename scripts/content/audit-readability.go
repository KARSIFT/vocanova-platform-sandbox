// Command audit-readability emits deterministic public-catalog editorial triage.
// Run from apps/api: go run ../../scripts/content/audit-readability.go -root ../..
// This command reads checked-in source only; it never contacts a database or API.
package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/learnerdefinitions"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaninguses"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
)

type word struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
type meaning struct {
	ID         string `json:"id"`
	WordID     string `json:"word_id"`
	POS        string `json:"part_of_speech"`
	Definition string `json:"short_definition"`
	Learner    string `json:"learner_definition"`
}
type ranked struct {
	Rank       int    `json:"rank"`
	Word       string `json:"word"`
	ID         string `json:"meaning_id"`
	WordID     string `json:"word_id"`
	POS        string `json:"part_of_speech"`
	Definition string `json:"definition"`
	Source     string `json:"source"`
	Sense      string `json:"source_sense_key"`
}
type row struct {
	MeaningID             string `json:"meaningId"`
	Word                  string `json:"word"`
	PartOfSpeech          string `json:"partOfSpeech"`
	Rank                  int    `json:"rank,omitempty"`
	Revision              string `json:"canonicalRevision"`
	Source                string `json:"source"`
	SourceSenseKey        string `json:"sourceSenseKey,omitempty"`
	Words                 int    `json:"definitionWords"`
	Semicolons            int    `json:"semicolons"`
	SynonymListCandidate  bool   `json:"shortListCandidate"`
	LongCandidate         bool   `json:"longDefinitionCandidate"`
	TeachingSenses        int    `json:"availableTeachingSenses"`
	TeachingPOS           int    `json:"availableTeachingPartsOfSpeech"`
	ReferenceSenses       int    `json:"historicalWordNetSenseMemberships"`
	ReferencePOS          int    `json:"historicalWordNetPartsOfSpeech"`
	RepeatedDefinition    int    `json:"sameTextDefinitionCount"`
	HasLearnerExplanation bool   `json:"hasLearnerExplanation"`
	HasReviewedOverlay    bool   `json:"hasReviewedOverlay"`
	definition            string
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func read(root, path string, out any, hashes map[string]string) {
	data, err := os.ReadFile(filepath.Join(root, path))
	must(err)
	sum := sha256.Sum256(data)
	hashes[path] = hex.EncodeToString(sum[:])
	if strings.HasSuffix(path, ".gz") {
		r, err := gzip.NewReader(bytes.NewReader(data))
		must(err)
		data, err = io.ReadAll(r)
		must(err)
		must(r.Close())
	}
	must(json.Unmarshal(data, out))
}
func main() {
	root := flag.String("root", "../..", "repository root, relative to current working directory")
	flag.Parse()
	hashes := map[string]string{}
	var catalog struct {
		Items []ranked `json:"items"`
	}
	read(*root, "apps/api/business/knowledgecatalog/catalog.json.gz", &catalog, hashes)
	if len(catalog.Items) != 10000 {
		panic("expected frozen 10,000-word catalog")
	}
	var seed struct {
		Words    []word    `json:"canonical_words"`
		Meanings []meaning `json:"word_meanings"`
	}
	read(*root, "apps/api/cmd/seed/voc026-p1.json", &seed, hashes)
	var aliases struct {
		Entries map[string][]string `json:"entries"`
	}
	read(*root, "apps/api/business/knowledgecatalog/semantic-aliases.json.gz", &aliases, hashes)
	var overlays []map[string]any
	read(*root, "apps/api/business/learnerdefinitions/definitions.json", &overlays, hashes)
	wordNames := map[string]string{}
	authored := map[string]string{}
	byID := map[string]row{}
	for _, w := range seed.Words {
		wordNames[w.ID] = w.Text
	}
	for _, m := range seed.Meanings {
		authored[m.ID] = m.Learner
		byID[m.ID] = row{MeaningID: m.ID, Word: wordNames[m.WordID], PartOfSpeech: m.POS, Source: "Vocanova authored", definition: m.Definition}
	}
	for _, item := range catalog.Items {
		text := item.Word
		if text == "i" {
			text = "I"
		}
		if existing, ok := byID[item.ID]; ok && (existing.Word != text || existing.PartOfSpeech != item.POS || existing.definition != item.Definition) {
			panic("authored/catalog identity mismatch: " + item.Word)
		}
		byID[item.ID] = row{MeaningID: item.ID, Word: text, PartOfSpeech: item.POS, Rank: item.Rank, Source: item.Source, SourceSenseKey: item.Sense, definition: item.Definition}
	}
	for _, use := range meaninguses.Entries() {
		authored[use.MeaningID.String()] = use.Definition // Supplemental seed stores the same learner_definition.
		if _, ok := byID[use.MeaningID.String()]; ok {
			panic("duplicate supplemental identity")
		}
		byID[use.MeaningID.String()] = row{MeaningID: use.MeaningID.String(), Word: use.Word, PartOfSpeech: use.PartOfSpeech, Source: "Vocanova supplemental use", SourceSenseKey: use.SourceSenseKey, definition: use.Definition}
	}
	senses := map[string]int{}
	parts := map[string]map[string]bool{}
	duplicates := map[string]int{}
	for _, item := range byID {
		key := strings.ToLower(item.Word)
		senses[key]++
		if parts[key] == nil {
			parts[key] = map[string]bool{}
		}
		parts[key][item.PartOfSpeech] = true
		duplicates[strings.ToLower(strings.Join(strings.Fields(item.definition), " "))]++
	}
	rows := make([]row, 0, len(byID))
	counts := map[string]int{}
	sourceCounts := map[string]int{}
	rankedSources := map[string]int{}
	for _, item := range byID {
		item.Revision = meaningidentity.Revision(item.MeaningID, item.Word, item.PartOfSpeech, item.definition)
		if item.Revision == "" {
			panic("incomplete meaning identity")
		}
		item.Words = len(strings.Fields(item.definition))
		item.Semicolons = strings.Count(item.definition, ";")
		item.LongCandidate = item.Words > 20
		item.SynonymListCandidate = item.Words <= 5 && strings.ContainsAny(item.definition, ",;")
		key := strings.ToLower(item.Word)
		item.TeachingSenses = senses[key]
		item.TeachingPOS = len(parts[key])
		item.ReferenceSenses = len(aliases.Entries[key])
		referencePOS := map[string]bool{}
		for _, sense := range aliases.Entries[key] {
			referencePOS[strings.SplitN(sense, ":", 2)[0]] = true
		}
		item.ReferencePOS = len(referencePOS)
		item.RepeatedDefinition = duplicates[strings.ToLower(strings.Join(strings.Fields(item.definition), " "))]
		item.HasReviewedOverlay = learnerdefinitions.Resolve(item.MeaningID, item.Word, item.PartOfSpeech, item.definition, "") != ""
		item.HasLearnerExplanation = learnerdefinitions.Resolve(item.MeaningID, item.Word, item.PartOfSpeech, item.definition, authored[item.MeaningID]) != ""
		counts["canonicalMeanings"]++
		sourceCounts[item.Source]++
		if item.Rank > 0 {
			counts["rankedMeanings"]++
			rankedSources[item.Source]++
			if item.LongCandidate {
				counts["rankedLongDefinitions"]++
			}
			if item.SynonymListCandidate {
				counts["rankedShortListCandidates"]++
			}
			if item.LongCandidate || item.SynonymListCandidate || strings.EqualFold(item.Word, "abandon") {
				counts["rankedReviewUnion"]++
			}
		}
		if item.HasReviewedOverlay {
			counts["reviewedOriginalExplanations"]++
		}
		if item.HasLearnerExplanation {
			counts["meaningsWithLearnerExplanation"]++
		}
		if item.TeachingSenses > 1 {
			counts["meaningsWithMultipleTeachingSenses"]++
		}
		if item.ReferenceSenses > 1 {
			counts["meaningsWithMultipleHistoricalReferenceSenses"]++
		}
		rows = append(rows, item)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Rank > 0 && rows[j].Rank > 0 {
			return rows[i].Rank < rows[j].Rank
		}
		if rows[i].Rank > 0 {
			return true
		}
		if rows[j].Rank > 0 {
			return false
		}
		if rows[i].Word != rows[j].Word {
			return rows[i].Word < rows[j].Word
		}
		return rows[i].MeaningID < rows[j].MeaningID
	})
	if counts["rankedMeanings"] != 10000 || counts["reviewedOriginalExplanations"] != len(overlays) {
		panic(fmt.Sprintf("unexpected audit coverage: %v", counts))
	}
	result := struct {
		Version       string            `json:"version"`
		Limitations   []string          `json:"limitations"`
		InputSHA256   map[string]string `json:"inputSHA256"`
		Counts        map[string]int    `json:"counts"`
		Sources       map[string]int    `json:"sources"`
		RankedSources map[string]int    `json:"rankedSources"`
		Rows          []row             `json:"rows"`
	}{"catalog-readability-triage-v1", []string{"Word-count and punctuation flags are editorial triage, not CEFR ratings or learner-comprehension evidence.", "Historical WordNet sense membership is not a recommendation to teach every sense or evidence that all uses are represented.", "Current teaching-sense counts describe checked-in seed sources, not a live database audit.", "Reviewed explanation overlays preserve canonical identity, grading, rank, approved media and learner state."}, hashes, counts, sourceCounts, rankedSources, rows}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	must(encoder.Encode(result))
}
