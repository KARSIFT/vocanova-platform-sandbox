import { createServerApiClient } from "./api-server";

export interface SignInAuthCapabilities {
  oauthEnabled: boolean;
}

/**
 * VOC-084-T01. Read deploy-derived OAuth availability from the API's
 * unauthenticated /healthz kill_switches.oauth_enabled signal. Fails
 * closed: any probe error or absent/false switch hides Google sign-in.
 */
export async function getSignInAuthCapabilities(): Promise<SignInAuthCapabilities> {
  try {
    const client = await createServerApiClient();
    const { data } = await client.getHealthz();
    return {
      oauthEnabled: data.kill_switches?.oauth_enabled === true,
    };
  } catch {
    return { oauthEnabled: false };
  }
}
