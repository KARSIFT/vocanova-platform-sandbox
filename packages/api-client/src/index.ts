/** Generated-contract foundation for GET /api/v1/me. */
export interface CurrentUser {
  email?: string;
  displayName?: string;
  avatarUrl?: string;
  emailVerifiedAt?: string;
}

/** Body for POST /api/v1/auth/magic-links. */
export interface RequestMagicLinkBody {
  email: string;
}

/** Body for POST /api/v1/auth/magic-links/consume. */
export interface ConsumeMagicLinkBody {
  token: string;
  email: string;
}

/** Body for POST /api/v1/auth/oauth/google/start. */
export interface OAuthStartBody {
  redirectUri: string;
}

/** Response for POST /api/v1/auth/oauth/google/start. */
export interface OAuthStartResponse {
  url: string;
}

/** Huma problem-details error model. */
export interface ApiError {
  type?: string;
  title?: string;
  status?: number;
  detail?: string;
  instance?: string;
  errors?: Array<{ location?: string; message?: string; value?: unknown }>;
}
