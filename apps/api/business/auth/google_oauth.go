package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GoogleOAuthProvider is the real, production-wiring Google OAuth
// implementation. It implements the OAuthProvider interface defined
// in auth.go and is the T15 counterpart to the existing
// NewFakeOAuthProvider used by the contract / test wiring.
//
// T15's scope is the "no real Google OAuth provider exists"
// portion of VOC-032-D10: the production deployment, including
// the staging host, currently has no way to actually exchange a
// Google authorization code for a verified identity. This
// implementation closes that gap by adding a Google-conformant
// OAuth 2.0 / OpenID Connect token exchange plus userinfo fetch,
// using only the credential fields the founder provisions under
// VOC-032-DEP-07 (GOOGLE_OAUTH_CLIENT_ID, GOOGLE_OAUTH_CLIENT_SECRET,
// OAUTH_REDIRECT_URI) and never a literal in code.
//
// GoogleOAuthProvider is provider-shaped after Google's
// documented endpoints (https://accounts.google.com/o/oauth2/v2/auth
// for authorization, https://oauth2.googleapis.com/token for the
// code exchange, https://openidconnect.googleapis.com/v1/userinfo for
// the verified identity) but takes its token and userinfo URLs
// from the config struct so a fake HTTP transport (the unit-test
// pattern T14 established with email.HTTPSender) can intercept
// every call without reaching a real Google endpoint from CI.
// The constructor defaults both URLs to the real Google
// endpoints when left empty, so the production wiring never has
// to set them.
//
// GoogleOAuthProvider is safe for concurrent use; the only
// mutable state is the *http.Client, which net/http documents as
// safe for concurrent use across requests.
type GoogleOAuthProvider struct {
	// TokenURL is the HTTPS endpoint GoogleOAuthProvider POSTs
	// the authorization-code exchange to. Defaults to Google's
	// real token endpoint when empty; tests substitute a fake
	// httptest.Server URL here.
	TokenURL string

	// UserInfoURL is the HTTPS endpoint GoogleOAuthProvider
	// queries for the verified identity after a successful code
	// exchange. Defaults to Google's real OIDC userinfo
	// endpoint when empty; tests substitute a fake URL.
	UserInfoURL string

	// AuthEndpoint is the HTTPS endpoint the AuthURL method
	// builds. Defaults to Google's real authorization endpoint
	// when empty.
	AuthEndpoint string

	// ClientID is the OAuth 2.0 client identifier Google issued
	// for this app. Read from GOOGLE_OAUTH_CLIENT_ID via the
	// production-wiring buildOAuthProvider, never from a literal
	// in code.
	ClientID string

	// ClientSecret is the OAuth 2.0 client secret Google issued
	// for this app. Read from GOOGLE_OAUTH_CLIENT_SECRET via the
	// production-wiring buildOAuthProvider, never from a literal
	// in code. GoogleOAuthProvider never logs ClientSecret and
	// never includes it in any returned error message.
	ClientSecret string

	// RedirectURI is the absolute URL the user-agent is sent
	// back to after completing the Google consent screen. Must
	// match exactly the URI registered for this client in
	// Google Cloud Console. The production-wiring reads this
	// from OAUTH_REDIRECT_URI (the same value the auth service's
	// Config uses) so web-side and api-side redirect handling
	// agree on a single value.
	RedirectURI string

	// Scopes are the space-separated OAuth 2.0 scopes
	// GoogleOAuthProvider requests at the authorization step.
	// Defaults to "openid email profile" when empty - the
	// minimum scope set that returns a verified email and
	// display name. The userinfo call's response carries
	// email_verified, which the auth service's
	// OAuthCallback path treats as the gating signal that
	// prevents account-takeover via unverified email.
	Scopes string

	// Client is the *http.Client used for the token and
	// userinfo requests. nil falls back to a default
	// *http.Client with a positive timeout.
	Client *http.Client
}

// GoogleOAuthConfig captures the configuration NewGoogleOAuthProvider
// needs. Callers that need a custom transport (e.g. a fake transport
// for tests) should construct the GoogleOAuthProvider directly
// rather than calling this constructor, so the test can substitute
// Client.
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       string
	Timeout      time.Duration
}

// DefaultGoogleOAuthEndpoints are the real Google endpoints the
// production wiring targets when the corresponding URL fields on
// GoogleOAuthProvider are empty. They are package-level constants
// (not config defaults applied by NewGoogleOAuthProvider) so a
// test that wants to verify "an empty TokenURL was correctly
// defaulted" can do so without comparing against a string the
// constructor itself filled in.
const (
	DefaultGoogleAuthEndpoint = "https://accounts.google.com/o/oauth2/v2/auth"
	DefaultGoogleTokenURL     = "https://oauth2.googleapis.com/token"
	DefaultGoogleUserInfoURL  = "https://openidconnect.googleapis.com/v1/userinfo"
	DefaultGoogleOAuthScopes  = "openid email profile"
	defaultGoogleOAuthTimeout = 8 * time.Second
)

// NewGoogleOAuthProvider returns a GoogleOAuthProvider with the
// real Google endpoints as defaults and a *http.Client whose
// timeout is set from cfg. Required fields (ClientID,
// ClientSecret, RedirectURI) are validated up front; an empty cfg
// returns a descriptive error rather than letting the
// misconfiguration surface at first Verify.
func NewGoogleOAuthProvider(cfg GoogleOAuthConfig) (*GoogleOAuthProvider, error) {
	if strings.TrimSpace(cfg.ClientID) == "" {
		return nil, errors.New("auth: GoogleOAuthConfig.ClientID is required")
	}
	if strings.TrimSpace(cfg.ClientSecret) == "" {
		return nil, errors.New("auth: GoogleOAuthConfig.ClientSecret is required")
	}
	if strings.TrimSpace(cfg.RedirectURI) == "" {
		return nil, errors.New("auth: GoogleOAuthConfig.RedirectURI is required")
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultGoogleOAuthTimeout
	}
	scopes := cfg.Scopes
	if strings.TrimSpace(scopes) == "" {
		scopes = DefaultGoogleOAuthScopes
	}
	return &GoogleOAuthProvider{
		AuthEndpoint: DefaultGoogleAuthEndpoint,
		TokenURL:     DefaultGoogleTokenURL,
		UserInfoURL:  DefaultGoogleUserInfoURL,
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURI:  cfg.RedirectURI,
		Scopes:       scopes,
		Client:       &http.Client{Timeout: timeout},
	}, nil
}

// googleTokenResponse is the JSON body Google returns from a
// successful authorization-code exchange. Only the access_token
// field is required for the T15 verify path; the others are
// parsed to make a future refresh-token flow easier to add
// without re-introducing a different shape.
type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	IDToken     string `json:"id_token"`
}

// googleUserInfo is the JSON body Google's OIDC userinfo endpoint
// returns. The auth service's OAuthCallback path requires
// email_verified=true to accept an identity; sub is the immutable
// per-user Google identifier that the auth service stores as
// ProviderSubject (so a later sign-in from the same Google
// account is correctly linked).
type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// AuthURL builds the Google authorization URL for the supplied
// state token and redirect URI. The state is included verbatim
// (URL-encoded) so the auth service can match it against the
// state cookie on callback; the redirect URI is included exactly
// as supplied so the value registered in Google Cloud Console
// matches the value the auth service sends. prompt=consent is
// requested so the user sees the consent screen on every sign-in
// (and Google therefore returns a fresh id_token whose
// email_verified reflects the user's current verified status);
// access_type=online is requested so Google does not return a
// refresh token, which T15 has no use for and which the auth
// service does not persist.
func (p *GoogleOAuthProvider) AuthURL(state, redirectURI string) string {
	endpoint := p.AuthEndpoint
	if endpoint == "" {
		endpoint = DefaultGoogleAuthEndpoint
	}
	scopes := p.Scopes
	if scopes == "" {
		scopes = DefaultGoogleOAuthScopes
	}
	q := url.Values{}
	q.Set("client_id", p.ClientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", scopes)
	q.Set("state", state)
	q.Set("access_type", "online")
	q.Set("prompt", "consent")
	q.Set("include_granted_scopes", "true")
	return endpoint + "?" + q.Encode()
}

// Verify exchanges the supplied authorization code with Google's
// token endpoint, then fetches the verified identity from
// Google's userinfo endpoint. It returns an *OAuthIdentity in the
// same shape NewFakeOAuthProvider already returns so the
// auth service's OAuthCallback path is provider-agnostic.
//
// Verify never logs the client secret and never includes it in
// any returned error message. Verify does not retry; the caller
// (the auth service) decides retry policy.
func (p *GoogleOAuthProvider) Verify(ctx context.Context, code, state, redirectURI string) (*OAuthIdentity, error) {
	if strings.TrimSpace(code) == "" {
		return nil, errors.New("auth: Verify called with empty code")
	}
	if strings.TrimSpace(state) == "" {
		return nil, errors.New("auth: Verify called with empty state")
	}
	if strings.TrimSpace(redirectURI) == "" {
		return nil, errors.New("auth: Verify called with empty redirectURI")
	}

	accessToken, err := p.exchangeCodeForToken(ctx, code, redirectURI)
	if err != nil {
		return nil, err
	}

	identity, err := p.fetchUserInfo(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	return identity, nil
}

// exchangeCodeForToken POSTs the authorization-code exchange to
// the configured token endpoint. The token endpoint URL defaults
// to Google's real endpoint; tests substitute a fake. The request
// body is form-encoded (the format Google documents for this
// endpoint). The client secret is sent only in the POST body
// and never logged or returned.
func (p *GoogleOAuthProvider) exchangeCodeForToken(ctx context.Context, code, redirectURI string) (string, error) {
	tokenURL := p.TokenURL
	if tokenURL == "" {
		tokenURL = DefaultGoogleTokenURL
	}
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", p.ClientID)
	form.Set("client_secret", p.ClientSecret)
	form.Set("redirect_uri", redirectURI)
	form.Set("grant_type", "authorization_code")

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("auth: build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "vocanova-api/1.0")

	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: defaultGoogleOAuthTimeout}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("auth: token request: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("auth: token endpoint returned status %d", resp.StatusCode)
	}

	var body googleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("auth: decode token response: %w", err)
	}
	if body.AccessToken == "" {
		return "", errors.New("auth: token endpoint returned no access_token")
	}
	return body.AccessToken, nil
}

// fetchUserInfo GETs the userinfo endpoint with the supplied
// access token as a Bearer credential, then returns the verified
// identity in the auth service's OAuthIdentity shape. A userinfo
// response whose email_verified is false is rejected here (not
// in the auth service) so the production-wiring path can return
// a stable, dedicated error without conflating "Google could
// not be reached" with "Google returned an unverified email".
func (p *GoogleOAuthProvider) fetchUserInfo(ctx context.Context, accessToken string) (*OAuthIdentity, error) {
	userInfoURL := p.UserInfoURL
	if userInfoURL == "" {
		userInfoURL = DefaultGoogleUserInfoURL
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("auth: build userinfo request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "vocanova-api/1.0")

	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: defaultGoogleOAuthTimeout}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("auth: userinfo request: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("auth: userinfo endpoint returned status %d", resp.StatusCode)
	}

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("auth: decode userinfo response: %w", err)
	}
	if info.Sub == "" {
		return nil, errors.New("auth: userinfo response missing sub")
	}
	if info.Email == "" {
		return nil, errors.New("auth: userinfo response missing email")
	}
	if !info.EmailVerified {
		return nil, errors.New("auth: userinfo response has email_verified=false")
	}
	return &OAuthIdentity{
		Subject:       info.Sub,
		Email:         info.Email,
		EmailVerified: true,
		DisplayName:   info.Name,
		AvatarURL:     info.Picture,
	}, nil
}
