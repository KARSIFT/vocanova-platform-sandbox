// VOC-031-T08 - Full core-loop end-to-end Playwright suite.
//
// This is the DOC-10 §7-documented end-to-end flow, built for
// the first time on the T07 Playwright install. The flow is the
// the functional-correctness counterpart to the T07a/T07b
// accessibility scans: every step is exercised against real
// client components, real network calls, and the real Next.js
// auth-gate middleware, with the mock API server (see
// mock-api-server.mjs) standing in for the Go/Postgres backend
// the same way it does for the T07 scans.
//
// Flow (one Playwright test, one browser context):
//
//   1.  Set the session + CSRF cookies so the learner is
//       authenticated without going through the magic-link UI
//       (the magic-link UI is a separate A1 component surface;
//       T08's "auth" step is "be authenticated", satisfied by
//       the session cookie). Set e2e_onboarding_status =
//       not_started so the auth-gate middleware funnels the
//       learner into /onboarding.
//
//   2.  Onboarding: fill the 5-step form, submit, expect
//       /onboarding -> /home.
//
//   3.  Discover: /home -> /discover -> /discover/ordering-at-a-
//       cafe -> /discover/ordering-at-a-cafe/pour.
//
//   4.  Save: click the Save button on the Word Detail screen,
//       wait for the button to read "Saved" (the T07b
//       accessibility scan already covered the static
//       keyboard/color-only surfaces; T08 covers the actual
//       save -> saved transition).
//
//   5.  Review: /reviews, complete the one due card via the
//       self-check flow (Show answer -> rate), expect the "all
//       caught up" empty state plus a sentence-feedback widget
//       for the just-reviewed card.
//
//   6.  Sentence feedback + deterministic AI: back on the Word
//       Detail screen, type a sentence containing the target
//       word, click Check, wait for the deterministic mock
//       feedback. The mock returns "correct" without calling a
//       paid/nondeterministic provider (DOC-10 §7).
//
//   7.  Progress update: /home -> verify the daily-mission
//       counter now shows the review counted.
//
//   8.  Settings change: /settings, change the display name,
//       click Save settings, verify the "saved" confirmation.
//
//   9.  Logout: click the AppHeader "Log out" button, expect
//       navigation to /signin.
//
//   10. Unauthenticated-access rejection: set
//       e2e_unauthenticated=1 (the same kind of test-only cookie
//       the T07b onboarding scan uses, see mock-api-server.mjs),
//       navigate to /home, expect the auth-gate middleware to
//       redirect to /signin.
//
// Project scope: T08 follows the T07a "one representative
// desktop width" pattern - mobile-360 / mobile-430 are
// accessibility-only, exercised by T07b. The functional flow
// runs on home-desktop-1280 only. The test self-skips on the
// mobile projects so adding the mobile projects to the
// Playwright config (T07b) does not silently expand T08's
// scope. Running the full functional flow on three projects
// would triple test time without coverage gain beyond what
// T07b's mobile accessibility scans already provide.

import { randomUUID } from "node:crypto";

import { expect, test, type Page } from "@playwright/test";

const ONBOARDING_COOKIE_VALUE = "not_started";
const CORE_LOOP_TEST_TIMEOUT_MS = 90_000;
const SENTENCE_PRIVACY_REMINDER =
  "For your privacy, do not include personal information such as phone numbers, addresses, or passwords.";

async function expectSentencePracticePrivacyReminder(page: Page) {
  await expect(page.getByText(SENTENCE_PRIVACY_REMINDER)).toBeVisible();
}

test.describe("Core loop end-to-end (VOC-031-T08)", () => {
  test("auth -> onboarding -> discover -> save -> review -> sentence -> AI feedback -> progress -> settings -> logout -> rejection", async ({
    page,
    context,
  }, testInfo) => {
    test.skip(
      testInfo.project.name !== "home-desktop-1280",
      "T08 scope is one representative desktop width >=1024px (mirrors T07a); mobile projects are T07b's accessibility scope.",
    );
    test.setTimeout(CORE_LOOP_TEST_TIMEOUT_MS);

    // ----- 1. Auth: set the session + CSRF + onboarding cookies.
    //
    // The session cookie value is unique per test run so this
    // test's in-memory mock state does not collide with the
    // T07a/T07b scans, which use the mock's default session
    // (no vocanova_session cookie). The CSRF cookie is set to
    // a value the test will echo as the X-CSRF-Token header on
    // every mutation; the mock's CSRF check requires the two
    // to match.
    const baseURL = testInfo.project.use.baseURL;
    if (!baseURL) {
      throw new Error(
        "Expected the Playwright project to configure use.baseURL so the test can scope its session + CSRF cookies to it.",
      );
    }
    const sessionValue = `test-session-${randomUUID()}`;
    const csrfValue = `test-csrf-${randomUUID()}`;
    await context.addCookies([
      { name: "vocanova_session", value: sessionValue, url: baseURL },
      { name: "vocanova_csrf", value: csrfValue, url: baseURL },
      {
        name: "e2e_onboarding_status",
        value: ONBOARDING_COOKIE_VALUE,
        url: baseURL,
      },
    ]);

    // ----- 2. Onboarding.
    //
    // /home is gated by the middleware, which fetches /api/v1/me
    // with the e2e_onboarding_status=not_started cookie, gets
    // onboardingStatus=not_started, and redirects to /onboarding.
    await page.goto("/home");
    await expect(page).toHaveURL(/\/onboarding(\?|$)/);
    await expect(
      page.getByRole("heading", { name: "Welcome to Vocanova", level: 1 }),
    ).toBeVisible();

    // Step 0: English level. Native radio-group semantics mean
    // Tab moves focus into the group once; the option labels
    // are reached with arrow keys, not Tab. The Continue button
    // is disabled until an option is selected.
    await page.getByRole("radio", { name: /A2 \u2014 Elementary/ }).check();
    await page.getByRole("button", { name: "Continue" }).click();

    // Step 1: native language. Free text. The step uses an
    // <h2> (not a fieldset legend) so a heading role match is
    // the right signal that the next step has rendered.
    await expect(
      page.getByRole("heading", { name: /What's your native language\?/ }),
    ).toBeVisible();
    await page.getByRole("textbox", { name: "Native language" }).fill("es");
    await page.getByRole("button", { name: "Continue" }).click();

    // Step 2: learning goal. Radio group rendered inside a
    // <fieldset>/<legend> (not an <h2>), so waiting for the
    // radio group to be visible is the right transition signal.
    await expect(
      page.getByRole("radiogroup", {
        name: /What's your main reason for learning\?/,
      }),
    ).toBeVisible();
    await page.getByRole("radio", { name: "General growth" }).check();
    await page.getByRole("button", { name: "Continue" }).click();

    // Step 3: main use case. Radio group inside a fieldset.
    await expect(
      page.getByRole("radiogroup", {
        name: /Where will you use English most\?/,
      }),
    ).toBeVisible();
    await page.getByRole("radio", { name: "Daily life" }).check();
    await page.getByRole("button", { name: "Continue" }).click();

    // Step 4: daily review target. The step has an <h2>
    // "Daily review target" (visible to the test) wrapping the
    // radio group, so either is a valid wait target.
    await expect(
      page.getByRole("heading", { name: "Daily review target" }),
    ).toBeVisible();
    // The underlying <input> is visually hidden (sr-only) behind a styled
    // <label> - checking it directly (even with force: true) reached the
    // DOM node but the click never propagated to a real POST /api/v1/
    // onboarding request (waitForResponse below then hung the full 90s).
    // Clicking the visible label text instead is a real user interaction -
    // the browser's native label-for-input association activates the
    // radio and fires onChange normally, same as a real click would.
    await page
      .getByRole("radiogroup", { name: "Daily review target" })
      .getByText("15", { exact: true })
      .click();
    const onboardingResponsePromise = page.waitForResponse(
      (response) =>
        response.url().includes("/api/v1/onboarding") &&
        response.request().method() === "POST",
    );
    await page.getByRole("button", { name: "Finish setup" }).click();
    const onboardingResponse = await onboardingResponsePromise;
    expect(onboardingResponse.status()).toBe(200);

    // The mock's /api/v1/onboarding endpoint honors the
    // e2e_onboarding_status cookie as an unconditional override
    // (the T07b scan depends on that behavior to land on the
    // onboarding form). The override would otherwise keep
    // /me returning onboardingStatus=not_started forever, and
    // the form's post-submit router.push("/home") would loop
    // /home -> /onboarding. Clear ONLY the override cookie
    // (the session + CSRF cookies must stay so the test's
    // per-session mock state and CSRF-protected mutations
    // continue to work). Then navigate to /home explicitly.
    await context.clearCookies({ name: "e2e_onboarding_status" });
    await page.goto("/home");
    await expect(page).toHaveURL(/\/home(\?|$)/);
    await expect(
      page.getByRole("heading", { name: "Today's Mission", level: 1 }),
    ).toBeVisible();

    // ----- 3. Discover.
    await page.goto("/discover");
    await expect(
      page.getByRole("heading", { name: "Journey", level: 1 }),
    ).toBeVisible();
    await page.getByRole("link", { name: /Ordering at a cafe/ }).click();
    await expect(page).toHaveURL(/\/discover\/ordering-at-a-cafe(\?|$)/);
    await expect(
      page.getByRole("heading", { name: "Ordering at a cafe", level: 1 }),
    ).toBeVisible();
    await page.getByRole("link", { name: /pour/ }).first().click();
    await expect(page).toHaveURL(/\/discover\/ordering-at-a-cafe\/pour(\?|$)/);
    await expect(
      page.getByRole("heading", { name: "pour", level: 1 }),
    ).toBeVisible();

    // PRD §2 MVP completion criteria requires the Word Detail screen to show
    // meaning, part of speech, examples, and usage notes (see
    // docs/product/01-mvp-prd.md §2, issue #1180). Assert each field's real
    // text - sourced from the mock's CANONICAL_WORDS.pour fixture, which
    // stands in for the Go content API - is actually on the page, so a
    // regression that silently drops one of these fields from the DTO,
    // client, or component fails this test instead of shipping unnoticed.
    await expect(page.getByText("verb", { exact: true })).toBeVisible();
    await expect(
      page.getByText("to make liquid flow into a container"),
    ).toBeVisible();
    await expect(
      page.getByText("Could you pour me a cup of coffee?"),
    ).toBeVisible();
    await expect(
      page.getByText("Common in everyday service contexts."),
    ).toBeVisible();

    // ----- 4. Save.
    //
    // The Save button's aria-label changes on toggle: "Save pour:
    // to make liquid flow into a container" -> "Remove pour from
    // saved words". The visible text toggles between "Save" and
    // "Saved" (see meaning-save-button.tsx). The mock's POST
    // /api/v1/user-words records the save.
    const saveButton = page.getByRole("button", {
      name: /^Save pour: to make liquid flow into a container$/,
    });
    await saveButton.click();
    await expect(
      page.getByRole("button", { name: "Remove pour from saved words" }),
    ).toBeVisible();
    await expect(page.getByText("Due today", { exact: true })).toBeVisible();

    // Issue #1181 (PRD §2): sentence practice is "surfaced from Home,
    // Word Detail, and Review Completion" - not a fourth tab. This is
    // entry point 1 of 3: the SentenceFeedback widget below a saved
    // meaning on the Word Detail screen. It's gated on the server-
    // rendered `meaning.saved`/`meaning.userWordId` fields (see
    // discover/[situation]/[word]/page.tsx), which meaning-save-button.tsx
    // now refreshes via router.refresh() right after a successful save -
    // without that refresh the button flips to "Saved" but this widget
    // never appears until a manual reload, silently breaking this entry
    // point. Assert it renders immediately, with no navigation in between.
    await expect(
      page.getByRole("heading", { name: /Practice with pour/ }),
    ).toBeVisible();
    await expectSentencePracticePrivacyReminder(page);

    // ----- 5. Review session.
    //
    // The mock returns the one saved word as a due word. With a
    // single due word, the review session uses the self_check
    // prompt type (no distractors to build a multiple-choice
    // question). The flow is: Show answer -> rate -> advance ->
    // refetch returns empty -> "all caught up" state.
    await page.goto("/reviews");
    await expect(
      page.getByRole("heading", { name: "Review", level: 1 }),
    ).toBeVisible();
    await expect(
      page.getByRole("heading", { name: "pour", level: 2 }),
    ).toBeVisible();
    await page.getByRole("button", { name: "Show answer" }).click();
    await expect(page.getByRole("button", { name: "Good" })).toBeVisible();
    await page.getByRole("button", { name: "Good" }).click();

    // "All caught up" empty state with a sentence-feedback
    // widget for the just-reviewed card. Issue #1181 (PRD §2),
    // entry point 3 of 3: SentenceFeedback surfaced from Review
    // Completion (see review-session.tsx's `completed` branch).
    await expect(
      page.getByRole("heading", { name: "You're all caught up", level: 2 }),
    ).toBeVisible();
    await expect(
      page.getByRole("heading", { name: /Practice with pour/ }),
    ).toBeVisible();
    await expectSentencePracticePrivacyReminder(page);

    // ----- 6. Sentence feedback + deterministic AI.
    //
    // The mock's evaluateSentenceFeedback is the deterministic
    // AI adapter DOC-10 §7 requires for CI: the same sentence
    // always produces the same feedback. A sentence containing
    // the target word "pour" and >= 3 words returns
    // { status: "correct", explanation: <fixed> }.
    await page
      .getByRole("textbox", { name: /Write a sentence using pour/ })
      .fill("I will pour the coffee into a cup.");
    await page.getByRole("button", { name: "Check my sentence" }).click();
    await expect(page.getByText("Correct", { exact: true })).toBeVisible();
    await expect(
      page.getByText("Your sentence uses the target word naturally."),
    ).toBeVisible();
    await expect(page.getByText(/Mission completed: Yes/)).toBeVisible();

    // A fresh result explicitly has reported=false and remains reportable.
    // A failed report preserves the feedback result and keeps the action
    // retryable; reporting never replaces the feedback under DOC-09 §16.
    await expect(
      page.getByRole("button", { name: "Report a problem" }),
    ).toBeVisible();
    const reportEndpoint = "**/api/v1/sentence-feedback/*/reports";
    await page.route(reportEndpoint, async (route) => {
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ error: "temporary_failure" }),
      });
    });
    await page.getByRole("button", { name: "Report a problem" }).click();
    await page.getByRole("button", { name: "Already correct" }).click();
    await expect(page.getByText("Unable to report. Try again.")).toBeVisible();
    await expect(page.getByText("Correct", { exact: true })).toBeVisible();
    await expect(
      page.getByText("Your sentence uses the target word naturally."),
    ).toBeVisible();
    await expect(page.getByText(/Mission completed: Yes/)).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Already correct" }),
    ).toBeVisible();

    await page.unroute(reportEndpoint);
    await page.getByRole("button", { name: "Already correct" }).click();
    await expect(page.getByText("Reported", { exact: true })).toBeVisible();

    // A deduplicated response represents feedback reported by an earlier
    // submission. It must hydrate that persisted state rather than offer the
    // five report reasons again.
    await page.getByRole("button", { name: "Check my sentence" }).click();
    await expect(page.getByText("Reported", { exact: true })).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Report a problem" }),
    ).toHaveCount(0);
    await expect(
      page.getByRole("button", { name: "Correction changed my meaning" }),
    ).toHaveCount(0);

    // ----- 7. Progress update.
    //
    // The mock's POST /api/v1/reviews/submissions increments
    // state.reviewedCount; the home page's daily-mission text
    // reflects that. After one review, the home page should
    // show "1 of 20 words reviewed today" (the default target).
    await page.goto("/home");
    await expect(page.getByText(/1 of 20 words reviewed today/)).toBeVisible();

    // Issue #1181 (PRD §2), entry point 2 of 3: the Home screen's
    // "Saved words" list renders a SentenceFeedback widget per saved
    // word (see home/page.tsx). "pour" was saved in step 4, so it
    // should appear here too, independent of the Word Detail and
    // Review Completion widgets already exercised above.
    await expect(
      page.getByRole("heading", { name: /Practice with pour/ }),
    ).toBeVisible();
    await expectSentencePracticePrivacyReminder(page);

    // ----- 8. Settings change.
    await page.goto("/settings");
    await expect(
      page.getByRole("heading", { name: "Settings", level: 1 }),
    ).toBeVisible();
    const displayNameInput = page.getByRole("textbox", {
      name: "Display name",
    });
    await displayNameInput.fill("Core Loop Fixture Updated");
    await page.getByRole("button", { name: "Save settings" }).click();
    await expect(
      page.getByText("Your settings have been saved."),
    ).toBeVisible();
    // PATCH /api/v1/settings updated the per-session state;
    // the form re-reads the response and the input now shows
    // the new value.
    await expect(displayNameInput).toHaveValue("Core Loop Fixture Updated");

    // ----- 9. Logout.
    //
    // The AppHeader's Log out button calls POST /api/v1/auth/logout
    // with the CSRF token, clears the local CSRF cookie, and
    // navigates to /signin.
    await page.getByRole("button", { name: "Log out" }).click();
    await expect(page).toHaveURL(/\/signin(\?|$)/);
    await expect(
      page.getByRole("heading", { name: "Sign in to Vocanova" }),
    ).toBeVisible();

    // ----- 10. Unauthenticated-access rejection.
    //
    // The mock returns 401 for /api/v1/me when the
    // e2e_unauthenticated=1 cookie is set (see
    // mock-api-server.mjs). The middleware uses that 401 to
    // redirect to /signin. The test sets the cookie, navigates
    // to /home, and expects the redirect.
    await context.addCookies([
      { name: "e2e_unauthenticated", value: "1", url: baseURL },
    ]);
    await page.goto("/home");
    await expect(page).toHaveURL(/\/signin\?returnTo=%2Fhome(\b|$)/);
  });
});
