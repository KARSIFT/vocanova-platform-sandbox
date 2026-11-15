import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { reviewedTodayEvidence, reviewedTodayIncrease } from "../staging-e2e/review-progress";
const mission = {localDate:"2026-11-15",reviewTarget:15,reviewsCompleted:15};
const goal = {localDate:"2026-11-15",target:15,reviewed:18};
describe("strict staging raw effort evidence", () => {
 it("renders raw 18 of 15 while validating capped mission credit of 15", () => {
  assert.deepEqual(reviewedTodayEvidence(mission,goal),{localDate:"2026-11-15",reviewed:18,target:15,homeLabel:"Reviewed today: 18 of 15"});
 });
 it("requires fresh 2 to increase raw 18 to at least 20 even after goal completion", () => {
  const before=reviewedTodayEvidence(mission,goal);
  const unchanged=reviewedTodayEvidence(mission,goal);
  assert.deepEqual(reviewedTodayIncrease(before,unchanged,2),{minimumExpected:20,reached:false});
  const after=reviewedTodayEvidence(mission,{...goal,reviewed:20});
  assert.deepEqual(reviewedTodayIncrease(before,after,2),{minimumExpected:20,reached:true});
 });
 it("rejects inconsistent mission caps and goal targets", () => {
  assert.throws(()=>reviewedTodayEvidence({...mission,reviewsCompleted:18},goal),/mission/);
  assert.throws(()=>reviewedTodayEvidence({...mission,reviewTarget:20},goal),/target/);
 });
 it("requires both reads and before/after comparisons to share a local day", () => {
  assert.throws(()=>reviewedTodayEvidence(mission,{...goal,localDate:"2026-11-16"}),/date/);
  const before=reviewedTodayEvidence(mission,goal);
  assert.throws(()=>reviewedTodayIncrease(before,{...before,localDate:"2026-11-16",reviewed:20},2),/date/);
 });
 it("rejects missing goal evidence and vacuous zero-card gates", () => {
  assert.throws(()=>reviewedTodayEvidence(mission,undefined),/goal/);
  const before=reviewedTodayEvidence(mission,goal);
  assert.throws(()=>reviewedTodayIncrease(before,before,0),/fresh/);
 });
});
