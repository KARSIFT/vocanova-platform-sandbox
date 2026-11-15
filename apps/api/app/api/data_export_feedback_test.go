package api

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFeedbackPersonalDataProjectionRetainsPositionAndEffort(t *testing.T) {
	raw := []byte(`{"discoveryPosition":{"catalogVersion":"english-10000-v1","band":44,"rank":4330,"active":true,"revision":2,"updatedAt":"2026-11-15T12:00:00Z"},"dailyReviewEffort":[{"wordId":"word","meaningId":"meaning","localDate":"2026-11-15","reviewAttemptId":"attempt","dailySessionId":"session","createdAt":"2026-11-15T12:00:00Z"}]}`)
	var dto PersonalDataExportDTO
	require.NoError(t, json.Unmarshal(raw, &dto))
	encoded, err := json.Marshal(dto)
	require.NoError(t, err)
	var actual, want map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &want))
	require.NoError(t, json.Unmarshal(encoded, &actual))
	for _, key := range []string{"discoveryPosition", "dailyReviewEffort"} {
		require.Contains(t, actual, key)
		require.JSONEq(t, string(want[key]), string(actual[key]))
	}
}
