import assert from "node:assert/strict";
import {
  mkdtempSync,
  mkdirSync,
  readFileSync,
  writeFileSync,
  rmSync,
} from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";
import test from "node:test";

const workflow = readFileSync(
  new URL("../../.github/workflows/deploy-staging.yml", import.meta.url),
  "utf8",
);
const routine = workflow.match(
  /            if docker inspect "\$SHARED_EDGE_CONTAINER" >\/dev\/null 2>&1; then\n([\s\S]*?)\n            else/,
)[1];

function exercise(t, options = {}) {
  const dir = mkdtempSync(path.join(tmpdir(), "shared-edge-base-"));
  t.after(() => rmSync(dir, { recursive: true, force: true }));
  mkdirSync(path.join(dir, "bin"));
  mkdirSync(path.join(dir, "nginx-shared"));
  writeFileSync(
    path.join(dir, "nginx-shared/nginx.conf"),
    "desired URI-only logging\n",
  );
  writeFileSync(
    path.join(dir, "mounted.conf"),
    options.same
      ? "desired URI-only logging\n"
      : "old inode with combined logging\n",
  );
  const stub = `#!/usr/bin/env node
const fs=require('node:fs'),crypto=require('node:crypto');
const args=process.argv.slice(2), cmd=args.join(' '), mode=JSON.parse(process.env.EDGE_CASE);
fs.appendFileSync('calls',cmd+'\\n');
function hash(file){return crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex')+'  '+file+'\\n'}
if(process.argv[1].endsWith('sha256sum')){
 if(mode.hostReadFail)process.exit(1);
 process.stdout.write(mode.hostInvalidHash?'not-a-hash\\n':hash(args[0]));process.exit(0);
}
if(args[0]==='compose'&&args.at(-1)==='config'&&mode.composeInvalid)process.exit(1);
if(cmd.includes(' sha256sum ')){
 if(mode.readFail)process.exit(1);
 process.stdout.write(mode.invalidHash?'not-a-hash\\n':hash('mounted.conf'));process.exit(0);
}
if(args[0]==='compose'&&args.includes('up')){
 if(mode.upFail)process.exit(1);
 if(args.includes('--force-recreate')&&!mode.staysStale)fs.copyFileSync('nginx-shared/nginx.conf','mounted.conf');
 process.exit(0);
}
if(args[0]==='inspect'){process.stdout.write('healthy\\n');process.exit(0)}
if(cmd.endsWith('nginx -t')&&mode.runningInvalid)process.exit(1);
process.exit(0);
`;
  for (const command of ["docker", "sha256sum"])
    writeFileSync(path.join(dir, "bin", command), stub, { mode: 0o755 });
  const result = spawnSync(
    "bash",
    [
      "-c",
      `set -euo pipefail
SHARED_EDGE_CONTAINER=vocanova-shared-edge-nginx
run_shared_edge_nginx_t() { printf 'desired-validation\\n' >> calls; ${options.desiredInvalid ? "return 1" : "return 0"}; }
${routine}`,
    ],
    {
      cwd: dir,
      encoding: "utf8",
      timeout: 10_000,
      env: {
        ...process.env,
        PATH: `${dir}/bin:${process.env.PATH}`,
        EDGE_CASE: JSON.stringify(options),
      },
    },
  );
  return {
    ...result,
    calls: readFileSync(path.join(dir, "calls"), "utf8").trim().split("\n"),
    mounted: readFileSync(path.join(dir, "mounted.conf"), "utf8"),
  };
}

test("stale single-file bind is recreated after validation and converges before reload", (t) => {
  const result = exercise(t);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.mounted, "desired URI-only logging\n");
  const up = result.calls.filter((call) => call.includes(" up "));
  assert.deepEqual(up, [
    "compose -f docker-compose.shared-edge.yml -p vocanova-shared-edge up -d --no-deps --force-recreate nginx",
  ]);
  assert.equal(result.calls[0], "desired-validation");
  assert.equal(
    result.calls.at(-1),
    "exec vocanova-shared-edge-nginx nginx -s reload",
  );
  assert.equal(
    result.calls.filter((call) => call.includes(" sha256sum ")).length,
    2,
  );
});

test("identical mounted config uses ordinary convergence without forced recreation", (t) => {
  const result = exercise(t, { same: true });
  assert.equal(result.status, 0, result.stderr);
  assert.ok(result.calls.some((call) => call.includes(" up -d")));
  assert.ok(result.calls.every((call) => !call.includes("--force-recreate")));
});

for (const option of [
  "desiredInvalid",
  "composeInvalid",
  "hostReadFail",
  "hostInvalidHash",
  "readFail",
  "invalidHash",
]) {
  test(`${option} fails closed before any edge replacement or reload`, (t) => {
    const result = exercise(t, { [option]: true });
    assert.notEqual(result.status, 0);
    assert.ok(
      result.calls.every(
        (call) => !call.includes(" up ") && !call.includes("nginx -s reload"),
      ),
    );
  });
}

for (const option of ["upFail", "staysStale", "runningInvalid"]) {
  test(`${option} aborts without reloading an unconfirmed configuration`, (t) => {
    const result = exercise(t, { [option]: true });
    assert.notEqual(result.status, 0);
    assert.ok(result.calls.every((call) => !call.includes("nginx -s reload")));
  });
}
