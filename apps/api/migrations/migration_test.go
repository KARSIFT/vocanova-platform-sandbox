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

func TestVOC026P1ContentMigrationCarriesDatabaseInvariants(t *testing.T) {
	sql, err := os.ReadFile("20260725100000_voc026_p1_content_tables.sql")
	if err != nil {
		t.Fatalf("read voc-026 p1 content migration: %v", err)
	}
	text := string(sql)
	required := []string{
		"CREATE TABLE canonical_words",
		"CREATE TABLE word_meanings",
		"CREATE TABLE word_examples",
		"CREATE TABLE usage_notes",
		"CREATE TABLE journey_situations",
		"CREATE TABLE journey_words",
		"CREATE TABLE user_words",
		"canonical_words (language_code, normalized_text)",
		"word_meanings (word_id, meaning_order)",
		"word_examples (meaning_id, example_order)",
		"usage_notes (meaning_id, note_order)",
		"journey_situations (slug)",
		"journey_words (journey_situation_id, meaning_id)",
		"user_words (user_id, meaning_id) WHERE deleted_at IS NULL",
		"review_step >= 0 AND review_step <= 7",
		"relevance_score >= 1 AND relevance_score <= 100",
		"correct_review_count <= total_review_count",
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

func TestVOC027P2ReviewAttemptsMigrationCarriesDatabaseInvariants(t *testing.T) {
	sql, err := os.ReadFile("20260725110000_voc027_p2_review_attempts.sql")
	if err != nil {
		t.Fatalf("read voc-027 p2 review attempts migration: %v", err)
	}
	text := string(sql)
	required := []string{
		"CREATE TABLE review_attempts",
		"REFERENCES users(id)",
		"REFERENCES user_words(id)",
		"REFERENCES word_meanings(id)",
		"ON DELETE RESTRICT",
		"prompt_type IN ('multiple_choice', 'self_check')",
		"result IN ('correct', 'incorrect', 'skipped')",
		"rating IS NULL OR rating IN ('again', 'hard', 'good', 'easy')",
		"review_step_before >= 0 AND review_step_before <= 7",
		"review_step_after >= 0 AND review_step_after <= 7",
		"response_time_ms >= 0",
		"ON review_attempts (user_id, client_attempt_id)\n  WHERE client_attempt_id IS NOT NULL",
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
	if strings.Contains(text, "ON DELETE CASCADE") {
		t.Errorf("review_attempts migration contains forbidden ON DELETE CASCADE")
	}
}

func TestVOC028P3LearnerSentencesMigrationCarriesDatabaseInvariants(t *testing.T) {
	sql, err := os.ReadFile("20260725120000_voc028_p3_learner_sentences.sql")
	if err != nil {
		t.Fatalf("read voc-028 p3 learner_sentences migration: %v", err)
	}
	text := string(sql)
	required := []string{
		"CREATE TABLE learner_sentences",
		"REFERENCES users(id)",
		"REFERENCES word_meanings(id)",
		"REFERENCES user_words(id)",
		"ON DELETE RESTRICT",
		"source IN ('word_detail', 'review', 'daily_mission', 'free_practice')",
		"status IN ('submitted', 'feedback_ready', 'feedback_failed', 'archived')",
		"char_length(sentence_text) <= 1000",
		"deleted_at",
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
	if strings.Contains(text, "ON DELETE CASCADE") {
		t.Errorf("learner_sentences migration contains forbidden ON DELETE CASCADE")
	}
}

func TestVOC028P3AIFeedbackAttemptsMigrationCarriesDatabaseInvariants(t *testing.T) {
	sql, err := os.ReadFile("20260725120001_voc028_p3_ai_feedback_attempts.sql")
	if err != nil {
		t.Fatalf("read voc-028 p3 ai_feedback_attempts migration: %v", err)
	}
	text := string(sql)
	required := []string{
		"CREATE TABLE ai_feedback_attempts",
		"REFERENCES learner_sentences(id)",
		"ON DELETE RESTRICT",
		"status IN ('pending', 'succeeded', 'failed', 'cancelled')",
		"provider text NOT NULL",
		"model text NOT NULL",
		"prompt_version text NOT NULL",
		"request_hash text NOT NULL",
		"feedback_json jsonb",
		"status <> 'succeeded' OR completed_at IS NOT NULL",
		"status <> 'failed' OR error_code IS NOT NULL",
		"ON ai_feedback_attempts (request_hash)",
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
	if strings.Contains(text, "ON DELETE CASCADE") {
		t.Errorf("ai_feedback_attempts migration contains forbidden ON DELETE CASCADE")
	}
}

func TestVOC026P1IdempotencyMigrationCarriesDatabaseInvariants(t *testing.T) {
	sql, err := os.ReadFile("20260725100001_voc026_p1_idempotency_keys.sql")
	if err != nil {
		t.Fatalf("read voc-026 p1 idempotency migration: %v", err)
	}
	text := string(sql)
	required := []string{
		"CREATE TABLE idempotency_keys",
		"user_id",
		"operation",
		"fingerprint",
		"idempotency_keys (user_id, operation, key)",
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
