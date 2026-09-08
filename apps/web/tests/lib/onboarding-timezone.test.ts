import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { getBrowserTimezone } from "../../src/app/onboarding/_components/onboarding-timezone";

describe("onboarding browser timezone", () => {
  it("returns the IANA timezone exposed by the browser runtime", () => {
    assert.equal(typeof getBrowserTimezone(), "string");
  });
});
