import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { getReviewsView } from "../../src/app/(app)/reviews/_components/reviews-view";

describe("reviews page view selection (VOC-1179)", () => {
  it("shows the empty state when there are zero due words", () => {
    assert.equal(getReviewsView(0), "empty");
  });

  it("shows the review session when at least one word is due", () => {
    assert.equal(getReviewsView(1), "session");
    assert.equal(getReviewsView(50), "session");
  });
});
