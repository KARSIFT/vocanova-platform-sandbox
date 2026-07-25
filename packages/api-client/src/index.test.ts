import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { ApiResponseError, VocanovaClient } from "./index.js";

describe("VocanovaClient", () => {
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
      assert.equal(init.body, JSON.stringify({ email: "user@example.com" }));
      return Promise.resolve(new Response(null, { status: 204 }));
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { response } = await client.requestMagicLink({
      email: "user@example.com",
    });
    assert.equal(response.status, 204);
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

  it("sends POST /api/v1/sentence-feedback with Idempotency-Key", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/sentence-feedback");
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

  it("sends POST /api/v1/sentence-feedback/{attemptId}/reports", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(
        url,
        "https://api.example.com/api/v1/sentence-feedback/00000000-0000-0000-0000-000000000011/reports",
      );
      assert.equal(init.method, "POST");
      assert.equal(
        init.body,
        JSON.stringify({
          reason: "The feedback is incorrect.",
          classification: "incorrect",
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
      { reason: "The feedback is incorrect.", classification: "incorrect" },
    );
    assert.equal(response.status, 204);
  });
});
