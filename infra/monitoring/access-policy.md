# monitor.vocanova.site — access exposure policy (VOC-081-DEP-00)

## Decision

**Private administrative surface via Cloudflare Access (Zero Trust).**

This is the package-recommended default for `VOC-081-DEP-00`. A proxied
Cloudflare DNS A record is **not** authorization for the Uptime Kuma UI
(`VOC-081-D02`). Origin nginx reverse-proxies Kuma only after Cloudflare
Access has authenticated the operator at the edge.

## Required controls

| Layer | Control | Owner |
| ----- | ------- | ----- |
| Edge | Cloudflare Access application for hostname `monitor.vocanova.site` | Founder / ops (Cloudflare Zero Trust dashboard) |
| Application | Uptime Kuma's own login remains enabled | Founder (Kuma admin credentials) |

The repository vhost lives at
`infra/nginx-shared/conf.d/30-monitor.vocanova.site.conf` and is loaded by
`vocanova-shared-edge-nginx`. It does **not** implement Access JWT validation
at origin — Access enforcement is expected at Cloudflare before requests reach
the origin on `:443`.

## Ops provisioning (not in git)

1. In Cloudflare Zero Trust, create (or restore) an Access application for
   `monitor.vocanova.site` with an identity provider and policy that denies
   unauthenticated public access.
2. Keep the DNS record proxied (orange-cloud) to `130.185.123.152` as today.
3. Confirm Kuma authentication is still required after passing Access.

Do not commit Cloudflare API tokens, Access service tokens, or Kuma admin
passwords to this repository.

## T04 verification probe (live)

Record redacted evidence in `t04-evidence.md` after T03 deploy:

```bash
# Unauthenticated — expect Cloudflare Access challenge or deny, not Kuma HTML.
curl -sS -o /dev/null -w '%{http_code}\n' https://monitor.vocanova.site/

# Authenticated operator path — expect Kuma UI HTML after Access + Kuma login.
# (browser session or Access service token; redact tokens in evidence)
```

**FAIL** if the only cited control is “DNS is proxied.”

## Stale production vhost retirement

The pre-VOC-081 host file
`/opt/vocanova/production/nginx/conf.d/30-monitor.conf` was never loaded by
shared edge (only production `10-*.conf` / `20-*.conf` globs apply) and must
not become a second source of truth. Repository marker:
`infra/nginx-production/conf.d/30-monitor.conf.superseded`. T03 deploy
convergence removes any live `30-monitor.conf` under the production nginx
tree when applying the repository bundle.
