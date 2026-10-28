// The password field reserves room for its Show/Hide toggle: typed characters
// must never run underneath the 44px control at any supported width.

import { expect, test } from "@playwright/test";

for (const theme of ["light", "dark"] as const) {
  test(`password text stays clear of the visibility toggle in ${theme}`, async ({
    page,
    context,
  }, testInfo) => {
    await context.addCookies([
      {
        name: "vocanova_theme",
        value: theme,
        url: testInfo.project.use.baseURL!,
      },
    ]);
    await page.goto("/signin");
    const password = page.getByLabel("Password", { exact: true });
    await password.fill("a".repeat(64));
    for (const name of ["Show password", "Hide password"]) {
      const toggle = page.getByRole("button", { name, exact: true });
      const field = await password.boundingBox();
      const control = await toggle.boundingBox();
      expect(field).not.toBeNull();
      expect(control).not.toBeNull();
      expect(control!.height).toBeGreaterThanOrEqual(44);
      expect(control!.width).toBeGreaterThanOrEqual(44);
      const paddingRight = await password.evaluate((element) =>
        Number.parseFloat(getComputedStyle(element).paddingRight),
      );
      // The text area ends where the right padding begins.
      expect(field!.x + field!.width - paddingRight).toBeLessThanOrEqual(
        control!.x,
      );
      // The toggle sits inside the field, vertically centred on it.
      expect(control!.y).toBeGreaterThanOrEqual(field!.y);
      expect(control!.y + control!.height).toBeLessThanOrEqual(
        field!.y + field!.height,
      );
      await toggle.click();
    }
    await expect(password).toHaveAttribute("type", "password");
  });
}
