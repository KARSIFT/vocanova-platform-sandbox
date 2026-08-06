import { readFile } from "node:fs/promises";
import { resolve } from "node:path";

import { expect, test } from "@playwright/test";

const MIDDLEWARE_PATH = resolve(process.cwd(), "src/middleware.ts");
const DESKTOP_PROJECT_NAME = "home-desktop-1280";
const ONBOARDING_STATUS_COOKIE_NAME = "e2e_onboarding_status";
const ONBOARDING_STATUS_NOT_STARTED = "not_started";

test.describe("Middleware runtime regression (VOC-039-T01)", () => {
  test("requires node runtime export and preserves auth-gate redirect behavior", async ({
    page,
    context,
  }, testInfo) => {
    test.skip(
      testInfo.project.name !== DESKTOP_PROJECT_NAME,
      "Run once on the representative desktop project to keep the regression deterministic and fast.",
    );

    // This assertion is intentionally failing-first before VOC-039-T00:
    // without `runtime = "nodejs"`, middleware runs on Edge and can
    // silently diverge from Node fetch behavior in production.
    const middlewareSource = await readFile(MIDDLEWARE_PATH, "utf8");
    expect(middlewareSource).toMatch(
      /\bexport\s+const\s+runtime\s*=\s*["']nodejs["']\s*;/,
    );

    const baseURL = testInfo.project.use.baseURL;
    if (!baseURL) {
      throw new Error(
        "Expected Playwright baseURL to be configured so middleware cookies can be scoped to the app origin.",
      );
    }

    await context.addCookies([
      {
        name: ONBOARDING_STATUS_COOKIE_NAME,
        value: ONBOARDING_STATUS_NOT_STARTED,
        url: baseURL,
      },
    ]);

    // The request goes through real Next middleware and its /api/v1/me
    // fetch. If auth-check fetch behavior regresses, this redirect check
    // catches it in the same end-to-end path production uses.
    await page.goto("/home");
    await expect(page).toHaveURL(/\/onboarding(\?|$)/);
  });
});
