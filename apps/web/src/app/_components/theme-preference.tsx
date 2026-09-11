"use client";

import type { ReactNode } from "react";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";

import {
  getThemeBootstrapScript,
  THEME_COOKIE_NAME,
  THEME_PREFERENCES,
  THEME_STORAGE_KEY,
  type ThemePreference,
  isThemePreference,
  resolveTheme,
} from "@/lib/theme-preference";

const ThemePreferenceContext = createContext<{
  preference: ThemePreference;
  setPreference: (preference: ThemePreference) => void;
} | null>(null);

function getDocumentPreference(): ThemePreference {
  if (typeof document === "undefined") {
    return "system";
  }
  const preference = document.documentElement.dataset.themePreference;
  return isThemePreference(preference) ? preference : "system";
}

function applyTheme(preference: ThemePreference): void {
  const root = document.documentElement;
  const resolved = resolveTheme(
    preference,
    window.matchMedia("(prefers-color-scheme: dark)").matches,
  );
  root.dataset.theme = resolved;
  root.dataset.themePreference = preference;
  root.style.colorScheme = resolved;
}

function persistPreference(preference: ThemePreference): void {
  try {
    window.localStorage.setItem(THEME_STORAGE_KEY, preference);
  } catch {
    // Appearance preferences are optional when browser storage is unavailable.
  }
  try {
    document.cookie = `${THEME_COOKIE_NAME}=${encodeURIComponent(preference)}; Path=/; Max-Age=31536000; SameSite=Lax`;
  } catch {
    // The cookie only improves first paint on later visits.
  }
}

export function ThemeBootstrap() {
  return (
    <script dangerouslySetInnerHTML={{ __html: getThemeBootstrapScript() }} />
  );
}

export function ThemeProvider({ children }: Readonly<{ children: ReactNode }>) {
  // The server renders System. The bootstrap script can already have applied a
  // saved choice before hydration, so wait to synchronize React state until
  // after hydration rather than overwriting that first paint with System.
  const [preference, setPreferenceState] = useState<ThemePreference>("system");
  const [isHydrated, setIsHydrated] = useState(false);

  useEffect(() => {
    setPreferenceState(getDocumentPreference());
    setIsHydrated(true);
  }, []);

  useEffect(() => {
    if (!isHydrated) {
      return;
    }

    applyTheme(preference);
    if (preference !== "system") {
      return;
    }

    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const updateSystemTheme = () => applyTheme("system");
    mediaQuery.addEventListener("change", updateSystemTheme);
    return () => mediaQuery.removeEventListener("change", updateSystemTheme);
  }, [isHydrated, preference]);

  const setPreference = useCallback((nextPreference: ThemePreference) => {
    persistPreference(nextPreference);
    applyTheme(nextPreference);
    setPreferenceState(nextPreference);
  }, []);

  return (
    <ThemePreferenceContext.Provider value={{ preference, setPreference }}>
      {children}
    </ThemePreferenceContext.Provider>
  );
}

export function useThemePreference() {
  const context = useContext(ThemePreferenceContext);
  if (!context) {
    throw new Error("useThemePreference must be used within ThemeProvider.");
  }
  return context;
}

const preferenceCopy: Record<ThemePreference, string> = {
  light: "Light",
  dark: "Dark",
  system: "System",
};

export function ThemePreferenceControl() {
  const { preference, setPreference } = useThemePreference();

  return (
    <section
      aria-labelledby="appearance-heading"
      className="mt-[var(--spacing-lg)] rounded-[var(--radius-lg)] border border-neutral-200 bg-white p-[var(--spacing-md)] shadow-sm sm:p-[var(--spacing-lg)]"
    >
      <h2
        id="appearance-heading"
        className="text-lg font-semibold text-neutral-900"
      >
        Appearance
      </h2>
      <p
        id="appearance-description"
        className="mt-[var(--spacing-xs)] text-base text-neutral-700"
      >
        Choose how VocaNova looks on this device.
      </p>
      <fieldset
        className="mt-[var(--spacing-md)]"
        aria-describedby="appearance-description"
      >
        <legend className="text-sm font-semibold text-neutral-900">
          Theme
        </legend>
        <div className="mt-[var(--spacing-sm)] grid gap-[var(--spacing-sm)] sm:grid-cols-3">
          {THEME_PREFERENCES.map((option) => (
            <label
              key={option}
              className="flex min-h-11 cursor-pointer items-center gap-[var(--spacing-sm)] rounded-md border border-neutral-300 bg-neutral-50 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 transition-colors hover:bg-neutral-100 has-[:checked]:border-primary-600 has-[:checked]:bg-primary-50"
            >
              <input
                type="radio"
                name="theme-preference"
                value={option}
                checked={preference === option}
                onChange={() => setPreference(option)}
                className="size-4 accent-primary-600"
              />
              {preferenceCopy[option]}
            </label>
          ))}
        </div>
      </fieldset>
      <p className="mt-[var(--spacing-sm)] text-sm text-neutral-600">
        System follows your device setting.
      </p>
    </section>
  );
}
