// Transport/UI fixtures only. Real grading and PostgreSQL durability are tested
// in apps/api/business/stories. Catalog is copied from the authored Go snapshot.
import { randomUUID } from "node:crypto";
import { readFileSync } from "node:fs";
const catalog = JSON.parse(
  readFileSync(
    new URL("./stories-catalog-fixture.json", import.meta.url),
    "utf8",
  ),
);

function project(st) {
  const completed = st.status === "completed";
  const index = st.completedSteps;
  const currentStep = completed ? null : st.snapshot.Steps[index].Public;
  return {
    id: st.id,
    storyKey: st.snapshot.Story.key,
    title: st.snapshot.Story.title,
    situation: st.snapshot.Story.situation,
    contentVersion: st.snapshot.ContentVersion,
    gradingVersion: st.snapshot.GradingVersion,
    status: st.status,
    revision: st.revision,
    completedSteps: index,
    totalSteps: st.snapshot.Steps.length,
    firstAnswersCorrect: st.firstAnswersCorrect,
    questionsAnswered: st.questionsAnswered,
    createdAt: st.createdAt,
    updatedAt: st.updatedAt,
    ...(completed ? { completedAt: st.completedAt } : {}),
    currentStep,
    feedback: completed ? null : st.feedback,
    canContinue:
      !completed && (currentStep.kind === "line" || !!st.feedback?.correct),
    visibleLines: st.snapshot.Steps.filter(
      (s, i) => i <= index && s.Public.line,
    ).map((s) => s.Public.line),
    vocabulary: st.snapshot.Story.vocabulary,
  };
}
function summary(st) {
  const out = project(st);
  for (const key of [
    "currentStep",
    "feedback",
    "canContinue",
    "visibleLines",
    "vocabulary",
  ])
    delete out[key];
  return out;
}
export async function handleStoryFeatures({
  req,
  res,
  url,
  cookies,
  state,
  jsonResponse,
  readJsonBody,
  checkCsrf,
  logLine,
}) {
  const path = url.pathname;
  if (
    !path.startsWith("/api/v1/stories") &&
    !path.startsWith("/api/v1/story-sessions")
  )
    return false;
  const send = (code, body) => {
    jsonResponse(res, code, body);
    return true;
  };
  if (!cookies.vocanova_session) return send(401, { detail: "Sign in" });
  if (cookies.e2e_stories === "unavailable")
    return send(503, { detail: "Stories unavailable" });
  state.storySessions ??= new Map();
  state.storyReceipts ??= new Map();
  const latest = (key) =>
    [...state.storySessions.values()]
      .filter((s) => s.snapshot.Story.key === key)
      .sort((a, b) => b.sequence - a.sequence)[0];
  if (req.method === "GET" && path === "/api/v1/stories")
    return send(200, {
      items:
        cookies.e2e_stories === "empty"
          ? []
          : catalog.map((s) => ({
              ...s.Story,
              ...(latest(s.Story.key)
                ? { latestSession: summary(latest(s.Story.key)) }
                : {}),
            })),
    });
  const reading = path.match(/^\/api\/v1\/stories\/([^/]+)$/u);
  if (req.method === "GET" && reading) {
    const snap = catalog.find(
      (s) => s.Story.key === decodeURIComponent(reading[1]),
    );
    return snap
      ? send(200, {
          ...snap.Story,
          lines: snap.Steps.filter((s) => s.Public.line).map(
            (s) => s.Public.line,
          ),
        })
      : send(404, { detail: "Story not found" });
  }
  const match = path.match(/^\/api\/v1\/story-sessions\/([^/]+)(\/actions)?$/u);
  if (req.method === "GET" && match && !match[2]) {
    const st = state.storySessions.get(match[1]);
    return st
      ? send(200, project(st))
      : send(404, { detail: "Story not found" });
  }
  if (req.method !== "POST") return send(404, { detail: "Story not found" });
  if (!checkCsrf(req, cookies, res, logLine)) return true;
  const key = req.headers["idempotency-key"];
  if (!key) return send(422, { detail: "Missing retry identity" });
  const body = await readJsonBody(req);
  if (path === "/api/v1/story-sessions") {
    const fp = JSON.stringify(body),
      receiptKey = `start:${key}`;
    const prior = state.storyReceipts.get(receiptKey);
    if (prior)
      return prior.fp === fp
        ? send(200, project(state.storySessions.get(prior.id)))
        : send(409, { detail: "Changed request" });
    const snapshot = catalog.find((s) => s.Story.key === body.storyKey);
    if (!snapshot) return send(404, { detail: "Story not found" });
    const now = new Date().toISOString();
    const st = {
      id: randomUUID(),
      snapshot: structuredClone(snapshot),
      sequence: state.storySessions.size,
      status: "in_progress",
      revision: 0,
      completedSteps: 0,
      firstAnswersCorrect: 0,
      questionsAnswered: 0,
      feedback: null,
      createdAt: now,
      updatedAt: now,
    };
    state.storySessions.set(st.id, st);
    state.storyReceipts.set(receiptKey, { id: st.id, fp });
    return send(200, project(st));
  }
  if (!match || !match[2]) return send(404, { detail: "Story not found" });
  const st = state.storySessions.get(match[1]);
  if (!st) return send(404, { detail: "Story not found" });
  const fp = JSON.stringify([st.id, body]);
  const receiptKey = `action:${key}`,
    clientKey = `client:${st.id}:${body.clientActionId}`;
  for (const k of [receiptKey, clientKey]) {
    const prior = state.storyReceipts.get(k);
    if (prior)
      return prior.fp === fp && prior.key === key
        ? send(200, project(st))
        : send(409, { detail: "Changed request" });
  }
  const step = st.snapshot.Steps[st.completedSteps];
  if (
    st.status === "completed" ||
    body.expectedRevision !== st.revision ||
    body.stepId !== step.Public.id
  )
    return send(409, { detail: "Story changed" });
  if (body.action === "continue") {
    if (body.choiceId || (step.Public.kind !== "line" && !st.feedback?.correct))
      return send(409, { detail: "Answer before continuing" });
    st.completedSteps++;
    st.feedback = null;
    if (st.completedSteps === st.snapshot.Steps.length) {
      st.status = "completed";
      st.completedAt = new Date().toISOString();
    }
  } else if (body.action === "answer") {
    if (
      step.Public.kind === "line" ||
      !step.Public.choices.some((c) => c.id === body.choiceId)
    )
      return send(400, { detail: "Invalid answer" });
    if (st.feedback?.correct)
      return send(409, { detail: "Answer already checked" });
    const correct = body.choiceId === step.CorrectChoice;
    if (!st.feedback) {
      st.questionsAnswered++;
      if (correct) st.firstAnswersCorrect++;
    }
    st.feedback = {
      stepId: step.Public.id,
      correct,
      answer: step.Public.choices.find((c) => c.id === step.CorrectChoice).text,
      explanation: step.Explanation,
    };
  } else return send(400, { detail: "Invalid action" });
  st.revision++;
  st.updatedAt = new Date().toISOString();
  const receipt = { id: st.id, fp, key };
  state.storyReceipts.set(receiptKey, receipt);
  state.storyReceipts.set(clientKey, receipt);
  return send(200, project(st));
}
