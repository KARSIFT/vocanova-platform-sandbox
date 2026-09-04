import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { getSituationDetailView } from "../../src/app/(app)/discover/[situation]/_components/situation-view";

describe("discover situation detail page view selection (VOC-1179)", () => {
  it("shows the empty state when a situation has zero meanings", () => {
    assert.equal(getSituationDetailView(0), "empty");
  });

  it("shows the meanings list when at least one meaning is available", () => {
    assert.equal(getSituationDetailView(1), "list");
    assert.equal(getSituationDetailView(8), "list");
  });
});
