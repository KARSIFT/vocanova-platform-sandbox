import assert from "node:assert/strict";
import test from "node:test";
import { explicitDiscoverySelection, resumableDiscoveryPosition } from "../../src/app/(app)/discovery/discovery-position";
const data = { catalogVersion: "english-10000-v1", bandSize:100, currentBand:{number:44,items:[{rank:4330,saved:true,selfReportedKnown:true}]}} as Parameters<typeof resumableDiscoveryPosition>[1];
const saved = {catalogVersion:data.catalogVersion,band:44,rank:4330,active:true,revision:7};
test("generic discovery resumes 4330 exactly even when known or saved",()=>{
 assert.equal(explicitDiscoverySelection({}),false);
 assert.deepEqual(resumableDiscoveryPosition(saved,data),saved);
});
test("explicit valid range/word selection wins over preference",()=>{
 assert.equal(explicitDiscoverySelection({band:"12",word:"1150"}),true);
 for(const band of ["0","101","wat","1.1"])assert.equal(explicitDiscoverySelection({band}),false);
});
test("missing, changed catalog, unavailable rank or wrong band cannot resume",()=>{
 assert.equal(resumableDiscoveryPosition(null,data),null);
 for(const p of [{...saved,catalogVersion:"old"},{...saved,rank:4331},{...saved,band:43},{...saved,rank:4330.5}])assert.equal(resumableDiscoveryPosition(p,data),null);
});
