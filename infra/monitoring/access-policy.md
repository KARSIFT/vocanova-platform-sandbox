# monitor.vocanova.site — access exposure policy (VOC-081-DEP-00)

## Decision

**Public Uptime Kuma login, protected by Kuma's own authentication.**

This preserves the pre-VOC-081 exposure model while moving routing under the
repository-managed shared edge. A proxied Cloudflare DNS A record provides
edge transport/proxying but is **not** authorization for the Uptime Kuma UI
(`VOC-081-D02`). Kuma's own login remains enabled and is the administrative
authorization boundary.

## Required controls

| Layer       | Control                                                           | Owner                    |
| ----------- | ----------------------------------------------------------------- | ------------------------ |
| Edge        | Proxied DNS/TLS for `monitor.vocanova.site`                       | Cloudflare configuration |
| Application | Uptime Kuma login remains enabled and required for administration | Kuma configuration       |

The repository vhost lives at
`infra/nginx-shared/conf.d/30-monitor.vocanova.site.conf` and is loaded by
`vocanova-shared-edge-nginx`. It does not claim or depend on a Cloudflare
Access application that is absent from the known configuration.

## Ops provisioning (not in git)

1. Keep the DNS record proxied (orange-cloud) to `130.185.123.152` as today.
2. Keep Kuma's own authentication enabled; do not enable anonymous
   administrative access.
3. Confirm an unauthenticated browser reaches Kuma's login rather than an
   authenticated dashboard, then confirm a valid Kuma operator can sign in.

Do not commit Cloudflare API tokens, session cookies, or Kuma admin passwords
to this repository.

## T04 verification probe (live)

After T03 deploy convergence succeeds, run:

```bash
infra/scripts/verify-voc081-monitor.sh
# VOC-086 full availability + topology closure:
infra/scripts/verify-voc086-monitoring.sh
```

Record redacted output in package evidence (`t04-evidence.md` for VOC-081;
`t05-evidence.md` for VOC-086). The VOC-081 script checks:

- four app hostnames still return HTTP 2xx on canonical `:443`
- `https://monitor.vocanova.site/` returns the Kuma application (follows
  `/` → `/dashboard`; not 520/502)
- socket.io Engine.IO polling handshake returns 2xx with a session id
  (WebSocket upgrade advertised)
- unauthenticated access shows the Kuma SPA / entry-page boundary
  (VOC-081-DEP-00), not an authenticated session payload

Offline harness (no real hostname required):

```bash
infra/scripts/verify-voc081-monitor.selftest.sh
```

**FAIL** if Kuma administration is anonymously accessible or the only cited
authorization control is “DNS is proxied.”

## Stale production vhost retirement

The pre-VOC-081 host file
`/opt/vocanova/production/nginx/conf.d/30-monitor.conf` was never loaded by
shared edge (only production `10-*.conf` / `20-*.conf` globs apply) and must
not become a second source of truth. Repository marker:
`infra/nginx-production/conf.d/30-monitor.conf.superseded`. T03 deploy
convergence removes any live `30-monitor.conf` under the production nginx
tree when applying the repository bundle.
