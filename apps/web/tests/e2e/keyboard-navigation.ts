import type { Page } from "@playwright/test";

/**
 * Use the engine's native sequential keyboard navigation, including links.
 * macOS WebKit uses Option-Tab; changing DOM focus would conceal tab-order bugs.
 * https://support.apple.com/guide/safari/keyboard-shortcuts-and-gestures-cpsh003/mac
 * Verified with the pinned WebKit on native input/button/link/checkbox/select
 * controls in both directions; no host preferences or launch defaults change.
 */
export async function pressTab(
  page: Page,
  { backwards = false }: { backwards?: boolean } = {},
): Promise<void> {
  const macWebKit =
    process.platform === "darwin" &&
    page.context().browser()?.browserType().name() === "webkit";
  const modifiers = [macWebKit ? "Alt" : "", backwards ? "Shift" : ""];
  await page.keyboard.press([...modifiers.filter(Boolean), "Tab"].join("+"));
}
