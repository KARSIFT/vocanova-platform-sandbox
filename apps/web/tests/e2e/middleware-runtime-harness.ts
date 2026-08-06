// VOC-039-T01 middleware runtime harness.
//
// Issue #297's root cause is not "middleware's logic is wrong" - the
// logic is identical either way. It is that Next.js runs
// `src/middleware.ts` on the Edge runtime unless the module exports
// `runtime = "nodejs"`, and an Edge bundle only sees the environment
// inlined into it at build time. This deployment supplies
// `API_BASE_URL` to the running container instead (see
// `playwright.config.ts`'s own webServer env and `infra/`), so inside
// the Edge sandbox `src/lib/env.ts`'s `getApiBaseURL()` silently falls
// back to `http://localhost:8080`, the `/api/v1/me` auth check never
// reaches the real API, and every authenticated learner is bounced
// back to `/signin`.
//
// A plain unit test of `middleware()` cannot see any of that: run
// under Node, the exact same source reads the real environment and
// passes whether or not the `runtime` export exists. So this harness
// executes the real, unmodified `src/middleware.ts` source twice -
// once inside the Edge sandbox Next.js itself ships
// (`next/dist/compiled/edge-runtime`) and once in a Node context -
// against a stub API server the test owns. The only difference between
// the two runs is the runtime, which is exactly the variable
// VOC-039-T00 changes.

import { createServer } from "node:http";
import type { AddressInfo, Server } from "node:net";
import { createRequire, stripTypeScriptTypes } from "node:module";
import { readFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { createContext, runInContext } from "node:vm";

const require = createRequire(import.meta.url);

const HERE = dirname(fileURLToPath(import.meta.url));
const WEB_APP_ROOT = resolve(HERE, "../..");
const MIDDLEWARE_SOURCE_PATH = resolve(WEB_APP_ROOT, "src/middleware.ts");
const ENV_MODULE_SOURCE_PATH = resolve(WEB_APP_ROOT, "src/lib/env.ts");

export const CURRENT_USER_PATH = "/api/v1/me";
export const SESSION_COOKIE_NAME = "vocanova_session";

/**
 * Next.js runs middleware on the Edge runtime unless the module
 * exports `runtime = "nodejs"` (supported for middleware since
 * Next.js 15.2; this app is on 16.2.10). "edge" is therefore the
 * runtime this repo had before VOC-039-T00.
 */
export type MiddlewareRuntime = "edge" | "nodejs";

export const DEFAULT_MIDDLEWARE_RUNTIME: MiddlewareRuntime = "edge";

/**
 * The environment an Edge middleware bundle can actually see: values
 * inlined at build time, not the variables handed to the running
 * container. Reproducing that gap faithfully is the whole point of
 * running the middleware inside the Edge sandbox here.
 */
const EDGE_BUILD_TIME_ENV: Record<string, string> = {
  NODE_ENV: "production",
};

export type MiddlewareOutcome =
  | { kind: "next" }
  | { kind: "redirect"; location: string }
  | { kind: "threw"; message: string };

export interface AuthCheckRequest {
  /** Absolute URL of the protected route being requested. */
  url: string;
  /** Cookie header the browser would send, e.g. a real session cookie. */
  cookieHeader: string;
  /** Origin the deployed API is reachable at, at run time only. */
  apiBaseURL: string;
}

export interface StubApiServer {
  /** Origin the middleware must reach for the auth check to succeed. */
  baseURL: string;
  /** Paths this server actually received, in arrival order. */
  receivedPaths: string[];
  close(): Promise<void>;
}

/**
 * startStubApiServer starts a throwaway stand-in for the deployed API
 * on an ephemeral port and records every request it receives.
 *
 * The ephemeral port is load-bearing: it is an address that exists
 * only at run time and can only be discovered through
 * `process.env.API_BASE_URL`, so "did this server receive the
 * `/api/v1/me` call" is a direct, unfakeable signal that the auth
 * check resolved the runtime environment. Nothing else can reach it by
 * accident - notably not `getApiBaseURL()`'s `http://localhost:8080`
 * fallback, which in this repo's e2e run is the mock API server's own
 * port.
 */
export async function startStubApiServer(): Promise<StubApiServer> {
  const receivedPaths: string[] = [];

  const server: Server = createServer((request, response) => {
    receivedPaths.push(request.url ?? "");
    const cookies = request.headers.cookie ?? "";
    if (!cookies.includes(`${SESSION_COOKIE_NAME}=`)) {
      response.writeHead(401);
      response.end();
      return;
    }
    response.writeHead(200, { "content-type": "application/json" });
    response.end(JSON.stringify({ onboardingStatus: "completed" }));
  });

  await new Promise<void>((resolveListening) => {
    server.listen(0, "127.0.0.1", resolveListening);
  });
  const { port } = server.address() as AddressInfo;

  return {
    baseURL: `http://127.0.0.1:${port}`,
    receivedPaths,
    close: () =>
      new Promise<void>((resolveClose, rejectClose) => {
        server.close((error) => (error ? rejectClose(error) : resolveClose()));
      }),
  };
}

/**
 * toRunnableScript turns an app source module into a script a sandbox
 * can evaluate: TypeScript annotations are stripped, the single-line
 * imports are dropped (the preamble provides `NextResponse`, and
 * `env.ts` is evaluated into the same scope), and `export` markers are
 * removed so the declarations land in the script's own scope.
 *
 * This deliberately rewrites no statement inside the module - the
 * environment lookup, the fetch call, its `try/catch`, and every
 * redirect branch execute verbatim. If either file ever grows a
 * multi-line or dynamic import, this transform breaks loudly in CI
 * instead of silently testing something other than the real source.
 */
function toRunnableScript(typeScriptSource: string): string {
  return stripTypeScriptTypes(typeScriptSource, { mode: "strip" })
    .replace(/^import\s[^;]*;\s*$/gm, "")
    .replace(/^export\s+/gm, "");
}

async function readMiddlewareScripts(): Promise<string> {
  const [envModuleSource, middlewareSource] = await Promise.all([
    readFile(ENV_MODULE_SOURCE_PATH, "utf8"),
    readFile(MIDDLEWARE_SOURCE_PATH, "utf8"),
  ]);
  return `${toRunnableScript(envModuleSource)}\n${toRunnableScript(middlewareSource)}`;
}

/**
 * NEXT_RESPONSE_RECORDER stands in for `next/server`'s `NextResponse`.
 * `NextResponse` is runtime-agnostic plumbing; what is under test here
 * is whether the auth check's environment lookup and `fetch()` can
 * reach the deployed API from the runtime middleware actually runs on,
 * so recording which branch was taken is all the harness needs.
 */
const NEXT_RESPONSE_RECORDER = `
  globalThis.NextResponse = {
    redirect: (url) => ({ kind: "redirect", location: String(url) }),
    next: () => ({ kind: "next" }),
  };
`;

function buildDriverScript(request: AuthCheckRequest): string {
  const requestUrl = new URL(request.url);
  return `
    (async () => {
      const request = {
        url: ${JSON.stringify(request.url)},
        nextUrl: {
          pathname: ${JSON.stringify(requestUrl.pathname)},
          search: ${JSON.stringify(requestUrl.search)},
        },
        headers: new Headers({ cookie: ${JSON.stringify(request.cookieHeader)} }),
      };
      try {
        return JSON.stringify(await middleware(request));
      } catch (error) {
        return JSON.stringify({ kind: "threw", message: String(error) });
      }
    })()
  `;
}

/**
 * runMiddlewareOnEdgeRuntime executes the real middleware source
 * inside the Edge sandbox Next.js itself ships, with the
 * build-time-only environment an Edge middleware bundle gets in
 * production.
 */
export async function runMiddlewareOnEdgeRuntime(
  request: AuthCheckRequest,
): Promise<MiddlewareOutcome> {
  const { EdgeRuntime } = require("next/dist/compiled/edge-runtime") as {
    EdgeRuntime: new () => { evaluate(code: string): Promise<string> };
  };

  const sandbox = new EdgeRuntime();
  const serialized = await sandbox.evaluate(`
    globalThis.process = { env: ${JSON.stringify(EDGE_BUILD_TIME_ENV)} };
    ${NEXT_RESPONSE_RECORDER}
    ${await readMiddlewareScripts()}
    ${buildDriverScript(request)}
  `);

  return JSON.parse(serialized) as MiddlewareOutcome;
}

/**
 * runMiddlewareOnNodeRuntime executes the same real middleware source
 * in a Node.js context: Node's own `fetch`, and a `process.env` that
 * carries the container-supplied `API_BASE_URL` the Edge bundle cannot
 * see. This is what Next.js gives middleware once the module exports
 * `runtime = "nodejs"`.
 *
 * The environment is passed as a copy rather than by mutating this
 * process's own `process.env`, so a failing run cannot leak the stub
 * API's address into any other test.
 */
export async function runMiddlewareOnNodeRuntime(
  request: AuthCheckRequest,
): Promise<MiddlewareOutcome> {
  const context = createNodeMiddlewareContext({
    ...process.env,
    API_BASE_URL: request.apiBaseURL,
  });
  runInContext(
    `${NEXT_RESPONSE_RECORDER}\n${await readMiddlewareScripts()}`,
    context,
  );
  const serialized = (await runInContext(
    buildDriverScript(request),
    context,
  )) as string;

  return JSON.parse(serialized) as MiddlewareOutcome;
}

function createNodeMiddlewareContext(env: NodeJS.ProcessEnv): object {
  return createContext({
    process: { env },
    fetch: globalThis.fetch,
    Headers: globalThis.Headers,
    Request: globalThis.Request,
    Response: globalThis.Response,
    URL: globalThis.URL,
    URLSearchParams: globalThis.URLSearchParams,
    TextEncoder: globalThis.TextEncoder,
    TextDecoder: globalThis.TextDecoder,
    console: globalThis.console,
  });
}

/**
 * readDeclaredMiddlewareRuntime reports the runtime Next.js will
 * actually use for `src/middleware.ts`, by evaluating the module and
 * reading the value of its `runtime` export rather than
 * pattern-matching source text. A module with no `runtime` export runs
 * on the Edge runtime.
 */
export async function readDeclaredMiddlewareRuntime(): Promise<MiddlewareRuntime> {
  const context = createNodeMiddlewareContext({});
  runInContext(
    `${NEXT_RESPONSE_RECORDER}\n${await readMiddlewareScripts()}`,
    context,
  );
  const declared = runInContext(
    `typeof runtime === "string" ? runtime : null`,
    context,
  ) as string | null;

  return declared === "nodejs" ? "nodejs" : DEFAULT_MIDDLEWARE_RUNTIME;
}

/**
 * runMiddlewareOnDeclaredRuntime runs the auth check the way the
 * deployed app will actually run it: under whichever runtime
 * `src/middleware.ts` declares.
 */
export async function runMiddlewareOnDeclaredRuntime(
  request: AuthCheckRequest,
): Promise<MiddlewareOutcome> {
  const declaredRuntime = await readDeclaredMiddlewareRuntime();
  return declaredRuntime === "nodejs"
    ? runMiddlewareOnNodeRuntime(request)
    : runMiddlewareOnEdgeRuntime(request);
}
