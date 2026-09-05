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
      assert.equal(new Headers(init.headers).get("Idempotency-Key"), "idem-key");
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
              { localDate: "2026-07-20", completed: true },
              { localDate: "2026-07-21", completed: true },
              { localDate: "2026-07-22", completed: false },
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
