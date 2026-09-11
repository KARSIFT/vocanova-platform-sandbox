const APP_ORIGIN = "https://vocanova.invalid";

// A completed sign-in must always take the learner somewhere useful. Sending
// them back to a public auth endpoint creates a confusing loop after a valid
// magic link is consumed. The email-change confirmation route is deliberately
// absent: it is a protected continuation that can require a fresh sign-in.
const PUBLIC_AUTH_PATHS = new Set([
  "/login",
  "/signin",
  "/magic-link",
  "/auth/magic",
]);

/**
 * Limits post-auth navigation to an app-relative route. Keeping this check in
 * the web app is defense in depth for destinations carried in a magic-link
 * URL, which can be edited before the learner opens it.
 */
export function normalizeReturnTo(value?: string | null): string {
  if (!value || typeof value !== "string") {
    return "/home";
  }

  const relative = value.trim();
  if (
    !relative.startsWith("/") ||
    relative.startsWith("//") ||
    relative.includes("\\")
  ) {
    return "/home";
  }

  try {
    const destination = new URL(relative, APP_ORIGIN);
    // Dot-segment normalization can produce //host even though the input
    // was a same-origin path. Never return a network-path reference.
    if (
      destination.origin !== APP_ORIGIN ||
      destination.pathname.startsWith("//") ||
      PUBLIC_AUTH_PATHS.has(destination.pathname)
    ) {
      return "/home";
    }
    return `${destination.pathname}${destination.search}${destination.hash}`;
  } catch {
    return "/home";
  }
}
