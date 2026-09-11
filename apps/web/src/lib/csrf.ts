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
    .then(() => getCookieValue(CSRF_COOKIE_NAME))
    .finally(() => {
      recoveryInFlight = null;
    });
  return recoveryInFlight;
}
