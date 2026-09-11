import assert from "node:assert/strict";
import { describe, it } from "node:test";

import {
  getThemePreferenceFromCookie,
  isThemePreference,
  resolveTheme,
} from "../../src/lib/theme-preference";

describe("theme preference", () => {
  it("defaults System to the active operating-system scheme", () => {
    assert.equal(resolveTheme("system", false), "light");
    assert.equal(resolveTheme("system", true), "dark");
    assert.equal(resolveTheme("light", true), "light");
    assert.equal(resolveTheme("dark", false), "dark");
  });

  it("accepts only supported persisted values", () => {
    assert.equal(isThemePreference("light"), true);
    assert.equal(isThemePreference("dark"), true);
    assert.equal(isThemePreference("system"), true);
    assert.equal(isThemePreference("sepia"), false);
    assert.equal(isThemePreference(undefined), false);
  });

  it("reads a valid theme cookie without trusting malformed values", () => {
    assert.equal(
      getThemePreferenceFromCookie("other=value; vocanova_theme=dark"),
      "dark",
    );
    assert.equal(
      getThemePreferenceFromCookie("vocanova_theme=system"),
      "system",
    );
    assert.equal(getThemePreferenceFromCookie("vocanova_theme=sepia"), null);
    assert.equal(getThemePreferenceFromCookie("vocanova_theme=%E0%A4%A"), null);
  });
});
