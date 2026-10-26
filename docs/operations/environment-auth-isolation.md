# Environment authentication boundaries

Staging and production select separate session and CSRF cookie names:

| Environment | Session                       | CSRF                       |
| ----------- | ----------------------------- | -------------------------- |
| Staging     | `vocanova_staging_session`    | `vocanova_staging_csrf`    |
| Production  | `vocanova_production_session` | `vocanova_production_csrf` |
| Local/test  | `vocanova_session`            | `vocanova_csrf`            |

For a local real API, set `ENVIRONMENT=local` and leave the web build identity local/unknown so both sides use legacy names. Local/test web forwarding preserves mock-fixture context; deployed tiers use strict allowlists.

The API selects names from its configured environment. The browser and server web bundle use the existing `NEXT_PUBLIC_VOCANOVA_BUILD_ENVIRONMENT` build identity. Deployed tiers accept their own names without a legacy fallback. OAuth state remains a host-only cookie on the corresponding API host; registered Google callback URLs stay unchanged.

The trusted shared nginx edge forwards only each tier's authentication cookies to that tier's API/web upstream. Web requests also retain the theme cookie. The monitor upstream receives no Cookie header; its Socket.IO authentication remains separate. Server-side web requests apply the same tier filter before calling the API.

This prevents ordinary cross-tier login overwrites and stops foreign authentication cookies reaching application upstreams. Parent-domain scope is still required by the current sibling API/web hostnames. Browsers therefore still send parent-domain cookies to the trusted edge, and sibling responses can set parent-domain cookies. Shared host administrators remain trusted. Full browser-origin or administrative isolation requires a separate staging domain or an API topology change.

## Request identities and limits

Public nginx ingress overwrites `X-Real-IP` and `X-Forwarded-For` with its normalized peer address. Existing Cloudflare real-IP restoration trusts only configured Cloudflare peer ranges. The API uses the normalized single `X-Real-IP`, validates it as an address, and otherwise uses the TCP peer; arbitrary forwarding chains never choose rate-limit identities. The API must remain reachable only through this trusted ingress.

Each in-memory limiter expires old buckets and holds at most 10,000 keys. When full of live keys it rejects new keys without evicting existing limits. Limits remain per process; a shared limiter is required before horizontally scaling API replicas.

## Rollout and rollback

Publish matching API, web bundle, nginx fragments and synthetic checks together using the repository deployment workflows. Existing browser sessions need a fresh sign-in after the name change. Preserve the configured Secure, HttpOnly and SameSite behavior and the existing CSRF/CORS checks.

Synthetic mint responses describe the actual cookie names alongside opaque tokens. Deployment and scheduled checks must validate the expected tier and use those names, rather than assume a legacy name. Missing or mismatched metadata fails the check.

Rollback must restore a complete matching release through the normal workflow, including API, browser bundle and edge configuration. Rolling back application images alone while retaining new cookie filters can leave authentication unavailable. A rollback can require another sign-in.

## Verification

Run the authentication package with the race detector, the web helper/middleware tests, the actual nginx proxy regression harness and the synthetic-mint helper tests. The nginx harness uses only isolated disposable resources and synthetic tokens. After staging succeeds, verify production through the manual deployment workflow, test authenticated server reads and CSRF-protected writes, and check that signing into both environments preserves both sessions. These checks do not establish full penetration-test or browser-origin isolation.
