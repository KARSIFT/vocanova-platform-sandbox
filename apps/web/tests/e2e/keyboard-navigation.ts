import { expect, type Locator, type Page } from "@playwright/test";

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

export async function expectKeyboardTargetClear(
  page: Page,
  target: Locator,
  { backwardsFirst = false }: { backwardsFirst?: boolean } = {},
) {
  await target.focus();
  await pressTab(page, { backwards: backwardsFirst });
  await pressTab(page, { backwards: !backwardsFirst });
  await expect(target).toBeFocused();
  await expect.poll(() => target.evaluate((element) => {
    const rect = element.getBoundingClientRect();
    const header = document.querySelector("header")?.getBoundingClientRect();
    const bottomNav = Array.from(document.querySelectorAll("nav")).find((nav) =>
      getComputedStyle(nav).position === "fixed" && nav.getBoundingClientRect().height > 0 &&
      nav.getBoundingClientRect().bottom >= innerHeight - 1)?.getBoundingClientRect();
    const hit = document.elementFromPoint(rect.x + rect.width / 2, rect.y + rect.height / 2);
    return rect.top >= Math.max(0, header?.bottom ?? 0) && rect.bottom <= (bottomNav?.top ?? innerHeight) &&
      rect.width >= 44 && rect.height >= 44 && !!hit && element.contains(hit);
  })).toBe(true);
}
