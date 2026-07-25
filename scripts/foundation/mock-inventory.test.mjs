import assert from "node:assert/strict";
import test from "node:test";

import { validateMockInventory } from "./mock-inventory.mjs";

test("A1 mock inventory is retained and no P1-P4 APIs are invented", () => {
  assert.deepEqual(validateMockInventory(), []);
});
