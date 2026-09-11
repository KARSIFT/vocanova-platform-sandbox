import { normalizeReturnTo } from "./return-to";

const STORAGE_KEY = "vocanova:oauth-continuation";
const CONTINUATION_TTL_MS = 15 * 60 * 1000;
const CONTINUATION_PATH_PREFIXES = [
  "/onboarding",
  "/home",
  "/discover",
  "/words",
  "/review",
  "/reviews",
  "/progress",
  "/settings",
] as const;

interface OAuthContinuation {
  createdAt: number;
  returnTo: string;
}

function getSessionStorage(): Storage | null {
  if (typeof window === "undefined") {
    return null;
  }
  try {
    return window.sessionStorage;
  } catch {
    return null;
  }
}

/**
 * Google OAuth is configured to return to a small server allowlist. Preserve a
 * safe route in this browser tab so the authenticated Home entry point can
 * continue to the learner's original destination after the callback.
 */
export function rememberOAuthContinuation(returnTo: string): void {
  const storage = getSessionStorage();
  if (!storage) {
    return;
  }
  const continuation: OAuthContinuation = {
    createdAt: Date.now(),
    returnTo: normalizeOAuthContinuation(returnTo),
  };
  try {
    storage.setItem(STORAGE_KEY, JSON.stringify(continuation));
  } catch {
    // Sign-in itself must continue when browser storage is unavailable.
  }
}

export function clearOAuthContinuation(): void {
  const storage = getSessionStorage();
  if (!storage) {
    return;
  }
  try {
    storage.removeItem(STORAGE_KEY);
  } catch {
    // Storage cleanup must never block an auth action.
  }
}

export function consumeOAuthContinuation(): string | null {
  const storage = getSessionStorage();
  if (!storage) {
    return null;
  }

  let raw: string | null;
  try {
    raw = storage.getItem(STORAGE_KEY);
    storage.removeItem(STORAGE_KEY);
  } catch {
    return null;
  }
  if (!raw) {
    return null;
  }

  try {
    const continuation = JSON.parse(raw) as OAuthContinuation;
    if (
      !Number.isFinite(continuation.createdAt) ||
      continuation.createdAt > Date.now() ||
      Date.now() - continuation.createdAt > CONTINUATION_TTL_MS
    ) {
      return null;
    }
    return normalizeOAuthContinuation(continuation.returnTo);
  } catch {
    return null;
  }
}

function normalizeOAuthContinuation(value: string): string {
  const returnTo = normalizeReturnTo(value);
  const pathname = new URL(returnTo, "https://vocanova.invalid").pathname;
  return CONTINUATION_PATH_PREFIXES.some(
    (prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`),
  )
    ? returnTo
    : "/home";
}
