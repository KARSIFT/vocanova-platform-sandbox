import assert from "node:assert/strict";
import { it } from "node:test";

import { isDisplayNameWithinLimit } from "../../src/lib/display-name";

it("uses the API's 80 Unicode code-point display-name limit", () => {
  for (const character of ["a", "𐐀", "😀"]) {
    assert.equal(isDisplayNameWithinLimit(character.repeat(80)), true);
    assert.equal(isDisplayNameWithinLimit(character.repeat(81)), false);
  }
  assert.equal(isDisplayNameWithinLimit(""), true);
});
