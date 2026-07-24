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
