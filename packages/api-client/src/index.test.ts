import assert from "node:assert/strict";
import { describe, it } from "node:test";

import {
  ApiResponseError,
  VocanovaClient,
  type KnowledgeSummary,
  type LessonAction,
  type LessonSession,
  type LessonSummary,
  type LessonRecommendationResponse,
  type ListSavedWordsResponse,
  type MeaningKnowledge,
  type VocabularySearchResponse,
} from "./index.js";

describe("VocanovaClient", () => {
  it("preserves literal saved-word filters and exposes full-filter counts and server review state", async () => {
    const expected: ListSavedWordsResponse = {
      items: [
        {
          userWordId: "saved-1",
          meaningId: "meaning-1",
          wordId: "word-1",
          wordText: "refund",
          wordSlug: "refund",
          partOfSpeech: "noun",
          shortDefinition: "Money returned after payment.",
          status: "learning",
          source: "manual",
          saved: true,
          addedAt: "2026-10-02T12:00:00Z",
          reviewState: "reviewing",
          due: true,
        },
      ],
      totalCount: 24,
      nextCursor: "next/+cursor=",
      hasMore: true,
    };
    const seen: URL[] = [];
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: async (input, init) => {
        const url = new URL(String(input));
        seen.push(url);
        assert.equal(init?.method, "GET");
        assert.equal(init?.body, undefined);
        assert.equal(url.pathname, "/api/v1/user-words");
        return new Response(JSON.stringify(expected), { status: 200 });
      },
    });
    assert.deepEqual(
      (
        await client.listSavedWords({
          q: "refund %_ &",
          stage: "reviewing",
          due: true,
          after: "opaque/+cursor=",
          limit: 1,
        })
      ).data,
      expected,
    );
    assert.deepEqual(Object.fromEntries(seen[0]!.searchParams), {
      q: "refund %_ &",
      stage: "reviewing",
      due: "true",
      after: "opaque/+cursor=",
      limit: "1",
    });
    await client.listSavedWords({ due: false });
    assert.equal(seen[1]!.searchParams.has("due"), false);
  });

  it("retains recommendation availability and nullable results without manufacturing a lesson", async () => {
    const recommended: LessonRecommendationResponse = {
      status: "recommended",
      recommendation: {
        lesson: {
          key: "conversation-start",
          version: "1",
          title: "Start a conversation",
          situationSlug: "daily-conversation",
          situationTitle: "Daily Conversation",
          description: "Useful words for meeting someone.",
          wordCount: 3,
          stepCount: 9,
          status: "not_started",
          completedSteps: 0,
        },
        reason: "focus_and_useful_words",
        usefulTargetCount: 1,
        totalTargetCount: 3,
        matchesFocus: true,
      },
    };
    const responses: LessonRecommendationResponse[] = [
      recommended,
      { status: "content_unavailable", recommendation: null },
      { status: "no_useful_targets", recommendation: null },
      { status: "no_unfinished_lessons", recommendation: null },
    ];
    let calls = 0;
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: async (input, init) => {
        assert.equal(
          String(input),
          "https://api.example.com/api/v1/lesson-recommendation",
        );
        assert.equal(init?.method, "GET");
        assert.equal(init?.body, undefined);
        return new Response(JSON.stringify(responses[calls++]), {
          status: 200,
        });
      },
    });
    for (const expected of responses)
      assert.deepEqual((await client.getLessonRecommendation()).data, expected);
    assert.equal(calls, responses.length);
  });

  it("reads nullable learning direction and sends the exact revisioned intent on explicit retry", async () => {
    const calls: { method: string | undefined; body: unknown }[] = [];
    const headers = new Headers({ "X-CSRF-Token": "synthetic-csrf" });
    const intent = {
      learningGoal: "conversation" as const,
      mainUseCase: "social" as const,
      expectedRevision: 0,
    };
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: async (url, init) => {
        assert.equal(
          String(url),
          "https://api.example.com/api/v1/learning-preferences",
        );
        if (init?.method === "PATCH")
          assert.equal(
            new Headers(init.headers).get("X-CSRF-Token"),
            "synthetic-csrf",
          );
        calls.push({
          method: init?.method,
          body: init?.body ? JSON.parse(String(init.body)) : null,
        });
        return new Response(
          JSON.stringify(
            init?.method === "PATCH"
              ? {
                  learningGoal: "conversation",
                  mainUseCase: "social",
                  revision: 1,
                }
              : { learningGoal: null, mainUseCase: null, revision: 0 },
          ),
          { status: 200 },
        );
      },
    });
    assert.deepEqual((await client.getLearningPreferences()).data, {
      learningGoal: null,
      mainUseCase: null,
      revision: 0,
    });
    for (let attempt = 0; attempt < 2; attempt++)
      assert.deepEqual(
        (await client.updateLearningPreferences(intent, { headers })).data,
        { learningGoal: "conversation", mainUseCase: "social", revision: 1 },
      );
    assert.deepEqual(
      calls.map((call) => call.body),
      [null, intent, intent],
    );
    assert.equal(calls[1]?.method, "PATCH");
    assert.deepEqual(
      [...headers.entries()],
      [["x-csrf-token", "synthetic-csrf"]],
    );
  });

  it("surfaces a learning-direction conflict without retrying or changing the requested revision", async () => {
    let calls = 0;
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: async (_url, init) => {
        calls++;
        assert.deepEqual(JSON.parse(String(init?.body)), {
          learningGoal: "work",
          mainUseCase: "work",
          expectedRevision: 3,
        });
        return new Response(
          JSON.stringify({ detail: "learning preferences changed" }),
          { status: 409 },
        );
      },
    });
    await assert.rejects(
      client.updateLearningPreferences({
        learningGoal: "work",
        mainUseCase: "work",
        expectedRevision: 3,
      }),
      (error: unknown) =>
        error instanceof ApiResponseError && error.status === 409,
    );
    assert.equal(calls, 1);
  });

  it("partially sets known status without sending a note and preserves caller headers on exact retry", async () => {
    const observed: {
      url: string;
      method: string | undefined;
      body: unknown;
      key: string | null;
    }[] = [];
    const headers = new Headers({
      "X-CSRF-Token": "synthetic-csrf",
      "Idempotency-Key": "caller-value",
    });
    const expected: MeaningKnowledge = {
      meaningId: "meaning/one",
      selfReportedKnown: true,
      note: "The current private note",
    };
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: async (input, init) => {
        const sentHeaders = new Headers(init?.headers);
        assert.equal(sentHeaders.get("X-CSRF-Token"), "synthetic-csrf");
        observed.push({
          url: String(input),
          method: init?.method,
          body: JSON.parse(String(init?.body)),
          key: sentHeaders.get("Idempotency-Key"),
        });
        return new Response(JSON.stringify(expected), { status: 200 });
      },
    });
    for (let attempt = 0; attempt < 2; attempt++) {
      assert.deepEqual(
        (
          await client.setMeaningKnown("meaning/one", true, "assessment-key", {
            headers,
          })
        ).data,
        expected,
      );
    }
    assert.deepEqual(
      observed,
      Array(2).fill({
        url: "https://api.example.com/api/v1/meaning-knowledge/meaning%2Fone",
        method: "PATCH",
        body: { selfReportedKnown: true },
        key: "assessment-key",
      }),
    );
    assert.equal(headers.get("Idempotency-Key"), "caller-value");
  });

  it("sends an explicit false assessment and exposes conflict without automatically retrying", async () => {
    let calls = 0;
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: async (_input, init) => {
        calls++;
        assert.equal(init?.method, "PATCH");
        assert.deepEqual(JSON.parse(String(init?.body)), {
          selfReportedKnown: false,
        });
        return new Response(
          JSON.stringify({ detail: "Idempotency key conflict" }),
          {
            status: 409,
            headers: { "Content-Type": "application/problem+json" },
          },
        );
      },
    });
    await assert.rejects(
      client.setMeaningKnown("meaning", false, "key"),
      (error: unknown) =>
        error instanceof ApiResponseError && error.status === 409,
    );
    assert.equal(calls, 1);
  });

  describe("vocabulary and guided learning contracts", () => {
    const word = {
      meaningId: "00000000-0000-0000-0000-000000000002",
      wordText: "invite",
      wordSlug: "invite",
      partOfSpeech: "verb",
      definition: "to ask someone to come to an event",
      example: "I will invite her to dinner.",
      usageNote: "Invite someone to an event.",
    };
    const session: LessonSession = {
      id: "00000000-0000-0000-0000-000000000010",
      lessonKey: "conversation-basics",
      lessonVersion: "1",
      title: "Make a plan with a friend",
      situationSlug: "daily-conversation",
      status: "in_progress",
      revision: 0,
      completedSteps: 0,
      totalSteps: 9,
      words: [word],
      currentStep: {
        id: "teach-1",
        kind: "teach",
        word,
        prompt: "Meet a useful word",
        choices: [],
      },
      feedback: null,
      canContinue: true,
      firstAnswersCorrect: 0,
      questionsAnswered: 0,
    };

    function jsonResponse(value: unknown): Response {
      return new Response(JSON.stringify(value), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    }

    it("encodes vocabulary filters and cursor without losing result identities or knowledge flags", async () => {
      const expected: VocabularySearchResponse = {
        items: [
          {
            meaningId: word.meaningId,
            wordId: "00000000-0000-0000-0000-000000000003",
            wordSlug: word.wordSlug,
            wordText: word.wordText,
            partOfSpeech: word.partOfSpeech,
            shortDefinition: word.definition,
            difficultyLevel: "a2",
            saved: true,
            selfReportedKnown: false,
            userWordId: "00000000-0000-0000-0000-000000000004",
            reviewState: "learning",
            due: true,
          },
        ],
        totalCount: 23,
        hasMore: true,
        nextCursor: "next+page/2=",
      };
      const client = new VocanovaClient({
        baseURL: "https://api.example.com",
        credentials: "include",
        fetch: async (input, init) => {
          const url = new URL(String(input));
          assert.equal(url.pathname, "/api/v1/canonical-words");
          assert.deepEqual(Object.fromEntries(url.searchParams), {
            q: "café & tea?",
            category: "social",
            level: "a2",
            knowledge: "saved",
            after: "page+1/==",
            limit: "7",
          });
          assert.equal(init?.method, "GET");
          assert.equal(init?.body, undefined);
          assert.equal(init?.credentials, "include");
          assert.equal(
            new Headers(init?.headers).get("Accept"),
            "application/json",
          );
          return jsonResponse(expected);
        },
      });
      const { data, response } = await client.searchVocabulary({
        q: "café & tea?",
        category: "social",
        level: "a2",
        knowledge: "saved",
        after: "page+1/==",
        limit: 7,
      });
      assert.deepEqual(data, expected);
      assert.equal(response.status, 200);
    });

    it("omits empty vocabulary filters and supports the default catalog request", async () => {
      let calls = 0;
      const empty: VocabularySearchResponse = {
        items: [],
        totalCount: 0,
        hasMore: false,
      };
      const client = new VocanovaClient({
        baseURL: "https://api.example.com",
        fetch: async (input, init) => {
          calls++;
          assert.equal(
            String(input),
            "https://api.example.com/api/v1/canonical-words",
          );
          assert.equal(init?.method, "GET");
          return jsonResponse(empty);
        },
      });
      assert.deepEqual((await client.searchVocabulary()).data, empty);
      assert.deepEqual(
        (
          await client.searchVocabulary({
            q: "",
            category: "",
            level: undefined,
            knowledge: "",
            after: "",
          })
        ).data,
        empty,
      );
      assert.equal(calls, 2);
    });

    it("reads knowledge counts without merging self-reported knowledge into reviewed mastery", async () => {
      const expected: KnowledgeSummary = {
        selfReportedKnown: 11,
        saved: 15,
        new: 3,
        learning: 4,
        reviewing: 5,
        mastered: 1,
        ignored: 2,
        archived: 0,
        due: 6,
      };
      const signal = new AbortController().signal;
      const client = new VocanovaClient({
        baseURL: "https://api.example.com",
        fetch: async (input, init) => {
          assert.equal(
            String(input),
            "https://api.example.com/api/v1/knowledge-summary",
          );
          assert.equal(init?.method, "GET");
          assert.equal(init?.signal, signal);
          assert.equal(
            new Headers(init?.headers).get("X-Test-Request"),
            "knowledge",
          );
          return jsonResponse(expected);
        },
      });
      const { data } = await client.getKnowledgeSummary({
        signal,
        headers: { "X-Test-Request": "knowledge" },
      });
      assert.deepEqual(data, expected);
    });

    it("reads private meaning knowledge using an encoded meaning identity", async () => {
      const expected: MeaningKnowledge = {
        meaningId: word.meaningId,
        selfReportedKnown: true,
        note: "Use this when planning dinner with friends.",
        updatedAt: "2026-10-02T12:00:00Z",
      };
      const client = new VocanovaClient({
        baseURL: "https://api.example.com",
        credentials: "include",
        fetch: async (input, init) => {
          assert.equal(
            String(input),
            "https://api.example.com/api/v1/meaning-knowledge/meaning%2Fone%3Fsource%3Dsearch",
          );
          assert.equal(init?.method, "GET");
          assert.equal(init?.body, undefined);
          assert.equal(init?.credentials, "include");
          assert.equal(
            new Headers(init?.headers).get("X-Test-Request"),
            "meaning-knowledge",
          );
          return jsonResponse(expected);
        },
      });
      const { data } = await client.getMeaningKnowledge(
        "meaning/one?source=search",
        { headers: { "X-Test-Request": "meaning-knowledge" } },
      );
      assert.deepEqual(data, expected);
    });

    it("updates known status and a private note with PUT, CSRF and the supplied retry key", async () => {
      const body = {
        selfReportedKnown: false,
        note: "Café invitation — practise again.",
      };
      const expected: MeaningKnowledge = {
        meaningId: word.meaningId,
        ...body,
        updatedAt: "2026-10-02T12:01:00Z",
      };
      const headers = new Headers({
        "X-CSRF-Token": "csrf-fixture",
        "Idempotency-Key": "previous-key",
      });
      const signal = new AbortController().signal;
      const client = new VocanovaClient({
        baseURL: "https://api.example.com",
        fetch: async (input, init) => {
          assert.equal(
            String(input),
            "https://api.example.com/api/v1/meaning-knowledge/meaning%2Fone",
          );
          assert.equal(init?.method, "PUT");
          assert.equal(init?.signal, signal);
          assert.equal(
            new Headers(init?.headers).get("Content-Type"),
            "application/json",
          );
          assert.equal(
            new Headers(init?.headers).get("X-CSRF-Token"),
            "csrf-fixture",
          );
          assert.equal(
            new Headers(init?.headers).get("Idempotency-Key"),
            "knowledge-retry-key",
          );
          assert.deepEqual(JSON.parse(String(init?.body)), body);
          return jsonResponse(expected);
        },
      });
      assert.deepEqual(
        (
          await client.updateMeaningKnowledge(
            "meaning/one",
            body,
            "knowledge-retry-key",
            { headers, signal },
          )
        ).data,
        expected,
      );
      assert.equal(headers.get("Idempotency-Key"), "previous-key");
    });

    it("clears meaning knowledge with an authenticated idempotent DELETE and accepts an empty 204", async () => {
      const headers = new Headers({
        "X-CSRF-Token": "csrf-fixture",
        "Idempotency-Key": "previous-key",
      });
      const noContent = new Response(null, { status: 204 });
      const client = new VocanovaClient({
        baseURL: "https://api.example.com",
        credentials: "include",
        fetch: async (input, init) => {
          assert.equal(
            String(input),
            "https://api.example.com/api/v1/meaning-knowledge/meaning%2Fone",
          );
          assert.equal(init?.method, "DELETE");
          assert.equal(init?.body, undefined);
          assert.equal(init?.credentials, "include");
          assert.equal(
            new Headers(init?.headers).get("X-CSRF-Token"),
            "csrf-fixture",
          );
          assert.equal(
            new Headers(init?.headers).get("Idempotency-Key"),
            "clear-retry-key",
          );
          return noContent;
        },
      });
      const result = await client.clearMeaningKnowledge(
        "meaning/one",
        "clear-retry-key",
        { headers },
      );
      assert.deepEqual(result, { response: noContent });
      assert.equal(result.response.status, 204);
      assert.equal(headers.get("Idempotency-Key"), "previous-key");
    });

    it("lists lesson availability with saved session identity and completed step counts", async () => {
      const summary: LessonSummary = {
        key: "conversation-basics",
        version: "1",
        title: "Make a plan with a friend",
        situationSlug: "daily-conversation",
        situationTitle: "Daily Conversation",
        description: "Invite a friend and agree on a plan.",
        wordCount: 3,
        stepCount: 9,
        status: "in_progress",
        sessionId: session.id,
        completedSteps: 4,
      };
      const expected = {
        items: [
          summary,
          {
            ...summary,
            key: "airport",
            status: "not_started" as const,
            sessionId: undefined,
            completedSteps: 0,
          },
        ],
      };
      const client = new VocanovaClient({
        baseURL: "https://api.example.com",
        fetch: async (input, init) => {
          assert.equal(String(input), "https://api.example.com/api/v1/lessons");
          assert.equal(init?.method, "GET");
          assert.equal(
            new Headers(init?.headers).get("X-Test-Request"),
            "lessons",
          );
          return jsonResponse(expected);
        },
      });
      const { data } = await client.listLessons({
        headers: { "X-Test-Request": "lessons" },
      });
      assert.deepEqual(data.items[0], summary);
      assert.equal(data.items[1]?.status, "not_started");
      assert.equal(data.items[1]?.sessionId, undefined);
      assert.equal(data.items[1]?.completedSteps, 0);
    });

    it("starts a lesson with its explicit retry key while retaining CSRF and caller options", async () => {
      const headers = new Headers({
        "X-CSRF-Token": "csrf-fixture",
        "Idempotency-Key": "old-key",
      });
      const signal = new AbortController().signal;
      const client = new VocanovaClient({
        baseURL: "https://api.example.com",
        credentials: "include",
        fetch: async (input, init) => {
          assert.equal(
            String(input),
            "https://api.example.com/api/v1/lessons/daily%20plans%2F1/sessions",
          );
          assert.equal(init?.method, "POST");
          assert.equal(init?.body, undefined);
          assert.equal(init?.signal, signal);
          assert.equal(init?.credentials, "include");
          assert.equal(
            new Headers(init?.headers).get("X-CSRF-Token"),
            "csrf-fixture",
          );
          assert.equal(
            new Headers(init?.headers).get("Idempotency-Key"),
            "start-retry-key",
          );
          return jsonResponse(session);
        },
      });
      const { data } = await client.startLesson(
        "daily plans/1",
        "start-retry-key",
        { headers, signal },
      );
      assert.deepEqual(data, session);
      assert.equal(
        headers.get("Idempotency-Key"),
        "old-key",
        "the caller's Headers object remains reusable",
      );
    });

    it("loads a completed lesson without inventing a current step or dropping completion evidence", async () => {
      const completed: LessonSession = {
        ...session,
        status: "completed",
        revision: 17,
        completedSteps: 9,
        currentStep: null,
        feedback: null,
        canContinue: false,
        firstAnswersCorrect: 5,
        questionsAnswered: 6,
        completedAt: "2026-10-02T12:00:00Z",
      };
      const client = new VocanovaClient({
        baseURL: "https://api.example.com",
        fetch: async (input, init) => {
          assert.equal(
            String(input),
            "https://api.example.com/api/v1/lesson-sessions/owned%2Fsession%3F1",
          );
          assert.equal(init?.method, "GET");
          assert.equal(init?.body, undefined);
          return jsonResponse(completed);
        },
      });
      const { data } = await client.getLessonSession("owned/session?1");
      assert.deepEqual(data, completed);
    });

    it("sends the same lesson answer identity, revision and choice on an explicit retry", async () => {
      const answer: LessonAction = {
        stepId: "context-1",
        expectedRevision: 12,
        clientActionId: "answer-action-1",
        action: "answer",
        choiceId: word.meaningId,
      };
      const expected: LessonSession = {
        ...session,
        revision: 13,
        completedSteps: 6,
        currentStep: {
          id: "context-1",
          kind: "context",
          word,
          prompt: "Choose a word",
          context: "Ask a friend to come to dinner.",
          choices: [{ id: word.meaningId, text: word.wordText }],
        },
        feedback: {
          stepId: "context-1",
          correct: true,
          explanation: "Invite asks someone to join you.",
          correctChoiceId: word.meaningId,
        },
        canContinue: true,
        firstAnswersCorrect: 4,
        questionsAnswered: 4,
      };
      const original = { ...answer };
      let calls = 0;
      const client = new VocanovaClient({
        baseURL: "https://api.example.com",
        fetch: async (input, init) => {
          calls++;
          assert.equal(
            String(input),
            `https://api.example.com/api/v1/lesson-sessions/${session.id}/actions`,
          );
          assert.equal(init?.method, "POST");
          assert.equal(
            new Headers(init?.headers).get("Content-Type"),
            "application/json",
          );
          assert.equal(
            new Headers(init?.headers).get("X-CSRF-Token"),
            "csrf-fixture",
          );
          assert.equal(
            new Headers(init?.headers).get("Idempotency-Key"),
            "answer-retry-key",
          );
          assert.deepEqual(JSON.parse(String(init?.body)), original);
          return jsonResponse(expected);
        },
      });
      for (let attempt = 0; attempt < 2; attempt++) {
        const { data } = await client.submitLessonAction(
          session.id,
          answer,
          "answer-retry-key",
          { headers: { "X-CSRF-Token": "csrf-fixture" } },
        );
        assert.deepEqual(data, expected);
      }
      assert.equal(calls, 2);
      assert.deepEqual(answer, original);
    });

    it("sends a lesson continue action without inventing an answer choice", async () => {
      const action: LessonAction = {
        stepId: "teach-1",
        expectedRevision: 0,
        clientActionId: "continue-action-1",
        action: "continue",
      };
      const expected = {
        ...session,
        revision: 1,
        completedSteps: 1,
        currentStep: { ...session.currentStep!, id: "teach-2" },
      };
      const client = new VocanovaClient({
        baseURL: "https://api.example.com",
        fetch: async (input, init) => {
          assert.equal(
            String(input),
            "https://api.example.com/api/v1/lesson-sessions/owned%2Fsession%3F1/actions",
          );
          assert.equal(init?.method, "POST");
          assert.equal(
            new Headers(init?.headers).get("Idempotency-Key"),
            action.clientActionId,
          );
          assert.deepEqual(JSON.parse(String(init?.body)), action);
          return jsonResponse(expected);
        },
      });
      assert.deepEqual(
        (
          await client.submitLessonAction(
            "owned/session?1",
            action,
            action.clientActionId,
          )
        ).data,
        expected,
      );
    });

    it("surfaces a stale lesson revision as a conflict without retrying or fabricating progress", async () => {
      let calls = 0;
      const client = new VocanovaClient({
        baseURL: "https://api.example.com",
        fetch: async () => {
          calls++;
          return new Response(
            JSON.stringify({
              detail: "This lesson changed. Reload it to continue.",
            }),
            {
              status: 409,
              headers: { "Content-Type": "application/problem+json" },
            },
          );
        },
      });
      await assert.rejects(
        client.submitLessonAction(
          session.id,
          {
            stepId: "teach-1",
            expectedRevision: 0,
            clientActionId: "stale-action",
            action: "continue",
          },
          "stale-action",
        ),
        (error: unknown) => {
          assert.ok(error instanceof ApiResponseError);
          assert.equal(error.status, 409);
          assert.equal(
            error.message,
            "This lesson changed. Reload it to continue.",
          );
          return true;
        },
      );
      assert.equal(calls, 1);
    });
  });

  it("sends GET /api/v1/me with Accept header", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/me");
      assert.equal(init.method, "GET");
      assert.equal(new Headers(init.headers).get("Accept"), "application/json");
      return Promise.resolve(
        new Response(JSON.stringify({ email: "user@example.com" }), {
          headers: { "Content-Type": "application/json" },
          status: 200,
        }),
      );
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.getCurrentUser();
    assert.equal(data.email, "user@example.com");
  });

  it("sends POST /api/v1/auth/magic-links with JSON body", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/auth/magic-links");
      assert.equal(init.method, "POST");
      assert.equal(
        new Headers(init.headers).get("Content-Type"),
        "application/json",
      );
      assert.equal(new Headers(init.headers).get("Accept"), "application/json");
      assert.equal(
        init.body,
        JSON.stringify({ email: "user@example.com", returnTo: "/reviews" }),
      );
      return Promise.resolve(new Response(null, { status: 204 }));
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { response } = await client.requestMagicLink({
      email: "user@example.com",
      returnTo: "/reviews",
    });
    assert.equal(response.status, 204);
  });

  it("sends password signup, login, verification, and reset requests", async () => {
    const calls: Array<{ url: string; init: RequestInit }> = [];
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      calls.push({ url, init });
      if (url.endsWith("/login")) {
        return Promise.resolve(
          new Response(JSON.stringify({ email: "user@example.com" }), {
            headers: { "Content-Type": "application/json" },
            status: 200,
          }),
        );
      }
      return Promise.resolve(new Response(null, { status: 204 }));
    };
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });

    await client.requestPasswordSignup({
      email: "user@example.com",
      password: "a long enough password",
      displayName: "Learner",
    });
    const { data } = await client.loginWithPassword({
      email: "user@example.com",
      password: "a long enough password",
    });
    await client.verifyPasswordSignup({ token: "signup-token" });
    await client.requestPasswordReset({ email: "user@example.com" });
    await client.resetPassword({
      token: "reset-token",
      password: "another long password",
    });

    assert.equal(data.email, "user@example.com");
    assert.deepEqual(
      calls.map(({ url, init }) => ({
        path: new URL(url).pathname,
        method: init.method,
      })),
      [
        { path: "/api/v1/auth/password/signups", method: "POST" },
        { path: "/api/v1/auth/password/login", method: "POST" },
        { path: "/api/v1/auth/password/signups/verify", method: "POST" },
        { path: "/api/v1/auth/password/reset-requests", method: "POST" },
        { path: "/api/v1/auth/password/resets", method: "POST" },
      ],
    );
  });

  it("sends CSRF header on logout", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/auth/logout");
      assert.equal(init.method, "POST");
      assert.equal(
        new Headers(init.headers).get("X-CSRF-Token"),
        "csrf-token-value",
      );
      return Promise.resolve(new Response(null, { status: 204 }));
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { response } = await client.logout({
      headers: { "X-CSRF-Token": "csrf-token-value" },
    });
    assert.equal(response.status, 204);
  });

  it("throws ApiResponseError for problem+json errors", async () => {
    const fetch = (): Promise<Response> =>
      Promise.resolve(
        new Response(JSON.stringify({ detail: "authentication required" }), {
          headers: { "Content-Type": "application/problem+json" },
          status: 401,
        }),
      );

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    await assert.rejects(client.getCurrentUser(), (error: unknown) => {
      if (!(error instanceof ApiResponseError)) {
        return false;
      }
      assert.equal(error.status, 401);
      assert.equal(error.message, "authentication required");
      return true;
    });
  });

  it("sends GET /healthz and parses kill_switches without throwing on 503", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/healthz");
      assert.equal(init.method, "GET");
      assert.equal(new Headers(init.headers).get("Accept"), "application/json");
      return Promise.resolve(
        new Response(
          JSON.stringify({
            status: "unhealthy",
            database: "unhealthy",
            kill_switches: { oauth_enabled: true },
          }),
          {
            headers: { "Content-Type": "application/json" },
            status: 503,
          },
        ),
      );
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data, response } = await client.getHealthz();
    assert.equal(response.status, 503);
    assert.equal(data.kill_switches?.oauth_enabled, true);
  });

  it("sends GET /api/v1/journey-situations", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/journey-situations");
      assert.equal(init.method, "GET");
      return Promise.resolve(
        new Response(
          JSON.stringify({
            items: [
              {
                id: "00000000-0000-0000-0000-000000000001",
                slug: "airport",
                title: "Airport",
                shortDescription: "Airport words.",
                category: "travel",
                displayOrder: 1,
              },
            ],
          }),
          { headers: { "Content-Type": "application/json" }, status: 200 },
        ),
      );
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.listJourneySituations();
    assert.equal(data.items[0]!.slug, "airport");
  });

  it("sends GET /api/v1/journey-situations/{slug}", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(
        url,
        "https://api.example.com/api/v1/journey-situations/airport",
      );
      assert.equal(init.method, "GET");
      return Promise.resolve(
        new Response(
          JSON.stringify({
            situation: {
              id: "00000000-0000-0000-0000-000000000001",
              slug: "airport",
              title: "Airport",
              shortDescription: "Airport words.",
              category: "travel",
              displayOrder: 1,
            },
            meanings: [
              {
                meaningId: "00000000-0000-0000-0000-000000000002",
                wordId: "00000000-0000-0000-0000-000000000003",
                wordSlug: "boarding-pass",
                wordText: "boarding pass",
                partOfSpeech: "noun",
                shortDefinition: "A document.",
                saved: false,
              },
            ],
          }),
          { headers: { "Content-Type": "application/json" }, status: 200 },
        ),
      );
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.getJourneySituation("airport");
    assert.equal(data.situation.slug, "airport");
    assert.equal(data.meanings[0]!.wordSlug, "boarding-pass");
  });

  it("sends GET /api/v1/canonical-words/{wordSlug}", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(
        url,
        "https://api.example.com/api/v1/canonical-words/boarding-pass",
      );
      assert.equal(init.method, "GET");
      return Promise.resolve(
        new Response(
          JSON.stringify({
            word: {
              id: "00000000-0000-0000-0000-000000000003",
              text: "boarding pass",
              slug: "boarding-pass",
              wordType: "phrase",
              meanings: [
                {
                  id: "00000000-0000-0000-0000-000000000002",
                  partOfSpeech: "noun",
                  shortDefinition: "A document.",
                  saved: true,
                  examples: [],
                  usageNotes: [],
                },
              ],
            },
          }),
          { headers: { "Content-Type": "application/json" }, status: 200 },
        ),
      );
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.getCanonicalWord("boarding-pass");
    assert.equal(data.word.slug, "boarding-pass");
    assert.equal(data.word.meanings[0]!.saved, true);
  });

  it("sends GET /api/v1/user-words", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/user-words");
      assert.equal(init.method, "GET");
      return Promise.resolve(
        new Response(
          JSON.stringify({
            items: [
              {
                userWordId: "00000000-0000-0000-0000-000000000001",
                meaningId: "00000000-0000-0000-0000-000000000002",
                wordId: "00000000-0000-0000-0000-000000000003",
                wordText: "boarding pass",
                wordSlug: "boarding-pass",
                partOfSpeech: "noun",
                shortDefinition: "A document.",
                status: "new",
                source: "journey",
                saved: true,
                addedAt: "2026-07-25T12:00:00Z",
              },
            ],
          }),
          { headers: { "Content-Type": "application/json" }, status: 200 },
        ),
      );
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.listSavedWords();
    assert.equal(data.items[0]!.wordSlug, "boarding-pass");
  });

  it("sends GET /api/v1/reviews/due", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/reviews/due?limit=2");
      assert.equal(init.method, "GET");
      return Promise.resolve(
        new Response(
          JSON.stringify({
            items: [
              {
                userWordId: "00000000-0000-0000-0000-000000000001",
                meaningId: "00000000-0000-0000-0000-000000000002",
                wordId: "00000000-0000-0000-0000-000000000003",
                wordText: "boarding pass",
                wordSlug: "boarding-pass",
                partOfSpeech: "noun",
                shortDefinition: "A document.",
                status: "new",
                reviewStep: 0,
              },
            ],
            nextCursor: "c",
            totalCount: 1,
          }),
          { headers: { "Content-Type": "application/json" }, status: 200 },
        ),
      );
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.listDueWords({ limit: 2 });
    assert.equal(data.items[0]!.wordSlug, "boarding-pass");
    assert.equal(data.items[0]!.reviewStep, 0);
    assert.equal(data.totalCount, 1);
  });

  it("sends GET /api/v1/user-words/records/{userWordId}", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(
        url,
        "https://api.example.com/api/v1/user-words/records/00000000-0000-0000-0000-000000000001",
      );
      assert.equal(init.method, "GET");
      return Promise.resolve(
        new Response(
          JSON.stringify({
            userWordId: "00000000-0000-0000-0000-000000000001",
            meaningId: "00000000-0000-0000-0000-000000000002",
            wordId: "00000000-0000-0000-0000-000000000003",
            wordText: "boarding pass",
            wordSlug: "boarding-pass",
            partOfSpeech: "noun",
            shortDefinition: "A document.",
            status: "new",
            source: "journey",
            saved: true,
            addedAt: "2026-07-25T12:00:00Z",
          }),
          { headers: { "Content-Type": "application/json" }, status: 200 },
        ),
      );
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.getSavedWord(
      "00000000-0000-0000-0000-000000000001",
    );
    assert.equal(data.meaningId, "00000000-0000-0000-0000-000000000002");
  });

  it("sends POST /api/v1/user-words with Idempotency-Key", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/user-words");
      assert.equal(init.method, "POST");
      assert.equal(
        new Headers(init.headers).get("Idempotency-Key"),
        "idem-key",
      );
      return Promise.resolve(
        new Response(
          JSON.stringify({
            userWordId: "00000000-0000-0000-0000-000000000001",
            meaningId: "00000000-0000-0000-0000-000000000002",
            wordId: "00000000-0000-0000-0000-000000000003",
            wordText: "boarding pass",
            wordSlug: "boarding-pass",
            partOfSpeech: "noun",
            shortDefinition: "A document.",
            status: "new",
            source: "journey",
            saved: true,
            addedAt: "2026-07-25T12:00:00Z",
          }),
          { headers: { "Content-Type": "application/json" }, status: 200 },
        ),
      );
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.saveUserWord(
      { meaningId: "00000000-0000-0000-0000-000000000002", source: "journey" },
      "idem-key",
    );
    assert.equal(data.wordSlug, "boarding-pass");
  });

  it("sends DELETE /api/v1/user-words/{meaningId}", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(
        url,
        "https://api.example.com/api/v1/user-words/00000000-0000-0000-0000-000000000002",
      );
      assert.equal(init.method, "DELETE");
      return Promise.resolve(new Response(null, { status: 204 }));
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { response } = await client.unsaveUserWord(
      "00000000-0000-0000-0000-000000000002",
    );
    assert.equal(response.status, 204);
  });

  it("sends POST /api/v1/reviews/submissions with Idempotency-Key", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/reviews/submissions");
      assert.equal(init.method, "POST");
      assert.equal(
        new Headers(init.headers).get("Idempotency-Key"),
        "idem-key",
      );
      return Promise.resolve(
        new Response(
          JSON.stringify({
            attemptId: "00000000-0000-0000-0000-000000000001",
            userWordId: "00000000-0000-0000-0000-000000000002",
            meaningId: "00000000-0000-0000-0000-000000000003",
            attemptType: "review",
            promptType: "multiple_choice",
            result: "correct",
            rating: "good",
            reviewStepBefore: 0,
            reviewStepAfter: 1,
            answeredAt: "2026-07-25T12:00:00Z",
            responseTimeMs: 1234,
            wasHintUsed: false,
            source: "review",
            clientAttemptId: "ca-1",
            nextReviewAt: "2026-07-25T13:00:00Z",
          }),
          { headers: { "Content-Type": "application/json" }, status: 200 },
        ),
      );
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.submitReview(
      {
        userWordId: "00000000-0000-0000-0000-000000000002",
        meaningId: "00000000-0000-0000-0000-000000000003",
        promptType: "multiple_choice",
        result: "correct",
        rating: "good",
        answeredAt: "2026-07-25T12:00:00Z",
        clientAttemptId: "ca-1",
      },
      "idem-key",
    );
    assert.equal(data.reviewStepAfter, 1);
    assert.equal(data.nextReviewAt, "2026-07-25T13:00:00Z");
  });

  it("sends POST /api/v1/learner-sentences with Idempotency-Key", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/learner-sentences");
      assert.equal(init.method, "POST");
      assert.equal(
        new Headers(init.headers).get("Idempotency-Key"),
        "idem-key",
      );
      assert.equal(
        init.body,
        JSON.stringify({
          sentenceText: "I work every day.",
          source: "word_detail",
          attemptId: "00000000-0000-0000-0000-000000000002",
        }),
      );
      return Promise.resolve(
        new Response(
          JSON.stringify({
            sentenceId: "00000000-0000-0000-0000-000000000010",
            attemptId: "00000000-0000-0000-0000-000000000011",
            status: "correct",
            originalSentence: "I work every day.",
            explanation: "The sentence uses the target word correctly.",
            missionCompleted: false,
            canRetry: false,
            reported: false,
          }),
          { headers: { "Content-Type": "application/json" }, status: 200 },
        ),
      );
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.submitSentenceFeedback(
      {
        sentenceText: "I work every day.",
        source: "word_detail",
        attemptId: "00000000-0000-0000-0000-000000000002",
      },
      "idem-key",
    );
    assert.equal(data.status, "correct");
    assert.equal(data.originalSentence, "I work every day.");
    assert.equal(data.missionCompleted, false);
  });

  it("lists and gets retained learner sentences", async () => {
    const sentence = {
      id: "00000000-0000-0000-0000-000000000010",
      processingStatus: "completed",
      status: "correct",
      originalSentence: "I work every day.",
      targetWordUsedCorrectly: true,
      grammarAcceptable: true,
      meaningClear: true,
      naturalness: "natural",
      reported: false,
      createdAt: "2026-09-10T12:00:00Z",
    };
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(init.method, "GET");
      if (url.endsWith("?after=next-page&limit=10")) {
        return Promise.resolve(
          Response.json({ items: [sentence], hasMore: false }),
        );
      }
      if (url.endsWith("?limit=0")) {
        return Promise.resolve(
          Response.json({ items: [sentence], hasMore: false }),
        );
      }
      assert.equal(
        url,
        "https://api.example.com/api/v1/learner-sentences/00000000-0000-0000-0000-000000000010",
      );
      return Promise.resolve(Response.json(sentence));
    };
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });

    const page = await client.listLearnerSentences({
      after: "next-page",
      limit: 10,
    });
    assert.equal(page.data.items[0]?.status, "correct");
    const explicitDefault = await client.listLearnerSentences({ limit: 0 });
    assert.equal(explicitDefault.data.items[0]?.status, "correct");
    const detail = await client.getLearnerSentence(sentence.id);
    assert.equal(detail.data.originalSentence, "I work every day.");
  });

  it("sends POST /api/v1/sentence-feedback/{attemptId}/reports", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(
        url,
        "https://api.example.com/api/v1/sentence-feedback/00000000-0000-0000-0000-000000000011/reports",
      );
      assert.equal(init.method, "POST");
      assert.equal(
        new Headers(init.headers).get("Idempotency-Key"),
        "idem-key",
      );
      assert.equal(
        init.body,
        JSON.stringify({
          reason: "already_correct",
        }),
      );
      return Promise.resolve(new Response(null, { status: 204 }));
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { response } = await client.reportSentenceFeedback(
      "00000000-0000-0000-0000-000000000011",
      { reason: "already_correct" },
      "idem-key",
    );
    assert.equal(response.status, 204);
  });

  it("sends GET /api/v1/daily-mission", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/daily-mission");
      assert.equal(init.method, "GET");
      return Promise.resolve(
        new Response(
          JSON.stringify({
            localDate: "2026-07-26",
            timezone: "UTC",
            reviewTarget: 20,
            reviewsCompleted: 5,
            policyVersion: "p4-mission-policy-v1",
            status: "open",
            graceApplied: false,
            streak: {
              currentStreakCount: 0,
              longestStreakCount: 0,
              status: "active",
              graceDayBalance: 0,
            },
          }),
          { headers: { "Content-Type": "application/json" }, status: 200 },
        ),
      );
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.getDailyMission();
    assert.equal(data.localDate, "2026-07-26");
    assert.equal(data.reviewTarget, 20);
    assert.equal(data.streak.status, "active");
  });

  it("sends GET /api/v1/daily-mission?timezone=America/New_York", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(
        url,
        "https://api.example.com/api/v1/daily-mission?timezone=America%2FNew_York",
      );
      assert.equal(init.method, "GET");
      return Promise.resolve(
        new Response(
          JSON.stringify({
            localDate: "2026-07-26",
            timezone: "America/New_York",
            reviewTarget: 20,
            reviewsCompleted: 0,
            policyVersion: "p4-mission-policy-v1",
            status: "open",
            graceApplied: false,
            streak: {
              currentStreakCount: 0,
              longestStreakCount: 0,
              status: "active",
              graceDayBalance: 0,
            },
          }),
          { headers: { "Content-Type": "application/json" }, status: 200 },
        ),
      );
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.getDailyMission({
      timezone: "America/New_York",
    });
    assert.equal(data.timezone, "America/New_York");
  });

  it("sends GET /api/v1/progress", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/progress");
      assert.equal(init.method, "GET");
      return Promise.resolve(
        new Response(
          JSON.stringify({
            confidencePointsBalance: 42,
            streak: {
              currentStreakCount: 3,
              longestStreakCount: 7,
              status: "active",
              graceDayBalance: 1,
            },
            completionHistory: [
              { localDate: "2026-07-20", completed: true, status: "completed" },
              { localDate: "2026-07-21", completed: true, status: "protected" },
              { localDate: "2026-07-22", completed: false, status: "missed" },
              { localDate: "2026-07-23", completed: true },
              { localDate: "2026-07-24", completed: true },
              { localDate: "2026-07-25", completed: true },
              { localDate: "2026-07-26", completed: false },
            ],
          }),
          { headers: { "Content-Type": "application/json" }, status: 200 },
        ),
      );
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.getProgress();
    assert.equal(data.confidencePointsBalance, 42);
    assert.equal(data.streak.currentStreakCount, 3);
    assert.equal(data.completionHistory.length, 7);
    assert.equal(data.completionHistory[0]?.completed, true);
    assert.equal(data.completionHistory[0]?.status, "completed");
    assert.equal(data.completionHistory[1]?.completed, true);
    assert.equal(data.completionHistory[1]?.status, "protected");
    assert.equal(data.completionHistory[2]?.status, "missed");
    assert.equal(data.completionHistory[3]?.status, undefined);
  });

  it("sends GET /api/v1/onboarding", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/onboarding");
      assert.equal(init.method, "GET");
      return Promise.resolve(
        new Response(JSON.stringify({ status: "not_started" }), {
          headers: { "Content-Type": "application/json" },
          status: 200,
        }),
      );
    };
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.getOnboarding();
    assert.equal(data.status, "not_started");
  });

  it("sends POST /api/v1/onboarding with full submission body", async () => {
    let capturedBody: unknown;
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/onboarding");
      assert.equal(init.method, "POST");
      capturedBody = JSON.parse(String(init.body));
      return Promise.resolve(
        new Response(
          JSON.stringify({
            status: "completed",
            englishLevel: "b1",
            nativeLanguage: "es",
            learningGoal: "general",
            mainUseCase: "daily_life",
            dailyReviewTarget: 25,
            completedAt: "2026-07-27T12:00:00Z",
          }),
          { headers: { "Content-Type": "application/json" }, status: 200 },
        ),
      );
    };
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.completeOnboarding({
      englishLevel: "b1",
      nativeLanguage: "es",
      learningGoal: "general",
      mainUseCase: "daily_life",
      dailyReviewTarget: 25,
    });
    assert.deepEqual(capturedBody, {
      englishLevel: "b1",
      nativeLanguage: "es",
      learningGoal: "general",
      mainUseCase: "daily_life",
      dailyReviewTarget: 25,
    });
    assert.equal(data.status, "completed");
    assert.equal(data.dailyReviewTarget, 25);
  });

  it("sends GET /api/v1/settings", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/settings");
      assert.equal(init.method, "GET");
      return Promise.resolve(
        new Response(
          JSON.stringify({
            dailyReviewTarget: 20,
            reviewIntervalPreset: "vocanova_default",
            appLanguage: "en",
            notificationsEnabled: true,
            marketingEmailsEnabled: false,
            displayName: "",
          }),
          { headers: { "Content-Type": "application/json" }, status: 200 },
        ),
      );
    };
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.getSettings();
    assert.equal(data.dailyReviewTarget, 20);
    assert.equal(data.reviewIntervalPreset, "vocanova_default");
    assert.equal(data.appLanguage, "en");
    assert.equal(data.notificationsEnabled, true);
    assert.equal(data.marketingEmailsEnabled, false);
    assert.equal(data.displayName, "");
  });

  it("sends PATCH /api/v1/settings with partial body and CSRF header", async () => {
    let capturedBody: unknown;
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/settings");
      assert.equal(init.method, "PATCH");
      assert.equal(
        new Headers(init.headers).get("X-CSRF-Token"),
        "csrf-token-value",
      );
      assert.equal(
        new Headers(init.headers).get("Content-Type"),
        "application/json",
      );
      capturedBody = JSON.parse(String(init.body));
      return Promise.resolve(
        new Response(
          JSON.stringify({
            dailyReviewTarget: 35,
            reviewIntervalPreset: "wordup_like",
            appLanguage: "en",
            notificationsEnabled: false,
            marketingEmailsEnabled: true,
            displayName: "Ada",
          }),
          { headers: { "Content-Type": "application/json" }, status: 200 },
        ),
      );
    };
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.updateSettings(
      {
        dailyReviewTarget: 35,
        reviewIntervalPreset: "wordup_like",
        notificationsEnabled: false,
        marketingEmailsEnabled: true,
        displayName: "Ada",
      },
      { headers: { "X-CSRF-Token": "csrf-token-value" } },
    );
    assert.deepEqual(capturedBody, {
      dailyReviewTarget: 35,
      reviewIntervalPreset: "wordup_like",
      notificationsEnabled: false,
      marketingEmailsEnabled: true,
      displayName: "Ada",
    });
    assert.equal(data.dailyReviewTarget, 35);
    assert.equal(data.displayName, "Ada");
  });

  it("sends PATCH /api/v1/settings with empty body for no-op read", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/settings");
      assert.equal(init.method, "PATCH");
      assert.equal(init.body, JSON.stringify({}));
      return Promise.resolve(
        new Response(
          JSON.stringify({
            dailyReviewTarget: 20,
            reviewIntervalPreset: "vocanova_default",
            appLanguage: "en",
            notificationsEnabled: true,
            marketingEmailsEnabled: false,
            displayName: "",
          }),
          { headers: { "Content-Type": "application/json" }, status: 200 },
        ),
      );
    };
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.updateSettings({});
    assert.equal(data.dailyReviewTarget, 20);
  });

  it("throws ApiResponseError for /api/v1/settings 422", async () => {
    const fetch = (): Promise<Response> =>
      Promise.resolve(
        new Response(
          JSON.stringify({
            detail: "daily review target 200 out of range [5,100]",
          }),
          {
            headers: { "Content-Type": "application/problem+json" },
            status: 422,
          },
        ),
      );
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    await assert.rejects(
      client.updateSettings({ dailyReviewTarget: 200 }),
      (error: unknown) => {
        if (!(error instanceof ApiResponseError)) {
          return false;
        }
        assert.equal(error.status, 422);
        assert.equal(
          error.message,
          "daily review target 200 out of range [5,100]",
        );
        return true;
      },
    );
  });

  it("sends POST /api/v1/settings/email-change-links with newEmail", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(
        url,
        "https://api.example.com/api/v1/settings/email-change-links",
      );
      assert.equal(init.method, "POST");
      assert.equal(
        new Headers(init.headers).get("X-CSRF-Token"),
        "csrf-token-value",
      );
      assert.equal(
        init.body,
        JSON.stringify({ newEmail: "new-address@example.com" }),
      );
      return Promise.resolve(new Response(null, { status: 204 }));
    };
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { response } = await client.requestEmailChangeLink(
      { newEmail: "new-address@example.com" },
      { headers: { "X-CSRF-Token": "csrf-token-value" } },
    );
    assert.equal(response.status, 204);
  });

  it("sends POST /api/v1/settings/email-change-links/consume with token", async () => {
    let capturedBody: unknown;
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(
        url,
        "https://api.example.com/api/v1/settings/email-change-links/consume",
      );
      assert.equal(init.method, "POST");
      assert.equal(
        new Headers(init.headers).get("X-CSRF-Token"),
        "csrf-token-value",
      );
      capturedBody = JSON.parse(String(init.body));
      return Promise.resolve(
        new Response(
          JSON.stringify({
            email: "new-address@example.com",
            previousEmail: "old-address@example.com",
            changedAt: "2026-07-27T12:00:00Z",
          }),
          { headers: { "Content-Type": "application/json" }, status: 200 },
        ),
      );
    };
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.consumeEmailChangeLink(
      { token: "the-token" },
      { headers: { "X-CSRF-Token": "csrf-token-value" } },
    );
    assert.deepEqual(capturedBody, { token: "the-token" });
    assert.equal(data.email, "new-address@example.com");
    assert.equal(data.previousEmail, "old-address@example.com");
    assert.equal(data.changedAt, "2026-07-27T12:00:00Z");
  });

  it("sends POST /api/v1/account-deletion-requests with Idempotency-Key and CSRF", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(
        url,
        "https://api.example.com/api/v1/account-deletion-requests",
      );
      assert.equal(init.method, "POST");
      assert.equal(
        new Headers(init.headers).get("Idempotency-Key"),
        "idem-key",
      );
      assert.equal(
        new Headers(init.headers).get("X-CSRF-Token"),
        "csrf-token-value",
      );
      return Promise.resolve(
        new Response(
          JSON.stringify({
            status: "deactivated",
            userId: "00000000-0000-0000-0000-000000000001",
            requestedAt: "2026-07-27T12:00:00Z",
            purgeAfter: "2026-08-26T12:00:00Z",
            idempotencyKey: "idem-key",
            replayed: false,
          }),
          { headers: { "Content-Type": "application/json" }, status: 200 },
        ),
      );
    };
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.createAccountDeletionRequest("idem-key", {
      headers: { "X-CSRF-Token": "csrf-token-value" },
    });
    assert.equal(data.status, "deactivated");
    assert.equal(data.replayed, false);
    assert.equal(data.purgeAfter, "2026-08-26T12:00:00Z");
  });

  it("sends POST /api/v1/personal-data-export with Idempotency-Key and CSRF", async () => {
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: async (input, init) => {
        assert.equal(
          input,
          "https://api.example.com/api/v1/personal-data-export",
        );
        assert.equal(init?.method, "POST");
        assert.equal(
          new Headers(init?.headers).get("Idempotency-Key"),
          "export-key",
        );
        assert.equal(
          new Headers(init?.headers).get("X-CSRF-Token"),
          "csrf-token-value",
        );
        return new Response(
          JSON.stringify({
            schemaVersion: "1.0",
            profile: {},
            settings: {},
            onboardingProfile: null,
            savedWords: [],
            reviewHistory: [],
            sentenceFeedbackHistory: [],
            dailyMissions: [],
            dailyActivity: [],
            confidencePointLedger: [],
            graceDayLedger: [],
            streakState: null,
          }),
        );
      },
    });
    const { data } = await client.exportPersonalData("export-key", {
      headers: { "X-CSRF-Token": "csrf-token-value" },
    });
    assert.equal(data.schemaVersion, "1.0");
  });

  // VOC-031-T06: the session-expiry mid-flow handler at
  // apps/web/src/lib/session.ts is a thin wrapper around
  // ApiResponseError.status === 401. The cross-cutting
  // guarantee (TEST-29) depends on this detection being
  // stable across the whole (app) surface, so the test below
  // pins the detection pattern the helper relies on. A
  // regression where ApiResponseError stopped reporting 401
  // would surface here before it reached the client.
  it("exposes a stable 401 detection for the session-expiry mid-flow helper", () => {
    const isSessionExpiredError = (error: unknown): boolean =>
      error instanceof ApiResponseError && error.status === 401;

    // 401 with a problem+json body must be detected.
    const expiredSession = new ApiResponseError(401, {
      detail: "authentication required",
    });
    assert.equal(
      isSessionExpiredError(expiredSession),
      true,
      "401 must be detected as a session expiry",
    );

    // Other 4xx statuses must NOT be detected as a session
    // expiry — the helper specifically routes 401 to
    // re-auth, never 403 (CSRF) or 404 (not found) or
    // 409 (idempotency conflict), which have their own
    // per-screen handling.
    assert.equal(
      isSessionExpiredError(new ApiResponseError(403, { detail: "csrf" })),
      false,
      "403 must not be treated as a session expiry",
    );
    assert.equal(
      isSessionExpiredError(new ApiResponseError(404, { detail: "missing" })),
      false,
      "404 must not be treated as a session expiry",
    );
    assert.equal(
      isSessionExpiredError(new ApiResponseError(409, { detail: "conflict" })),
      false,
      "409 must not be treated as a session expiry",
    );
    assert.equal(
      isSessionExpiredError(new ApiResponseError(500, { detail: "oops" })),
      false,
      "500 must not be treated as a session expiry",
    );

    // Non-ApiResponseError values (network failures, JSON
    // parse errors, plain Error, undefined) must not be
    // detected as a session expiry.
    assert.equal(
      isSessionExpiredError(new Error("network failed")),
      false,
      "a plain Error must not be treated as a session expiry",
    );
    assert.equal(isSessionExpiredError(undefined), false);
    assert.equal(isSessionExpiredError(null), false);
    assert.equal(isSessionExpiredError("401"), false);
  });
});

describe("private list and story contracts", () => {
  it("sends list revisions, safe encoded IDs, retry keys and preserves current membership", async () => {
    const seen: { url: URL; init?: RequestInit }[] = [];
    const detail = {
      id: "list",
      name: "Travel",
      revision: 3,
      memberCount: 0,
      usableMemberCount: 0,
      members: [],
      createdAt: "now",
      updatedAt: "now",
    };
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: async (input, init) => {
        seen.push({ url: new URL(String(input)), init });
        return init?.method === "DELETE" && seen.length === 7
          ? new Response(null, { status: 204 })
          : new Response(JSON.stringify(detail), { status: 200 });
      },
    });
    await client.listWordLists();
    await client.getWordList("list/one");
    const created = await client.putWordList(
      "list/one",
      { name: "Travel", expectedRevision: 0 },
      "create-key",
      { headers: { "X-CSRF-Token": "token" } },
    );
    assert.deepEqual(created.data, detail);
    await client.putWordListMember(
      "list/one",
      "meaning/two",
      { expectedRevision: 1 },
      "add-key",
    );
    const removed = await client.deleteWordListMember(
      "list/one",
      "meaning/two",
      2,
      "remove-key",
    );
    assert.deepEqual(removed.data.members, []);
    await client.startPracticeSession(
      { mode: "typed_recall", listId: "list", listRevision: 3 },
      "start-key",
    );
    await client.deleteWordList("list/one", 3, "delete-key");
    assert.equal(seen[1]!.url.pathname, "/api/v1/word-lists/list%2Fone");
    assert.equal(
      seen[3]!.url.pathname,
      "/api/v1/word-lists/list%2Fone/members/meaning%2Ftwo",
    );
    assert.equal(seen[4]!.url.searchParams.get("expectedRevision"), "2");
    assert.equal(seen[6]!.url.searchParams.get("expectedRevision"), "3");
    assert.equal(
      new Headers(seen[2]!.init?.headers).get("Idempotency-Key"),
      "create-key",
    );
    assert.equal(
      new Headers(seen[2]!.init?.headers).get("X-CSRF-Token"),
      "token",
    );
    assert.deepEqual(JSON.parse(String(seen[5]!.init?.body)), {
      mode: "typed_recall",
      listId: "list",
      listRevision: 3,
    });
  });
  it("preserves story content, session action revisions and retry headers", async () => {
    const seen: { url: URL; init?: RequestInit }[] = [];
    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: async (input, init) => {
        seen.push({ url: new URL(String(input)), init });
        return new Response(JSON.stringify({ items: [] }), { status: 200 });
      },
    });
    await client.listStories();
    await client.getStory("travel/story");
    await client.startStorySession({ storyKey: "travel" }, "start");
    await client.getStorySession("session/one");
    await client.submitStoryAction(
      "session/one",
      {
        stepId: "question",
        expectedRevision: 2,
        clientActionId: "answer-one",
        action: "answer",
        choiceId: "one",
      },
      "answer",
    );
    assert.equal(seen[1]!.url.pathname, "/api/v1/stories/travel%2Fstory");
    assert.equal(seen[3]!.url.pathname, "/api/v1/story-sessions/session%2Fone");
    assert.equal(
      new Headers(seen[4]!.init?.headers).get("Idempotency-Key"),
      "answer",
    );
    assert.equal(JSON.parse(String(seen[4]!.init?.body)).expectedRevision, 2);
  });
});
