import assert from "node:assert/strict";
import { it } from "node:test";
import { isSameAppChunkLoadError } from "../../src/ui/route-chunk-error";
const origin = "https://production.vocanova.site";
const chunk = "/_next/static/chunks/settings-123.js";
function failed(url: string) {
  return Object.assign(new Error(`Failed to load chunk ${url} from module [project]/settings`), {name: "ChunkLoadError"});
}
it("recognizes the emitted same-app Turbopack chunk failure", () => {
  assert.equal(isSameAppChunkLoadError(failed(chunk), origin), true);
  assert.equal(isSameAppChunkLoadError(failed(origin + chunk), origin), true);
});
for (const [label, error] of [
  ["ordinary API failure",new Error("Service unavailable")],
  ["message without chunk error identity",new Error(`Failed to load chunk ${chunk}`)],
  ["name without recognized message",Object.assign(new Error("Network unavailable"),{name:"ChunkLoadError"})],
  ["external origin",failed("https://external.example" + chunk)],
  ["protocol-relative external URL",failed("//external.example" + chunk)],
  ["non-chunk app resource",failed("/api/v1/settings")],
  ["non-JavaScript resource",failed("/_next/static/chunks/settings.css")],
  ["nested unrelated error",{cause:failed(chunk)}],
  ["missing error",undefined],
] as const) it(`keeps soft recovery for ${label}`,()=>assert.equal(isSameAppChunkLoadError(error,origin),false));
