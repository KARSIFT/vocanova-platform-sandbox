import assert from "node:assert/strict";
import { test } from "node:test";

import {
  sanitizeTelemetryBreadcrumb,
  sanitizeTelemetryEvent,
} from "../../src/lib/telemetry-privacy";

test("error telemetry excludes proof queries, password bodies and session headers", () => {
  const result = sanitizeTelemetryEvent({
    event_id: "safe-event-id",
    request: {
      url: "https://example.test/auth/password/reset?token=secret-proof#fragment",
      method: "POST",
      query_string: "token=secret-proof",
      data: { password: "secret-password" },
      cookies: { session: "secret-session" },
      headers: { Authorization: "Bearer secret-session" },
    },
  });
  assert.deepEqual(result.request, {
    url: "https://example.test/auth/password/reset",
    method: "POST",
  });
  assert.equal(result.event_id, "safe-event-id");
  assert.doesNotMatch(JSON.stringify(result), /secret-/);
});

test("navigation and fetch breadcrumbs cannot retain proof URLs", () => {
  const result = sanitizeTelemetryBreadcrumb({
    category: "navigation",
    data: {
      from: "/auth/password/verify?token=secret-proof",
      to: "/login?returnTo=%2Freviews",
      url: "https://example.test/auth/password/reset?token=secret-proof",
      status_code: 204,
    },
  });
  assert.doesNotMatch(JSON.stringify(result), /secret-proof|returnTo/);
  assert.equal(result.data?.status_code, 204);
});
