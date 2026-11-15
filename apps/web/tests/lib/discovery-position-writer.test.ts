import assert from "node:assert/strict";
import test from "node:test";
import { DiscoveryPositionWriter } from "../../src/app/(app)/discovery/discovery-position-writer";
const position={catalogVersion:"english-10000-v1",band:44,rank:4330,active:true};
test("position writes coalesce and preserve CAS revision across navigation",async()=>{
 const sent: Array<{rank:number;expectedRevision:number}>=[];
 const writer=new DiscoveryPositionWriter("owner-a", 0,{read:async()=>0,write:async(value)=>{sent.push(value);return value.expectedRevision+1},status:()=>{},accountChanged:()=>false,conflict:()=>false});
 writer.select(position);writer.select({...position,rank:4331});await writer.flush();
 writer.select({...position,rank:4332});await writer.flush();writer.dispose();
 assert.deepEqual(sent.map(({rank,expectedRevision})=>({rank,expectedRevision})),[{rank:4331,expectedRevision:0},{rank:4332,expectedRevision:1}]);
});
test("a stale tab never retries automatically with a newly read revision",async()=>{
 let calls=0,reads=0;
 const writer=new DiscoveryPositionWriter("owner-a", 0,{read:async()=>{reads++;return 9},write:async()=>{calls++;throw new Error("stale")},status:()=>{},accountChanged:()=>false,conflict:()=>true});
 writer.select(position);await writer.flush();writer.select({...position,rank:4331});await writer.flush();
 assert.equal(calls,1);assert.equal(reads,0);
 await writer.retry();assert.equal(reads,1);assert.equal(calls,2);writer.dispose();
});
test("a disposed account cannot issue a write after a late preference read",async()=>{
 let release!:(value:number)=>void;let writes=0;
 const read=new Promise<number>((resolve)=>{release=resolve});
 const writer=new DiscoveryPositionWriter("owner-a", null,{read:()=>read,write:async()=>{writes++;return 1},status:()=>{},accountChanged:()=>false,conflict:()=>false});
 writer.select(position);const pending=writer.flush();writer.dispose();release(0);await pending;
 assert.equal(writes,0);
});

for (const failure of ["error", "conflict"] as const) {
 test(`saved A, failed B, A then C retains ${failure} until explicit retry confirms C`,async()=>{
  const sent:Array<{rank:number;expectedRevision:number}>=[];
  const statuses:string[]=[];
  let fail=false,reads=0;
  const writer=new DiscoveryPositionWriter("owner-a", 0,{
   read:async()=>{reads++;return 8;},
   write:async(value)=>{sent.push(value);if(fail)throw new Error(failure);return value.expectedRevision+1;},
   status:(status)=>statuses.push(status),accountChanged:()=>false,conflict:()=>failure==="conflict",
  });
  try {
   writer.select(position);await writer.flush();assert.equal(statuses.at(-1),"saved");
   fail=true;writer.select({...position,rank:4331});await writer.flush();assert.equal(statuses.at(-1),failure);
   writer.select(position);
   fail=false;writer.select({...position,rank:4332});await writer.flush();
   assert.equal(statuses.at(-1),failure,"retry must remain reachable for the unconfirmed current word");
   assert.deepEqual(sent.map(({rank})=>rank),[4330,4331],"blocked navigation must not overwrite automatically");
   assert.equal(reads,0,"conflicts must not acquire a newer revision automatically");
   await writer.retry();
   assert.equal(reads,1);assert.deepEqual(sent.at(-1),{expectedAccountId:"owner-a",...position,rank:4332,expectedRevision:8});
   assert.equal(statuses.at(-1),"saved");
  } finally {writer.dispose();}
 });
}

test("returning to confirmed A while B is in flight stays saving until A is confirmed again",async()=>{
 const sent:number[]=[],statuses:string[]=[];
 let release!:(value:number)=>void;
 const pendingB=new Promise<number>((resolve)=>{release=resolve;});
 const writer=new DiscoveryPositionWriter("owner-a", 0,{
  read:async()=>0,
  write:async(value)=>{sent.push(value.rank);if(value.rank===4331)return pendingB;return value.expectedRevision+1;},
  status:(status)=>statuses.push(status),accountChanged:()=>false,conflict:()=>false,
 });
 try {
  writer.select(position);await writer.flush();
  writer.select({...position,rank:4331});const pending=writer.flush();
  writer.select(position);
  assert.equal(statuses.at(-1),"saving","old confirmation cannot race a different in-flight write");
  release(2);await pending;
  assert.deepEqual(sent,[4330,4331,4330]);assert.equal(statuses.at(-1),"saved");
 } finally {release(2);writer.dispose();}
});

test("explicit retry re-confirms A after an uncertain B even if A was confirmed before",async()=>{
 const sent:Array<{rank:number;expectedRevision:number}>=[];
 let fail=false;
 const writer=new DiscoveryPositionWriter("owner-a", 0,{
  read:async()=>9,
  write:async(value)=>{sent.push(value);if(fail)throw new Error("response lost after commit");return value.expectedRevision+1;},
  status:()=>{},accountChanged:()=>false,conflict:()=>false,
 });
 try {
  writer.select(position);await writer.flush();
  fail=true;writer.select({...position,rank:4331});await writer.flush();
  writer.select(position);fail=false;await writer.retry();
  assert.deepEqual(sent.at(-1),{expectedAccountId:"owner-a",...position,expectedRevision:9},"fresh revision alone cannot prove an old local confirmation is still the server position");
 } finally {writer.dispose();}
});


test("owner fence survives unknown initial revision and retry; account change is terminal", async () => {
 const sent: Array<{expectedAccountId:string; expectedRevision:number}> = [], statuses:string[]=[];
 let fail = true, reads = 0;
 const writer = new DiscoveryPositionWriter("owner-a", null, {
  read: async () => {reads++;return 1;},
  write: async value => {sent.push(value);throw new Error(fail ? "network" : "account changed");},
  status: value => statuses.push(value), conflict: () => false,
  accountChanged: error => error instanceof Error && error.message === "account changed",
 });
 try {
  writer.select(position);await writer.flush();
  assert.equal(statuses.at(-1),"error");
  fail = false;await writer.retry();
  assert.deepEqual(sent.map(({expectedAccountId,expectedRevision})=>({expectedAccountId,expectedRevision})), [
   {expectedAccountId:"owner-a",expectedRevision:1}, {expectedAccountId:"owner-a",expectedRevision:1},
  ]);
  assert.equal(statuses.at(-1),"account-changed");
  writer.select({...position,rank:4331});await writer.flush();await writer.retry();
  assert.equal(sent.length,2);assert.equal(reads,2);
  assert.equal(statuses.at(-1),"account-changed");
 } finally {writer.dispose();}
});
