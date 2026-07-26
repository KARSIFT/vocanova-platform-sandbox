import assert from "node:assert/strict";
import test from "node:test";

import { validateMockInventory } from "./mock-inventory.mjs";

test("VOC-030-T00 mock dispositions and protected boundaries are respected", () => {
  assert.deepEqual(validateMockInventory(), []);
});
