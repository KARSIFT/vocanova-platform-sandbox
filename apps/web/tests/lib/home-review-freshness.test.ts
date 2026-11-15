import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { homeReviewSnapshot } from "../../src/app/(app)/home/_components/home-review-snapshot";

const goal = { localDate: "2026-11-15", timeZone: "Pacific/Honolulu", target: 5, reviewed: 5, remaining: 0, complete: true };
describe("Home review freshness snapshot", () => {
  it("detects a server-confirmed new day without deriving credit", () => {
    const before = homeReviewSnapshot({ goal, dueCount: 0, session: null });
    const after = homeReviewSnapshot({ goal: { ...goal, localDate: "2026-11-16", reviewed: 0, remaining: 5, complete: false }, dueCount: 5, session: null });
    assert.notEqual(before, after);
    assert.equal(homeReviewSnapshot({ goal: { ...goal }, dueCount: 0, session: null }), before);
  });
  it("refreshes when confirmed progress, availability or active-session status changes", () => {
    const before = homeReviewSnapshot({ goal, dueCount: 0, session: { id: "lesson", status: "completed" } });
    assert.notEqual(homeReviewSnapshot({ goal, dueCount: 1, session: { id: "lesson", status: "completed" } }), before);
    assert.notEqual(homeReviewSnapshot({ goal, dueCount: 0, session: { id: "lesson", status: "in_progress" } }), before);
    assert.notEqual(homeReviewSnapshot({ goal: { ...goal, reviewed: 4, complete: false, remaining: 1 }, dueCount: 0, session: { id: "lesson", status: "completed" } }), before);
  });
});
