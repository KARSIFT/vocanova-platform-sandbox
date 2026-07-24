/** Generated-contract foundation for GET /api/v1/me. */
export interface CurrentUser {
  email?: string;
  displayName?: string;
  avatarUrl?: string;
  emailVerifiedAt?: string;
}

export interface ApiError {
  code: string;
  message: string;
}
