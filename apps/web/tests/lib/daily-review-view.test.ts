import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { draftReviewDate, dailyReviewView, relativeReviewDate, reviewGoalDayChanged } from "../../src/app/(app)/reviews/_components/daily-review-view";

describe("authoritative daily review view", () => {
  it("uses a target of 15 without replacing effort with successful answers", () => {
    const goal = { localDate: "2026-11-15", timeZone: "America/New_York", target: 15, reviewed: 15, remaining: 0, complete: true };
    const view = dailyReviewView({ goal, dueCount: 8, session: { status: "completed" } });
    assert.equal(view.complete, true);
    assert.equal(view.active, false);
    assert.equal(view.reviewed, 15);
    assert.equal(view.target, 15);
    assert.equal(view.canReviewExtra, true);
    assert.equal(view.discoveryIsNext, true);
  });
  it("does not treat a truthy completed session as active", () => {
    const view = dailyReviewView({ dueCount: 0, session: { status: "completed" } });
    assert.equal(view.active, false);
    assert.equal(view.discoveryIsNext, true);
  });
  it("distinguishes all available words finished from goal completion", () => {
    const view = dailyReviewView({ dueCount: 0, goal: { localDate: "2026-11-15", timeZone: "America/New_York", target: 15, reviewed: 4, remaining: 11, complete: false } });
    assert.equal(view.complete, false);
    assert.equal(view.allAvailableFinished, true);
    assert.equal(view.reviewed, 4);
  });
  it("preserves legacy queue behavior when goal is absent", () => {
    assert.equal(dailyReviewView({ dueCount: 2 }).discoveryIsNext, false);
    assert.equal(dailyReviewView({ dueCount: 2, session: { status: "in_progress" } }).active, true);
  });
});
describe("relative review schedule in learner timezone", () => {
  const label = (instant: string, now: string, zone = "America/New_York") => relativeReviewDate(instant, zone, new Date(now));
  it("uses calendar tomorrow across the spring DST transition", () => {
    assert.equal(label("2026-03-08T13:00:00Z", "2026-03-07T14:00:00Z")?.label, "Tomorrow");
  });
  it("uses calendar tomorrow across the autumn DST transition", () => {
    assert.equal(label("2026-11-01T14:00:00Z", "2026-10-31T13:00:00Z")?.label, "Tomorrow");
  });
  it("uses the learner zone instead of UTC dates", () => {
    assert.equal(label("2026-11-16T01:00:00Z", "2026-11-15T23:00:00Z")?.label, "Today");
    assert.equal(label("2026-11-16T01:00:00Z", "2026-11-15T23:00:00Z", "Asia/Tehran")?.label, "Today");
  });
  it("shows weeks and calendar months with an exact accessible date", () => {
    assert.equal(label("2026-11-29T14:00:00Z", "2026-11-15T14:00:00Z")?.label, "In 2 weeks");
    assert.equal(label("2026-12-15T14:00:00Z", "2026-11-15T14:00:00Z")?.label, "In 1 month");
    assert.match(label("2026-12-15T14:00:00Z", "2026-11-15T14:00:00Z")!.exact, /December 15, 2026/);
  });
  it("does not invent a date without a valid timestamp and timezone", () => {
    assert.equal(relativeReviewDate("invalid", "UTC"), undefined);
    assert.equal(relativeReviewDate("2026-11-16T00:00:00Z", undefined), undefined);
    assert.equal(relativeReviewDate("2026-11-16T00:00:00Z", "not/a/zone"), undefined);
  });
});

describe("overnight goal freshness", () => {
  it("refreshes at learner midnight, not UTC midnight", () => {
    const goal = { localDate: "2026-11-15", timeZone: "America/New_York" };
    assert.equal(reviewGoalDayChanged(goal, new Date("2026-11-16T04:59:00Z")), false);
    assert.equal(reviewGoalDayChanged(goal, new Date("2026-11-16T05:00:00Z")), true);
  });
  it("ignores malformed legacy dates and zones", () => {
    assert.equal(reviewGoalDayChanged({ localDate: "bad", timeZone: "UTC" }), false);
    assert.equal(reviewGoalDayChanged({ localDate: "2026-11-15", timeZone: "invalid" }), false);
  });
});


describe("draft schedule preserves elapsed durations across DST", () => {
  it("labels spring 1d as two calendar dates later with an exact date", () => {
    const draft = draftReviewDate("1d", "America/New_York", new Date("2026-03-08T04:30:00Z"));
    assert.equal(draft?.label, "In 2 days");
    assert.match(draft!.exact, /March 9, 2026/);
  });
  it("labels autumn 1d as today with an exact date", () => {
    const draft = draftReviewDate("1d", "America/New_York", new Date("2026-11-01T04:30:00Z"));
    assert.equal(draft?.label, "Today");
    assert.match(draft!.exact, /November 1, 2026/);
  });
  it("does not invent a calendar label without a learner timezone", () => {
    assert.equal(draftReviewDate("1d", undefined), undefined);
  });
});
