import assert from "node:assert/strict";
import test from "node:test";
import { DiscoveryPositionWriter } from "../../src/app/(app)/discovery/discovery-position-writer";
const position={catalogVersion:"english-10000-v1",band:44,rank:4330,active:true};
test("position writes coalesce and preserve CAS revision across navigation",async()=>{
 const sent: Array<{rank:number;expectedRevision:number}>=[];
 const writer=new DiscoveryPositionWriter(0,{read:async()=>0,write:async(value)=>{sent.push(value);return value.expectedRevision+1},status:()=>{},conflict:()=>false});
 writer.select(position);writer.select({...position,rank:4331});await writer.flush();
 writer.select({...position,rank:4332});await writer.flush();writer.dispose();
 assert.deepEqual(sent.map(({rank,expectedRevision})=>({rank,expectedRevision})),[{rank:4331,expectedRevision:0},{rank:4332,expectedRevision:1}]);
});
test("a stale tab never retries automatically with a newly read revision",async()=>{
 let calls=0,reads=0;
 const writer=new DiscoveryPositionWriter(0,{read:async()=>{reads++;return 9},write:async()=>{calls++;throw new Error("stale")},status:()=>{},conflict:()=>true});
 writer.select(position);await writer.flush();writer.select({...position,rank:4331});await writer.flush();
 assert.equal(calls,1);assert.equal(reads,0);
 await writer.retry();assert.equal(reads,1);assert.equal(calls,2);writer.dispose();
});
test("a disposed account cannot issue a write after a late preference read",async()=>{
 let release!:(value:number)=>void;let writes=0;
 const read=new Promise<number>((resolve)=>{release=resolve});
 const writer=new DiscoveryPositionWriter(null,{read:()=>read,write:async()=>{writes++;return 1},status:()=>{},conflict:()=>false});
 writer.select(position);const pending=writer.flush();writer.dispose();release(0);await pending;
 assert.equal(writes,0);
});
