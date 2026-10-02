// Browser transport fixtures only. Real grading, ownership and persistence are
// verified by the Go/PostgreSQL tests; this module is not production logic.
import { randomUUID } from "node:crypto";
import { handlePracticeSessions } from "./practice-sessions-fixture.mjs";
import { savedVocabularyFixture } from "./saved-vocabulary-fixture.mjs";

export async function handleLearningFeatures({ req, res, url, cookies, state, canonicalWords, jsonResponse, readJsonBody, checkCsrf, logLine }) {
  if (await handlePracticeSessions({ req, res, url, cookies, state, canonicalWords, jsonResponse, readJsonBody, checkCsrf, logLine })) return true;
  const send = (code, data) => { jsonResponse(res, code, data); return true; };
  if (req.method === "GET" && url.pathname === "/api/v1/user-words") {
    const result = savedVocabularyFixture({ url, cookies, state, canonicalWords });
    return send(result.status, result.body);
  }
  if (url.pathname === "/api/v1/learning-preferences") {
    const current = () => state.learningPreferences ?? { learningGoal: cookies.e2e_learning_preferences === "unset" ? null : "general", mainUseCase: cookies.e2e_learning_preferences === "unset" ? null : cookies.e2e_onboarding_focus ?? "daily_life", revision: 0 };
    if (req.method === "GET") return send(200, current());
    if (req.method === "PATCH") {
      if (!checkCsrf(req, cookies, res, logLine)) return true;
      const body = await readJsonBody(req);
      if (current().learningGoal === body.learningGoal && current().mainUseCase === body.mainUseCase) return send(200, current());
      if (body.expectedRevision !== current().revision) return send(409, { detail: "Newer choices exist" });
      state.learningPreferences = { learningGoal: body.learningGoal, mainUseCase: body.mainUseCase, revision: current().revision + 1 };
      return send(200, current());
    }
  }
  if (url.pathname.startsWith("/api/v1/meaning-knowledge/")) {
    const meaningId = url.pathname.slice("/api/v1/meaning-knowledge/".length);
    if (!Object.values(canonicalWords).some(word => word.meanings.some(meaning => meaning.id === meaningId))) return send(404, { detail: "not found" });
    state.wordKnowledge ??= new Map();
    state.wordKnowledgeReceipts ??= new Map();
    const current = () => state.wordKnowledge.get(meaningId) ?? { meaningId, selfReportedKnown: false, note: "" };
    if (req.method === "GET") return send(200, current());
    if (req.method === "PUT" || req.method === "PATCH" || req.method === "DELETE") {
      if (!checkCsrf(req, cookies, res, logLine)) return true;
      const key = req.headers["idempotency-key"];
      const body = req.method !== "DELETE" ? await readJsonBody(req) : null;
      const fingerprint = JSON.stringify([req.method, meaningId, body]);
      if (!key) return send(422, { detail: "missing key" });
      if (state.wordKnowledgeReceipts.has(key) && state.wordKnowledgeReceipts.get(key) !== fingerprint) return send(409, { detail: "changed request" });
      if (!state.wordKnowledgeReceipts.has(key)) {
        if (req.method === "DELETE") state.wordKnowledge.delete(meaningId);
        else state.wordKnowledge.set(meaningId, { meaningId, selfReportedKnown: body.selfReportedKnown, note: req.method === "PATCH" ? current().note : body.note.trim(), updatedAt: new Date().toISOString() });
        state.wordKnowledgeReceipts.set(key, fingerprint);
      }
      if (req.method === "DELETE") { res.writeHead(204); res.end(); return true; }
      return send(200, current());
    }
  }
  if (req.method === "GET" && url.pathname === "/api/v1/achievements") {
    if (cookies.e2e_achievements === "unavailable") return send(503, { detail: "Fixture unavailable" });
    const earned = cookies.e2e_achievements === "earned";
    // Fixed UI examples only; production metric qualification is tested in PostgreSQL.
    const definitions = [
      ["guided-first", "First lesson", "Complete one guided lesson.", "lessons", "participation", 1],
      ["guided-three", "Three lessons", "Complete three different guided lessons.", "lessons", "participation", 3],
      ["guided-seven", "Seven lessons", "Complete seven different guided lessons.", "lessons", "participation", 7],
      ["recall-independent-first", "Recall on your own", "Complete typed recall without showing answers.", "practice", "unaided_recall", 1],
      ["practice-five", "Keep practising", "Complete five practice sessions. Practice with help counts.", "practice", "participation", 5],
      ["reviews-ten", "Ten reviews", "Answer ten scheduled review prompts. Skipped prompts do not count.", "reviews", "participation", 10],
      ["reviews-fifty", "Fifty reviews", "Answer fifty scheduled review prompts.", "reviews", "participation", 50],
      ["writing-first", "First sentence checked", "Write a sentence and receive saved feedback.", "writing", "participation", 1],
    ];
    return send(200, { catalogVersion: "1", items: definitions.map(([id, label, description, category, criterion, target]) => {
      const unlocked = earned && (id === "guided-first" || id === "recall-independent-first");
      return { id, label, description, category, criterion, target, current: unlocked ? target : earned && category === "lessons" ? 1 : 0, earned: unlocked, ...(unlocked ? { earnedAt: "2026-10-02T10:00:00Z" } : {}) };
    }) });
  }
  const rawWords = ["invite", "confirm", "reschedule"].map((slug) => canonicalWords[slug]);
  const words = rawWords.map((word) => ({ meaningId: word.meanings[0].id, wordText: word.text, wordSlug: word.slug, partOfSpeech: word.meanings[0].partOfSpeech, definition: word.meanings[0].shortDefinition, example: word.meanings[0].examples[0].exampleText, usageNote: word.meanings[0].usageNotes[0]?.noteText ?? "" }));
  const summary = { key: "conversation-basics", version: "1", title: "Make a plan with a friend", situationSlug: "daily-conversation", situationTitle: "Daily Conversation", description: "Invite a friend, confirm a time, and change a plan.", wordCount: 3, stepCount: 9 };
  const steps = ["teach", "recall", "context"].flatMap((kind) => words.map((word, index) => ({ id: `${kind}-${index + 1}`, kind, word, prompt: kind === "teach" ? "Meet a useful word" : kind === "recall" ? `Which meaning matches “${word.wordText}”?` : `Choose a word for this situation.`, ...(kind === "context" ? { context: ["Ask your friend to come to dinner.", "Say that the meeting time is definite.", "Move your meeting to another day."][index] } : {}), choices: kind === "teach" ? [] : words.map((choice) => ({ id: choice.meaningId, text: kind === "recall" ? choice.definition : choice.wordText })) })));
  const unavailable = cookies.e2e_lessons === "unavailable";
  if (req.method === "GET" && url.pathname === "/api/v1/lesson-recommendation") {
    if (cookies.e2e_recommendation_auth === "expired") return send(401, { detail: "Fixture expired session" });
    if (unavailable) return send(503, { detail: "Fixture recommendation unavailable" });
    if (cookies.e2e_lessons === "empty") return send(200, { status: "content_unavailable", recommendation: null });
    const usefulTargetCount = words.filter(word => !state.wordKnowledge?.get(word.meaningId)?.selfReportedKnown).length;
    const lesson = { ...summary, status: state.lesson?.status ?? "not_started", completedSteps: state.lesson?.completedSteps ?? 0, ...(state.lesson ? { sessionId: state.lesson.id } : {}) };
    const resume = lesson.status === "in_progress";
    if (lesson.status === "completed") return send(200, { status: "no_unfinished_lessons", recommendation: null });
    if (!resume && usefulTargetCount === 0) return send(200, { status: "no_useful_targets", recommendation: null });
    const focus = state.learningPreferences?.mainUseCase ?? (cookies.e2e_learning_preferences === "unset" ? null : cookies.e2e_onboarding_focus ?? "daily_life");
    const matchesFocus = focus === "social";
    return send(200, { status: "recommended", recommendation: { lesson, reason: resume ? "resume" : matchesFocus ? "focus_and_useful_words" : "useful_words", usefulTargetCount, totalTargetCount: 3, matchesFocus } });
  }
  if (req.method === "GET" && url.pathname === "/api/v1/lessons") {
    if (unavailable) return send(503, { detail: "fixture unavailable" });
    if (cookies.e2e_lessons === "empty") return send(200, { items: [] });
    if (cookies.e2e_lessons === "expanded") return send(200, { items: Array.from({ length: 30 }, (_, index) => ({ ...summary, key: index === 0 ? summary.key : `fixture-lesson-${index}`, title: index === 0 ? summary.title : `Fixture lesson ${index + 1}`, situationSlug: index < 6 ? "daily-conversation" : "travel-basics", situationTitle: index < 6 ? "Daily Conversation" : "Travel basics", status: "not_started", completedSteps: 0 })) });
    return send(200, { items: [{ ...summary, status: state.lesson?.status ?? "not_started", completedSteps: state.lesson?.completedSteps ?? 0, ...(state.lesson ? { sessionId: state.lesson.id } : {}) }] });
  }
  if (req.method === "POST" && url.pathname === `/api/v1/lessons/${summary.key}/sessions`) {
    if (!checkCsrf(req, cookies, res, logLine)) return true;
    state.lesson ??= { id: randomUUID(), lessonKey: summary.key, lessonVersion: "1", title: summary.title, situationSlug: summary.situationSlug, status: "in_progress", revision: 0, completedSteps: 0, totalSteps: 9, words, currentStep: steps[0], feedback: null, canContinue: true, firstAnswersCorrect: 0, questionsAnswered: 0 };
    state.lessonReceipts ??= new Map();
    state.lessonAnswered ??= new Set();
    return send(200, state.lesson);
  }
  if (url.pathname.startsWith("/api/v1/lesson-sessions/")) {
    const [id, action] = url.pathname.slice("/api/v1/lesson-sessions/".length).split("/");
    if (!state.lesson || state.lesson.id !== id) return send(404, { detail: "not found" });
    if (req.method === "GET" && !action) return send(200, state.lesson);
    if (req.method === "POST" && action === "actions") {
      if (!checkCsrf(req, cookies, res, logLine)) return true;
      const body = await readJsonBody(req);
      const key = req.headers["idempotency-key"];
      const fingerprint = JSON.stringify(body);
      if (state.lessonReceipts.has(key)) return state.lessonReceipts.get(key) === fingerprint ? send(200, state.lesson) : send(409, { detail: "changed action" });
      const session = state.lesson;
      const step = session.currentStep;
      if (!step || body.expectedRevision !== session.revision || body.stepId !== step.id) return send(409, { detail: "stale revision" });
      if (body.action === "continue" && session.canContinue) {
        session.completedSteps++;
        session.currentStep = steps[session.completedSteps] ?? null;
        session.feedback = null;
        session.canContinue = session.currentStep?.kind === "teach";
        if (!session.currentStep) { session.status = "completed"; session.completedAt = new Date().toISOString(); }
      } else if (body.action === "answer" && step.kind !== "teach" && !session.canContinue && step.choices.some((item) => item.id === body.choiceId)) {
        const correct = body.choiceId === step.word.meaningId;
        if (!state.lessonAnswered.has(step.id)) { session.questionsAnswered++; if (correct) session.firstAnswersCorrect++; state.lessonAnswered.add(step.id); }
        session.feedback = { stepId: step.id, correct, explanation: `${step.word.wordText}: ${step.word.definition}`, correctChoiceId: step.word.meaningId };
        session.canContinue = correct;
      } else return send(400, { detail: "invalid action" });
      session.revision++;
      state.lessonReceipts.set(key, fingerprint);
      return send(200, session);
    }
  }
  if (req.method === "GET" && url.pathname === "/api/v1/knowledge-summary") {
    const saved = state.savedMeaningIds.size;
    const learning = [...state.savedMeaningIds].filter((id) => state.reviewedMeaningIds.has(id)).length;
    return send(200, { selfReportedKnown: [...(state.wordKnowledge?.values() ?? [])].filter(item => item.selfReportedKnown).length, saved, new: saved - learning, learning, reviewing: 0, mastered: 0, ignored: 0, archived: 0, due: saved - learning });
  }
  if (req.method === "GET" && url.pathname === "/api/v1/canonical-words") {
    if (cookies.e2e_vocabulary === "unavailable") return send(503, { detail: "Fixture vocabulary unavailable" });
    const q = (url.searchParams.get("q") ?? "").toLowerCase();
    const category = url.searchParams.get("category");
    const level = url.searchParams.get("level");
    const knowledge = url.searchParams.get("knowledge");
    const all = Object.values(canonicalWords).flatMap((word) => word.meanings.map((meaning) => ({ meaningId: meaning.id, wordId: word.id, wordSlug: word.slug, wordText: word.text, partOfSpeech: meaning.partOfSpeech, shortDefinition: meaning.shortDefinition, difficultyLevel: word.difficultyLevel?.toLowerCase() ?? "unknown", saved: state.savedMeaningIds.has(meaning.id), due: state.savedMeaningIds.has(meaning.id) && !state.reviewedMeaningIds.has(meaning.id), ...(state.savedMeaningIds.has(meaning.id) ? { userWordId: `uw-${meaning.id}`, reviewState: state.reviewedMeaningIds.has(meaning.id) ? "learning" : "new" } : {}) })));
    for (const word of all) word.selfReportedKnown = state.wordKnowledge?.get(word.meaningId)?.selfReportedKnown ?? false;
    const filtered = all.filter((word) => `${word.wordText} ${word.shortDefinition}`.toLowerCase().includes(q) && (!level || word.difficultyLevel === level) && (!category || category === (word.wordSlug === "pour" ? "daily_life" : "social")) && (!knowledge || (knowledge === "known" && word.selfReportedKnown) || (knowledge === "saved" && word.saved) || (knowledge === "unexplored" && !word.saved && !word.selfReportedKnown)));
    const offset = Number(url.searchParams.get("after") ?? 0);
    const limit = Number(url.searchParams.get("limit") ?? 20);
    const items = filtered.slice(offset, offset + limit);
    const hasMore = offset + limit < filtered.length;
    return send(200, { items, totalCount: filtered.length, hasMore, ...(hasMore ? { nextCursor: String(offset + limit) } : {}) });
  }
  return false;
}
