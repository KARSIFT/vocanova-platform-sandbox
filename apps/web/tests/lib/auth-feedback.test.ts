import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { ApiResponseError } from "@vocanova/api-client";

import {
  getAuthErrorMessage,
  getOAuthCallbackMessage,
} from "../../src/lib/auth-feedback";

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
});
