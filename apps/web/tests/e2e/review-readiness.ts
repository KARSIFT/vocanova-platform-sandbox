import { expect, type Locator, type Page } from "@playwright/test";

export type ReviewReadiness = "prompt" | "terminal";

export function getTerminalReviewHeading(page: Page): Locator {
  return page
    .getByRole("heading", { name: "You're all caught up", level: 2 })
    .or(page.getByRole("heading", { name: "Review complete", level: 2 }))
    .or(
      page.getByRole("heading", {
        name: "Today's review target is complete",
        level: 2,
      }),
    );
}

export function getEnabledReviewPromptControl(page: Page): Locator {
  const showAnswerButton = page.getByRole("button", {
    name: "Show answer",
    disabled: false,
  });
  const enabledMultipleChoiceOption = page
    .getByRole("group", { name: /^Choose the meaning for / })
    .getByRole("button", { disabled: false })
    .first();

  return showAnswerButton.or(enabledMultipleChoiceOption).first();
}

export async function waitForReviewReadiness(
  page: Page,
  timeout?: number,
): Promise<ReviewReadiness> {
  const terminalReviewHeading = getTerminalReviewHeading(page);
  const readiness = getEnabledReviewPromptControl(page)
    .or(terminalReviewHeading)
    .first();
  if (timeout === undefined) {
    await expect(readiness).toBeVisible();
  } else {
    await expect(readiness).toBeVisible({ timeout });
  }

  return (await terminalReviewHeading.isVisible()) ? "terminal" : "prompt";
}
