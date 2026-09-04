import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { getDiscoverListView } from "../../src/app/(app)/discover/_components/discover-view";

describe("discover list page view selection (VOC-1179)", () => {
  it("shows the empty state when there are zero situations", () => {
    assert.equal(getDiscoverListView(0), "empty");
  });

  it("shows the situations list when at least one situation is available", () => {
    assert.equal(getDiscoverListView(1), "list");
    assert.equal(getDiscoverListView(12), "list");
  });
});
