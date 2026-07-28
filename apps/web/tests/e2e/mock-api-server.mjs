// VOC-031-T07a mock API server for the accessibility harness.
//
// This is a deliberately tiny, dependency-free node:http server that
// returns the minimum canned JSON the Home page server component
// needs to render in isolation. The whole point of T07a is to prove
// the Playwright + axe-core harness works end-to-end; full API
// coverage and the real Postgres-backed /api/v1 belong to the
// Go API service and to T08 (the DOC-10 §7 core-loop E2E suite).
//
// Endpoints (matching the openapi-generated @vocanova/api-client
// shape, NOT introducing any new path):
//
//   GET /api/v1/me
//     200 -> { onboardingStatus: "completed", ... }
//     401 if ?fail=me is set (used to verify the auth-gate path)
//
//   GET /api/v1/user-words
//     200 -> { items: [] }
//
//   GET /api/v1/reviews/due
//     200 -> { items: [], totalCount: 0 }
//
//   GET /api/v1/daily-mission
//     200 -> DailyMission with reviewsCompleted=0, streak=0
//
//   GET /healthz
//     200 -> { status: "ok" }  (Playwright webServer probe URL)
//
// Anything else returns 404. The server logs each request to stderr
// so a CI failure's log can be matched against the harness's
// expectations without enabling extra debug output.

import { createServer } from "node:http";

const PORT = Number(process.env.MOCK_API_PORT ?? 8080);
const HOST = process.env.MOCK_API_HOST ?? "127.0.0.1";

function jsonResponse(res, status, body) {
  const payload = JSON.stringify(body);
  res.writeHead(status, {
    "Content-Type": "application/json; charset=utf-8",
    "Content-Length": Buffer.byteLength(payload),
    "Cache-Control": "no-store",
  });
  res.end(payload);
}

const currentUser = {
  email: "accessibility-fixture@example.test",
  displayName: "Accessibility Fixture",
  emailVerifiedAt: "2026-01-01T00:00:00Z",
  onboardingStatus: "completed",
};

const emptySavedWords = { items: [], nextCursor: undefined };

const emptyDueWords = { items: [], nextCursor: undefined, totalCount: 0 };

const emptyDailyMission = {
  localDate: "2026-01-01",
  timezone: "UTC",
  reviewTarget: 20,
  reviewsCompleted: 0,
  newWordTarget: 5,
  newWordsCompleted: 0,
  sentencePracticeTarget: 3,
  sentencePracticesCompleted: 0,
  policyVersion: "t07a-fixture-v1",
  status: "open",
  graceApplied: false,
  streak: {
    currentStreakCount: 0,
    longestStreakCount: 0,
    status: "active",
    graceDayBalance: 0,
  },
};

function logLine(req, status, extra) {
  const tag = extra ? ` ${JSON.stringify(extra)}` : "";
  process.stderr.write(
    `[mock-api] ${req.method} ${req.url} -> ${status}${tag}\n`,
  );
}

const server = createServer((req, res) => {
  if (!req.url) {
    res.writeHead(400);
    res.end();
    return;
  }

  const url = new URL(req.url, `http://${req.headers.host ?? `${HOST}:${PORT}`}`);

  if (req.method === "GET" && url.pathname === "/healthz") {
    logLine(req, 200);
    jsonResponse(res, 200, { status: "ok" });
    return;
  }

  if (req.method === "GET" && url.pathname === "/api/v1/me") {
    if (url.searchParams.get("fail") === "me") {
      logLine(req, 401, { reason: "fixture-forced-401" });
      jsonResponse(res, 401, { error: "unauthorized" });
      return;
    }
    logLine(req, 200);
    jsonResponse(res, 200, currentUser);
    return;
  }

  if (req.method === "GET" && url.pathname === "/api/v1/user-words") {
    logLine(req, 200);
    jsonResponse(res, 200, emptySavedWords);
    return;
  }

  if (req.method === "GET" && url.pathname === "/api/v1/reviews/due") {
    logLine(req, 200);
    jsonResponse(res, 200, emptyDueWords);
    return;
  }

  if (req.method === "GET" && url.pathname === "/api/v1/daily-mission") {
    logLine(req, 200);
    jsonResponse(res, 200, emptyDailyMission);
    return;
  }

  logLine(req, 404);
  jsonResponse(res, 404, { error: "not_found", path: url.pathname });
});

server.listen(PORT, HOST, () => {
  process.stderr.write(`[mock-api] listening on http://${HOST}:${PORT}\n`);
});

function shutdown(signal) {
  process.stderr.write(`[mock-api] received ${signal}, shutting down\n`);
  server.close(() => {
    process.exit(0);
  });
}

process.on("SIGINT", () => shutdown("SIGINT"));
process.on("SIGTERM", () => shutdown("SIGTERM"));
