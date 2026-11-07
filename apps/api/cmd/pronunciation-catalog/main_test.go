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

func TestRecoveryArgumentsReachDisabledConfigurationBeforeDatabase(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("PRONUNCIATION_PROVIDER", "openai")
	t.Setenv("PRONUNCIATION_API_KEY", "")
	t.Setenv("AI_FEATURES_ENABLED", "false")
	require.ErrorContains(t, run([]string{
		"--generate", "--job", "10000000-0000-4000-8000-000000000001",
		"--recover-failed-attempt", "20000000-0000-4000-8000-000000000002",
		"--max-generations", "1", "--max-characters", "100", "--max-audio-seconds", "60",
	}), "speech configuration unavailable")
}

func TestRecoveryRejectsUnsafeArgumentsBeforeConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("PRONUNCIATION_API_KEY", "")
	t.Setenv("AI_FEATURES_ENABLED", "false")
	const job = "abcdef00-0000-4000-8000-000000000001"
	const original = "abcdef00-0000-4000-8000-000000000002"
	args := func(jobID, attemptID, limit string) []string {
		return []string{"--generate", "--job", jobID, "--recover-failed-attempt", attemptID, "--max-generations", limit, "--max-characters", "100", "--max-audio-seconds", "60"}
	}
	for _, tc := range []struct {
		name    string
		args    []string
		message string
	}{
		{"without explicit generate", []string{"--job", job, "--recover-failed-attempt", original}, "requires explicit --generate"},
		{"empty explicit target", args(job, "", "1"), "canonical non-nil original attempt"},
		{"nil target", args(job, "00000000-0000-0000-0000-000000000000", "1"), "canonical non-nil original attempt"},
		{"malformed target", args(job, "bad", "1"), "canonical non-nil original attempt"},
		{"uppercase target", args(job, "ABCDEF00-0000-4000-8000-000000000002", "1"), "canonical non-nil original attempt"},
		{"URN target", args(job, "urn:uuid:"+original, "1"), "canonical non-nil original attempt"},
		{"uppercase job", args("ABCDEF00-0000-4000-8000-000000000001", original, "1"), "canonical non-nil new job"},
		{"missing job", args("", original, "1"), "canonical non-nil new job"},
		{"same UUID", args(original, original, "1"), "distinct from the original attempt"},
		{"two attempts", args(job, original, "2"), "max-generations exactly 1"},
		{"zero attempts", args(job, original, "0"), "max-generations exactly 1"},
		{"read-only validation combined", append(args(job, original, "1"), "--validate-audio"), "read-only"},
	} {
		t.Run(tc.name, func(t *testing.T) { require.ErrorContains(t, run(tc.args), tc.message) })
	}
}

func TestRecoverySpeechConfigurationFailsClosedWithoutDatabase(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	args := []string{"--generate", "--job", "10000000-0000-4000-8000-000000000001", "--recover-failed-attempt", "20000000-0000-4000-8000-000000000002", "--max-generations", "1", "--max-characters", "100", "--max-audio-seconds", "60"}
	for _, tc := range []struct{ name, provider, key, enabled string }{
		{"missing key", "openai", "", "true"},
		{"whitespace key", "openai", " \t", "true"},
		{"wrong provider", "cloudflare", "synthetic-not-used", "true"},
		{"disabled", "openai", "synthetic-not-used", "false"},
		{"disabled numeric", "openai", "synthetic-not-used", "0"},
		{"invalid flag", "openai", "synthetic-not-used", "invalid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("PRONUNCIATION_PROVIDER", tc.provider)
			t.Setenv("PRONUNCIATION_API_KEY", tc.key)
			t.Setenv("AI_FEATURES_ENABLED", tc.enabled)
			require.ErrorContains(t, run(args), "speech configuration unavailable")
		})
	}
}
