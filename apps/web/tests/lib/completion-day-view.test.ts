import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { getCompletionDayView } from "../../src/app/(app)/progress/completion-day-view";

describe("progress mission-day presentation", () => {
  const day = { localDate: "2026-09-30", completed: true };

  it("distinguishes protected streak credit from an actually completed mission", () => {
    const completed = getCompletionDayView({ ...day, status: "completed" });
    const protectedDay = getCompletionDayView({ ...day, status: "protected" });
    assert.equal(completed.label, "Completed");
    assert.equal(protectedDay.label, "Streak protected");
    assert.notEqual(protectedDay.className, completed.className);
  });

  it("uses explicit status even when the legacy boolean disagrees", () => {
    for (const status of ["open", "missed"] as const) {
      assert.equal(
        getCompletionDayView({ ...day, status }).label,
        "Not complete",
      );
    }
    assert.equal(
      getCompletionDayView({ ...day, completed: false, status: "completed" })
        .label,
      "Completed",
    );
    assert.equal(
      getCompletionDayView({ ...day, completed: false, status: "protected" })
        .label,
      "Streak protected",
    );
  });

  it("does not invent a precise outcome when an older API omits status", () => {
    assert.equal(getCompletionDayView(day).label, "Completed or protected");
    assert.equal(
      getCompletionDayView({ ...day, completed: false }).label,
      "Not complete",
    );
  });
});
