import { expect, type Locator, type Page } from "@playwright/test";

// Prefer an unsaved meaning; when every meaning is saved, the caller may
// deliberately remove and restore one to prepare a scheduled review.
export async function chooseSituationWordLink(page: Page): Promise<Locator> {
  const words = page.getByRole("main").getByRole("list", {
    name: "Words in this situation",
    exact: true,
  });
  await expect(words).toHaveCount(1);
  await expect(words).toBeVisible();
  const wordItems = words
    .getByRole("listitem")
    .filter({ has: page.getByRole("link") });
  const count = await wordItems.count();
  expect(count).toBeGreaterThan(0);
  for (let index = 0; index < count; index++) {
    const item = wordItems.nth(index);
    if ((await item.getByText(/^(?:✓\s*)?Saved$/).count()) === 0) {
      return item.getByRole("link").first();
    }
  }
  return wordItems.first().getByRole("link").first();
}

// Pronunciation speed and knowledge controls also use aria-pressed. Scope to
// an authored meaning card and its explicit saved-vocabulary action instead.
export function getFirstMeaningSaveControl(page: Page): Locator {
  return page
    .getByRole("main")
    .locator('li[id^="meaning-"]')
    .first()
    .getByRole("button", {
      name: /^(?:Save .+: |Remove .+ from saved words$)/,
    });
}
