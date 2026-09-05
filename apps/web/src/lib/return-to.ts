const APP_ORIGIN = "https://vocanova.invalid";

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
      destination.pathname.startsWith("//")
    ) {
      return "/home";
    }
    return `${destination.pathname}${destination.search}${destination.hash}`;
  } catch {
    return "/home";
  }
}
