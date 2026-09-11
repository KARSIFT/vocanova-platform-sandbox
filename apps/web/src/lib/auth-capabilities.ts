import { createServerApiClient } from "./api-server";

export interface SignInAuthCapabilities {
  magicLinkEnabled: boolean;
  oauthEnabled: boolean;
  passwordEnabled: boolean;
}

/**
 * Read deploy-derived sign-in availability from the API's unauthenticated
 * health signal. Fails closed: when the capability probe cannot establish a
 * method is enabled, the screen avoids offering a flow that will fail.
 */
export async function getSignInAuthCapabilities(): Promise<SignInAuthCapabilities> {
  try {
    const client = await createServerApiClient();
    const { data } = await client.getHealthz();
    return {
      magicLinkEnabled: data.kill_switches?.magic_link_enabled === true,
      oauthEnabled: data.kill_switches?.oauth_enabled === true,
      passwordEnabled: data.kill_switches?.password_enabled === true,
    };
  } catch {
    return {
      magicLinkEnabled: false,
      oauthEnabled: false,
      passwordEnabled: false,
    };
  }
}
