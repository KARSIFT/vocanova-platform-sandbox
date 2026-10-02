import { randomUUID } from "node:crypto";

import { expect, test } from "@playwright/test";
import type { TestInfo } from "@playwright/test";

function desktopOnly(testInfo: TestInfo) {
  test.skip(
    testInfo.project.name !== "home-desktop-1280",
    "Functional profile regressions run once at the representative desktop viewport.",
  );
}

async function authenticatedCookies(testInfo: TestInfo) {
  const baseURL = testInfo.project.use.baseURL;
  if (!baseURL) throw new Error("Expected a configured base URL.");
  return {
    baseURL,
    cookies: [
      {
        name: "vocanova_session",
        value: `profile-${randomUUID()}`,
        url: baseURL,
      },
      {
        name: "vocanova_csrf",
        value: `profile-csrf-${randomUUID()}`,
        url: baseURL,
      },
    ],
  };
}

test.describe("Profile and password security", () => {
  test("does not describe an unverified server email as verified", async ({
    page,
    context,
  }, testInfo) => {
    desktopOnly(testInfo);
    const { baseURL, cookies } = await authenticatedCookies(testInfo);
    await context.addCookies([
      ...cookies,
      { name: "e2e_email_verified", value: "false", url: baseURL },
    ]);
    await page.goto("/settings/profile");

    await expect(
      page.getByRole("heading", { name: "Email", level: 2 }),
    ).toBeVisible();
    await expect(
      page.getByText("This email has not been verified yet."),
    ).toBeVisible();
  });

  test("saves the display name and preserves it after reload", async ({
    page,
    context,
  }, testInfo) => {
    desktopOnly(testInfo);
    const { cookies } = await authenticatedCookies(testInfo);
    await context.addCookies(cookies);
    await page.goto("/settings/profile");

    const displayName = page.getByRole("textbox", { name: "Display name" });
    await displayName.fill("  Profile learner  ");
    await page.getByRole("button", { name: "Save profile" }).click();
    await expect(
      page
        .getByRole("status")
        .filter({ hasText: "Your profile has been saved." }),
    ).toBeVisible();
    await expect(displayName).toHaveValue("Profile learner");
    await page.reload();
    await expect(displayName).toHaveValue("Profile learner");
    await expect(
      page.getByText("core-loop-fixture@example.test"),
    ).toBeVisible();
  });

  test("routes an expired profile update back to sign-in without claiming success", async ({
    page,
    context,
  }, testInfo) => {
    desktopOnly(testInfo);
    const { baseURL, cookies } = await authenticatedCookies(testInfo);
    await context.addCookies(cookies);
    await page.goto("/settings/profile");
    await page
      .getByRole("textbox", { name: "Display name" })
      .fill("Needs auth");
    await context.addCookies([
      { name: "e2e_unauthenticated", value: "1", url: baseURL },
    ]);

    await page.getByRole("button", { name: "Save profile" }).click();
    await expect(page).toHaveURL(/\/login\?returnTo=%2Fsettings%2Fprofile/);
    await expect(
      page.getByText("Your session expired. Sign in again to continue."),
    ).toBeVisible();
  });

  test("offers a Google or magic-link account an email-confirmed add-password action", async ({
    page,
    context,
  }, testInfo) => {
    desktopOnly(testInfo);
    const { baseURL, cookies } = await authenticatedCookies(testInfo);
    await context.addCookies([
      ...cookies,
      { name: "e2e_has_password", value: "false", url: baseURL },
    ]);
    await page.goto("/settings/account");
    await expect(
      page.getByRole("heading", { name: "Password", level: 2 }),
    ).toBeVisible();
    await expect(page.getByText("No password is set.")).toBeVisible();
    await expect(page.getByLabel("Email address")).toHaveValue(
      "core-loop-fixture@example.test",
    );
    await page
      .getByRole("button", { name: "Email link to add a password" })
      .click();
    await expect(
      page.getByRole("status").filter({ hasText: /sent instructions/ }),
    ).toBeVisible();
  });
});

const PROFILE_NEWER_CHANGES_MESSAGE =
  "Your earlier profile changes were saved. You have newer changes to save.";

for (const theme of ["light", "dark"] as const) {
  test(`preserves a newer profile name when an earlier save completes (${theme})`, async ({
    page,
    context,
  }, testInfo) => {
    const { baseURL, cookies } = await authenticatedCookies(testInfo);
    await context.addCookies([
      ...cookies,
      { name: "vocanova_theme", value: theme, url: baseURL },
    ]);
    const requests: Record<string, unknown>[] = [];
    let earlierSaveApplied = false;
    let releaseSave!: () => void;
    const pendingSave = new Promise<void>((resolve) => {
      releaseSave = resolve;
    });
    await page.route("**/api/v1/settings", async (route) => {
      if (route.request().method() === "PATCH") {
        requests.push(route.request().postDataJSON());
        if (requests.length === 1) {
          const response = await route.fetch();
          expect(response.ok()).toBeTruthy();
          expect(await response.json()).toMatchObject({
            displayName: "Earlier name",
          });
          earlierSaveApplied = true;
          await pendingSave;
          await route.fulfill({ response });
          return;
        }
      }
      await route.continue();
    });
    await page.goto("/settings/profile");
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    const name = page.getByRole("textbox", { name: "Display name" });
    await name.fill("Earlier name");
    await page.getByRole("button", { name: "Save profile" }).click();
    await expect.poll(() => earlierSaveApplied).toBe(true);
    try {
      await name.fill("Later name");
      await expect
        .soft(page.getByRole("button", { name: "Saving..." }))
        .toBeDisabled();
    } finally {
      releaseSave();
    }
    await expect(
      page.getByRole("button", { name: "Save profile" }),
    ).toBeEnabled();
    await expect(page.getByRole("status")).toHaveText(
      PROFILE_NEWER_CHANGES_MESSAGE,
    );
    await expect(name).toHaveValue("Later name");
    await expect(page.getByText("Your profile has been saved.")).toHaveCount(0);
    expect(requests).toEqual([{ displayName: "Earlier name" }]);
    await name.locator("xpath=ancestor::form").screenshot({
      path: testInfo.outputPath(`profile-newer-changes-${theme}.png`),
    });
    await page.getByRole("button", { name: "Save profile" }).click();
    await expect(page.getByRole("status")).toHaveText(
      "Your profile has been saved.",
    );
    expect(requests).toEqual([
      { displayName: "Earlier name" },
      { displayName: "Later name" },
    ]);
    await page.reload();
    await expect(name).toHaveValue("Later name");
  });
}

test("blocks overlapping profile submissions while typing during a save", async ({
  page,
  context,
}, testInfo) => {
  const { cookies } = await authenticatedCookies(testInfo);
  await context.addCookies(cookies);
  const requests: Record<string, unknown>[] = [];
  let releaseSave!: () => void;
  const pendingSave = new Promise<void>((resolve) => {
    releaseSave = resolve;
  });
  await page.route("**/api/v1/settings", async (route) => {
    if (route.request().method() === "PATCH") {
      requests.push(route.request().postDataJSON());
      await pendingSave;
    }
    await route.continue();
  });
  await page.goto("/settings/profile");
  const name = page.getByRole("textbox", { name: "Display name" });
  await name.fill("Earlier name");
  await page.getByRole("button", { name: "Save profile" }).click();
  await expect.poll(() => requests.length).toBe(1);
  try {
    await name.fill("Later name");
    await name.locator("xpath=ancestor::form").evaluate((node) => {
      const form = node as HTMLFormElement;
      form.requestSubmit();
      form.requestSubmit();
    });
  } finally {
    releaseSave();
  }
  await expect(
    page.getByRole("button", { name: "Save profile" }),
  ).toBeEnabled();
  await expect(page.getByRole("status")).toBeVisible();
  expect(requests).toEqual([{ displayName: "Earlier name" }]);
  await expect(name).toHaveValue("Later name");
});

test("retains the newer profile name after a failed save and submits it on retry", async ({
  page,
  context,
}, testInfo) => {
  const { cookies } = await authenticatedCookies(testInfo);
  await context.addCookies(cookies);
  const requests: Record<string, unknown>[] = [];
  let releaseFailure!: () => void;
  const pendingFailure = new Promise<void>((resolve) => {
    releaseFailure = resolve;
  });
  await page.route("**/api/v1/settings", async (route) => {
    if (route.request().method() === "PATCH") {
      requests.push(route.request().postDataJSON());
      if (requests.length === 1) {
        await pendingFailure;
        await route.abort("failed");
        return;
      }
    }
    await route.continue();
  });
  await page.goto("/settings/profile");
  const name = page.getByRole("textbox", { name: "Display name" });
  await name.fill("Earlier name");
  await page.getByRole("button", { name: "Save profile" }).click();
  await expect.poll(() => requests.length).toBe(1);
  try {
    await name.fill("Later name");
  } finally {
    releaseFailure();
  }
  await expect(
    name.locator("xpath=ancestor::form").getByRole("alert"),
  ).toContainText("We couldn't save your profile.");
  await expect(name).toHaveValue("Later name");
  await expect(page.getByText("Your profile has been saved.")).toHaveCount(0);
  await page.getByRole("button", { name: "Save profile" }).click();
  await expect(page.getByRole("status")).toHaveText(
    "Your profile has been saved.",
  );
  expect(requests).toEqual([
    { displayName: "Earlier name" },
    { displayName: "Later name" },
  ]);
  await page.reload();
  await expect(name).toHaveValue("Later name");
});
