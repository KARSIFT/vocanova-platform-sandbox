package migrations_test

import (
	"os"
	"strings"
	"testing"
)

func TestIdentityMigrationCarriesDatabaseInvariants(t *testing.T) {
	sql, err := os.ReadFile("20260724210000_identity_foundation.sql")
	if err != nil {
		t.Fatalf("read identity migration: %v", err)
	}
	text := string(sql)
	required := []string{
		"CREATE TABLE users",
		"CREATE TABLE external_identities",
		"CREATE TABLE sessions",
		"CREATE TABLE magic_links",
		"lower(email)",
		"octet_length(token_hash) = 32",
		"interval '30 days'",
		"interval '15 minutes'",
		"consumed_at IS NULL OR revoked_at IS NULL",
	}
	for _, invariant := range required {
		if !strings.Contains(text, invariant) {
			t.Errorf("migration missing invariant %q", invariant)
		}
	}
	for _, forbidden := range []string{"session_token", "magic_link_token", "access_token", "refresh_token"} {
		if strings.Contains(text, forbidden) {
			t.Errorf("migration contains forbidden raw bearer column %q", forbidden)
		}
	}
}

func TestOAuthStateMigrationCarriesDatabaseInvariants(t *testing.T) {
	sql, err := os.ReadFile("20260724210001_oauth_state.sql")
	if err != nil {
		t.Fatalf("read oauth state migration: %v", err)
	}
	text := string(sql)
	required := []string{
		"CREATE TABLE oauth_states",
		"octet_length(token_hash) = 32",
		"interval '10 minutes'",
		"consumed_at IS NULL",
		"app_return_url",
	}
	for _, invariant := range required {
		if !strings.Contains(text, invariant) {
			t.Errorf("migration missing invariant %q", invariant)
		}
	}
	for _, forbidden := range []string{"session_token", "magic_link_token", "access_token", "refresh_token", "oauth_state_token"} {
		if strings.Contains(text, forbidden) {
			t.Errorf("migration contains forbidden raw bearer column %q", forbidden)
		}
	}
}
