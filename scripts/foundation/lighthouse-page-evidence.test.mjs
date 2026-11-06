import assert from "node:assert/strict";
import { test } from "node:test";

import { assertAuditedPage } from "../../apps/web/tests/lighthouse/assertions.mjs";

const requestedUrl = "http://127.0.0.1:3000/home";
function report(url = requestedUrl, statusCode = 200) {
  return {
    finalDisplayedUrl: url,
    audits: {
      "network-requests": {
        details: { items: [{ resourceType: "Document", url, statusCode }] },
      },
    },
  };
}

test("learner performance evidence requires the requested successful page", () => {
  assert.doesNotThrow(() =>
    assertAuditedPage({ requestedUrl, report: report() }),
  );
  for (const url of ["/login?returnTo=%2Fhome", "/onboarding"]) {
    assert.throws(
      () =>
        assertAuditedPage({
          requestedUrl,
          report: report(`http://127.0.0.1:3000${url}`),
        }),
      /instead of/,
    );
  }
  assert.throws(
    () =>
      assertAuditedPage({ requestedUrl, report: report(requestedUrl, 503) }),
    /returned 503/,
  );
  assert.throws(
    () => assertAuditedPage({ requestedUrl, report: {} }),
    /unknown URL/,
  );
  assert.throws(
    () =>
      assertAuditedPage({
        requestedUrl,
        report: {
          ...report(),
          runtimeError: { code: "FAILED_DOCUMENT_REQUEST" },
        },
      }),
    /could not load/,
  );
});

// A screen-specific heading must not accept a successful HTTP200 auth or error
// shell. These checks prevent broad heading patterns from laundering such pages
// into performance evidence; the runner verifies the actual browser document.
test("full learner audit expectations reject generic auth and error shells", async () => {
  const { FULL_SCREENS } =
    await import("../../apps/web/tests/lighthouse/full-screens.mjs");
  const shells = [
    "Sign in to Vocanova",
    "Something went wrong",
    "We couldn't connect",
    "We couldn't load your account",
    "Your story could not load",
    "Writing practice could not load",
    "We couldn't load your sentence history",
  ];
  for (const screen of FULL_SCREENS.filter((screen) => screen.authenticated)) {
    for (const heading of shells) {
      const accepted =
        typeof screen.heading === "string"
          ? screen.heading === heading
          : screen.heading.test(heading);
      assert.equal(accepted, false, `${screen.name} must reject ${heading}`);
    }
    // A changed final route must fail even when its document succeeds.
    const url = `http://127.0.0.1:3100${screen.path}`;
    assert.throws(
      () =>
        assertAuditedPage({
          requestedUrl: url,
          report: report("http://127.0.0.1:3100/login"),
        }),
      /instead of/,
    );
  }
});
