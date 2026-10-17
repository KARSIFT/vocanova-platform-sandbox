import { expect, type Locator, type Page } from "@playwright/test";

export type ReviewReadiness = "prompt" | "terminal";

export function getTerminalReviewHeading(page: Page): Locator {
  return page
    .getByRole("heading", { name: "You’re ready for new words", level: 1 })
    .or(
      page.getByRole("heading", {
        name: /^(?:Review saved|Review set complete|You’ve finished your review)$/,
        level: 1,
      }),
    );
}

export function getEnabledReviewPromptControl(page: Page): Locator {
  return page.getByRole("radio", { disabled: false }).first();
}

export async function startReviewLesson(page: Page, timeout?: number) {
  const start = page.getByRole("button", { name: "Start review", exact: true });
  await expect(
    start
      .or(getEnabledReviewPromptControl(page))
      .or(getTerminalReviewHeading(page))
      .first(),
  ).toBeVisible(timeout === undefined ? {} : { timeout });
  if (await start.isVisible()) await start.click();
}

export async function answerReviewWord(page: Page, meaning: string) {
  await page.getByRole("radio", { name: meaning, exact: true }).check();
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  await expect(
    page.getByRole("status").filter({ hasText: /^Correct/ }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Continue", exact: true }).click();
}

export async function waitForReviewReadiness(
  page: Page,
  timeout?: number,
): Promise<ReviewReadiness> {
  await startReviewLesson(page, timeout);
  const terminal = getTerminalReviewHeading(page);
  await expect(
    getEnabledReviewPromptControl(page).or(terminal).first(),
  ).toBeVisible(timeout === undefined ? {} : { timeout });
  return (await terminal.isVisible()) ? "terminal" : "prompt";
}
