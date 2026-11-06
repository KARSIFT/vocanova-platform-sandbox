import { createApiClient } from "./api";
import { CSRF_COOKIE_NAME, getCookieValue } from "./cookies";

let recoveryInFlight: Promise<string | null> | null = null;

/**
 * Restores a missing browser-session CSRF cookie for an already authenticated
 * learner. The API only issues it after validating the HttpOnly session on
 * GET /me; protected writes remain protected by the double-submit check.
 */
export async function getOrRefreshCSRFToken(): Promise<string | null> {
  const existing = getCookieValue(CSRF_COOKIE_NAME);
  if (existing) {
    return existing;
  }

  if (recoveryInFlight) {
    return recoveryInFlight;
  }

  recoveryInFlight = createApiClient()
    .getCurrentUser()
    .then(async () => {
      const refreshed = getCookieValue(CSRF_COOKIE_NAME);
      if (refreshed) return refreshed;

      // WebKit can finish /me before invalidating document.cookie's earlier
      // missing-cookie snapshot. Give the browser one task to expose Set-Cookie;
      // an absent cookie after that still means the session is not ready.
      await new Promise<void>((resolve) => setTimeout(resolve, 0));
      return getCookieValue(CSRF_COOKIE_NAME);
    })
    .finally(() => {
      recoveryInFlight = null;
    });
  return recoveryInFlight;
}
