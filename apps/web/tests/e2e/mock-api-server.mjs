// VOC-031-T07a + VOC-031-T07b mock API server for the accessibility
// harness.
//
// T07a added the minimum canned JSON the Home page server
// component needs to render. T07b extends the server so every
// remaining core-loop page (Discover, Discover/[situation],
// Discover/[situation]/[word], Reviews, Progress, Onboarding,
// Settings, Settings/account) can render with deterministic
// fixture data - the T07b scans do not exercise form submissions
// or POST endpoints, so this server intentionally returns
// read-only fixtures and 404s on writes.
//
// The whole point of this server is to prove the Playwright +
// axe-core harness works end-to-end; full API coverage and the
// real Postgres-backed /api/v1 belong to the Go API service and
// to T08 (the DOC-10 §7 core-loop E2E suite).
//
// Endpoints (matching the openapi-generated @vocanova/api-client
// shape, NOT introducing any new path):
//
//   GET /api/v1/me
//     200 -> { onboardingStatus: "completed", ... }
//     401 if ?fail=me is set (used to verify the auth-gate path)
//     onboardingStatus is overridden to the value of the
//     `e2e_onboarding_status` cookie when present (see the
//     onboarding-accessibility.spec.ts comment for why a cookie,
//     not page.route(), is the only way a Playwright test can
//     influence this response: both the Next.js middleware and
//     the /onboarding Server Component fetch this endpoint
//     directly, server-to-server, which never passes through the
//     browser's network stack that page.route() intercepts).
//
//   GET /api/v1/onboarding
//     200 -> { status: "not_started", ... }  (no completed profile yet)
//
//   GET /api/v1/user-words
//     200 -> { items: [] }  (no saved words yet - matches the
//                            empty-state rendering)
//
//   GET /api/v1/reviews/due
//     200 -> { items: [], totalCount: 0 }  (all-caught-up state)
//
//   GET /api/v1/daily-mission
//     200 -> DailyMission with reviewsCompleted=0, streak=0
//
//   GET /api/v1/journey-situations
//     200 -> { items: [cafe, airport, ...] }
//
//   GET /api/v1/journey-situations/:slug
//     200 -> { situation: {...}, meanings: [...] }
//     404 -> { error: "not_found" } for unknown slugs
//
//   GET /api/v1/canonical-words/:slug
//     200 -> { word: { text, wordType, meanings: [...] } }
//     404 -> { error: "not_found" } for unknown slugs
//
//   GET /api/v1/progress
//     200 -> { confidencePointsBalance, streak, completionHistory }
//
//   GET /api/v1/settings
//     200 -> { dailyReviewTarget, reviewIntervalPreset, appLanguage,
//               notificationsEnabled, marketingEmailsEnabled,
//               displayName }
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

const ONBOARDING_STATUSES = new Set(["not_started", "in_progress", "completed"]);

function parseCookies(header) {
  const cookies = {};
  if (!header) {
    return cookies;
  }
  for (const part of header.split(";")) {
    const idx = part.indexOf("=");
    if (idx === -1) {
      continue;
    }
    const key = part.slice(0, idx).trim();
    const value = part.slice(idx + 1).trim();
    if (key) {
      cookies[key] = decodeURIComponent(value);
    }
  }
  return cookies;
}

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

const onboardingProfile = {
  status: "not_started",
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
  policyVersion: "t07b-fixture-v1",
  status: "open",
  graceApplied: false,
  streak: {
    currentStreakCount: 0,
    longestStreakCount: 0,
    status: "active",
    graceDayBalance: 0,
  },
};

const journeySituations = {
  items: [
    {
      id: "sit-cafe",
      slug: "ordering-at-a-cafe",
      title: "Ordering at a cafe",
      shortDescription: "Polite everyday phrases for your morning coffee.",
      levelBand: "A1",
      category: "daily_life",
      displayOrder: 1,
    },
    {
      id: "sit-airport",
      slug: "navigating-an-airport",
      title: "Navigating an airport",
      shortDescription: "Check-in, security, and boarding conversations.",
      levelBand: "A2",
      category: "travel",
      displayOrder: 2,
    },
  ],
};

const journeySituationsBySlug = {
  "ordering-at-a-cafe": {
    situation: {
      id: "sit-cafe",
      slug: "ordering-at-a-cafe",
      title: "Ordering at a cafe",
      shortDescription: "Polite everyday phrases for your morning coffee.",
      levelBand: "A1",
      category: "daily_life",
      displayOrder: 1,
    },
    meanings: [
      {
        meaningId: "mean-pour",
        wordId: "word-pour",
        wordSlug: "pour",
        wordText: "pour",
        partOfSpeech: "verb",
        shortDefinition: "to make liquid flow into a container",
        saved: false,
      },
      {
        meaningId: "mean-receipt",
        wordId: "word-receipt",
        wordSlug: "receipt",
        wordText: "receipt",
        partOfSpeech: "noun",
        shortDefinition: "a paper that shows what you paid for",
        saved: true,
      },
    ],
  },
};

const canonicalWordsBySlug = {
  pour: {
    word: {
      id: "word-pour",
      text: "pour",
      slug: "pour",
      wordType: "verb",
      difficultyLevel: "A2",
      meanings: [
        {
          id: "mean-pour",
          partOfSpeech: "verb",
          shortDefinition: "to make liquid flow into a container",
          saved: false,
          examples: [
            {
              id: "ex-pour-1",
              exampleText: "Could you pour me a cup of coffee?",
            },
          ],
          usageNotes: [
            {
              id: "note-pour-1",
              noteType: "register",
              noteText: "Common in everyday service contexts.",
            },
          ],
        },
      ],
    },
  },
};

const progress = {
  confidencePointsBalance: 120,
  streak: {
    currentStreakCount: 3,
    longestStreakCount: 7,
    status: "active",
    graceDayBalance: 1,
  },
  completionHistory: [
    { localDate: "2026-01-01", completed: true },
    { localDate: "2026-01-02", completed: true },
    { localDate: "2026-01-03", completed: true },
    { localDate: "2026-01-04", completed: false },
    { localDate: "2026-01-05", completed: false },
    { localDate: "2026-01-06", completed: false },
    { localDate: "2026-01-07", completed: false },
  ],
};

const settings = {
  dailyReviewTarget: 20,
  reviewIntervalPreset: "vocanova_default",
  appLanguage: "en",
  notificationsEnabled: true,
  marketingEmailsEnabled: false,
  displayName: "Accessibility Fixture",
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
    const cookies = parseCookies(req.headers.cookie);
    const onboardingStatusOverride = cookies.e2e_onboarding_status;
    const responseBody = ONBOARDING_STATUSES.has(onboardingStatusOverride)
      ? { ...currentUser, onboardingStatus: onboardingStatusOverride }
      : currentUser;
    logLine(req, 200, { onboardingStatus: responseBody.onboardingStatus });
    jsonResponse(res, 200, responseBody);
    return;
  }

  if (req.method === "GET" && url.pathname === "/api/v1/onboarding") {
    logLine(req, 200);
    jsonResponse(res, 200, onboardingProfile);
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

  if (req.method === "GET" && url.pathname === "/api/v1/journey-situations") {
    logLine(req, 200);
    jsonResponse(res, 200, journeySituations);
    return;
  }

  if (
    req.method === "GET" &&
    url.pathname.startsWith("/api/v1/journey-situations/")
  ) {
    const slug = decodeURIComponent(
      url.pathname.slice("/api/v1/journey-situations/".length),
    );
    const situation = journeySituationsBySlug[slug];
    if (!situation) {
      logLine(req, 404, { slug });
      jsonResponse(res, 404, { error: "not_found", slug });
      return;
    }
    logLine(req, 200, { slug });
    jsonResponse(res, 200, situation);
    return;
  }

  if (
    req.method === "GET" &&
    url.pathname.startsWith("/api/v1/canonical-words/")
  ) {
    const slug = decodeURIComponent(
      url.pathname.slice("/api/v1/canonical-words/".length),
    );
    const word = canonicalWordsBySlug[slug];
    if (!word) {
      logLine(req, 404, { slug });
      jsonResponse(res, 404, { error: "not_found", slug });
      return;
    }
    logLine(req, 200, { slug });
    jsonResponse(res, 200, word);
    return;
  }

  if (req.method === "GET" && url.pathname === "/api/v1/progress") {
    logLine(req, 200);
    jsonResponse(res, 200, progress);
    return;
  }

  if (req.method === "GET" && url.pathname === "/api/v1/settings") {
    logLine(req, 200);
    jsonResponse(res, 200, settings);
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
