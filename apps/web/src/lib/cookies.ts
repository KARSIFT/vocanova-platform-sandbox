import { THEME_COOKIE_NAME } from "./theme-preference";

export const LEGACY_SESSION_COOKIE_NAME = "vocanova_session";
export const LEGACY_CSRF_COOKIE_NAME = "vocanova_csrf";
export const OAUTH_STATE_COOKIE_NAME = "vocanova_oauth_state";

export type CookieEnvironment = "staging" | "production" | "unknown";

export interface AuthCookieNames {
  session: string;
  csrf: string;
}

// Staging and production share the parent cookie domain, so each deployed
// tier needs its own session/CSRF names. Keep in sync with the API's
// auth.EnvironmentCookieNames and the nginx Cookie allowlists.
const DEPLOYED_AUTH_COOKIE_NAMES: Record<
  Exclude<CookieEnvironment, "unknown">,
  AuthCookieNames
> = {
  staging: {
    session: "vocanova_staging_session",
    csrf: "vocanova_staging_csrf",
  },
  production: {
    session: "vocanova_production_session",
    csrf: "vocanova_production_csrf",
  },
};

export function normalizeCookieEnvironment(
  value: string | undefined,
): CookieEnvironment {
  const normalized = value?.trim().toLowerCase();
  return normalized === "staging" || normalized === "production"
    ? normalized
    : "unknown";
}

export function authCookieNamesFor(
  environment: CookieEnvironment,
): AuthCookieNames {
  if (environment === "unknown") {
    return {
      session: LEGACY_SESSION_COOKIE_NAME,
      csrf: LEGACY_CSRF_COOKIE_NAME,
    };
  }
  return DEPLOYED_AUTH_COOKIE_NAMES[environment];
}

/**
 * The deployed tier, from the existing build-time variable. Keep the literal
 * process.env.NEXT_PUBLIC_VOCANOVA_BUILD_ENVIRONMENT access so Next.js
 * inlines it into client bundles; local and test builds leave it unset.
 */
export function buildCookieEnvironment(): CookieEnvironment {
  return normalizeCookieEnvironment(
    process.env.NEXT_PUBLIC_VOCANOVA_BUILD_ENVIRONMENT,
  );
}

export const SESSION_COOKIE_NAME = authCookieNamesFor(
  buildCookieEnvironment(),
).session;
export const CSRF_COOKIE_NAME = authCookieNamesFor(
  buildCookieEnvironment(),
).csrf;

function cookiePairs(cookieHeader: string): Array<[string, string]> {
  const pairs: Array<[string, string]> = [];
  for (const segment of cookieHeader.split(";")) {
    const separator = segment.indexOf("=");
    const name = segment.slice(0, Math.max(separator, 0)).trim();
    if (separator < 0 || name === "") {
      continue;
    }
    pairs.push([name, segment.slice(separator + 1).trim()]);
  }
  return pairs;
}

/**
 * Builds the Cookie header the web server forwards to the API. Deployed tiers
 * retain only their session/CSRF cookies and theme, matched by exact name;
 * foreign/legacy names and empty values are dropped. The first non-empty
 * value of a name wins. Local/test builds retain their original fixture context.
 */
export function filterForwardedCookieHeader(
  cookieHeader: string,
  environment: CookieEnvironment = buildCookieEnvironment(),
): string {
  // Preserve local development and mock-fixture request context. Deployed
  // builds select an explicit tier and always apply the strict allowlist.
  if (environment === "unknown") return cookieHeader;
  const names = authCookieNamesFor(environment);
  const allowed = new Set([names.session, names.csrf, THEME_COOKIE_NAME]);
  const forwarded = new Map<string, string>();
  for (const [name, value] of cookiePairs(cookieHeader)) {
    if (allowed.has(name) && value !== "" && !forwarded.has(name)) {
      forwarded.set(name, value);
    }
  }
  return [...forwarded].map(([name, value]) => `${name}=${value}`).join("; ");
}

export function getCookieValue(name: string): string | null {
  if (typeof document === "undefined") {
    return null;
  }
  const value = cookiePairs(document.cookie).find(
    ([cookieName, cookieValue]) => cookieName === name && cookieValue !== "",
  )?.[1];
  if (value === undefined) {
    return null;
  }
  // A malformed percent escape must not throw out of callers such as the
  // CSRF recovery path; treat it as a missing cookie.
  try {
    return decodeURIComponent(value);
  } catch {
    return null;
  }
}

export function deleteCookie(name: string, domain?: string, path = "/") {
  if (typeof document === "undefined") {
    return;
  }
  const domainPart = domain ? `; domain=${domain}` : "";
  document.cookie = `${name}=; Max-Age=0; path=${path}${domainPart}`;
}
