# VOC-095-T02 — Live staging core-loop verification

Task: `VOC-095-T02`
Evidence class: `VOC-095-EV-02`
Observed: 2026-08-20 (Asia/Tehran; GitHub timestamps are 2026-08-19 UTC)

## Integration revision

- T01 pull request: `#799`
- Reviewed T01 head: `bc061dc2f02ac231bde159798f221eb3fc10f795`
- `develop` merge commit: `abf268f1d7f396894107cd375ecceab70aa02220`
- Deploy workflow: `deploy-staging`
- Run: <https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32309952858>
- Job: `deploy to staging` (`96250729047`)
- Conclusion: `success`

The workflow run and job metadata both bind the deployment to the integration
commit above. The job ran from 2026-08-19 22:41:24 UTC through 22:45:04 UTC
(3 minutes 40 seconds).

## Browser setup and mandatory journey

Public step metadata records:

| Step                                                                      | UTC interval      | Conclusion |
| ------------------------------------------------------------------------- | ----------------- | ---------- |
| Restore Playwright Chromium browser cache for the staging core-loop check | 22:44:17–22:44:22 | success    |
| Install Playwright Chromium for the staging core-loop check               | 22:44:22–22:44:35 | success    |
| Run the staging core-loop journey                                         | 22:44:35–22:44:49 | success    |

The repository-managed installer therefore completed in 13 seconds and the
mandatory browser journey completed immediately afterward. The job did not
time out or get cancelled during browser installation.

The live runner did not reproduce the hosted apt-mirror stall, so the HTTPS
fallback was not needed in this run. The deterministic primary-timeout to
fallback-success and all-sources-fail-closed fixtures remain recorded in
`t00-evidence.md` and locked by `voc095-playwright-install.test.mjs`.

## Independent readiness check

After the deployment, the repository-managed staging OAuth-start verifier
passed without following the external identity-provider login:

- OAuth start returned HTTP 200;
- the authorization destination was the expected Google host;
- the redirect URI was the canonical staging callback;
- API health returned HTTP 200; and
- `controlled_signup_ready` was true without exposing cohort metadata.

No identity, OAuth state, authorization code, cookie, token, secret, SSH output,
or application log was collected or copied into this evidence.

## Closure

`VOC-095-AC-05` and `VOC-095-TEST-11` are satisfied by the exact integration
revision and successful live steps above. Issue `#792` can close after this
evidence change passes exact-SHA CI and independent review, merges to `develop`,
and task issue `#796` closes through the governed roster path.
