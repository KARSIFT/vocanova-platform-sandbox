import { randomUUID } from "node:crypto";

import { expect, test, type BrowserContext, type Page } from "@playwright/test";
import { formatViolations, scanForAxeViolations } from "./axe-helper";
import { pressTab } from "./keyboard-navigation";

const PASSWORD = "a password long enough";
const PREFIX = "vocanova:preferences-draft:";

async function authenticate(context: BrowserContext, baseURL?: string) {
  if (!baseURL) throw new Error("Expected a test URL");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
  ]);
}

async function checkRecoveryUI(
  page: Page,
  screenshotPath: string,
  theme: "light" | "dark",
) {
  await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
  await page.addStyleTag({ content: "html { font-size: 200% !important; }" });
  const main = page.getByRole("main");
  const recoveryButtons = main.getByRole("button", {
    name: /Resend|Use a different email|Sending\.\.\./,
  });
  await expect(recoveryButtons).toHaveCount(2);
  for (const button of await recoveryButtons.all()) {
    await expect(button).toBeVisible();
    const box = await button.boundingBox();
    expect(box?.height).toBeGreaterThanOrEqual(44);
    expect(box?.width).toBeGreaterThanOrEqual(44);
  }
  const different = main.getByRole("button", {
    name: "Use a different email",
    exact: true,
  });
  await different.focus();
  await pressTab(page, { backwards: true });
  await pressTab(page);
  await expect(different).toBeFocused();
  await different.scrollIntoViewIfNeeded();
  expect(
    await different.evaluate((button) => {
      const box = button.getBoundingClientRect();
      const hit = document.elementFromPoint(
        box.left + box.width / 2,
        box.top + box.height / 2,
      );
      return Boolean(hit && (hit === button || button.contains(hit)));
    }),
  ).toBe(true);
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth,
    ),
  ).toBe(true);
  // The keyboard and hit-target checks above verify the operative controls.
  // Normalize scroll before the whole-document contrast/accessibility scan.
  await page.evaluate(() => window.scrollTo(0, 0));
  const scan = await scanForAxeViolations(page);
  expect(
    scan.criticalOrSerious,
    formatViolations(scan.criticalOrSerious).join("\n"),
  ).toEqual([]);
  await page.screenshot({ path: screenshotPath, fullPage: true });
}

test("account drafts stay private across account changes and explicit logout", async ({
  page,
  context,
  baseURL,
}) => {
  await authenticate(context, baseURL);
  await page.goto("/settings/profile");
  await page
    .getByRole("textbox", { name: "Display name", exact: true })
    .fill("Private profile draft");
  await page.goto("/settings/account");
  await page
    .getByRole("textbox", { name: "New sign-in email" })
    .fill("private@example.test");
  const draftCount = () =>
    page.evaluate(
      (prefix) =>
        Object.keys(sessionStorage).filter((key) => key.startsWith(prefix))
          .length,
      PREFIX,
    );
  await expect.poll(draftCount).toBe(2);
  await context.addCookies([
    { name: "e2e_user_id", value: "alternate", url: baseURL! },
  ]);
  await page.reload();
  await expect(
    page.getByRole("textbox", { name: "New sign-in email" }),
  ).toHaveValue("");
  await expect.poll(draftCount).toBe(0);
  await page.goto("/settings/profile");
  await expect(
    page.getByRole("textbox", { name: "Display name", exact: true }),
  ).not.toHaveValue("Private profile draft");
  await page
    .getByRole("textbox", { name: "Display name", exact: true })
    .fill("Another private draft");
  await page.getByRole("button", { name: "Log out", exact: true }).click();
  await expect(page).toHaveURL(/\/signin(\?|$)|\/login(\?|$)/);
  await expect.poll(draftCount).toBe(0);
});

test("a newer profile edit survives an older confirmed save", async ({
  page,
  context,
  baseURL,
}) => {
  await authenticate(context, baseURL);
  let release: (() => void) | undefined;
  const pending = new Promise<void>((resolve) => {
    release = resolve;
  });
  let committed = false;
  let first = true;
  await page.route("**/api/v1/settings", async (route) => {
    if (route.request().method() === "PATCH" && first) {
      first = false;
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      committed = true;
      await pending;
      return route.fulfill({ response });
    }
    await route.continue();
  });
  await page.goto("/settings/profile");
  const name = page.getByRole("textbox", { name: "Display name", exact: true });
  await name.fill("Earlier confirmed name");
  await page.getByRole("button", { name: "Save profile", exact: true }).click();
  await expect.poll(() => committed).toBe(true);
  try {
    await name.fill("Newer unsaved name");
  } finally {
    release?.();
  }
  await expect(page.getByRole("status")).toContainText("newer changes");
  await expect(name).toHaveValue("Newer unsaved name");
  await page.reload();
  await expect(name).toHaveValue("Newer unsaved name");
  await expect(page.getByRole("status")).toContainText("restored");
  await page.getByRole("button", { name: "Save profile", exact: true }).click();
  await expect(page.getByRole("status")).toContainText(
    "Your profile has been saved",
  );
});

test("a reverted profile draft survives a committed save with a lost response", async ({
  page,
  context,
  baseURL,
}) => {
  await authenticate(context, baseURL);
  let release: (() => void) | undefined;
  const pending = new Promise<void>((resolve) => {
    release = resolve;
  });
  let committed = false;
  let first = true;
  await page.route("**/api/v1/settings", async (route) => {
    if (route.request().method() === "PATCH" && first) {
      first = false;
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      committed = true;
      await pending;
      return route.abort("failed");
    }
    await route.continue();
  });
  await page.goto("/settings/profile");
  const name = page.getByRole("textbox", { name: "Display name", exact: true });
  const original = await name.inputValue();
  const avatar = page
    .getByRole("region", { name: "Your profile", exact: true })
    .locator('[aria-hidden="true"]');
  const originalInitials = await avatar.innerText();
  await name.fill("Committed before response failed");
  await page.getByRole("button", { name: "Save profile", exact: true }).click();
  await expect.poll(() => committed).toBe(true);
  try {
    await name.fill(original);
  } finally {
    release?.();
  }
  await expect(
    page
      .getByRole("main")
      .getByRole("alert")
      .filter({ hasText: "We couldn't save your profile. Please try again." }),
  ).toBeVisible();
  await page.reload();
  await expect(name).toHaveValue(original);
  await expect(page.getByRole("status")).toContainText("restored");
  // The lost response followed a real committed write. Confirm that reload
  // rendered its avatar before explicitly restoring the learner's draft.
  await expect(avatar).toHaveText("CB");
  await page.getByRole("button", { name: "Save profile", exact: true }).click();
  await expect(page.getByRole("status")).toContainText(
    "Your profile has been saved",
  );
  // Save acknowledgement precedes router.refresh. Wait for its observable
  // server-rendered avatar before starting another document navigation.
  await expect(avatar).toHaveText(originalInitials);
  await page.reload();
  await expect(name).toHaveValue(original);
  await expect(
    page.getByRole("status").filter({ hasText: "restored" }),
  ).toHaveCount(0);
});

for (const theme of ["light", "dark"] as const) {
  test(`password signup can resend privately and change address (${theme})`, async ({
    page,
    context,
    baseURL,
  }, testInfo) => {
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    await page.addInitScript(
      (value) => localStorage.setItem("vocanova:theme-preference", value),
      theme,
    );
    const bodies: Array<Record<string, string>> = [];
    await page.route("**/api/v1/auth/password/signups", async (route) => {
      bodies.push(route.request().postDataJSON());
      await route.fulfill({
        status: 202,
        contentType: "application/json",
        body: "{}",
      });
    });
    await page.goto("/signup");
    await expect(
      page.getByRole("textbox", {
        name: "Display name (optional)",
        exact: true,
      }),
    ).toHaveAccessibleDescription("Up to 80 characters.");
    await page.clock.install();
    await page.getByLabel("Email address").fill("learner@example.test");
    await page
      .getByRole("textbox", { name: "Password", exact: true })
      .fill(PASSWORD);
    await page
      .getByRole("button", { name: "Create account", exact: true })
      .click();
    await expect(
      page
        .getByRole("status")
        .filter({ hasText: "If this address can create an account" }),
    ).toBeVisible();
    await expect(
      page.getByRole("button", { name: /Resend available in/ }),
    ).toBeDisabled();
    await page.clock.fastForward(60_000);
    await page
      .getByRole("button", { name: "Resend verification link", exact: true })
      .click();
    await expect.poll(() => bodies.length).toBe(2);
    expect(bodies[1]).toEqual(bodies[0]);
    expect(
      await page.evaluate(() =>
        JSON.stringify({ sessionStorage, localStorage }),
      ),
    ).not.toContain(PASSWORD);
    await checkRecoveryUI(
      page,
      testInfo.outputPath(`signup-recovery-${theme}.png`),
      theme,
    );
    await page
      .getByRole("button", { name: "Use a different email", exact: true })
      .press("Enter");
    await expect(page.getByLabel("Email address")).toBeFocused();
    await expect(
      page.getByRole("textbox", { name: "Password", exact: true }),
    ).toHaveValue("");
  });

  test(`password reset keeps recovery controls after resend errors (${theme})`, async ({
    page,
    context,
    baseURL,
  }, testInfo) => {
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    await page.addInitScript(
      (value) => localStorage.setItem("vocanova:theme-preference", value),
      theme,
    );
    let requests = 0;
    await page.route(
      "**/api/v1/auth/password/reset-requests",
      async (route) => {
        requests++;
        await route.fulfill({
          status: requests === 2 ? 503 : 202,
          contentType: "application/json",
          body: requests === 2 ? '{"error":"unavailable"}' : "{}",
        });
      },
    );
    await page.goto("/auth/password/reset");
    await page.clock.install();
    await page.getByLabel("Email address").fill("learner@example.test");
    await page
      .getByRole("button", { name: "Send password reset link", exact: true })
      .click();
    await expect(
      page.getByRole("status").filter({ hasText: /sent instructions/ }),
    ).toBeVisible();
    await expect(
      page.getByRole("button", { name: /Resend available in/ }),
    ).toBeDisabled();
    await page.clock.fastForward(60_000);
    await page
      .getByRole("button", { name: "Resend password reset link", exact: true })
      .click();
    await expect(
      page
        .getByRole("main")
        .getByRole("alert")
        .filter({ hasText: "temporarily unavailable" }),
    ).toBeVisible();
    await expect(
      page.getByRole("status").filter({ hasText: /sent instructions/ }),
    ).toBeVisible();
    await checkRecoveryUI(
      page,
      testInfo.outputPath(`reset-error-${theme}.png`),
      theme,
    );
    await expect(
      page
        .getByRole("main")
        .getByRole("alert")
        .filter({ hasText: "temporarily unavailable" }),
    ).toBeVisible();
    await page
      .getByRole("button", { name: "Use a different email", exact: true })
      .press("Enter");
    await expect(page.getByLabel("Email address")).toBeEditable();
    await expect(page.getByLabel("Email address")).toBeFocused();
  });

  test(`profile restores a supplementary-letter name after reauthentication (${theme})`, async ({
    page,
    context,
    baseURL,
  }) => {
    await authenticate(context, baseURL);
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    let expires = true;
    await page.route("**/api/v1/settings", async (route) => {
      if (route.request().method() === "PATCH" && expires) {
        expires = false;
        return route.fulfill({
          status: 401,
          contentType: "application/json",
          body: '{"error":"unauthorized"}',
        });
      }
      await route.continue();
    });
    await page.goto("/settings/profile");
    const name = page.getByRole("textbox", {
      name: "Display name",
      exact: true,
    });
    await expect(name).toHaveAccessibleDescription("Up to 80 characters.");
    const supplementaryName = "𐐀".repeat(80);
    await name.fill(supplementaryName);
    await expect(name).toHaveValue(supplementaryName);
    await page
      .getByRole("button", { name: "Save profile", exact: true })
      .click();
    await expect(page).toHaveURL(/\/login\?.*reason=session-expired/);
    await authenticate(context, baseURL);
    await page.goto("/settings/profile");
    await expect(name).toHaveValue(supplementaryName);
    await expect(page.getByRole("status")).toContainText("restored");
    await page
      .getByRole("button", { name: "Save profile", exact: true })
      .click();
    await expect(page.getByRole("status")).toContainText(
      "Your profile has been saved",
    );
    await expect
      .poll(() =>
        page.evaluate(
          (prefix) =>
            Object.keys(sessionStorage).filter((key) => key.startsWith(prefix))
              .length,
          PREFIX,
        ),
      )
      .toBe(0);
  });

  test(`email change restores only an unsent same-user address (${theme})`, async ({
    page,
    context,
    baseURL,
  }) => {
    await authenticate(context, baseURL);
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    let expires = true;
    await page.route("**/api/v1/settings/email-change-links", async (route) => {
      if (expires) {
        expires = false;
        return route.fulfill({
          status: 401,
          contentType: "application/json",
          body: '{"error":"unauthorized"}',
        });
      }
      await route.continue();
    });
    await page.goto("/settings/account");
    await page
      .getByRole("textbox", { name: "New sign-in email" })
      .fill("unsent@example.test");
    await page
      .getByRole("button", { name: "Send confirmation link", exact: true })
      .click();
    await expect(page).toHaveURL(/\/login\?.*reason=session-expired/);
    await authenticate(context, baseURL);
    await page.goto("/settings/account");
    await expect(
      page.getByRole("textbox", { name: "New sign-in email" }),
    ).toHaveValue("unsent@example.test");
    await expect(
      page.getByRole("status").filter({ hasText: "restored" }),
    ).toBeVisible();
    await page
      .getByRole("button", { name: "Send confirmation link", exact: true })
      .click();
    await expect(
      page.getByRole("status").filter({
        hasText: "If this address can be used, we've sent a confirmation link",
      }),
    ).toBeVisible();
    await expect(
      page
        .getByRole("status")
        .filter({ hasText: /^We sent a confirmation link/ }),
    ).toHaveCount(0);
    await expect
      .poll(() =>
        page.evaluate(
          (prefix) =>
            Object.keys(sessionStorage).filter((key) => key.startsWith(prefix))
              .length,
          PREFIX,
        ),
      )
      .toBe(0);
    await page.reload();
    await expect(
      page.getByRole("textbox", { name: "New sign-in email" }),
    ).toHaveValue("");
  });
}
