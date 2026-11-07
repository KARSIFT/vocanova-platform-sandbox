package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInvalidCatalogArgumentsNeverReachDatabaseOrProvider(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("PRONUNCIATION_API_KEY", "")
	for _, args := range [][]string{{"--generate"}, {"--job", "not-a-uuid"}, {"--max-generations", "5"}, {"--generate", "--job", "00000000-0000-0000-0000-000000000000"}, {"--unknown"}} {
		require.Error(t, run(args))
	}
	require.ErrorContains(t, run(nil), "DATABASE_URL")
}

func TestDisabledOrInvalidSpeechFlagStopsBeforeDatabase(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("PRONUNCIATION_PROVIDER", "openai")
	t.Setenv("PRONUNCIATION_API_KEY", "synthetic-key-not-used")
	for _, value := range []string{"false", "False", "FALSE", "0", "f", "F", "invalid"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("AI_FEATURES_ENABLED", value)
			require.ErrorContains(t, run([]string{"--generate", "--job", "10000000-0000-4000-8000-000000000001", "--max-generations", "1", "--max-characters", "100", "--max-audio-seconds", "60"}), "speech configuration unavailable")
		})
	}
}
