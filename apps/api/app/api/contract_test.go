package api

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestContractContainsCurrentUserWithoutSensitiveFields(t *testing.T) {
	document, err := json.Marshal(NewContractAPI().OpenAPI())
	if err != nil {
		t.Fatalf("marshal OpenAPI: %v", err)
	}
	contract := string(document)
	for _, expected := range []string{"GetCurrentUser", "/api/v1/me", "displayName"} {
		if !strings.Contains(contract, expected) {
			t.Errorf("OpenAPI missing %q", expected)
		}
	}
	for _, forbidden := range []string{"token_hash", "provider_subject", "revoked_at"} {
		if strings.Contains(contract, forbidden) {
			t.Errorf("OpenAPI exposed internal field %q", forbidden)
		}
	}
}

func TestContractContainsDiscoveryEndpoints(t *testing.T) {
	document, err := json.Marshal(NewContractAPI().OpenAPI())
	if err != nil {
		t.Fatalf("marshal OpenAPI: %v", err)
	}
	contract := string(document)
	for _, expected := range []string{"ListJourneySituations", "GetJourneySituation", "GetCanonicalWord", "/api/v1/journey-situations", "/api/v1/canonical-words/{wordSlug}"} {
		if !strings.Contains(contract, expected) {
			t.Errorf("OpenAPI missing %q", expected)
		}
	}
	for _, forbidden := range []string{"token_hash", "provider_subject", "revoked_at", "deleted_at", "user_id", "meaning_id"} {
		if strings.Contains(contract, forbidden) {
			t.Errorf("OpenAPI exposed internal field %q", forbidden)
		}
	}
}

func TestContractContainsLearningEndpoints(t *testing.T) {
	document, err := json.Marshal(NewContractAPI().OpenAPI())
	if err != nil {
		t.Fatalf("marshal OpenAPI: %v", err)
	}
	contract := string(document)
	for _, expected := range []string{"ListSavedWords", "SaveUserWord", "UnsaveUserWord", "/api/v1/user-words", "Idempotency-Key"} {
		if !strings.Contains(contract, expected) {
			t.Errorf("OpenAPI missing %q", expected)
		}
	}
	for _, forbidden := range []string{"token_hash", "provider_subject", "revoked_at", "deleted_at", "user_id"} {
		if strings.Contains(contract, forbidden) {
			t.Errorf("OpenAPI exposed internal field %q", forbidden)
		}
	}
}

func TestContractContainsReviewEndpoints(t *testing.T) {
	document, err := json.Marshal(NewContractAPI().OpenAPI())
	if err != nil {
		t.Fatalf("marshal OpenAPI: %v", err)
	}
	contract := string(document)
	for _, expected := range []string{"GetReviewsDue", "SubmitReview", "/api/v1/reviews/due", "/api/v1/reviews/submissions", "Idempotency-Key"} {
		if !strings.Contains(contract, expected) {
			t.Errorf("OpenAPI missing %q", expected)
		}
	}
	for _, forbidden := range []string{"token_hash", "provider_subject", "revoked_at", "deleted_at", "user_id"} {
		if strings.Contains(contract, forbidden) {
			t.Errorf("OpenAPI exposed internal field %q", forbidden)
		}
	}
}

func TestContractContainsMissionsEndpoints(t *testing.T) {
	document, err := json.Marshal(NewContractAPI().OpenAPI())
	if err != nil {
		t.Fatalf("marshal OpenAPI: %v", err)
	}
	contract := string(document)
	for _, expected := range []string{
		"GetDailyMission", "GetProgress",
		"/api/v1/daily-mission", "/api/v1/progress",
		"DailyMissionDTO", "ProgressDTO", "StreakDTO", "CompletionDayDTO",
		"confidencePointsBalance", "completionHistory", "graceDayBalance",
	} {
		if !strings.Contains(contract, expected) {
			t.Errorf("OpenAPI missing %q", expected)
		}
	}
	for _, forbidden := range []string{"token_hash", "provider_subject", "revoked_at", "deleted_at", "user_id", "idempotency_key", "source_id"} {
		if strings.Contains(contract, forbidden) {
			t.Errorf("OpenAPI exposed internal field %q", forbidden)
		}
	}
}
