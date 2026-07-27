import assert from "node:assert/strict";
import test from "node:test";

import { validateMockInventory } from "./mock-inventory.mjs";

// VOC-030-T04 baseline: the existing allow list (T00–T04
// boundaries, T05 mock retirements, no-P5) is enforced.
test("VOC-030-T04 mock dispositions and protected boundaries are respected", () => {
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
