import assert from "node:assert/strict";
import { afterEach, describe, it } from "node:test";

import {
  authCookieNamesFor,
  filterForwardedCookieHeader,
  getCookieValue,
  normalizeCookieEnvironment,
} from "../../src/lib/cookies";

const originalDocument = globalThis.document;

function installCookies(cookie: string): void {
  Object.defineProperty(globalThis, "document", {
    configurable: true,
    value: { cookie },
  });
}

afterEach(() => {
  Object.defineProperty(globalThis, "document", {
    configurable: true,
    value: originalDocument,
  });
});

describe("environment cookie names", () => {
  it("uses tier-specific names for the deployed environments", () => {
    assert.deepEqual(authCookieNamesFor("staging"), {
      session: "vocanova_staging_session",
      csrf: "vocanova_staging_csrf",
    });
    assert.deepEqual(authCookieNamesFor("production"), {
      session: "vocanova_production_session",
      csrf: "vocanova_production_csrf",
    });
  });

  it("keeps the legacy names for local, test and unknown environments", () => {
    for (const value of [undefined, "", "local", "test", "prod", "staging-2"]) {
      assert.equal(normalizeCookieEnvironment(value), "unknown");
    }
    assert.deepEqual(authCookieNamesFor("unknown"), {
      session: "vocanova_session",
      csrf: "vocanova_csrf",
    });
  });

  it("normalizes case and surrounding whitespace like the API", () => {
    assert.equal(normalizeCookieEnvironment(" Production\n"), "production");
    assert.equal(normalizeCookieEnvironment("STAGING"), "staging");
  });
});

describe("filterForwardedCookieHeader", () => {
  const everyTier = [
    "vocanova_theme=dark",
    "vocanova_production_session=prod-session==",
    "  vocanova_staging_session=stage-session  ",
    "vocanova_session=legacy-session",
    "vocanova_production_csrf=prod-csrf",
    "vocanova_staging_csrf=stage-csrf",
    "vocanova_csrf=legacy-csrf",
    "vocanova_oauth_state=state",
    "tracking=1",
  ].join(";");

  it("keeps only production credentials and the theme in production", () => {
    assert.equal(
      filterForwardedCookieHeader(everyTier, "production"),
      "vocanova_theme=dark; vocanova_production_session=prod-session==; vocanova_production_csrf=prod-csrf",
    );
  });

  it("keeps only staging credentials and the theme in staging", () => {
    assert.equal(
      filterForwardedCookieHeader(everyTier, "staging"),
      "vocanova_theme=dark; vocanova_staging_session=stage-session; vocanova_staging_csrf=stage-csrf",
    );
  });

  it("preserves local mock context only for unknown environments", () => {
    assert.equal(
      filterForwardedCookieHeader(everyTier, "unknown"),
      everyTier,
    );
  });

  it("matches names exactly, never by prefix, suffix or case", () => {
    const lookalikes =
      "xvocanova_staging_session=a; vocanova_staging_session_old=b; Vocanova_Staging_Session=c; vocanova_staging_csrf";
    assert.equal(filterForwardedCookieHeader(lookalikes, "staging"), "");
  });

  it("drops empty values and keeps the first non-empty duplicate", () => {
    assert.equal(
      filterForwardedCookieHeader(
        "vocanova_staging_session=; vocanova_staging_session=first; vocanova_staging_session=second",
        "staging",
      ),
      "vocanova_staging_session=first",
    );
  });

  it("returns an empty header when no allowed cookie is present", () => {
    assert.equal(filterForwardedCookieHeader("", "production"), "");
    assert.equal(
      filterForwardedCookieHeader(
        "vocanova_session=legacy; vocanova_staging_session=stage",
        "production",
      ),
      "",
    );
  });
});

describe("getCookieValue", () => {
  it("returns null outside the browser", () => {
    Object.defineProperty(globalThis, "document", {
      configurable: true,
      value: undefined,
    });
    assert.equal(getCookieValue("vocanova_csrf"), null);
  });

  it("finds a cookie regardless of position and whitespace", () => {
    installCookies("vocanova_theme=dark;vocanova_staging_csrf=stage  ;  other=1");
    assert.equal(getCookieValue("vocanova_staging_csrf"), "stage");
    installCookies("  vocanova_staging_csrf=first; vocanova_theme=dark");
    assert.equal(getCookieValue("vocanova_staging_csrf"), "first");
  });

  it("ignores foreign tiers and look-alike names", () => {
    installCookies(
      "vocanova_production_csrf=prod; xvocanova_staging_csrf=x; vocanova_staging_csrf_old=old",
    );
    assert.equal(getCookieValue("vocanova_staging_csrf"), null);
    installCookies("vocanovaXcsrf=regex-lookalike");
    assert.equal(getCookieValue("vocanova.csrf"), null);
  });

  it("decodes percent escapes and treats malformed ones as missing", () => {
    installCookies("vocanova_csrf=a%2Bb%3D");
    assert.equal(getCookieValue("vocanova_csrf"), "a+b=");
    for (const malformed of ["%E0%A4%A", "%", "abc%zz"]) {
      installCookies(`vocanova_theme=dark; vocanova_csrf=${malformed}`);
      assert.equal(getCookieValue("vocanova_csrf"), null);
    }
  });

  it("skips empty values", () => {
    installCookies("vocanova_csrf=; vocanova_csrf=later");
    assert.equal(getCookieValue("vocanova_csrf"), "later");
    installCookies("vocanova_csrf=");
    assert.equal(getCookieValue("vocanova_csrf"), null);
  });
});
