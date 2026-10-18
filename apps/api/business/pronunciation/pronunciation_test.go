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
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"strings"
)

var testMP3 = []byte{'I', 'D', '3', 4, 0, 0, 0, 0, 0, 1, 0}

type fakeProvider struct {
	calls    int
	identity ProviderIdentity
	fn       func(context.Context, Utterance) ([]byte, error)
}

func (p *fakeProvider) Generate(ctx context.Context, text Utterance) ([]byte, error) {
	p.calls++
	if p.fn != nil {
		return p.fn(ctx, text)
	}
	return testMP3, nil
}

func (p *fakeProvider) Identity() ProviderIdentity {
	if p.identity.Provider != "" {
		return p.identity
	}
	return ProviderIdentity{Provider: "openai", Model: OpenAIModel, Voice: OpenAIVoice, InstructionVersion: OpenAIInstructionVersion}
}

type fakeRepo struct {
	utterance                       Utterance
	keys                            []string
	allowed                         bool
	audio                           []byte
	reserveErr                      error
	reserved, authorized, completed int
	finished                        []byte
	onReserve                       func()
}

func (r *fakeRepo) ResolveUtterance(_ context.Context, _ uuid.UUID, text, meaningID string) (Utterance, error) {
	r.authorized++
	if !r.allowed {
		return Utterance{}, ErrNotFound
	}
	u := r.utterance
	u.Text = text
	if meaningID != "" {
		u.MeaningID = meaningID
	}
	return u, nil
}
func (r *fakeRepo) Reserve(_ context.Context, key string, _ int, _ time.Time) (reservation, error) {
	r.reserved++
	r.keys = append(r.keys, key)
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
	provider.fn = func(ctx context.Context, text Utterance) ([]byte, error) {
		require.NoError(t, ctx.Err(), "paid request must finish after card navigation aborts caller")
		require.Equal(t, "greeting", text.Text)
		return testMP3, nil
	}
	audio, err = svc.Speak(parent, u, "greeting")
	require.NoError(t, err)
	require.False(t, audio.CacheHit)
	require.Equal(t, 1, provider.calls)
	require.Equal(t, testMP3, repo.finished)
	require.ErrorIs(t, parent.Err(), context.Canceled)
	provider.fn = func(context.Context, Utterance) ([]byte, error) {
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
			out, err := p.Generate(t.Context(), Utterance{Text: "greeting"})
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
	_, err := p.Generate(t.Context(), Utterance{Text: "greeting"})
	require.ErrorIs(t, err, ErrUnavailable)
}

func TestMeaningAudioIdentitySeparatesAllSensesAndProviderGuidance(t *testing.T) {
	u := Utterance{MeaningID: uuid.NewString(), Text: "close", Kind: "word", PartOfSpeech: "verb", Definition: "Move to a shut position.", Context: "Close the door.", Locale: "en-US"}
	identity := NewOpenAIProvider(OpenAIConfig{}).Identity()
	key := utteranceCacheKey(identity, u)
	require.Len(t, key, 64)
	require.Equal(t, key, utteranceCacheKey(identity, u))
	for _, mutate := range []func(*Utterance){
		func(v *Utterance) { v.MeaningID = uuid.NewString() },
		func(v *Utterance) { v.PartOfSpeech = "noun" },
		func(v *Utterance) { v.PartOfSpeech = "adjective" },
		func(v *Utterance) { v.PartOfSpeech = "adverb" },
		func(v *Utterance) { v.Kind = "example" },
		func(v *Utterance) { v.Text = "closed" },
		func(v *Utterance) { v.Definition = "Near in distance." },
		func(v *Utterance) { v.Context = "A close friend." },
		func(v *Utterance) { v.Locale = "en-GB" },
	} {
		v := u
		mutate(&v)
		require.NotEqual(t, key, utteranceCacheKey(identity, v))
	}
	for _, mutate := range []func(*ProviderIdentity){
		func(v *ProviderIdentity) { v.Provider = "other" },
		func(v *ProviderIdentity) { v.Model = "other" },
		func(v *ProviderIdentity) { v.Voice = "other" },
		func(v *ProviderIdentity) { v.InstructionVersion = "other" },
	} {
		v := identity
		mutate(&v)
		require.NotEqual(t, key, utteranceCacheKey(v, u))
	}
	repo := &fakeRepo{allowed: true, utterance: u}
	provider := &fakeProvider{}
	svc := newService(repo, provider, true, nil)
	user := uuid.New()
	for _, pos := range []string{"noun", "verb", "adjective", "adverb", "preposition", "other"} {
		repo.utterance.PartOfSpeech = pos
		id := uuid.NewString()
		out, err := svc.SpeakMeaning(t.Context(), user, "close", id)
		require.NoError(t, err)
		require.Equal(t, id, out.MeaningID)
		require.Equal(t, pos, out.PartOfSpeech)
	}
	keys := map[string]bool{}
	for _, key := range repo.keys {
		require.False(t, keys[key])
		keys[key] = true
	}
	before := repo.authorized
	_, err := svc.SpeakMeaning(t.Context(), user, "close", "arbitrary context")
	require.ErrorIs(t, err, ErrInvalid)
	require.Equal(t, before, repo.authorized)
}

func TestOpenAIProviderSenseInstructionsBoundedSingleRequest(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		mime   string
		body   []byte
		want   bool
	}{
		{"mpeg", 200, "audio/mpeg", testMP3, true},
		{"binary", 200, "application/octet-stream", testMP3, true},
		{"json", 200, "application/json", []byte(`{"secret":"provider failure"}`), false},
		{"invalid", 200, "audio/mpeg", []byte("not mp3"), false},
		{"oversize", 200, "audio/mpeg", append(append([]byte{}, testMP3...), make([]byte, MaxAudioBytes)...), false},
		{"error", 429, "application/json", []byte("private provider failure"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				require.Equal(t, "/audio/speech", r.URL.Path)
				require.Equal(t, "POST", r.Method)
				require.Equal(t, "Bearer private-test-key", r.Header.Get("Authorization"))
				var body struct {
					Model, Voice, Input, Instructions string
					ResponseFormat                    string `json:"response_format"`
					Speed                             float64
				}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				require.Equal(t, OpenAIModel, body.Model)
				require.Equal(t, "marin", body.Voice)
				require.Equal(t, "record", body.Input)
				require.Equal(t, "mp3", body.ResponseFormat)
				require.Equal(t, float64(1), body.Speed)
				require.Contains(t, body.Instructions, `"PartOfSpeech":"verb"`)
				require.Contains(t, body.Instructions, "Store sound for later use.")
				require.Contains(t, body.Instructions, "Speak only the exact input text")
				require.Contains(t, body.Instructions, "en-US")
				w.Header().Set("Content-Type", tc.mime)
				w.WriteHeader(tc.status)
				_, _ = w.Write(tc.body)
			}))
			defer server.Close()
			p := NewOpenAIProvider(OpenAIConfig{APIKey: "private-test-key", BaseURL: server.URL})
			out, err := p.Generate(t.Context(), Utterance{Text: "record", PartOfSpeech: "verb", Definition: "Store sound for later use.", Context: "Please record this lesson.", Locale: "en-US"})
			if tc.want {
				require.NoError(t, err)
				require.Equal(t, testMP3, out)
			} else {
				require.ErrorIs(t, err, ErrUnavailable)
				require.Nil(t, out)
				require.NotContains(t, err.Error(), "private")
			}
			require.Equal(t, 1, calls, "never retry a paid generation")
		})
	}
	calls := 0
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Redirect(w, r, "https://invalid.example/private", 307)
	}))
	defer redirect.Close()
	p := NewOpenAIProvider(OpenAIConfig{APIKey: "test", BaseURL: redirect.URL})
	_, err := p.Generate(t.Context(), Utterance{Text: "record"})
	require.ErrorIs(t, err, ErrUnavailable)
	require.Equal(t, 1, calls)
	missing := NewOpenAIProvider(OpenAIConfig{BaseURL: redirect.URL})
	_, err = missing.Generate(t.Context(), Utterance{Text: "record"})
	require.ErrorIs(t, err, ErrUnavailable)
	require.Equal(t, 1, calls, "missing credentials must not make a request")
}

func TestCloudflareLegacyRequestsCannotMasqueradeAsSelectedMeaning(t *testing.T) {
	repo := &fakeRepo{allowed: true}
	provider := &fakeProvider{identity: ProviderIdentity{Provider: "cloudflare", Model: Model, Voice: Voice, InstructionVersion: "1"}}
	svc := newService(repo, provider, true, nil)
	user := uuid.New()
	_, err := svc.SpeakMeaning(t.Context(), user, "record", uuid.NewString())
	require.ErrorIs(t, err, ErrUnavailable)
	require.Zero(t, repo.reserved, "incompatible provider must not consume paid budget")
	require.Zero(t, provider.calls)
	_, err = svc.Speak(t.Context(), user, "greeting")
	require.NoError(t, err, "unambiguous legacy text remains supported")
	require.Equal(t, 1, provider.calls)
}

func TestMeaningRevisionRejectsStalePOSDefinitionAndCanonicalWordBeforePayment(t *testing.T) {
	id := uuid.NewString()
	frozen := Utterance{MeaningID: id, WordText: "record", Text: "record", PartOfSpeech: "noun", Definition: "Information kept so it can be checked later.", Kind: "word", Locale: "en-US"}
	revision := meaningidentity.Revision(frozen.MeaningID, frozen.WordText, frozen.PartOfSpeech, frozen.Definition)
	require.Len(t, revision, 64)
	for _, mutate := range []func(*Utterance){
		func(u *Utterance) { u.PartOfSpeech = "verb" },
		func(u *Utterance) { u.Definition = "To store sound for later use." },
		func(u *Utterance) { u.WordText = "other" },
	} {
		t.Run("stale", func(t *testing.T) {
			live := frozen
			mutate(&live)
			repo := &fakeRepo{allowed: true, utterance: live, audio: testMP3}
			provider := &fakeProvider{}
			svc := newService(repo, provider, true, nil)
			_, err := svc.SpeakMeaningRevision(t.Context(), uuid.New(), "record", id, revision)
			require.ErrorIs(t, err, ErrNotFound)
			require.Zero(t, repo.reserved, "even a cached clip must not be looked up for stale exercise identity")
			require.Zero(t, provider.calls)
		})
	}
	repo := &fakeRepo{allowed: true, utterance: frozen, audio: testMP3}
	provider := &fakeProvider{}
	svc := newService(repo, provider, true, nil)
	audio, err := svc.SpeakMeaningRevision(t.Context(), uuid.New(), "record", id, revision)
	require.NoError(t, err)
	require.Equal(t, revision, audio.MeaningRevision)
	require.True(t, audio.CacheHit)
	for _, bad := range []struct{ id, revision string }{{"", revision}, {id, "not a hash"}, {id, strings.ToUpper(revision)}} {
		before := repo.authorized
		_, err = svc.SpeakMeaningRevision(t.Context(), uuid.New(), "record", bad.id, bad.revision)
		require.ErrorIs(t, err, ErrInvalid)
		require.Equal(t, before, repo.authorized)
	}
	// Only the outgoing provider guidance is truncated. A change beyond the first
	// 500 characters still invalidates the exercise's full catalog fingerprint.
	frozen.Definition = strings.Repeat("a", 500) + " original"
	repo.utterance = frozen
	revision = meaningidentity.Revision(id, frozen.WordText, frozen.PartOfSpeech, frozen.Definition)
	repo.utterance.Definition = strings.Repeat("a", 500) + " edited"
	before := repo.reserved
	_, err = svc.SpeakMeaningRevision(t.Context(), uuid.New(), "record", id, revision)
	require.ErrorIs(t, err, ErrNotFound)
	require.Equal(t, before, repo.reserved)
}

func TestOpenAIProviderBoundsReferenceWithoutChangingSpokenText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct{ Input, Instructions string }
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, "record", body.Input)
		start := strings.Index(body.Instructions, `{"PartOfSpeech"`)
		require.GreaterOrEqual(t, start, 0)
		var reference map[string]string
		require.NoError(t, json.Unmarshal([]byte(body.Instructions[start:]), &reference))
		require.Equal(t, 500, len([]rune(reference["Definition"])))
		require.Equal(t, 750, len([]rune(reference["Context"])))
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write(testMP3)
	}))
	defer server.Close()
	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "synthetic", BaseURL: server.URL})
	_, err := provider.Generate(t.Context(), Utterance{Text: "record", WordText: "record", PartOfSpeech: "verb", Definition: strings.Repeat("é", 600), Context: strings.Repeat("b", 900)})
	require.NoError(t, err)
}
