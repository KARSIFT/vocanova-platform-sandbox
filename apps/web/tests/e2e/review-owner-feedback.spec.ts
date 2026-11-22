import { createHash, randomUUID } from "node:crypto";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { expect, test, type Locator, type Page } from "@playwright/test";
import type { DailyLesson, MeaningImage } from "@vocanova/api-client";
import { formatViolations, scanForAxeViolations } from "./axe-helper";
import { pressTab } from "./keyboard-navigation";

const attribution = { title: "Word sources", sourceUrl: "https://wordnet.princeton.edu/", licenseUrl: "https://wordnet.princeton.edu/license-and-commercial-use", wordnetLicense: "WordNet" };
const revision = "a".repeat(64);
const meaningId = "329c4ec9-6562-568e-9e71-9134da2f0d8d";
const goal = { localDate: "2026-11-15", timeZone: "America/New_York", target: 15, reviewed: 15, remaining: 0, complete: true };
const image = (index: number): MeaningImage => {
  const id = `00000000-0000-4000-8000-00000000000${index}`;
  return { meaningId: id, meaningRevision: revision, imageUrl: `https://images.vocanova.site/meanings/${id}/${revision}/${revision}.webp`, alt: `Scene ${index}`, width: 960, height: 640, sha256: revision };
};
function lesson(overrides: Partial<DailyLesson> = {}): DailyLesson {
  return { id: "owner-review", attribution, status: "in_progress", revision: 1, completedWords: 0, totalWords: 15, firstAnswersCorrect: 0, questionsAnswered: 0, creditedToday: 0,
    goal: { ...goal, reviewed: 0, remaining: 15, complete: false },
    currentStep: { id: "owner-step", kind: "meaning_choice", wordText: "menu", wordSlug: "menu", partOfSpeech: "noun", pronunciationMeaningId: meaningId, pronunciationRevision: revision, choices: ["a list of food", "a place to sleep", "a time to leave", "a way to travel"].map((text, i) => ({ id: `choice-${i}`, text })) },
    feedback: null, canContinue: false, createdAt: "2026-11-15T12:00:00Z", updatedAt: "2026-11-15T12:00:00Z", ...overrides };
}
async function setup(page: Page) {
  const baseURL = test.info().project.use.baseURL!;
  await page.context().addCookies([
    { name: "vocanova_session", value: `owner-${randomUUID()}`, url: baseURL },
    { name: "vocanova_csrf", value: "owner-feedback", url: baseURL },
    { name: "e2e_review_fixture_count", value: "20", url: baseURL },
  ]);
}
async function open(page: Page, value: DailyLesson) {
  await setup(page);
  await page.route("**/api/v1/daily-lessons", (route) => route.fulfill({ json: value }));
  await page.goto("/review");
  await page.getByRole("button", { name: "Start review", exact: true }).click();
}
async function completionRead(page: Page, value: DailyLesson, dueCount = 8) {
  await page.route("**/api/v1/daily-lessons/current", (route) => route.fulfill({ json: { session: value, dueCount, goal: value.goal } }));
}

test("wrong answers complete the authoritative goal and extra review stays optional", async ({ page }) => {
  const completed = lesson({ status: "completed", goal, completedWords: 15, questionsAnswered: 15, firstAnswersCorrect: 0, creditedToday: 0, currentStep: null });
  await completionRead(page, completed);
  await open(page, completed);
  await expect(page.getByRole("heading", { name: "Today’s review complete" })).toBeVisible();
  await expect(page.getByText("15 of 15 words reviewed today", { exact: true })).toBeVisible();
  await expect(page.getByText("0 correct on the first try.", { exact: true })).toBeVisible();
  await expect(page.getByRole("link", { name: "Discover words", exact: true })).toBeVisible();
  const extra = page.getByRole("button", { name: "Review more words", exact: true });
  await expect(extra).toBeVisible();
  const request = page.waitForRequest((request) => request.method() === "POST" && new URL(request.url()).searchParams.get("additional") === "true");
  await extra.click();
  await request;
});
test("completed session truthiness does not promise another required review", async ({ page }) => {
  const completed = lesson({ status: "completed", goal: { ...goal, reviewed: 4, remaining: 11, complete: false }, totalWords: 4, completedWords: 4, currentStep: null });
  await completionRead(page, completed, 0);
  await open(page, completed);
  await expect(page.getByRole("heading", { name: "All available reviews finished" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Keep reviewing", exact: true })).toHaveCount(0);
  await expect(page.getByText("4 of 15 words reviewed today", { exact: true })).toBeVisible();
});
test("gap question conceals the missing answer and pronunciation until checked", async ({ page }) => {
  const value = lesson();
  value.currentStep = { ...value.currentStep!, kind: "gap_choice", context: "Please bring the ____ so we can choose our food.", choices: ["menu", "bed", "train", "clock"].map((text, i) => ({ id: `choice-${i}`, text })) };
  await open(page, value);
  await expect(page.getByRole("heading", { name: "Complete the sentence", exact: true })).toBeVisible();
  await expect(page.getByText(value.currentStep.context!, { exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: /Listen|Automatic audio/ })).toHaveCount(0);
  await expect(page.getByRole("heading", { name: "menu", exact: true })).toHaveCount(0);
  await expect(page.getByRole("radio")).toHaveCount(4);
});
test("four picture choices can switch together to equivalent text", async ({ page }) => {
  const value = lesson();
  value.currentStep = { ...value.currentStep!, kind: "picture_choice", choices: value.currentStep!.choices.map((option, i) => ({ ...option, image: image(i + 1) })) };
  await page.route("https://images.vocanova.site/**", (route) => route.abort());
  await open(page, value);
  await expect(page.getByRole("button", { name: "Use text options", exact: true })).toBeVisible();
  await page.getByRole("button", { name: "Use text options", exact: true }).click();
  await expect(page.getByRole("radio", { name: "a list of food", exact: true })).toBeVisible();
  await expect(page.getByRole("radio")).toHaveCount(4);
  await expect(page.getByRole("img")).toHaveCount(0);
});
test("post-answer teaching uses canonical meanings and optional failure cannot block Continue", async ({ page }) => {
  const value = lesson();
  let canonicalReads = 0;
  await page.route("**/api/v1/canonical-words/menu", (route) => {
    canonicalReads++;
    return route.fulfill({ json: { word: { id: "menu", text: "menu", slug: "menu", wordType: "word", meanings: [{ id: meaningId, meaningRevision: revision, partOfSpeech: "noun", shortDefinition: "a list of food", saved: true, selfReportedKnown: false, examples: [{ id: "menu-food", exampleText: "Can I see the lunch menu?" }], usageNotes: [] }, { id: "other-real", partOfSpeech: "noun", shortDefinition: "a list of computer commands", saved: false, selfReportedKnown: false, examples: [{ id: "menu-software", exampleText: "Open the settings menu." }], usageNotes: [] }] } } });
  });
  await page.route("**/api/v1/meaning-images**", (route) => route.abort());
  await page.route("**/api/v1/daily-lessons/*/actions", (route) => route.fulfill({ json: { ...value, revision: 2, feedback: { stepId: value.currentStep!.id, correct: false, answer: "a list of food", explanation: "", meaningId, wordText: "menu", credited: false, skipped: false, correctChoiceId: "choice-0", selectedChoiceId: "choice-1" }, canContinue: true } }));
  await open(page, value);
  expect(canonicalReads).toBe(0);
  await expect(page.getByText("Can I see the lunch menu?", { exact: true })).toHaveCount(0);
  await page.getByRole("radio").nth(1).check();
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  await expect(page.getByRole("status").filter({ hasText: "Not quite" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Continue", exact: true })).toBeEnabled();
  await expect(page.getByText("Can I see the lunch menu?", { exact: true })).toBeVisible();
  await expect(page.getByText("Open the settings menu.", { exact: true })).toHaveCount(0);
  await expect(page.getByText("Other meanings of menu", { exact: true })).toBeVisible();
  await page.getByText("Other meanings of menu", { exact: true }).click();
  await expect(page.getByText("a list of computer commands", { exact: true })).toBeVisible();
  await expect(page.getByText("Word sources", { exact: true })).toHaveCount(0);
});

for (const optional of ["single", "failure", "revised"] as const) {
  test(`post-answer ${optional} canonical read leaves the checked answer and Continue available`, async ({ page }) => {
    const value = lesson();
    await page.route("**/api/v1/canonical-words/menu", (route) => optional === "failure"
      ? route.abort()
      : route.fulfill({ json: { word: { id: "menu", text: "menu", slug: "menu", wordType: "word", meanings: [{ id: meaningId, meaningRevision: optional === "revised" ? "b".repeat(64) : revision, partOfSpeech: "noun", shortDefinition: "a list of food", saved: true, selfReportedKnown: false, examples: optional === "revised" ? [{ id: "later", exampleText: "A sentence from the revised meaning." }] : [], usageNotes: [] }] } } }));
    await page.route("**/api/v1/meaning-images**", (route) => route.abort());
    await page.route("**/api/v1/daily-lessons/*/actions", (route) => route.fulfill({ json: { ...value, revision: 2, feedback: { stepId: value.currentStep!.id, correct: true, answer: "a list of food", explanation: "", meaningId, wordText: "menu", credited: true, skipped: false, correctChoiceId: "choice-0", selectedChoiceId: "choice-0" }, canContinue: true } }));
    await open(page, value);
    await page.getByRole("radio").first().check();
    await page.getByRole("button", { name: "Check answer", exact: true }).click();
    await expect(page.getByRole("status").filter({ hasText: /^Correct/ })).toContainText("a list of food");
    await expect(page.getByText(optional === "failure" ? "Other meanings could not load. You can continue." : "This is the only meaning available for menu.", { exact: true })).toBeVisible();
    await expect(page.getByRole("button", { name: "Continue", exact: true })).toBeEnabled();
    await expect(page.getByText("A sentence from the revised meaning.", { exact: true })).toHaveCount(0);
    await expect(page.getByRole("main").getByRole("alert")).toHaveCount(0);
    await expect(page.getByRole("status")).toHaveCount(1);
  });
}

test("late optional teaching cannot bleed into the next question", async ({ page }) => {
  const value = lesson();
  let release!: () => void;
  const gate = new Promise<void>((resolve) => { release = resolve; });
  await page.route("**/api/v1/canonical-words/menu", async (route) => {
    await gate;
    await route.fulfill({ json: { word: { id: "menu", text: "menu", slug: "menu", wordType: "word", meanings: [{ id: "late-meaning", shortDefinition: "A late meaning that belongs only to menu", partOfSpeech: "noun", saved: false, selfReportedKnown: false, examples: [], usageNotes: [] }] } } });
  });
  await page.route("**/api/v1/meaning-images**", (route) => route.abort());
  await page.route("**/api/v1/daily-lessons/*/actions", (route) => {
    const action = route.request().postDataJSON() as { action: string };
    return route.fulfill({ json: action.action === "answer"
      ? { ...value, revision: 2, feedback: { stepId: value.currentStep!.id, correct: true, answer: "a list of food", explanation: "", meaningId, wordText: "menu", credited: true, skipped: false, correctChoiceId: "choice-0", selectedChoiceId: "choice-0" }, canContinue: true }
      : { ...value, revision: 3, completedWords: 1, currentStep: { ...value.currentStep!, id: "next-step", wordText: "restaurant", wordSlug: "restaurant" } } });
  });
  try {
    await open(page, value);
    await page.getByRole("radio").first().check();
    const read = page.waitForRequest("**/api/v1/canonical-words/menu");
    await page.getByRole("button", { name: "Check answer", exact: true }).click();
    await read;
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await expect(page.getByRole("heading", { name: "restaurant", exact: true })).toBeVisible();
    release();
    await expect(page.getByRole("region", { name: "Learn more about menu", exact: true })).toHaveCount(0);
    await expect(page.getByText("A late meaning that belongs only to menu", { exact: true })).toHaveCount(0);
    await expect(page.getByRole("button", { name: "Check answer", exact: true })).toBeDisabled();
  } finally {
    release();
    await page.unrouteAll({ behavior: "wait" });
  }
});

test("checked teaching restores after a real synthetic session reload", async ({ page }) => {
  await setup(page);
  let canonicalReads = 0;
  await page.route("**/api/v1/canonical-words/fixture-word-1", (route) => {
    canonicalReads++;
    return route.fulfill({ json: { word: { id: "fixture-word-1", text: "Review word 1", slug: "fixture-word-1", wordType: "word", meanings: [{ id: "fixture-meaning-1", partOfSpeech: "noun", shortDefinition: "definition for review word 1", saved: true, selfReportedKnown: false, examples: [], usageNotes: [] }] } } });
  });
  await page.goto("/review");
  await page.getByRole("button", { name: "Start review", exact: true }).click();
  await page.getByRole("radio", { name: "definition for review word 1", exact: true }).check();
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  await expect(page.getByRole("region", { name: "Learn more about Review word 1", exact: true })).toBeVisible();
  await page.reload();
  await expect(page.getByRole("status").filter({ hasText: /^Correct/ })).toBeVisible();
  await expect(page.getByRole("region", { name: "Learn more about Review word 1", exact: true })).toBeVisible();
  await expect.poll(() => canonicalReads).toBe(2);
  await expect(page.getByRole("button", { name: "Continue", exact: true })).toBeEnabled();
});

for (const theme of ["light", "dark"] as const) {
  test(`mixed review choices support keyboard and 200 percent text in ${theme}`, async ({ page }) => {
    const value = lesson();
    value.currentStep = { ...value.currentStep!, kind: "picture_choice", choices: value.currentStep!.choices.map((option, i) => ({ ...option, image: image(i + 1) })) };
    await page.context().addCookies([{ name: "vocanova_theme", value: theme, url: test.info().project.use.baseURL! }]);
    await page.route("https://images.vocanova.site/**", (route) => route.fulfill({ contentType: "image/svg+xml", body: '<svg xmlns="http://www.w3.org/2000/svg" width="960" height="640"><rect width="960" height="640" fill="#d0e6df"/></svg>' }));
    await open(page, value);
    await page.addStyleTag({ content: "html { font-size: 200% !important; }" });
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    await expect(page.getByRole("img")).toHaveCount(4);
    const useText = page.getByRole("button", { name: "Use text options", exact: true });
    await useText.focus();
    await useText.press("Enter");
    const option = page.getByRole("radio").first();
    await option.focus();
    await option.press("Space");
    await expect(option).toBeChecked();
    await expect(page.getByRole("button", { name: "Check answer", exact: true })).toBeEnabled();
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(page.viewportSize()!.width);
    await page.screenshot({ path: test.info().outputPath(`review-${theme}-200-percent.png`), fullPage: true });
  });
}


// Use the repository's reviewed artwork/meaning revisions for visual inspection.
// The HTTP delivery is synthetic; production eligibility remains a backend test.
const reviewedPictures = JSON.parse(readFileSync(new URL("../../src/lib/meaning-pictures.runtime.json", import.meta.url), "utf8")) as Record<string, { alt: string; meaningRevision: string }>;
const reviewedChoices = [
  [meaningId, "A list of food and drinks you can order."],
  ["5d21c02b-ef51-5545-bd8a-01f5814d3a67", "The statement showing how much you need to pay."],
  ["2534b539-a80f-5001-b9a4-27a5e0679539", "A small dish eaten before the main meal."],
  ["bd5eacb6-be32-5e7b-a8d7-d56db1908347", "Extra money given to thank someone for service."],
].map(([id, text], index) => {
  const path = fileURLToPath(new URL(`../../public/images/meanings/${id}.webp`, import.meta.url));
  const meta = reviewedPictures[id!]!;
  const sha256 = createHash("sha256").update(readFileSync(path)).digest("hex");
  return { id: `reviewed-${index}`, text: text!, path, image: { meaningId: id!, meaningRevision: meta.meaningRevision, sha256, alt: meta.alt, width: 960, height: 640, imageUrl: `https://images.vocanova.site/meanings/${id}/${meta.meaningRevision}/${sha256}.webp` } };
});
for (const theme of ["light", "dark"] as const) {
  test(`all three review modes and checked teaching have accessible visual layouts in ${theme}`, async ({ page }, testInfo) => {
    const kinds = ["meaning_choice", "picture_choice", "gap_choice"] as const;
    const sessions = kinds.map((kind, index) => {
      const base = lesson();
      return { ...base, revision: 1 + index * 2, completedWords: index, currentStep: {
        ...base.currentStep!, id: `visual-${index}`, kind,
        ...(kind === "gap_choice" ? { wordText: "", wordSlug: "", pronunciationMeaningId: undefined, pronunciationRevision: undefined, context: "Please bring the ____ so we can choose our food." } : { pronunciationRevision: reviewedPictures[meaningId]!.meaningRevision }),
        choices: kind === "gap_choice" ? ["menu", "bed", "train", "clock"].map((text, index) => ({ id: `reviewed-${index}`, text })) : reviewedChoices.map(({ id, text, image }) => ({ id, text, ...(kind === "picture_choice" ? { image } : {}) })),
      } } satisfies DailyLesson;
    });
    let index = 0;
    await page.context().addCookies([{ name: "vocanova_theme", value: theme, url: testInfo.project.use.baseURL! }]);
    await page.route("https://images.vocanova.site/**", (route) => {
      const choice = reviewedChoices.find((choice) => choice.image.imageUrl === route.request().url());
      return choice ? route.fulfill({ path: choice.path, contentType: "image/webp" }) : route.abort();
    });
    await page.route("**/api/v1/meaning-images**", (route) => route.fulfill({ json: { items: [reviewedChoices[0]!.image] } }));
    await page.route("**/api/v1/canonical-words/menu", (route) => route.fulfill({ json: { word: { id: "menu", text: "menu", slug: "menu", wordType: "word", meanings: [{ id: meaningId, meaningRevision: reviewedPictures[meaningId]!.meaningRevision, image: reviewedChoices[0]!.image, partOfSpeech: "noun", shortDefinition: reviewedChoices[0]!.text, saved: true, selfReportedKnown: false, examples: [{ id: "menu-food", exampleText: "Can I see the lunch menu?" }], usageNotes: [] }] } } }));
    await page.route("**/api/v1/daily-lessons/*/actions", (route) => {
      if (route.request().postDataJSON().action === "continue") return route.fulfill({ json: sessions[++index] });
      const value = sessions[index]!;
      return route.fulfill({ json: { ...value, revision: value.revision + 1, canContinue: true, feedback: { stepId: value.currentStep!.id, correct: true, answer: reviewedChoices[0]!.text, explanation: "", meaningId, meaningRevision: reviewedPictures[meaningId]!.meaningRevision, wordText: "menu", wordSlug: "menu", credited: true, skipped: false, correctChoiceId: "reviewed-0", selectedChoiceId: "reviewed-0", nextReviewAt: "2026-11-16T14:00:00Z", reviewStep: 1 } } });
    });
    await open(page, sessions[0]!);
    for (let mode = 0; mode < kinds.length; mode++) {
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      await expect(page.getByRole("radio")).toHaveCount(4);
      if (kinds[mode] === "picture_choice") {
        for (const picture of await page.getByRole("img").all()) await expect.poll(() => picture.evaluate((element) => (element as HTMLImageElement).naturalWidth)).toBeGreaterThan(0);
      }
      const promptScan = await scanForAxeViolations(page);
      expect(promptScan.criticalOrSerious, formatViolations(promptScan.criticalOrSerious).join("\n")).toEqual([]);
      await page.screenshot({ path: testInfo.outputPath(`${theme}-${kinds[mode]}-viewport.png`) });
      await page.screenshot({ path: testInfo.outputPath(`${theme}-${kinds[mode]}-full.png`), fullPage: true });
      const first = page.getByRole("radio").first();
      await first.focus();
      await first.press("Space");
      await page.getByRole("button", { name: "Check answer", exact: true }).click();
      const teaching = page.getByRole("region", { name: "Learn more about menu", exact: true });
      await expect(teaching.getByText("Can I see the lunch menu?", { exact: true })).toBeVisible();
      await expect(teaching.getByRole("img")).toBeVisible();
      await expect.poll(() => teaching.getByRole("img").evaluate((element) => (element as HTMLImageElement).naturalWidth)).toBeGreaterThan(0);
      await expect(page.getByRole("button", { name: "Continue", exact: true })).toBeEnabled();
      const focusedFeedback = page.getByRole("status").filter({ hasText: /^Correct/ });
      await expect(focusedFeedback).toBeFocused();
      const feedbackBounds = (await focusedFeedback.boundingBox())!;
      const headerBounds = (await page.getByRole("banner").boundingBox())!;
      const railBounds = (await page.getByRole("region", { name: "Session actions", exact: true }).boundingBox())!;
      expect(feedbackBounds.y).toBeGreaterThanOrEqual(headerBounds.y + headerBounds.height + 8);
      expect(feedbackBounds.y + feedbackBounds.height).toBeLessThanOrEqual(railBounds.y - 8);
      const feedbackScan = await scanForAxeViolations(page);
      expect(feedbackScan.criticalOrSerious, formatViolations(feedbackScan.criticalOrSerious).join("\n")).toEqual([]);
      await page.screenshot({ path: testInfo.outputPath(`${theme}-${kinds[mode]}-checked-viewport.png`) });
      await page.screenshot({ path: testInfo.outputPath(`${theme}-${kinds[mode]}-checked-full.png`), fullPage: true });
      if (kinds[mode] === "picture_choice") {
        // Use native full-navigation keys on every engine. Sequential focus
        // from a negative-tabindex status can begin at different anchors.
        const fallback = page.getByRole("button", { name: "Use text options", exact: true });
        const audio = page.getByRole("button", { name: "Automatic audio off", exact: true });
        const visited: string[] = [];
        let sawFallback = false;
        let sawAudio = false;
        const assertClearTarget = async (target: Locator) => {
          await expect(target).toBeFocused();
          const bounds = (await target.boundingBox())!;
          expect(bounds.width).toBeGreaterThanOrEqual(44);
          expect(bounds.height).toBeGreaterThanOrEqual(44);
          expect(bounds.y).toBeGreaterThanOrEqual(headerBounds.y + headerBounds.height + 8);
          expect(bounds.y + bounds.height).toBeLessThanOrEqual(railBounds.y - 8);
          expect(await target.evaluate((element) => {
            const rect = element.getBoundingClientRect();
            const hit = document.elementFromPoint(rect.x + rect.width / 2, rect.y + rect.height / 2);
            return !!hit && element.contains(hit);
          })).toBe(true);
        };
        for (let step = 0; step < 24; step++) {
          await pressTab(page, { backwards: true });
          visited.push(await page.evaluate(() => {
            const element = document.activeElement;
            return `${element?.tagName}: ${element?.getAttribute("aria-label") ?? element?.textContent?.trim().slice(0, 60)}`;
          }));
          if (await fallback.evaluate((element) => element === document.activeElement)) {
            await assertClearTarget(fallback);
            sawFallback = true;
          }
          if (sawFallback && await audio.evaluate((element) => element === document.activeElement)) {
            await assertClearTarget(audio);
            sawAudio = true;
            break;
          }
        }
        expect(sawFallback, `Reverse-Tab sequence: ${visited.join(" → ")}`).toBe(true);
        expect(sawAudio, `Reverse-Tab sequence: ${visited.join(" → ")}`).toBe(true);
        expect((await scanForAxeViolations(page)).criticalOrSerious).toEqual([]);
      }
      if (mode < kinds.length - 1) await page.getByRole("button", { name: "Continue", exact: true }).click();
    }
  });
}
