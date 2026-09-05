//go:build integration

package accounts

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// TestPostgreSQLRepositoryExportAndAnonymization exercises the real SQL
// projection and purge ordering against a separately migrated disposable
// database. It is opt-in because migrations are intentionally not run by the
// ordinary unit-test suite.
func TestPostgreSQLRepositoryExportAndAnonymization(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	userA, userB := uuid.New(), uuid.New()
	sentenceA, sentenceB := uuid.New(), uuid.New()
	attemptA := uuid.New()
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	for _, user := range []struct {
		id    uuid.UUID
		email string
	}{{userA, userA.String() + "@example.test"}, {userB, userB.String() + "@example.test"}} {
		sessionHash := sha256.Sum256([]byte(user.id.String() + "session"))
		magicHash := sha256.Sum256([]byte(user.id.String() + "magic"))
		emailChangeHash := sha256.Sum256([]byte(user.id.String() + "email-change"))
		exec(`INSERT INTO users (id, email, status, created_at, updated_at) VALUES ($1, $2, 'deleted', $3, $3)`, user.id, user.email, now)
		exec(`INSERT INTO external_identities (id, user_id, provider, provider_subject, created_at, updated_at) VALUES ($1, $2, 'email', $3, $4, $4)`, uuid.New(), user.id, user.id.String(), now)
		exec(`INSERT INTO sessions (id, user_id, token_hash, created_at, expires_at) VALUES ($1, $2, $3, $4, $5)`, uuid.New(), user.id, sessionHash[:], now, now.Add(time.Hour))
		exec(`INSERT INTO magic_links (id, user_id, email, token_hash, environment, created_at, expires_at) VALUES ($1, $2, $3, $4, 'test', $5, $6)`, uuid.New(), user.id, user.email, magicHash[:], now, now.Add(time.Minute))
		exec(`INSERT INTO email_change_links (id, user_id, new_email, token_hash, environment, created_at, expires_at) VALUES ($1, $2, $3, $4, 'test', $5, $6)`, uuid.New(), user.id, "new-"+user.email, emailChangeHash[:], now, now.Add(time.Minute))
		exec(`INSERT INTO idempotency_keys (id, user_id, operation, key, fingerprint, created_at) VALUES ($1, $2, 'test', $3, 'fingerprint', $4)`, uuid.New(), user.id, user.id.String(), now)
	}
	exec(`INSERT INTO user_settings (id, user_id, timezone, daily_review_target, review_interval_preset, notifications_enabled, marketing_emails_enabled, app_language, created_at, updated_at) VALUES ($1, $2, 'Asia/Tehran', 25, 'wordup_like', false, true, 'en', $3, $3)`, uuid.New(), userA, now)
	for _, sentence := range []struct {
		id     uuid.UUID
		userID uuid.UUID
		text   string
	}{{sentenceA, userA, "A private learner sentence."}, {sentenceB, userB, "B private learner sentence."}} {
		exec(`INSERT INTO learner_sentences (id, user_id, sentence_text, normalized_sentence_text, source, status, submitted_at, created_at, updated_at) VALUES ($1, $2, $3, lower($3), 'free_practice', 'feedback_ready', $4, $4, $4)`, sentence.id, sentence.userID, sentence.text, now)
	}
	exec(`INSERT INTO ai_feedback_attempts (id, learner_sentence_id, status, provider, model, prompt_version, request_hash, feedback_json, completed_at, created_at, updated_at) VALUES ($1, $2, 'succeeded', 'test', 'test', 'v1', $3, '{"status":"correct"}', $4, $4, $4)`, attemptA, sentenceA, attemptA.String(), now)
	exec(`INSERT INTO ai_feedback_quality_review_reports (id, ai_feedback_attempt_id, user_id, reason, classification, state, created_at, updated_at) VALUES ($1, $2, $3, 'already_correct', 'unnecessary_correction', 'open', $4, $4)`, uuid.New(), attemptA, userA, now)
	exec(`INSERT INTO account_deletion_requests (id, user_id, status, requested_at, purge_after, idempotency_key, created_at, updated_at) VALUES ($1, $2, 'deactivated', $3, $4, 'delete-key-a', $3, $3)`, uuid.New(), userA, now, now.Add(time.Hour))

	repo := NewPostgreSQLRepository(db)
	payload, err := repo.ExportPersonalData(ctx, userA)
	if err != nil {
		t.Fatal(err)
	}
	var export map[string]any
	if err := json.Unmarshal(payload, &export); err != nil {
		t.Fatal(err)
	}
	settings := export["settings"].(map[string]any)
	if settings["timezone"] != "Asia/Tehran" || settings["dailyReviewTarget"] != float64(25) {
		t.Fatalf("unexpected real settings projection: %#v", settings)
	}
	otherPayload := repoMustExport(t, repo, ctx, userB)
	if string(payload) == "" || string(payload) == string(otherPayload) {
		t.Fatal("exports must be non-empty and requester-scoped")
	}
	var otherExport map[string]any
	if err := json.Unmarshal(otherPayload, &otherExport); err != nil {
		t.Fatal(err)
	}
	otherSettings := otherExport["settings"].(map[string]any)
	if otherSettings["timezone"] != "UTC" || otherSettings["dailyReviewTarget"] != float64(20) {
		t.Fatalf("unexpected schema-default settings projection: %#v", otherSettings)
	}

	counters, err := repo.AnonymizeUserData(ctx, userA)
	if err != nil {
		t.Fatal(err)
	}
	if counters.LearnerSentences != 1 || counters.AIFeedbackAttempts != 1 || counters.AIQualityReviewReports != 1 || counters.ExternalIdentities != 1 {
		t.Fatalf("unexpected purge counters: %#v", counters)
	}
	var aRemaining, bRemaining int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM learner_sentences WHERE user_id = $1`, userA).Scan(&aRemaining); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM learner_sentences WHERE user_id = $1`, userB).Scan(&bRemaining); err != nil {
		t.Fatal(err)
	}
	if aRemaining != 0 || bRemaining != 1 {
		t.Fatalf("purge isolation failed: A=%d B=%d", aRemaining, bRemaining)
	}
	for _, table := range []string{"sessions", "magic_links", "email_change_links"} {
		var aTokens, bTokens int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM `+table+` WHERE user_id = $1`, userA).Scan(&aTokens); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM `+table+` WHERE user_id = $1`, userB).Scan(&bTokens); err != nil {
			t.Fatal(err)
		}
		if aTokens != 0 || bTokens != 1 {
			t.Fatalf("%s purge isolation failed: A=%d B=%d", table, aTokens, bTokens)
		}
	}
}

// TestPostgreSQLRepositoryReclaimsOnlyStaleDeletionClaims exercises the
// production query and atomic claim predicate against the migrated schema.
func TestPostgreSQLRepositoryReclaimsOnlyStaleDeletionClaims(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	staleBefore := now.Add(-DefaultAccountDeletionClaimTimeout)
	staleUser, freshUser, dueUser := uuid.New(), uuid.New(), uuid.New()
	staleRequest, freshRequest, dueRequest := uuid.New(), uuid.New(), uuid.New()
	for _, userID := range []uuid.UUID{staleUser, freshUser, dueUser} {
		if _, err := db.ExecContext(ctx, `INSERT INTO users (id, email, status, created_at, updated_at) VALUES ($1, $2, 'deleted', $3, $3)`, userID, userID.String()+"@example.test", now); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		for _, id := range []uuid.UUID{staleRequest, freshRequest, dueRequest} {
			_, _ = db.ExecContext(context.Background(), `DELETE FROM account_deletion_requests WHERE id = $1`, id)
		}
		for _, id := range []uuid.UUID{staleUser, freshUser, dueUser} {
			_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, id)
		}
	})
	for _, row := range []struct {
		id, userID uuid.UUID
		status     string
		updatedAt  time.Time
	}{
		{staleRequest, staleUser, "anonymizing", staleBefore.Add(-time.Minute)},
		{freshRequest, freshUser, "anonymizing", now},
		{dueRequest, dueUser, "deactivated", now},
	} {
		if _, err := db.ExecContext(ctx, `INSERT INTO account_deletion_requests (id, user_id, status, requested_at, purge_after, idempotency_key, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $4, $7)`, row.id, row.userID, row.status, now.Add(-31*24*time.Hour), now.Add(-time.Hour), row.id.String(), row.updatedAt); err != nil {
			t.Fatal(err)
		}
	}

	repo := NewPostgreSQLRepository(db)
	candidates, err := repo.ListDeactivatedRequestsDueForPurge(ctx, now, staleBefore, 10)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[uuid.UUID]bool{}
	for _, candidate := range candidates {
		seen[candidate.ID] = true
	}
	if !seen[staleRequest] || !seen[dueRequest] || seen[freshRequest] {
		t.Fatalf("unexpected recovery candidates: %#v", seen)
	}

	// Two recovery workers race on the same stale lease. PostgreSQL's guarded
	// UPDATE must grant exactly one ownership claim.
	start := make(chan struct{})
	results := make(chan bool, 2)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			claimed, err := repo.ClaimAccountDeletionRequestForAnonymization(ctx, staleRequest, now, staleBefore)
			results <- claimed
			errs <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	close(errs)
	winners := 0
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	for claimed := range results {
		if claimed {
			winners++
		}
	}
	if winners != 1 {
		t.Fatalf("stale claim has %d winners, want exactly one", winners)
	}

	// The original owner resumes after reclamation. Its obsolete claim token
	// must fail before the purge mutates learner data.
	_, completed, err := repo.FinalizeAccountDeletionClaim(ctx, staleRequest, staleUser, staleBefore.Add(-time.Minute), now)
	if err != nil || completed {
		t.Fatalf("stale owner finalized reclaimed request: completed=%t err=%v", completed, err)
	}
	var email string
	if err := db.QueryRowContext(ctx, `SELECT email FROM users WHERE id = $1`, staleUser).Scan(&email); err != nil || email == "" {
		t.Fatalf("stale owner mutated user data: email=%q err=%v", email, err)
	}
	_, completed, err = repo.FinalizeAccountDeletionClaim(ctx, staleRequest, staleUser, now, now)
	if err != nil || !completed {
		t.Fatalf("current owner did not finalize claim: completed=%t err=%v", completed, err)
	}
}

// TestUpdateUserEmailMovesEmailProviderIdentityPostgreSQL ensures a confirmed
// email change releases the old mailbox at the provider-identity boundary.
// A later magic-link sign-in for that old mailbox can therefore belong to a
// newly created account instead of being rejected as another user's identity.
func TestUpdateUserEmailMovesEmailProviderIdentityPostgreSQL(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}

	userID := uuid.New()
	oldEmail := userID.String() + "-old@example.test"
	newEmail := userID.String() + "-new@example.test"
	now := time.Now().UTC()
	if _, err := db.ExecContext(ctx, `INSERT INTO users (id, email, status, created_at, updated_at) VALUES ($1, $2, 'active', $3, $3)`, userID, oldEmail, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO external_identities (id, user_id, provider, provider_subject, provider_email, provider_email_verified, created_at, updated_at) VALUES ($1, $2, 'email', $3, $3, true, $4, $4)`, uuid.New(), userID, oldEmail, now); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM external_identities WHERE user_id = $1`, userID)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})

	repo := NewPostgreSQLRepository(db)
	if err := repo.UpdateUserEmail(ctx, userID, newEmail, now); err != nil {
		t.Fatal(err)
	}
	var subject, providerEmail string
	var verified bool
	if err := db.QueryRowContext(ctx, `SELECT provider_subject, provider_email, provider_email_verified FROM external_identities WHERE user_id = $1 AND provider = 'email' AND deleted_at IS NULL`, userID).Scan(&subject, &providerEmail, &verified); err != nil {
		t.Fatal(err)
	}
	if subject != newEmail || providerEmail != newEmail || !verified {
		t.Fatalf("email identity not moved: subject=%q email=%q verified=%t", subject, providerEmail, verified)
	}
	var oldCount int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM external_identities WHERE provider = 'email' AND provider_subject = $1 AND deleted_at IS NULL`, oldEmail).Scan(&oldCount); err != nil {
		t.Fatal(err)
	}
	if oldCount != 0 {
		t.Fatalf("old email identity remains active: %d", oldCount)
	}
}

func repoMustExport(t *testing.T, repo *PostgreSQLRepository, ctx context.Context, userID uuid.UUID) []byte {
	t.Helper()
	payload, err := repo.ExportPersonalData(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}
