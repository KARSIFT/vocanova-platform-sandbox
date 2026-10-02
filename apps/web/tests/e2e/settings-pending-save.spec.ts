import { randomUUID } from "node:crypto";

import { expect, test, type Route } from "@playwright/test";

const NEWER_CHANGES_MESSAGE =
  "Your earlier changes were saved. You have newer changes to save.";

test.beforeEach(async ({ context, baseURL }) => {
  if (!baseURL) throw new Error("A test app URL is required");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
  ]);
});

for (const theme of ["light", "dark"] as const) {
  test(`preserves newer text, target and preference edits when an earlier save completes (${theme})`, async ({
    page,
  }, testInfo) => {
    await page.emulateMedia({ colorScheme: theme });
    const requests: Record<string, unknown>[] = [];
    let releaseFirstSave: (() => void) | undefined;
    const firstSavePending = new Promise<void>((resolve) => {
      releaseFirstSave = resolve;
    });
    await page.route("**/api/v1/settings", async (route) => {
      if (route.request().method() === "PATCH") {
        requests.push(route.request().postDataJSON());
        if (requests.length === 1) await firstSavePending;
      }
      await route.continue();
    });

    await page.goto("/settings");
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    const name = page.getByRole("textbox", { name: "Display name" });
    const reminders = page.getByRole("switch", {
      name: /^Daily review reminder/,
    });
    await expect(reminders).toHaveAttribute("aria-checked", "true");
    await name.fill("Earlier name");
    await page.getByRole("button", { name: "Save settings" }).click();
    await expect.poll(() => requests.length).toBe(1);
    try {
      await expect(
        page.getByRole("button", { name: "Saving..." }),
      ).toBeDisabled();
      await name.fill("Newer name");
      await page.getByRole("radio", { name: "10", exact: true }).press("Space");
      await expect(
        page.getByRole("radio", { name: "10", exact: true }),
      ).toBeChecked();
      await reminders.click();
    } finally {
      releaseFirstSave?.();
    }

    await expect(
      page.getByRole("button", { name: "Save settings" }),
    ).toBeEnabled();
    await expect(name).toHaveValue("Newer name");
    await expect(
      page.getByRole("radio", { name: "10", exact: true }),
    ).toBeChecked();
    await expect(reminders).toHaveAttribute("aria-checked", "false");
    await expect(page.getByRole("status")).toHaveText(NEWER_CHANGES_MESSAGE);
    await expect(page.getByText("Your settings have been saved.")).toHaveCount(
      0,
    );
    expect(requests).toEqual([{ displayName: "Earlier name" }]);
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth),
    ).toBeLessThanOrEqual(page.viewportSize()!.width);
    await page.getByRole("form", { name: "Practice settings" }).screenshot({
      path: testInfo.outputPath(`settings-newer-changes-${theme}.png`),
    });

    await page.getByRole("button", { name: "Save settings" }).click();
    await expect(
      page.getByText("Your settings have been saved."),
    ).toBeVisible();
    expect(requests).toEqual([
      { displayName: "Earlier name" },
      {
        displayName: "Newer name",
        dailyReviewTarget: 10,
        notificationsEnabled: false,
      },
    ]);
    await page.reload();
    await expect(name).toHaveValue("Newer name");
    await expect(
      page.getByRole("radio", { name: "10", exact: true }),
    ).toBeChecked();
    await expect(reminders).toHaveAttribute("aria-checked", "false");
  });
}

test("keeps a reversion made while saving and persists it against the confirmed baseline", async ({
  page,
}) => {
  let firstSave: Route | undefined;
  const requests: Record<string, unknown>[] = [];
  await page.route("**/api/v1/settings", async (route) => {
    if (route.request().method() === "PATCH") {
      requests.push(route.request().postDataJSON());
      if (requests.length === 1) {
        firstSave = route;
        return;
      }
    }
    await route.continue();
  });

  await page.goto("/settings");
  const name = page.getByRole("textbox", { name: "Display name" });
  const initialName = await name.inputValue();
  await name.fill("Temporary name");
  await page.getByRole("button", { name: "Save settings" }).click();
  await expect.poll(() => Boolean(firstSave)).toBe(true);
  await name.fill(initialName);
  await firstSave!.continue();
  await expect(
    page.getByRole("button", { name: "Save settings" }),
  ).toBeEnabled();
  await expect(name).toHaveValue(initialName);
  await expect(page.getByRole("status")).toHaveText(NEWER_CHANGES_MESSAGE);

  await page.getByRole("button", { name: "Save settings" }).click();
  await expect(page.getByText("Your settings have been saved.")).toBeVisible();
  expect(requests).toEqual([
    { displayName: "Temporary name" },
    { displayName: initialName },
  ]);
  await page.reload();
  await expect(name).toHaveValue(initialName);
});

test("retains newer edits after a pending save fails and retries all unsaved changes", async ({
  page,
}) => {
  let firstSave: Route | undefined;
  const requests: Record<string, unknown>[] = [];
  await page.route("**/api/v1/settings", async (route) => {
    if (route.request().method() === "PATCH") {
      requests.push(route.request().postDataJSON());
      if (requests.length === 1) {
        firstSave = route;
        return;
      }
    }
    await route.continue();
  });

  await page.goto("/settings");
  const name = page.getByRole("textbox", { name: "Display name" });
  await name.fill("Earlier name");
  await page.getByRole("button", { name: "Save settings" }).click();
  await expect.poll(() => Boolean(firstSave)).toBe(true);
  await name.fill("Name after failure");
  await page.getByRole("radio", { name: "5", exact: true }).press("Space");
  await expect(
    page.getByRole("radio", { name: "5", exact: true }),
  ).toBeChecked();
  await firstSave!.fulfill({
    status: 500,
    contentType: "application/json",
    body: JSON.stringify({ error: "temporary_failure" }),
  });
  await expect(
    page.getByRole("form", { name: "Practice settings" }).getByRole("alert"),
  ).toContainText("HTTP 500");
  await expect(name).toHaveValue("Name after failure");
  await expect(
    page.getByRole("radio", { name: "5", exact: true }),
  ).toBeChecked();
  await expect(page.getByText("Your settings have been saved.")).toHaveCount(0);

  await page.getByRole("button", { name: "Save settings" }).click();
  await expect(page.getByText("Your settings have been saved.")).toBeVisible();
  expect(requests).toEqual([
    { displayName: "Earlier name" },
    { displayName: "Name after failure", dailyReviewTarget: 5 },
  ]);
  await page.reload();
  await expect(name).toHaveValue("Name after failure");
  await expect(
    page.getByRole("radio", { name: "5", exact: true }),
  ).toBeChecked();
});

test("retries a reverted value after the server applies a save but its response is lost", async ({
  page,
}) => {
  const requests: Record<string, unknown>[] = [];
  let firstSaveApplied = false;
  let loseFirstResponse: (() => void) | undefined;
  const firstResponsePending = new Promise<void>((resolve) => {
    loseFirstResponse = resolve;
  });
  await page.route("**/api/v1/settings", async (route) => {
    if (route.request().method() === "PATCH") {
      requests.push(route.request().postDataJSON());
      if (requests.length === 1) {
        // The synthetic backend commits the write, but the browser never
        // receives confirmation. This is an ambiguous transport failure.
        const response = await route.fetch();
        expect(response.ok()).toBeTruthy();
        expect(await response.json()).toMatchObject({
          displayName: "Applied name",
        });
        firstSaveApplied = true;
        await firstResponsePending;
        await route.abort("failed");
        return;
      }
    }
    await route.continue();
  });

  await page.goto("/settings");
  const name = page.getByRole("textbox", { name: "Display name" });
  const initialName = await name.inputValue();
  await name.fill("Applied name");
  await page.getByRole("button", { name: "Save settings" }).click();
  await expect.poll(() => firstSaveApplied).toBe(true);
  try {
    await name.fill(initialName);
  } finally {
    loseFirstResponse?.();
  }
  await expect(
    page.getByRole("form", { name: "Practice settings" }).getByRole("alert"),
  ).toBeVisible();
  await expect(name).toHaveValue(initialName);

  await page.getByRole("button", { name: "Save settings" }).click();
  await expect.poll(() => requests.length).toBe(2);
  await expect(page.getByText("Your settings have been saved.")).toBeVisible();
  expect(requests).toEqual([
    { displayName: "Applied name" },
    { displayName: initialName },
  ]);
  // A confirmed response resolves the uncertainty; another unchanged save
  // must not send the attempted fields again.
  await page.getByRole("button", { name: "Save settings" }).click();
  expect(requests).toHaveLength(2);
  await page.reload();
  await expect(name).toHaveValue(initialName);
});
