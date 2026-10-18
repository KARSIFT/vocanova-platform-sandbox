package pronunciation

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var testMP3 = []byte{'I', 'D', '3', 4, 0, 0, 0, 0, 0, 1, 0}

type fakeProvider struct {
	calls int
	fn    func(context.Context, string) ([]byte, error)
}

func (p *fakeProvider) Generate(ctx context.Context, text string) ([]byte, error) {
	p.calls++
	if p.fn != nil {
		return p.fn(ctx, text)
	}
	return testMP3, nil
}

type fakeRepo struct {
	allowed                         bool
	audio                           []byte
	reserveErr                      error
	reserved, authorized, completed int
	finished                        []byte
	onReserve                       func()
}

func (r *fakeRepo) AuthorizeText(context.Context, uuid.UUID, string) (bool, error) {
	r.authorized++
	return r.allowed, nil
}
func (r *fakeRepo) Reserve(context.Context, string, int, time.Time) (reservation, error) {
	r.reserved++
	if r.onReserve != nil {
		r.onReserve()
	}
	return reservation{audio: r.audio, generation: uuid.New()}, r.reserveErr
}
func (r *fakeRepo) Complete(ctx context.Context, _ string, _ uuid.UUID, data []byte, _ time.Time) error {
	r.completed++
	r.finished = data
	return ctx.Err()
}

func TestCanonicalAudioServiceGatesCacheAndCancellation(t *testing.T) {
	u := uuid.New()
	now := clock.Fixed{T: time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)}
	repo := &fakeRepo{allowed: true}
	provider := &fakeProvider{}
	svc := newService(repo, provider, false, now)
	_, err := svc.Speak(t.Context(), u, "greeting")
	require.ErrorIs(t, err, ErrUnavailable)
	require.Zero(t, repo.authorized)
	require.Zero(t, provider.calls)
	svc.enabled = true
	_, err = svc.Speak(t.Context(), u, " my private draft ")
	require.ErrorIs(t, err, ErrInvalid)
	require.Zero(t, repo.authorized)
	repo.allowed = false
	_, err = svc.Speak(t.Context(), u, "My private learner sentence.")
	require.ErrorIs(t, err, ErrNotFound)
	require.Zero(t, repo.reserved)
	require.Zero(t, provider.calls)
	repo.allowed = true
	repo.audio = testMP3
	audio, err := svc.Speak(t.Context(), u, "greeting")
	require.NoError(t, err)
	require.True(t, audio.CacheHit)
	require.Equal(t, "audio/mpeg", audio.MimeType)
	require.Zero(t, provider.calls)
	repo.audio = nil
	parent, cancel := context.WithCancel(t.Context())
	repo.onReserve = cancel
	provider.fn = func(ctx context.Context, text string) ([]byte, error) {
		require.NoError(t, ctx.Err(), "paid request must finish after card navigation aborts caller")
		require.Equal(t, "greeting", text)
		return testMP3, nil
	}
	audio, err = svc.Speak(parent, u, "greeting")
	require.NoError(t, err)
	require.False(t, audio.CacheHit)
	require.Equal(t, 1, provider.calls)
	require.Equal(t, testMP3, repo.finished)
	require.ErrorIs(t, parent.Err(), context.Canceled)
	provider.fn = func(context.Context, string) ([]byte, error) {
		return nil, errors.New("private provider response must not escape")
	}
	_, err = svc.Speak(t.Context(), u, "greeting")
	require.ErrorIs(t, err, ErrUnavailable)
	require.NotContains(t, err.Error(), "private")
	require.Nil(t, repo.finished)
}

func TestCloudflareAudioProviderSingleBoundedRequest(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		mime   string
		body   []byte
		want   bool
	}{
		{"mpeg", 200, "audio/mpeg", testMP3, true},
		{"binary", 200, "application/octet-stream", testMP3, true},
		{"json error", 200, "application/json", []byte(`{"success":false,"errors":[{"message":"private"}]}`), false},
		{"bad audio", 200, "audio/mpeg", []byte("not mp3"), false},
		{"oversize", 200, "audio/mpeg", append(append([]byte{}, testMP3...), make([]byte, MaxAudioBytes)...), false},
		{"provider failure", 503, "application/json", []byte("private provider failure"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				require.Equal(t, "/accounts/test-account/ai/run/"+Model, r.URL.Path)
				require.Equal(t, "Bearer private-test-token", r.Header.Get("Authorization"))
				var body map[string]string
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				require.Equal(t, map[string]string{"text": "greeting", "speaker": "luna", "encoding": "mp3"}, body)
				w.Header().Set("Content-Type", tc.mime)
				w.WriteHeader(tc.status)
				_, _ = w.Write(tc.body)
			}))
			defer server.Close()
			p := NewCloudflareProvider(CloudflareConfig{APIToken: "private-test-token", AccountID: "test-account", BaseURL: server.URL})
			out, err := p.Generate(t.Context(), "greeting")
			if tc.want {
				require.NoError(t, err)
				require.Equal(t, testMP3, out)
			} else {
				require.ErrorIs(t, err, ErrUnavailable)
				require.Nil(t, out)
				require.NotContains(t, err.Error(), "private")
			}
			require.Equal(t, 1, calls, "no paid provider retry")
		})
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://invalid.example/private", http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	p := NewCloudflareProvider(CloudflareConfig{APIToken: "token", AccountID: "account", BaseURL: server.URL})
	_, err := p.Generate(t.Context(), "greeting")
	require.ErrorIs(t, err, ErrUnavailable)
}
