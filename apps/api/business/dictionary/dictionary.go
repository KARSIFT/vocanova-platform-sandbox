// Package dictionary serves an immutable local WordNet excerpt. Entries are
// reading aids, not canonical learning meanings, and have no saved/grading IDs.
package dictionary

import (
	"bytes"
	"compress/gzip"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strings"
	"sync"
)

var (
	ErrInvalid     = errors.New("enter one English word, up to 48 letters with optional internal apostrophes or hyphens")
	ErrNotFound    = errors.New("dictionary word not found")
	ErrUnavailable = errors.New("dictionary temporarily unavailable")
	wordPattern    = regexp.MustCompile(`^[a-z]+(?:['-][a-z]+)*$`)
)

const SourceURL = "https://wordnet.princeton.edu/"
const LicenseURL = "https://wordnet.princeton.edu/license-and-commercial-use"

//go:embed wordnet.json.gz
var compressed []byte

type Definition struct{ Definition, Example string }
type Meaning struct {
	PartOfSpeech string
	Definitions  []Definition
}
type Entry struct {
	Word     string
	Meanings []Meaning
	License  string
}

type index struct {
	License    string              `json:"license"`
	Entries    map[string][]int    `json:"entries"`
	Senses     [][3]string         `json:"senses"`
	Exceptions map[string][]string `json:"exceptions"`
}

var loadOnce sync.Once
var bundled *index

type Service struct{ data *index }

// New loads the bounded embedded index once at API startup. Lookup itself does
// no network, disk access, decompression, or learner-state reads or writes.
func New() *Service {
	loadOnce.Do(func() {
		reader, err := gzip.NewReader(bytes.NewReader(compressed))
		if err != nil {
			return
		}
		defer reader.Close()
		data, err := io.ReadAll(io.LimitReader(reader, 16*1024*1024+1))
		if err != nil || len(data) > 16*1024*1024 {
			return
		}
		var parsed index
		if json.Unmarshal(data, &parsed) != nil || parsed.License == "" || len(parsed.Entries) == 0 {
			return
		}
		bundled = &parsed
	})
	return &Service{data: bundled}
}

func NormalizeWord(raw string) (string, error) {
	word := strings.ToLower(strings.TrimSpace(raw))
	if len(word) == 0 || len(word) > 48 || !wordPattern.MatchString(word) {
		return "", ErrInvalid
	}
	return word, nil
}

func (s *Service) Lookup(ctx context.Context, raw string) (Entry, error) {
	word, err := NormalizeWord(raw)
	if err != nil {
		return Entry{}, err
	}
	if ctx.Err() != nil || s.data == nil {
		return Entry{}, ErrUnavailable
	}
	ids := s.data.Entries[word]
	if len(ids) == 0 {
		for _, base := range s.data.Exceptions[word] {
			if len(s.data.Entries[base]) > 0 {
				word, ids = base, s.data.Entries[base]
				break
			}
		}
	}
	if len(ids) == 0 {
		// WordNet's documented suffix substitutions are candidates only. A
		// candidate must exist with the corresponding part of speech; never
		// invent a word or definition by blindly removing a suffix.
		for _, rule := range morphology {
			if !strings.HasSuffix(word, rule.suffix) {
				continue
			}
			base := strings.TrimSuffix(word, rule.suffix) + rule.replacement
			for _, id := range s.data.Entries[base] {
				if s.data.Senses[id][0] == rule.part {
					word, ids = base, s.data.Entries[base]
					break
				}
			}
			if len(ids) > 0 {
				break
			}
		}
	}
	if len(ids) == 0 {
		return Entry{}, ErrNotFound
	}
	out := Entry{Word: word, Meanings: []Meaning{}, License: s.data.License}
	for _, id := range ids {
		if id < 0 || id >= len(s.data.Senses) {
			return Entry{}, ErrUnavailable
		}
		sense := s.data.Senses[id]
		position := -1
		for i, meaning := range out.Meanings {
			if meaning.PartOfSpeech == sense[0] {
				position = i
				break
			}
		}
		if position < 0 {
			out.Meanings = append(out.Meanings, Meaning{PartOfSpeech: sense[0], Definitions: []Definition{}})
			position = len(out.Meanings) - 1
		}
		if len(out.Meanings[position].Definitions) < 2 {
			out.Meanings[position].Definitions = append(out.Meanings[position].Definitions, Definition{Definition: sense[1], Example: sense[2]})
		}
	}
	return out, nil
}

var morphology = []struct{ part, suffix, replacement string }{
	{"noun", "s", ""}, {"noun", "ses", "s"}, {"noun", "xes", "x"},
	{"noun", "zes", "z"}, {"noun", "ches", "ch"}, {"noun", "shes", "sh"},
	{"noun", "men", "man"}, {"noun", "ies", "y"},
	{"verb", "s", ""}, {"verb", "ies", "y"}, {"verb", "es", "e"},
	{"verb", "es", ""}, {"verb", "ed", "e"}, {"verb", "ed", ""},
	{"verb", "ing", "e"}, {"verb", "ing", ""},
	{"adjective", "er", ""}, {"adjective", "est", ""},
	{"adjective", "er", "e"}, {"adjective", "est", "e"},
}
