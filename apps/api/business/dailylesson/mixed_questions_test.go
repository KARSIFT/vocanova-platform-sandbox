package dailylesson

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaningmedia"
	"github.com/stretchr/testify/require"
)

func TestMixedKindWeightedPrivateDraw(t *testing.T) {
	counts := map[string]int{}
	for n := byte(0); n < 100; n++ {
		kind, err := chooseMixedKind(bytes.NewReader([]byte{n}), true, true)
		require.NoError(t, err)
		counts[kind]++
	}
	require.Equal(t, map[string]int{"picture_choice": 60, "gap_choice": 30, "meaning_choice": 10}, counts)
	for _, tc := range []struct {
		picture, gap bool
		draw         byte
		want         string
	}{
		{true, false, 59, "picture_choice"}, {true, false, 60, "meaning_choice"},
		{false, true, 29, "gap_choice"}, {false, true, 30, "meaning_choice"},
		{false, false, 0, "meaning_choice"},
	} {
		got, err := chooseMixedKind(bytes.NewReader([]byte{tc.draw}), tc.picture, tc.gap)
		require.NoError(t, err)
		require.Equal(t, tc.want, got)
	}
	_, err := chooseMixedKind(bytes.NewReader(nil), true, true)
	require.True(t, errors.Is(err, io.EOF))
}

func TestMixedClozeRequiresExactlyOneWholeWord(t *testing.T) {
	for _, tc := range []struct{ sentence, target, want string }{
		{"Our flight leaves from gate twelve.", "gate", "Our flight leaves from _____ twelve."},
		{"Check-out is at eleven.", "check-out", "_____ is at eleven."},
		{"My key card stopped working.", "key card", "My _____ stopped working."},
		{"The gateway is open.", "gate", ""},
		{"The gate and gate are open.", "gate", ""},
		{"The gate's hinge broke.", "gate", ""},
		{"The key-card reader failed.", "card", ""},
		{"A gatekeeper saw the gate.", "gate", "A gatekeeper saw the _____."},
	} {
		got, ok := mixedCloze(tc.sentence, tc.target)
		require.Equal(t, tc.want != "", ok, tc.sentence)
		require.Equal(t, tc.want, got, tc.sentence)
	}
}

func TestMixedPairsRejectAmbiguityAndDuplicateContent(t *testing.T) {
	base := mixedCandidate{ID: "target", Word: "parcel", POS: "noun", Definition: "A wrapped object sent in the post.", SourceSense: "noun:parcel", Image: &meaningmedia.Image{SHA256: "target-hash"}}
	safe := mixedCandidate{ID: "alternative", Word: "menu", POS: "noun", Definition: "A list of dishes available to order.", SourceSense: "noun:menu", Image: &meaningmedia.Image{SHA256: "other-hash"}}
	require.True(t, safeMixedPair(base, safe))
	for _, mutate := range []func(*mixedCandidate){
		func(c *mixedCandidate) { c.ID = base.ID },
		func(c *mixedCandidate) { c.Word = "PARCEL" },
		func(c *mixedCandidate) { c.Word = "package" },
		func(c *mixedCandidate) { c.POS = "verb" },
		func(c *mixedCandidate) { c.Definition = base.Definition },
		func(c *mixedCandidate) { c.SourceSense = base.SourceSense },
		func(c *mixedCandidate) { c.Image = &meaningmedia.Image{SHA256: base.Image.SHA256} },
	} {
		candidate := safe
		mutate(&candidate)
		require.False(t, safeMixedPair(base, candidate))
	}
}
