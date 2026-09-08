/**
 * Returns the browser's IANA timezone when the runtime exposes one. Some
 * embedded or privacy-restricted clients do not, so callers intentionally
 * fall back to the API's documented UTC behavior instead of guessing.
 */
export function getBrowserTimezone(): string | undefined {
  try {
    const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
    return timezone || undefined;
  } catch {
    return undefined;
  }
}
