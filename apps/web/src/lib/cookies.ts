export const SESSION_COOKIE_NAME = "vocanova_session";
export const CSRF_COOKIE_NAME = "vocanova_csrf";
export const OAUTH_STATE_COOKIE_NAME = "vocanova_oauth_state";

export function getCookieValue(name: string): string | null {
  if (typeof document === "undefined") {
    return null;
  }
  const match = document.cookie.match(new RegExp(`(?:^|;\\s*)${name}=([^;]+)`));
  if (match?.[1] === undefined) {
    return null;
  }
  return decodeURIComponent(match[1]);
}

export function deleteCookie(name: string, domain?: string, path = "/") {
  if (typeof document === "undefined") {
    return;
  }
  const domainPart = domain ? `; domain=${domain}` : "";
  document.cookie = `${name}=; Max-Age=0; path=${path}${domainPart}`;
}
