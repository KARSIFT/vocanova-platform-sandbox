export const THEME_COOKIE_NAME = "vocanova_theme";
export const THEME_STORAGE_KEY = "vocanova:theme-preference";

export const THEME_PREFERENCES = ["light", "dark", "system"] as const;

export type ThemePreference = (typeof THEME_PREFERENCES)[number];
export type ResolvedTheme = Exclude<ThemePreference, "system">;

export function isThemePreference(value: unknown): value is ThemePreference {
  return (
    typeof value === "string" &&
    THEME_PREFERENCES.includes(value as ThemePreference)
  );
}

export function resolveTheme(
  preference: ThemePreference,
  systemPrefersDark: boolean,
): ResolvedTheme {
  if (preference === "system") {
    return systemPrefersDark ? "dark" : "light";
  }
  return preference;
}

export function getThemePreferenceFromCookie(
  cookieValue: string,
): ThemePreference | null {
  const encodedName = `${THEME_COOKIE_NAME}=`;
  const cookie = cookieValue
    .split(";")
    .map((entry) => entry.trim())
    .find((entry) => entry.startsWith(encodedName));
  if (!cookie) {
    return null;
  }

  try {
    const value = decodeURIComponent(cookie.slice(encodedName.length));
    return isThemePreference(value) ? value : null;
  } catch {
    return null;
  }
}

/** Runs in the document head before the application hydrates. */
export function getThemeBootstrapScript(): string {
  return `(() => {
  const storageKey = ${JSON.stringify(THEME_STORAGE_KEY)};
  const cookieName = ${JSON.stringify(THEME_COOKIE_NAME)};
  const allowed = ["light", "dark", "system"];
  let preference;
  try { preference = window.localStorage.getItem(storageKey); } catch {}
  if (!allowed.includes(preference)) {
    try {
      const prefix = cookieName + "=";
      const cookie = document.cookie.split(";").map((entry) => entry.trim()).find((entry) => entry.startsWith(prefix));
      preference = cookie ? decodeURIComponent(cookie.slice(prefix.length)) : undefined;
    } catch {}
  }
  if (!allowed.includes(preference)) preference = "system";
  const resolved = preference === "system"
    ? (window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light")
    : preference;
  const root = document.documentElement;
  root.dataset.theme = resolved;
  root.dataset.themePreference = preference;
  root.style.colorScheme = resolved;
})();`;
}
