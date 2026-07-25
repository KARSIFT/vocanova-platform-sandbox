import assert from "node:assert/strict";
import test from "node:test";

import { validateMockInventory } from "./mock-inventory.mjs";

test("VOC-026 P1 mocks and the authorized VOC-027-T00 boundary are respected", () => {
  assert.deepEqual(validateMockInventory(), []);
});
