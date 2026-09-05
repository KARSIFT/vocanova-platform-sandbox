package aifeedback

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMemoryRepositoryCreateQualityReviewReportIdempotency exercises the
// repository contract independently of the service's ownership checks. The
// memory implementation is deliberately small, but must preserve the same
// user/key/fingerprint and rollback-like ordering rules as PostgreSQL.
func TestMemoryRepositoryCreateQualityReviewReportIdempotency(t *testing.T) {
	repo := NewMemoryRepository(MemoryRepositoryData{})
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	user, otherUser := uuid.New(), uuid.New()
	firstAttempt, secondAttempt, thirdAttempt := uuid.New(), uuid.New(), uuid.New()
	report := func(userID, attemptID uuid.UUID, reason string, at time.Time) QualityReviewReport {
		return QualityReviewReport{ID: uuid.New(), UserID: userID, AttemptID: attemptID, Reason: reason, State: QualityReviewStateOpen, CreatedAt: at}
	}

	created, err := repo.CreateQualityReviewReport(t.Context(), report(user, firstAttempt, ReportReasonAlreadyCorrect, now), "shared")
	require.NoError(t, err)
	assert.True(t, created)

	// An exact retry is a no-op, and changing the attempt/reason with the
	// still-active key is a conflict.
	created, err = repo.CreateQualityReviewReport(t.Context(), report(user, firstAttempt, ReportReasonAlreadyCorrect, now), "shared")
	require.NoError(t, err)
	assert.False(t, created)
	created, err = repo.CreateQualityReviewReport(t.Context(), report(user, firstAttempt, ReportReasonAlreadyCorrect, now), "different-key")
	require.NoError(t, err)
	assert.False(t, created, "one report per attempt applies even with a distinct key")
	_, err = repo.CreateQualityReviewReport(t.Context(), report(user, secondAttempt, ReportReasonAlreadyCorrect, now), "shared")
	assert.ErrorIs(t, err, ErrReportIdempotencyConflict)
	assert.Len(t, repo.QualityReviewReports(), 1)

	// A key is scoped to its learner: another learner can use the same key.
	created, err = repo.CreateQualityReviewReport(t.Context(), report(otherUser, uuid.New(), ReportReasonAlreadyCorrect, now), "shared")
	require.NoError(t, err)
	assert.True(t, created)

	// The exact 24-hour boundary expires a claim and permits a fresh report.
	created, err = repo.CreateQualityReviewReport(t.Context(), report(user, secondAttempt, ReportReasonAlreadyCorrect, now.Add(24*time.Hour)), "shared")
	require.NoError(t, err)
	assert.True(t, created)

	// A reason conflict for an already reported attempt must not claim its
	// fresh key. That key remains usable for a legitimate different attempt.
	_, err = repo.CreateQualityReviewReport(t.Context(), report(user, firstAttempt, ReportReasonExplanationUnclear, now), "unpoisoned")
	assert.True(t, errors.Is(err, ErrReportIdempotencyConflict))
	created, err = repo.CreateQualityReviewReport(t.Context(), report(user, thirdAttempt, ReportReasonExplanationUnclear, now), "unpoisoned")
	require.NoError(t, err)
	assert.True(t, created)
	assert.Len(t, repo.QualityReviewReports(), 4)
}
