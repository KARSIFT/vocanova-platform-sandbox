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

func TestVOC030P4UserSettingsMigrationCarriesDatabaseInvariants(t *testing.T) {
	sql, err := os.ReadFile("20260725130000_voc030_p4_user_settings.sql")
	if err != nil {
		t.Fatalf("read voc-030 p4 user_settings migration: %v", err)
	}
	text := string(sql)
	required := []string{
		"CREATE TABLE user_settings",
		"REFERENCES users(id)",
		"ON DELETE RESTRICT",
		"timezone text NOT NULL DEFAULT 'UTC'",
		"daily_review_target integer NOT NULL DEFAULT 20",
		"daily_review_target >= 5 AND daily_review_target <= 100",
		"review_interval_preset IN ('vocanova_default', 'wordup_like', 'custom')",
		"app_language ~ '^[A-Za-z]{2,8}$'",
		"char_length(timezone) > 0",
		"ON user_settings (user_id)",
	}
	for _, invariant := range required {
		if !strings.Contains(text, invariant) {
			t.Errorf("migration missing invariant %q", invariant)
		}
	}
	if strings.Contains(text, "ON DELETE CASCADE") {
		t.Errorf("user_settings migration contains forbidden ON DELETE CASCADE")
	}
}

func TestVOC030P4MissionTablesMigrationCarriesDatabaseInvariants(t *testing.T) {
	sql, err := os.ReadFile("20260725130001_voc030_p4_mission_tables.sql")
	if err != nil {
		t.Fatalf("read voc-030 p4 mission tables migration: %v", err)
	}
	text := string(sql)
	required := []string{
		"CREATE TABLE daily_mission_snapshots",
		"CREATE TABLE daily_activity_summaries",
		"REFERENCES users(id)",
		"ON DELETE RESTRICT",
		"local_date date NOT NULL",
		"review_target >= 5 AND review_target <= 100",
		"reviews_completed >= 0 AND reviews_completed <= review_target",
		"status IN ('open', 'completed', 'missed', 'protected')",
		"status <> 'completed' OR completed_at IS NOT NULL",
		"ON daily_mission_snapshots (user_id, local_date)",
		"ON daily_activity_summaries (user_id, local_date)",
	}
	for _, invariant := range required {
		if !strings.Contains(text, invariant) {
			t.Errorf("mission_tables migration missing invariant %q", invariant)
		}
	}
	if strings.Contains(text, "ON DELETE CASCADE") {
		t.Errorf("voc-030 mission_tables migration contains forbidden ON DELETE CASCADE")
	}
}

func TestVOC030P4GamificationTablesMigrationCarriesDatabaseInvariants(t *testing.T) {
	sql, err := os.ReadFile("20260725130002_voc030_p4_gamification_tables.sql")
	if err != nil {
		t.Fatalf("read voc-030 p4 gamification tables migration: %v", err)
	}
	text := string(sql)
	required := []string{
		"CREATE TABLE confidence_point_ledger",
		"CREATE TABLE streak_states",
		"CREATE TABLE grace_day_ledger",
		"amount <> 0",
		"reason IN ('word_added', 'review_correct', 'daily_mission_completed', 'sentence_submitted', 'ai_feedback_received', 'streak_bonus', 'admin_adjustment')",
		"source_type IN ('user_word', 'review_attempt', 'daily_mission', 'learner_sentence', 'ai_feedback_attempt', 'streak', 'admin')",
		"longest_streak_count >= current_streak_count",
		"status IN ('active', 'at_risk', 'broken')",
		"reason IN ('earned_by_streak', 'manual_grant', 'used_for_missed_day', 'expired', 'admin_adjustment')",
		"source_type IN ('daily_mission', 'streak', 'admin')",
		"ON confidence_point_ledger (user_id, idempotency_key)\n  WHERE idempotency_key IS NOT NULL",
		"ON grace_day_ledger (user_id, idempotency_key)\n  WHERE idempotency_key IS NOT NULL",
		"  user_id uuid NOT NULL UNIQUE",
	}
	for _, invariant := range required {
		if !strings.Contains(text, invariant) {
			t.Errorf("gamification_tables migration missing invariant %q", invariant)
		}
	}
	if strings.Contains(text, "ON DELETE CASCADE") {
		t.Errorf("voc-030 gamification_tables migration contains forbidden ON DELETE CASCADE (ledgers must be immutable)")
	}
}

// TestVOC031P5UserOnboardingProfilesMigrationCarriesDatabaseInvariants
// covers VOC-031-TEST-00. The migration is the T00 deliverable: it
// creates the user_onboarding_profiles table (DOC-05 §6) and
// backfills pre-existing users' onboarding_status to 'completed'
// (VOC-031-D03). Like every other migration in this repository, no
// existing A1–P4 table, column, or constraint is altered — the
// backfill is purely additive on users.onboarding_status.
func TestVOC031P5UserOnboardingProfilesMigrationCarriesDatabaseInvariants(t *testing.T) {
	sql, err := os.ReadFile("20260725140000_voc031_p5_user_onboarding_profiles.sql")
	if err != nil {
		t.Fatalf("read voc-031 p5 user_onboarding_profiles migration: %v", err)
	}
	text := string(sql)
	required := []string{
		"CREATE TABLE user_onboarding_profiles",
		"REFERENCES users(id)",
		"ON DELETE RESTRICT",
		"english_level IN ('a1', 'a2', 'b1', 'b2', 'unknown')",
		"native_language",
		"char_length(native_language) > 0",
		"learning_goal IN ('general', 'work', 'travel', 'study', 'conversation', 'exam')",
		"main_use_case IN ('daily_life', 'work', 'travel', 'study', 'social')",
		"daily_review_target >= 5 AND daily_review_target <= 100",
		"completed_at timestamptz",
		"user_id uuid NOT NULL UNIQUE",
		"ON user_onboarding_profiles (user_id)",
		"UPDATE users",
		"SET onboarding_status = 'completed'",
		"WHERE onboarding_status = 'not_started'",
		"created_at < NOW()",
	}
	for _, invariant := range required {
		if !strings.Contains(text, invariant) {
			t.Errorf("user_onboarding_profiles migration missing invariant %q", invariant)
		}
	}
	if strings.Contains(text, "ON DELETE CASCADE") {
		t.Errorf("user_onboarding_profiles migration contains forbidden ON DELETE CASCADE")
	}
}

// TestVOC031P5EmailChangeLinksMigrationCarriesDatabaseInvariants
// covers the migration invariants for VOC-031-T03. The migration
// creates the email_change_links table per VOC-031-D05: it mirrors
// magic_links' token-hash, 15-minute expiry, environment-scoping,
// and single-use discipline exactly, with three deliberate
// differences (user_id NOT NULL, new_email instead of email, no
// email-uniqueness constraint) that are themselves invariant
// requirements.
func TestVOC031P5EmailChangeLinksMigrationCarriesDatabaseInvariants(t *testing.T) {
	sql, err := os.ReadFile("20260725140001_voc031_p5_email_change_links.sql")
	if err != nil {
		t.Fatalf("read voc-031 p5 email_change_links migration: %v", err)
	}
	text := string(sql)
	required := []string{
		"CREATE TABLE email_change_links",
		"user_id uuid NOT NULL REFERENCES users(id)",
		"ON DELETE RESTRICT",
		"new_email text NOT NULL",
		"new_email <> ''",
		"token_hash bytea NOT NULL UNIQUE",
		"octet_length(token_hash) = 32",
		"environment text NOT NULL",
		"environment <> ''",
		"expires_at > created_at",
		"expires_at <= created_at + interval '15 minutes'",
		"consumed_at IS NULL OR consumed_at >= created_at",
		"revoked_at IS NULL OR revoked_at >= created_at",
		"consumed_at IS NULL OR revoked_at IS NULL",
		"ON email_change_links (expires_at)",
		"WHERE consumed_at IS NULL AND revoked_at IS NULL",
		"ON email_change_links (user_id)",
	}
	for _, invariant := range required {
		if !strings.Contains(text, invariant) {
			t.Errorf("email_change_links migration missing invariant %q", invariant)
		}
	}
	for _, forbidden := range []string{
		"session_token", "magic_link_token", "access_token", "refresh_token",
		"oauth_state_token", "email_change_link_token",
	} {
		if strings.Contains(text, forbidden) {
			t.Errorf("email_change_links migration contains forbidden raw bearer column %q", forbidden)
		}
	}
	if strings.Contains(text, "ON DELETE CASCADE") {
		t.Errorf("email_change_links migration contains forbidden ON DELETE CASCADE")
	}
}

// TestVOC031P5AccountDeletionRequestsMigrationCarriesDatabaseInvariants
// covers the migration invariants for VOC-031-T04. The
// migration creates the account_deletion_requests table per
// DOC-05 §16 / DOC-06 §14 / VOC-031-D07: one row per user
// (user_id UNIQUE), a 30-day-default purge_after the
// service-layer code reads, a (status, purge_after) partial
// index that makes the sweep's "find due rows" an index
// scan, and a check constraint pair that ensures
// status='completed' is paired with completed_at IS NOT NULL
// (and the converse). No existing A1–P4 table, column, or
// constraint is altered; no ON DELETE CASCADE is introduced
// (the per-table disposition runs in code, not as a database
// cascade).
func TestVOC031P5AccountDeletionRequestsMigrationCarriesDatabaseInvariants(t *testing.T) {
	sql, err := os.ReadFile("20260725140002_voc031_p5_account_deletion_requests.sql")
	if err != nil {
		t.Fatalf("read voc-031 p5 account_deletion_requests migration: %v", err)
	}
	text := string(sql)
	required := []string{
		"CREATE TABLE account_deletion_requests",
		"user_id uuid NOT NULL UNIQUE REFERENCES users(id)",
		"ON DELETE RESTRICT",
		"status text NOT NULL DEFAULT 'deactivated'",
		"status IN ('deactivated', 'anonymizing', 'completed')",
		"requested_at timestamptz NOT NULL",
		"purge_after timestamptz NOT NULL",
		"completed_at timestamptz",
		"idempotency_key text NOT NULL",
		"idempotency_key <> ''",
		"purge_after > requested_at",
		"purge_after <= requested_at + interval '365 days'",
		"status = 'completed' OR completed_at IS NULL",
		"status <> 'completed' OR completed_at IS NOT NULL",
		"ON account_deletion_requests (status, purge_after)",
		"WHERE status = 'deactivated'",
		"ON account_deletion_requests (user_id)",
	}
	for _, invariant := range required {
		if !strings.Contains(text, invariant) {
			t.Errorf("account_deletion_requests migration missing invariant %q", invariant)
		}
	}
	for _, forbidden := range []string{
		"session_token", "magic_link_token", "access_token", "refresh_token",
		"oauth_state_token", "email_change_link_token", "account_deletion_token",
	} {
		if strings.Contains(text, forbidden) {
			t.Errorf("account_deletion_requests migration contains forbidden raw bearer column %q", forbidden)
		}
	}
	if strings.Contains(text, "ON DELETE CASCADE") {
		// The header comment is allowed to mention the
		// phrase "ON DELETE CASCADE" (it explicitly
		// states the migration does NOT introduce it).
		// Detect the actual SQL pattern: a space (or
		// start-of-line) before "ON DELETE CASCADE" is
		// what we forbid.
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "--") {
				continue
			}
			if strings.Contains(line, "ON DELETE CASCADE") {
				t.Errorf("account_deletion_requests migration contains forbidden ON DELETE CASCADE at line %d: %q", i+1, line)
			}
		}
	}
}
