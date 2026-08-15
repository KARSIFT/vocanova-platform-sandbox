// VOC-084-T01 sign-in OAuth capability gating.
//
// The sign-in page reads /healthz kill_switches.oauth_enabled during SSR
// and hides "Continue with Google" when OAuth is disabled. The mock API
// defaults MOCK_OAUTH_ENABLED=false so this suite matches staging posture.

import { expect, test } from "@playwright/test";

test.describe("Sign-in OAuth capability (VOC-084-T01)", () => {
  test("hides Continue with Google when oauth_enabled is false", async ({
    page,
  }) => {
    await page.goto("/signin");

    await expect(
      page.getByRole("heading", { name: "Sign in to Vocanova", level: 1 }),
    ).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Continue with Google" }),
    ).toHaveCount(0);
    await expect(page.getByLabel("Email address")).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Send sign-in link" }),
    ).toBeVisible();
  });
});
