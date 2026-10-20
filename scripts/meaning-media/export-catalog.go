// Export the public, canonical content union without a database or credentials.
// Run from apps/api: go run ../../scripts/meaning-media/export-catalog.go -root ../.. -out ../../.work/meaning-media/catalog.json
package main

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaninguses"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
)

type meaning struct {
	MeaningID       string   `json:"meaningId"`
	WordID          string   `json:"wordId"`
	Word            string   `json:"word"`
	PartOfSpeech    string   `json:"partOfSpeech"`
	Definition      string   `json:"definition"`
	MeaningRevision string   `json:"meaningRevision"`
	Sources         []string `json:"sources"`
	Rank            int      `json:"rank,omitempty"`
	Existing        *picture `json:"existing,omitempty"`
}
type picture struct {
	Src        string `json:"src"`
	Alt        string `json:"alt"`
	Status     string `json:"status"`
	Word       string `json:"word"`
	Definition string `json:"definition"`
}

func main() {
	root := flag.String("root", ".", "Repository root")
	out := flag.String("out", "", "Output path (required; keep working queues outside Git)")
	flag.Parse()
	if *out == "" {
		fail(fmt.Errorf("-out is required"))
	}
	rows := map[string]*meaning{}
	add := func(m meaning, source string) {
		m.MeaningRevision = meaningidentity.Revision(m.MeaningID, m.Word, m.PartOfSpeech, m.Definition)
		if old, ok := rows[m.MeaningID]; ok {
			if old.MeaningRevision != m.MeaningRevision || old.WordID != m.WordID {
				fail(fmt.Errorf("conflicting canonical meaning %s", m.MeaningID))
			}
			old.Sources = append(old.Sources, source)
			return
		}
		m.Sources = []string{source}
		rows[m.MeaningID] = &m
	}
	var seed struct {
		Words    []struct{ ID, Text, Status string } `json:"canonical_words"`
		Meanings []struct {
			ID         string
			WordID     string `json:"word_id"`
			POS        string `json:"part_of_speech"`
			Definition string `json:"short_definition"`
			Status     string
		} `json:"word_meanings"`
	}
	readJSON(filepath.Join(*root, "apps/api/cmd/seed/voc026-p1.json"), &seed)
	words := map[string]string{}
	for _, w := range seed.Words {
		if w.Status == "active" {
			words[w.ID] = w.Text
		}
	}
	// The authored curriculum is seeded first and the map deliberately preserves it.
	for _, m := range seed.Meanings {
		if m.Status == "active" {
			word, ok := words[m.WordID]
			if !ok {
				fail(fmt.Errorf("missing active seed word"))
			}
			add(meaning{MeaningID: m.ID, WordID: m.WordID, Word: word, PartOfSpeech: m.POS, Definition: m.Definition}, "curriculum")
		}
	}
	f, err := os.Open(filepath.Join(*root, "apps/api/business/knowledgecatalog/catalog.json.gz"))
	if err != nil {
		fail(err)
	}
	z, err := gzip.NewReader(f)
	if err != nil {
		fail(err)
	}
	var catalog struct {
		Items []struct {
			MeaningID  string `json:"meaning_id"`
			WordID     string `json:"word_id"`
			Word       string `json:"word"`
			POS        string `json:"part_of_speech"`
			Definition string `json:"definition"`
			Rank       int    `json:"rank"`
		} `json:"items"`
	}
	if err = json.NewDecoder(z).Decode(&catalog); err != nil {
		fail(err)
	}
	z.Close()
	f.Close()
	for _, m := range catalog.Items {
		word := m.Word
		if word == "i" {
			word = "I"
		}
		add(meaning{MeaningID: m.MeaningID, WordID: m.WordID, Word: word, PartOfSpeech: m.POS, Definition: m.Definition, Rank: m.Rank}, "knowledge-map")
		rows[m.MeaningID].Rank = m.Rank
	}
	for _, m := range meaninguses.Entries() {
		add(meaning{MeaningID: m.MeaningID.String(), WordID: m.WordID.String(), Word: m.Word, PartOfSpeech: m.PartOfSpeech, Definition: m.Definition}, "reviewed-supplement")
	}
	var pictures map[string]picture
	readJSON(filepath.Join(*root, "apps/web/src/lib/meaning-pictures.json"), &pictures)
	ready := 0
	for id, p := range pictures {
		if p.Status != "ready" {
			continue
		}
		m, ok := rows[id]
		if !ok || m.Word != p.Word || m.Definition != p.Definition {
			fail(fmt.Errorf("stale existing picture %s", id))
		}
		p := p
		m.Existing = &p
		ready++
	}
	list := make([]*meaning, 0, len(rows))
	for _, m := range rows {
		list = append(list, m)
	}
	sort.Slice(list, func(i, j int) bool {
		a, b := list[i], list[j]
		if (a.Existing != nil) != (b.Existing != nil) {
			return a.Existing != nil
		}
		if (a.Rank > 0) != (b.Rank > 0) {
			return a.Rank > 0
		}
		if a.Rank != b.Rank {
			return a.Rank < b.Rank
		}
		return a.MeaningID < b.MeaningID
	})
	data, err := json.MarshalIndent(struct {
		SchemaVersion int        `json:"schemaVersion"`
		Meanings      []*meaning `json:"meanings"`
	}{1, list}, "", "  ")
	if err != nil {
		fail(err)
	}
	if err = os.MkdirAll(filepath.Dir(*out), 0700); err != nil {
		fail(err)
	}
	if err = os.WriteFile(*out, append(data, '\n'), 0600); err != nil {
		fail(err)
	}
	fmt.Printf("Canonical meanings: %d; reviewed existing images: %d; missing: %d\n", len(list), ready, len(list)-ready)
}
func readJSON(path string, v any) {
	b, err := os.ReadFile(path)
	if err != nil {
		fail(err)
	}
	if err = json.Unmarshal(b, v); err != nil {
		fail(err)
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, "meaning-media export:", err); os.Exit(1) }
