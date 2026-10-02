// VOC-031-T07b accessibility scan for /settings.
//
// /settings is the screen that renders every editable Settings
// field (daily review target, review rhythm, app language,
// product emails, display name) and the
// "Account security" link to the deeper account sub-screen.
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
  for (const theme of ["light", "dark"] as const) {
    test(`the product email switch provides a large touch target and retains explicit save in ${theme} mode`, async ({
      page,
      context,
      baseURL,
    }, testInfo) => {
      if (!baseURL) throw new Error("A test app URL is required");
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
        { name: "vocanova_theme", value: theme, url: baseURL },
      ]);
      const submitted: Record<string, unknown>[] = [];
      page.on("request", (request) => {
        if (
          request.method() === "PATCH" &&
          new URL(request.url()).pathname === "/api/v1/settings"
        ) {
          submitted.push(request.postDataJSON());
        }
      });
      await page.goto("/settings");
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      const form = page.getByRole("form", { name: "Practice settings" });
      const preferences = [
        { name: "Product news and tips", field: "marketingEmailsEnabled" },
      ];
      const changes: Record<string, boolean> = {};
      for (const preference of preferences) {
        const control = form.getByRole("switch", {
          name: preference.name,
          exact: true,
        });
        await expect(control).toBeVisible();
        await control.scrollIntoViewIfNeeded();
        const box = await control.boundingBox();
        expect(box).not.toBeNull();
        expect(box!.width).toBeGreaterThanOrEqual(44);
        expect(box!.height).toBeGreaterThanOrEqual(44);
        const initiallyChecked =
          (await control.getAttribute("aria-checked")) === "true";
        changes[preference.field] = !initiallyChecked;

        // The area above the compact visual track must also activate the switch.
        await control.click({ position: { x: box!.width / 2, y: 2 } });
        await expect(control).toHaveAttribute(
          "aria-checked",
          String(!initiallyChecked),
        );
        await control.press("Space");
        await expect(control).toBeFocused();
        await expect(control).toHaveAttribute(
          "aria-checked",
          String(initiallyChecked),
        );
        await control.press("Space");
        await expect(control).toHaveAttribute(
          "aria-checked",
          String(!initiallyChecked),
        );
      }
      expect(submitted).toEqual([]);
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth),
      ).toBeLessThanOrEqual(page.viewportSize()!.width);
      await form
        .getByRole("group", { name: "Notifications and emails" })
        .screenshot({
          path: testInfo.outputPath(`settings-switch-targets-${theme}.png`),
        });
      await form.getByRole("button", { name: "Save settings" }).click();
      await expect(form.getByRole("status")).toHaveText(
        "Your settings have been saved.",
      );
      expect(submitted).toEqual([changes]);
      await page.reload();
      for (const preference of preferences) {
        await expect(
          form.getByRole("switch", { name: preference.name, exact: true }),
        ).toHaveAttribute("aria-checked", String(changes[preference.field]));
      }
    });
  }

  test("keeps settings reading order consistent with the visual layout", async ({
    page,
  }) => {
    await page.goto("/settings");
    const account = page.getByRole("complementary");
    const learning = page.getByRole("heading", {
      name: "Learning preferences",
    });
    const [accountBox, learningBox] = await Promise.all([
      account.boundingBox(),
      learning.boundingBox(),
    ]);
    expect(accountBox).not.toBeNull();
    expect(learningBox).not.toBeNull();
    const accountComesFirst = await account.evaluate((aside) => {
      const form = document.querySelector(
        'form[aria-label="Practice settings"]',
      )!;
      return Boolean(
        aside.compareDocumentPosition(form) & Node.DOCUMENT_POSITION_FOLLOWING,
      );
    });
    expect(accountComesFirst).toBe(true);
    if (page.viewportSize()!.width >= 1024) {
      expect(accountBox!.x + accountBox!.width).toBeLessThanOrEqual(
        learningBox!.x,
      );
    } else {
      expect(accountBox!.y + accountBox!.height).toBeLessThanOrEqual(
        learningBox!.y,
      );
    }
  });

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
    const customNotice = page
      .getByRole("form", { name: "Practice settings" })
      .getByText(
        "Your saved custom rhythm stays saved until you choose one of the available options.",
      );
    await expect(customNotice).toBeVisible();
    await expect(page.getByRole("radio", { name: /Custom/ })).toHaveCount(0);
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
    await expect(customNotice).toBeVisible();
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
    // radios, the product email switch, display name, calendar reminder fields,
    // save/download buttons and account links provide many focusable elements.
    // Use a conservative floor.
    await assertKeyboardReachable(page, { minFocusable: 10 });

    await assertNonColorOnlyFeedback(page, {
      contextLabel: "/settings",
      requireText: [
        "text=Daily review target",
        "text=Review rhythm",
        "text=Account security",
        "text=Vocanova default",
        "text=Faster reviews",
      ],
    });

    await expect(page.getByRole("radio", { name: /Custom/ })).toHaveCount(0);
  });
});
