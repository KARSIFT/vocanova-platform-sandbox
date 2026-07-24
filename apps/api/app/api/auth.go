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

// LogoutInput logs out the current session.
type LogoutInput struct {
	// XCSRFToken is the double-submit CSRF token matching the cookie.
	XCSRFToken string `header:"X-CSRF-Token" doc:"Double-submit CSRF token"`
}

// LogoutOutput is an empty successful response.
type LogoutOutput struct{}

// RegisterAuth registers the T01 magic-link, session, and logout routes.
func RegisterAuth(api huma.API, svc *auth.Service) {
	huma.Register(api, huma.Operation{
		OperationID: "RequestMagicLink",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/magic-links",
		Summary:     "Request a magic sign-in link",
		Tags:        []string{"Authentication"},
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
		OperationID: "Logout",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/logout",
		Summary:     "Log out the current session",
		Tags:        []string{"Authentication"},
	}, func(ctx context.Context, input *LogoutInput) (*LogoutOutput, error) {
		c := authHumaContext(ctx)
		if !svc.ValidateCSRF(csrfCookieValue(c, svc.CSRFCookieName()), c.Header("X-CSRF-Token")) {
			return nil, huma.Error403Forbidden("invalid csrf token")
		}
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
		return &LogoutOutput{}, nil
	})
}

func currentUserFromAuth(u *auth.User) CurrentUser {
	return CurrentUser{
		Email:           &u.Email,
		EmailVerifiedAt: u.EmailVerifiedAt,
	}
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
	default:
		return huma.Error500InternalServerError("internal error")
	}
}
