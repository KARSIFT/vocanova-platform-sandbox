// VOC-031-T07b accessibility scan for /settings.
//
// /settings is the screen that renders every editable Settings
// field (daily review target, review rhythm, app language,
// notifications, marketing emails, display name) and the
// "Manage account" link to the deeper account sub-screen.
// /settings/account coverage lives in settings-account-accessibility.spec.ts
// (VOC-073-T03).

import { expect, test } from "@playwright/test";
import { randomUUID } from "node:crypto";

import {
  assertKeyboardReachable,
  assertNonColorOnlyFeedback,
  formatViolations,
  scanForAxeViolations,
} from "./axe-helper.js";

test.describe("Settings accessibility (VOC-031-T07b)", () => {
  test("preserves a stored Custom preset when saving unrelated settings", async ({
    page,
    context,
    baseURL,
  }) => {
    if (!baseURL) throw new Error("A test app URL is required");
    const csrf = `test-csrf-${randomUUID()}`;
    await context.addCookies([
      {
        name: "vocanova_session",
        value: `custom-preset-${randomUUID()}`,
        url: baseURL,
      },
      { name: "vocanova_csrf", value: csrf, url: baseURL },
    ]);
    const mockAPI = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;
    const seed = await page.request.patch(`${mockAPI}/api/v1/settings`, {
      headers: { "X-CSRF-Token": csrf },
      data: { reviewIntervalPreset: "custom" },
    });
    expect(seed.ok()).toBeTruthy();
    await page.goto("/settings");
    const custom = page.getByRole("radio", { name: "Custom (coming soon)" });
    await expect(custom).toBeChecked();
    await expect(custom).toBeDisabled();
    const patches: unknown[] = [];
    page.on("request", (request) => {
      if (
        request.method() === "PATCH" &&
        new URL(request.url()).pathname === "/api/v1/settings"
      ) {
        patches.push(request.postDataJSON());
      }
    });
    await page.getByRole("button", { name: "Save settings" }).click();
    await expect(
      page.getByText("Your settings have been saved."),
    ).toBeVisible();
    expect(patches).toEqual([]);
    await page.getByLabel("Display name").fill("Custom preset learner");
    const submission = page.waitForRequest(
      (request) =>
        request.method() === "PATCH" &&
        new URL(request.url()).pathname === "/api/v1/settings",
    );
    await page.getByRole("button", { name: "Save settings" }).click();
    expect((await submission).postDataJSON()).not.toHaveProperty(
      "reviewIntervalPreset",
    );
    await expect(
      page.getByText("Your settings have been saved."),
    ).toBeVisible();
    await page.reload();
    await expect(custom).toBeChecked();
    await expect(custom).toBeDisabled();
  });

  test("/settings renders with zero critical/serious axe violations, is keyboard reachable, and uses text-based state", async ({
    page,
  }) => {
    await page.goto("/settings");

    await expect(
      page.getByRole("heading", { name: "Settings", level: 1 }),
    ).toBeVisible();

    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(
      criticalOrSerious,
      `Expected zero critical or serious axe-core violations on /settings; found:\n${formatViolations(
        criticalOrSerious,
      ).join("\n")}`,
    ).toEqual([]);

    // /settings has 8 daily-review-target radios + 2 available review-rhythm
    // radios (the unshipped custom choice is disabled) + 2 checkbox toggles + 1 display-name input + 1 save
    // button + 2 links (Back to Home, Manage account) = 17+
    // focusable elements. Use a conservative floor.
    await assertKeyboardReachable(page, { minFocusable: 10 });

    await assertNonColorOnlyFeedback(page, {
      contextLabel: "/settings",
      requireText: [
        "text=Daily review target",
        "text=Review rhythm",
        "text=Manage account",
        "text=Vocanova default",
        "text=Faster reminders",
      ],
    });

    const customReviewRhythm = page.getByRole("radio", {
      name: "Custom (coming soon)",
    });
    await expect(customReviewRhythm).toBeDisabled();
    await expect(
      page.getByText("Custom review schedules are not available yet."),
    ).toBeVisible();
  });
});
