import assert from "node:assert/strict";
import test from "node:test";

import { validateMockInventory } from "./mock-inventory.mjs";

// VOC-031-T03: the protected-boundary allow list now includes the
// T03 email-change routes (`/api/v1/settings/email-change-links`),
// the `accounts` business module, the `email_change_links` Ent
// schema, and the `email_change_links` migration, in addition to
// the previously adopted A1/P1/P2/P4-T00/P5-T01/T02 boundaries.
test("VOC-031-T03 mock inventory accepts the email-change backend boundary", () => {
  assert.deepEqual(validateMockInventory(), []);
});

// VOC-030-T06: the T06 cross-cutting safety test files are present
// and the P5-forbidden invariant holds across the T06 deliverable.
// This test is the same code path as the one above; it is
// separately named to make the T06 acceptance criterion visible in
// the test runner output (the production script's success message
// already prints "VOC-030-T06 mock inventory validation passed").
test("VOC-030-T06 mock inventory accepts the T06 cross-cutting test files", () => {
  assert.deepEqual(validateMockInventory(), []);
});
