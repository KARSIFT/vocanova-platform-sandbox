import { expect, test, type Page } from "@playwright/test";

type FeedbackSample = {
  at: number;
  source: "mutation" | "frame";
  barActive: boolean;
  barVisible: boolean;
  loader: boolean;
  healthy: boolean;
};
type FeedbackObservation = {
  samples: FeedbackSample[];
  gaps: FeedbackSample[];
  prematureBarClears: FeedbackSample[];
  firstFeedback: number | null;
  firstHealthy: number | null;
};
type FeedbackWindow = Window & {
  __navigationFeedbackAudit?: FeedbackObservation & { stop: () => void };
};

/** Real client navigation may retain the prior screen instead of committing a
 * route skeleton. Require continuous visible feedback through either path,
 * including the handoff when the global bar clears on a pathname commit. */
export async function expectNavigationFeedback(
  page: Page,
  { loadingStatusId, heading, healthyTimeout = 10_000 }: {
    loadingStatusId: string;
    heading: string;
    healthyTimeout?: number;
  },
  activate: () => Promise<unknown>,
) {
  await page.evaluate(({ loadingStatusId, heading }) => {
    const browser = window as FeedbackWindow;
    browser.__navigationFeedbackAudit?.stop();
    const observation: FeedbackObservation = {
      samples: [], gaps: [], prematureBarClears: [],
      firstFeedback: null, firstHealthy: null,
    };
    let started: number | null = null;
    let previous: FeedbackSample | undefined;
    let previousFrame: FeedbackSample | undefined;
    let frame = 0;
    let stopped = false;
    const visible = (element: Element | null): boolean => {
      if (!element || element.closest("[hidden]")) return false;
      const style = getComputedStyle(element);
      const box = element.getBoundingClientRect();
      return style.display !== "none" && style.visibility !== "hidden"
        && Number(style.opacity) > 0 && box.width > 0 && box.height > 0;
    };
    // Start at the real user activation, excluding locator focus/actionability
    // and host IPC time from the acknowledgement budget.
    const activate = (event: Event) => {
      if (started !== null || !event.isTrusted) return;
      if (event instanceof KeyboardEvent && event.key !== "Enter") return;
      started = performance.now();
    };
    document.addEventListener("click", activate, true);
    document.addEventListener("keydown", activate, true);
    const sample = (source: FeedbackSample["source"]) => {
      if (stopped || started === null) return;
      const bar = document.querySelector(".navigation-progress");
      const section = document.getElementById(loadingStatusId)?.closest("section") ?? null;
      const state: FeedbackSample = {
        at: Math.round(performance.now() - started),
        source,
        barActive: bar?.getAttribute("data-active") === "true",
        barVisible: bar?.getAttribute("data-active") === "true" && visible(bar),
        loader: section?.getAttribute("aria-busy") === "true" && visible(section),
        healthy: [...document.querySelectorAll("main h1")].some(element =>
          element.textContent?.replace(/\s+/g, " ").trim() === heading && visible(element)),
      };
      // DOM mutations explain transitions; only paint-frame observations can
      // establish a learner-visible gap, not an intermediate React commit.
      if (source === "frame") {
        if ((state.barVisible || state.loader) && observation.firstFeedback === null)
          observation.firstFeedback = state.at;
        if (state.healthy && observation.firstHealthy === null)
          observation.firstHealthy = state.at;
        if (observation.firstFeedback !== null && !state.healthy && !state.barVisible && !state.loader)
          observation.gaps.push(state);
        if (previousFrame?.barActive && !state.barActive && !state.loader && !state.healthy)
          observation.prematureBarClears.push(state);
        previousFrame = state;
      }
      if (!previous || ["source", "barActive", "barVisible", "loader", "healthy"].some(key =>
        state[key as keyof FeedbackSample] !== previous![key as keyof FeedbackSample]))
        observation.samples.push(state);
      previous = state;
    };
    const observer = new MutationObserver(() => sample("mutation"));
    observer.observe(document.documentElement, { childList: true, subtree: true, attributes: true });
    const tick = () => { sample("frame"); if (!stopped) frame = requestAnimationFrame(tick); };
    browser.__navigationFeedbackAudit = {
      ...observation,
      // Preserve the mutable scalar fields as well as the collected arrays.
      get firstFeedback() { return observation.firstFeedback; },
      get firstHealthy() { return observation.firstHealthy; },
      stop() {
        sample("mutation"); stopped = true; observer.disconnect(); cancelAnimationFrame(frame);
        document.removeEventListener("click", activate, true);
        document.removeEventListener("keydown", activate, true);
      },
    };
    frame = requestAnimationFrame(tick);
  }, { loadingStatusId, heading });

  let result: FeedbackObservation | undefined;
  try {
    await activate();
    await expect.poll(() => page.evaluate(() =>
      (window as FeedbackWindow).__navigationFeedbackAudit!.firstFeedback),
    { timeout: 500, message: "Navigation must visibly acknowledge the action before the destination arrives" })
      .not.toBeNull();
    await expect(page.getByRole("main").getByRole("heading", { name: heading, level: 1, exact: true }))
      .toBeVisible({ timeout: healthyTimeout });
    await expect(page.locator(".navigation-progress")).not.toHaveAttribute("data-active", "true");
    await page.evaluate(() => new Promise<void>(resolve => requestAnimationFrame(() => resolve())));
  } finally {
    result = await page.evaluate(() => {
      const audit = (window as FeedbackWindow).__navigationFeedbackAudit!;
      audit.stop();
      return {
        samples: audit.samples, gaps: audit.gaps, prematureBarClears: audit.prematureBarClears,
        firstFeedback: audit.firstFeedback, firstHealthy: audit.firstHealthy,
      };
    });
    await test.info().attach("navigation-feedback-lifecycle", {
      body: JSON.stringify({ loadingStatusId, heading, ...result }, null, 2), contentType: "application/json",
    });
  }
  expect(result.firstFeedback, "Visible feedback must precede healthy destination content").not.toBeNull();
  expect(result.firstFeedback!).toBeLessThanOrEqual(500);
  expect(result.firstFeedback!).toBeLessThan(result.firstHealthy!);
  expect(result.gaps, "Feedback disappeared before the destination became usable").toEqual([]);
  expect(result.prematureBarClears, "The bar may clear only when a loading shell or healthy destination is visible").toEqual([]);
}
