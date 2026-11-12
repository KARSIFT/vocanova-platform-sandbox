import { randomUUID } from "node:crypto";

import {
  expect,
  test,
  type JSHandle,
  type Page,
  type Request,
} from "@playwright/test";

const PREFIX = "vocanova:preferences-draft:";

function draftCount(departingDocument: JSHandle<Window>) {
  return departingDocument.evaluate(
    (win, prefix) =>
      Object.keys(win.sessionStorage).filter((key) => key.startsWith(prefix))
        .length,
    PREFIX,
  );
}

async function settleReleasedResponse(departingDocument: JSHandle<Window>) {
  // Wait for a browser rendering opportunity after the held response finishes,
  // so the fetch/JSON continuation and its synchronous storage writes have run.
  await departingDocument.evaluate(
    (win) =>
      new Promise<void>((resolve) =>
        win.requestAnimationFrame(() => resolve()),
      ),
  );
}

function waitForTerminalRequest(page: Page, target: Request): Promise<void> {
  return new Promise((resolve) => {
    const settled = (request: Request) => {
      if (request !== target) return;
      page.off("requestfinished", settled);
      page.off("requestfailed", settled);
      resolve();
    };
    page.on("requestfinished", settled);
    page.on("requestfailed", settled);
  });
}

for (const theme of ["light", "dark"] as const) {
  test.describe(`logout draft boundary (${theme})`, () => {
    test.beforeEach(async ({ context, baseURL, page }) => {
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL! },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
        { name: "vocanova_theme", value: theme, url: baseURL! },
      ]);
      await page.addInitScript(
        (value) => localStorage.setItem("vocanova:theme-preference", value),
        theme,
      );
    });

    test(`successful logout prevents a held profile response from recreating drafts (${theme})`, async ({
      page,
    }) => {
      let releaseSave!: () => void;
      const heldSave = new Promise<void>((resolve) => {
        releaseSave = resolve;
      });
      let committed = false;
      let heldRequest: Request | undefined;
      await page.route("**/api/v1/settings", async (route) => {
        if (route.request().method() !== "PATCH") return route.continue();
        heldRequest = route.request();
        const response = await route.fetch();
        expect(response.ok()).toBe(true);
        committed = true;
        await heldSave;
        await route.fulfill({ response });
      });
      let navigationStarted = false;
      await page.route("**/login?signedOut=1", async (route) => {
        navigationStarted = true;
        // Model Stop/offline navigation after the server has logged out:
        // the old document remains alive after the provisional load aborts.
        await route.abort("aborted");
      });
      let lateRefreshes = 0;
      page.on("request", (request) => {
        if (
          navigationStarted &&
          new URL(request.url()).pathname === "/settings/profile" &&
          request.headers().rsc === "1"
        )
          lateRefreshes++;
      });
      try {
        await page.goto("/settings/profile");
        await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
        // Capture the old document; aborting the sign-out destination below
        // leaves this document alive after successful server-side logout.
        const departingDocument = await page.evaluateHandle(() => window);
        const name = page.getByRole("textbox", {
          name: "Display name",
          exact: true,
        });
        await name.fill("Earlier profile save");
        await page
          .getByRole("button", { name: "Save profile", exact: true })
          .click();
        await expect.poll(() => committed).toBe(true);
        await name.fill("Private newer draft");
        await expect.poll(() => draftCount(departingDocument)).toBe(1);
        if (!heldRequest)
          throw new Error("Expected the pending profile request");
        const terminal = waitForTerminalRequest(page, heldRequest);
        const navigationFailed = page.waitForEvent("requestfailed", {
          predicate: (request) =>
            new URL(request.url()).pathname === "/login" &&
            new URL(request.url()).searchParams.get("signedOut") === "1",
        });
        await page
          .getByRole("button", { name: "Log out", exact: true })
          .click({ noWaitAfter: true });
        await navigationFailed;
        expect(navigationStarted).toBe(true);
        releaseSave();
        await terminal;
        await settleReleasedResponse(departingDocument);
        expect(await draftCount(departingDocument)).toBe(0);
        const lateUI = await departingDocument.evaluate((win) => ({
          statuses: Array.from(
            win.document.querySelectorAll('main [role="status"]'),
            (node) => node.textContent ?? "",
          ),
          alerts: win.document.querySelectorAll('main [role="alert"]').length,
        }));
        expect(
          lateUI.statuses.filter((text) => /saved|newer changes/.test(text)),
        ).toEqual([]);
        expect(lateUI.alerts).toBe(0);
        expect(lateRefreshes).toBe(0);
      } finally {
        releaseSave();
      }
    });

    test(`successful logout prevents a held email failure from restoring private input (${theme})`, async ({
      page,
    }) => {
      let releaseEmail!: () => void;
      const heldEmail = new Promise<void>((resolve) => {
        releaseEmail = resolve;
      });
      let requestStarted = false;
      let heldRequest: Request | undefined;
      await page.route(
        "**/api/v1/settings/email-change-links",
        async (route) => {
          heldRequest = route.request();
          requestStarted = true;
          await heldEmail;
          await route.fulfill({
            status: 503,
            contentType: "application/problem+json",
            body: JSON.stringify({ detail: "Deferred confirmation failure" }),
          });
        },
      );
      let navigationStarted = false;
      await page.route("**/login?signedOut=1", async (route) => {
        navigationStarted = true;
        // Model Stop/offline navigation after the server has logged out:
        // the old document remains alive after the provisional load aborts.
        await route.abort("aborted");
      });
      try {
        await page.goto("/settings/account");
        await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
        const departingDocument = await page.evaluateHandle(() => window);
        await page
          .getByRole("textbox", { name: "New sign-in email", exact: true })
          .fill("private-unsent@example.test");
        await page
          .getByRole("button", { name: "Send confirmation link", exact: true })
          .click();
        await expect.poll(() => requestStarted).toBe(true);
        await expect.poll(() => draftCount(departingDocument)).toBe(1);
        if (!heldRequest) throw new Error("Expected the pending email request");
        const terminal = waitForTerminalRequest(page, heldRequest);
        const navigationFailed = page.waitForEvent("requestfailed", {
          predicate: (request) =>
            new URL(request.url()).pathname === "/login" &&
            new URL(request.url()).searchParams.get("signedOut") === "1",
        });
        await page
          .getByRole("button", { name: "Log out", exact: true })
          .click({ noWaitAfter: true });
        await navigationFailed;
        expect(navigationStarted).toBe(true);
        releaseEmail();
        await terminal;
        await settleReleasedResponse(departingDocument);
        expect(await draftCount(departingDocument)).toBe(0);
        const lateUI = await departingDocument.evaluate((win) => ({
          statuses: Array.from(
            win.document.querySelectorAll('main [role="status"]'),
            (node) => node.textContent ?? "",
          ),
          alerts: Array.from(
            win.document.querySelectorAll('main [role="alert"]'),
            (node) => node.textContent ?? "",
          ),
        }));
        expect(
          lateUI.statuses.filter((text) =>
            /confirmation link|restored/.test(text),
          ),
        ).toEqual([]);
        expect(
          lateUI.alerts.filter((text) =>
            text.includes("Deferred confirmation failure"),
          ),
        ).toEqual([]);
      } finally {
        releaseEmail();
      }
    });
  });
}
