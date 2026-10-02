import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { ApiResponseError } from "@vocanova/api-client";

import {
  getAuthErrorMessage,
  getOAuthCallbackMessage,
} from "../../src/lib/auth-feedback";

const googleOnly = {
  magicLinkEnabled: false,
  oauthEnabled: true,
  passwordEnabled: false,
};
const passwordOnly = {
  magicLinkEnabled: false,
  oauthEnabled: false,
  passwordEnabled: true,
};
const magicLinkOnly = {
  magicLinkEnabled: true,
  oauthEnabled: false,
  passwordEnabled: false,
};
const noMethods = {
  magicLinkEnabled: false,
  oauthEnabled: false,
  passwordEnabled: false,
};

describe("authentication feedback", () => {
  it("maps rate limits to useful magic-link guidance", () => {
    assert.equal(
      getAuthErrorMessage(new ApiResponseError(429, null), "magic-request"),
      "Too many sign-in links were requested. Please wait a few minutes, then try again.",
    );
  });

  it("does not display raw magic-link API messages", () => {
    assert.equal(
      getAuthErrorMessage(
        new ApiResponseError(401, null, "invalid or expired magic link"),
        "magic-consume",
      ),
      "This sign-in link is invalid, expired, or has already been used. Request a new link.",
    );
  });

  it("distinguishes unavailable and rate-limited password sign-in", () => {
    assert.equal(
      getAuthErrorMessage(new ApiResponseError(503, null), "password-login"),
      "Password sign-in is temporarily unavailable. Please try again later.",
    );
    assert.equal(
      getAuthErrorMessage(new ApiResponseError(429, null), "password-login"),
      "Too many password sign-in attempts were made. Please wait a few minutes, then try again.",
    );
  });

  it("explains each supported OAuth callback outcome", () => {
    assert.match(getOAuthCallbackMessage("cancelled") ?? "", /cancelled/);
    assert.match(getOAuthCallbackMessage("expired") ?? "", /expired/);
    assert.match(getOAuthCallbackMessage("unavailable") ?? "", /unavailable/);
    assert.match(getOAuthCallbackMessage("failed") ?? "", /could not complete/);
    assert.equal(getOAuthCallbackMessage("unexpected"), null);
  });

  for (const status of [404, 503, 500]) {
    it(`does not suggest disabled email methods after a Google-only start failure (${status})`, () => {
      const message = getAuthErrorMessage(
        new ApiResponseError(status, null),
        "oauth-start",
        googleOnly,
      );
      assert.match(message, /Google sign-in/);
      assert.match(message, /try again/i);
      assert.doesNotMatch(message, /email|password|sign-in link/i);
    });
  }

  for (const outcome of ["cancelled", "unavailable", "failed"]) {
    it(`does not suggest disabled email methods after a Google-only callback (${outcome})`, () => {
      const message = getOAuthCallbackMessage(outcome, googleOnly) ?? "";
      assert.match(message, /Google/);
      assert.match(message, /try again/i);
      assert.doesNotMatch(message, /email|password|sign-in link/i);
    });
  }

  it("uses retry guidance when sign-in capabilities are unknown or all disabled", () => {
    for (const capabilities of [undefined, noMethods]) {
      assert.doesNotMatch(
        getAuthErrorMessage(
          new ApiResponseError(503, null),
          "oauth-start",
          capabilities,
        ),
        /use email|email instead|continue with Google|sign-in link/i,
      );
      assert.doesNotMatch(
        getOAuthCallbackMessage("cancelled", capabilities) ?? "",
        /use email|email instead|continue with Google|sign-in link/i,
      );
    }
  });

  it("names email and password only when password sign-in is available", () => {
    for (const message of [
      getAuthErrorMessage(
        new ApiResponseError(503, null),
        "oauth-start",
        passwordOnly,
      ),
      getOAuthCallbackMessage("failed", passwordOnly) ?? "",
    ]) {
      assert.match(message, /email and password/i);
      assert.doesNotMatch(message, /sign-in link/i);
    }
  });

  it("names a sign-in link only when magic-link sign-in is available", () => {
    for (const message of [
      getAuthErrorMessage(new Error("network error"), "oauth-start", magicLinkOnly),
      getOAuthCallbackMessage("cancelled", magicLinkOnly) ?? "",
    ]) {
      assert.match(message, /sign-in link/i);
      assert.doesNotMatch(message, /password/i);
    }
  });

  it("keeps rate-limit guidance without advertising another disabled method", () => {
    const message = getAuthErrorMessage(
      new ApiResponseError(429, null),
      "oauth-start",
      googleOnly,
    );
    assert.match(message, /wait a few minutes/);
    assert.doesNotMatch(message, /email|password|sign-in link/i);
  });

  for (const status of [404, 503, 500]) {
    it(`suggests creating an account after a Google signup failure when its password form is available (${status})`, () => {
      const message = getAuthErrorMessage(
        new ApiResponseError(status, null, "private-provider-diagnostic"),
        "oauth-start",
        { ...passwordOnly, oauthEnabled: true },
        "signup",
      );
      assert.match(message, /try again/i);
      assert.match(message, /create (?:an|your) account with (?:your )?email and password/i);
      assert.doesNotMatch(message, /sign in with|sign-in link|private-provider-diagnostic/i);
    });

    it(`does not advertise the absent email-link form after a Google signup failure (${status})`, () => {
      const message = getAuthErrorMessage(
        new ApiResponseError(status, null, "private-provider-diagnostic"),
        "oauth-start",
        { ...magicLinkOnly, oauthEnabled: true },
        "signup",
      );
      assert.match(message, /try again/i);
      assert.doesNotMatch(message, /email|password|sign-in link|private-provider-diagnostic/i);
    });
  }

  it("keeps Google signup rate-limit guidance without offering an immediate alternative", () => {
    const message = getAuthErrorMessage(
      new ApiResponseError(429, null),
      "oauth-start",
      { ...passwordOnly, oauthEnabled: true },
      "signup",
    );
    assert.match(message, /wait a few minutes/);
    assert.doesNotMatch(message, /email|password|sign-in link/i);
  });
});
