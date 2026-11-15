// Real staging only. No local mock, email, third-party signup or owner account.
import { readFile } from "node:fs/promises";
import { createHash } from "node:crypto";
import { expect, test, type BrowserContext } from "@playwright/test";
import type { CurrentUser, WordDetailResponse, SentenceFeedbackResult } from "@vocanova/api-client";

const API = "https://api-staging.vocanova.site";
const WEB = "https://staging.vocanova.site";
const EMAIL = "smoke-test-bot@synthetic.vocanova.invalid";
const MEANING = "329c4ec9-6562-568e-9e71-9134da2f0d8d";
const SESSION = "vocanova_staging_session";
const CSRF = "vocanova_staging_csrf";
const CANONICAL_PATHS = {
  parcel: "/api/v1/canonical-words/parcel",
  menu: "/api/v1/canonical-words/menu",
  "key-card": "/api/v1/canonical-words/key-card",
  appetizer: "/api/v1/canonical-words/appetizer",
  gate: "/api/v1/canonical-words/gate",
} as const;
const STAGING_READ_PATHS = new Set<string>([
  "/api/v1/me",
  ...Object.values(CANONICAL_PATHS),
]);

// The service deduplicates equivalent user/meaning/sentence requests even with
// a new idempotency key. A workflow invocation therefore needs distinct authored
// text to exercise a fresh provider generation instead of earlier stored output.
function providerSentenceForRun(): string {
  const runId = process.env.GITHUB_RUN_ID;
  const attempt = process.env.GITHUB_RUN_ATTEMPT;
  if (!runId || !/^[1-9]\d{0,19}$/.test(runId) || !attempt || !/^[1-9]\d{0,3}$/.test(attempt)) {
    throw new Error("Fresh provider acceptance requires numeric GitHub run and attempt identifiers before any mutation.");
  }
  // The separator keeps (run 12, attempt 34) distinct from (run 123, attempt 4).
  return `I read the menu before choosing a bowl of soup for order number ${runId}-${attempt}.`;
}

test.describe.configure({ retries: 0 });
test.use({ trace: "off", video: "off", screenshot: "off" });
test.beforeEach(async ({ context, baseURL }) => {
  test.skip(process.env.E2E_PROVIDER_DATA_ACCEPTANCE !== "true", "Explicit synthetic staging provider/data acceptance opt-in required.");
  if (baseURL !== WEB || (process.env.STAGING_API_BASE_URL && process.env.STAGING_API_BASE_URL !== API))
    throw new Error("Provider/data acceptance only accepts the exact staging origins.");
  if (process.env.E2E_SESSION_COOKIE_NAME !== SESSION || process.env.E2E_CSRF_COOKIE_NAME !== CSRF)
    throw new Error("Only staging-scoped minted cookie names are accepted.");
  const session = process.env.E2E_SESSION_COOKIE;
  const csrf = process.env.E2E_CSRF_TOKEN;
  if (!session || !csrf) throw new Error("A fresh CI-minted synthetic staging session is required.");
  await context.addCookies([
    { name: SESSION, value: session, domain: ".vocanova.site", path: "/", secure: true, sameSite: "Lax" },
    { name: CSRF, value: csrf, domain: ".vocanova.site", path: "/", secure: true, sameSite: "Lax" },
  ]);
  const me = await readStaging<CurrentUser>(context, "/api/v1/me");
  expect(me.email === EMAIL, "Refuse all account access outside the reserved synthetic identity.").toBe(true);
});

// Node fetch keeps credentials out of Playwright's retained API-call diagnostics.
// Only fixed identity and reviewed-content endpoints; errors contain no bodies.
async function readStaging<T>(context: BrowserContext, path: "/api/v1/me" | typeof CANONICAL_PATHS[keyof typeof CANONICAL_PATHS]): Promise<T> {
  try {
    if (!STAGING_READ_PATHS.has(path)) throw new Error("Outside acceptance scope");
    const cookies = (await context.cookies(API)).filter((cookie) => cookie.name === SESSION || cookie.name === CSRF);
    const response = await fetch(`${API}${path}`, {
      headers: { Cookie: cookies.map((cookie) => `${cookie.name}=${cookie.value}`).join("; ") },
      redirect: "error", signal: AbortSignal.timeout(20_000),
    });
    if (response.status !== 200) throw new Error("Read failed");
    const body = await response.text();
    if (body.length > 1_000_000) throw new Error("Read exceeded bound");
    return JSON.parse(body) as T;
  } catch { throw new Error("Synthetic staging read failed; response and credentials withheld."); }
}

function publicSentenceProjection(value: unknown): boolean {
  if (!value || typeof value !== "object" || Array.isArray(value)) return false;
  const keys = new Set(["id", "meaningId", "userWordId", "sentenceText", "source", "status", "submittedAt", "createdAt", "updatedAt"]);
  return Object.keys(value).every((key) => keys.has(key));
}

function publicFeedbackHistory(value: unknown): boolean {
  if (!value || typeof value !== "object" || Array.isArray(value)) return false;
  const entry = value as Record<string, unknown>;
  if (!Object.keys(entry).every((key) => ["sentence", "feedbackAttempts"].includes(key)) || !publicSentenceProjection(entry.sentence) || !Array.isArray(entry.feedbackAttempts)) return false;
  return entry.feedbackAttempts.every((attempt: unknown) => {
    if (!attempt || typeof attempt !== "object" || Array.isArray(attempt)) return false;
    const fields = attempt as Record<string, unknown>;
    if (!Object.keys(fields).every((key) => ["status", "feedback", "feedbackText", "startedAt", "completedAt", "createdAt", "updatedAt"].includes(key))) return false;
    return fields.feedback === null || (typeof fields.feedback === "object" && fields.feedback !== null && !Array.isArray(fields.feedback) && Object.keys(fields.feedback).every((key) => ["status", "targetWordUsedCorrectly", "correctedSentence", "explanation", "improvementTip"].includes(key)));
  });
}

test("fresh AI generation returns the expected verdict and browser export contains that exact submission", async ({ page, context }) => {
  const sentence = providerSentenceForRun();
  const me = await readStaging<CurrentUser>(context, "/api/v1/me");
  const word = await readStaging<WordDetailResponse>(context, "/api/v1/canonical-words/menu");
  expect(word.word.meanings.some((meaning) => meaning.id === MEANING)).toBe(true);
  let submissions = 0;
  let forbidden = 0;
  await page.route("**/api/v1/**", async (route) => {
    const request = route.request();
    const url = new URL(request.url());
    if (url.origin !== API) { forbidden += 1; return route.abort("blockedbyclient"); }
    if (request.method() === "POST" && url.pathname === "/api/v1/learner-sentences") {
      submissions += 1;
      if (submissions > 1 || request.postDataJSON()?.sentenceText !== sentence) {
        forbidden += 1; return route.abort("blockedbyclient");
      }
    }
    if (request.method() === "POST" && url.pathname === "/api/v1/user-words" && request.postDataJSON()?.meaningId !== MEANING) {
      forbidden += 1; return route.abort("blockedbyclient");
    }
    // Other allowed mutations are the explicit menu save and personal export.
    if (!["GET", "HEAD", "OPTIONS"].includes(request.method()) &&
        !(request.method() === "POST" && ["/api/v1/user-words", "/api/v1/learner-sentences", "/api/v1/personal-data-export"].includes(url.pathname))) {
      forbidden += 1; return route.abort("blockedbyclient");
    }
    await route.continue();
  });
  await page.goto(`/vocabulary/menu#meaning-${MEANING}`);
  const meaning = page.locator(`[id="meaning-${MEANING}"]`);
  await expect(meaning).toBeVisible();
  const save = meaning.getByRole("button", { name: /^Save menu:/ });
  if (await save.count()) await save.click();
  await expect(meaning.getByRole("button", { name: "Remove menu from saved words" })).toBeVisible();
  const tools = meaning.locator("summary").filter({ hasText: "Personal tools" });
  if (await tools.locator("..").getAttribute("open") === null) await tools.click();
  const writing = meaning.locator("summary").filter({ hasText: "Optional writing" });
  if (await writing.locator("..").getAttribute("open") === null) await writing.click();
  const editor = meaning.getByRole("textbox", { name: "Write a sentence using menu" });
  await editor.fill(sentence);
  const submitted = page.waitForResponse((r) => r.request().method() === "POST" && new URL(r.url()).pathname === "/api/v1/learner-sentences");
  await meaning.getByRole("button", { name: "Check my sentence", exact: true }).click();
  const response = await submitted;
  expect(response.status(), "Real provider acceptance failed; distinguish provider availability from frontend correctness.").toBe(200);
  const result = await response.json() as SentenceFeedbackResult;
  expect(result.processingStatus === "completed", "Real AI must complete; an error alert is not successful provider acceptance.").toBe(true);
  // One run-specific, grammatical example is a bounded semantic smoke check,
  // not evidence of general pedagogical quality. A wrong verdict is a real
  // provider/quality acceptance gap, never converted into an allowed alert.
  expect(result.status === "correct", "The authored menu example must receive the expected correct verdict.").toBe(true);
  expect(result.targetWordUsedCorrectly === true, "The provider must recognize correct use of the restaurant-menu meaning.").toBe(true);
  expect(result.originalSentence === sentence && Boolean(result.sentenceId)).toBe(true);
  expect(Boolean(result.errorCode || result.errorMessage)).toBe(false);
  await expect(meaning.getByRole("status", { name: /^Feedback result: / })).toBeVisible();
  expect(submissions).toBe(1);

  await page.goto("/settings/account");
  const downloading = page.waitForEvent("download");
  await page.getByRole("button", { name: "Download my data", exact: true }).click();
  const download = await downloading;
  try {
    expect(await download.failure() === null).toBe(true);
    const path = await download.path();
    if (!path) throw new Error("Synthetic export download was unavailable.");
    const raw = await readFile(path, "utf8");
    expect(raw.length < 20_000_000, "Synthetic export exceeds acceptance size bound.").toBe(true);
    let data;
    try { data = JSON.parse(raw); }
    catch { throw new Error("Synthetic export was not valid JSON; contents withheld."); }
    expect(data.schemaVersion === "1.7").toBe(true);
    expect(data.profile?.id === me.id && data.profile?.email === EMAIL, "Export must belong to the verified synthetic requester.").toBe(true);
    const history = data.sentenceFeedbackHistory;
    expect(Array.isArray(history)).toBe(true);
    const record = history.find((entry: { sentence?: { id?: string } }) => entry.sentence?.id === result.sentenceId);
    expect(Boolean(record) && record.sentence.sentenceText === sentence && record.sentence.meaningId === MEANING).toBe(true);
    // Exact DTO projection excludes private request/receipt/provider fields.
    expect(history.every(publicFeedbackHistory)).toBe(true);
    const session = process.env.E2E_SESSION_COOKIE!;
    const csrf = process.env.E2E_CSRF_TOKEN!;
    expect(raw.includes(session) || raw.includes(csrf), "Export must not contain authentication values.").toBe(false);
  } finally { await download.delete(); }
  expect(forbidden).toBe(0);
});

test("reviewed picture bank loads approved public CDN assets and authored gap contexts match canonical content", async ({ page, context }) => {
  // This proves live metadata/assets and authored content eligibility, not that
  // a mixed session was selected for this account. PostgreSQL covers grading.
  const runtime = JSON.parse(await readFile(
    new URL("../../src/lib/meaning-pictures.runtime.json", import.meta.url), "utf8",
  )) as Record<string, { alt: string; meaningRevision: string }>;
  // Exact meanings from dailylesson/mixed_questions.go's reviewed picture set.
  const pictures = [
    { slug: "parcel", text: "parcel", meaningId: "0b593408-6679-5fb8-a21c-b73a525c76e5" },
    { slug: "menu", text: "menu", meaningId: MEANING },
    { slug: "key-card", text: "key card", meaningId: "d990ce6e-4756-52e0-96e8-9926442963f2" },
    { slug: "appetizer", text: "appetizer", meaningId: "2534b539-a80f-5001-b9a4-27a5e0679539" },
  ] as const;
  const canonical = new Map<string, WordDetailResponse["word"]>();
  for (const picture of pictures) {
    const { word } = await readStaging<WordDetailResponse>(context, CANONICAL_PATHS[picture.slug]);
    canonical.set(picture.slug, word);
    expect(word.slug).toBe(picture.slug);
    expect(word.text).toBe(picture.text);
    const matches = word.meanings.filter((meaning) => meaning.id === picture.meaningId);
    expect(matches).toHaveLength(1);
    const meaning = matches[0]!;
    const committed = runtime[picture.meaningId];
    expect(committed, "Every reviewed picture must exist in the committed runtime manifest.").toBeDefined();
    expect(meaning.partOfSpeech).toBe("noun");
    expect(meaning.meaningRevision).toBe(committed!.meaningRevision);
    const asset = meaning.image;
    expect(asset, "All four reviewed picture options require approved staging image metadata.").toBeDefined();
    const sha256 = createHash("sha256").update(await readFile(
      new URL(`../../public/images/meanings/${picture.meaningId}.webp`, import.meta.url),
    )).digest("hex");
    const expectedURL = `https://images.vocanova.site/meanings/${picture.meaningId}/${committed!.meaningRevision}/${sha256}.webp`;
    expect(asset!.meaningId).toBe(picture.meaningId);
    expect(asset!.meaningRevision).toBe(committed!.meaningRevision);
    expect(asset!.sha256).toBe(sha256);
    expect(asset!.alt).toBe(committed!.alt);
    expect(asset!.imageUrl).toBe(expectedURL);
    const delivered = page.waitForResponse((response) => response.url() === expectedURL);
    await page.goto(`/vocabulary/${picture.slug}#meaning-${picture.meaningId}`);
    const image = page.locator(`[id="meaning-${picture.meaningId}"] img[src="${expectedURL}"]`).filter({ visible: true });
    await expect(image).toBeVisible();
    expect(await image.evaluate(async (node) => {
      const image = node as HTMLImageElement;
      await image.decode();
      return image.complete && image.naturalWidth > 0 && image.naturalHeight > 0;
    })).toBe(true);
    expect((await delivered).status()).toBe(200);
  }

  // These exact active examples/revisions are the two pinned gap targets in
  // dailylesson/mixed_questions.go, not examples generated during acceptance.
  const gaps = [
    {
      slug: "gate", text: "gate", meaningId: "6a61d337-80e7-5d50-bc61-8b45b6a054a1",
      revision: "112a5a995ccdcaa2b3b3b231a0dcc165926b98b4732daf032eae37e0d0170a0b",
      example: "Our flight leaves from gate twelve.",
      context: "Our flight leaves from _____ twelve.",
    },
    {
      slug: "key-card", text: "key card", meaningId: "d990ce6e-4756-52e0-96e8-9926442963f2",
      revision: "068a52d16045d06ce7f8e5fe0575add338bd96df5e76aee46359ecf350d6fa7d",
      example: "My key card stopped working.",
      context: "My _____ stopped working.",
    },
  ] as const;
  for (const gap of gaps) {
    const word = canonical.get(gap.slug) ?? (
      await readStaging<WordDetailResponse>(context, CANONICAL_PATHS[gap.slug])
    ).word;
    expect(word.slug).toBe(gap.slug);
    expect(word.text).toBe(gap.text);
    const matches = word.meanings.filter((meaning) => meaning.id === gap.meaningId);
    expect(matches).toHaveLength(1);
    const meaning = matches[0]!;
    expect(meaning.partOfSpeech).toBe("noun");
    expect(meaning.meaningRevision).toBe(gap.revision);
    expect(meaning.examples.some((example) => example.exampleText === gap.example),
      "The exact reviewed gap example must be present in active canonical content.").toBe(true);
    expect(gap.example.split(gap.text)).toHaveLength(2);
    expect(gap.example.replace(gap.text, "_____")).toBe(gap.context);
  }
});
