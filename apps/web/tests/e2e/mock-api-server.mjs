// VOC-031-T07a + VOC-031-T07b + VOC-031-T08 mock API server for the
// e2e harness.
//
// T07a added the minimum canned JSON the Home page server
// component needs to render. T07b extended the server so every
// remaining core-loop page (Discover, Discover/[situation],
// Discover/[situation]/[word], Reviews, Progress, Onboarding,
// Settings, Settings/account) could render with deterministic
// fixture data - the T07b scans do not exercise form submissions
// or POST endpoints, so that server returned read-only fixtures
// and 404'd on writes.
//
// T08 (this task) extends the server with the full mutation
// surface the DOC-10 §7 end-to-end flow needs against the same
// Playwright install: magic-link request/consume (issuing a
// session cookie + double-submit CSRF cookie), onboarding
// completion, save / unsave, review submission, deterministic
// AI sentence feedback, settings PATCH, and logout. The session
// cookie is `vocanova_session`; the CSRF cookie is
// `vocanova_csrf`. Mutations (POST/PATCH/DELETE) require the
// `X-CSRF-Token` header to match the `vocanova_csrf` cookie
// value, exactly as the real backend's `CSRFMiddleware` does
// (apps/api/app/api/middleware.go). The /api/v1/auth/* routes
// are exempt because they predate / establish the session.
//
// State (saved words, completed onboarding, per-session
// settings, per-session progress) is tracked in memory keyed by
// the session cookie value. The server is single-process and
// the T08 suite runs as a single worker (`workers: 1` in
// playwright.config.ts), so the in-memory state is consistent
// across the test run and is reset on server restart. The
// existing T07a/T07b scan specs do not mutate state, so the
// per-session default fixtures continue to apply to them.
//
// AI feedback is deterministic: a fixed rule table based on the
// input sentence (length, presence of the target word, blank).
// This is the "deterministic AI adapter, not a
// paid/nondeterministic provider call" pattern DOC-10 §7
// requires for CI.
//
// Endpoints:
//
//   GET    /healthz                                 -> 200 { status: "ok" }
//
//   GET    /api/v1/me                                -> 200 CurrentUser
//             401 if ?fail=me (T07b auth-gate path)
//             onboardingStatus overridden by the
//             `e2e_onboarding_status` cookie when present
//
//   GET    /api/v1/onboarding                        -> 200 OnboardingProfile
//   POST   /api/v1/onboarding                        -> 200 OnboardingProfile
//                                                       (sets onboardingStatus
//                                                        = "completed" in /me)
//
//   POST   /api/v1/auth/magic-links                  -> 200 (no CSRF check)
//   POST   /api/v1/auth/magic-links/consume          -> 200 CurrentUser
//                                                       (sets session + CSRF
//                                                        cookies; exempt from
//                                                        CSRF check)
//   POST   /api/v1/auth/oauth/google/start          -> 200 { url: ... }
//   POST   /api/v1/auth/logout                       -> 204
//                                                       (clears session cookie)
//
//   GET    /api/v1/user-words                        -> 200 { items, nextCursor }
//   POST   /api/v1/user-words                        -> 200 SavedMeaning
//   DELETE /api/v1/user-words/:meaningId             -> 204
//
//   GET    /api/v1/reviews/due                       -> 200 { items, nextCursor, totalCount }
//   POST   /api/v1/reviews/submissions               -> 200 ReviewAttempt
//
//   POST   /api/v1/sentence-feedback                 -> 200 SentenceFeedbackResult
//                                                       (deterministic rule table)
//   POST   /api/v1/sentence-feedback/:attemptId/reports -> 204
//
//   GET    /api/v1/daily-mission                     -> 200 DailyMission
//   GET    /api/v1/progress                          -> 200 Progress
//
//   GET    /api/v1/journey-situations                -> 200 { items: [...] }
//   GET    /api/v1/journey-situations/:slug          -> 200 SituationResponse
//   GET    /api/v1/canonical-words/:slug             -> 200 WordDetailResponse
//
//   GET    /api/v1/settings                          -> 200 Settings
//   PATCH  /api/v1/settings                          -> 200 Settings
//
//   POST   /api/v1/settings/email-change-links       -> 204
//   POST   /api/v1/settings/email-change-links/consume -> 200 ConsumeEmailChangeLinkResult
//
//   POST   /api/v1/account-deletion-requests         -> 200 CreateAccountDeletionRequestResult
//
// Anything else returns 404. Mutations without a matching
// X-CSRF-Token return 403 (matches the backend's
// CSRFMiddleware). Mutations on /api/v1/auth/* are exempt.
// The server logs each request to stderr so a CI failure's
// log can be matched against the harness's expectations
// without enabling extra debug output.

import { createServer } from "node:http";

const PORT = Number(process.env.MOCK_API_PORT ?? 8080);
const HOST = process.env.MOCK_API_HOST ?? "127.0.0.1";

const ONBOARDING_STATUSES = new Set(["not_started", "in_progress", "completed"]);

const SESSION_COOKIE_NAME = "vocanova_session";
const CSRF_COOKIE_NAME = "vocanova_csrf";
const SESSION_DEFAULT_VALUE = "test-session-default";

const DEFAULT_USER = {
  email: "core-loop-fixture@example.test",
  displayName: "Core Loop Fixture",
  emailVerifiedAt: "2026-01-01T00:00:00Z",
};

const DEFAULT_SETTINGS = {
  dailyReviewTarget: 20,
  reviewIntervalPreset: "vocanova_default",
  appLanguage: "en",
  notificationsEnabled: true,
  marketingEmailsEnabled: false,
  displayName: DEFAULT_USER.displayName,
};

const DEFAULT_PROGRESS = {
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

const DEFAULT_DAILY_MISSION = {
  localDate: "2026-01-01",
  timezone: "UTC",
  reviewTarget: 20,
  reviewsCompleted: 0,
  newWordTarget: 5,
  newWordsCompleted: 0,
  sentencePracticeTarget: 3,
  sentencePracticesCompleted: 0,
  policyVersion: "t08-fixture-v1",
  status: "open",
  graceApplied: false,
  streak: DEFAULT_PROGRESS.streak,
};

const CANONICAL_WORDS = {
  pour: {
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
};

const JOURNEY_SITUATIONS = [
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
];

const SITUATIONS_BY_SLUG = {
  "ordering-at-a-cafe": {
    situation: JOURNEY_SITUATIONS[0],
    meanings: [
      {
        meaningId: "mean-pour",
        wordId: CANONICAL_WORDS.pour.id,
        wordSlug: CANONICAL_WORDS.pour.slug,
        wordText: CANONICAL_WORDS.pour.text,
        partOfSpeech: "verb",
        shortDefinition: "to make liquid flow into a container",
        saved: false,
      },
    ],
  },
  "navigating-an-airport": {
    situation: JOURNEY_SITUATIONS[1],
    meanings: [],
  },
};

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

function readJsonBody(req) {
  return new Promise((resolve, reject) => {
    const chunks = [];
    req.on("data", (chunk) => chunks.push(chunk));
    req.on("end", () => {
      const raw = Buffer.concat(chunks).toString("utf8");
      if (!raw) {
        resolve({});
        return;
      }
      try {
        resolve(JSON.parse(raw));
      } catch (error) {
        reject(error);
      }
    });
    req.on("error", reject);
  });
}

function jsonResponse(res, status, body, extraHeaders) {
  const payload = body === undefined ? "" : JSON.stringify(body);
  const headers = {
    "Content-Type": "application/json; charset=utf-8",
    "Cache-Control": "no-store",
  };
  if (body !== undefined) {
    headers["Content-Length"] = Buffer.byteLength(payload);
  }
  if (extraHeaders) {
    for (const [key, value] of Object.entries(extraHeaders)) {
      headers[key] = value;
    }
  }
  res.writeHead(status, headers);
  res.end(payload);
}

function emptyResponse(res, status, extraHeaders) {
  const headers = {
    "Cache-Control": "no-store",
  };
  if (extraHeaders) {
    for (const [key, value] of Object.entries(extraHeaders)) {
      headers[key] = value;
    }
  }
  res.writeHead(status, headers);
  res.end();
}

function buildSessionCookie(value) {
  // `HttpOnly` is intentionally omitted because the frontend's
  // `getCookieValue("vocanova_session")` is never called - the
  // frontend only reads the CSRF cookie. Session stays server-side
  // only. The real backend's session cookie is HttpOnly; mirroring
  // that here would force the mock to track an extra auth
  // relationship for the same security guarantee.
  return `${SESSION_COOKIE_NAME}=${value}; Path=/; SameSite=Lax`;
}

function buildCsrfCookie(value) {
  // CSRF cookie is intentionally NOT HttpOnly so the frontend can
  // read it via document.cookie and echo it back as X-CSRF-Token.
  return `${CSRF_COOKIE_NAME}=${value}; Path=/; SameSite=Lax`;
}

function buildClearSessionCookie() {
  return `${SESSION_COOKIE_NAME}=; Path=/; Max-Age=0; SameSite=Lax`;
}

// --- per-session mutable state ---------------------------------

const sessions = new Map();

function getSessionState(cookies) {
  const sessionId = cookies[SESSION_COOKIE_NAME] ?? SESSION_DEFAULT_VALUE;
  let state = sessions.get(sessionId);
  if (!state) {
    state = createInitialState();
    sessions.set(sessionId, state);
  }
  return state;
}

function createInitialState() {
  return {
    onboardingCompleted: true,
    savedMeaningIds: new Set(),
    // A meaning enters reviewedMeaningIds after a successful
    // review submission. The /api/v1/reviews/due response is
    // `savedMeaningIds - reviewedMeaningIds` so the review
    // session's "advance" refetch (which fires when the last
    // card is rated) returns an empty list and the page
    // transitions to the "all caught up" state, mirroring the
    // real backend's review-step advance.
    reviewedMeaningIds: new Set(),
    settings: { ...DEFAULT_SETTINGS },
    progress: cloneProgress(DEFAULT_PROGRESS),
    dailyMission: { ...DEFAULT_DAILY_MISSION },
    reviewAttempts: [],
    sentenceAttempts: [],
    lastReviewAttemptId: null,
    sentenceCount: 0,
    reviewedCount: 0,
  };
}

function cloneProgress(progress) {
  return {
    confidencePointsBalance: progress.confidencePointsBalance,
    streak: { ...progress.streak },
    completionHistory: progress.completionHistory.map((day) => ({ ...day })),
  };
}

function buildCurrentUser(state) {
  return {
    email: DEFAULT_USER.email,
    displayName: state.settings.displayName,
    emailVerifiedAt: DEFAULT_USER.emailVerifiedAt,
    onboardingStatus: state.onboardingCompleted ? "completed" : "not_started",
  };
}

function buildSettings(state) {
  return { ...state.settings };
}

function buildProgress(state) {
  return cloneProgress(state.progress);
}

function buildDailyMission(state) {
  const streak = { ...state.progress.streak };
  return {
    ...state.dailyMission,
    reviewsCompleted: state.reviewedCount,
    streak,
  };
}

function buildSavedWords(state) {
  const items = [];
  for (const meaningId of state.savedMeaningIds) {
    const word = CANONICAL_WORDS.pour;
    const meaning = word.meanings.find((m) => m.id === meaningId);
    if (!meaning) {
      continue;
    }
    items.push({
      userWordId: `uw-${meaningId}`,
      meaningId: meaning.id,
      wordId: word.id,
      wordSlug: word.slug,
      wordText: word.text,
      partOfSpeech: meaning.partOfSpeech,
      shortDefinition: meaning.shortDefinition,
      status: "saved",
      source: "journey",
      saved: true,
      addedAt: new Date().toISOString(),
    });
  }
  return { items, nextCursor: undefined };
}

function buildDueWords(state) {
  const items = [];
  for (const meaningId of state.savedMeaningIds) {
    if (state.reviewedMeaningIds.has(meaningId)) {
      continue;
    }
    const word = CANONICAL_WORDS.pour;
    const meaning = word.meanings.find((m) => m.id === meaningId);
    if (!meaning) {
      continue;
    }
    items.push({
      userWordId: `uw-${meaningId}`,
      meaningId: meaning.id,
      wordId: word.id,
      wordSlug: word.slug,
      wordText: word.text,
      partOfSpeech: meaning.partOfSpeech,
      shortDefinition: meaning.shortDefinition,
      status: "due",
      reviewStep: 0,
    });
  }
  return {
    items,
    nextCursor: undefined,
    totalCount: items.length,
  };
}

function buildWordDetailResponse(state, slug) {
  const word = CANONICAL_WORDS[slug];
  if (!word) {
    return null;
  }
  const meanings = word.meanings.map((meaning) => ({
    ...meaning,
    saved: state.savedMeaningIds.has(meaning.id),
    userWordId: state.savedMeaningIds.has(meaning.id)
      ? `uw-${meaning.id}`
      : undefined,
  }));
  return {
    word: {
      ...word,
      meanings,
    },
  };
}

function buildSituationResponse(slug) {
  const fixture = SITUATIONS_BY_SLUG[slug];
  if (!fixture) {
    return null;
  }
  return {
    situation: fixture.situation,
    meanings: fixture.meanings.map((meaning) => ({ ...meaning })),
  };
}

function buildJourneySituations() {
  return { items: JOURNEY_SITUATIONS.map((s) => ({ ...s })) };
}

function generateId(prefix) {
  return `${prefix}-${Math.random().toString(36).slice(2, 10)}-${Date.now().toString(36)}`;
}

// --- deterministic AI feedback --------------------------------

function evaluateSentenceFeedback({ sentence, targetWord }) {
  const trimmed = (sentence ?? "").trim();
  if (trimmed.length === 0) {
    return {
      status: "incorrect",
      errorCode: "too_short",
      errorMessage: "Your sentence is too short. Write at least 3 words.",
    };
  }
  const words = trimmed.split(/\s+/).filter(Boolean);
  if (words.length < 3) {
    return {
      status: "incorrect",
      errorCode: "too_short",
      errorMessage: "Your sentence is too short. Write at least 3 words.",
    };
  }
  if (trimmed.length > 300) {
    return {
      status: "needs_improvement",
      errorCode: "too_long",
      errorMessage: "Your sentence is too long. Keep it under 300 characters.",
    };
  }
  const containsTarget = targetWord
    ? new RegExp(`\\b${targetWord}\\b`, "i").test(trimmed)
    : true;
  if (!containsTarget) {
    return {
      status: "needs_improvement",
      errorCode: "missing_target",
      errorMessage: `Your sentence is missing the target word "${targetWord}".`,
    };
  }
  return {
    status: "correct",
    explanation: "Your sentence uses the target word naturally.",
  };
}

function applySettingsPatch(settings, patch) {
  const allowedKeys = new Set([
    "dailyReviewTarget",
    "reviewIntervalPreset",
    "appLanguage",
    "notificationsEnabled",
    "marketingEmailsEnabled",
    "displayName",
  ]);
  const result = { ...settings };
  for (const [key, value] of Object.entries(patch ?? {})) {
    if (allowedKeys.has(key)) {
      result[key] = value;
    }
  }
  if (result.dailyReviewTarget !== undefined) {
    if (
      typeof result.dailyReviewTarget !== "number" ||
      result.dailyReviewTarget < 5 ||
      result.dailyReviewTarget > 100
    ) {
      const error = new Error("dailyReviewTarget out of range");
      error.code = 400;
      throw error;
    }
  }
  if (
    result.reviewIntervalPreset !== undefined &&
    !["vocanova_default", "wordup_like", "custom"].includes(
      result.reviewIntervalPreset,
    )
  ) {
    const error = new Error("invalid reviewIntervalPreset");
    error.code = 400;
    throw error;
  }
  if (result.appLanguage !== undefined && result.appLanguage !== "en") {
    const error = new Error("appLanguage not supported");
    error.code = 400;
    throw error;
  }
  return result;
}

function logLine(req, status, extra) {
  const tag = extra ? ` ${JSON.stringify(extra)}` : "";
  process.stderr.write(
    `[mock-api] ${req.method} ${req.url} -> ${status}${tag}\n`,
  );
}

const server = createServer(async (req, res) => {
  if (!req.url) {
    res.writeHead(400);
    res.end();
    return;
  }

  const url = new URL(req.url, `http://${req.headers.host ?? `${HOST}:${PORT}`}`);
  const cookies = parseCookies(req.headers.cookie);

  if (req.method === "GET" && url.pathname === "/healthz") {
    logLine(req, 200);
    jsonResponse(res, 200, { status: "ok" });
    return;
  }

  // ----- auth (CSRF-exempt) ----------------------------------

  if (req.method === "POST" && url.pathname === "/api/v1/auth/magic-links") {
    logLine(req, 200);
    jsonResponse(res, 200, {});
    return;
  }

  if (
    req.method === "POST" &&
    url.pathname === "/api/v1/auth/magic-links/consume"
  ) {
    const sessionValue = generateId("session");
    const csrfValue = generateId("csrf");
    sessions.set(sessionValue, createInitialState());
    logLine(req, 200, { session: "issued" });
    jsonResponse(
      res,
      200,
      {
        ...DEFAULT_USER,
        onboardingStatus: "completed",
      },
      {
        "Set-Cookie": [
          buildSessionCookie(sessionValue),
          buildCsrfCookie(csrfValue),
        ].join(", "),
      },
    );
    return;
  }

  if (
    req.method === "POST" &&
    url.pathname === "/api/v1/auth/oauth/google/start"
  ) {
    logLine(req, 200);
    jsonResponse(res, 200, {
      url: "https://accounts.google.com/o/oauth2/v2/auth?redirected=true",
    });
    return;
  }

  if (req.method === "POST" && url.pathname === "/api/v1/auth/logout") {
    const sessionId = cookies[SESSION_COOKIE_NAME];
    if (sessionId) {
      sessions.delete(sessionId);
    }
    logLine(req, 204, { action: "logout" });
    emptyResponse(res, 204, { "Set-Cookie": buildClearSessionCookie() });
    return;
  }

  // ----- authenticated routes (CSRF enforced for mutations) --

  if (req.method === "GET" && url.pathname === "/api/v1/me") {
    if (url.searchParams.get("fail") === "me") {
      logLine(req, 401, { reason: "fixture-forced-401" });
      jsonResponse(res, 401, { error: "unauthorized" });
      return;
    }
    if (cookies.e2e_unauthenticated === "1") {
      // T08: the unauthenticated-access rejection step sets this
      // cookie after logout to make /api/v1/me return 401, so the
      // Next.js auth-gate middleware (apps/web/src/middleware.ts)
      // routes the learner to /signin. The cookie is unset by the
      // test before the next test that needs an authenticated
      // session, so the existing T07a/T07b scans continue to see
      // a 200 here without changing their own setup.
      logLine(req, 401, { reason: "e2e-unauthenticated-override" });
      jsonResponse(res, 401, { error: "unauthorized" });
      return;
    }
    const onboardingStatusOverride = cookies.e2e_onboarding_status;
    if (ONBOARDING_STATUSES.has(onboardingStatusOverride)) {
      logLine(req, 200, { onboardingStatus: onboardingStatusOverride });
      jsonResponse(res, 200, {
        ...DEFAULT_USER,
        onboardingStatus: onboardingStatusOverride,
      });
      return;
    }
    const state = getSessionState(cookies);
    const user = buildCurrentUser(state);
    logLine(req, 200, { onboardingStatus: user.onboardingStatus });
    jsonResponse(res, 200, user);
    return;
  }

  if (req.method === "GET" && url.pathname === "/api/v1/onboarding") {
    const state = getSessionState(cookies);
    logLine(req, 200, { completed: state.onboardingCompleted });
    jsonResponse(res, 200, {
      status: state.onboardingCompleted ? "completed" : "not_started",
      englishLevel: "a2",
      nativeLanguage: "es",
      learningGoal: "general",
      mainUseCase: "daily_life",
      dailyReviewTarget: state.settings.dailyReviewTarget,
      completedAt: state.onboardingCompleted
        ? new Date().toISOString()
        : undefined,
    });
    return;
  }

  if (req.method === "POST" && url.pathname === "/api/v1/onboarding") {
    if (!checkCsrf(req, cookies, res, logLine)) {
      return;
    }
    const state = getSessionState(cookies);
    let body = {};
    try {
      body = await readJsonBody(req);
    } catch {
      logLine(req, 400, { reason: "malformed-json" });
      jsonResponse(res, 400, { error: "invalid_json" });
      return;
    }
    state.onboardingCompleted = true;
    if (typeof body.dailyReviewTarget === "number") {
      // VOC-031-D04 seed rule: only seed dailyReviewTarget when no
      // customized value exists yet. The mock starts with the
      // schema default (20) for every fresh session, so the seed
      // fires for the very first onboarding write.
      if (state.settings.dailyReviewTarget === DEFAULT_SETTINGS.dailyReviewTarget) {
        state.settings.dailyReviewTarget = body.dailyReviewTarget;
      }
    }
    logLine(req, 200, { action: "complete-onboarding" });
    jsonResponse(res, 200, {
      status: "completed",
      englishLevel: body.englishLevel,
      nativeLanguage: body.nativeLanguage,
      learningGoal: body.learningGoal,
      mainUseCase: body.mainUseCase,
      dailyReviewTarget: state.settings.dailyReviewTarget,
      completedAt: new Date().toISOString(),
    });
    return;
  }

  if (req.method === "GET" && url.pathname === "/api/v1/user-words") {
    const state = getSessionState(cookies);
    const data = buildSavedWords(state);
    logLine(req, 200, { count: data.items.length });
    jsonResponse(res, 200, data);
    return;
  }

  if (req.method === "POST" && url.pathname === "/api/v1/user-words") {
    if (!checkCsrf(req, cookies, res, logLine)) {
      return;
    }
    let body = {};
    try {
      body = await readJsonBody(req);
    } catch {
      logLine(req, 400, { reason: "malformed-json" });
      jsonResponse(res, 400, { error: "invalid_json" });
      return;
    }
    const state = getSessionState(cookies);
    const meaningId = body.meaningId;
    if (!meaningId) {
      logLine(req, 400, { reason: "missing-meaning-id" });
      jsonResponse(res, 400, { error: "missing_meaning_id" });
      return;
    }
    state.savedMeaningIds.add(meaningId);
    logLine(req, 200, { action: "save", meaningId });
    const word = CANONICAL_WORDS.pour;
    const meaning = word.meanings.find((m) => m.id === meaningId);
    jsonResponse(res, 200, {
      userWordId: `uw-${meaningId}`,
      meaningId: meaning.id,
      wordId: word.id,
      wordSlug: word.slug,
      wordText: word.text,
      partOfSpeech: meaning.partOfSpeech,
      shortDefinition: meaning.shortDefinition,
      status: "saved",
      source: body.source ?? "journey",
      saved: true,
      addedAt: new Date().toISOString(),
    });
    return;
  }

  if (
    req.method === "DELETE" &&
    url.pathname.startsWith("/api/v1/user-words/")
  ) {
    if (!checkCsrf(req, cookies, res, logLine)) {
      return;
    }
    const meaningId = decodeURIComponent(
      url.pathname.slice("/api/v1/user-words/".length),
    );
    const state = getSessionState(cookies);
    state.savedMeaningIds.delete(meaningId);
    logLine(req, 204, { action: "unsave", meaningId });
    emptyResponse(res, 204);
    return;
  }

  if (req.method === "GET" && url.pathname === "/api/v1/reviews/due") {
    const state = getSessionState(cookies);
    const data = buildDueWords(state);
    logLine(req, 200, { count: data.items.length });
    jsonResponse(res, 200, data);
    return;
  }

  if (req.method === "POST" && url.pathname === "/api/v1/reviews/submissions") {
    if (!checkCsrf(req, cookies, res, logLine)) {
      return;
    }
    let body = {};
    try {
      body = await readJsonBody(req);
    } catch {
      logLine(req, 400, { reason: "malformed-json" });
      jsonResponse(res, 400, { error: "invalid_json" });
      return;
    }
    const state = getSessionState(cookies);
    const attemptId = generateId("att");
    const reviewedMeaningId = body.meaningId;
    if (reviewedMeaningId) {
      state.reviewedMeaningIds.add(reviewedMeaningId);
    }
    state.reviewedCount += 1;
    state.lastReviewAttemptId = attemptId;
    state.reviewAttempts.push({
      attemptId,
      meaningId: reviewedMeaningId,
      result: body.result,
      rating: body.rating,
      answeredAt: new Date().toISOString(),
    });
    // The real backend advances the word's review schedule but
    // keeps it in the saved set; the T08 "review -> sentence"
    // step relies on the word remaining in savedMeaningIds after
    // the review submission so the sentence-feedback widget on
    // /discover/.../[word] still renders the saved meaning.
    logLine(req, 200, {
      action: "submit-review",
      meaningId: reviewedMeaningId,
      reviewedCount: state.reviewedCount,
    });
    jsonResponse(res, 200, {
      attemptId,
      userWordId: body.userWordId,
      meaningId: body.meaningId,
      attemptType: "review",
      promptType: body.promptType,
      result: body.result,
      rating: body.rating,
      reviewStepBefore: 0,
      reviewStepAfter: 1,
      answeredAt: new Date().toISOString(),
      responseTimeMs: body.responseTimeMs ?? 0,
      selectedOptionMeaningId: body.selectedOptionMeaningId,
      wasHintUsed: body.wasHintUsed ?? false,
      source: body.source ?? "review_session",
      clientAttemptId: body.clientAttemptId,
      nextReviewAt: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
    });
    return;
  }

  if (req.method === "POST" && url.pathname === "/api/v1/sentence-feedback") {
    if (!checkCsrf(req, cookies, res, logLine)) {
      return;
    }
    let body = {};
    try {
      body = await readJsonBody(req);
    } catch {
      logLine(req, 400, { reason: "malformed-json" });
      jsonResponse(res, 400, { error: "invalid_json" });
      return;
    }
    const state = getSessionState(cookies);
    const sentenceId = generateId("sent");
    const targetWord = body.attemptId
      ? lookupTargetWord(body.attemptId)
      : "pour";
    const evaluation = evaluateSentenceFeedback({
      sentence: body.sentenceText,
      targetWord,
    });
    state.sentenceAttempts.push({
      sentenceId,
      attemptId: body.attemptId,
      sentenceText: body.sentenceText,
      evaluation,
    });
    if (!evaluation.errorCode) {
      state.sentenceCount += 1;
    }
    logLine(req, 200, {
      action: "submit-sentence",
      status: evaluation.status,
      errorCode: evaluation.errorCode,
    });
    jsonResponse(res, 200, {
      sentenceId,
      attemptId: body.attemptId,
      status: evaluation.status,
      originalSentence: body.sentenceText,
      correctedSentence: evaluation.status === "correct" ? body.sentenceText : undefined,
      explanation: evaluation.explanation,
      improvementTip: evaluation.status === "needs_improvement"
        ? "Try using the target word naturally in your sentence."
        : undefined,
      missionCompleted: !evaluation.errorCode,
      canRetry: Boolean(evaluation.errorCode),
      reported: false,
      errorCode: evaluation.errorCode,
      errorMessage: evaluation.errorMessage,
    });
    return;
  }

  if (
    req.method === "POST" &&
    url.pathname.startsWith("/api/v1/sentence-feedback/") &&
    url.pathname.endsWith("/reports")
  ) {
    if (!checkCsrf(req, cookies, res, logLine)) {
      return;
    }
    logLine(req, 204, { action: "report-sentence" });
    emptyResponse(res, 204);
    return;
  }

  if (req.method === "GET" && url.pathname === "/api/v1/daily-mission") {
    const state = getSessionState(cookies);
    const mission = buildDailyMission(state);
    logLine(req, 200, { reviewsCompleted: mission.reviewsCompleted });
    jsonResponse(res, 200, mission);
    return;
  }

  if (req.method === "GET" && url.pathname === "/api/v1/progress") {
    const state = getSessionState(cookies);
    logLine(req, 200, { reviewedCount: state.reviewedCount });
    jsonResponse(res, 200, buildProgress(state));
    return;
  }

  if (req.method === "GET" && url.pathname === "/api/v1/journey-situations") {
    logLine(req, 200);
    jsonResponse(res, 200, buildJourneySituations());
    return;
  }

  if (
    req.method === "GET" &&
    url.pathname.startsWith("/api/v1/journey-situations/")
  ) {
    const slug = decodeURIComponent(
      url.pathname.slice("/api/v1/journey-situations/".length),
    );
    const response = buildSituationResponse(slug);
    if (!response) {
      logLine(req, 404, { slug });
      jsonResponse(res, 404, { error: "not_found", slug });
      return;
    }
    logLine(req, 200, { slug });
    jsonResponse(res, 200, response);
    return;
  }

  if (
    req.method === "GET" &&
    url.pathname.startsWith("/api/v1/canonical-words/")
  ) {
    const slug = decodeURIComponent(
      url.pathname.slice("/api/v1/canonical-words/".length),
    );
    const state = getSessionState(cookies);
    const response = buildWordDetailResponse(state, slug);
    if (!response) {
      logLine(req, 404, { slug });
      jsonResponse(res, 404, { error: "not_found", slug });
      return;
    }
    logLine(req, 200, { slug });
    jsonResponse(res, 200, response);
    return;
  }

  if (req.method === "GET" && url.pathname === "/api/v1/settings") {
    const state = getSessionState(cookies);
    logLine(req, 200);
    jsonResponse(res, 200, buildSettings(state));
    return;
  }

  if (req.method === "PATCH" && url.pathname === "/api/v1/settings") {
    if (!checkCsrf(req, cookies, res, logLine)) {
      return;
    }
    let body = {};
    try {
      body = await readJsonBody(req);
    } catch {
      logLine(req, 400, { reason: "malformed-json" });
      jsonResponse(res, 400, { error: "invalid_json" });
      return;
    }
    const state = getSessionState(cookies);
    try {
      state.settings = applySettingsPatch(state.settings, body);
    } catch (error) {
      const status = error.code === 400 ? 400 : 500;
      logLine(req, status, { reason: "invalid-settings-patch" });
      jsonResponse(res, status, { error: error.message });
      return;
    }
    logLine(req, 200, { action: "patch-settings" });
    jsonResponse(res, 200, buildSettings(state));
    return;
  }

  if (
    req.method === "POST" &&
    url.pathname === "/api/v1/settings/email-change-links"
  ) {
    if (!checkCsrf(req, cookies, res, logLine)) {
      return;
    }
    logLine(req, 204, { action: "request-email-change" });
    emptyResponse(res, 204);
    return;
  }

  if (
    req.method === "POST" &&
    url.pathname === "/api/v1/settings/email-change-links/consume"
  ) {
    if (!checkCsrf(req, cookies, res, logLine)) {
      return;
    }
    logLine(req, 200, { action: "consume-email-change" });
    jsonResponse(res, 200, {
      email: DEFAULT_USER.email,
      previousEmail: DEFAULT_USER.email,
      changedAt: new Date().toISOString(),
    });
    return;
  }

  if (
    req.method === "POST" &&
    url.pathname === "/api/v1/account-deletion-requests"
  ) {
    if (!checkCsrf(req, cookies, res, logLine)) {
      return;
    }
    const sessionId = cookies[SESSION_COOKIE_NAME];
    if (sessionId) {
      sessions.delete(sessionId);
    }
    logLine(req, 200, { action: "account-deletion" });
    const requestedAt = new Date();
    const purgeAfter = new Date(requestedAt.getTime() + 30 * 24 * 60 * 60 * 1000);
    jsonResponse(
      res,
      200,
      {
        status: "deactivated",
        userId: "user-fixture",
        requestedAt: requestedAt.toISOString(),
        purgeAfter: purgeAfter.toISOString(),
        idempotencyKey: req.headers["idempotency-key"] ?? "",
        replayed: false,
      },
      { "Set-Cookie": buildClearSessionCookie() },
    );
    return;
  }

  logLine(req, 404);
  jsonResponse(res, 404, { error: "not_found", path: url.pathname });
});

function lookupTargetWord(attemptId) {
  // attemptId == userWordId == "uw-<meaningId>" for this fixture.
  if (typeof attemptId !== "string") {
    return "pour";
  }
  const meaningId = attemptId.startsWith("uw-")
    ? attemptId.slice("uw-".length)
    : attemptId;
  const word = CANONICAL_WORDS.pour;
  const meaning = word.meanings.find((m) => m.id === meaningId);
  return meaning ? word.text : "pour";
}

function checkCsrf(req, cookies, res, logLineRef) {
  const method = req.method ?? "";
  if (!["POST", "PUT", "PATCH", "DELETE"].includes(method)) {
    return true;
  }
  const csrfCookie = cookies[CSRF_COOKIE_NAME];
  const headerToken = req.headers["x-csrf-token"];
  if (csrfCookie && headerToken && csrfCookie === headerToken) {
    return true;
  }
  logLineRef(req, 403, { reason: "csrf-failed" });
  jsonResponse(res, 403, { error: "invalid_csrf_token" });
  return false;
}

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
