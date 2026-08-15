// VOC-081-T02 — monitor.vocanova.site vhost and access exposure policy.
//
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { existsSync, readFileSync, readdirSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const monitorVhostPath = path.join(
  repositoryRoot,
  "infra/nginx-shared/conf.d/30-monitor.vocanova.site.conf",
);
const sharedNginxConfPath = path.join(
  repositoryRoot,
  "infra/nginx-shared/nginx.conf",
);
const productionConfDir = path.join(
  repositoryRoot,
  "infra/nginx-production/conf.d",
);
const accessPolicyPath = path.join(
  repositoryRoot,
  "infra/monitoring/access-policy.md",
);
const supersededMarkerPath = path.join(
  productionConfDir,
  "30-monitor.conf.superseded",
);

const MONITOR_HOSTNAME = "monitor.vocanova.site";
const KUMA_UPSTREAM = "vocanova-uptime-kuma:3001";

test("VOC-081-TEST-03: repository vhost declares monitor.vocanova.site with TLS and proxy headers", () => {
  assert.ok(
    existsSync(monitorVhostPath),
    "monitor vhost must exist under infra/nginx-shared/conf.d/",
  );

  const vhost = readFileSync(monitorVhostPath, "utf8");

  assert.match(
    vhost,
    new RegExp(`server_name\\s+${MONITOR_HOSTNAME.replace(".", "\\.")}\\s*;`),
    "vhost must declare server_name monitor.vocanova.site",
  );
  assert.match(
    vhost,
    /ssl_certificate\s+\/etc\/nginx\/certs\/production\/cert\.pem;/,
    "vhost must use production TLS certificate path",
  );
  assert.match(
    vhost,
    /ssl_certificate_key\s+\/etc\/nginx\/certs\/production\/key\.pem;/,
    "vhost must use production TLS private key path",
  );
  assert.match(
    vhost,
    new RegExp(KUMA_UPSTREAM.replace(":", "\\:")),
    "upstream must target Kuma over monitoring-network DNS",
  );

  for (const header of [
    "proxy_set_header Host",
    "proxy_set_header X-Real-IP",
    "proxy_set_header X-Forwarded-For",
    "proxy_set_header X-Forwarded-Proto",
  ]) {
    assert.match(vhost, new RegExp(header), `vhost must set ${header}`);
  }

  assert.match(
    vhost,
    /proxy_http_version\s+1\.1;/,
    "vhost must use HTTP/1.1 for WebSocket upgrades",
  );
  assert.match(
    vhost,
    /proxy_set_header Upgrade\s+\$http_upgrade;/,
    "vhost must forward Upgrade for WebSockets",
  );
  assert.match(
    vhost,
    /proxy_set_header Connection\s+\$monitor_connection_upgrade;/,
    "vhost must set Connection upgrade semantics for Kuma",
  );
  assert.match(
    vhost,
    /map \$http_upgrade \$monitor_connection_upgrade/,
    "vhost must map Connection for upgrade vs close",
  );
});

test("VOC-081-TEST-03: shared-edge nginx.conf loads the monitor vhost via shared/*.conf", () => {
  const nginxConf = readFileSync(sharedNginxConfPath, "utf8");

  assert.match(
    nginxConf,
    /include\s+\/etc\/nginx\/conf\.d\/shared\/\*\.conf;/,
    "shared-edge must include all shared conf.d fragments",
  );
  assert.doesNotMatch(
    nginxConf,
    /include\s+\/etc\/nginx\/conf\.d\/production\/30-/,
    "monitor vhost must not rely on unloaded production 30-*.conf globs",
  );

  const sharedConfFiles = readdirSync(
    path.join(repositoryRoot, "infra/nginx-shared/conf.d"),
  );
  assert.ok(
    sharedConfFiles.includes("30-monitor.vocanova.site.conf"),
    "monitor vhost fragment must be present in nginx-shared/conf.d",
  );
});

test("VOC-081-TEST-03: production conf.d has no active 30-monitor.conf fragment", () => {
  const productionConfFiles = readdirSync(productionConfDir);
  assert.ok(
    !productionConfFiles.includes("30-monitor.conf"),
    "production tree must not ship an active 30-monitor.conf vhost",
  );
  assert.ok(
    existsSync(supersededMarkerPath),
    "stale 30-monitor.conf must have a repository retirement marker",
  );
});

test("VOC-081-TEST-04: access exposure policy is explicit (public Kuma login, not DNS alone)", () => {
  assert.ok(
    existsSync(accessPolicyPath),
    "access policy document must exist at infra/monitoring/access-policy.md",
  );

  const policy = readFileSync(accessPolicyPath, "utf8");

  assert.match(
    policy,
    /Public Uptime Kuma login|public Kuma login/i,
    "policy must select the public Kuma login exposure model",
  );
  assert.match(
    policy,
    new RegExp(MONITOR_HOSTNAME.replace(".", "\\.")),
    "policy must name monitor.vocanova.site",
  );
  assert.match(
    policy,
    /not.*authorization|not authorization/i,
    "policy must state proxied DNS is not authorization",
  );
  assert.match(
    policy,
    /Kuma.*auth|Uptime Kuma.*login/i,
    "policy must require Kuma authentication to remain enabled",
  );
  assert.match(
    policy,
    /does not claim or depend on a Cloudflare\s+Access application/i,
    "policy must not invent an unverified Cloudflare Access dependency",
  );
  assert.match(
    policy,
    /T04|verification probe|curl/i,
    "policy must define a live verification probe for T04",
  );
  assert.doesNotMatch(
    policy,
    /only.*proxied.*dns.*(?:is|as).*control/i,
    "policy must not cite proxied DNS alone as the access control",
  );
});
