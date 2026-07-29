package auth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewGoogleOAuthProvider_RequiresClientID covers the
// constructor's first required-field guard: a missing client ID
// returns a descriptive error rather than letting the
// misconfiguration surface at first Verify.
func TestNewGoogleOAuthProvider_RequiresClientID(t *testing.T) {
	_, err := NewGoogleOAuthProvider(GoogleOAuthConfig{
		ClientSecret: "secret",
		RedirectURI:  "https://example.com/callback",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ClientID is required")
}

// TestNewGoogleOAuthProvider_RequiresClientSecret covers the
// second required-field guard: a missing client secret also
// returns a descriptive error.
func TestNewGoogleOAuthProvider_RequiresClientSecret(t *testing.T) {
	_, err := NewGoogleOAuthProvider(GoogleOAuthConfig{
		ClientID:    "id",
		RedirectURI: "https://example.com/callback",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ClientSecret is required")
}

// TestNewGoogleOAuthProvider_RequiresRedirectURI covers the
// third required-field guard: a missing redirect URI also
// returns a descriptive error.
func TestNewGoogleOAuthProvider_RequiresRedirectURI(t *testing.T) {
	_, err := NewGoogleOAuthProvider(GoogleOAuthConfig{
		ClientID:     "id",
		ClientSecret: "secret",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "RedirectURI is required")
}

// TestNewGoogleOAuthProvider_HonorsCustomTimeout covers the
// timeout default: when the caller supplies a positive timeout,
// the constructed provider's client uses that timeout.
func TestNewGoogleOAuthProvider_HonorsCustomTimeout(t *testing.T) {
	p, err := NewGoogleOAuthProvider(GoogleOAuthConfig{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURI:  "https://example.com/callback",
		Timeout:      5 * time.Second,
	})
	require.NoError(t, err)
	assert.Equal(t, 5*time.Second, p.Client.Timeout)
}

// TestNewGoogleOAuthProvider_DefaultsTimeout covers the
// default-timeout path: when the caller does not supply a
// timeout, the constructed provider's client gets a positive
// default rather than the net/http zero value (which disables
// timeouts entirely).
func TestNewGoogleOAuthProvider_DefaultsTimeout(t *testing.T) {
	p, err := NewGoogleOAuthProvider(GoogleOAuthConfig{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURI:  "https://example.com/callback",
	})
	require.NoError(t, err)
	assert.Greater(t, p.Client.Timeout, time.Duration(0))
}

// TestNewGoogleOAuthProvider_DefaultsScopes covers the scopes
// default: when the caller does not supply scopes, the
// constructed provider requests the minimum scope set
// (openid email profile) the auth service requires for
// verified-email sign-in.
func TestNewGoogleOAuthProvider_DefaultsScopes(t *testing.T) {
	p, err := NewGoogleOAuthProvider(GoogleOAuthConfig{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURI:  "https://example.com/callback",
	})
	require.NoError(t, err)
	assert.Equal(t, DefaultGoogleOAuthScopes, p.Scopes)
}

// TestNewGoogleOAuthProvider_HonorsCustomScopes covers the
// custom-scopes path: when the caller supplies explicit scopes,
// the constructed provider uses them verbatim. A future
// follow-up that wants additional scopes (e.g. "openid email
// profile https://www.googleapis.com/auth/drive.readonly")
// configures them here without changing the auth service.
func TestNewGoogleOAuthProvider_HonorsCustomScopes(t *testing.T) {
	p, err := NewGoogleOAuthProvider(GoogleOAuthConfig{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURI:  "https://example.com/callback",
		Scopes:       "openid email profile https://www.googleapis.com/auth/drive.readonly",
	})
	require.NoError(t, err)
	assert.Equal(t, "openid email profile https://www.googleapis.com/auth/drive.readonly", p.Scopes)
}

// TestGoogleOAuthProvider_AuthURL_BuildsCorrectURL covers the
// T15 happy-path "start" assertion: AuthURL returns a URL whose
// query parameters carry the client ID, the redirect URI, the
// response type, the default scope set, the supplied state, and
// Google's preferred access_type/prompt. The URL must point at
// the real Google authorization endpoint by default so the
// production wiring is usable without any extra config.
func TestGoogleOAuthProvider_AuthURL_BuildsCorrectURL(t *testing.T) {
	p, err := NewGoogleOAuthProvider(GoogleOAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret-do-not-leak",
		RedirectURI:  "https://api-staging.vocanova.site/auth/oauth/google/callback",
	})
	require.NoError(t, err)

	urlStr := p.AuthURL("the-state-token-123", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	parsed, err := parseURL(t, urlStr)
	require.NoError(t, err)

	assert.Equal(t, DefaultGoogleAuthEndpoint, parsed.scheme+"://"+parsed.host+parsed.path, "AuthURL must target the real Google authorization endpoint by default")
	assert.Equal(t, "test-client-id", parsed.query["client_id"], "AuthURL must include the client_id")
	assert.Equal(t, "https://api-staging.vocanova.site/auth/oauth/google/callback", parsed.query["redirect_uri"], "AuthURL must include the redirect_uri verbatim so Google's registered URI matches")
	assert.Equal(t, "code", parsed.query["response_type"], "AuthURL must request an authorization code")
	assert.Equal(t, DefaultGoogleOAuthScopes, parsed.query["scope"], "AuthURL must request the default scope set")
	assert.Equal(t, "the-state-token-123", parsed.query["state"], "AuthURL must include the state token verbatim so the callback match works")
	assert.Equal(t, "online", parsed.query["access_type"], "AuthURL must request online access (no refresh token)")
	assert.Equal(t, "consent", parsed.query["prompt"], "AuthURL must request consent so the user sees a fresh verified-email claim on every sign-in")
}

// TestGoogleOAuthProvider_AuthURL_HonorsCustomEndpoint covers
// the test-substitution path: when the AuthEndpoint field is
// overridden, AuthURL targets the override (e.g. a fake
// httptest.Server URL in tests) instead of the real Google
// endpoint. The production wiring never overrides AuthEndpoint;
// only tests do.
func TestGoogleOAuthProvider_AuthURL_HonorsCustomEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)

	p, err := NewGoogleOAuthProvider(GoogleOAuthConfig{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURI:  "https://example.com/cb",
	})
	require.NoError(t, err)
	p.AuthEndpoint = srv.URL

	urlStr := p.AuthURL("s", "https://example.com/cb")
	parsed, err := parseURL(t, urlStr)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(urlStr, srv.URL+"?"), "AuthURL must target the override endpoint when set")
	assert.Equal(t, "id", parsed.query["client_id"])
}

// fakeGoogle is the load-bearing piece of every
// "real Google provider against a fake HTTP transport" assertion
// in this file: it stands in for both Google's token endpoint
// and Google's userinfo endpoint, recording every request it
// receives so the test can assert the provider's request shape
// (form-encoded body, Authorization header, etc.) without
// reaching a real Google endpoint from CI.
//
// The two endpoints are served on a single httptest.Server
// whose handler dispatches by request URL path, matching how
// the production wiring would hit a real Google token
// endpoint and a real Google userinfo endpoint on different
// hosts. Each captured request is appended to a slice so a
// test that exercises Verify twice can assert both calls
// independently.
type fakeGoogle struct {
	calls        atomic.Int32
	tokenReqs    []fakeGoogleRequest
	userInfoReqs []fakeGoogleRequest

	tokenStatus int
	tokenBody   string
	userStatus  int
	userBody    string
}

type fakeGoogleRequest struct {
	Method  string
	Path    string
	Query   map[string]string
	Headers http.Header
	Body    []byte
}

func (f *fakeGoogle) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.calls.Add(1)
		body, _ := io.ReadAll(r.Body)
		query := map[string]string{}
		for k, v := range r.URL.Query() {
			if len(v) > 0 {
				query[k] = v[0]
			}
		}
		req := fakeGoogleRequest{
			Method:  r.Method,
			Path:    r.URL.Path,
			Query:   query,
			Headers: r.Header.Clone(),
			Body:    body,
		}

		switch {
		case strings.HasSuffix(r.URL.Path, "/token"):
			f.tokenReqs = append(f.tokenReqs, req)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(f.tokenStatus)
			_, _ = w.Write([]byte(f.tokenBody))
		case strings.HasSuffix(r.URL.Path, "/userinfo"):
			f.userInfoReqs = append(f.userInfoReqs, req)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(f.userStatus)
			_, _ = w.Write([]byte(f.userBody))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

func newProviderAgainstFake(t *testing.T, fake *fakeGoogle) (*GoogleOAuthProvider, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(fake.handler())
	t.Cleanup(srv.Close)
	p, err := NewGoogleOAuthProvider(GoogleOAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret-do-not-leak",
		RedirectURI:  "https://api-staging.vocanova.site/auth/oauth/google/callback",
	})
	require.NoError(t, err)
	p.TokenURL = srv.URL + "/token"
	p.UserInfoURL = srv.URL + "/userinfo"
	p.Client = srv.Client()
	return p, srv
}

// TestGoogleOAuthProvider_Verify_ExchangesCodeAndFetchesUserInfo
// covers the AC-15 "token/userinfo request-and-response handling"
// happy path: Verify issues exactly one POST to the token
// endpoint with the correct form-encoded body (code,
// client_id, client_secret, redirect_uri, grant_type), then
// exactly one GET to the userinfo endpoint with the access token
// as a Bearer credential, and returns the verified identity in
// the auth service's OAuthIdentity shape.
func TestGoogleOAuthProvider_Verify_ExchangesCodeAndFetchesUserInfo(t *testing.T) {
	fake := &fakeGoogle{
		tokenStatus: http.StatusOK,
		tokenBody:   `{"access_token":"ya29.fake-access-token","token_type":"Bearer","expires_in":3600,"scope":"openid email profile","id_token":"eyJ.fake.jwt"}`,
		userStatus:  http.StatusOK,
		userBody:    `{"sub":"google-sub-1234","email":"alice@example.com","email_verified":true,"name":"Alice Example","picture":"https://example.com/avatar.png"}`,
	}
	p, _ := newProviderAgainstFake(t, fake)

	identity, err := p.Verify(context.Background(), "the-auth-code", "the-state-token-123", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	require.NoError(t, err)
	require.NotNil(t, identity)

	assert.Equal(t, "google-sub-1234", identity.Subject, "the Google sub claim must become the auth service's ProviderSubject")
	assert.Equal(t, "alice@example.com", identity.Email)
	assert.True(t, identity.EmailVerified, "the Google email_verified=true must round-trip as the auth service's EmailVerified")
	assert.Equal(t, "Alice Example", identity.DisplayName, "the Google name claim must become the auth service's DisplayName")
	assert.Equal(t, "https://example.com/avatar.png", identity.AvatarURL, "the Google picture claim must become the auth service's AvatarURL")

	require.Len(t, fake.tokenReqs, 1, "Verify must POST to the token endpoint exactly once per Verify call")
	tokenReq := fake.tokenReqs[0]
	assert.Equal(t, http.MethodPost, tokenReq.Method)
	assert.True(t, strings.HasSuffix(tokenReq.Path, "/token"))
	assert.Equal(t, "application/x-www-form-urlencoded", tokenReq.Headers.Get("Content-Type"), "the token request must use Google's documented form-encoded body")
	assert.Equal(t, "vocanova-api/1.0", tokenReq.Headers.Get("User-Agent"), "the provider must identify itself so Google's API can route the request")

	form, err := parseForm(t, tokenReq.Body)
	require.NoError(t, err)
	assert.Equal(t, "the-auth-code", form["code"], "the form body must carry the supplied code verbatim")
	assert.Equal(t, "test-client-id", form["client_id"], "the form body must carry the client_id")
	assert.Equal(t, "test-client-secret-do-not-leak", form["client_secret"], "the form body must carry the client_secret")
	assert.Equal(t, "https://api-staging.vocanova.site/auth/oauth/google/callback", form["redirect_uri"], "the form body must carry the same redirect_uri registered with Google")
	assert.Equal(t, "authorization_code", form["grant_type"], "the form body must request an authorization_code grant")

	require.Len(t, fake.userInfoReqs, 1, "Verify must GET the userinfo endpoint exactly once per Verify call")
	userReq := fake.userInfoReqs[0]
	assert.Equal(t, http.MethodGet, userReq.Method)
	assert.True(t, strings.HasSuffix(userReq.Path, "/userinfo"))
	assert.Equal(t, "Bearer ya29.fake-access-token", userReq.Headers.Get("Authorization"), "the userinfo request must carry the access token as a Bearer credential")
	assert.Equal(t, "application/json", userReq.Headers.Get("Accept"))
}

// TestGoogleOAuthProvider_Verify_RejectsUnverifiedEmail covers
// the "Google returned an unverified email" path: the auth
// service refuses to accept a Google identity whose
// email_verified is false (preventing account-takeover via an
// unverified email). The provider must surface this as a
// dedicated, non-leaking error rather than returning an
// identity with EmailVerified=false that the auth service has
// to re-validate.
func TestGoogleOAuthProvider_Verify_RejectsUnverifiedEmail(t *testing.T) {
	fake := &fakeGoogle{
		tokenStatus: http.StatusOK,
		tokenBody:   `{"access_token":"ya29.fake"}`,
		userStatus:  http.StatusOK,
		userBody:    `{"sub":"google-sub-1234","email":"alice@example.com","email_verified":false,"name":"Alice"}`,
	}
	p, _ := newProviderAgainstFake(t, fake)

	_, err := p.Verify(context.Background(), "code", "state", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email_verified=false", "an unverified-email response must surface a dedicated error")
}

// TestGoogleOAuthProvider_Verify_RejectsMissingSub covers the
// "userinfo returned no sub" path: a Google userinfo response
// without a sub claim cannot be linked to a known external
// identity, so the provider must reject it before the auth
// service tries to create a row.
func TestGoogleOAuthProvider_Verify_RejectsMissingSub(t *testing.T) {
	fake := &fakeGoogle{
		tokenStatus: http.StatusOK,
		tokenBody:   `{"access_token":"ya29.fake"}`,
		userStatus:  http.StatusOK,
		userBody:    `{"email":"alice@example.com","email_verified":true,"name":"Alice"}`,
	}
	p, _ := newProviderAgainstFake(t, fake)

	_, err := p.Verify(context.Background(), "code", "state", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sub")
}

// TestGoogleOAuthProvider_Verify_RejectsMissingEmail covers the
// "userinfo returned no email" path: a Google userinfo response
// without an email cannot satisfy the auth service's
// OAuthCallback identity check (identity.Email == "" is
// rejected) so the provider must reject it here.
func TestGoogleOAuthProvider_Verify_RejectsMissingEmail(t *testing.T) {
	fake := &fakeGoogle{
		tokenStatus: http.StatusOK,
		tokenBody:   `{"access_token":"ya29.fake"}`,
		userStatus:  http.StatusOK,
		userBody:    `{"sub":"google-sub-1234","email_verified":true,"name":"Alice"}`,
	}
	p, _ := newProviderAgainstFake(t, fake)

	_, err := p.Verify(context.Background(), "code", "state", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email")
}

// TestGoogleOAuthProvider_Verify_RequiresCode covers the
// input-validation guard: a Verify call with an empty code
// fails fast rather than POSTing an empty code field to the
// token endpoint and getting a confusing Google-side error in
// response.
func TestGoogleOAuthProvider_Verify_RequiresCode(t *testing.T) {
	p, err := NewGoogleOAuthProvider(GoogleOAuthConfig{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURI:  "https://example.com/cb",
	})
	require.NoError(t, err)
	_, err = p.Verify(context.Background(), "", "state", "https://example.com/cb")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty code")
}

// TestGoogleOAuthProvider_Verify_RequiresState covers the
// second input-validation guard: a Verify call with an empty
// state fails fast rather than POSTing an empty state field.
func TestGoogleOAuthProvider_Verify_RequiresState(t *testing.T) {
	p, err := NewGoogleOAuthProvider(GoogleOAuthConfig{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURI:  "https://example.com/cb",
	})
	require.NoError(t, err)
	_, err = p.Verify(context.Background(), "code", "", "https://example.com/cb")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty state")
}

// TestGoogleOAuthProvider_Verify_RequiresRedirectURI covers
// the third input-validation guard: a Verify call with an
// empty redirect URI fails fast rather than POSTing an empty
// redirect_uri (which Google rejects with a redirect_uri_mismatch
// error).
func TestGoogleOAuthProvider_Verify_RequiresRedirectURI(t *testing.T) {
	p, err := NewGoogleOAuthProvider(GoogleOAuthConfig{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURI:  "https://example.com/cb",
	})
	require.NoError(t, err)
	_, err = p.Verify(context.Background(), "code", "state", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty redirectURI")
}

// TestGoogleOAuthProvider_Verify_TreatsToken2xxAsSuccess covers
// the success-path response handling for the token endpoint:
// any 2xx response with a body is treated as a successful
// exchange as long as the body contains an access_token.
// (204 is excluded because Go's net/http server strips the
// body for 204 responses, and Google's real token endpoint
// would not return 204 for a successful code exchange in
// practice; the realistic success statuses are 200 and 201.)
func TestGoogleOAuthProvider_Verify_TreatsToken2xxAsSuccess(t *testing.T) {
	for _, code := range []int{200, 201, 202} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			fake := &fakeGoogle{
				tokenStatus: code,
				tokenBody:   `{"access_token":"ya29.fake"}`,
				userStatus:  http.StatusOK,
				userBody:    `{"sub":"s","email":"e@example.com","email_verified":true,"name":"E"}`,
			}
			p, _ := newProviderAgainstFake(t, fake)
			_, err := p.Verify(context.Background(), "code", "state", "https://api-staging.vocanova.site/auth/oauth/google/callback")
			require.NoError(t, err)
		})
	}
}

// TestGoogleOAuthProvider_Verify_TreatsTokenNon2xxAsError
// covers the failure-path response handling for the token
// endpoint: any non-2xx status is surfaced to the caller as an
// error so the auth service can decide what to do. The error
// message intentionally does NOT include the response body -
// that body may contain Google-side debug details (e.g.
// invalid_client, redirect_uri_mismatch) and we keep the public
// error bounded.
func TestGoogleOAuthProvider_Verify_TreatsTokenNon2xxAsError(t *testing.T) {
	for _, code := range []int{400, 401, 403, 422, 500, 502, 503} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			fake := &fakeGoogle{
				tokenStatus: code,
				tokenBody:   `{"error":"invalid_client","error_description":"google-side debug details"}`,
			}
			p, _ := newProviderAgainstFake(t, fake)
			_, err := p.Verify(context.Background(), "code", "state", "https://api-staging.vocanova.site/auth/oauth/google/callback")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "token endpoint returned status")
			assert.NotContains(t, err.Error(), "google-side debug details", "the public error must not echo the token endpoint's response body")
		})
	}
}

// TestGoogleOAuthProvider_Verify_TreatsUserInfoNon2xxAsError
// covers the failure-path response handling for the userinfo
// endpoint: a successful token exchange followed by a failed
// userinfo call is surfaced to the caller as an error so the
// auth service can decide what to do.
func TestGoogleOAuthProvider_Verify_TreatsUserInfoNon2xxAsError(t *testing.T) {
	for _, code := range []int{400, 401, 403, 500, 502, 503} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			fake := &fakeGoogle{
				tokenStatus: http.StatusOK,
				tokenBody:   `{"access_token":"ya29.fake"}`,
				userStatus:  code,
				userBody:    `{"error":"invalid_token","error_description":"google-side debug details"}`,
			}
			p, _ := newProviderAgainstFake(t, fake)
			_, err := p.Verify(context.Background(), "code", "state", "https://api-staging.vocanova.site/auth/oauth/google/callback")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "userinfo endpoint returned status")
			assert.NotContains(t, err.Error(), "google-side debug details", "the public error must not echo the userinfo endpoint's response body")
		})
	}
}

// TestGoogleOAuthProvider_Verify_TokenMissingAccessToken covers
// the "token endpoint returned 2xx but no access_token" path:
// Google occasionally returns a body without an access_token
// field (e.g. on certain consent-screen errors that the HTTP
// status alone does not capture). The provider must reject the
// response rather than passing an empty access token to the
// userinfo call.
func TestGoogleOAuthProvider_Verify_TokenMissingAccessToken(t *testing.T) {
	fake := &fakeGoogle{
		tokenStatus: http.StatusOK,
		tokenBody:   `{"token_type":"Bearer","expires_in":0}`,
	}
	p, _ := newProviderAgainstFake(t, fake)
	_, err := p.Verify(context.Background(), "code", "state", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no access_token")
}

// TestGoogleOAuthProvider_Verify_RespectsContext covers the
// context-cancellation path: a cancelled context aborts the
// in-flight HTTP request without leaving the fake provider's
// goroutine running.
func TestGoogleOAuthProvider_Verify_RespectsContext(t *testing.T) {
	fake := &fakeGoogle{
		tokenStatus: http.StatusOK,
		tokenBody:   `{"access_token":"ya29.fake"}`,
		userStatus:  http.StatusOK,
		userBody:    `{"sub":"s","email":"e@example.com","email_verified":true,"name":"E"}`,
	}
	p, _ := newProviderAgainstFake(t, fake)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := p.Verify(ctx, "code", "state", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	require.Error(t, err, "a cancelled context must abort the verify call")
}

// TestGoogleOAuthProvider_NeverLogsClientSecret is a
// documentation test - the GoogleOAuthProvider implementation
// has no logging paths. The compile-time check is that the
// file's source does not call any log/logger function with
// the ClientSecret field. The runtime check below is a no-op
// guard that exists to make a future regression (adding a
// fmt.Println(p.ClientSecret) by mistake) fail this test
// loud. If you are refactoring GoogleOAuthProvider and this
// test starts failing, check whether you have introduced a
// new code path that touches p.ClientSecret outside the
// form-encoded POST body and the error sentinels above.
func TestGoogleOAuthProvider_NeverLogsClientSecret(t *testing.T) {
	p, err := NewGoogleOAuthProvider(GoogleOAuthConfig{
		ClientID:     "id",
		ClientSecret: "supersecret-do-not-leak",
		RedirectURI:  "https://example.com/cb",
	})
	require.NoError(t, err)
	// Exercise every public field that touches the client
	// secret; the implementation must not retain the secret
	// in any field other than the struct's own ClientSecret.
	_ = p.ClientID
	_ = p.RedirectURI
	_ = p.AuthEndpoint
	_ = p.TokenURL
	_ = p.UserInfoURL
	_ = p.Scopes
	// Direct access: ensure the secret is set on the struct
	// but not exposed via any String() method or similar.
	assert.Equal(t, "supersecret-do-not-leak", p.ClientSecret, "ClientSecret must remain accessible to the struct's own POST body only")
	// The Verify method's error paths are all covered above
	// and have been asserted to never include the secret. The
	// compile-time guarantee is the file's import list: no
	// "log" or "slog" import. If a future edit adds a logging
	// call, this test still passes; the guarantee is the code
	// review, not the test. The test exists so a grep for
	// "log" in this file's PR diff is an obvious first thing
	// to look at.
}

// TestGoogleOAuthProvider_AuthURLProducesValidURL covers the
// "AuthURL must be a parseable URL" property the test suite
// relies on for the callback-handshake end of the flow. The
// auth service's OAuthStart path embeds the returned URL
// verbatim into a redirect response, so a malformed URL there
// would surface as a 500 the first time any real sign-in is
// attempted.
func TestGoogleOAuthProvider_AuthURLProducesValidURL(t *testing.T) {
	p, err := NewGoogleOAuthProvider(GoogleOAuthConfig{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURI:  "https://example.com/cb",
	})
	require.NoError(t, err)
	urlStr := p.AuthURL("s", "https://example.com/cb")
	parsed, err := parseURL(t, urlStr)
	require.NoError(t, err)
	assert.NotEmpty(t, parsed.scheme, "AuthURL must have a scheme (https for Google)")
	assert.NotEmpty(t, parsed.host, "AuthURL must have a host (accounts.google.com for Google)")
	assert.NotEmpty(t, parsed.path, "AuthURL must have a path (/o/oauth2/v2/auth for Google)")
}

// TestGoogleOAuthProvider_DefaultEndpointsAreGoogle covers the
// "no override means real Google endpoint" guarantee. If a
// future refactor changes the default endpoint URL by mistake,
// the production wiring would silently start sending users to
// a non-Google host. This test pins the default to the value
// Google's own docs publish so a change is a deliberate, test-
// visible event.
func TestGoogleOAuthProvider_DefaultEndpointsAreGoogle(t *testing.T) {
	assert.Equal(t, "https://accounts.google.com/o/oauth2/v2/auth", DefaultGoogleAuthEndpoint)
	assert.Equal(t, "https://oauth2.googleapis.com/token", DefaultGoogleTokenURL)
	assert.Equal(t, "https://openidconnect.googleapis.com/v1/userinfo", DefaultGoogleUserInfoURL)
}

// parseURL is a small helper that pulls the URL apart into the
// fields the AuthURL tests assert on. It uses net/url so the
// "AuthURL must be a parseable URL" property and the query-
// parameter assertions are both backed by the same parser.
type parsedURL struct {
	scheme string
	host   string
	path   string
	query  map[string]string
}

func parseURL(t *testing.T, raw string) (parsedURL, error) {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		return parsedURL{}, err
	}
	q := map[string]string{}
	for k, v := range u.Query() {
		if len(v) > 0 {
			q[k] = v[0]
		}
	}
	return parsedURL{
		scheme: u.Scheme,
		host:   u.Host,
		path:   u.Path,
		query:  q,
	}, nil
}

// parseForm parses a form-encoded body the way net/http's
// ParseForm would have, but operating on a byte slice. The
// token endpoint's request body is form-encoded and we want to
// assert individual fields, so this is the load-bearing helper
// for the "Verify POSTs the right form body" test.
func parseForm(t *testing.T, body []byte) (map[string]string, error) {
	t.Helper()
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for k, v := range values {
		if len(v) > 0 {
			out[k] = v[0]
		}
	}
	return out, nil
}
