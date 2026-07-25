import assert from "node:assert/strict";
import test from "node:test";

import { validateMockInventory } from "./mock-inventory.mjs";

test("mock inventory is retained and APIs do not exceed VOC-026-T01", () => {
  assert.deepEqual(validateMockInventory(), []);
});
