package accounts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryRepository is a deterministic in-memory implementation of
// Repository for service tests. It also implements the
// UpdateUserEmail discipline the SQL repository applies: a write
// that would collide with another user's active email is rejected
// with ErrEmailAlreadyRegistered, never a generic 500, so service
// tests can exercise the same race-handling the production path
// will.
type MemoryRepository struct {
	mu     sync.Mutex
	byID   map[uuid.UUID]*EmailChangeLink
	byHash map[string]*EmailChangeLink
	byUser map[uuid.UUID]map[uuid.UUID]*EmailChangeLink
	// usersByEmail mirrors the lower(email) WHERE deleted_at IS NULL
	// partial unique index users_active_email_key enforces in SQL.
	// The map holds the current email for every active user; nil
	// emails are not represented.
	usersByEmail map[string]uuid.UUID
	users        map[uuid.UUID]*memoryUser
	// collisions is incremented every time UpdateUserEmail rejects
	// a write because newEmail is already taken. Tests assert on
	// it directly; production code never reads it.
	collisions int
	// deletionRequests holds every account_deletion_requests
	// row the in-memory store has seen, keyed by the row's
	// primary id. The (user_id) UNIQUE constraint is enforced
	// at write time.
	deletionRequests map[uuid.UUID]*AccountDeletionRequest
	// sessionsRevoked counts the (user, session) pairs the
	// deactivation transaction has revoked. Tests assert on
	// it directly.
	sessionsRevoked int
	// magicLinksRevoked counts the (user, magic-link) pairs
	// the deactivation transaction has revoked.
	magicLinksRevoked int
	// emailChangeLinksRevoked counts the (user,
	// email-change-link) pairs the deactivation transaction
	// has revoked.
	emailChangeLinksRevoked int
	// anonymizeCounters aggregates the per-table counts the
	// in-memory AnonymizeUserData has produced. Tests assert
	// on it directly to confirm the per-table disposition
	// actually ran.
	anonymizeCounters AnonymizationCounters
}

// memoryUser is the minimal projection the in-memory repository
// needs about a user: the current email, a soft-delete flag
// matching users.deleted_at IS NULL, and the deleted_at
// timestamp the deactivation transaction writes.
type memoryUser struct {
	ID        uuid.UUID
	Email     string
	Deleted   bool
	DeletedAt *time.Time
}

// NewMemoryRepository creates an empty in-memory Repository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		byID:             make(map[uuid.UUID]*EmailChangeLink),
		byHash:           make(map[string]*EmailChangeLink),
		byUser:           make(map[uuid.UUID]map[uuid.UUID]*EmailChangeLink),
		usersByEmail:     make(map[string]uuid.UUID),
		users:            make(map[uuid.UUID]*memoryUser),
		deletionRequests: make(map[uuid.UUID]*AccountDeletionRequest),
	}
}

// ExportPersonalData provides a minimal deterministic contract fixture. The
// production repository is responsible for enumerating persisted history.
func (r *MemoryRepository) ExportPersonalData(ctx context.Context, userID uuid.UUID) (json.RawMessage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[userID]
	if !ok {
		return nil, ErrUserNotFound
	}
	return json.Marshal(map[string]any{
		"schemaVersion": "1.0", "profile": map[string]any{"id": u.ID.String(), "email": u.Email},
		"settings": map[string]any{"timezone": "UTC", "dailyReviewTarget": 20, "reviewIntervalPreset": "vocanova_default", "notificationsEnabled": true, "marketingEmailsEnabled": false, "appLanguage": "en", "createdAt": nil, "updatedAt": nil}, "onboardingProfile": nil, "savedWords": []any{}, "reviewHistory": []any{},
		"sentenceFeedbackHistory": []any{}, "dailyMissions": []any{}, "dailyActivity": []any{},
		"confidencePointLedger": []any{}, "graceDayLedger": []any{}, "streakState": nil,
	})
}

// CreateEmailChangeLink inserts one row. Mirrors the SQL
// migration's invariants (user_id is required, new_email is
// non-empty, token_hash is unique).
func (r *MemoryRepository) CreateEmailChangeLink(ctx context.Context, userID uuid.UUID, newEmail string, tokenHash []byte, environment string, createdAt, expiresAt time.Time) (*EmailChangeLink, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if userID == uuid.Nil {
		return nil, errors.New("user id required")
	}
	if newEmail == "" {
		return nil, errors.New("new email required")
	}
	if _, exists := r.byHash[string(tokenHash)]; exists {
		return nil, errors.New("token hash collision")
	}
	link := &EmailChangeLink{
		ID:          uuid.New(),
		UserID:      userID,
		NewEmail:    strings.ToLower(newEmail),
		Environment: environment,
		CreatedAt:   createdAt.UTC(),
		ExpiresAt:   expiresAt.UTC(),
	}
	r.byID[link.ID] = link
	r.byHash[string(tokenHash)] = link
	if _, ok := r.byUser[userID]; !ok {
		r.byUser[userID] = make(map[uuid.UUID]*EmailChangeLink)
	}
	r.byUser[userID][link.ID] = link
	return cloneLink(link), nil
}

// GetEmailChangeLinkByTokenHash returns the projection or a
// not-found error.
func (r *MemoryRepository) GetEmailChangeLinkByTokenHash(ctx context.Context, tokenHash []byte) (*EmailChangeLink, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.byHash[string(tokenHash)]
	if !ok {
		return nil, errors.New("email change link not found")
	}
	return cloneLink(l), nil
}

// ConsumeEmailChangeLink marks the row consumed.
func (r *MemoryRepository) ConsumeEmailChangeLink(ctx context.Context, id uuid.UUID, consumedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.byID[id]
	if !ok {
		return errors.New("email change link not found")
	}
	t := consumedAt.UTC()
	l.ConsumedAt = &t
	return nil
}

// RevokeAllEmailChangeLinksForUser revokes every unconsumed link
// for the user. Used by the account-deletion path (T04).
func (r *MemoryRepository) RevokeAllEmailChangeLinksForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	links, ok := r.byUser[userID]
	if !ok {
		return 0, nil
	}
	var n int64
	t := revokedAt.UTC()
	for _, l := range links {
		if l.ConsumedAt == nil && l.RevokedAt == nil {
			l.RevokedAt = &t
			n++
		}
	}
	return n, nil
}

// UpdateUserEmail applies the new email to the user, enforcing the
// same uniqueness discipline the SQL partial unique index provides:
// another active user already owns lower(newEmail), so we reject
// with ErrEmailAlreadyRegistered. Soft-deleted users are not
// considered active, matching the partial-index WHERE clause.
func (r *MemoryRepository) UpdateUserEmail(ctx context.Context, userID uuid.UUID, newEmail string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	newEmail = strings.ToLower(strings.TrimSpace(newEmail))
	if newEmail == "" {
		return errors.New("new email required")
	}
	u, ok := r.users[userID]
	if !ok {
		return ErrUserNotFound
	}
	if u.Deleted {
		return ErrUserNotFound
	}
	if owner, exists := r.usersByEmail[newEmail]; exists && owner != userID {
		// The owning user is checked for "deleted" so the
		// partial-index semantics translate exactly: a
		// soft-deleted user does not occupy the email.
		if other, ok := r.users[owner]; !ok || !other.Deleted {
			r.collisions++
			return ErrEmailAlreadyRegistered
		}
	}
	// Release the old email from the index.
	if u.Email != "" {
		delete(r.usersByEmail, u.Email)
	}
	u.Email = newEmail
	r.usersByEmail[newEmail] = userID
	return nil
}

// SetUser is a test helper that seeds the in-memory users table
// with an email identity. It is not part of the Repository
// contract; the production wiring uses auth.GetUserByID for
// identity and never lets the accounts.Repository mutate users
// directly.
func (r *MemoryRepository) SetUser(userID uuid.UUID, email string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[userID] = &memoryUser{ID: userID, Email: strings.ToLower(email)}
	if email != "" {
		r.usersByEmail[strings.ToLower(email)] = userID
	}
}

// SetUserDeleted marks a user as soft-deleted so the partial-
// unique-index semantics are observable in tests.
func (r *MemoryRepository) SetUserDeleted(userID uuid.UUID, deleted bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u, ok := r.users[userID]; ok {
		u.Deleted = deleted
	}
}

// UserEmail returns the current in-memory email for a user. Used
// only by tests; not part of the contract.
func (r *MemoryRepository) UserEmail(userID uuid.UUID) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u, ok := r.users[userID]; ok {
		return u.Email
	}
	return ""
}

// Collisions returns the number of duplicate-email rejections the
// repository has produced since construction. Tests assert on it
// directly to confirm the partial-index discipline fires.
func (r *MemoryRepository) Collisions() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.collisions
}

// LinksForUser returns the in-memory email_change_links for a
// user, sorted by CreatedAt. Test helper only.
func (r *MemoryRepository) LinksForUser(userID uuid.UUID) []EmailChangeLink {
	r.mu.Lock()
	defer r.mu.Unlock()
	links, ok := r.byUser[userID]
	if !ok {
		return nil
	}
	out := make([]EmailChangeLink, 0, len(links))
	for _, l := range links {
		out = append(out, *cloneLink(l))
	}
	return out
}

func cloneLink(l *EmailChangeLink) *EmailChangeLink {
	c := *l
	if l.ConsumedAt != nil {
		t := *l.ConsumedAt
		c.ConsumedAt = &t
	}
	if l.RevokedAt != nil {
		t := *l.RevokedAt
		c.RevokedAt = &t
	}
	return &c
}

// CreateAccountDeletionRequest performs the deactivation
// transaction in-memory. The semantics match the SQL
// implementation: deactivate the user (status='deleted',
// deleted_at=now), revoke every active session, every
// unconsumed magic link, every unconsumed email change link,
// and insert the account_deletion_requests row. The
// (user_id) UNIQUE constraint is enforced at the in-memory
// write so a second deactivation for the same user surfaces
// ErrAccountDeletionAlreadyInFlight exactly the way the SQL
// path does.
func (r *MemoryRepository) CreateAccountDeletionRequest(ctx context.Context, userID uuid.UUID, idempotencyKey string, now time.Time, purgeDelay time.Duration) (*AccountDeletionRequest, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user id required")
	}
	if idempotencyKey == "" {
		return nil, errors.New("idempotency key required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	// Pre-check: a deletion for this user is already in
	// flight. The SQL path relies on a unique-violation
	// translation; the memory path translates the
	// pre-check directly.
	for _, row := range r.deletionRequests {
		if row.UserID == userID {
			if row.IdempotencyKey == idempotencyKey {
				result := cloneDeletion(row)
				result.Replayed = true
				return result, nil
			}
			return nil, fmt.Errorf("%w: deletion already in flight", ErrAccountDeletionAlreadyInFlight)
		}
	}

	u, ok := r.users[userID]
	if !ok || u.Deleted {
		return nil, ErrUserNotFound
	}
	now = now.UTC()
	t := now
	u.Deleted = true
	u.DeletedAt = &t
	// Release the email from the partial-unique-index
	// discipline: a deactivated user does not occupy the
	// email. The email string is kept on the projection so
	// existing readers see the last-known value, but
	// usersByEmail is the authoritative source for "can
	// another user claim this address".
	if u.Email != "" {
		delete(r.usersByEmail, u.Email)
	}

	// The session / magic-link / email-change-link revocation
	// counters are incremented here. The test reads them to
	// confirm the transaction invoked each step. Production
	// would call the corresponding auth.Repository methods;
	// the in-memory stand-in counts the calls instead.
	r.sessionsRevoked++
	r.magicLinksRevoked++
	r.emailChangeLinksRevoked++

	row := &AccountDeletionRequest{
		ID:             uuid.New(),
		UserID:         userID,
		Status:         "deactivated",
		RequestedAt:    now,
		PurgeAfter:     now.Add(purgeDelay),
		IdempotencyKey: idempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	r.deletionRequests[row.ID] = row
	return cloneDeletion(row), nil
}

// GetAccountDeletionRequestByUserID returns the current row
// for the user, or an error.
func (r *MemoryRepository) GetAccountDeletionRequestByUserID(ctx context.Context, userID uuid.UUID) (*AccountDeletionRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, row := range r.deletionRequests {
		if row.UserID == userID {
			return cloneDeletion(row), nil
		}
	}
	return nil, errors.New("account deletion request not found")
}

// ListDeactivatedRequestsDueForPurge returns due deactivated rows plus stale
// anonymizing claims. Fresh claims are deliberately excluded so a live worker
// cannot lose ownership to another sweep.
func (r *MemoryRepository) ListDeactivatedRequestsDueForPurge(ctx context.Context, now, staleBefore time.Time, limit int) ([]AccountDeletionRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now = now.UTC()
	if limit <= 0 {
		limit = 100
	}
	var out []AccountDeletionRequest
	for _, row := range r.deletionRequests {
		if (row.Status == "deactivated" && !row.PurgeAfter.After(now)) ||
			(row.Status == "anonymizing" && !row.UpdatedAt.After(staleBefore)) {
			out = append(out, *cloneDeletion(row))
		}
	}
	// Sort by PurgeAfter ascending to match the SQL ORDER BY.
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].PurgeAfter.Before(out[i].PurgeAfter) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// ClaimAccountDeletionRequestForAnonymization atomically
// transitions a due row from 'deactivated' to 'anonymizing', or reclaims a
// stale anonymizing claim.
// Returns true when this caller now owns the row, false when it is missing,
// completed, or still covered by a fresh claim.
func (r *MemoryRepository) ClaimAccountDeletionRequestForAnonymization(ctx context.Context, id uuid.UUID, now, staleBefore time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	row, ok := r.deletionRequests[id]
	if !ok {
		return false, nil
	}
	if row.Status != "deactivated" && (row.Status != "anonymizing" || row.UpdatedAt.After(staleBefore)) {
		return false, nil
	}
	row.Status = "anonymizing"
	row.UpdatedAt = now.UTC()
	return true, nil
}

// AnonymizeUserData runs the per-table disposition for
// the user. The memory implementation does not have
// underlying tables to mutate (the per-table
// de-identification only exists at the SQL level); the
// counters it returns reflect the work that *would* have
// run, so tests can assert on the same shape the
// production sweep would produce. A test helper
// (AnonymizeCounters) exposes the per-table counts.
func (r *MemoryRepository) AnonymizeUserData(ctx context.Context, userID uuid.UUID) (AnonymizationCounters, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if userID == uuid.Nil {
		return AnonymizationCounters{}, errors.New("user id required")
	}
	// The in-memory store has no per-table data to mutate.
	// Each table's counter is incremented by 1 to record
	// that the disposition ran for at least one row of
	// that class; tests that need a non-zero counter use
	// SetAnonymizeCounters to seed the value before the
	// sweep runs.
	r.anonymizeCounters.ExternalIdentities++
	r.anonymizeCounters.UserWords++
	r.anonymizeCounters.LearnerSentences++
	r.anonymizeCounters.ReviewAttempts++
	r.anonymizeCounters.AIFeedbackAttempts++
	r.anonymizeCounters.ConfidencePointLedger++
	r.anonymizeCounters.GraceDayLedger++
	r.anonymizeCounters.UserOnboardingProfiles++
	r.anonymizeCounters.UserSettings++
	r.anonymizeCounters.DailyMissionSnapshots++
	r.anonymizeCounters.DailyActivitySummaries++
	r.anonymizeCounters.StreakStates++
	return r.anonymizeCounters, nil
}

// FinalizeAccountDeletionClaim keeps ownership validation, the in-memory
// purge model, and completion under one mutex just as PostgreSQL keeps them
// under the request-row lock and transaction.
func (r *MemoryRepository) FinalizeAccountDeletionClaim(ctx context.Context, id, userID uuid.UUID, claimedAt, now time.Time) (AnonymizationCounters, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	row, ok := r.deletionRequests[id]
	if !ok || row.UserID != userID || row.Status != "anonymizing" || !row.UpdatedAt.Equal(claimedAt) {
		return AnonymizationCounters{}, false, nil
	}
	r.anonymizeCounters.ExternalIdentities++
	r.anonymizeCounters.UserWords++
	r.anonymizeCounters.LearnerSentences++
	r.anonymizeCounters.ReviewAttempts++
	r.anonymizeCounters.AIFeedbackAttempts++
	r.anonymizeCounters.ConfidencePointLedger++
	r.anonymizeCounters.GraceDayLedger++
	r.anonymizeCounters.UserOnboardingProfiles++
	r.anonymizeCounters.UserSettings++
	r.anonymizeCounters.DailyMissionSnapshots++
	r.anonymizeCounters.DailyActivitySummaries++
	r.anonymizeCounters.StreakStates++
	row.Status = "completed"
	completedAt := now.UTC()
	row.CompletedAt = &completedAt
	row.UpdatedAt = completedAt
	return r.anonymizeCounters, true, nil
}

// MarkAccountDeletionRequestCompleted transitions a row from 'anonymizing' to
// 'completed' and stamps completed_at. FinalizeAccountDeletionClaim is the
// fenced service path; this compatibility helper is idempotent.
func (r *MemoryRepository) MarkAccountDeletionRequestCompleted(ctx context.Context, id uuid.UUID, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	row, ok := r.deletionRequests[id]
	if !ok {
		return errors.New("account deletion request not found")
	}
	if row.Status == "completed" {
		return nil
	}
	row.Status = "completed"
	t := now.UTC()
	row.CompletedAt = &t
	row.UpdatedAt = t
	return nil
}

// SessionsRevoked is a test helper that exposes the
// per-call revocation count the in-memory deactivation
// transaction has produced.
func (r *MemoryRepository) SessionsRevoked() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return int64(r.sessionsRevoked)
}

// MagicLinksRevoked is a test helper that exposes the
// per-call magic-link revocation count.
func (r *MemoryRepository) MagicLinksRevoked() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return int64(r.magicLinksRevoked)
}

// EmailChangeLinksRevoked is a test helper that exposes
// the per-call email-change-link revocation count.
func (r *MemoryRepository) EmailChangeLinksRevoked() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return int64(r.emailChangeLinksRevoked)
}

// DeletionRequest is a test helper that returns the
// in-memory account_deletion_requests row for a user, if
// any.
func (r *MemoryRepository) DeletionRequest(userID uuid.UUID) *AccountDeletionRequest {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, row := range r.deletionRequests {
		if row.UserID == userID {
			return cloneDeletion(row)
		}
	}
	return nil
}

func cloneDeletion(r *AccountDeletionRequest) *AccountDeletionRequest {
	c := *r
	if r.CompletedAt != nil {
		t := *r.CompletedAt
		c.CompletedAt = &t
	}
	return &c
}
