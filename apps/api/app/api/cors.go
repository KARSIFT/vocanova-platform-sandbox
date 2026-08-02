package api

import (
	"net/http"
	"net/url"
	"strings"
)

// corsAllowedOrigins derives the set of browser origins (scheme://host[:port],
// no path) allowed to make cross-origin requests against this API, from the
// already-founder-approved OAuth return URLs (OAUTH_REDIRECT_ALLOWLIST /
// OAUTH_REDIRECT_URI). Those URLs are the web app's known trusted hostnames -
// reusing them here avoids introducing a second, independently-configurable
// trust boundary that could silently drift from the first. Malformed entries
// are skipped rather than causing a startup failure: an origin-parsing
// mistake in one entry should not take down CORS (and therefore the whole
// interactive app) for every other, correctly-formed entry.
func corsAllowedOrigins(returnURLs []string) []string {
	seen := make(map[string]bool)
	var origins []string
	for _, raw := range returnURLs {
		u, err := url.Parse(raw)
		if err != nil || u.Scheme == "" || u.Host == "" {
			continue
		}
		origin := u.Scheme + "://" + u.Host
		if !seen[origin] {
			seen[origin] = true
			origins = append(origins, origin)
		}
	}
	return origins
}

// corsMiddleware answers CORS preflight (OPTIONS) requests and annotates
// every response with the appropriate Access-Control-* headers, restricted
// to an explicit origin allowlist (never a wildcard: this API relies on
// credentialed, cookie-based sessions, and the CORS spec forbids combining
// a wildcard Access-Control-Allow-Origin with Access-Control-Allow-Credentials
// anyway - browsers reject that combination outright).
//
// docs/engineering/06-backend-design.md §16 requires a "CORS allowlist" as
// part of this API's baseline security posture; this was never implemented,
// which meant every cross-origin browser request from the deployed web app
// (a different subdomain than the API, per VOC-037-D00) failed its preflight
// with no Access-Control-Allow-Origin header - discovered live on 2026-08-02
// while verifying the GOOGLE_OAUTH_ENABLED kill switch against production
// with real credentials: the browser's OAuth-start preflight received a bare
// 405 with no CORS headers at all, and every other credentialed cross-origin
// POST (magic-link request, sentence submission, review submission, etc.)
// has the identical failure mode.
func corsMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			isAllowed := origin != "" && allowed[origin]

			if isAllowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
			}

			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				if isAllowed {
					reqHeaders := r.Header.Get("Access-Control-Request-Headers")
					if reqHeaders == "" {
						reqHeaders = "Content-Type, Authorization"
					}
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
					w.Header().Set("Access-Control-Max-Age", "600")
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// corsOriginsSummary is a small logging helper so a misconfigured (empty)
// allowlist is visible in the startup log rather than silently breaking
// every cross-origin request with no diagnostic trail.
func corsOriginsSummary(origins []string) string {
	if len(origins) == 0 {
		return "(none - all cross-origin requests will be rejected)"
	}
	return strings.Join(origins, ", ")
}
