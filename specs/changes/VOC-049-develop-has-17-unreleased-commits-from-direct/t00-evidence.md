# VOC-049-EV-00 — T00 `main` vs `develop` re-verification

Evidence for `VOC-049-T00`, covering `VOC-049-AC-00` and `VOC-049-AC-03`.

## Comparison metadata

- Comparison timestamp (UTC): `2026-08-08T02:25:16Z`
- Source: GitHub compare API endpoint
  `https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/compare/main...develop`
- Comparison summary at snapshot time:
  - `status=diverged`
  - `ahead_by=23` (commits in `develop` not in `main`)
  - `behind_by=11` (commits in `main` not in `develop`)
  - `total_commits=23`
  - `merge_base=890a50d83464f649e582ebf8f089563d8da68f2a`
  - `base(main)=079de2899f374cad57b208606b1268157281fe77`

## Re-verified commit set for promotion scope (`main..develop`)

1. `9ef35a6532cd68c705f0f5b4f720b2095b8828da` — 2026-08-07T20:26:50Z — VOC-044: VOC-044-T01 (attempt 1) (#339)
2. `cbe5e1fb48a8d993c6a0c4f355a5d79f96cbb0b6` — 2026-08-07T23:15:10Z — Reconcile stale governance docs, drop false private-repo claim, tighten direct-commit exception
3. `66e361af6759b49f8bc1715f94ff70190802d08b` — 2026-08-07T23:17:24Z — Fix invalid YAML in VOC-038's change.yaml
4. `b7e3382b7c7c5c2830d851902c69bec641b6eb33` — 2026-08-07T23:19:59Z — Resolve duplicate change IDs VOC-043 and VOC-045
5. `1796fe0588c7a2bc60e4094d541f45c93e1cdebc` — 2026-08-07T23:35:21Z — Sync 16 stale package statuses: implementation-ready -> completed
6. `f6d0f4448338c3eff4fd2ffdba0dc6e3d622c4fb` — 2026-08-07T23:46:45Z — Authorize automatic release-to-main promotion and production deployment
7. `2c1f8579d38a247ed4368b747114937cb69ab01d` — 2026-08-07T23:53:13Z — Trigger re-check with corrected risk classification (R3)
8. `001d98c5e00617e652e9e2a6447724f040c9e1dc` — 2026-08-07T23:54:11Z — Trigger re-check (2)
9. `a1515747f4986a83cd6787833b5df8893e06c561` — 2026-08-07T23:55:07Z — Merge pull request #367 from KARSIFT/fix/voc-038-invalid-yaml
10. `1ddb100f16475282a48c1e63a8209f69a60100c3` — 2026-08-07T23:55:10Z — Merge pull request #368 from KARSIFT/fix/duplicate-change-ids
11. `f84b64d2b0840e9af6a499524800cb1e5ed9748e` — 2026-08-07T23:55:14Z — Merge pull request #369 from KARSIFT/sync/stale-package-statuses
12. `fb3c4207ef3695268dc2875d475c94da275152ba` — 2026-08-07T23:55:27Z — Merge pull request #370 from KARSIFT/auto-release-enabled
13. `c48e8cbbb6ccef1a77dac0bc3859abf38dfe3880` — 2026-08-07T23:55:57Z — Merge pull request #365 from KARSIFT/docs/governance-doc-reconciliation
14. `78f3d871d80eb98a1788beefefb60c645337c02a` — 2026-08-08T00:06:20Z — Sync 4 more stale package statuses: VOC-006 through VOC-009
15. `96eacd91142860be0732eb7d22a984faba829d5c` — 2026-08-08T00:11:03Z — Merge pull request #372 from KARSIFT/sync/voc-006-009-statuses
16. `17246249dd565bf063d84d84bf219fdb4ed5437b` — 2026-08-08T00:14:34Z — test: verify branch ruleset doesn't break the merge pipeline
17. `0914ea7d9d41d6ef6cd4fa4c3530ecb858c877b0` — 2026-08-08T00:16:45Z — Merge pull request #374 from KARSIFT/test/ruleset-verification
18. `e15b2a16a5eebb916a3e90a3c0c0e60957ae9299` — 2026-08-08T02:03:40Z — Plan: VOC-049 - develop-has-17-unreleased-commits-from-direct
19. `46599e82aa517d17eb46d8976786437c9efd4227` — 2026-08-08T02:05:26Z — Merge pull request #376 from KARSIFT/plan/voc-049-develop-has-17-unreleased-commits-from-direct
20. `576e66562e14b0428d6965c6b6b63bae52dac3cb` — 2026-08-08T02:10:35Z — VOC-049: record adoption (founder-delegate merged PR #376)
21. `eb01684fccfca4eb9d29ef85f864287f4642d35b` — 2026-08-08T02:12:53Z — Merge pull request #377 from KARSIFT/adopt-voc-049
22. `e33af4a38256ea416fd2b995232aaba350e38b7b` — 2026-08-08T02:22:21Z — VOC-049: recover roster/authority_issue after adopt.yml's push-auth bug
23. `ebd1c731eb7165d7bf1b166a0da376d5869aef00` — 2026-08-08T02:24:18Z — Merge pull request #384 from KARSIFT/recover-voc-049-roster

## Outcome for task gating

`VOC-049-T00` confirms a non-zero promotion gap at implementation time, so the
zero-gap closure path in `VOC-049-AC-03` does not apply and `VOC-049-T01`
remains in scope.
