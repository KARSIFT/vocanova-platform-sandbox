import {
  expect,
  type Locator,
  type Page,
  type Request,
} from "@playwright/test";

const SITUATION_PATH = "/discover/daily-conversation";
const PRACTICE_NAME = "Choose the word for the situation";

// Independent acceptance expectations: do not derive answers or destinations
// from the product's activity factory or the response under test.
const WORDS = [
  "greeting",
  "small talk",
  "casual",
  "weekend plans",
  "available",
  "invite",
  "join",
  "suggest",
  "sounds good",
  "arrange",
  "meet up",
  "confirm",
  "on time",
  "reschedule",
  "cancel",
  "catch up",
  "keep in touch",
  "farewell",
];

const EXAMPLES = [
  {
    title: "Ask a friend to dinner",
    context: "You are planning dinner at home. You want Sam to come.",
    prompt: "Complete your message: “I'd like to ___ you to dinner.”",
    choices: ["invite", "join"],
    correct: "invite",
    wrong: "join",
    explanations: [
      "Invite is what you do when you ask Sam to come.",
      "Join means taking part with other people. Sam joins you if he takes part.",
    ],
    href: `${SITUATION_PATH}/invite`,
    canonicalExample: "I want to invite you to dinner on Friday.",
  },
  {
    title: "Respond to an idea",
    context: "Sam says, “How about lunch on Saturday?” You like the idea.",
    prompt: "Which reply fits?",
    choices: ["keep in touch", "sounds good"],
    correct: "sounds good",
    wrong: "keep in touch",
    explanations: [
      "Keep in touch means continuing to contact someone over time. It does not say that you like this lunch suggestion.",
      "Sounds good is a friendly way to say you like the idea. You may still need to agree on the time and place.",
    ],
    href: `${SITUATION_PATH}/sounds-good`,
    canonicalExample:
      "A picnic sounds good, but let's check the weather first.",
  },
  {
    title: "Move lunch to another day",
    context:
      "Tuesday is no longer possible. You still want lunch with Sam on Friday.",
    prompt: "Complete your message: “Could we ___ lunch for Friday?”",
    choices: ["reschedule", "cancel"],
    correct: "reschedule",
    wrong: "cancel",
    explanations: [
      "Reschedule means moving a planned event to a different time or day. You want to keep the lunch plan and move it to Friday.",
      "Cancel means stopping a plan from going ahead. It does not by itself say that the event has a new date.",
    ],
    href: `${SITUATION_PATH}/reschedule`,
    canonicalExample: "Could we reschedule our lunch for Friday?",
  },
];

async function tabTo(page: Page, target: Locator): Promise<void> {
  // The starting focus varies between a fresh document and client navigation.
  // Bound the real Tab walk; do not repair an unreachable control with focus().
  for (let attempt = 0; attempt < 40; attempt++) {
    if (
      await target.evaluate((element) => element === document.activeElement)
    ) {
      return;
    }
    await page.keyboard.press("Tab");
  }
  await expect(target).toBeFocused();
}

/**
 * The same non-mutating phase runs with local fixtures and the existing minted
 * staging session. Authentication, backend selection and logout belong to the
 * caller. No saved-word state is assumed or reset here.
 */
export async function verifyContextPracticeJourney(page: Page): Promise<void> {
  let apiMutationCount = 0;
  const context = page.context();
  const countMutation = (request: Request) => {
    if (
      new URL(request.url()).pathname.startsWith("/api/v1/") &&
      !["GET", "HEAD", "OPTIONS"].includes(request.method())
    ) {
      apiMutationCount++;
    }
  };
  // Retain only a context-wide count, never request bodies, headers or identifiers.
  context.on("request", countMutation);

  try {
    await page.goto("/discover");
    await expect(
      page.getByRole("heading", { level: 1, name: "Journey", exact: true }),
    ).toBeVisible();
    await page.getByRole("link", { name: /^Daily Conversation\b/ }).click();
    await expect(page).toHaveURL(new RegExp(`${SITUATION_PATH}$`));
    await expect(
      page.getByRole("heading", {
        level: 1,
        name: "Daily Conversation",
        exact: true,
      }),
    ).toBeVisible();
    const words = page.getByRole("list", { name: "Words in this situation" });
    await expect(words.getByRole("heading", { level: 2 })).toHaveText(WORDS);
    const savedProgress = page
      .getByRole("region", { name: "Situation progress" })
      .getByText(/^\d+ of 18 words saved$/);
    await expect(savedProgress).toBeVisible();
    const originalSavedProgress = await savedProgress.innerText();
    const practice = page.getByRole("region", { name: PRACTICE_NAME });
    await expect(practice).toBeVisible();
    const start = practice.getByRole("button", {
      name: "Start context practice",
    });
    await tabTo(page, start);
    await page.keyboard.press("Enter");

    for (const [index, example] of EXAMPLES.entries()) {
      await expect(
        practice.getByRole("heading", { name: example.title, exact: true }),
      ).toBeFocused();
      await expect(practice).toContainText(`Example ${index + 1} of 3`);
      await expect(
        practice.getByText(example.context, { exact: true }),
      ).toBeVisible();
      await expect(
        practice.getByText(example.prompt, { exact: true }),
      ).toBeVisible();
      const choices = practice
        .getByRole("group", { name: example.prompt, exact: true })
        .getByRole("button");
      await expect(choices).toHaveText(example.choices);
      for (const explanation of example.explanations) {
        await expect(practice).not.toContainText(explanation);
      }
      await page.keyboard.press("Tab");
      await expect(choices.first()).toBeFocused();
      if (example.choices[0] !== example.wrong) {
        await page.keyboard.press("Tab");
      }
      await expect(
        practice.getByRole("button", { name: example.wrong, exact: true }),
      ).toBeFocused();
      await page.keyboard.press("Enter");
      const feedback = practice.getByRole("status");
      await expect(feedback).toContainText(
        "Not quite. Compare the two choices.",
      );
      for (const explanation of example.explanations) {
        await expect(
          feedback.getByText(explanation, { exact: true }),
        ).toBeVisible();
      }
      await expect(choices.first()).toBeDisabled();
      await expect(choices.last()).toBeDisabled();
      await expect(
        practice.getByRole("button", {
          name: /^(Next example|Finish practice)$/,
        }),
      ).toHaveCount(0);

      // Answer buttons are now disabled: the next native Tab must reach retry.
      await page.keyboard.press("Tab");
      await expect(
        practice.getByRole("button", { name: "Try again" }),
      ).toBeFocused();
      await page.keyboard.press("Enter");
      await expect(choices.first()).toBeFocused();
      for (const explanation of example.explanations) {
        await expect(practice).not.toContainText(explanation);
      }
      if (example.choices[0] !== example.correct) {
        await page.keyboard.press("Tab");
      }
      await expect(
        practice.getByRole("button", { name: example.correct, exact: true }),
      ).toBeFocused();
      await page.keyboard.press("Enter");
      await expect(feedback).toContainText("That fits this situation.");
      for (const explanation of example.explanations) {
        await expect(
          feedback.getByText(explanation, { exact: true }),
        ).toBeVisible();
      }
      await page.keyboard.press("Tab");
      await expect(
        practice.getByRole("button", {
          name:
            index === EXAMPLES.length - 1 ? "Finish practice" : "Next example",
        }),
      ).toBeFocused();
      await page.keyboard.press("Enter");
    }

    await expect(
      practice.getByRole("heading", {
        name: "You’ve explored three situations",
      }),
    ).toBeFocused();
    const destinations: string[] = [];
    for (const example of EXAMPLES) {
      const link = practice.getByRole("link", {
        name: `Practice with ${example.correct}`,
        exact: true,
      });
      await expect(link).toHaveAttribute("href", example.href);
      destinations.push((await link.getAttribute("href"))!);
    }

    await tabTo(
      page,
      practice.getByRole("button", { name: "Restart practice" }),
    );
    await page.keyboard.press("Enter");
    await expect(
      practice.getByRole("heading", { name: EXAMPLES[0]!.title, exact: true }),
    ).toBeFocused();
    await expect(practice).toContainText("Example 1 of 3");
    await expect(practice.getByRole("status")).toBeEmpty();
    // Prompt -> first choice -> second choice -> Exit to words.
    await page.keyboard.press("Tab");
    await page.keyboard.press("Tab");
    await page.keyboard.press("Tab");
    await expect(
      practice.getByRole("button", { name: "Exit to words" }),
    ).toBeFocused();
    await page.keyboard.press("Enter");
    await expect(words).toBeFocused();
    await expect(start).toBeVisible();
    await expect(savedProgress).toHaveText(originalSavedProgress);

    await page.keyboard.press("Shift+Tab");
    await expect(start).toBeFocused();
    await page.keyboard.press("Enter");
    await page.keyboard.press("Tab");
    await expect(
      practice.getByRole("button", { name: "invite", exact: true }),
    ).toBeFocused();
    await page.keyboard.press("Enter");
    await page.keyboard.press("Tab");
    await expect(
      practice.getByRole("button", { name: "Next example" }),
    ).toBeFocused();
    await page.keyboard.press("Enter");
    await expect(
      practice.getByRole("heading", { name: EXAMPLES[1]!.title, exact: true }),
    ).toBeFocused();
    await page.reload();
    await expect(start).toBeVisible();
    await expect(
      practice.getByRole("heading", { name: EXAMPLES[1]!.title, exact: true }),
    ).toHaveCount(0);
    // A concurrent synthetic journey can legitimately change the server's
    // saved count. Only the mounted activity above must preserve its baseline.
    await expect(savedProgress).toBeVisible();

    // Complete again after checking recovery so an ordinary completion-link
    // click can leave this page. Avoid browser-specific modifier/new-tab behavior.
    await start.click();
    for (const [index, example] of EXAMPLES.entries()) {
      await expect(
        practice.getByRole("heading", { name: example.title, exact: true }),
      ).toBeFocused();
      await expect(
        practice.getByText(example.prompt, { exact: true }),
      ).toBeVisible();
      const feedback = practice.getByRole("status");
      await expect(feedback).toBeEmpty();
      await practice
        .getByRole("button", { name: example.correct, exact: true })
        .click();
      await expect(feedback).toContainText("That fits this situation.");
      const next = practice.getByRole("button", {
        name:
          index === EXAMPLES.length - 1 ? "Finish practice" : "Next example",
      });
      await expect(next).toBeEnabled();
      await next.click();
    }
    await expect(
      practice.getByRole("heading", {
        name: "You’ve explored three situations",
      }),
    ).toBeFocused();
    const inviteLink = practice.getByRole("link", {
      name: "Practice with invite",
      exact: true,
    });
    await expect(inviteLink).toHaveAttribute("href", destinations[0]!);
    await inviteLink.click();
    for (const [index, example] of EXAMPLES.entries()) {
      if (index > 0) await page.goto(destinations[index]!);
      await expect(page).toHaveURL(new RegExp(`${example.href}$`));
      const main = page.getByRole("main");
      await expect(
        main.getByRole("heading", {
          level: 1,
          name: example.correct,
          exact: true,
        }),
      ).toBeVisible();
      // Server streaming can leave hidden copies outside or within main.
      // Ignore those copies, but reject duplicate examples visible anywhere.
      const moreExamples = main
        .locator("summary")
        .filter({ hasText: "More examples" });
      if (await moreExamples.count()) {
        await expect(moreExamples.locator("..")).not.toHaveAttribute("open");
        await moreExamples.focus();
        await page.keyboard.press("Enter");
        await expect(moreExamples.locator("..")).toHaveAttribute("open", "");
      }
      await expect(
        page
          .getByText(example.canonicalExample, { exact: true })
          .filter({ visible: true }),
        "Canonical example must have exactly one globally visible copy",
      ).toHaveCount(1);
      await expect(
        main
          .getByText(example.canonicalExample, { exact: true })
          .filter({ visible: true }),
      ).toBeVisible();
    }
    expect(
      apiMutationCount,
      "Context practice must not mutate learning state",
    ).toBe(0);
  } finally {
    context.off("request", countMutation);
  }
}
