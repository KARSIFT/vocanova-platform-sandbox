import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";

const verifyEndpoint = "**/api/v1/auth/password/signups/verify";
const consumeEndpoint = "**/api/v1/settings/email-change-links/consume";
const longEmail = `${"learner".repeat(8)}@${"example".repeat(8)}.test`;

async function expectNoOverflow(page: Page) {
  expect(await page.evaluate(() => document.documentElement.scrollWidth))
    .toBeLessThanOrEqual(page.viewportSize()!.width);
}

// Isolated frontend responses establish recovery and focus, not live email
// delivery or proof durability after an uncertain server response.
for (const theme of ["light", "dark"] as const) {
  test.describe(`confirmation recovery (${theme})`, () => {
    test.beforeEach(async ({ page, context, baseURL }) => {
      await page.addInitScript((preference) => {
        localStorage.setItem("vocanova:theme-preference", preference);
      }, theme);
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL! },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
        { name: "vocanova_theme", value: theme, url: baseURL! },
      ]);
    });

    for (const failure of ["connection", "unavailable", "rate-limited"] as const) {
      test(`password verification retries ${failure} with the original proof`, async ({ page }) => {
        const token = randomUUID();
        let requests = 0;
        await page.route(verifyEndpoint, async (route) => {
          requests += 1;
          expect(route.request().postDataJSON()).toEqual({ token });
          if (requests > 1) return route.fulfill({ status: 204 });
          if (failure === "connection") return route.abort("failed");
          return route.fulfill({
            status: failure === "unavailable" ? 503 : 429,
            json: { detail: "Temporary synthetic failure" },
          });
        });
        await page.goto(`/auth/password/verify?token=${token}`);
        await page.getByRole("button", { name: "Verify email", exact: true }).click();
        const retry = page.getByRole("button", { name: "Retry verification", exact: true });
        await expect(retry).toBeVisible();
        await expect(retry).toBeFocused();
        await expect(page.getByRole("main").getByRole("alert")).not.toContainText("invalid, expired");
        expect(new URL(page.url()).searchParams.get("token")).toBe(token);
        await expectNoOverflow(page);
        await retry.press("Enter");
        await expect(page.getByRole("status")).toContainText("password is ready to use");
        await expect(page.getByRole("link", { name: "Sign in", exact: true })).toBeFocused();
        await expect(page).toHaveURL(/\/auth\/password\/verify$/);
        expect(requests).toBe(2);
      });
    }

    test("password verification rejects a definitively invalid proof", async ({ page }) => {
      await page.route(verifyEndpoint, (route) => route.fulfill({
        status: 401, json: { detail: "invalid credentials or expired link" },
      }));
      await page.goto(`/auth/password/verify?token=${randomUUID()}`);
      await page.getByRole("button", { name: "Verify email", exact: true }).click();
      await expect(page.getByRole("main").getByRole("alert")).toContainText("invalid, expired, or has already been used");
      await expect(page.getByRole("button", { name: "Retry verification" })).toHaveCount(0);
      await expect(page.getByRole("link", { name: "Create account", exact: true })).toBeFocused();
    });

    for (const phase of ["identity", "consume"] as const) {
      for (const failure of ["connection", "unavailable"] as const) {
        test(`email confirmation retries ${phase} ${failure} without losing its proof`, async ({ page }) => {
          const token = randomUUID();
          let identityRequests = 0;
          let consumes = 0;
          await page.route("**/api/v1/me", async (route) => {
            identityRequests += 1;
            if (phase === "identity" && identityRequests === 1) {
              if (failure === "connection") return route.abort("failed");
              return route.fulfill({ status: 503, json: { detail: "Unavailable" } });
            }
            return route.fulfill({ json: { id: "confirmation-learner", email: "learner@example.test" } });
          });
          await page.route(consumeEndpoint, async (route) => {
            consumes += 1;
            expect(route.request().postDataJSON()).toEqual({ token });
            if (phase === "consume" && consumes === 1) {
              if (failure === "connection") return route.abort("failed");
              return route.fulfill({ status: 503, json: { detail: "Unavailable" } });
            }
            return route.fulfill({ json: { email: longEmail } });
          });
          await page.goto(`/auth/email-change?token=${token}`);
          const retry = page.getByRole("button", { name: "Retry confirmation", exact: true });
          await expect(retry).toBeFocused();
          await expect(page.getByRole("main").getByRole("alert")).not.toContainText("invalid or has expired");
          expect(new URL(page.url()).searchParams.get("token")).toBe(token);
          await retry.press("Enter");
          await expect(page.getByRole("status")).toContainText(`Your sign-in email is now ${longEmail}`);
          await expect(page.getByRole("link", { name: "Return to account settings" })).toBeFocused();
          await expectNoOverflow(page);
          expect(consumes).toBe(phase === "consume" ? 2 : 1);
        });
      }
    }

    test("email confirmation keeps an uncertain identity recheck recoverable", async ({ page }) => {
      const token = randomUUID();
      let identityRequests = 0;
      let consumes = 0;
      await page.route("**/api/v1/me", async (route) => {
        identityRequests += 1;
        return identityRequests === 2
          ? route.fulfill({ status: 503, json: { detail: "Unavailable" } })
          : route.fulfill({ json: { id: "confirmation-learner" } });
      });
      await page.route(consumeEndpoint, (route) => {
        consumes += 1;
        expect(route.request().postDataJSON()).toEqual({ token });
        return consumes === 1
          ? route.fulfill({ status: 401, json: { detail: "invalid or expired email change link" } })
          : route.fulfill({ json: { email: "new-learner@example.test" } });
      });
      await page.goto(`/auth/email-change?token=${token}`);
      const retry = page.getByRole("button", { name: "Retry confirmation" });
      await expect(retry).toBeFocused();
      await retry.press("Enter");
      await expect(page.getByRole("status")).toContainText("new-learner@example.test");
    });

    test("email confirmation offers account recovery for a confirmed conflict", async ({ page }) => {
      await page.route("**/api/v1/me", (route) => route.fulfill({ json: { id: "confirmation-learner" } }));
      await page.route(consumeEndpoint, (route) => route.fulfill({
        status: 409, json: { detail: "new email is already registered to another account" },
      }));
      await page.goto(`/auth/email-change?token=${randomUUID()}`);
      await expect(page.getByRole("main").getByRole("alert")).toContainText("could not be used");
      await expect(page.getByRole("button", { name: "Retry confirmation" })).toHaveCount(0);
      await expect(page.getByRole("link", { name: "Go to account settings" })).toBeFocused();
    });

    test("pending email-change address wraps on small screens", async ({ page }) => {
      await page.route("**/api/v1/settings/email-change-links", (route) => route.fulfill({ status: 204 }));
      await page.goto("/settings/account");
      await page.getByLabel("New sign-in email", { exact: true }).fill(longEmail);
      await page.getByRole("button", { name: "Send confirmation link", exact: true }).click();
      await expect(page.getByRole("status")).toContainText(longEmail);
      await page.addStyleTag({ content: "html { font-size: 200% !important; }" });
      await expectNoOverflow(page);
    });
  });
}
