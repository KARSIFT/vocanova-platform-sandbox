package aifeedback

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type temporarilyUnavailableMission struct {
	calls   int
	applied bool
}

func (m *temporarilyUnavailableMission) Update(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	m.calls++
	if m.calls == 1 {
		return false, errors.New("mission transaction unavailable")
	}
	m.applied = true
	return false, nil
}

func TestFeedbackRetryRecoversFailedMissionAccounting(t *testing.T) {
	f := newServiceFixture(t)
	mission := &temporarilyUnavailableMission{}
	f.service.mission = mission
	req := f.request("I work every day.")
	_, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.Error(t, err)
	require.False(t, mission.applied)

	// A transport replay must not expose a committed success with missing
	// accounting. It returns the recorded, retryable failure instead.
	replay, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	require.Equal(t, ErrorCodeTemporaryFailure, replay.ErrorCode)
	require.True(t, replay.CanRetry)
	require.False(t, mission.applied)

	// DOC-09 requires an explicit fresh key for a retry, which creates a new
	// generation while retaining that failure in history.
	req.IdempotencyKey = "mission-accounting-retry"
	result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, mission.applied, "a successful retry must not permanently skip failed mission/reward accounting")
}
