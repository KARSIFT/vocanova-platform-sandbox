// Command pronunciation-audition attempts at most six public, fixed teaching
// samples in GitHub CI. It never accepts learner text, provider configuration,
// model choices or credentials through command-line arguments.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaninguses"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/pronunciation"
)

type auditionCase struct {
	Name                 string                      `json:"name"`
	Utterance            pronunciation.Utterance     `json:"utterance"`
	Status               string                      `json:"status"`
	File                 string                      `json:"file,omitempty"`
	Bytes                int                         `json:"bytes,omitempty"`
	SHA256               string                      `json:"sha256,omitempty"`
	Quality              *pronunciation.AudioQuality `json:"quality,omitempty"`
	DurationMilliseconds int64                       `json:"durationMilliseconds,omitempty"`
}
type manifest struct {
	SchemaVersion int                            `json:"schemaVersion"`
	GeneratedAt   time.Time                      `json:"generatedAt"`
	AIGenerated   bool                           `json:"aiGenerated"`
	Identity      pronunciation.ProviderIdentity `json:"identity"`
	Cases         []auditionCase                 `json:"cases"`
}

func fixedCases() []auditionCase {
	entries := []struct{ name, id, text, pos, definition, example string }{
		{"record-noun", "009949f4-3605-530f-84f4-301a6beeb809", "record", "noun", "Information kept so that it can be checked later.", "Keep a record of every payment."},
		{"record-verb", "45cb435d-695a-5acb-bd3e-f227ab6fe217", "record", "verb", "To store information, sound, or pictures so that they can be used later.", "We record every lesson."},
		{"light-adjective", "6f4e156d-64fd-5005-9ed5-f0f7d18157d1", "light", "adjective", "Having little weight.", "This light bag is easy to carry."},
		{"fast-adverb", "1187a4d8-c4df-5f17-acb2-86b2382b12c1", "fast", "adverb", "At high speed.", "The train moves fast."},
		{"read-present", "9e8813c5-81d2-584e-8197-882110c19402", "read", "verb", "To look at written words and understand what they mean.", "I read a book every evening."},
		{"read-past", "732bb481-ef1b-57c7-aa9c-7aff62c2391e", "read", "verb", "Past tense or past participle of read: looked at and understood written words before now.", "Yesterday, I read the whole story."},
	}
	out := make([]auditionCase, 0, len(entries))
	for _, e := range entries {
		out = append(out, auditionCase{Name: e.name, Status: "not_attempted", Utterance: pronunciation.Utterance{MeaningID: e.id, WordText: e.text, Text: e.text, Kind: "word", PartOfSpeech: e.pos, Definition: e.definition, Context: e.example, Locale: "en-US"}})
	}
	return out
}

// writeManifest atomically preserves prior successes after any later failure.
// All fields are fixed public teaching content or redacted generation status.
func writeManifest(directory string, result manifest) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(directory, ".audition-manifest-*.json")
	if err != nil {
		return err
	}
	name := temporary.Name()
	defer os.Remove(name)
	if _, err = temporary.Write(append(data, '\n')); err != nil {
		temporary.Close()
		return err
	}
	if err = temporary.Close(); err != nil {
		return err
	}
	return os.Rename(name, filepath.Join(directory, "manifest.json"))
}
func run(args []string) int {
	fs := flag.NewFlagSet("pronunciation-audition", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	directory := fs.String("output-dir", "", "Directory for six public MP3 samples and manifest.json")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 || strings.TrimSpace(*directory) == "" {
		fmt.Fprintln(os.Stderr, "usage: pronunciation-audition --output-dir DIRECTORY")
		return 2
	}
	if os.Getenv("CI") != "true" || os.Getenv("GITHUB_ACTIONS") != "true" || strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) == "" {
		fmt.Fprintln(os.Stderr, "audition configuration unavailable; GitHub CI and its existing OpenAI secret are required")
		return 2
	}
	cases := fixedCases()
	if len(cases) != 6 {
		fmt.Fprintln(os.Stderr, "invalid fixed audition set")
		return 2
	}
	// Refuse drift between audition guidance and the server's reviewed catalog.
	for _, sample := range cases {
		reviewed, ok := meaninguses.ReviewedContext(sample.Utterance.MeaningID, sample.Utterance.Text, sample.Utterance.Definition)
		if !ok || reviewed != sample.Utterance.Context {
			fmt.Fprintln(os.Stderr, "audition guidance does not match reviewed catalog")
			return 2
		}
	}

	if err := os.MkdirAll(*directory, 0700); err != nil {
		fmt.Fprintln(os.Stderr, "audition output unavailable")
		return 2
	}
	// Never overwrite a previous run or silently pay again into its output set.
	for _, name := range append([]string{"manifest.json"}, "record-noun.mp3", "record-verb.mp3", "light-adjective.mp3", "fast-adverb.mp3", "read-present.mp3", "read-past.mp3") {
		if _, err := os.Lstat(filepath.Join(*directory, name)); !os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "audition output must be fresh")
			return 2
		}
	}
	provider := pronunciation.NewOpenAIProvider(pronunciation.OpenAIConfig{APIKey: os.Getenv("OPENAI_API_KEY")})
	result := manifest{SchemaVersion: 1, GeneratedAt: time.Now().UTC(), AIGenerated: true, Identity: provider.Identity(), Cases: cases}
	if err := writeManifest(*directory, result); err != nil {
		fmt.Fprintln(os.Stderr, "audition manifest unavailable")
		return 2
	}
	failed := false
	for i := range result.Cases {
		sample := &result.Cases[i]
		started := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		data, err := provider.Generate(ctx, sample.Utterance)
		cancel()
		sample.DurationMilliseconds = time.Since(started).Milliseconds()
		if err != nil {
			sample.Status = "provider_unavailable"
			failed = true
		} else {
			quality, qualityErr := pronunciation.ValidateAudio(data)
			sample.Quality = &quality
			if qualityErr != nil {
				sample.Status = "audio_quality_failed"
				failed = true
			}
		}
		if !failed || (sample.Quality != nil && sample.Quality.SampleRate > 0) {
			name := sample.Name + ".mp3"
			if failed {
				name = "rejected-" + name
			}
			file, writeErr := os.OpenFile(filepath.Join(*directory, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
			if writeErr == nil {
				_, writeErr = file.Write(data)
				closeErr := file.Close()
				if writeErr == nil {
					writeErr = closeErr
				}
			}
			if writeErr != nil {
				// Remove only a partial file created by this attempt; never remove a
				// pre-existing file when exclusive creation itself failed.
				if file != nil {
					_ = os.Remove(filepath.Join(*directory, name))
				}
				sample.Status = "output_unavailable"
				failed = true
			} else {
				sum := sha256.Sum256(data)
				sample.Status = "generated"
				if failed {
					sample.Status = "audio_rejected"
				}
				sample.File = name
				sample.Bytes = len(data)
				sample.SHA256 = hex.EncodeToString(sum[:])
			}
		}
		fmt.Fprintf(os.Stdout, "%s: %s\n", sample.Name, sample.Status)
		if err := writeManifest(*directory, result); err != nil {
			fmt.Fprintln(os.Stderr, "audition manifest unavailable; prior public samples remain saved")
			return 2
		}
		// Stop paid work at the first failure. Later cases remain not_attempted,
		// and the manifest records both this failure and all earlier saved clips.
		if failed {
			return 1
		}
	}
	return 0
}
func main() { os.Exit(run(os.Args[1:])) }
