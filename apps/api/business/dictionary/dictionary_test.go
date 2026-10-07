package dictionary

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDictionaryBundledProvenanceAndIntegrity(t *testing.T) {
	s := New()
	require.NotNil(t, s.data)
	require.Equal(t, "814ced411a9fe144c21920ae4b39054ef8c2b047309b40288707eba7d877f091", fmt.Sprintf("%x", sha256.Sum256(compressed)))
	notice, err := os.ReadFile("LICENSE.wordnet")
	require.NoError(t, err)
	require.Equal(t, string(notice), s.data.License)
	require.Len(t, s.data.Entries, 82710)
	require.Len(t, s.data.Exceptions, 4820)
	for word, ids := range s.data.Entries {
		_, err := NormalizeWord(word)
		require.NoError(t, err)
		require.NotEmpty(t, ids)
		counts := map[string]int{}
		for _, id := range ids {
			require.GreaterOrEqual(t, id, 0)
			require.Less(t, id, len(s.data.Senses))
			sense := s.data.Senses[id]
			require.Contains(t, []string{"noun", "verb", "adjective", "adverb"}, sense[0])
			require.NotEmpty(t, sense[1])
			counts[sense[0]]++
			require.LessOrEqual(t, counts[sense[0]], 2)
		}
	}
	for _, bases := range s.data.Exceptions {
		for _, base := range bases {
			require.NotEmpty(t, s.data.Entries[base])
		}
	}
}

func TestDictionaryRealEntriesAndInflections(t *testing.T) {
	s := New()
	for input, expected := range map[string]string{
		" Serendipity ": "serendipity", "APPLE": "apple", "book": "book",
		"books": "book", "went": "go", "children": "child", "walking": "walking",
		"apologized": "apologize", "quicker": "quicker", "tallest": "tall", "mother-in-law": "mother-in-law",
	} {
		t.Run(input, func(t *testing.T) {
			entry, err := s.Lookup(context.Background(), input)
			require.NoError(t, err)
			require.Equal(t, expected, entry.Word)
			require.NotEmpty(t, entry.Meanings)
			require.Contains(t, entry.License, "Copyright 2006")
		})
	}
	entry, err := s.Lookup(context.Background(), "serendipity")
	require.NoError(t, err)
	require.Equal(t, "good luck in making unexpected and fortunate discoveries", entry.Meanings[0].Definitions[0].Definition)
}

func TestDictionaryInvalidMissingAndUnavailableAreDistinct(t *testing.T) {
	s := New()
	for _, input := range []string{"", "  ", "hello world", "https://example.com", "<script>", "café", "123", "word\nword", "word--word", strings.Repeat("a", 49)} {
		_, err := s.Lookup(context.Background(), input)
		require.ErrorIs(t, err, ErrInvalid)
	}
	_, err := s.Lookup(context.Background(), "zzqvnonexistent")
	require.ErrorIs(t, err, ErrNotFound)
	_, err = (&Service{}).Lookup(context.Background(), "word")
	require.ErrorIs(t, err, ErrUnavailable)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = s.Lookup(ctx, "word")
	require.ErrorIs(t, err, ErrUnavailable)
}

func TestDictionaryReturnedEntriesCannotMutateSharedIndex(t *testing.T) {
	s := New()
	entry, err := s.Lookup(context.Background(), "serendipity")
	require.NoError(t, err)
	entry.Meanings[0].Definitions[0].Definition = "changed"
	entry.Meanings[0].PartOfSpeech = "changed"
	entry.License = "changed"
	next, err := s.Lookup(context.Background(), "serendipity")
	require.NoError(t, err)
	require.NotEqual(t, "changed", next.Meanings[0].Definitions[0].Definition)
	require.Equal(t, "noun", next.Meanings[0].PartOfSpeech)
	require.Contains(t, next.License, "Copyright 2006")
}

func TestDictionaryParallelReadAndStartup(t *testing.T) {
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			entry, err := New().Lookup(context.Background(), "serendipity")
			if err != nil || entry.Word != "serendipity" {
				t.Errorf("parallel read: %v", err)
			}
		})
	}
	wg.Wait()
}

func BenchmarkDictionaryLookup(b *testing.B) {
	s := New()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := s.Lookup(context.Background(), "serendipity"); err != nil {
			b.Fatal(err)
		}
	}
}
