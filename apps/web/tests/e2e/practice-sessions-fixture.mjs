// Synthetic transport fixtures, not production grading or durability evidence.
// Keep this small: three canonical meanings, deterministic order, opaque IDs.
import { createHash, randomUUID } from "node:crypto";

const normalize = (value) => value.trim().replace(/\s+/gu, " ").toLowerCase();

export async function handlePracticeSessions({
  req,
  res,
  url,
  cookies,
  state,
  canonicalWords,
  jsonResponse,
  readJsonBody,
  checkCsrf,
  logLine,
}) {
  if (!/^\/api\/v1\/practice-sessions(?:\/|$)/u.test(url.pathname))
    return false;
  const send = (status, body) => {
    jsonResponse(res, status, body);
    return true;
  };
  if (req.method === "GET" && url.pathname === "/api/v1/practice-sessions") {
    const delay = Number(cookies.e2e_practice_sessions_delay_ms ?? 0);
    if (Number.isInteger(delay) && delay > 0 && delay <= 5000)
      await new Promise((resolve) => setTimeout(resolve, delay));
    const status = Number(cookies.e2e_practice_sessions_status ?? 0);
    if (status === 401 || status === 503)
      return send(status, {
        detail:
          status === 401
            ? "Session expired"
            : "Practice session read unavailable",
      });
  }
  state.practiceSessions ??= new Map();
  state.practiceReceipts ??= new Map();
  state.practiceMistakes ??= new Map();
  const words = ["invite", "confirm", "reschedule"].map((slug) => {
    const word = canonicalWords[slug];
    const meaning = word.meanings[0];
    return {
      meaningId: meaning.id,
      wordText: word.text,
      wordSlug: slug,
      partOfSpeech: meaning.partOfSpeech,
      definition: meaning.shortDefinition,
    };
  });
  const projectStep = (entry) => {
    const frozen = entry.steps[entry.summary.completedSteps];
    if (!frozen) return null;
    const step = { ...frozen.public };
    delete step.meaningId;
    delete step.pronunciationMeaningId;
    delete step.pronunciationRevision;
    step.partOfSpeech = frozen.word.partOfSpeech;
    if (step.speechText) {
      step.pronunciationMeaningId = frozen.word.meaningId;
      const identity = [
        frozen.word.meaningId,
        frozen.word.wordText,
        frozen.word.partOfSpeech,
        frozen.word.definition,
      ];
      if (
        identity.every((value) => typeof value === "string" && value.length > 0)
      ) {
        // Match meaningidentity.Revision, including Go encoding/json's HTML escaping.
        const encoded = JSON.stringify(identity).replace(
          /[<>&\u2028\u2029]/gu,
          (character) =>
            `\\u${character.charCodeAt(0).toString(16).padStart(4, "0")}`,
        );
        step.pronunciationRevision = createHash("sha256")
          .update(encoded)
          .digest("hex");
      }
    }
    if (step.kind === "teach") step.meaningId = frozen.word.meaningId;
    return step;
  };
  const project = (entry) => ({
    ...entry.summary,
    currentStep: projectStep(entry),
    feedback: entry.feedback,
    canContinue:
      entry.steps[entry.summary.completedSteps]?.public.kind === "teach" ||
      (entry.feedback !== null &&
        (entry.summary.mode !== "journey" ||
          entry.feedback.correct ||
          entry.feedback.revealed)),
  });
  const availableMistakes = () =>
    [...state.practiceMistakes.values()].filter((source) => !source.resolved)
      .length;
  const parts = url.pathname
    .slice("/api/v1/practice-sessions".length)
    .split("/")
    .filter(Boolean);
  if (req.method === "GET") {
    if (parts.length === 0)
      return send(200, {
        items: [...state.practiceSessions.values()]
          .reverse()
          .map((entry) => entry.summary)
          .sort(
            (left, right) =>
              Number(right.status === "in_progress") -
                Number(left.status === "in_progress") ||
              right.updatedAt.localeCompare(left.updatedAt),
          )
          .slice(0, 20),
        availableMistakes: availableMistakes(),
        journey: {
          completedLessons: [...state.practiceSessions.values()].filter(
            (entry) =>
              entry.summary.mode === "journey" &&
              entry.summary.status === "completed",
          ).length,
          todayCompletedLessons: [...state.practiceSessions.values()].filter(
            (entry) =>
              entry.summary.mode === "journey" &&
              entry.summary.status === "completed",
          ).length,
          dailyGoal: 3,
          currentSession:
            [...state.practiceSessions.values()].find(
              (entry) =>
                entry.summary.mode === "journey" &&
                entry.summary.status === "in_progress",
            )?.summary ?? null,
          nextLesson:
            [...state.practiceSessions.values()].filter(
              (entry) =>
                entry.summary.mode === "journey" &&
                entry.summary.status === "completed",
            ).length + 1,
          supportedKnownWords: 3,
          totalKnownWords: 3,
        },
      });
    const entry = state.practiceSessions.get(parts[0]);
    return entry && parts.length === 1
      ? send(200, project(entry))
      : send(404, { detail: "Practice session not found" });
  }
  if (req.method !== "POST") return send(405, { detail: "Method not allowed" });
  if (!checkCsrf(req, cookies, res, logLine)) return true;
  const key = req.headers["idempotency-key"];
  if (!key) return send(400, { detail: "Missing idempotency key" });
  const body = await readJsonBody(req);
  const fingerprint = JSON.stringify([url.pathname, body]);
  const receipt = state.practiceReceipts.get(key);
  if (receipt) {
    if (receipt.fingerprint !== fingerprint)
      return send(409, { detail: "Changed request" });
    return send(200, project(state.practiceSessions.get(receipt.sessionId)));
  }
  if (parts.length === 0) {
    if (
      !["typed_recall", "listening_choice", "mistakes", "journey"].includes(
        body.mode,
      )
    )
      return send(400, { detail: "Invalid mode" });
    if (
      body.lessonKey &&
      !["daily-conversation", "conversation-basics"].includes(body.lessonKey)
    )
      return send(503, { detail: "Fixture lesson unavailable" });
    if (body.mode === "journey") {
      const current = [...state.practiceSessions.values()].find(
        (entry) =>
          entry.summary.mode === "journey" &&
          entry.summary.status === "in_progress",
      );
      if (current) {
        state.practiceReceipts.set(key, {
          fingerprint,
          sessionId: current.summary.id,
        });
        return send(200, project(current));
      }
    }
    const selectedList = body.listId ? state.wordLists?.get(body.listId) : null;
    if (
      (body.listId !== undefined) !== (body.listRevision !== undefined) ||
      (body.listId && (body.lessonKey || body.mode === "mistakes"))
    )
      return send(422, { detail: "Invalid list selection" });
    if (
      body.listId &&
      (!selectedList || selectedList.revision !== body.listRevision)
    )
      return send(409, { detail: "Selected list changed or was deleted" });
    const selected =
      body.mode === "mistakes"
        ? words.filter((word) => {
            const source = state.practiceMistakes.get(word.meaningId);
            return source && !source.resolved;
          })
        : selectedList
          ? words.filter((word) => selectedList.members.has(word.meaningId))
          : words;
    if (!selected.length)
      return send(409, {
        detail: selectedList
          ? "This list has no supported meanings to practise."
          : "There are no supported mistakes to practise right now.",
      });
    const now = new Date().toISOString();
    const id = randomUUID();
    const entry = {
      summary: {
        id,
        mode: body.mode,
        ...(body.lessonKey ? { lessonKey: body.lessonKey } : {}),
        ...(selectedList
          ? {
              listId: selectedList.id,
              listName: selectedList.name,
              listRevision: selectedList.revision,
            }
          : {}),
        contentVersion: "starter-21-v1",
        gradingVersion: "exact-recall-v1",
        status: "in_progress",
        revision: 0,
        completedSteps: 0,
        totalSteps: selected.length,
        firstAnswersCorrect: 0,
        questionsAnswered: 0,
        createdAt: now,
        updatedAt: now,
      },
      feedback: null,
      steps: selected.map((word) => {
        const choices =
          body.mode === "listening_choice"
            ? words.map((option) => ({
                id: randomUUID(),
                text: option.definition,
              }))
            : [];
        return {
          word,
          sourceId:
            body.mode === "mistakes"
              ? state.practiceMistakes.get(word.meaningId).id
              : null,
          correctChoice: choices[words.indexOf(word)]?.id,
          public: {
            id: randomUUID(),
            kind:
              body.mode === "listening_choice"
                ? "listening_choice"
                : "typed_recall",
            prompt:
              body.mode === "listening_choice"
                ? "Listen and choose the meaning."
                : `Type the word or phrase you learned: ${word.definition}`,
            choices,
            ...(body.mode === "listening_choice"
              ? { speechText: word.wordText, speechLanguage: "en-US" }
              : {}),
          },
        };
      }),
    };
    if (body.mode === "journey") {
      const word = words[0];
      const choices = words.map((item) => ({
        id: randomUUID(),
        text: item.definition,
      }));
      const sentence = "We invite our friends.";
      const tokens = sentence
        .split(" ")
        .map((text) => ({ id: randomUUID(), text }));
      const make = (kind, prompt, extra = {}) => ({
        word,
        sourceId: null,
        correctChoice: choices[0].id,
        public: { id: randomUUID(), kind, prompt, choices, ...extra },
      });
      entry.steps = [
        make("teach", word.definition, {
          meaningId: word.meaningId,
          wordText: word.wordText,
          context: sentence,
          speechText: word.wordText,
          choices: [],
        }),
        make("listening_choice", "Listen and choose the meaning", {
          speechText: word.wordText,
        }),
        make("gap_choice", "Complete the sentence", {
          context: "We _____ our friends.",
        }),
        {
          ...make("sentence_build", "Put the sentence together", {
            speechText: sentence,
            choices: [...tokens].reverse(),
          }),
          correctTokenIds: tokens.map((token) => token.id),
          sentence,
        },
        make("meaning_choice", "Choose the meaning", {
          wordText: word.wordText,
        }),
      ];
      entry.summary.totalSteps = entry.steps.length;
      entry.summary.journeyLesson =
        [...state.practiceSessions.values()].filter(
          (item) =>
            item.summary.mode === "journey" &&
            item.summary.status === "completed",
        ).length + 1;
      entry.summary.newWords = 1;
    }
    state.practiceSessions.set(id, entry);
    state.practiceReceipts.set(key, { fingerprint, sessionId: id });
    return send(200, project(entry));
  }
  const entry = state.practiceSessions.get(parts[0]);
  if (!entry || parts.length !== 2 || parts[1] !== "actions")
    return send(404, { detail: "Practice session not found" });
  const step = entry.steps[entry.summary.completedSteps];
  if (
    !step ||
    body.expectedRevision !== entry.summary.revision ||
    body.stepId !== step.public.id
  )
    return send(409, {
      detail: "This practice changed. Reload it to continue.",
    });
  if (!body.clientActionId)
    return send(400, { detail: "Missing action identity" });
  if (body.action === "continue") {
    if (
      (step.public.kind !== "teach" &&
        (!entry.feedback ||
          (entry.summary.mode === "journey" &&
            !entry.feedback.correct &&
            !entry.feedback.revealed))) ||
      body.typedAnswer ||
      body.choiceId ||
      body.tokenIds
    )
      return send(409, { detail: "No confirmed answer" });
    entry.summary.completedSteps += 1;
    entry.feedback = null;
    if (entry.summary.completedSteps === entry.summary.totalSteps) {
      entry.summary.status = "completed";
      entry.summary.completedAt = new Date().toISOString();
    }
  } else if (["answer", "reveal"].includes(body.action)) {
    if (
      entry.feedback &&
      (entry.feedback.correct ||
        (body.action === "reveal" &&
          (entry.summary.mode !== "journey" || entry.feedback.revealed)))
    )
      return send(409, { detail: "Answer already recorded" });
    const first = entry.feedback === null;
    const assisted = !first || body.action === "reveal";
    let correct = false;
    if (body.action === "answer") {
      if (step.public.kind === "typed_recall") {
        if (
          typeof body.typedAnswer !== "string" ||
          !normalize(body.typedAnswer) ||
          body.choiceId
        )
          return send(400, { detail: "Invalid typed answer" });
        correct = normalize(body.typedAnswer) === normalize(step.word.wordText);
      } else if (step.public.kind === "sentence_build") {
        if (
          !Array.isArray(body.tokenIds) ||
          body.tokenIds.some(
            (id) => !step.public.choices.some((token) => token.id === id),
          ) ||
          new Set(body.tokenIds).size !== body.tokenIds.length
        )
          return send(400, { detail: "Invalid token order" });
        correct =
          JSON.stringify(body.tokenIds) ===
          JSON.stringify(step.correctTokenIds);
      } else {
        if (
          !step.public.choices.some((choice) => choice.id === body.choiceId) ||
          body.typedAnswer
        )
          return send(400, { detail: "Invalid choice" });
        correct = body.choiceId === step.correctChoice;
      }
    } else if (body.typedAnswer || body.choiceId)
      return send(400, { detail: "Invalid reveal" });
    if (first) {
      entry.summary.questionsAnswered += 1;
      if (correct && !assisted) entry.summary.firstAnswersCorrect += 1;
    }
    if (body.action === "answer" && !correct) {
      state.practiceMistakes.set(step.word.meaningId, {
        id: randomUUID(),
        resolved: false,
      });
    }
    if (first && correct && step.sourceId) {
      const latest = state.practiceMistakes.get(step.word.meaningId);
      if (latest?.id === step.sourceId) latest.resolved = true;
    }
    entry.feedback = {
      stepId: step.public.id,
      correct,
      ...(step.public.kind === "sentence_build"
        ? {
            correctTokenIds: step.correctTokenIds,
            selectedTokenIds: body.tokenIds ?? [],
          }
        : {}),
      ...(correct &&
      ["listening_choice", "meaning_choice", "gap_choice"].includes(
        step.public.kind,
      )
        ? { correctChoiceId: step.correctChoice }
        : {}),
      assisted,
      ...(entry.summary.mode === "journey"
        ? { revealed: body.action === "reveal" }
        : {}),
      answer: step.sentence ?? step.word.wordText,
      explanation: `The word from this lesson is “${step.word.wordText}”. ${step.word.definition}${assisted ? " This answer was practised with help." : ""}`,
      wordText: step.word.wordText,
      wordSlug: step.word.wordSlug,
      meaningId: step.word.meaningId,
    };
  } else return send(400, { detail: "Invalid action" });
  entry.summary.revision += 1;
  entry.summary.updatedAt = new Date().toISOString();
  state.practiceReceipts.set(key, { fingerprint, sessionId: entry.summary.id });
  return send(200, project(entry));
}
