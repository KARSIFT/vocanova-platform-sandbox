// Real staging only. No local mock, email, third-party signup or owner account.
import { readFile } from "node:fs/promises";
import { expect, test, type BrowserContext } from "@playwright/test";
import type { CurrentUser, WordDetailResponse, SentenceFeedbackResult } from "@vocanova/api-client";

const API = "https://api-staging.vocanova.site";
const WEB = "https://staging.vocanova.site";
const EMAIL = "smoke-test-bot@synthetic.vocanova.invalid";
const MEANING = "329c4ec9-6562-568e-9e71-9134da2f0d8d";
const SESSION = "vocanova_staging_session";
const CSRF = "vocanova_staging_csrf";

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
// This helper only permits two fixed read endpoints; errors contain no bodies.
async function readStaging<T>(context: BrowserContext, path: "/api/v1/me" | "/api/v1/canonical-words/menu"): Promise<T> {
  try {
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

test("configured canonical image loads and decodes from its approved public CDN", async ({ page, context }) => {
  const word = await readStaging<WordDetailResponse>(context, "/api/v1/canonical-words/menu");
  const asset = word.word.meanings.find((meaning) => meaning.id === MEANING)?.image;
  test.skip(!asset, "Acceptance gap: this staging meaning has no configured approved image; CDN delivery is unverified.");
  if (!asset) return;
  const expectedURL = `https://images.vocanova.site/meanings/${MEANING}/${asset.meaningRevision}/${asset.sha256}.webp`;
  expect(asset.imageUrl === expectedURL).toBe(true);
  let imageStatus: number | undefined;
  page.on("response", (response) => { if (response.url() === expectedURL) imageStatus = response.status(); });
  await page.goto(`/vocabulary/menu#meaning-${MEANING}`);
  const image = page.locator("img").filter({ visible: true }).and(page.locator(`img[src="${expectedURL}"]`));
  await expect(image).toBeVisible();
  expect(await image.evaluate(async (node) => {
    const image = node as HTMLImageElement;
    await image.decode();
    return image.complete && image.naturalWidth > 0 && image.naturalHeight > 0;
  })).toBe(true);
  expect(imageStatus).toBe(200);
});
