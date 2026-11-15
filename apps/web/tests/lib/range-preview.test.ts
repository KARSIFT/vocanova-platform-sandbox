import assert from "node:assert/strict";
import test from "node:test";
import { rangeProgressLabel, rangePreviewPaths } from "../../src/app/(app)/knowledge-map/range-preview-model";
const band={number:1,startRank:1,endRank:100,totalWords:100,knownCount:1,learningCount:2,unseenCount:97,wordStates:"klm"+"n".repeat(97)};
test("miniature per-word states agree with readable known, learning, learned and new counts",()=>{
 assert.equal(rangeProgressLabel(band," · ","already know"),"1 already know · 1 learning · 1 learned · 97 new");
 assert.deepEqual(rangePreviewPaths(band),{k:"M1 1h8v8h-8z",l:"M11 1h8v8h-8z",m:"M21 1h8v8h-8z"});
});
test("legacy aggregates remain honest without inventing per-word states",()=>{
 const legacy={...band,wordStates:undefined};
 assert.equal(rangeProgressLabel(legacy," · ","already know"),"1 already know · 2 saved · 97 new");
 assert.equal(rangePreviewPaths(legacy),null);
 assert.deepEqual(rangePreviewPaths({...legacy,knownCount:0,learningCount:0,unseenCount:100}),{k:"",l:"",m:""});
});
