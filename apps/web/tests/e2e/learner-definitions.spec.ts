import { createHash, randomUUID } from "node:crypto";
import type {
  KnowledgeMap,
  KnowledgeMapDecision,
  KnowledgeMapItem,
} from "@vocanova/api-client";
import { expect, test } from "@playwright/test";

for (const theme of ["light", "dark"] as const) {
  test(`current discovery details retain an authored explanation missing from the card in ${theme}`, async ({
    page,
    baseURL,
  }) => {
    await page.addInitScript(() =>
      Object.defineProperty(window, "speechSynthesis", {
        value: undefined,
        configurable: true,
      }),
    );
    await page.context().addCookies([
      {
        name: "vocanova_session",
        value: `learner-explanation-${randomUUID()}`,
        url: baseURL!,
      },
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    let reads = 0;
    const teaching =
      "Use this word when both people already know which person or thing you mean.";
    await page.route("**/api/v1/canonical-words/map-word-1", async (route) => {
      reads++;
      await route.fulfill({
        json: {
          word: {
            id: "map-word-1",
            text: "the",
            slug: "map-word-1",
            wordType: "word",
            meanings: [
              {
                id: "map-meaning-1",
                partOfSpeech: "article",
                shortDefinition:
                  "Used before a person or thing that is already clear.",
                learnerDefinition: teaching,
                saved: false,
                selfReportedKnown: false,
                examples: [],
                usageNotes: [],
              },
            ],
          },
        },
      });
    });
    await page.goto("/discovery?band=1&word=1&active=1");
    await expect(
      page.getByRole("heading", { name: "the", exact: true }),
    ).toBeVisible();
    expect(reads).toBe(0);
    const details = page.locator("details").filter({
      has: page.getByText("Word details & other meanings", { exact: true }),
    });
    await details.locator("summary").press("Enter");
    await expect(details.getByText(teaching, { exact: true })).toBeVisible();
    await expect(
      details.getByText("article · This meaning", { exact: true }),
    ).toBeVisible();
    expect(reads).toBe(1);
  });
}

for (const theme of ["light", "dark"] as const) {
  test(`plain map explanations preserve Learn and Already know identities in ${theme}`, async ({
    page,
    baseURL,
  }) => {
    await page.addInitScript(() =>
      Object.defineProperty(window, "speechSynthesis", {
        value: undefined,
        configurable: true,
      }),
    );
    await page.context().addCookies([
      {
        name: "vocanova_session",
        value: `plain-map-${randomUUID()}`,
        url: baseURL!,
      },
      { name: "vocanova_csrf", value: "plain-map-decisions", url: baseURL! },
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    const canonical: KnowledgeMapItem[] = [];
    const decisions: KnowledgeMapDecision[] = [];
    const teaching = (rank: number) =>
      `A plain explanation for the selected meaning at position ${rank}.`;
    await page.route("**/api/v1/knowledge-map?band=2", async (route) => {
      const response = await route.fetch();
      const original: KnowledgeMap = await response.json();
      if (!canonical.length) {
        canonical.push(
          ...structuredClone(original.currentBand.items.slice(0, 2)),
        );
      }
      const displayed = structuredClone(original);
      for (const item of displayed.currentBand.items.slice(0, 2)) {
        item.learnerDefinition = teaching(item.rank);
        // The shared fixture omits optional catalog metadata. Match the real
        // API's canonical revision derivation without changing source wording.
        item.meaningRevision ??= createHash("sha256")
          .update(
            JSON.stringify([
              item.meaningId,
              item.wordText,
              item.partOfSpeech,
              item.shortDefinition,
            ]),
          )
          .digest("hex");
      }
      await route.fulfill({ json: displayed });
    });
    await page.route(
      "**/api/v1/canonical-words/map-word-101",
      async (route) => {
        const item = canonical[0]!;
        const meaningRevision =
          item.meaningRevision ??
          createHash("sha256")
            .update(
              JSON.stringify([
                item.meaningId,
                item.wordText,
                item.partOfSpeech,
                item.shortDefinition,
              ]),
            )
            .digest("hex");
        await route.fulfill({
          json: {
            word: {
              id: item.wordId,
              text: item.wordText,
              slug: item.wordSlug,
              wordType: "word",
              meanings: [
                {
                  id: item.meaningId,
                  meaningRevision,
                  partOfSpeech: item.partOfSpeech,
                  shortDefinition: item.shortDefinition,
                  learnerDefinition: teaching(item.rank),
                  saved: item.saved,
                  selfReportedKnown: item.selfReportedKnown,
                  examples: [],
                  usageNotes: [],
                },
              ],
            },
          },
        });
      },
    );
    await page.route("**/api/v1/knowledge-map/decisions", async (route) => {
      decisions.push(route.request().postDataJSON());
      await route.continue();
    });
    await page.goto("/discovery?band=1&word=1&active=1");
    await expect(
      page.getByRole("heading", { name: "the", exact: true }),
    ).toBeVisible();
    await page.getByRole("button", { name: "Word range", exact: true }).click();
    await page
      .getByRole("dialog", { name: "Choose a range" })
      .getByRole("button", { name: "101–200", exact: true })
      .click();
    for (const [index, action, decision] of [
      [0, "Learn", "learn"],
      [1, "Already know", "known"],
    ] as const) {
      const item = canonical[index]!;
      await expect(
        page.getByRole("heading", { name: item.wordText, exact: true }),
      ).toBeVisible();
      await expect(
        page.getByText(teaching(item.rank), { exact: true }),
      ).toBeVisible();
      await expect(
        page.getByText(item.shortDefinition, { exact: true }),
      ).toHaveCount(0);
      if (index === 0) {
        const details = page
          .locator("details")
          .filter({
            has: page.getByText("Word details & other meanings", {
              exact: true,
            }),
          });
        await details.locator("summary").press("Enter");
        // Matching the untouched canonical metadata must pass the existing
        // detail-revision gate even while the card displays a plain explanation.
        await expect(
          details.getByText("This is the only meaning in Vocanova.", {
            exact: true,
          }),
        ).toBeVisible();
        await expect(
          details.getByRole("button", { name: "Reload word", exact: true }),
        ).toHaveCount(0);
        await expect(
          details.getByText(teaching(item.rank), { exact: true }),
        ).toHaveCount(0);
      }
      await page.getByRole("button", { name: action, exact: true }).click();
      await expect.poll(() => decisions.length).toBe(index + 1);
      expect(decisions[index]).toEqual({
        meaningId: item.meaningId,
        decision,
        expectedRevision: item.revision,
      });
      await expect(page).toHaveURL(new RegExp(`word=${item.rank + 1}(?:&|$)`));
    }
    await expect(
      page.getByRole("heading", { name: "Map word 103", exact: true }),
    ).toBeVisible();
    expect(canonical[0]!.shortDefinition).not.toBe(
      teaching(canonical[0]!.rank),
    );
    expect(canonical[1]!.shortDefinition).not.toBe(
      teaching(canonical[1]!.rank),
    );
  });
}
