import type {
  CurrentUser,
  DailyLesson,
  DailyMission,
  KnowledgeMap,
  KnowledgeMapItem,
  KnowledgeMapDecisionResult,
  Progress,
  WordDetailResponse,
} from "@vocanova/api-client";
// VOC-050-T02 - the core-loop journey run against a real deployed
// environment (staging), as deploy-staging.yml's post-deploy gate.
//
// Relationship to tests/e2e/core-loop.spec.ts (VOC-031-T08): same
// journey, different backend. That spec runs at PR time against
// mock-api-server.mjs. Both use the named situation vocabulary list for word
// navigation, keeping guidebook links separate. This spec runs after a staging deploy against the real Go API and real Postgres,
// so the two differ in the three places where the mock's shortcuts
// have no real-backend equivalent (specification.md's open
// questions 2 and 3):
//
//   1. Authentication. The mock accepts any vocanova_session value;
//      the real API requires a session row created by
//      CreateSession. This spec therefore consumes a session minted
//      server-side by VOC-050-T01's token-gated
//      POST /ops/synthetic-smoke-test/session for the seeded
//      synthetic account (VOC-050-T00), passed in as
//      E2E_SESSION_COOKIE / E2E_CSRF_TOKEN.
//   2. Onboarding. No e2e_onboarding_status override cookie is used
//      or expected - the real backend does not implement one, and
//      this task deliberately does not ask it to. The synthetic
//      account is seeded with onboarding_status = 'completed', so
//      the journey starts on /home. The onboarding form is still
//      driven for real if the app redirects there, so a
//      regression in onboarding is caught rather than hidden.
//   3. Fixture content. The mock serves fixed fixtures; staging
//      serves canonical content from Postgres. The save/review loop
//      remains content-agnostic (first situation, first word,
//      the supplied meaning choices). The context-practice phase
//      checks the authored Daily Conversation curriculum explicitly.
//
// Rerun safety: this journey runs on every staging deploy against
// an account that keeps its state between runs. Each step is
// written to be correct whether or not a previous run already left
// the account with saved words, a due-review backlog, or none.

import { expect, test } from "@playwright/test";
import type { Page, Route, TestInfo } from "@playwright/test";

import { verifyContextPracticeJourney } from "../e2e/context-practice-journey";
import { completeOnboardingIfRedirected } from "../e2e/onboarding-journey";
import {
  getEnabledReviewPromptControl,
  getTerminalReviewHeading,
  waitForReviewReadiness,
} from "../e2e/review-readiness";
import {
  eligibleReviewSeeds,
  MAX_REVIEW_SEED_CANDIDATES,
  REVIEW_SEED_BANDS,
  reviewedMeaningAnswer,
  reviewSeedActions,
} from "./review-seed";
import {
  chooseSituationWordLink,
  getFirstMeaningSaveControl,
} from "../e2e/situation-word-navigation";

const SESSION_COOKIE_ENV = "E2E_SESSION_COOKIE";
const CSRF_COOKIE_ENV = "E2E_CSRF_TOKEN";
// Deployed staging names are authoritative mint metadata. Require the exact
// staging names so a legacy or wrong-tier name never silently passes.
const SESSION_COOKIE_NAME = requiredCookieName(
  "E2E_SESSION_COOKIE_NAME",
  "vocanova_staging_session",
);
const CSRF_COOKIE_NAME = requiredCookieName(
  "E2E_CSRF_COOKIE_NAME",
  "vocanova_staging_csrf",
);

// The review session refetches after the last card in its batch, so
// "review everything due" is a loop. The bound keeps a large
// accumulated backlog from consuming the whole deploy budget; when
// it is hit, the journey still proves the review path works and
// records why it stopped short of the caught-up state.
const MAX_REVIEW_CARDS = 8;

type ReviewedTodayProgress = {
  reviewed: number;
  target: number;
};
// VOC-063-T01: bounded step-7 re-read when the first /home load may be stale.
// Four attempts (initial + three retries) with 1.5s between failures keeps the
// loop under ~5s of deliberate waiting, well inside the staging journey timeout.
const STEP_7_REVIEWED_COUNT_MAX_ATTEMPTS = 4;
const STEP_7_REVIEWED_COUNT_RETRY_DELAY_MS = 1_500;

// Cookies must be scoped to the shared parent domain, not just the web
// app's own host: the app's server-side rendering/fetch paths send these
// cookies on to api-staging.vocanova.site (a different subdomain from
// staging.vocanova.site), mirroring deploy-production.yml's own
// `SESSION_COOKIE_DOMAIN=.${HOST#*.}` computation and apps/api/.env.example's
// documented `.vocanova.site` convention. Using Playwright's `url`-only
// cookie form (as an earlier revision did) scopes the cookie to exactly
// the web app's host and never reaches the API host, which independent
// review flagged as a High finding blocking the real authenticated
// journey from working as wired.
function cookieDomainFor(baseURL: string): string {
  const override = process.env.STAGING_SESSION_COOKIE_DOMAIN;
  if (override) {
    return override;
  }
  const host = new URL(baseURL).hostname;
  const labels = host.split(".");
  return `.${labels.slice(1).join(".")}`;
}

function requiredEnv(name: string): string {
  const value = process.env[name];
  if (!value) {
    throw new Error(
      `${name} must be set for the staging core-loop journey. deploy-staging.yml supplies it from the VOC-050-T01 session-mint endpoint.`,
    );
  }
  return value;
}

function requiredCookieName(name: string, expected: string): string {
  const value = requiredEnv(name);
  if (value !== expected) {
    throw new Error(
      `${name} must be the deployed staging cookie name ${expected}.`,
    );
  }
  return value;
}

// Reads use Node fetch so failed Playwright APIRequestContext diagnostics never
// retain authenticated headers. No writes, redirects, fixtures or response logs.
async function readStagingState<T>(
  page: Page,
  apiOrigin: string,
  path: string,
): Promise<T> {
  try {
    const url = new URL(path, apiOrigin);
    if (url.origin !== apiOrigin || !url.pathname.startsWith("/api/v1/"))
      throw new Error("Outside staging scope");
    const cookies = (await page.context().cookies(apiOrigin)).filter(
      (cookie) => cookie.name === SESSION_COOKIE_NAME,
    );
    if (!cookies.length) throw new Error("Session missing");
    const response = await fetch(url, {
      headers: {
        Cookie: cookies
          .map((cookie) => `${cookie.name}=${cookie.value}`)
          .join("; "),
      },
      redirect: "error",
      signal: AbortSignal.timeout(20_000),
    });
    if (response.status !== 200) throw new Error("Staging read failed");
    return (await response.json()) as T;
  } catch {
    // Never preserve causes, headers, cookie values or response bodies.
    throw new Error("The bounded authenticated staging read failed.");
  }
}

function stagingOrigin(raw: string, hostname: string): string {
  const url = new URL(raw);
  if (
    url.protocol !== "https:" ||
    url.hostname !== hostname ||
    url.port ||
    url.username ||
    url.password ||
    url.pathname !== "/" ||
    url.search ||
    url.hash
  ) {
    throw new Error(
      `The core journey only accepts the HTTPS ${hostname} origin.`,
    );
  }
  return url.origin;
}

async function verifyRealKnowledgeMap(page: Page, apiOrigin: string) {
  const [map, beforeMission, beforeProgress] = await Promise.all([
    readStagingState<KnowledgeMap>(
      page,
      apiOrigin,
      "/api/v1/knowledge-map?band=100",
    ),
    readStagingState<DailyMission>(page, apiOrigin, "/api/v1/daily-mission"),
    readStagingState<Progress>(page, apiOrigin, "/api/v1/progress"),
  ]);
  expect(map.totalWords).toBe(10_000);
  expect(map.bandSize).toBe(100);
  expect(map.bands).toHaveLength(100);
  for (const [index, band] of map.bands.entries()) {
    expect([
      band.number,
      band.startRank,
      band.endRank,
      band.totalWords,
    ]).toEqual([index + 1, index * 100 + 1, (index + 1) * 100, 100]);
    expect(band.knownCount + band.learningCount + band.unseenCount).toBe(100);
  }
  expect(map.currentBand.number).toBe(100);
  expect(map.currentBand.items.map((item) => item.rank)).toEqual(
    Array.from({ length: 100 }, (_, index) => 9901 + index),
  );
  expect(
    new Set(map.currentBand.items.map((item) => item.meaningId)).size,
  ).toBe(100);

  // Never pause an existing saved review merely to exercise assessment. A new
  // known word is intentional synthetic-account state; reruns can safely
  // reapply Known once this bounded range has no unseen unsaved words left.
  const chosen =
    map.currentBand.items.find(
      (item) =>
        item.saved === false &&
        item.due === false &&
        item.selfReportedKnown === false,
    ) ??
    map.currentBand.items.find(
      (item) =>
        item.saved === false &&
        item.due === false &&
        item.selfReportedKnown === true,
    );
  expect(
    Boolean(chosen),
    "The synthetic account needs an unsaved word in the last range; do not reset its saved history.",
  ).toBe(true);
  if (!chosen) throw new Error("No safe synthetic map word available.");
  expect(Number.isSafeInteger(beforeMission.reviewsCompleted)).toBe(true);
  expect(Number.isSafeInteger(beforeProgress.confidencePointsBalance)).toBe(
    true,
  );

  // The word map is part of Journey; the earlier address still opens it.
  await page.goto("/knowledge-map");
  await expect(page).toHaveURL(/\/discover\?band=1$/);
  await expect(
    page.getByRole("heading", { name: "Journey", level: 1, exact: true }),
  ).toBeVisible();
  const ranges = page
    .getByRole("navigation", { name: "Word ranges", exact: true })
    .getByRole("link");
  await expect(ranges).toHaveCount(100);
  expect(
    await ranges.evaluateAll((links) =>
      links.map((link) => link.getAttribute("href")),
    ),
  ).toEqual(
    Array.from(
      { length: 100 },
      (_, index) => `/discover?band=${index + 1}#range-words`,
    ),
  );
  await ranges.last().click();
  await expect(page).toHaveURL(/\/discover\?band=100#range-words$/);
  await expect(
    page.getByRole("heading", { name: "Words 9,901–10,000", exact: true }),
  ).toBeVisible();
  const range = page.getByRole("region", {
    name: "Words 9,901–10,000",
    exact: true,
  });
  // The range keeps its word tiles behind the "Words in this range" disclosure.
  const openRangeWords = async () => {
    const summary = range
      .locator("summary")
      .filter({ hasText: "Words in this range" });
    if ((await summary.locator("..").getAttribute("open")) === null)
      await summary.click();
    await expect(summary.locator("..")).toHaveAttribute("open", "");
  };
  const tile = () =>
    range.locator(`a[href="/discovery?band=100&word=${chosen.rank}"]`);
  await openRangeWords();
  await expect(
    tile().getByText(chosen.wordText, { exact: true }),
  ).toBeVisible();
  await tile().click();
  await page
    .getByRole("button", { name: "Start discovering", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: chosen.wordText, level: 1, exact: true }),
  ).toBeFocused();
  await expect(
    page
      .getByRole("article", { name: chosen.wordText, exact: true })
      .getByText(chosen.shortDefinition, { exact: true }),
  ).toBeVisible();

  // Pass through to the real API; block an incorrectly configured frontend
  // before this new assessment can write to any origin except staging.
  let forbiddenWrite = false;
  const guard = async (route: Route) => {
    if (new URL(route.request().url()).origin !== apiOrigin) {
      forbiddenWrite = true;
      await route.abort("blockedbyclient");
    } else await route.continue();
  };
  const decisionPattern = "**/api/v1/knowledge-map/decisions";
  await page.route(decisionPattern, guard);
  try {
    const acknowledged = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        new URL(response.url()).origin === apiOrigin &&
        new URL(response.url()).pathname === "/api/v1/knowledge-map/decisions",
    );
    await page
      .getByRole("button", { name: "Already know", exact: true })
      .click();
    const response = await acknowledged;
    expect(
      forbiddenWrite,
      "Assessment must only write to the staging API.",
    ).toBe(false);
    expect(response.status()).toBe(200);
    expect(response.request().postDataJSON()).toEqual({
      meaningId: chosen.meaningId,
      decision: "known",
      expectedRevision: chosen.revision,
    });
    expect(Boolean(response.request().headers()["idempotency-key"])).toBe(true);
    const result = (await response.json()) as KnowledgeMapDecisionResult;
    expect(result.meaningId).toBe(chosen.meaningId);
    expect(result.decision).toBe("known");
    expect(result.revision).toMatch(/^[a-f0-9]{64}$/);

    // Guided discovery skips assessed gaps. Derive the next unseen word from
    // fresh real server state, rather than assuming rank + 1 is still new.
    const acknowledgedMap = await readStagingState<KnowledgeMap>(
      page,
      apiOrigin,
      "/api/v1/knowledge-map?band=100",
    );
    const currentUnseen = acknowledgedMap.currentBand.items.find(
      (item) => !item.selfReportedKnown && !item.saved,
    );
    const nextBand = currentUnseen
      ? acknowledgedMap.currentBand.number
      : acknowledgedMap.bands.find((band) => band.unseenCount > 0)?.number;
    if (nextBand === undefined) {
      await expect(
        page.getByRole("heading", {
          name: "You’ve explored the whole map",
          exact: true,
        }),
      ).toBeVisible();
      await page.getByRole("link", { name: "View map", exact: true }).click();
    } else {
      const nextMap =
        nextBand === acknowledgedMap.currentBand.number
          ? acknowledgedMap
          : await readStagingState<KnowledgeMap>(
              page,
              apiOrigin,
              `/api/v1/knowledge-map?band=${nextBand}`,
            );
      const next = nextMap.currentBand.items.find(
        (item) => !item.selfReportedKnown && !item.saved,
      );
      expect(
        next,
        "A range reported as unseen must contain a new word.",
      ).toBeDefined();
      if (!next) throw new Error("The real next unseen word is missing.");
      await expect(
        page.getByRole("heading", {
          name: next.wordText,
          level: 1,
          exact: true,
        }),
      ).toBeFocused();
      await expect(page).toHaveURL(
        new RegExp(`/discovery\\?band=${nextBand}&word=${next.rank}$`),
      );
      // The session's exit returns to the range now on screen.
      const back = page.getByRole("link", {
        name: "Back to Journey",
        exact: true,
      });
      await expect(back).toHaveAttribute(
        "href",
        `/discover?band=${nextBand}#range-words`,
      );
      await back.click();
      await expect(page).toHaveURL(
        new RegExp(`/discover\\?band=${nextBand}#range-words$`),
      );
      if (nextBand !== 100) {
        await page
          .getByRole("navigation", { name: "Word ranges", exact: true })
          .getByRole("link")
          .last()
          .click();
        await expect(page).toHaveURL(/\/discover\?band=100#range-words$/);
      }
    }
    await openRangeWords();
    await expect(
      tile().getByText("Already know", { exact: true }),
    ).toBeVisible();
    await page.reload();
    await openRangeWords();
    await expect(
      tile().getByText(chosen.wordText, { exact: true }),
    ).toBeVisible();
    await expect(
      tile().getByText("Already know", { exact: true }),
    ).toBeVisible();

    const [persisted, afterMission, afterProgress] = await Promise.all([
      readStagingState<KnowledgeMap>(
        page,
        apiOrigin,
        "/api/v1/knowledge-map?band=100",
      ),
      readStagingState<DailyMission>(page, apiOrigin, "/api/v1/daily-mission"),
      readStagingState<Progress>(page, apiOrigin, "/api/v1/progress"),
    ]);
    const saved = persisted.currentBand.items.find(
      (item) => item.meaningId === chosen.meaningId,
    );
    expect(saved?.selfReportedKnown).toBe(true);
    expect(saved?.saved).toBe(false);
    expect(saved?.due).toBe(false);
    expect(saved?.revision).toBe(result.revision);
    expect(persisted.bands[99]!.knownCount).toBe(
      map.bands[99]!.knownCount + (chosen.selfReportedKnown ? 0 : 1),
    );
    expect(persisted.bands[99]!.learningCount).toBe(
      map.bands[99]!.learningCount,
    );
    expect(
      afterMission.localDate,
      "A date change prevents a stable assessment-credit comparison.",
    ).toBe(beforeMission.localDate);
    expect(
      afterMission.reviewsCompleted,
      "Self-assessment must not grant review credit.",
    ).toBe(beforeMission.reviewsCompleted);
    expect(
      afterProgress.confidencePointsBalance,
      "Self-assessment must not award confidence points, even when the daily target is already capped.",
    ).toBe(beforeProgress.confidencePointsBalance);
    expect(afterMission.streak.currentStreakCount).toBe(
      beforeMission.streak.currentStreakCount,
    );
  } finally {
    await page.unroute(decisionPattern, guard);
  }
  await page
    .getByRole("navigation", { name: "Primary", exact: true })
    .getByRole("link", { name: "Home", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "Today", exact: true }),
  ).toBeVisible();
}

async function readReviewedTodayProgress(
  page: Page,
): Promise<ReviewedTodayProgress> {
  const apiOrigin = stagingOrigin(
    process.env.STAGING_API_BASE_URL ?? "https://api-staging.vocanova.site",
    "api-staging.vocanova.site",
  );
  const mission = await readStagingState<DailyMission>(
    page,
    apiOrigin,
    "/api/v1/daily-mission",
  );
  const counter = page
    .getByRole("region", { name: "Review words", exact: true })
    .getByText(/^Reviewed today:/);
  await expect(counter).toHaveText(
    `Reviewed today: ${mission.reviewsCompleted}`,
  );
  const reviewed = mission.reviewsCompleted;
  const target = mission.reviewTarget;
  expect(Number.isFinite(reviewed)).toBe(true);
  expect(target).toBeGreaterThan(0);
  return { reviewed, target };
}

async function readReviewedTodayCount(page: Page): Promise<number> {
  const { reviewed } = await readReviewedTodayProgress(page);
  return reviewed;
}

async function readReviewedTodayCountAfterReviews(
  page: Page,
  testInfo: TestInfo,
  reviewedBefore: number,
  reviewedCards: number,
  reviewTarget: number,
): Promise<number> {
  // reviews_completed caps at review_target; extra submissions after the
  // completing review do not raise the home counter (VOC-082-T01).
  const minimumExpected = Math.min(
    reviewedBefore + reviewedCards,
    reviewTarget,
  );
  const attemptValues: number[] = [];

  for (
    let attempt = 1;
    attempt <= STEP_7_REVIEWED_COUNT_MAX_ATTEMPTS;
    attempt++
  ) {
    await page.goto("/home");
    const reviewedAfter = await readReviewedTodayCount(page);
    attemptValues.push(reviewedAfter);

    if (reviewedAfter >= minimumExpected) {
      const countsSummary =
        `reviewedBefore=${reviewedBefore}, reviewedCards=${reviewedCards}, ` +
        `reviewedAfter=${reviewedAfter}, minimumExpected=${minimumExpected}`;
      testInfo.annotations.push({
        type: "step-7-reviewed-counts",
        description:
          attempt > 1
            ? `${countsSummary}, attempts=${attempt}, values=${attemptValues.join(",")}`
            : countsSummary,
      });
      console.log(
        `[staging core-loop] step 7 reviewed counts: ${countsSummary}`,
      );
      if (attempt > 1) {
        testInfo.annotations.push({
          type: "step-7-retry",
          description:
            `reviewed-count read needed ${attempt} attempt(s) ` +
            `(values=${attemptValues.join(",")})`,
        });
      }
      return reviewedAfter;
    }

    if (attempt < STEP_7_REVIEWED_COUNT_MAX_ATTEMPTS) {
      await page.waitForTimeout(STEP_7_REVIEWED_COUNT_RETRY_DELAY_MS);
    }
  }

  throw new Error(
    `Step 7 reviewed-today counter did not reach the expected minimum after ` +
      `${STEP_7_REVIEWED_COUNT_MAX_ATTEMPTS} attempt(s): ` +
      `reviewedBefore=${reviewedBefore}, reviewedCards=${reviewedCards}, ` +
      `minimumExpected=${minimumExpected}, observed=[${attemptValues.join(", ")}]`,
  );
}

// Save a fresh public meaning through the real UI. Never remove a saved word,
// clear an audit, reset clocks or infer eligibility from a successful Save alone.
async function prepareEligibleReview(
  page: Page,
  apiOrigin: string,
): Promise<KnowledgeMapItem> {
  let attempted = 0;
  for (const band of REVIEW_SEED_BANDS) {
    const map = await readStagingState<KnowledgeMap>(
      page,
      apiOrigin,
      `/api/v1/knowledge-map?band=${band}`,
    );
    expect(
      map.currentBand.number === band,
      "The candidate range must match the requested public catalog band.",
    ).toBe(true);
    const candidates = eligibleReviewSeeds(
      map.currentBand.items,
      MAX_REVIEW_SEED_CANDIDATES - attempted,
    );
    for (const candidate of candidates) {
      attempted++;
      const { word } = await readStagingState<WordDetailResponse>(
        page,
        apiOrigin,
        `/api/v1/canonical-words/${encodeURIComponent(candidate.wordSlug)}`,
      );
      const meaning = word.meanings.find(
        (item) => item.id === candidate.meaningId,
      );
      if (
        word.id !== candidate.wordId ||
        word.text !== candidate.wordText ||
        word.slug !== candidate.wordSlug ||
        meaning?.partOfSpeech !== candidate.partOfSpeech ||
        meaning.shortDefinition !== candidate.shortDefinition ||
        meaning.saved ||
        meaning.selfReportedKnown
      )
        continue;
      await page.goto(`/vocabulary/${encodeURIComponent(word.slug)}`);
      const card = page
        .getByRole("main")
        .locator(`li[id="meaning-${candidate.meaningId}"]`);
      const save = card.getByRole("button", {
        name: `Save ${word.text}: ${meaning.shortDefinition}`,
        exact: true,
      });
      await expect(save).toBeVisible();
      const saved = page.waitForResponse(
        (response) =>
          response.request().method() === "POST" &&
          new URL(response.url()).origin === apiOrigin &&
          new URL(response.url()).pathname === "/api/v1/user-words" &&
          response.request().postDataJSON()?.meaningId === candidate.meaningId,
      );
      await save.click();
      expect(
        (await saved).ok(),
        "The real UI must acknowledge saving this exact synthetic review meaning.",
      ).toBe(true);
      await expect(
        card.getByRole("button", {
          name: `Remove ${word.text} from saved words`,
          exact: true,
        }),
      ).toHaveAttribute("aria-pressed", "true");
      const confirmed = await readStagingState<KnowledgeMap>(
        page,
        apiOrigin,
        `/api/v1/knowledge-map?band=${band}`,
      );
      const current = confirmed.currentBand.items.find(
        (item) => item.meaningId === candidate.meaningId,
      );
      if (current?.saved && current.due && !current.selfReportedKnown)
        return current;
      // A previously removed word may still have today's immutable review audit.
      // Keep it saved and try another candidate; restoration must not bypass it.
    }
    if (attempted >= MAX_REVIEW_SEED_CANDIDATES) break;
  }
  throw new Error(
    "No fresh eligible review meaning was confirmed within the bounded synthetic catalog candidates. Preserve today's audit and inspect the synthetic account's remaining unsaved words; do not reset its history.",
  );
}

// Only confirmed correct answers newly credited today count toward the gate.
// Read the exact public meaning before grading; never guess and retry today.
async function reviewOneCard(page: Page, apiOrigin: string): Promise<boolean> {
  for (let attempt = 0; attempt < 40; attempt++) {
    const resume = page.getByRole("button", {
      name: "Continue",
      exact: true,
      disabled: false,
    });
    await expect(
      resume
        .or(getEnabledReviewPromptControl(page))
        .or(getTerminalReviewHeading(page))
        .or(page.getByRole("button", { name: "Start review", exact: true }))
        .first(),
    ).toBeVisible({ timeout: 120_000 });
    if (await resume.isVisible()) {
      // An interrupted earlier run may already have an acknowledged answer.
      // Advance it without counting its pre-existing credit toward this run.
      await resume.click();
      continue;
    }
    const readiness = await waitForReviewReadiness(page, 120_000);
    if (readiness === "terminal") {
      const next = page.getByRole("button", {
        name: "Keep reviewing",
        exact: true,
        disabled: false,
      });
      const caughtUp = page.getByRole("link", {
        name: "Discover words",
        exact: true,
      });
      const error = page.getByRole("main").getByRole("alert");
      // The completion heading arrives before the authoritative remaining-word
      // read settles. Neither a stale nor disabled Next button proves readiness.
      await expect(next.or(caughtUp).or(error).first()).toBeVisible({
        timeout: 120_000,
      });
      if (await error.isVisible())
        throw new Error(
          "The completed lesson could not confirm its next action.",
        );
      if (await next.isVisible()) {
        await next.click();
        continue;
      }
      return false;
    }
    const word = await page
      .getByRole("main")
      .getByRole("heading", { level: 1 })
      .innerText();
    const current = await readStagingState<{
      session: DailyLesson | null;
      dueCount: number;
    }>(page, apiOrigin, "/api/v1/daily-lessons/current");
    const step = current.session?.currentStep;
    if (
      !step ||
      step.wordText !== word ||
      current.session?.id !== new URL(page.url()).searchParams.get("lesson")
    )
      throw new Error(
        "The visible review question does not match the authoritative saved lesson. Refuse to guess an answer.",
      );
    const canonical = await readStagingState<WordDetailResponse>(
      page,
      apiOrigin,
      `/api/v1/canonical-words/${encodeURIComponent(step.wordSlug)}`,
    );
    const answer = reviewedMeaningAnswer(step, canonical.word);
    if (!answer) {
      // A stale frozen definition cannot be guessed from an edited live word.
      // Skip without grading and continue the bounded search for a safe card.
      await page
        .getByRole("button", { name: "Skip for now", exact: true })
        .click();
      continue;
    }
    const option = page.getByRole("radio", {
      name: answer.definition,
      exact: true,
    });
    await expect(option).toHaveValue(answer.choiceId);
    await option.check();
    const graded = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        /\/api\/v1\/daily-lessons\/[^/]+\/actions$/.test(
          new URL(response.url()).pathname,
        ),
    );
    await page
      .getByRole("button", { name: "Check answer", exact: true })
      .click();
    const response = await graded;
    expect(response.ok()).toBe(true);
    const lesson = (await response.json()) as DailyLesson;
    expect(lesson.feedback).toBeTruthy();
    const feedback = lesson.feedback!;
    if (!feedback.skipped) {
      expect(
        feedback.meaningId === answer.meaningId,
        "Grading must confirm the exact pre-read meaning.",
      ).toBe(true);
      expect(
        feedback.correct,
        "The exact canonical meaning must be answered correctly without a retry.",
      ).toBe(true);
      expect(
        feedback.answer === answer.definition,
        "The confirmed answer must match the meaning read before grading.",
      ).toBe(true);
    }
    await expect(
      page
        .getByRole("main")
        .getByRole("status")
        .filter({ hasText: feedback.answer }),
    ).toBeVisible();
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await waitForReviewReadiness(page, 120_000);
    if (feedback.skipped)
      expect(
        feedback.credited,
        "Skipped cards must never satisfy the real-review gate.",
      ).toBe(false);
    if (feedback.correct && !feedback.skipped && feedback.credited) return true;
  }
  throw new Error(
    "No new correct daily review credit after 40 bounded lesson transitions. Do not guess answers, retry reviewed words today or count attempts as learned words.",
  );
}

test.describe("Core loop against real staging (VOC-050-T02)", () => {
  test("authenticated journey: home -> discover -> save -> review -> sentence feedback -> progress -> settings -> logout -> auth gate", async ({
    page,
    context,
  }, testInfo) => {
    const sessionCookie = requiredEnv(SESSION_COOKIE_ENV);
    const csrfToken = requiredEnv(CSRF_COOKIE_ENV);
    const baseURL = testInfo.project.use.baseURL;
    if (!baseURL) {
      throw new Error(
        "Expected playwright.staging.config.ts to configure use.baseURL so cookies can be scoped to the deployed origin.",
      );
    }
    stagingOrigin(baseURL, "staging.vocanova.site");
    const apiOrigin = stagingOrigin(
      process.env.STAGING_API_BASE_URL ?? "https://api-staging.vocanova.site",
      "api-staging.vocanova.site",
    );

    await test.step("1. authenticate with the minted synthetic session", async () => {
      const domain = cookieDomainFor(baseURL);
      await context.addCookies([
        {
          name: SESSION_COOKIE_NAME,
          value: sessionCookie,
          domain,
          path: "/",
          secure: true,
          sameSite: "Lax",
        },
        {
          name: CSRF_COOKIE_NAME,
          value: csrfToken,
          domain,
          path: "/",
          secure: true,
          sameSite: "Lax",
        },
      ]);
      await page.goto("/home");
      await completeOnboardingIfRedirected(page);
      await expect(page).toHaveURL(/\/home(\?|$)/);
      await expect(
        page.getByRole("region", {
          name: "Review words",
          exact: true,
        }),
      ).toBeVisible();
    });

    const { reviewed: reviewedBefore, target: reviewTarget } =
      await test.step("2. read the daily-mission baseline", async () =>
        readReviewedTodayProgress(page));

    await test.step("2a. explore Daily Conversation without learning mutations", async () => {
      await verifyContextPracticeJourney(page);
    });

    await test.step("2b. assess a real map word without review credit and retain it after reload", async () => {
      const expectedEmail =
        process.env.VOCANOVA_SYNTHETIC_SMOKE_TEST_EMAIL ??
        "smoke-test-bot@synthetic.vocanova.invalid";
      expect(
        expectedEmail.endsWith("@synthetic.vocanova.invalid"),
        "Only the designated synthetic identity may assess map words in this gate.",
      ).toBe(true);
      const me = await readStagingState<CurrentUser>(
        page,
        apiOrigin,
        "/api/v1/me",
      );
      expect(
        me.email === expectedEmail,
        "Refuse assessment outside the designated synthetic account.",
      ).toBe(true);
      await verifyRealKnowledgeMap(page, apiOrigin);
      expect(await readReviewedTodayCount(page)).toBe(reviewedBefore);
    });

    let openedWordPath = "";
    await test.step("3. open situation courses from Journey and discover a word", async () => {
      const journey = page
        .getByRole("navigation", { name: "Primary", exact: true })
        .getByRole("link", { name: "Journey", exact: true });
      await expect(journey).toHaveAttribute("href", "/discover");
      await journey.click();
      await expect(page).toHaveURL(/\/discover$/);
      await expect(
        page.getByRole("heading", { name: "Journey", level: 1, exact: true }),
      ).toBeVisible();
      const courses = page
        .getByRole("navigation", { name: "More ways to learn", exact: true })
        .getByRole("link", { name: "Situation courses", exact: true });
      await expect(courses).toHaveAttribute("href", "/courses");
      await courses.click();
      await expect(page).toHaveURL(/\/courses$/);
      await expect(
        page.getByRole("heading", { name: "Situation courses", level: 1 }),
      ).toBeVisible();

      const situations = page
        .getByRole("region", { name: "Guided lessons", exact: true })
        .locator("details")
        .filter({
          has: page.getByRole("link", {
            name: /^Explore words in /,
            includeHidden: true,
          }),
        });
      expect(await situations.count()).toBeGreaterThan(0);
      const firstSituation = situations.first();
      if (
        !(await firstSituation.evaluate((element) =>
          element.hasAttribute("open"),
        ))
      ) {
        await firstSituation.locator("summary").focus();
        await page.keyboard.press("Enter");
        await expect(firstSituation).toHaveAttribute("open", "");
      }
      const explore = firstSituation.getByRole("link", {
        name: /^Explore words in /,
      });
      await expect(explore).toBeVisible();
      await expect(explore).toHaveAttribute("href", /^\/discover\/[^/]+$/);
      await explore.click();
      await expect(page).toHaveURL(/\/discover\/[^/]+(\?|$)/);
      const allCourses = page.getByRole("link", {
        name: "All courses",
        exact: true,
      });
      await expect(allCourses).toBeVisible();
      await expect(allCourses).toHaveAttribute("href", "/courses");

      const wordLink = await chooseSituationWordLink(page);
      await wordLink.click();
      await expect(page).toHaveURL(/\/discover\/[^/]+\/[^/]+(\?|$)/);
      openedWordPath = new URL(page.url()).pathname;
    });

    await test.step("4. save the word", async () => {
      const meaningCard = page
        .getByRole("main")
        .locator('li[id^="meaning-"]')
        .first();
      const saveButton = getFirstMeaningSaveControl(page);
      await expect(saveButton).toBeVisible();
      const meaningId = (await meaningCard.getAttribute("id"))?.slice(
        "meaning-".length,
      );
      expect(meaningId).toBeTruthy();

      const initiallySaved =
        (await saveButton.getAttribute("aria-pressed")) === "true";
      for (const action of reviewSeedActions(initiallySaved)) {
        const response = page.waitForResponse((result) => {
          const path = new URL(result.url()).pathname;
          return (
            action === "save" &&
            result.request().method() === "POST" &&
            path === "/api/v1/user-words" &&
            result.request().postDataJSON()?.meaningId === meaningId
          );
        });
        await saveButton.click();
        expect((await response).ok()).toBe(true);
        await expect(saveButton).toHaveAttribute("aria-pressed", "true");
      }

      await expect(saveButton).toHaveAttribute("aria-pressed", "true");

      // Saving exposes optional writing on this exact meaning without an
      // extra reload. Open its native disclosure as a learner would; a hidden
      // editor is not evidence that the writing entry point is usable.
      const tools = meaningCard
        .locator("summary")
        .filter({ hasText: "Personal tools" });
      if ((await tools.locator("..").getAttribute("open")) === null)
        await tools.click();
      const wordPractice = meaningCard.locator("summary").filter({
        hasText: "Optional writing",
      });
      await expect(wordPractice).toBeVisible();
      const disclosure = wordPractice.locator("..");
      if ((await disclosure.getAttribute("open")) === null) {
        await wordPractice.focus();
        await page.keyboard.press("Enter");
      }
      await expect(disclosure).toHaveAttribute("open", "");
      await expect(
        meaningCard.getByRole("textbox", {
          name: /^Write a sentence using /,
        }),
      ).toBeVisible();
    });

    await test.step("4b. confirm a fresh meaning is truly due without resetting review history", async () => {
      await prepareEligibleReview(page, apiOrigin);
    });

    const reviewedCards =
      await test.step("5. work the review queue", async () => {
        await page.goto("/reviews");
        let reviewed = 0;
        while (reviewed < MAX_REVIEW_CARDS) {
          const didReview = await reviewOneCard(page, apiOrigin);
          if (!didReview) {
            // reviewOneCard independently found a terminal review state - trust
            // it over this loop's own now-stale check above.
            break;
          }
          reviewed++;
          // The submission either advances to the next card or empties
          // the queue; both are settled states, so wait for one of them
          // instead of a fixed delay.
          await waitForReviewReadiness(page);
        }

        // VOC-074-T02: step 7 must not pass vacuously when the queue was empty or
        // caught up (reviewedAfter >= reviewedBefore + 0). Fail here, not in step 7.
        expect(
          reviewed,
          "step 5 must review at least one card for the staging core-loop gate (VOC-074-T02)",
        ).toBeGreaterThanOrEqual(1);

        const step5Summary = `reviewedCards=${reviewed}`;
        testInfo.annotations.push({
          type: "step-5-reviewed-cards",
          description: step5Summary,
        });
        console.log(`[staging core-loop] step 5 ${step5Summary}`);

        return reviewed;
      });

    await test.step("6. optional sentence feedback on the saved word", async () => {
      // Recall no longer embeds a compulsory editor. Revisit the saved word
      // deliberately; this gate still exercises real sentence feedback.
      expect(openedWordPath).toMatch(/^\/discover\/[^/]+\/[^/]+$/);
      await page.goto(`${openedWordPath}#sentence-practice`);
      const feedbackHeading = page
        .getByRole("heading", { name: /^Practice with / })
        .first();
      await expect(feedbackHeading).toBeVisible();
      const feedback = page.getByRole("region", {
        name: await feedbackHeading.innerText(),
        exact: true,
      });
      const sentenceInput = feedback.getByRole("textbox", {
        name: /^Write a sentence using /,
      });
      const targetWord = (
        (await sentenceInput.getAttribute("placeholder")) ?? ""
      ).replace(/^Type a sentence using "(.*)"\.\.\.$/, "$1");
      await sentenceInput.fill(
        `Today I practise the word ${targetWord} in a full sentence.`,
      );

      const feedbackResponse = page.waitForResponse(
        (response) =>
          response.url().includes("/api/v1/learner-sentences") &&
          response.request().method() === "POST",
      );
      await feedback
        .getByRole("button", { name: "Check my sentence", exact: true })
        .click();
      expect((await feedbackResponse).status()).toBe(200);

      // The real evaluator is not deterministic, so the assertion is
      // that the widget settled on a rendered outcome - a verdict or
      // an explicit error - not which verdict it was.
      const verdict = feedback.getByRole("status", {
        name: /^Feedback result: /,
      });
      await expect(
        verdict.or(feedback.getByRole("alert")).first(),
      ).toBeVisible();
    });

    await test.step("7. progress reflects the completed reviews", async () => {
      const reviewedAfter = await readReviewedTodayCountAfterReviews(
        page,
        testInfo,
        reviewedBefore,
        reviewedCards,
        reviewTarget,
      );
      expect(reviewedAfter).toBeGreaterThanOrEqual(
        Math.min(reviewedBefore + reviewedCards, reviewTarget),
      );
    });

    await test.step("8. change a setting", async () => {
      await page.goto("/settings");
      await expect(
        page.getByRole("heading", { name: "Settings", level: 1 }),
      ).toBeVisible();
      const displayNameInput = page.getByRole("textbox", {
        name: "Display name",
      });
      const displayName = `Synthetic smoke-test ${Date.now()}`;
      await displayNameInput.fill(displayName);
      await page.getByRole("button", { name: "Save settings" }).click();
      await expect(
        page.getByText("Your settings have been saved."),
      ).toBeVisible();
      await expect(displayNameInput).toHaveValue(displayName);
    });

    await test.step("9. log out", async () => {
      // A real evaluator may leave a recoverable, unsent draft. The learner
      // must explicitly confirm discarding it before logout can proceed.
      page.once("dialog", async (dialog) => {
        expect(dialog.type()).toBe("confirm");
        expect(dialog.message()).toContain("unsent practice draft");
        await dialog.accept();
      });
      await page.getByRole("button", { name: "Log out" }).click();
      await expect(page).toHaveURL(/\/login(\?|$)/);
      await expect(
        page.getByRole("heading", { name: "Sign in to Vocanova" }),
      ).toBeVisible();
    });

    await test.step("10. the auth gate rejects the now-invalid session", async () => {
      // No e2e_unauthenticated override cookie: logout revoked the
      // session server-side, so this is the real auth gate reacting
      // to a real 401 from /api/v1/me.
      await page.goto("/home");
      await expect(page).toHaveURL(/\/login\?returnTo=%2Fhome(\b|$)/);
    });
  });
});
