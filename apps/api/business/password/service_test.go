package password

import (
	"context"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/email"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingLimiter struct {
	keys []string
	ok   bool
}

func (l *recordingLimiter) Allow(_ context.Context, key string) (bool, error) {
	l.keys = append(l.keys, key)
	return l.ok, nil
}

func testService(l auth.RateLimiter, enabled func() bool, signup, identity func(string) bool) *Service {
	return NewService(nil, &email.Fake{}, &clock.Fixed{T: time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)}, l, "https://example.test", "test", time.Hour, enabled, signup, identity)
}

func TestServiceDisabledBeforeAnyDatabaseOrPasswordWork(t *testing.T) {
	limiter := &recordingLimiter{ok: true}
	svc := testService(limiter, func() bool { return false }, func(string) bool { return true }, func(string) bool { return true })
	token, _, err := auth.NewTokenAndHash()
	require.NoError(t, err)
	assert.ErrorIs(t, svc.RequestSignup(t.Context(), "127.0.0.1", "learner@example.test", "correct horse battery staple", "Learner"), ErrDisabled)
	assert.ErrorIs(t, svc.VerifySignup(t.Context(), "127.0.0.1", token), ErrDisabled)
	assert.ErrorIs(t, svc.RequestReset(t.Context(), "127.0.0.1", "learner@example.test"), ErrDisabled)
	assert.ErrorIs(t, svc.Reset(t.Context(), "127.0.0.1", token, "correct horse battery staple"), ErrDisabled)
	_, _, _, err = svc.Login(t.Context(), "127.0.0.1", "learner@example.test", "correct horse battery staple")
	assert.ErrorIs(t, err, ErrDisabled)
	assert.Empty(t, limiter.keys)
}

func TestServicePoliciesStillPassThroughRateLimits(t *testing.T) {
	limiter := &recordingLimiter{ok: true}
	svc := testService(limiter, func() bool { return true }, func(string) bool { return false }, func(string) bool { return false })
	assert.NoError(t, svc.RequestSignup(t.Context(), "127.0.0.1", "reserved@example.test", "correct horse battery staple", "Learner"))
	assert.NoError(t, svc.RequestReset(t.Context(), "127.0.0.1", "reserved@example.test"))
	_, _, _, err := svc.Login(t.Context(), "127.0.0.1", "reserved@example.test", "correct horse battery staple")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Len(t, limiter.keys, 6, "each policy-denied anonymous action consumes IP and hashed-email budget")
	for _, key := range limiter.keys {
		assert.NotContains(t, key, "reserved@example.test")
	}
}

func TestServiceRejectsMalformedProofBeforeLimiter(t *testing.T) {
	limiter := &recordingLimiter{ok: true}
	svc := testService(limiter, func() bool { return true }, nil, nil)
	assert.ErrorIs(t, svc.VerifySignup(t.Context(), "127.0.0.1", "not-a-token"), ErrInvalidToken)
	assert.ErrorIs(t, svc.Reset(t.Context(), "127.0.0.1", "not-a-token", "correct horse battery staple"), ErrInvalidToken)
	assert.Empty(t, limiter.keys)
}
