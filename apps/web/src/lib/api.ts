import { VocanovaClient } from "@vocanova/api-client";

import { getApiBaseURL } from "./env";

export function createApiClient(): VocanovaClient {
  return new VocanovaClient({
    baseURL: getApiBaseURL(),
    credentials: "include",
  });
}
