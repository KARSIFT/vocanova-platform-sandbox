export interface CurrentUser {
  email?: string;
  displayName?: string;
  avatarUrl?: string;
  emailVerifiedAt?: string;
  /**
   * VOC-031-T01 additive field. Always present in the response.
   * The Next.js middleware uses it to gate the core-loop routes
   * on whether the learner has completed onboarding (DOC-03 §3).
   */
  onboardingStatus: "not_started" | "in_progress" | "completed";
}

export type EnglishLevel = "a1" | "a2" | "b1" | "b2" | "unknown";

export type LearningGoal =
  "general" | "work" | "travel" | "study" | "conversation" | "exam";

export type MainUseCase = "daily_life" | "work" | "travel" | "study" | "social";

export interface OnboardingProfile {
  status: "not_started" | "in_progress" | "completed";
  englishLevel?: EnglishLevel;
  nativeLanguage?: string;
  learningGoal?: LearningGoal;
  mainUseCase?: MainUseCase;
  dailyReviewTarget?: number;
  completedAt?: string;
}

export interface CompleteOnboardingBody {
  englishLevel: EnglishLevel;
  nativeLanguage: string;
  learningGoal: LearningGoal;
  mainUseCase: MainUseCase;
  dailyReviewTarget: number;
}

export interface RequestMagicLinkBody {
  email: string;
}

export interface ConsumeMagicLinkBody {
  token: string;
  email: string;
}

export interface OAuthStartBody {
  redirectUri: string;
}

export interface OAuthStartResponse {
  url: string;
}

/**
 * VOC-084-T01. Kill-switch state reported by GET /healthz. Field
 * names match the API's snake_case JSON (apps/api production
 * HealthzOutput).
 */
export interface KillSwitchStatus {
  magic_link_enabled?: boolean;
  oauth_enabled?: boolean;
  new_signups_enabled?: boolean;
  ai_enabled?: boolean;
}

/**
 * VOC-084-T01. Unauthenticated liveness probe body. The handler
 * returns HTTP 200 when healthy and 503 when unhealthy, but
 * kill_switches are present in both cases.
 */
export interface HealthzResponse {
  status: string;
  database?: string;
  timestamp?: string;
  kill_switches?: KillSwitchStatus;
}

export interface Situation {
  id: string;
  slug: string;
  title: string;
  shortDescription: string;
  levelBand?: string;
  category: string;
  displayOrder: number;
}

export interface SituationMeaning {
  meaningId: string;
  wordId: string;
  wordSlug: string;
  wordText: string;
  partOfSpeech: string;
  shortDefinition: string;
  saved: boolean;
}

export interface WordExample {
  id: string;
  exampleText: string;
  situationLabel?: string;
}

export interface WordUsageNote {
  id: string;
  noteType: string;
  noteText: string;
}

export interface WordMeaning {
  id: string;
  partOfSpeech: string;
  shortDefinition: string;
  learnerDefinition?: string;
  saved: boolean;
  userWordId?: string;
  examples: WordExample[];
  usageNotes: WordUsageNote[];
}

export interface WordDetail {
  id: string;
  text: string;
  slug: string;
  wordType: string;
  difficultyLevel?: string;
  meanings: WordMeaning[];
}

export interface ListSituationsResponse {
  items: Situation[];
  nextCursor?: string;
}

export interface SituationResponse {
  situation: Situation;
  meanings: SituationMeaning[];
}

export interface WordDetailResponse {
  word: WordDetail;
}

export interface SavedMeaning {
  userWordId: string;
  meaningId: string;
  wordId: string;
  wordText: string;
  wordSlug: string;
  partOfSpeech: string;
  shortDefinition: string;
  status: string;
  source: string;
  saved: boolean;
  addedAt: string;
}

export interface ListSavedWordsResponse {
  items: SavedMeaning[];
  nextCursor?: string;
}

export interface SaveUserWordBody {
  meaningId: string;
  source: "journey" | "search" | "manual";
}

export interface DueWord {
  userWordId: string;
  meaningId: string;
  wordId: string;
  wordText: string;
  wordSlug: string;
  partOfSpeech: string;
  shortDefinition: string;
  status: string;
  reviewStep: number;
}

export interface ListDueWordsResponse {
  items: DueWord[];
  nextCursor?: string;
  totalCount: number;
}

export interface ReviewAttempt {
  attemptId: string;
  userWordId: string;
  meaningId: string;
  attemptType: string;
  promptType: "multiple_choice" | "self_check";
  result: "correct" | "incorrect" | "skipped";
  rating?: "again" | "hard" | "good" | "easy";
  reviewStepBefore: number;
  reviewStepAfter: number;
  answeredAt: string;
  responseTimeMs: number;
  selectedOptionMeaningId?: string;
  typedAnswer?: string;
  wasHintUsed: boolean;
  source: string;
  clientAttemptId: string;
  nextReviewAt: string;
}

export interface SubmitReviewBody {
  userWordId: string;
  meaningId: string;
  attemptType?: "review";
  promptType: "multiple_choice" | "self_check";
  result: "correct" | "incorrect" | "skipped";
  rating?: "again" | "hard" | "good" | "easy";
  answeredAt: string;
  responseTimeMs?: number;
  selectedOptionMeaningId?: string;
  typedAnswer?: string;
  wasHintUsed?: boolean;
  source?: "review" | "review_session";
  clientAttemptId: string;
  metadata?: Record<string, unknown>;
}

export interface SentenceFeedbackResult {
  sentenceId?: string;
  attemptId?: string;
  status?: "correct" | "needs_improvement" | "incorrect";
  originalSentence: string;
  correctedSentence?: string;
  explanation?: string;
  improvementTip?: string;
  missionCompleted: boolean;
  canRetry: boolean;
  reported: boolean;
  errorCode?: string;
  errorMessage?: string;
  crisisResourceMessage?: string;
}

export interface SubmitSentenceFeedbackBody {
  sentenceText: string;
  source: "word_detail" | "review" | "daily_mission" | "free_practice";
  attemptId: string;
}

export interface ReportSentenceFeedbackBody {
  reason: string;
  classification?: string;
}

export interface Streak {
  currentStreakCount: number;
  longestStreakCount: number;
  status: "active" | "at_risk" | "broken";
  graceDayBalance: number;
}

export interface DailyMission {
  localDate: string;
  timezone: string;
  reviewTarget: number;
  reviewsCompleted: number;
  newWordTarget?: number;
  newWordsCompleted?: number;
  sentencePracticeTarget?: number;
  sentencePracticesCompleted?: number;
  policyVersion: string;
  status: "open" | "completed" | "missed" | "protected";
  completedAt?: string;
  graceApplied: boolean;
  streak: Streak;
}

export interface CompletionDay {
  localDate: string;
  completed: boolean;
}

export interface Progress {
  confidencePointsBalance: number;
  streak: Streak;
  completionHistory: CompletionDay[];
}

export type ReviewIntervalPreset =
  "vocanova_default" | "wordup_like" | "custom";

/**
 * VOC-031-T02 additive type. The persisted app language
 * preference; only "en" is accepted at launch (VOC-031-D06),
 * because no i18n infrastructure exists in this repository
 * today.
 */
export type AppLanguage = "en";

/**
 * VOC-031-T02. The public Settings projection returned by
 * GET /api/v1/settings. The /api/v1/settings/account frontend
 * reads this for every editable Settings field.
 */
export interface Settings {
  dailyReviewTarget: number;
  reviewIntervalPreset: ReviewIntervalPreset;
  appLanguage: AppLanguage;
  notificationsEnabled: boolean;
  marketingEmailsEnabled: boolean;
  displayName: string;
}

/**
 * VOC-031-T02. The partial-update payload for
 * PATCH /api/v1/settings. Every field is optional; the API
 * only writes the fields the caller supplies. The DOC-07 §3
 * "no-op PATCH is a well-formed read" rule is honored, so
 * an empty body returns the current state.
 */
export interface UpdateSettingsBody {
  dailyReviewTarget?: number;
  reviewIntervalPreset?: ReviewIntervalPreset;
  appLanguage?: AppLanguage;
  notificationsEnabled?: boolean;
  marketingEmailsEnabled?: boolean;
  displayName?: string;
}

/**
 * VOC-031-T03. The request body for
 * POST /api/v1/settings/email-change-links. The new email is
 * the destination the requester wants to switch to; the
 * current sign-in address is taken from the session and is
 * never trusted from the body. The request is unconditionally
 * generic on the server side, so the registration status of
 * the new address is never observable through the request
 * outcome (anti-enumeration posture, VOC-031-D05).
 */
export interface RequestEmailChangeLinkBody {
  newEmail: string;
}

/**
 * VOC-031-T03. The request body for
 * POST /api/v1/settings/email-change-links/consume. The token
 * is the only form the requester supplies; the API never sees
 * the email itself at consume time (the server resolved it
 * when the request was issued).
 */
export interface ConsumeEmailChangeLinkBody {
  token: string;
}

/**
 * VOC-031-T03. The post-confirm response from
 * POST /api/v1/settings/email-change-links/consume. The
 * server returns the new email and the previous email so the
 * frontend can show the learner which address the security
 * notification was dispatched to; the notification itself is
 * owned by the backend, not the frontend.
 */
export interface ConsumeEmailChangeLinkResult {
  email: string;
  previousEmail: string;
  changedAt: string;
}

/**
 * VOC-031-T04. The post-deactivation response from
 * POST /api/v1/account-deletion-requests. The user is already
 * deactivated at this point: status is 'deactivated', every
 * active session and every unconsumed auth/email-change
 * token is revoked, and the purge_after clock is running. The
 * frontend uses the dates to render a clear "your account
 * has been scheduled for deletion" confirmation and to
 * initiate logout. `replayed` is true when the call was a
 * no-op because the (user, idempotency-key) pair already
 * matched a prior request — the frontend uses it to
 * suppress duplicate toasts on a retry.
 */
export interface CreateAccountDeletionRequestResult {
  status: string;
  userId: string;
  requestedAt: string;
  purgeAfter: string;
  idempotencyKey: string;
  replayed: boolean;
}

export interface ApiError {
  type?: string;
  title?: string;
  status?: number;
  detail?: string;
  instance?: string;
  errors?: Array<{ location?: string; message?: string; value?: unknown }>;
}

export class ApiResponseError extends Error {
  constructor(
    readonly status: number,
    readonly body: ApiError | null,
    message?: string,
  ) {
    super(message ?? body?.detail ?? `HTTP ${status}`);
  }
}

export interface VocanovaClientOptions {
  baseURL: string;
  credentials?: RequestCredentials;
  fetch?: typeof fetch;
}

export class VocanovaClient {
  private readonly fetch: typeof fetch;

  constructor(private readonly options: VocanovaClientOptions) {
    this.fetch = options.fetch ?? globalThis.fetch.bind(globalThis);
  }

  /**
   * VOC-084-T01. Fetch GET /healthz for deploy-derived capability
   * signals (kill switches). Unlike other client methods, this does
   * not throw on HTTP 503 so callers can still read kill_switches
   * when the database probe is unhealthy.
   */
  async getHealthz(init?: RequestInit): Promise<{
    data: HealthzResponse;
    response: Response;
  }> {
    const url = new URL("/healthz", this.options.baseURL);
    const headers = new Headers(init?.headers);
    if (!headers.has("Accept")) {
      headers.set("Accept", "application/json");
    }

    const response = await this.fetch(url.toString(), {
      ...init,
      method: "GET",
      headers,
      credentials: this.options.credentials,
    });
    const data = (await response.json()) as HealthzResponse;
    return { data, response };
  }

  async getCurrentUser(init?: RequestInit): Promise<{
    data: CurrentUser;
    response: Response;
  }> {
    const response = await this.request("GET", "/api/v1/me", undefined, init);
    const data = (await response.json()) as CurrentUser;
    return { data, response };
  }

  async getOnboarding(init?: RequestInit): Promise<{
    data: OnboardingProfile;
    response: Response;
  }> {
    const response = await this.request(
      "GET",
      "/api/v1/onboarding",
      undefined,
      init,
    );
    const data = (await response.json()) as OnboardingProfile;
    return { data, response };
  }

  async completeOnboarding(
    body: CompleteOnboardingBody,
    init?: RequestInit,
  ): Promise<{ data: OnboardingProfile; response: Response }> {
    const response = await this.request(
      "POST",
      "/api/v1/onboarding",
      body,
      init,
    );
    const data = (await response.json()) as OnboardingProfile;
    return { data, response };
  }

  async requestMagicLink(
    body: RequestMagicLinkBody,
    init?: RequestInit,
  ): Promise<{ response: Response }> {
    const response = await this.request(
      "POST",
      "/api/v1/auth/magic-links",
      body,
      init,
    );
    return { response };
  }

  async consumeMagicLink(
    body: ConsumeMagicLinkBody,
    init?: RequestInit,
  ): Promise<{ data: CurrentUser; response: Response }> {
    const response = await this.request(
      "POST",
      "/api/v1/auth/magic-links/consume",
      body,
      init,
    );
    const data = (await response.json()) as CurrentUser;
    return { data, response };
  }

  async startOAuth(
    body: OAuthStartBody,
    init?: RequestInit,
  ): Promise<{ data: OAuthStartResponse; response: Response }> {
    const response = await this.request(
      "POST",
      "/api/v1/auth/oauth/google/start",
      body,
      init,
    );
    const data = (await response.json()) as OAuthStartResponse;
    return { data, response };
  }

  async logout(init?: RequestInit): Promise<{ response: Response }> {
    const response = await this.request(
      "POST",
      "/api/v1/auth/logout",
      undefined,
      init,
    );
    return { response };
  }

  async listJourneySituations(
    params?: { after?: string; limit?: number },
    init?: RequestInit,
  ): Promise<{ data: ListSituationsResponse; response: Response }> {
    const query = new URLSearchParams();
    if (params?.after) {
      query.set("after", params.after);
    }
    if (params?.limit !== undefined) {
      query.set("limit", String(params.limit));
    }
    const path =
      "/api/v1/journey-situations" +
      (query.toString() ? `?${query.toString()}` : "");
    const response = await this.request("GET", path, undefined, init);
    const data = (await response.json()) as ListSituationsResponse;
    return { data, response };
  }

  async getJourneySituation(
    slug: string,
    init?: RequestInit,
  ): Promise<{ data: SituationResponse; response: Response }> {
    const response = await this.request(
      "GET",
      `/api/v1/journey-situations/${encodeURIComponent(slug)}`,
      undefined,
      init,
    );
    const data = (await response.json()) as SituationResponse;
    return { data, response };
  }

  async getCanonicalWord(
    wordSlug: string,
    init?: RequestInit,
  ): Promise<{ data: WordDetailResponse; response: Response }> {
    const response = await this.request(
      "GET",
      `/api/v1/canonical-words/${encodeURIComponent(wordSlug)}`,
      undefined,
      init,
    );
    const data = (await response.json()) as WordDetailResponse;
    return { data, response };
  }

  async listSavedWords(
    params?: { after?: string; limit?: number },
    init?: RequestInit,
  ): Promise<{ data: ListSavedWordsResponse; response: Response }> {
    const query = new URLSearchParams();
    if (params?.after) {
      query.set("after", params.after);
    }
    if (params?.limit !== undefined) {
      query.set("limit", String(params.limit));
    }
    const path =
      "/api/v1/user-words" + (query.toString() ? `?${query.toString()}` : "");
    const response = await this.request("GET", path, undefined, init);
    const data = (await response.json()) as ListSavedWordsResponse;
    return { data, response };
  }

  async listDueWords(
    params?: { after?: string; limit?: number },
    init?: RequestInit,
  ): Promise<{ data: ListDueWordsResponse; response: Response }> {
    const query = new URLSearchParams();
    if (params?.after) {
      query.set("after", params.after);
    }
    if (params?.limit !== undefined) {
      query.set("limit", String(params.limit));
    }
    const path =
      "/api/v1/reviews/due" + (query.toString() ? `?${query.toString()}` : "");
    const response = await this.request("GET", path, undefined, init);
    const data = (await response.json()) as ListDueWordsResponse;
    return { data, response };
  }

  async saveUserWord(
    body: SaveUserWordBody,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: SavedMeaning; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request("POST", "/api/v1/user-words", body, {
      ...init,
      headers,
    });
    const data = (await response.json()) as SavedMeaning;
    return { data, response };
  }

  async unsaveUserWord(
    meaningId: string,
    init?: RequestInit,
  ): Promise<{ response: Response }> {
    const response = await this.request(
      "DELETE",
      `/api/v1/user-words/${encodeURIComponent(meaningId)}`,
      undefined,
      init,
    );
    return { response };
  }

  async submitReview(
    body: SubmitReviewBody,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: ReviewAttempt; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "POST",
      "/api/v1/reviews/submissions",
      body,
      {
        ...init,
        headers,
      },
    );
    const data = (await response.json()) as ReviewAttempt;
    return { data, response };
  }

  async submitSentenceFeedback(
    body: SubmitSentenceFeedbackBody,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: SentenceFeedbackResult; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "POST",
      "/api/v1/sentence-feedback",
      body,
      {
        ...init,
        headers,
      },
    );
    const data = (await response.json()) as SentenceFeedbackResult;
    return { data, response };
  }

  async reportSentenceFeedback(
    attemptId: string,
    body: ReportSentenceFeedbackBody,
    init?: RequestInit,
  ): Promise<{ response: Response }> {
    const response = await this.request(
      "POST",
      `/api/v1/sentence-feedback/${encodeURIComponent(attemptId)}/reports`,
      body,
      init,
    );
    return { response };
  }

  async getDailyMission(
    params?: { timezone?: string },
    init?: RequestInit,
  ): Promise<{ data: DailyMission; response: Response }> {
    const query = new URLSearchParams();
    if (params?.timezone) {
      query.set("timezone", params.timezone);
    }
    const path =
      "/api/v1/daily-mission" +
      (query.toString() ? `?${query.toString()}` : "");
    const response = await this.request("GET", path, undefined, init);
    const data = (await response.json()) as DailyMission;
    return { data, response };
  }

  async getProgress(
    params?: { timezone?: string },
    init?: RequestInit,
  ): Promise<{ data: Progress; response: Response }> {
    const query = new URLSearchParams();
    if (params?.timezone) {
      query.set("timezone", params.timezone);
    }
    const path =
      "/api/v1/progress" + (query.toString() ? `?${query.toString()}` : "");
    const response = await this.request("GET", path, undefined, init);
    const data = (await response.json()) as Progress;
    return { data, response };
  }

  /**
   * VOC-031-T02. Get the requester's settings. The response is
   * a stable Settings projection — every field is always
   * present, with schema defaults for any unset value.
   */
  async getSettings(init?: RequestInit): Promise<{
    data: Settings;
    response: Response;
  }> {
    const response = await this.request(
      "GET",
      "/api/v1/settings",
      undefined,
      init,
    );
    const data = (await response.json()) as Settings;
    return { data, response };
  }

  /**
   * VOC-031-T02. Update the requester's settings via a partial
   * PATCH. Only the fields supplied in `body` are written; every
   * other field is preserved. An empty body is a well-formed
   * no-op read and returns the current state. The `init.headers`
   * are forwarded so the caller can attach a CSRF token.
   */
  async updateSettings(
    body: UpdateSettingsBody,
    init?: RequestInit,
  ): Promise<{ data: Settings; response: Response }> {
    const response = await this.request(
      "PATCH",
      "/api/v1/settings",
      body,
      init,
    );
    const data = (await response.json()) as Settings;
    return { data, response };
  }

  /**
   * VOC-031-T03. Request a single-use email-change link. The
   * request is unconditionally generic on the server side
   * (anti-enumeration posture, VOC-031-D05): whether the
   * requested new email is already registered is never
   * observable through this response. The `init.headers` are
   * forwarded so the caller can attach a CSRF token.
   */
  async requestEmailChangeLink(
    body: RequestEmailChangeLinkBody,
    init?: RequestInit,
  ): Promise<{ response: Response }> {
    const response = await this.request(
      "POST",
      "/api/v1/settings/email-change-links",
      body,
      init,
    );
    return { response };
  }

  /**
   * VOC-031-T03. Consume a single-use email-change link. The
   * server validates the token's hash, expiry, single-use
   * `consumed_at`, and environment, re-checks new-email
   * uniqueness atomically at confirm time, and updates
   * `users.email`. The `init.headers` are forwarded so the
   * caller can attach a CSRF token.
   */
  async consumeEmailChangeLink(
    body: ConsumeEmailChangeLinkBody,
    init?: RequestInit,
  ): Promise<{
    data: ConsumeEmailChangeLinkResult;
    response: Response;
  }> {
    const response = await this.request(
      "POST",
      "/api/v1/settings/email-change-links/consume",
      body,
      init,
    );
    const data = (await response.json()) as ConsumeEmailChangeLinkResult;
    return { data, response };
  }

  /**
   * VOC-031-T04. Deactivate the requester's account and
   * schedule anonymization. Requires a CSRF token and a
   * unique Idempotency-Key (DOC-07). A replay with the same
   * key returns the existing row with `replayed: true`, so
   * the frontend can suppress duplicate "your account was
   * deleted" toasts on a retry. The user is already
   * deactivated at this point: every active session is
   * revoked, and the `purgeAfter` clock is running. The
   * frontend should follow up with a logout request to clear
   * the session cookie.
   */
  async createAccountDeletionRequest(
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{
    data: CreateAccountDeletionRequestResult;
    response: Response;
  }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "POST",
      "/api/v1/account-deletion-requests",
      undefined,
      { ...init, headers },
    );
    const data = (await response.json()) as CreateAccountDeletionRequestResult;
    return { data, response };
  }

  private async request(
    method: string,
    path: string,
    body?: unknown,
    init?: RequestInit,
  ): Promise<Response> {
    const url = new URL(path, this.options.baseURL);
    const headers = new Headers(init?.headers);
    if (!headers.has("Accept")) {
      headers.set("Accept", "application/json");
    }
    if (body !== undefined && !headers.has("Content-Type")) {
      headers.set("Content-Type", "application/json");
    }

    const response = await this.fetch(url.toString(), {
      ...init,
      method,
      headers,
      credentials: this.options.credentials,
      body: body === undefined ? undefined : JSON.stringify(body),
    });

    if (!response.ok) {
      const apiError = await parseProblemDetails(response).catch(() => null);
      throw new ApiResponseError(
        response.status,
        apiError,
        apiError?.detail ?? `HTTP ${response.status}`,
      );
    }

    return response;
  }
}

async function parseProblemDetails(
  response: Response,
): Promise<ApiError | null> {
  const contentType = response.headers.get("Content-Type") ?? "";
  if (!contentType.includes("application/problem+json")) {
    return null;
  }
  try {
    return (await response.json()) as ApiError;
  } catch {
    return null;
  }
}
