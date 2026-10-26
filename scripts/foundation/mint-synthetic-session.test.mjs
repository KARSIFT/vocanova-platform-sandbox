// Focused tests for infra/scripts/mint-synthetic-session.sh. A fake curl on
// PATH returns fixture mint responses; nothing touches the network.

import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const root = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);
const mintScript = path.join(root, "infra/scripts/mint-synthetic-session.sh");
const read = (file) => readFileSync(path.join(root, file), "utf8");

// Synthetic sentinels: any appearance in stdout/stderr outside a mask command
// is a leak.
const MINT_TOKEN = "fixture-mint-token-7c1e";
const SESSION = "fixture-session-value-91af";
const CSRF = "fixture-csrf-value-4d22";
const INJECTED = "evil-injected-3b8d";

function mintResponse(tier, overrides = {}) {
  return {
    session_cookie: SESSION,
    csrf_token: CSRF,
    session_cookie_name: `vocanova_${tier}_session`,
    csrf_cookie_name: `vocanova_${tier}_csrf`,
    ...overrides,
  };
}

function runMint(t, body, env = {}) {
  const directory = mkdtempSync(path.join(tmpdir(), "vocanova-mint-"));
  t.after(() => rmSync(directory, { recursive: true, force: true }));
  const responseFile = path.join(directory, "response.json");
  const outputFile = path.join(directory, "github-output");
  writeFileSync(
    responseFile,
    typeof body === "string" ? body : JSON.stringify(body),
  );
  writeFileSync(outputFile, "");
  writeFileSync(
    path.join(directory, "curl"),
    '#!/bin/sh\ncat "$FAKE_MINT_RESPONSE"\n',
    { mode: 0o700 },
  );
  const childEnv = {
    PATH: `${directory}${path.delimiter}${process.env.PATH}`,
    FAKE_MINT_RESPONSE: responseFile,
    GITHUB_OUTPUT: outputFile,
    SMOKE_TEST_SESSION_MINT_TOKEN: MINT_TOKEN,
    SYNTHETIC_COOKIE_ENVIRONMENT: "staging",
    ...env,
  };
  for (const [key, value] of Object.entries(childEnv)) {
    if (value === undefined) delete childEnv[key];
  }
  const result = spawnSync(
    "bash",
    [mintScript, "https://api.example.invalid"],
    {
      env: childEnv,
      encoding: "utf8",
      timeout: 10_000,
    },
  );
  assert.equal(result.error, undefined);
  return { ...result, output: readFileSync(outputFile, "utf8") };
}

for (const tier of ["staging", "production"]) {
  test(`mint writes validated ${tier} names, opaque values and a separate header`, (t) => {
    const result = runMint(t, mintResponse(tier), {
      SYNTHETIC_COOKIE_ENVIRONMENT: tier,
    });
    assert.equal(result.status, 0, result.stderr);
    assert.equal(
      result.output,
      [
        `session_cookie=${SESSION}`,
        `csrf_token=${CSRF}`,
        `session_cookie_name=vocanova_${tier}_session`,
        `csrf_cookie_name=vocanova_${tier}_csrf`,
        `session_cookie_header=vocanova_${tier}_session=${SESSION}`,
        "",
      ].join("\n"),
    );
    assert.deepEqual(
      result.stdout
        .split("\n")
        .filter((line) => line.includes(SESSION) || line.includes(CSRF)),
      [`::add-mask::${SESSION}`, `::add-mask::${CSRF}`],
    );
    assert.ok(
      !result.stderr.includes(SESSION) && !result.stderr.includes(CSRF),
    );
    assert.ok(!`${result.stdout}${result.stderr}`.includes(MINT_TOKEN));
  });
}

for (const [label, body, env] of [
  [
    "missing session cookie name",
    mintResponse("staging", { session_cookie_name: undefined }),
  ],
  [
    "missing CSRF cookie name",
    mintResponse("staging", { csrf_cookie_name: undefined }),
  ],
  [
    "missing session token",
    mintResponse("staging", { session_cookie: undefined }),
  ],
  ["empty CSRF token", mintResponse("staging", { csrf_token: "" })],
  [
    "legacy cookie names",
    mintResponse("staging", {
      session_cookie_name: "vocanova_session",
      csrf_cookie_name: "vocanova_csrf",
    }),
  ],
  ["production names for the staging tier", mintResponse("production")],
  [
    "staging names for the production tier",
    mintResponse("staging"),
    { SYNTHETIC_COOKIE_ENVIRONMENT: "production" },
  ],
  [
    "mixed-tier names",
    mintResponse("staging", { csrf_cookie_name: "vocanova_production_csrf" }),
  ],
  [
    "cookie-name injection",
    mintResponse("staging", {
      session_cookie_name: `vocanova_staging_session; ${INJECTED}=1`,
    }),
  ],
  [
    "CR/LF in the session value",
    mintResponse("staging", {
      session_cookie: `${SESSION}\r\nsession_cookie_header=${INJECTED}`,
    }),
  ],
  [
    "semicolon in the CSRF value",
    mintResponse("staging", { csrf_token: `${CSRF};${INJECTED}=1` }),
  ],
  ["non-JSON response", `not json ${SESSION}`],
  [
    "missing expected tier",
    mintResponse("staging"),
    { SYNTHETIC_COOKIE_ENVIRONMENT: undefined },
  ],
  [
    "unknown expected tier",
    mintResponse("staging"),
    { SYNTHETIC_COOKIE_ENVIRONMENT: "local" },
  ],
  [
    "legacy SESSION_COOKIE_PREFIX",
    mintResponse("staging"),
    { SESSION_COOKIE_PREFIX: "vocanova_session" },
  ],
]) {
  test(`mint fails closed without leaking values: ${label}`, (t) => {
    const result = runMint(t, body, env);
    assert.notEqual(result.status, 0);
    assert.equal(result.output, "", "no step outputs on failure");
    assert.notEqual(result.stderr.trim(), "");
    const logs = `${result.stdout}${result.stderr}`;
    for (const value of [SESSION, CSRF, MINT_TOKEN, INJECTED]) {
      assert.ok(!logs.includes(value), `${label} leaked a fixture value`);
    }
  });
}

test("deploy workflows mint through the shared helper with their expected tier", () => {
  const production = read(".github/workflows/deploy-production.yml");
  const staging = read(".github/workflows/deploy-staging.yml");
  for (const source of [production, staging]) {
    assert.doesNotMatch(
      source,
      /json\.load\(sys\.stdin\)\["session_cookie"\]/,
      "deploy workflows must not parse the mint response inline",
    );
  }
  assert.match(
    production,
    /SMOKE_TEST_SESSION_MINT_TOKEN: \$\{\{ secrets\.PRODUCTION_SMOKE_TEST_SESSION_MINT_TOKEN \}\}\s+SYNTHETIC_COOKIE_ENVIRONMENT: production\s/,
  );
  assert.match(
    production,
    /bash infra\/scripts\/mint-synthetic-session\.sh "\$PRODUCTION_API_BASE_URL"/,
  );
  assert.match(
    production,
    /SMOKE_TEST_SESSION_COOKIE: \$\{\{ steps\.mint_production_session\.outputs\.session_cookie_header \}\}/,
  );
  assert.equal(
    staging.match(
      /SMOKE_TEST_SESSION_MINT_TOKEN: \$\{\{ secrets\.STAGING_SMOKE_TEST_SESSION_MINT_TOKEN \}\}\s+SYNTHETIC_COOKIE_ENVIRONMENT: staging\s/g,
    )?.length,
    2,
  );
  assert.equal(
    staging.match(
      /bash infra\/scripts\/mint-synthetic-session\.sh \\\s+"https:\/\/api-staging\.vocanova\.site" \\\s+"api-staging\.vocanova\.site"/g,
    )?.length,
    2,
  );
  for (const id of ["mint_staging_session", "mint_staging_maturity_session"]) {
    for (const [variable, output] of [
      ["E2E_SESSION_COOKIE", "session_cookie"],
      ["E2E_CSRF_TOKEN", "csrf_token"],
      ["E2E_SESSION_COOKIE_NAME", "session_cookie_name"],
      ["E2E_CSRF_COOKIE_NAME", "csrf_cookie_name"],
    ]) {
      assert.match(
        staging,
        new RegExp(
          `${variable}: \\$\\{\\{ steps\\.${id}\\.outputs\\.${output} \\}\\}`,
        ),
      );
    }
  }
});
