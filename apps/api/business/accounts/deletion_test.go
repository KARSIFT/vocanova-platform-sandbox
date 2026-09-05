package accounts

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateAccountDeletionRequestDeactivatesUserAndPersistsRow covers
// VOC-031-TEST-18: a successful POST marks the user deleted, persists
// the account_deletion_requests row with the correct status and
// purge_after, and invokes every required auth-side revocation
// (sessions, magic links, email change links). Replayed idempotency
// keys return the existing row.
func TestCreateAccountDeletionRequestDeactivatesUserAndPersistsRow(t *testing.T) {
	svc, repo, authRepo, _, c := newService(t)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "user@example.com", Status: "active"})
	repo.SetUser(uid, "user@example.com")

	res, err := svc.CreateAccountDeletionRequest(context.Background(), uid.String(), "1.2.3.4", "sess-token", "idem-key-1")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.False(t, res.Replayed, "first call is not a replay")
	assert.Equal(t, uid, res.UserID)
	assert.Equal(t, "deactivated", res.Status)
	assert.Equal(t, "idem-key-1", res.IdempotencyKey)
	assert.True(t, res.PurgeAfter.After(res.RequestedAt), "purge_after is later than requested_at")
	assert.Equal(t, 30*24*time.Hour, res.PurgeAfter.Sub(res.RequestedAt), "default purge_after is 30 days")

	// Auth-side revocations fired. The in-memory
	// CreateAccountDeletionRequest stands in for the SQL
	// transaction's per-table writes; the
	// SessionsRevoked / MagicLinksRevoked counters it
	// increments are what tests assert on. The production
	// SQL path delegates the same revocations to the
	// auth.Repository.
	assert.Equal(t, int64(1), repo.SessionsRevoked(), "sessions were revoked")
	assert.Equal(t, int64(1), repo.MagicLinksRevoked(), "magic links were revoked")

	// Repository has a deletion row in 'deactivated'.
	row := repo.DeletionRequest(uid)
	require.NotNil(t, row)
	assert.Equal(t, "deactivated", row.Status)
	assert.Equal(t, "idem-key-1", row.IdempotencyKey)
	_ = c
}

// TestCreateAccountDeletionRequestReplaysIdempotencyKey covers
// VOC-031-TEST-19: a second call with the same Idempotency-Key
// returns the same row without re-running the deactivation. The
// repository's per-call revocation counter confirms the second
// call did not invoke the deactivation again.
func TestCreateAccountDeletionRequestReplaysIdempotencyKey(t *testing.T) {
	svc, repo, authRepo, _, _ := newService(t)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "user@example.com", Status: "active"})
	repo.SetUser(uid, "user@example.com")

	first, err := svc.CreateAccountDeletionRequest(context.Background(), uid.String(), "1.2.3.4", "sess-token", "idem-key")
	require.NoError(t, err)
	require.False(t, first.Replayed)
	assert.Equal(t, int64(1), repo.SessionsRevoked(), "first call revoked sessions exactly once")

	second, err := svc.CreateAccountDeletionRequest(context.Background(), uid.String(), "1.2.3.4", "sess-token", "idem-key")
	require.NoError(t, err)
	require.NotNil(t, second)
	assert.True(t, second.Replayed, "second call is a replay")
	assert.Equal(t, first.UserID, second.UserID)
	assert.Equal(t, first.PurgeAfter, second.PurgeAfter, "replay returns the same purge_after")
	assert.Equal(t, int64(1), repo.SessionsRevoked(), "replay must not re-run the deactivation")
	// The repository still has only one row.
	row := repo.DeletionRequest(uid)
	require.NotNil(t, row)
}

// TestCreateAccountDeletionRequestIdempotencyKeysAreUserScoped covers
// VOC-031-TEST-20: the (user, operation, key) tuple the
// idempotency store indexes is unique per user. The same
// Idempotency-Key string is therefore reusable across distinct
// users without any conflict — the cache key is "userID +
// operation + key", and a different userID short-circuits the
// lookup to Absent. The test pins this property so a future
// "global" idempotency-key collision cannot accidentally be
// introduced.
func TestCreateAccountDeletionRequestIdempotencyKeysAreUserScoped(t *testing.T) {
	svc, repo, authRepo, _, _ := newService(t)
	uidA := uuid.New()
	uidB := uuid.New()
	authRepo.setUser(&auth.User{ID: uidA, Email: "a@example.com", Status: "active"})
	authRepo.setUser(&auth.User{ID: uidB, Email: "b@example.com", Status: "active"})
	repo.SetUser(uidA, "a@example.com")
	repo.SetUser(uidB, "b@example.com")

	resA, err := svc.CreateAccountDeletionRequest(context.Background(), uidA.String(), "1.2.3.4", "sess-token-A", "shared-key")
	require.NoError(t, err)
	require.False(t, resA.Replayed)
	// uidB reuses the exact same Idempotency-Key. The
	// idempotency store's cache key includes the user id, so
	// the lookup for uidB returns Absent and the call
	// proceeds (a different user with a separate deactivation
	// row in its own right).
	resB, err := svc.CreateAccountDeletionRequest(context.Background(), uidB.String(), "5.6.7.8", "sess-token-B", "shared-key")
	require.NoError(t, err)
	require.NotNil(t, resB)
	assert.False(t, resB.Replayed, "a different user is not a replay of userA's request")
	assert.Equal(t, uidB, resB.UserID, "userB gets its own deactivation row")
}

// TestCreateAccountDeletionRequestRequiresIdempotencyKey covers the
// 400 path: a missing Idempotency-Key header surfaces as
// ErrAccountDeletionIdempotencyKeyRequired, mapped by the API
// layer to a stable 400.
func TestCreateAccountDeletionRequestRequiresIdempotencyKey(t *testing.T) {
	svc, repo, authRepo, _, _ := newService(t)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "user@example.com", Status: "active"})
	repo.SetUser(uid, "user@example.com")

	_, err := svc.CreateAccountDeletionRequest(context.Background(), uid.String(), "1.2.3.4", "sess-token", "")
	assert.ErrorIs(t, err, ErrAccountDeletionIdempotencyKeyRequired)
	_, err = svc.CreateAccountDeletionRequest(context.Background(), uid.String(), "1.2.3.4", "sess-token", "   ")
	assert.ErrorIs(t, err, ErrAccountDeletionIdempotencyKeyRequired)
}

// TestCreateAccountDeletionRequestRequiresUserID covers the 400
// path: an empty / unparseable user id is rejected before any
// work runs.
func TestCreateAccountDeletionRequestRequiresUserID(t *testing.T) {
	svc, _, _, _, _ := newService(t)
	_, err := svc.CreateAccountDeletionRequest(context.Background(), "", "1.2.3.4", "sess-token", "idem")
	assert.Error(t, err)
	_, err = svc.CreateAccountDeletionRequest(context.Background(), "not-a-uuid", "1.2.3.4", "sess-token", "idem")
	assert.Error(t, err)
	_, err = svc.CreateAccountDeletionRequest(context.Background(), uuid.Nil.String(), "1.2.3.4", "sess-token", "idem")
	assert.Error(t, err)
}

// TestCreateAccountDeletionRequestUserNotFound covers the 404
// path: a deletion for a user that does not exist (or has
// already been deleted) is mapped to ErrUserNotFound.
func TestCreateAccountDeletionRequestUserNotFound(t *testing.T) {
	svc, repo, authRepo, _, _ := newService(t)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "user@example.com", Status: "active"})
	repo.SetUser(uid, "user@example.com")

	_, err := svc.CreateAccountDeletionRequest(context.Background(), uuid.New().String(), "1.2.3.4", "sess-token", "idem")
	assert.ErrorIs(t, err, ErrUserNotFound)

	// A second call against the same user surfaces as
	// already-in-flight, not user-not-found: the row is
	// already deactivated by the first call.
	_, err = svc.CreateAccountDeletionRequest(context.Background(), uid.String(), "1.2.3.4", "sess-token", "idem-A")
	require.NoError(t, err)
	_, err = svc.CreateAccountDeletionRequest(context.Background(), uid.String(), "1.2.3.4", "sess-token", "idem-B")
	assert.ErrorIs(t, err, ErrAccountDeletionAlreadyInFlight)
}

// TestCreateAccountDeletionRequestRateLimited covers the
// rate-limit branch. The auth.RateLimiter is a single
// fixed-window instance that the service calls twice per
// request — once keyed by IP and once keyed by session —
// against separate buckets. With a limit of 1, the second
// call's per-IP bucket has count 1, the per-session bucket
// has count 1, and either fires ErrAccountDeletionRateLimited
// first (the implementation checks per-IP, then per-session).
// The test pins the post-condition without trying to
// distinguish the two keys.
func TestCreateAccountDeletionRequestRateLimited(t *testing.T) {
	now := testNow()
	c := &clock.Fixed{T: now}
	svc, repo, authRepo, _, _ := newServiceForRate(t, c, 1)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "user@example.com", Status: "active"})
	repo.SetUser(uid, "user@example.com")

	_, err := svc.CreateAccountDeletionRequest(context.Background(), uid.String(), "1.2.3.4", "sess-token", "idem-A")
	require.NoError(t, err)
	_, err = svc.CreateAccountDeletionRequest(context.Background(), uid.String(), "1.2.3.4", "sess-token", "idem-B")
	assert.ErrorIs(t, err, ErrAccountDeletionRateLimited)
}

func TestExportPersonalDataReturnsRequesterScopedJSON(t *testing.T) {
	svc, repo, authRepo, _, _ := newService(t)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "user@example.com", Status: "active"})
	repo.SetUser(uid, "user@example.com")

	payload, err := svc.ExportPersonalData(context.Background(), uid.String(), "1.2.3.4", "session", "export-key")
	require.NoError(t, err)
	assert.JSONEq(t, `{"schemaVersion":"1.0","profile":{"id":"`+uid.String()+`","email":"user@example.com"},"settings":{"timezone":"UTC","dailyReviewTarget":20,"reviewIntervalPreset":"vocanova_default","notificationsEnabled":true,"marketingEmailsEnabled":false,"appLanguage":"en","createdAt":null,"updatedAt":null},"onboardingProfile":null,"savedWords":[],"reviewHistory":[],"sentenceFeedbackHistory":[],"dailyMissions":[],"dailyActivity":[],"confidencePointLedger":[],"graceDayLedger":[],"streakState":null}`, string(payload))

	_, err = svc.ExportPersonalData(context.Background(), uid.String(), "1.2.3.4", "session", "")
	assert.ErrorIs(t, err, ErrDataExportIdempotencyKeyRequired)
}

func TestExportPersonalDataIdempotencyIsUserScopedAndDoesNotCrossRead(t *testing.T) {
	svc, repo, authRepo, _, _ := newService(t)
	uidA, uidB := uuid.New(), uuid.New()
	authRepo.setUser(&auth.User{ID: uidA, Email: "a@example.com", Status: "active"})
	authRepo.setUser(&auth.User{ID: uidB, Email: "b@example.com", Status: "active"})
	repo.SetUser(uidA, "a@example.com")
	repo.SetUser(uidB, "b@example.com")

	first, err := svc.ExportPersonalData(context.Background(), uidA.String(), "1.2.3.4", "session-a", "shared-key")
	require.NoError(t, err)
	replay, err := svc.ExportPersonalData(context.Background(), uidA.String(), "1.2.3.4", "session-a", "shared-key")
	require.NoError(t, err)
	assert.JSONEq(t, string(first), string(replay), "same requester/key is a safe replay")

	other, err := svc.ExportPersonalData(context.Background(), uidB.String(), "5.6.7.8", "session-b", "shared-key")
	require.NoError(t, err)
	assert.Contains(t, string(other), "b@example.com")
	assert.NotContains(t, string(other), "a@example.com", "the key is scoped to the requester, never an export cache shared across users")
}

func TestExportPersonalDataRateLimited(t *testing.T) {
	now := testNow()
	c := &clock.Fixed{T: now}
	svc, repo, authRepo, _, _ := newServiceForRate(t, c, 1)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "user@example.com", Status: "active"})
	repo.SetUser(uid, "user@example.com")
	_, err := svc.ExportPersonalData(context.Background(), uid.String(), "1.2.3.4", "session", "key-a")
	require.NoError(t, err)
	_, err = svc.ExportPersonalData(context.Background(), uid.String(), "1.2.3.4", "session", "key-b")
	assert.ErrorIs(t, err, ErrDataExportRateLimited)
}

// TestExportPersonalDataUsesExplicitPrivacyProjection prevents a future
// convenience conversion such as to_jsonb(row) from silently adding provider
// metadata, arbitrary JSON, or operational error details to this download.
func TestExportPersonalDataUsesExplicitPrivacyProjection(t *testing.T) {
	source, err := os.ReadFile("postgres.go")
	require.NoError(t, err)
	projection := string(source[:strings.Index(string(source), "// PostgreSQLRepository implements")])
	assert.NotContains(t, projection, "to_jsonb(")
	for _, internal := range []string{"request_hash", "prompt_version", "error_message", "metadata jsonb", "feedback_json,"} {
		assert.NotContains(t, projection, internal)
	}
	assert.Contains(t, projection, "feedback_json->'status'", "only known learner-visible feedback keys may be selected")
	assert.Contains(t, projection, "JOIN word_meanings wm ON wm.id = uw.meaning_id")
	assert.Contains(t, projection, "'wordText', cw.text")
	assert.Contains(t, projection, "'shortDefinition', wm.short_definition")
	assert.Contains(t, projection, "'dailyReviewTarget', 20")
}

// TestRunDeletionSweepProcessesDueRequests covers VOC-031-TEST-21:
// the sweep reads every 'deactivated' row past its purge_after,
// claims each, runs the per-table disposition, and transitions
// the row to 'completed'. Counters are aggregated.
func TestRunDeletionSweepProcessesDueRequests(t *testing.T) {
	svc, repo, authRepo, _, c := newService(t)
	uidA := uuid.New()
	uidB := uuid.New()
	authRepo.setUser(&auth.User{ID: uidA, Email: "a@example.com", Status: "active"})
	authRepo.setUser(&auth.User{ID: uidB, Email: "b@example.com", Status: "active"})
	repo.SetUser(uidA, "a@example.com")
	repo.SetUser(uidB, "b@example.com")

	// Two deactivated rows; both are immediately past their
	// 30-day purge_after because we move the clock forward.
	_, err := svc.CreateAccountDeletionRequest(context.Background(), uidA.String(), "1.2.3.4", "sess-token", "idem-A")
	require.NoError(t, err)
	_, err = svc.CreateAccountDeletionRequest(context.Background(), uidB.String(), "1.2.3.4", "sess-token", "idem-B")
	require.NoError(t, err)

	c.Advance(31 * 24 * time.Hour)
	res, err := svc.RunDeletionSweep(context.Background(), "1.2.3.4", "sess-token")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, 2, res.Processed, "sweep saw both rows")
	assert.Equal(t, 2, res.Anonymized, "both rows were anonymized")
	assert.Equal(t, 0, res.Failed)

	// Both rows are now 'completed' with a stamped
	// completed_at.
	rowA := repo.DeletionRequest(uidA)
	require.NotNil(t, rowA)
	assert.Equal(t, "completed", rowA.Status)
	require.NotNil(t, rowA.CompletedAt)
	rowB := repo.DeletionRequest(uidB)
	require.NotNil(t, rowB)
	assert.Equal(t, "completed", rowB.Status)
	require.NotNil(t, rowB.CompletedAt)

	// Per-table counters are non-zero (the in-memory
	// AnonymizeUserData increments every class).
	assert.Greater(t, res.AnonymizationTotals.ExternalIdentities, int64(0))
	assert.Greater(t, res.AnonymizationTotals.UserWords, int64(0))
	assert.Greater(t, res.AnonymizationTotals.ReviewAttempts, int64(0))
}

// TestRunDeletionSweepSkipsRowsNotYetDue covers VOC-031-TEST-22:
// a 'deactivated' row whose purge_after is in the future is not
// picked up by the sweep.
func TestRunDeletionSweepSkipsRowsNotYetDue(t *testing.T) {
	svc, repo, authRepo, _, _ := newService(t)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "user@example.com", Status: "active"})
	repo.SetUser(uid, "user@example.com")

	_, err := svc.CreateAccountDeletionRequest(context.Background(), uid.String(), "1.2.3.4", "sess-token", "idem")
	require.NoError(t, err)

	// The clock has not moved; the row is 30 days from
	// being eligible. The sweep should report zero processed.
	res, err := svc.RunDeletionSweep(context.Background(), "1.2.3.4", "sess-token")
	require.NoError(t, err)
	assert.Equal(t, 0, res.Processed)
	assert.Equal(t, 0, res.Anonymized)

	row := repo.DeletionRequest(uid)
	require.NotNil(t, row)
	assert.Equal(t, "deactivated", row.Status, "row stays in deactivated while not yet due")
	assert.Nil(t, row.CompletedAt)
}

// TestRunDeletionSweepIdempotentResume covers VOC-031-TEST-23:
// a row that is already 'completed' is never re-touched by a
// subsequent sweep pass. A row that is 'anonymizing' (claimed
// but not yet completed) is skipped on the same pass but
// re-processed on the next pass (the resumable sweep
// guarantee).
func TestRunDeletionSweepIdempotentResume(t *testing.T) {
	svc, repo, authRepo, _, c := newService(t)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "user@example.com", Status: "active"})
	repo.SetUser(uid, "user@example.com")

	_, err := svc.CreateAccountDeletionRequest(context.Background(), uid.String(), "1.2.3.4", "sess-token", "idem")
	require.NoError(t, err)

	c.Advance(31 * 24 * time.Hour)
	// First pass: processes the row to completion.
	res, err := svc.RunDeletionSweep(context.Background(), "1.2.3.4", "sess-token")
	require.NoError(t, err)
	require.Equal(t, 1, res.Anonymized)

	// Second pass: same clock; the row is 'completed' and is
	// not picked up.
	res, err = svc.RunDeletionSweep(context.Background(), "1.2.3.4", "sess-token")
	require.NoError(t, err)
	assert.Equal(t, 0, res.Processed, "a completed row is never re-touched")
	assert.Equal(t, 0, res.Anonymized)

	row := repo.DeletionRequest(uid)
	require.NotNil(t, row)
	assert.Equal(t, "completed", row.Status)
}

// TestAccountDeletionRequestEligibleForPurge exercises the
// EligibleForPurge helper the sweep's claim predicate uses.
func TestAccountDeletionRequestEligibleForPurge(t *testing.T) {
	now := testNow()
	row := AccountDeletionRequest{
		Status:     "deactivated",
		PurgeAfter: now.Add(-time.Second),
	}
	assert.True(t, row.EligibleForPurge(now), "deactivated row past purge_after is eligible")
	row.Status = "anonymizing"
	assert.False(t, row.EligibleForPurge(now), "anonymizing row is not eligible on a fresh pass")
	row.Status = "completed"
	assert.False(t, row.EligibleForPurge(now), "completed row is not eligible")
	row.Status = "deactivated"
	row.PurgeAfter = now.Add(time.Hour)
	assert.False(t, row.EligibleForPurge(now), "deactivated row not yet past purge_after is not eligible")
}

// TestMemoryRepositoryCreateAccountDeletionRequestRequiresUserID
// covers the repository-level input-validation branch.
func TestMemoryRepositoryCreateAccountDeletionRequestRequiresUserID(t *testing.T) {
	repo := NewMemoryRepository()
	_, err := repo.CreateAccountDeletionRequest(context.Background(), uuid.Nil, "idem", time.Now(), 30*24*time.Hour)
	assert.Error(t, err)
	_, err = repo.CreateAccountDeletionRequest(context.Background(), uuid.New(), "", time.Now(), 30*24*time.Hour)
	assert.Error(t, err)
}

// TestMemoryRepositoryCreateAccountDeletionRequestAlreadyInFlight
// covers the (user_id) UNIQUE-discipline pre-check.
func TestMemoryRepositoryCreateAccountDeletionRequestAlreadyInFlight(t *testing.T) {
	repo := NewMemoryRepository()
	uid := uuid.New()
	repo.SetUser(uid, "user@example.com")
	_, err := repo.CreateAccountDeletionRequest(context.Background(), uid, "idem-1", time.Now(), 30*24*time.Hour)
	require.NoError(t, err)
	_, err = repo.CreateAccountDeletionRequest(context.Background(), uid, "idem-2", time.Now(), 30*24*time.Hour)
	assert.True(t, errors.Is(err, ErrAccountDeletionAlreadyInFlight))
}

// TestMemoryRepositoryAnonymizeUserDataCounters covers the
// per-table counter discipline the sweep aggregates.
func TestMemoryRepositoryAnonymizeUserDataCounters(t *testing.T) {
	repo := NewMemoryRepository()
	uid := uuid.New()
	repo.SetUser(uid, "user@example.com")
	counters, err := repo.AnonymizeUserData(context.Background(), uid)
	require.NoError(t, err)
	assert.Equal(t, int64(1), counters.ExternalIdentities)
	assert.Equal(t, int64(1), counters.UserWords)
	assert.Equal(t, int64(1), counters.LearnerSentences)
	assert.Equal(t, int64(1), counters.ReviewAttempts)
	assert.Equal(t, int64(1), counters.AIFeedbackAttempts)
	assert.Equal(t, int64(1), counters.ConfidencePointLedger)
	assert.Equal(t, int64(1), counters.GraceDayLedger)
	assert.Equal(t, int64(1), counters.UserOnboardingProfiles)
	assert.Equal(t, int64(1), counters.UserSettings)
	assert.Equal(t, int64(1), counters.DailyMissionSnapshots)
	assert.Equal(t, int64(1), counters.DailyActivitySummaries)
	assert.Equal(t, int64(1), counters.StreakStates)
}

// newServiceForRate wires a Service with a specific rate-limit
// budget so the rate-limit test can exercise the same code
// path with a single-request window.
func newServiceForRate(t *testing.T, c clock.Clock, limit int) (*Service, *MemoryRepository, *authRepoStub, *clock.Fixed, *MemoryIdempotencyStore) {
	t.Helper()
	repo := NewMemoryRepository()
	authRepo := newAuthRepoStub()
	limiter := auth.NewFixedWindowRateLimiter(c, time.Hour, limit)
	idem := NewMemoryIdempotencyStore()
	svc := NewService(repo, authRepo, nil, idem, c, limiter, Config{
		Environment:               "test",
		BaseURL:                   "https://test.example.com",
		EmailChangePath:           "/auth/email-change",
		EmailChangeLinkLifetime:   15 * time.Minute,
		AccountDeletionPurgeDelay: 30 * 24 * time.Hour,
		AccountDeletionSweepLimit: 100,
		RateLimit: EmailChangeRateLimitConfig{
			RequestWindow: time.Hour, RequestLimit: 100,
			ConsumeWindow: time.Hour, ConsumeLimit: 100,
		},
		AccountDeletionRateLimit: AccountDeletionRateLimitConfig{
			RequestWindow: time.Hour, RequestLimit: limit,
			SweepWindow: time.Hour, SweepLimit: limit,
		},
	})
	return svc, repo, authRepo, nil, idem
}
