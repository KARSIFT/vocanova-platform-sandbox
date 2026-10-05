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
