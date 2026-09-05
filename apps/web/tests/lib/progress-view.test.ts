import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { getSavedVocabularySummary } from "../../src/app/(app)/progress/_components/progress-view";

describe("progress saved vocabulary summary", () => {
  it("labels the bounded recent-word sample without claiming a total", () => {
    assert.equal(
      getSavedVocabularySummary(10),
      "Showing up to 10 recently saved words.",
    );
  });
});
