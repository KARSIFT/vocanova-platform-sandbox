import {
  expect,
  test,
  type BrowserContext,
  type Page,
  type TestInfo,
} from "@playwright/test";

import { scanForAxeViolations } from "./axe-helper.js";

type Capabilities = {
  magicLinkEnabled: boolean;
  oauthEnabled: boolean;
  passwordEnabled: boolean;
};

async function setCapabilities(
  context: BrowserContext,
  baseURL: string | undefined,
  capabilities: Capabilities,
) {
  if (!baseURL) throw new Error("Expected a configured base URL.");
  await context.addCookies([
    {
      name: "e2e_magic_link_enabled",
      value: String(capabilities.magicLinkEnabled),
      url: baseURL,
    },
    {
      name: "e2e_oauth_enabled",
      value: String(capabilities.oauthEnabled),
      url: baseURL,
    },
    {
      name: "e2e_password_enabled",
      value: String(capabilities.passwordEnabled),
      url: baseURL,
    },
  ]);
}

async function rejectOAuthStart(page: Page) {
  await page.route("**/api/v1/auth/oauth/google/start", (route) =>
    route.fulfill({
      status: 503,
      contentType: "application/json",
      body: JSON.stringify({ message: "private-provider-diagnostic" }),
    }),
  );
}

async function captureLayout(page: Page, testInfo: TestInfo, name: string) {
  const overflow = await page.evaluate(
    () => document.documentElement.scrollWidth - window.innerWidth,
  );
  expect(overflow, "the authentication page must fit the viewport").toBeLessThanOrEqual(1);
  await page.screenshot({ path: testInfo.outputPath(`${name}.png`), fullPage: true });
}

for (const colorScheme of ["light", "dark"] as const) {
  test.describe(`${colorScheme} authentication capability guidance`, () => {
    test.beforeEach(async ({ page }) => {
      await page.emulateMedia({ colorScheme });
    });

    test("Google-only cancellation offers retry without unavailable email methods", async ({
      page,
      context,
      baseURL,
    }) => {
      await setCapabilities(context, baseURL, {
        magicLinkEnabled: false,
        oauthEnabled: true,
        passwordEnabled: false,
      });
      await page.goto("/signin?oauth=cancelled&returnTo=%2Freviews%3Fmode%3Ddue");

      const alert = page.getByRole("main").getByRole("alert");
      await expect(alert).toContainText("Google sign-in was cancelled.");
      await expect(alert).toContainText(/try again/i);
      await expect(alert).not.toContainText(/email|password|sign-in link/i);
      await expect(page.getByRole("textbox")).toHaveCount(0);
      await expect(
        page.getByRole("button", { name: "Continue with Google" }),
      ).toBeEnabled();

      const { criticalOrSerious } = await scanForAxeViolations(page);
      expect(criticalOrSerious).toEqual([]);
    });

    for (const route of ["/signin", "/signup"]) {
      test(`${route} keeps Google-only start failures actionable without exposing diagnostics`, async ({
        page,
        context,
        baseURL,
      }, testInfo) => {
        await setCapabilities(context, baseURL, {
          magicLinkEnabled: false,
          oauthEnabled: true,
          passwordEnabled: false,
        });
        await rejectOAuthStart(page);
        await page.goto(route);
        const google = page.getByRole("button", { name: "Continue with Google" });
        await google.focus();
        await page.keyboard.press("Enter");

        const alert = page.getByRole("main").getByRole("alert");
        await expect(alert).toContainText("Google sign-in is unavailable right now.");
        await expect(alert).toContainText(/try again/i);
        await expect(alert).not.toContainText(/email|password|sign-in link/i);
        await expect(alert).not.toContainText("private-provider-diagnostic");
        await expect(google).toBeEnabled();
        await captureLayout(page, testInfo, "google-only-failure");
      });
    }

    for (const alternate of ["password", "magic-link"] as const) {
      test(`Google start errors point to the available ${alternate} method`, async ({
        page,
        context,
        baseURL,
      }) => {
        await setCapabilities(context, baseURL, {
          magicLinkEnabled: alternate === "magic-link",
          oauthEnabled: true,
          passwordEnabled: alternate === "password",
        });
        await rejectOAuthStart(page);
        await page.goto("/signin");
        await page.getByRole("button", { name: "Continue with Google" }).click();

        const alert = page.getByRole("main").getByRole("alert");
        if (alternate === "password") {
          await expect(alert).toContainText(/email and password/i);
          await expect(page.getByLabel("Password", { exact: true })).toBeVisible();
          await expect(page.getByRole("button", { name: "Send sign-in link" })).toHaveCount(0);
        } else {
          await expect(alert).toContainText(/email sign-in link/i);
          await expect(page.getByRole("button", { name: "Send sign-in link" })).toBeVisible();
          await expect(page.getByLabel("Password", { exact: true })).toHaveCount(0);
        }
      });
    }

    test("unavailable email-link recovery leads to enabled password sign-in and preserves the destination", async ({
      page,
      context,
      baseURL,
    }, testInfo) => {
      await setCapabilities(context, baseURL, {
        magicLinkEnabled: false,
        oauthEnabled: false,
        passwordEnabled: true,
      });
      await page.goto("/login?magicOnly=1&returnTo=%2Freviews%3Fmode%3Ddue");
      const status = page.getByRole("main").getByRole("status");
      await expect(status).toContainText("Email sign-in links are unavailable right now.");
      await expect(status).toContainText(/email and password/i);
      await expect(status).not.toContainText(/Google/i);
      await expect(page.getByRole("button", { name: "Continue with Google" })).toHaveCount(0);
      await captureLayout(page, testInfo, "password-recovery-guidance");

      const standardSignIn = page.getByRole("link", { name: "use the standard sign-in page" });
      await standardSignIn.focus();
      await page.keyboard.press("Enter");
      await expect(page).toHaveURL(/\/login\?returnTo=%2Freviews%3Fmode%3Ddue$/);
      await expect(page.getByLabel("Password", { exact: true })).toBeVisible();
      await expect(page.getByRole("button", { name: "Send sign-in link" })).toHaveCount(0);
      await captureLayout(page, testInfo, "password-recovery-destination");
    });

    test("unavailable recovery does not invent a method when all are disabled", async ({
      page,
      context,
      baseURL,
    }) => {
      await setCapabilities(context, baseURL, {
        magicLinkEnabled: false,
        oauthEnabled: false,
        passwordEnabled: false,
      });
      await page.goto("/login?magicOnly=1");
      const main = page.getByRole("main");
      await expect(main.getByRole("status")).toContainText("Please try again later.");
      await expect(main.getByRole("status")).not.toContainText(/Google|email and password/);
      await expect(main.getByRole("alert")).toHaveText("Sign-in is temporarily unavailable. Please try again later.");
      await expect(page.getByRole("textbox")).toHaveCount(0);
      await expect(page.getByRole("button", { name: "Continue with Google" })).toHaveCount(0);
      await expect(page.getByRole("link", { name: "use the standard sign-in page" })).toHaveCount(0);
    });

    test("password-only signup does not advertise disabled Google", async ({
      page,
      context,
      baseURL,
    }) => {
      await setCapabilities(context, baseURL, {
        magicLinkEnabled: false,
        oauthEnabled: false,
        passwordEnabled: true,
      });
      await page.goto("/signup");
      await expect(page.getByRole("main")).toContainText("We'll email a verification link before your password can be used.");
      await expect(page.getByRole("main")).not.toContainText(/Google/);
      await expect(page.getByRole("button", { name: "Create account" })).toBeVisible();
    });
  });
}
