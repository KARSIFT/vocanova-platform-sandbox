import { expect, test } from "@playwright/test";

import { formatViolations, scanForAxeViolations } from "./axe-helper.js";

async function useDarkTheme(page: import("@playwright/test").Page) {
  await page.addInitScript(() => {
    window.localStorage.setItem("vocanova:theme-preference", "dark");
  });
}

test.describe("Theme preference", () => {
  test("restores a stored dark choice before hydration without console errors", async ({
    page,
  }) => {
    const runtimeErrors: string[] = [];
    page.on("pageerror", (error) => runtimeErrors.push(error.message));
    page.on("console", (message) => {
      if (message.type() === "error") {
        runtimeErrors.push(message.text());
      }
    });
    await useDarkTheme(page);

    await page.goto("/settings");
    await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");
    await expect(page.getByRole("radio", { name: "Dark" })).toBeChecked();
    expect(runtimeErrors).toEqual([]);
  });

  test("defaults to System, supports keyboard selection, and follows only System changes", async ({
    page,
  }) => {
    await page.emulateMedia({ colorScheme: "dark" });
    await page.goto("/settings");

    const root = page.locator("html");
    await expect(root).toHaveAttribute("data-theme", "dark");
    await expect(root).toHaveAttribute("data-theme-preference", "system");
    await expect(page.getByRole("radio", { name: "System" })).toBeChecked();

    const light = page.getByRole("radio", { name: "Light" });
    await light.focus();
    await page.keyboard.press("Space");
    await expect(light).toBeChecked();
    await expect(root).toHaveAttribute("data-theme", "light");
    await expect
      .poll(() =>
        page.evaluate(() =>
          window.localStorage.getItem("vocanova:theme-preference"),
        ),
      )
      .toBe("light");

    await page.emulateMedia({ colorScheme: "dark" });
    await expect(root).toHaveAttribute("data-theme", "light");

    await page.getByRole("radio", { name: "System" }).check();
    await expect(root).toHaveAttribute("data-theme", "dark");
  });

  test("keeps dark theme accessible across public, auth, and app surfaces", async ({
    page,
  }) => {
    await useDarkTheme(page);

    for (const route of ["/", "/signin", "/home"]) {
      await page.goto(route);
      await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");

      const { criticalOrSerious } = await scanForAxeViolations(page);
      expect(
        criticalOrSerious,
        `Expected zero critical or serious axe-core violations in dark mode on ${route}; found:\n${formatViolations(
          criticalOrSerious,
        ).join("\n")}`,
      ).toEqual([]);
    }
  });

  test("keeps dark theme accessible across authenticated account and practice surfaces", async ({
    page,
  }) => {
    await useDarkTheme(page);

    for (const route of [
      "/settings",
      "/settings/profile",
      "/settings/account",
      "/words",
      "/reviews",
      "/review",
    ]) {
      await page.goto(route);
      await expect.poll(() => new URL(page.url()).pathname).toBe(route);
      await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");

      const { criticalOrSerious } = await scanForAxeViolations(page);
      expect(
        criticalOrSerious,
        `Expected zero critical or serious axe-core violations in dark mode on ${route}; found:\n${formatViolations(
          criticalOrSerious,
        ).join("\n")}`,
      ).toEqual([]);
    }
  });

  test("keeps an invalid password message readable in dark mode", async ({
    page,
  }) => {
    await useDarkTheme(page);
    await page.goto("/signup");
    await page.getByLabel("Email address").fill("learner@example.com");
    await page.getByLabel("Password", { exact: true }).fill("too short");
    await page.getByRole("button", { name: "Create account" }).click();
    await expect(
      page
        .getByRole("alert")
        .filter({ hasText: "Use a password between 15 and 128 characters." }),
    ).toBeVisible();

    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(
      criticalOrSerious,
      `Expected zero critical or serious axe-core violations in dark mode on invalid password feedback; found:\n${formatViolations(
        criticalOrSerious,
      ).join("\n")}`,
    ).toEqual([]);
  });
});
