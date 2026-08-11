import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import test from "node:test";

const PRODUCTION_COMPOSE_PATH = path.resolve(
  "infra/docker-compose.production.yml",
);
const WEB_SERVICE_MARKER = "  web:";
const WEB_ENVIRONMENT_MARKER = "    environment:";
const WEB_HEALTHCHECK_MARKER = "    healthcheck:";
const API_BASE_URL_KEY = "      API_BASE_URL:";
const EXPECTED_API_BASE_URL = "https://api-production.vocanova.site";

function extractWebEnvironmentBlock(composeSource) {
  const webServiceIndex = composeSource.indexOf(WEB_SERVICE_MARKER);
  assert.notEqual(
    webServiceIndex,
    -1,
    "production compose file is missing the web service marker",
  );

  const environmentIndex = composeSource.indexOf(
    WEB_ENVIRONMENT_MARKER,
    webServiceIndex,
  );
  assert.notEqual(
    environmentIndex,
    -1,
    "production compose file is missing the web environment marker",
  );

  const healthcheckIndex = composeSource.indexOf(
    WEB_HEALTHCHECK_MARKER,
    environmentIndex,
  );
  assert.notEqual(
    healthcheckIndex,
    -1,
    "production compose file is missing the web healthcheck marker after environment",
  );

  return composeSource.slice(environmentIndex, healthcheckIndex);
}

function extractApiBaseUrl(webEnvironmentBlock) {
  const line = webEnvironmentBlock
    .split("\n")
    .find((candidateLine) => candidateLine.startsWith(API_BASE_URL_KEY));

  assert.ok(line, "production compose web environment is missing API_BASE_URL");

  const apiBaseUrl = line.slice(API_BASE_URL_KEY.length).trim();
  assert.ok(
    apiBaseUrl.length > 0,
    "production compose API_BASE_URL is present but empty",
  );

  return apiBaseUrl;
}

test("VOC-067-TEST-05: web API_BASE_URL uses ordinary :443 hostname (no :8443)", () => {
  const composeSource = readFileSync(PRODUCTION_COMPOSE_PATH, "utf8");
  const webEnvironmentBlock = extractWebEnvironmentBlock(composeSource);
  const apiBaseUrl = extractApiBaseUrl(webEnvironmentBlock);

  assert.equal(
    apiBaseUrl,
    EXPECTED_API_BASE_URL,
    `Expected web API_BASE_URL to be ${EXPECTED_API_BASE_URL}, got ${apiBaseUrl}`,
  );
  assert.doesNotMatch(
    apiBaseUrl,
    /:8443/,
    "production compose API_BASE_URL must not include the retired :8443 cutover port",
  );
});
