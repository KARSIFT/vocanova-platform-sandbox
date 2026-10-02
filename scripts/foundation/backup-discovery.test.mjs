import { spawnSync } from "node:child_process";
import assert from "node:assert/strict";
import { test } from "node:test";

test("backup discovery bounds observations and protects private SSH data", () => {
  const result = spawnSync(
    "python3",
    [
      "-m",
      "unittest",
      "discover",
      "-s",
      "infra/scripts",
      "-p",
      "test_*backup*.py",
    ],
    { encoding: "utf8", timeout: 60000 },
  );
  assert.equal(result.status, 0, result.stderr || result.error?.message);
});
