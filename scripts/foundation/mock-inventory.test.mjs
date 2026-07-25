import assert from "node:assert/strict";
import test from "node:test";

import { validateMockInventory } from "./mock-inventory.mjs";

test("VOC-026 P1 mock inventory and API/schema boundary are respected", () => {
  assert.deepEqual(validateMockInventory(), []);
});
