import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

test("deployments retain the feedback adapter's supported Cloudflare default", async () => {
  const root = new URL("../../", import.meta.url);
  const [adapter, runtime, staging, production] = await Promise.all(
    [
      "apps/api/business/aifeedback/cloudflare.go",
      "apps/api/app/api/production.go",
      ".github/workflows/deploy-staging.yml",
      ".github/workflows/deploy-production.yml",
    ].map((path) => readFile(new URL(path, root), "utf8")),
  );
  const model = adapter.match(/defaultCloudflareModel\s*=\s*"([^"]+)"/)?.[1];
  assert.ok(model, "adapter must declare its default model");
  assert.equal(
    runtime.match(/defaultCloudflareModel\s*=\s*"([^"]+)"/)?.[1],
    model,
  );
  for (const [name, workflow] of [
    ["staging", staging],
    ["production", production],
  ]) {
    const configured = [
      ...workflow.matchAll(/echo "AI_PROVIDER_MODEL=([^"]+)"/g),
    ].map((match) => match[1]);
    assert.deepEqual(
      configured,
      [model],
      `${name} must not replace the supported adapter model with a different identifier`,
    );
  }
});
