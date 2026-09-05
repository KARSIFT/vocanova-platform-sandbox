package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/danielgtaylor/huma/v2"
)

// authHumaContext returns the huma.Context or nil. It panics in tests if the
// middleware is not installed, making missing setup obvious.
func authHumaContext(ctx context.Context) huma.Context {
	if c := HumaContext(ctx); c != nil {
		return c
	}
	panic("huma context middleware not installed")
}

// sessionTokenFromHuma reads the raw session bearer token from the request
// cookie, for handlers that need to rate-limit per-session (in addition to
// per-IP) rather than just resolving the requester's identity.
func sessionTokenFromHuma(c huma.Context, svc *auth.Service) string {
	return sessionCookieValue(c, svc.SessionCookieName())
}

// clientIPFromHuma extracts a best-effort client IP from the Huma context.
func clientIPFromHuma(c huma.Context) string {
	if xff := c.Header("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xri := c.Header("X-Real-Ip"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, _ := strings.Cut(c.RemoteAddr(), ":")
	return host
}

// csrfCookieValue reads the configured double-submit CSRF cookie from the request.
func csrfCookieValue(c huma.Context, name string) string {
	ck, err := huma.ReadCookie(c, name)
	if err != nil {
		return ""
	}
	return ck.Value
}

// oauthStateCookieValue reads the OAuth state cookie from the request.
func oauthStateCookieValue(c huma.Context, name string) string {
	ck, err := huma.ReadCookie(c, name)
	if err != nil {
		return ""
	}
	return ck.Value
}

// RequestMagicLinkInput requests a single-use sign-in link.
type RequestMagicLinkInput struct {
	Body struct {
		Email string `json:"email" format:"email" doc:"Email address to send the sign-in link to"`
	}
}

// RequestMagicLinkOutput is intentionally empty to avoid account enumeration.
type RequestMagicLinkOutput struct{}

// ConsumeMagicLinkInput consumes a single-use sign-in link.
type ConsumeMagicLinkInput struct {
	Body struct {
		Token string `json:"token" doc:"The single-use token from the sign-in link"`
		Email string `json:"email" format:"email" doc:"The email address the link was sent to"`
	}
}

// ConsumeMagicLinkOutput returns the authenticated user and sets a session cookie.
type ConsumeMagicLinkOutput struct {
	Body CurrentUser
}

// OAuthStartInput begins a Google OAuth flow.
type OAuthStartInput struct {
	Body struct {
		RedirectURI string `json:"redirectUri" format:"uri" doc:"The allowed OAuth redirect URI to return to"`
	}
}

// OAuthStartOutput returns the provider authorization URL.
type OAuthStartOutput struct {
	Body struct {
		URL string `json:"url" format:"uri" doc:"The Google OAuth authorization URL to redirect the user to"`
	}
}

// OAuthCallbackInput receives the OAuth provider callback.
type OAuthCallbackInput struct {
	Code  string `query:"code" doc:"The authorization code from the OAuth provider"`
	State string `query:"state" doc:"The state parameter returned by the OAuth provider"`
	Error string `query:"error" doc:"An error returned by the OAuth provider"`
}

// OAuthCallbackOutput redirects the browser to the authenticated application.
type OAuthCallbackOutput struct {
	Status int
	// Location is set by the handler.
}

// LogoutInput logs out the current session.
type LogoutInput struct {
	// XCSRFToken is the double-submit CSRF token matching the cookie.
	XCSRFToken string `header:"X-CSRF-Token" doc:"Double-submit CSRF token"`
}

// LogoutOutput is an empty successful response.
type LogoutOutput struct{}

// RegisterAuth registers the T01 magic-link, T02 OAuth, session, and logout routes.
func RegisterAuth(api huma.API, svc *auth.Service) {
	huma.Register(api, huma.Operation{
		OperationID: "RequestMagicLink",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/magic-links",
		Summary:     "Request a magic sign-in link",
		Tags:        []string{"Authentication"},
		Responses: map[string]*huma.Response{
			"429": {Description: "Rate limited"},
			"503": {Description: "Magic link sign-in is disabled"},
		},
	}, func(ctx context.Context, input *RequestMagicLinkInput) (*RequestMagicLinkOutput, error) {
		c := authHumaContext(ctx)
		if err := svc.RequestMagicLink(ctx, clientIPFromHuma(c), input.Body.Email); err != nil {
			return nil, mapAuthError(err)
		}
		return &RequestMagicLinkOutput{}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "ConsumeMagicLink",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/magic-links/consume",
		Summary:     "Consume a magic sign-in link",
		Tags:        []string{"Authentication"},
		Responses: map[string]*huma.Response{
			"401": {Description: "Invalid or expired magic link, or the account is disabled"},
			"429": {Description: "Rate limited"},
			"503": {Description: "Magic link sign-in, or new sign-ups, is disabled"},
		},
	}, func(ctx context.Context, input *ConsumeMagicLinkInput) (*ConsumeMagicLinkOutput, error) {
		c := authHumaContext(ctx)
		user, session, token, err := svc.ConsumeMagicLink(ctx, clientIPFromHuma(c), input.Body.Token, input.Body.Email)
		if err != nil {
			return nil, mapAuthError(err)
		}
		// Issue session cookie and double-submit CSRF cookie for later
		// authenticated requests (e.g., logout).
		cookie := svc.SessionCookie(token, session.ExpiresAt)
		c.AppendHeader("Set-Cookie", cookie.String())
		if _, csrfCookie := svc.IssueCSRFCookie(); csrfCookie != nil {
			c.AppendHeader("Set-Cookie", csrfCookie.String())
		}
		return &ConsumeMagicLinkOutput{Body: currentUserFromAuth(user)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "OAuthStart",
		Method:        http.MethodPost,
		Path:          "/api/v1/auth/oauth/google/start",
		Summary:       "Start a Google OAuth sign-in flow",
		Tags:          []string{"Authentication"},
		DefaultStatus: 200,
		Responses: map[string]*huma.Response{
			"401": {Description: "The redirect URI is not allowed"},
			"404": {Description: "OAuth provider not configured"},
			"429": {Description: "Rate limited"},
			"503": {Description: "Google OAuth sign-in is disabled"},
		},
	}, func(ctx context.Context, input *OAuthStartInput) (*OAuthStartOutput, error) {
		c := authHumaContext(ctx)
		url, stateToken, err := svc.OAuthStart(ctx, clientIPFromHuma(c), input.Body.RedirectURI)
		if err != nil {
			return nil, mapAuthError(err)
		}
		stateCookie := svc.OAuthStateCookie(stateToken, svc.Clock().Now().Add(svc.OAuthStateLifetime()))
		c.AppendHeader("Set-Cookie", stateCookie.String())
		return &OAuthStartOutput{Body: struct {
			URL string `json:"url" format:"uri" doc:"The Google OAuth authorization URL to redirect the user to"`
		}{URL: url}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "OAuthCallback",
		Method:      http.MethodGet,
		Path:        "/api/v1/auth/oauth/google/callback",
		Summary:     "Handle the Google OAuth provider callback",
		Tags:        []string{"Authentication"},
		Responses: map[string]*huma.Response{
			"302": {Description: "Redirect to the authenticated application"},
			"401": {Description: "Invalid or expired OAuth state, or the provider/account failed verification"},
			"404": {Description: "OAuth provider not configured"},
			"429": {Description: "Rate limited"},
			"503": {Description: "Google OAuth sign-in, or new sign-ups, is disabled"},
		},
	}, func(ctx context.Context, input *OAuthCallbackInput) (*OAuthCallbackOutput, error) {
		c := authHumaContext(ctx)
		if input.Error != "" {
			return nil, mapAuthError(auth.ErrOAuthProviderFailed)
		}
		cookieState := oauthStateCookieValue(c, svc.OAuthStateCookieName())
		_, session, token, returnURL, err := svc.OAuthCallback(ctx, clientIPFromHuma(c), input.Code, input.State, cookieState)
		if err != nil {
			return nil, mapAuthError(err)
		}
		c.AppendHeader("Set-Cookie", svc.SessionCookie(token, session.ExpiresAt).String())
		if _, csrfCookie := svc.IssueCSRFCookie(); csrfCookie != nil {
			c.AppendHeader("Set-Cookie", csrfCookie.String())
		}
		c.AppendHeader("Set-Cookie", svc.ClearOAuthStateCookie().String())
		c.AppendHeader("Location", returnURL)
		return &OAuthCallbackOutput{Status: http.StatusFound}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "Logout",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/logout",
		Summary:     "Log out the current session",
		Tags:        []string{"Authentication"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(svc)},
		Responses: map[string]*huma.Response{
			"401": {Description: "Authentication is required"},
			"403": {Description: "Invalid CSRF token"},
			"429": {Description: "Rate limited"},
		},
	}, func(ctx context.Context, input *LogoutInput) (*LogoutOutput, error) {
		c := authHumaContext(ctx)
		sessionCookie, err := huma.ReadCookie(c, svc.SessionCookieName())
		token := ""
		if err == nil {
			token = sessionCookie.Value
		}
		if err := svc.Logout(ctx, token); err != nil {
			return nil, mapAuthError(err)
		}
		cookie := svc.ClearSessionCookie()
		c.AppendHeader("Set-Cookie", cookie.String())
		c.AppendHeader("Set-Cookie", svc.ClearCSRFCookie().String())
		return &LogoutOutput{}, nil
	})
}

func currentUserFromAuth(u *auth.User) CurrentUser {
	cu := CurrentUser{
		Email:           &u.Email,
		EmailVerifiedAt: u.EmailVerifiedAt,
	}
	if u.DisplayName != "" {
		cu.DisplayName = &u.DisplayName
	}
	if u.AvatarURL != "" {
		cu.AvatarURL = &u.AvatarURL
	}
	return cu
}

func mapAuthError(err error) huma.StatusError {
	switch {
	case err == nil:
		return nil
	case err == auth.ErrInvalidMagicLink:
		return huma.Error401Unauthorized("invalid or expired magic link")
	case err == auth.ErrAuthenticationRequired:
		return huma.Error401Unauthorized("authentication required")
	case err == auth.ErrUserDisabled:
		return huma.Error401Unauthorized("authentication required")
	case err == auth.ErrRateLimited:
		return huma.Error429TooManyRequests("rate limited")
	case err == auth.ErrInvalidOAuthState:
		return huma.Error401Unauthorized("invalid or expired oauth state")
	case err == auth.ErrOAuthProviderFailed:
		return huma.Error401Unauthorized("oauth provider failed")
	case err == auth.ErrOAuthNotConfigured:
		return huma.Error404NotFound("oauth provider not configured")
	case err == auth.ErrMagicLinkDisabled:
		return huma.Error503ServiceUnavailable("magic link sign-in is disabled")
	case err == auth.ErrOAuthDisabled:
		return huma.Error503ServiceUnavailable("google oauth sign-in is disabled")
	case err == auth.ErrSignupsDisabled:
		return huma.Error503ServiceUnavailable("new sign-ups are disabled")
	default:
		return huma.Error500InternalServerError("internal error")
	}
}
