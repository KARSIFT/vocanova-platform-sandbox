export interface CurrentUser {
  email?: string;
  displayName?: string;
  avatarUrl?: string;
  emailVerifiedAt?: string;
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

  async getCurrentUser(init?: RequestInit): Promise<{
    data: CurrentUser;
    response: Response;
  }> {
    const response = await this.request("GET", "/api/v1/me", undefined, init);
    const data = (await response.json()) as CurrentUser;
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
