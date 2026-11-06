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

    for (const name of ["Light", "Dark", "System"]) {
      const labelFits = await page.getByRole("radio", { name }).evaluate((radio) => {
        const label = radio.closest("label")!;
        return label.scrollWidth <= label.clientWidth;
      });
      expect(labelFits, `${name} must fit inside its clickable label`).toBe(true);
    }

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

    for (const [route, heading] of [
      ["/", /Learn the words you will actually use\./],
      ["/signin", "Sign in to Vocanova"],
      ["/home", "Today"],
    ] as const) {
      await page.goto(route);
      await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");
      await expect(page.getByRole("heading", { level: 1, name: heading, exact: true })).toBeVisible();

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
    context,
    baseURL,
  }) => {
    await context.addCookies([
      { name: "vocanova_session", value: crypto.randomUUID(), url: baseURL! },
      { name: "vocanova_csrf", value: crypto.randomUUID(), url: baseURL! },
    ]);
    await useDarkTheme(page);

    for (const [route, heading] of [
      ["/settings", "Settings"],
      ["/settings/profile", "Profile"],
      ["/settings/account", "Account"],
      ["/words", "Saved vocabulary"],
      ["/reviews", /^(Keep your words fresh|You’re ready for new words)$/],
      ["/review", /^(Keep your words fresh|You’re ready for new words)$/],
    ] as const) {
      await page.goto(route);
      await expect.poll(() => new URL(page.url()).pathname).toBe(route);
      await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");
      // Scan the finished destination; held loading states have their own
      // explicit accessibility regression in route-loading.spec.ts.
      await expect(page.getByRole("heading", { level: 1, name: heading, exact: true })).toBeVisible();

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
