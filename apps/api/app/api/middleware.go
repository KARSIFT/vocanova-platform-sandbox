package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/danielgtaylor/huma/v2"
)

// AuthMiddleware validates the session cookie and injects the requester into
// the context. It never fails on its own; routes that require authentication
// must add RequireAuth.
func AuthMiddleware(svc *auth.Service) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		// Authentication responses may establish, inspect, or invalidate a
		// browser session. Keep them out of browser and intermediary caches
		// even when a handler returns an error before its normal response.
		if isSensitiveAuthPath(ctx.URL().Path) {
			ctx.SetHeader("Cache-Control", "no-store")
			ctx.SetHeader("Pragma", "no-cache")
		}
		token := sessionCookieValue(ctx, svc.SessionCookieName())
		if token != "" {
			if user, err := svc.ValidateSession(ctx.Context(), token); err == nil {
				ctx = huma.WithValue(ctx, requesterKey{}, user)
				// A session outlives the browser-session CSRF cookie. Let an
				// authenticated browser refresh that cookie through /me before it
				// makes a protected write, without accepting a missing token on the
				// write itself.
				if ctx.Method() == http.MethodGet && ctx.URL().Path == "/api/v1/me" && csrfCookieValue(ctx, svc.CSRFCookieName()) == "" {
					if _, csrfCookie := svc.IssueCSRFCookie(); csrfCookie != nil {
						ctx.AppendHeader("Set-Cookie", csrfCookie.String())
					}
				}
			}
		}
		next(ctx)
	}
}

func isSensitiveAuthPath(path string) bool {
	return path == "/api/v1/me" || strings.HasPrefix(path, "/api/v1/auth/")
}

// RequireAuth aborts the request with 401 when no authenticated requester was
// set by AuthMiddleware.
func RequireAuth() func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		if Requester(ctx.Context()) == nil {
			writeHumaError(ctx, huma.Error401Unauthorized("authentication required"))
			return
		}
		next(ctx)
	}
}

// CSRFMiddleware enforces the double-submit CSRF token for unsafe (state-changing)
// HTTP methods. Safe methods (GET, HEAD, OPTIONS) and cookie credentials are
// skipped.
func CSRFMiddleware(svc *auth.Service) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		if !isUnsafeMethod(ctx.Method()) {
			next(ctx)
			return
		}
		cookie := csrfCookieValue(ctx, svc.CSRFCookieName())
		header := ctx.Header("X-CSRF-Token")
		if !svc.ValidateCSRF(cookie, header) {
			writeHumaError(ctx, huma.Error403Forbidden("invalid csrf token"))
			return
		}
		next(ctx)
	}
}

// sessionCookieValue reads the session bearer cookie from the request.
func sessionCookieValue(c huma.Context, name string) string {
	ck, err := huma.ReadCookie(c, name)
	if err != nil {
		return ""
	}
	return ck.Value
}

func isUnsafeMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}

func writeHumaError(ctx huma.Context, err huma.StatusError) {
	status := err.GetStatus()
	ctx.SetStatus(status)
	ctx.SetHeader("Content-Type", "application/problem+json")
	em := huma.ErrorModel{
		Type:   "about:blank",
		Title:  http.StatusText(status),
		Status: status,
		Detail: err.Error(),
	}
	_ = json.NewEncoder(ctx.BodyWriter()).Encode(em)
}
