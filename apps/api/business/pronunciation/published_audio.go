package pronunciation

import (
	_ "embed"
	"encoding/hex"
	"sort"
	"strings"
)

// The generated index stores sorted pairs of 32-byte cache keys and MP3
// hashes. The cache key binds provider, meaning, exact text and context.
// See docs/operations/pronunciation-catalog.md for the refresh procedure.
//
//go:embed published_audio_index.bin
var publishedAudioIndex string

const publishedAudioBase = "https://audio.vocanova.site/meanings/"
const audioIndexRecordSize = 64

func publishedRecordingURL(meaningID, revision, cacheKey string) string {
	var key [32]byte
	if len(cacheKey) != hex.EncodedLen(len(key)) || len(publishedAudioIndex)%audioIndexRecordSize != 0 {
		return ""
	}
	if _, err := hex.Decode(key[:], []byte(cacheKey)); err != nil {
		return ""
	}
	target := string(key[:])
	count := len(publishedAudioIndex) / audioIndexRecordSize
	record := sort.Search(count, func(i int) bool {
		start := i * audioIndexRecordSize
		return strings.Compare(publishedAudioIndex[start:start+len(key)], target) >= 0
	})
	if record == count {
		return ""
	}
	start := record * audioIndexRecordSize
	if publishedAudioIndex[start:start+len(key)] != target {
		return ""
	}
	checksum := hex.EncodeToString([]byte(publishedAudioIndex[start+len(key) : start+audioIndexRecordSize]))
	return publishedAudioBase + meaningID + "/" + revision + "/" + cacheKey + "/" + checksum + ".mp3"
}
