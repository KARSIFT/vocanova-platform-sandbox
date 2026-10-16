package knowledgecatalog

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSemanticAliasesIgnoreEditorialProvenanceAndSelectedSense(t *testing.T) {
	for _, pair := range [][2]string{{"baggage", "luggage"}, {"bag", "luggage"}, {"bags", "luggage"}, {"luggage", "bags"}, {"dog", "animal"}, {"animal", "dog"}, {"answer", "reply"}, {"colleague", "co-worker"}, {"reservation", "booking"}, {"men", "man"}, {"Check-out", "check out"}, {"on_time", "on time"}} {
		same, err := SemanticAliases(pair[0], pair[1])
		require.NoError(t, err)
		require.True(t, same, pair)
	}
	for _, pair := range [][2]string{{"baggage", "menu"}, {"dog", "cat"}, {"refund", "landlord"}, {"signature", "pharmacy"}, {"", ""}} {
		same, err := SemanticAliases(pair[0], pair[1])
		require.NoError(t, err)
		require.False(t, same, pair)
	}
}
