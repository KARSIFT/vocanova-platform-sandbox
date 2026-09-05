package aifeedback

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type finalizerAwareRepository struct {
	Repository
	finalize func(context.Context, PendingAttempt, *ProviderFeedback, time.Time, SuccessfulFeedbackCompletion) (bool, error)
}

func (r *finalizerAwareRepository) CompleteSuccessfulFeedbackAttempt(ctx context.Context, pending PendingAttempt, feedback *ProviderFeedback, now time.Time, completion SuccessfulFeedbackCompletion) (bool, error) {
	return r.finalize(ctx, pending, feedback, now, completion)
}

type countedMissionAccounting struct{ calls int }

func (m *countedMissionAccounting) Update(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	m.calls++
	return false, nil
}

func TestTransactionalFinalizerRejectsNontransactionalMission(t *testing.T) {
	f := newServiceFixture(t)
	mission := &countedMissionAccounting{}
	f.service.mission = mission
	f.service.repo = &finalizerAwareRepository{Repository: f.repo, finalize: func(ctx context.Context, _ PendingAttempt, _ *ProviderFeedback, _ time.Time, completion SuccessfulFeedbackCompletion) (bool, error) {
		// The guard must reject this configuration before touching the SQL tx
		// or calling an updater that would open an independent transaction.
		return completion(ctx, &sql.Tx{})
	}}
	_, err := f.service.SubmitSentenceFeedback(t.Context(), f.request("I work every day."))
	require.ErrorContains(t, err, "transaction-aware mission updater required")
	require.Zero(t, mission.calls)
	require.Equal(t, AttemptStatusFailed, f.repo.attempts[0].Status)
}

func TestAtomicFinalizerCancellationSettlesPendingWithoutAccounting(t *testing.T) {
	f := newServiceFixture(t)
	mission := &countedMissionAccounting{}
	f.service.mission = mission
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	f.service.repo = &finalizerAwareRepository{Repository: f.repo, finalize: func(ctx context.Context, pending PendingAttempt, feedback *ProviderFeedback, now time.Time, completion SuccessfulFeedbackCompletion) (bool, error) {
		cancel()
		return f.repo.CompleteSuccessfulFeedbackAttempt(ctx, pending, feedback, now, completion)
	}}
	result, err := f.service.SubmitSentenceFeedback(ctx, f.request("I work every day."))
	require.NoError(t, err)
	require.Equal(t, ErrorCodeTemporaryFailure, result.ErrorCode)
	require.Zero(t, mission.calls)
	require.Equal(t, AttemptStatusFailed, f.repo.attempts[0].Status)
}

func TestAtomicFinalizerAmbiguousCommitPreservesSuccessAndAccounting(t *testing.T) {
	f := newServiceFixture(t)
	mission := &countedMissionAccounting{}
	f.service.mission = mission
	f.service.repo = &finalizerAwareRepository{Repository: f.repo, finalize: func(ctx context.Context, pending PendingAttempt, feedback *ProviderFeedback, now time.Time, completion SuccessfulFeedbackCompletion) (bool, error) {
		completed, err := f.repo.CompleteSuccessfulFeedbackAttempt(ctx, pending, feedback, now, completion)
		if err != nil {
			return false, err
		}
		return completed, errors.New("commit acknowledgement unavailable")
	}}
	req := f.request("I work every day.")
	_, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.ErrorContains(t, err, "commit acknowledgement unavailable")
	require.Equal(t, AttemptStatusSucceeded, f.repo.attempts[0].Status)
	require.Equal(t, 1, mission.calls)
	result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	require.Equal(t, LearningStatusCorrect, result.Status)
	require.Equal(t, 1, mission.calls)
	require.Equal(t, 1, f.provider.calls)
}
