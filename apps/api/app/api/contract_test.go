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
	for _, expected := range []string{"GetCurrentUser", "/api/v1/me", "displayName", "onboardingStatus"} {
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

// TestContractContainsOnboardingEndpoints pins the OpenAPI shape
// for VOC-031-T01: the /api/v1/onboarding GET/POST routes, the
// DTOs, and the enum constraints.
func TestContractContainsOnboardingEndpoints(t *testing.T) {
	document, err := json.Marshal(NewContractAPI().OpenAPI())
	if err != nil {
		t.Fatalf("marshal OpenAPI: %v", err)
	}
	contract := string(document)
	for _, expected := range []string{
		"GetOnboarding",
		"CompleteOnboarding",
		"/api/v1/onboarding",
		"OnboardingProfileDTO",
		"englishLevel",
		"nativeLanguage",
		"learningGoal",
		"mainUseCase",
		"dailyReviewTarget",
		"completedAt",
		// Enum constraints the API uses to validate the inbound
		// payload at the Huma boundary.
		"a1", "a2", "b1", "b2", "unknown",
		"general", "work", "travel", "study", "conversation", "exam",
		"daily_life", "social",
	} {
		if !strings.Contains(contract, expected) {
			t.Errorf("OpenAPI missing onboarding contract element %q", expected)
		}
	}
	for _, forbidden := range []string{
		// Onboarding answers are learner free-text-free; the
		// contract must not expose any internal user_id / FK
		// to the requester.
		"user_id",
		"token_hash",
		"deleted_at",
	} {
		if strings.Contains(contract, forbidden) {
			t.Errorf("OpenAPI exposed internal onboarding field %q", forbidden)
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
