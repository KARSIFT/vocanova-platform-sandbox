import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { getBuildIdentity } from "../../src/lib/build-identity";

describe("build identity", () => {
  it("preserves valid immutable release metadata", () => {
    assert.deepEqual(
      getBuildIdentity({
        version: "0.2.0-rc.1",
        commit: "0123456789abcdef0123456789abcdef01234567",
        environment: "staging",
        builtAt: "2026-09-12T13:45:00+03:30",
      }),
      {
        version: "0.2.0-rc.1",
        commit: "0123456789abcdef0123456789abcdef01234567",
        environment: "staging",
        builtAt: "2026-09-12T10:15:00.000Z",
      },
    );
  });

  it("uses unknown instead of guessing local or malformed metadata", () => {
    assert.deepEqual(
      getBuildIdentity({
        version: `1.0.0-${"a".repeat(60)}`,
        commit: "0123456",
        environment: "preview",
        builtAt: "tomorrow",
      }),
      {
        version: "unknown",
        commit: "unknown",
        environment: "unknown",
        builtAt: "unknown",
      },
    );
  });

  it("labels an explicitly injected local build as development", () => {
    assert.equal(
      getBuildIdentity({ environment: "development" }).environment,
      "development",
    );
  });

  it("accepts the API contract's test environment", () => {
    assert.equal(getBuildIdentity({ environment: "test" }).environment, "test");
  });
});
