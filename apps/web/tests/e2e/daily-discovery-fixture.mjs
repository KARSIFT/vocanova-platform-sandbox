// Synthetic transport state for browser interaction. PostgreSQL tests establish
// real grading, scheduling, ownership and daily credit; this fixture does not.
import { createHash, randomUUID } from "node:crypto";
import { readFileSync } from "node:fs";

const attribution = {
  title:
    "FrequencyWords / OpenSubtitles 2018, WordNet 3.0 and Vocanova editorial content",
  sourceUrl: "https://github.com/hermitdave/FrequencyWords",
  licenseUrl: "https://creativecommons.org/licenses/by-sa/4.0/",
  wordnetLicense: readFileSync(
    new URL(
      "../../../api/business/knowledgecatalog/LICENSE.wordnet",
      import.meta.url,
    ),
    "utf8",
  ),
};
const revision = (value) => String(value).padStart(64, "0");

export async function handleDailyDiscovery({
  req,
  res,
  url,
  cookies,
  state,
  jsonResponse,
  readJsonBody,
  checkCsrf,
  logLine,
  dueWords,
}) {
  const path = url.pathname;
  const removal =
    req.method === "DELETE" &&
    /^\/api\/v1\/user-words\/map-meaning-(\d+)$/.exec(path);
  if (
    !removal &&
    !path.startsWith("/api/v1/knowledge-map") &&
    !path.startsWith("/api/v1/daily-lessons")
  )
    return false;
  const send = (status, body) => {
    jsonResponse(res, status, body);
    return true;
  };
  // Follow the mock's existing default-learner convention. The central
  // e2e_unauthenticated override exercises a signed-out response for all routes.
  state.mapDecisions ??= new Map();
  state.dailySessions ??= new Map();
  state.dailyReceipts ??= new Map();
  state.dailyStartReceipts ??= new Map();
  state.mapReceipts ??= new Map();
  state.dailyDeferredMeaningIds ??= new Set();
  state.dailyReviewSchedules ??= new Map();
  // The complete-map case avoids 10,000 setup writes while keeping its range
  // metadata and individual word states consistent. Ordinary tests use CAS.
  const completeMap = cookies.e2e_knowledge_map === "complete";
  // Simulate another tab finishing range2 between the two server reads. After
  // that single change the fixture stays stable, so explicit retry can recover.
  const rolloverRace = cookies.e2e_knowledge_map === "rollover-race";
  if (
    rolloverRace &&
    req.method === "GET" &&
    path === "/api/v1/knowledge-map" &&
    url.searchParams.get("band") === "2"
  )
    state.mapRolloverRaceResolved = true;
  const completeThrough = completeMap
    ? 100
    : cookies.e2e_knowledge_map === "first-40-known"
      ? 40
      : rolloverRace
        ? state.mapRolloverRaceResolved
          ? 2
          : 1
        : 0;
  const legacyProgress = cookies.e2e_knowledge_map === "legacy-progress";
  const mixedProgress =
    legacyProgress || cookies.e2e_knowledge_map === "mixed-progress"
      ? {
          1: { decision: "known", revision: 0 },
          3: { decision: "learn", revision: 0 },
          12: { decision: "learn", learned: true, revision: 0 },
          105: { decision: "learn", learned: true, revision: 0 },
        }
      : {};
  const mapEntry = (rank) =>
    state.mapDecisions.get(rank) ??
    mixedProgress[rank] ??
    (rank <= completeThrough * 100
      ? { decision: "known", revision: 0 }
      : cookies.e2e_knowledge_map === "gapped-ranges" &&
          ((rank >= 1101 && rank <= 1302 && rank !== 1200) ||
            (rank >= 9901 && rank <= 9998))
        ? { decision: rank === 1302 ? "learn" : "known", revision: 0 }
        : undefined);
  const mapCode = (rank) => {
    const entry = mapEntry(rank);
    if (entry?.decision === "known") return "k";
    if (entry?.decision === "learn") return entry.learned ? "m" : "l";
    return "n";
  };
  const map = (band) => ({
    attribution,
    catalogVersion: "synthetic-browser-v1",
    totalWords: 10000,
    bandSize: 100,
    ordering: {
      title: "Common words first",
      description:
        "Synthetic interface fixture. The released catalog uses selected English subtitle frequency, not a personalized or CEFR ranking.",
      sourceUrl: attribution.sourceUrl,
    },
    bands: Array.from({ length: 100 }, (_, index) => {
      const number = index + 1;
      const wordStates = Array.from({ length: 100 }, (_, i) =>
        mapCode(index * 100 + i + 1),
      ).join("");
      const knownCount = wordStates.match(/k/g)?.length ?? 0;
      const learningCount = wordStates.match(/[lm]/g)?.length ?? 0;
      return {
        number,
        startRank: index * 100 + 1,
        endRank: number * 100,
        totalWords: 100,
        knownCount,
        learningCount,
        unseenCount: 100 - knownCount - learningCount,
        ...(!legacyProgress && (knownCount || learningCount)
          ? { wordStates }
          : {}),
      };
    }),
    currentBand: {
      number: band,
      items: Array.from({ length: 100 }, (_, index) => {
        const rank = (band - 1) * 100 + index + 1;
        const entry = mapEntry(rank);
        return {
          rank,
          meaningId: `map-meaning-${rank}`,
          wordId: `map-word-${rank}`,
          wordText: rank === 1 ? "the" : rank === 2 ? "be" : `Map word ${rank}`,
          wordSlug: `map-word-${rank}`,
          partOfSpeech: rank === 1 ? "determiner" : "verb",
          shortDefinition:
            rank === 1
              ? "Used before a person or thing that is already clear."
              : "A synthetic meaning for this browser fixture.",
          // One realistic example keeps the card as long as real content.
          ...(rank === 1
            ? { exampleSentence: "The bus to the station is late again." }
            : {}),
          selfReportedKnown: entry?.decision === "known",
          saved: entry?.decision === "learn",
          ...(!legacyProgress && entry?.learned ? { learned: true } : {}),
          due: entry?.decision === "learn" && !entry.learned,
          revision: revision(entry?.revision ?? 0),
        };
      }),
    },
  });
  // Like the API's decision result: the meaning as it is now.
  const decisionState = (rank, decision) => {
    const entry = mapEntry(rank);
    return {
      meaningId: `map-meaning-${rank}`,
      decision,
      revision: revision(entry?.revision ?? 0),
      selfReportedKnown: entry?.decision === "known",
      saved: entry?.decision === "learn",
      ...(!legacyProgress && entry?.learned ? { learned: true } : {}),
      due: entry?.decision === "learn" && !entry.learned,
    };
  };
  if (removal) {
    if (!checkCsrf(req, cookies, res, logLine)) return true;
    const rank = Number(removal[1]);
    const entry = mapEntry(rank);
    // Removing a saved word leaves no learning or known state behind.
    if (entry?.decision === "learn")
      state.mapDecisions.set(rank, {
        decision: "removed",
        revision: entry.revision + 1,
      });
    return send(200, {});
  }
  if (req.method === "GET" && path === "/api/v1/knowledge-map") {
    if (cookies.e2e_knowledge_map === "unavailable")
      return send(503, { detail: "Map unavailable" });
    // The identity check still passes; only this read reports expiry, so the
    // page's own sign-in return address is observable.
    if (cookies.e2e_knowledge_map === "expired")
      return send(401, { detail: "Session expired" });
    const band = Number(url.searchParams.get("band") ?? 1);
    if (!Number.isInteger(band) || band < 1 || band > 100)
      return send(400, { detail: "Invalid range" });
    const displayed = map(band);
    // Cookie-scoped read projection works through SSR range navigation too.
    // map() creates fresh items; canonical fixture state and CAS tokens stay put.
    if (cookies.e2e_learner_explanation === "true") {
      for (const item of displayed.currentBand.items.slice(0, 2)) {
        item.learnerDefinition = `A plain explanation for the selected meaning at position ${item.rank}.`;
        item.meaningRevision = createHash("sha256")
          .update(
            JSON.stringify([
              item.meaningId,
              item.wordText,
              item.partOfSpeech,
              item.shortDefinition,
            ]),
          )
          .digest("hex");
      }
    }
    return send(200, displayed);
  }
  if (req.method === "GET" && path === "/api/v1/daily-lessons/current") {
    if (cookies.e2e_daily_lesson_auth === "expired")
      return send(401, { detail: "Session expired" });
    return send(200, {
      session:
        [...state.dailySessions.values()].find(
          (s) => s.status === "in_progress",
        ) ?? null,
      dueCount: dueWords().totalCount,
    });
  }
  const sessionMatch = path.match(
    /^\/api\/v1\/daily-lessons\/([^/]+)(\/actions)?$/,
  );
  if (req.method === "GET" && sessionMatch)
    return state.dailySessions.has(sessionMatch[1])
      ? send(200, state.dailySessions.get(sessionMatch[1]))
      : send(404, { detail: "Lesson not found" });
  if (req.method !== "POST") return send(404, { detail: "Not found" });
  if (!checkCsrf(req, cookies, res, logLine)) return true;
  const key = req.headers["idempotency-key"];
  if (!key) return send(400, { detail: "A retry key is required" });
  const body = await readJsonBody(req);
  if (path === "/api/v1/knowledge-map/decisions") {
    const prior = state.mapReceipts.get(key);
    // A replayed receipt reports the meaning's current state with the original
    // decision, as the API does; it never re-applies an older choice.
    if (prior)
      return prior.body === JSON.stringify(body)
        ? send(200, decisionState(prior.rank, body.decision))
        : send(409, { detail: "Conflicting retry" });
    const rank = Number(body.meaningId?.replace("map-meaning-", ""));
    const entry = mapEntry(rank);
    if (
      !Number.isInteger(rank) ||
      rank < 1 ||
      rank > 10000 ||
      !["known", "learn"].includes(body.decision)
    )
      return send(400, { detail: "Invalid choice" });
    if (body.expectedRevision !== revision(entry?.revision ?? 0))
      return send(409, { detail: "Word changed" });
    const next = (entry?.revision ?? 0) + 1;
    state.mapDecisions.set(rank, { decision: body.decision, revision: next });
    state.mapReceipts.set(key, { body: JSON.stringify(body), rank });
    return send(200, decisionState(rank, body.decision));
  }
  const step = (word) => ({
    id: randomUUID(),
    kind: "meaning_choice",
    wordText: word.wordText,
    wordSlug: word.wordSlug,
    partOfSpeech: word.partOfSpeech,
    choices: [
      { id: "meaning-a", text: word.shortDefinition },
      { id: "meaning-b", text: "A different meaning." },
      { id: "meaning-c", text: "Another unrelated meaning." },
      { id: "meaning-d", text: "Something else entirely." },
    ],
  });
  if (path === "/api/v1/daily-lessons") {
    const replayId = state.dailyStartReceipts.get(key);
    if (replayId) return send(200, state.dailySessions.get(replayId));
    const active = [...state.dailySessions.values()].find(
      (s) => s.status === "in_progress",
    );
    if (active) {
      state.dailyStartReceipts.set(key, active.id);
      return send(200, active);
    }
    const words = dueWords().items.slice(0, 10);
    if (!words.length) {
      jsonResponse(
        res,
        409,
        {
          status: 409,
          detail: "No words are due. Discover a word to start learning.",
        },
        { "Content-Type": "application/problem+json" },
      );
      return true;
    }
    const now = new Date().toISOString();
    const session = {
      attribution,
      id: randomUUID(),
      status: "in_progress",
      revision: 1,
      completedWords: 0,
      totalWords: words.length,
      firstAnswersCorrect: 0,
      questionsAnswered: 0,
      creditedToday: 0,
      currentStep: step(words[0]),
      feedback: null,
      canContinue: false,
      createdAt: now,
      updatedAt: now,
    };
    state.dailyWords ??= new Map();
    state.dailyRetries ??= new Map();
    state.dailyWords.set(session.id, words);
    state.dailyRetries.set(session.id, new Set());
    state.dailySessions.set(session.id, session);
    state.dailyStartReceipts.set(key, session.id);
    return send(200, session);
  }
  if (!sessionMatch?.[2]) return send(404, { detail: "Not found" });
  const sessionId = sessionMatch[1];
  const session = state.dailySessions.get(sessionId);
  const fingerprint = JSON.stringify([sessionId, body]);
  const prior = state.dailyReceipts.get(key);
  if (prior)
    return prior.fingerprint === fingerprint
      ? send(200, prior.result)
      : send(409, { detail: "Conflicting retry" });
  if (!session) return send(404, { detail: "Lesson not found" });
  if (
    body.expectedRevision !== session.revision ||
    body.stepId !== session.currentStep?.id ||
    session.status !== "in_progress"
  )
    return send(409, { detail: "Lesson changed" });
  const words = state.dailyWords.get(sessionId);
  const word = words[session.completedWords];
  const retries = state.dailyRetries.get(sessionId);
  if (body.action === "continue" && session.feedback) {
    const manual =
      body.nextReview !== undefined && body.nextReview !== "automatic";
    const intervals = {
      "1d": 1,
      "3d": 3,
      "7d": 7,
      "14d": 14,
      "30d": 30,
      "60d": 60,
      "120d": 120,
    };
    if (
      manual &&
      ((!intervals[body.nextReview] && body.nextReview !== "known") ||
        session.feedback.skipped ||
        cookies.e2e_daily_schedule === "legacy")
    )
      return send(400, { detail: "Invalid review schedule" });
    if (manual) {
      state.dailyDeferredMeaningIds.add(word.meaningId);
      state.dailyReviewSchedules.set(word.meaningId, {
        choice: body.nextReview,
        ...(body.nextReview === "known"
          ? { known: true }
          : {
              nextReviewAt: new Date(
                Date.now() + intervals[body.nextReview] * 86400000,
              ).toISOString(),
            }),
      });
    }
    // Every outcome finishes this word for today's review. Practice is separate.
    session.completedWords += 1;
    session.feedback = null;
    session.canContinue = false;
    if (session.completedWords === session.totalWords) {
      session.status = "completed";
      session.currentStep = null;
      session.completedAt = new Date().toISOString();
    } else session.currentStep = step(words[session.completedWords]);
  } else if (["answer", "skip"].includes(body.action) && !session.feedback) {
    const correct = body.action === "answer" && body.choiceId === "meaning-a";
    state.dailyDeferredMeaningIds.add(word.meaningId);
    if (body.action === "answer") {
      session.questionsAnswered += 1;
      if (correct && !retries.has(word.meaningId))
        session.firstAnswersCorrect += 1;
      retries.add(word.meaningId);
    }
    if (correct && !state.reviewedMeaningIds.has(word.meaningId)) {
      session.creditedToday += 1;
      state.reviewedMeaningIds.add(word.meaningId);
      state.reviewedCount += 1;
    }
    session.feedback = {
      ...(body.action === "answer" && cookies.e2e_daily_schedule !== "legacy"
        ? {
            reviewStep: correct ? 1 : 0,
            nextReviewAt:
              cookies.e2e_daily_review_at === "2026-10-18T22:30:00.000Z"
                ? cookies.e2e_daily_review_at
                : new Date(Date.now() + 86400000).toISOString(),
          }
        : {}),
      stepId: session.currentStep.id,
      correct,
      answer: word.shortDefinition,
      explanation: "",
      meaningId: word.meaningId,
      wordText: word.wordText,
      credited: correct,
      skipped: body.action === "skip",
      reason: body.action === "skip" ? "learner_skip" : undefined,
      correctChoiceId: "meaning-a",
      selectedChoiceId: body.action === "answer" ? body.choiceId : undefined,
    };
    session.canContinue = true;
  } else return send(400, { detail: "Invalid action" });
  session.revision += 1;
  session.updatedAt = new Date().toISOString();
  const result = structuredClone(session);
  state.dailyReceipts.set(key, { fingerprint, result });
  return send(200, result);
}
