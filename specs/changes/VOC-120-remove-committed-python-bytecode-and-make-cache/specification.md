# VOC-120 — Remove committed Python bytecode and make cache ignores portable: Specification

## Objective and requirement source

Remove committed Python bytecode from the caller protected tree and make Python
cache ignore rules portable in both the caller and `KARSIFT/karsift-ai-infra`,
with deterministic validation so the same class of generated artifact cannot be
re-committed unnoticed.

**Requirement source:** [GitHub issue #987](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/987).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #987)

| Item | Value |
|------|-------|
| Caller `origin/develop` | `d67e34b149ad45fd8dcd0e3860a0bbfd1d3da0fb` |
| Caller `origin/main` | `48d7bb5a410035d4d1f12e5c34860399509d9215` |
| Tracked path | `infra/scripts/__pycache__/cloudflare_origin_port_remap.cpython-312.pyc` |
| Tracked blob | 5,825-byte CPython 3.12 timestamp-based bytecode |
| Introducing commit | `2118865b6acc4bbd533f62b0cf12330f6cf123d0` |
| Caller ignore rules already present | `__pycache__/` and `*.py[cod]` in repository `.gitignore` |
| Shared infra `origin/main` | `37b06aa95030e235b7311b3c14ee23977f62ac76` |
| Infra ignore gap | `git cat-file -e origin/main:.gitignore` fails; no repository `.gitignore` |
| Non-portable workaround | private `.git/info/exclude` entries (`__pycache__/`, `*.py[cod]`) |

## Scope and non-goals

### In scope

1. Untrack the committed caller bytecode blob and any other tracked Python
   bytecode/cache artifacts discovered at implementation time in the same task.
2. Leave `infra/scripts/cloudflare_origin_port_remap.py` and other Python source
   behavior unchanged.
3. Keep and verify the caller repository `.gitignore` rules `__pycache__/` and
   `*.py[cod]`.
4. Add a repository-root `.gitignore` to `KARSIFT/karsift-ai-infra` that covers
   `__pycache__/`, `*.pyc`, and related Python bytecode variants.
5. Add deterministic validation in both repositories:
   - caller protected/tracked source tree contains no tracked Python
     bytecode/cache artifacts;
   - caller ignore rules still cover the required patterns;
   - shared-infrastructure source ignore rules cover `__pycache__/`, `*.pyc`, and
     related variants;
   - negative cases fail closed.
6. Validate both repositories, independently review exact revisions, merge the
   infrastructure PR first if the caller fixture/pin consumes that change, pin any
   consumed caller mirror to the exact infrastructure merge SHA, and record
   rollback/evidence.
7. Preserve independent exact-SHA review, risk classification, protected checks,
   and retry limits.

### Non-goals / explicitly excluded

- Changing application runtime behavior, deployment topology, credentials,
  providers, or model routing.
- Rewriting git history to purge the historical bytecode blob.
- Replacing repository ignore rules with `PYTHONDONTWRITEBYTECODE` or private
  `.git/info/exclude` as the portable control.
- Expanding infra `.gitignore` into a full Python packaging ignore template
  (virtualenvs, eggs, `dist/`, coverage, and similar unrelated artifacts).
- Weakening governance, exact-SHA review, protected checks, risk classification,
  or retry limits.
- Treating the untracked local `karsift-ai-infra/` checkout, if present, as this
  repository's tracked tree.
- Self-adoption or self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: `scripts/governance/`, `tooling/governance/`, and caller
  `infra/` (bytecode lives under `infra/scripts/__pycache__/`). Shared-infra
  repository ignore rules and tests are in `KARSIFT/karsift-ai-infra`.
- Protected technical effect: source-tree hygiene and ignore/validation
  enforcement only; no product runtime effect is intended.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but governance validation changes still require
  exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a determination.
The path-based classifier and independent verifier remain authoritative.

## Decisions

`VOC-120-D00`: Untrack committed Python bytecode without changing Python source
behavior. The known blob is
`infra/scripts/__pycache__/cloudflare_origin_port_remap.cpython-312.pyc`. If
implementation `git ls-files` finds additional tracked `__pycache__/`, `*.pyc`,
`*.pyo`, or `*.pyd` paths, untrack those in the same task. Do not rewrite git
history. Do not modify `infra/scripts/cloudflare_origin_port_remap.py` except
insofar as an accidental source edit would be out of scope and must not land.

`VOC-120-D01`: Keep the existing caller repository `.gitignore` entries
`__pycache__/` and `*.py[cod]`. Verify they remain. Ignore rules are not a
substitute for untracking an already committed file.

`VOC-120-D02`: Add a repository-root `.gitignore` to `KARSIFT/karsift-ai-infra`
that covers at least `__pycache__/`, `*.pyc`, and related bytecode variants. The
portable form already used by the caller (`__pycache__/` plus `*.py[cod]`)
satisfies this. Private `.git/info/exclude` and `PYTHONDONTWRITEBYTECODE` are
not portable substitutes.

`VOC-120-D03`: Related bytecode variants for ignore rules and validation are
`__pycache__/`, `*.pyc`, `*.pyo`, and `*.pyd`. The gitignore character class
`*.py[cod]` covers those file extensions. Closely related `*$py.class` may be
included. Unrelated Python packaging ignores are out of scope.

`VOC-120-D04`: Add deterministic, fail-closed validation:

1. Caller: `git ls-files` on the protected source tree reports no tracked
   `__pycache__/` paths and no tracked `*.pyc` / `*.pyo` / `*.pyd` artifacts.
2. Caller: repository `.gitignore` still contains `__pycache__/` and a rule that
   covers `*.pyc` and related variants (`*.py[cod]` is sufficient).
3. Infra: repository `.gitignore` exists and contains those same required
   patterns.
4. Negative coverage: a synthetic tracked bytecode/cache path fails closed; a
   missing required ignore pattern fails closed.

Caller validation must run from the documented governance command
(`bash scripts/governance/validate-governance.sh`) and from
`tooling/governance/tests/`. Infra validation must run from the infra unit suite
(`python3 -m unittest discover -s tests -p 'test_*.py'`).

`VOC-120-D05`: One end-to-end implementation task covers both repositories.
Coordinated pull requests may remain one task. Merge infrastructure first if the
caller `tooling/governance/fixtures/karsift-ai-infra/` pin consumes the change,
and pin `PINNED_SHA.txt` to that exact reviewed infra merge. The fixture is
currently a policy-contract subset without a root `.gitignore` (pin
`37b06aa95030e235b7311b3c14ee23977f62ac76`). Do not expand that fixture to
include `.gitignore` merely to force a pin. If implementation adds a mirrored
test that actually consumes the new ignore file, then the pin is applicable and
required.

`VOC-120-D06`: Independent exact-SHA review, deterministic risk floors, protected
checks, and retry limits remain unchanged. This package does not authorize
application behavior, deployment topology, credential, provider, or
model-routing changes.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is generated-artifact cleanup,
ignore-rule, and hygiene-validation work only. No database, schema, seed,
analytics instrumentation, or user-interface accessibility effect.

## Security, privacy, and authorization

No new secrets, tokens, or privileged automation are introduced. The tracked
bytecode is a generated interpreter artifact, not a credential store, but it is
still a committed binary on protected branches and must be removed.

Abuse/process risks:

1. Re-committing bytecode because ignore rules were assumed to untrack files —
   mitigated by explicit untrack plus fail-closed `git ls-files` validation.
2. Treating local exclude-file or `PYTHONDONTWRITEBYTECODE` as sufficient —
   mitigated by requiring repository-level ignore rules and tests.
3. Expanding into unrelated Python ignore or fixture-pin churn — mitigated by
   `VOC-120-D03` and `VOC-120-D05`.

## Contradictions and open questions

1. **Closely related extra ignore pattern:** `*$py.class` is optional. T00 may
   include it as a related bytecode variant; it is not required if `__pycache__/`
   and `*.py[cod]` are present and tested.
2. **Additional tracked blobs:** issue #987 documents one path. If T00 finds
   more tracked bytecode/cache artifacts, untracking them is in-scope causal
   hygiene under `VOC-120-D00`, not a new package.
3. **Fixture pin applicability:** resolved in `VOC-120-D05` as conditional on
   actual fixture consumption. If implementation discovers the mirrored test
   tree cannot assert infra ignore rules without copying `.gitignore` into the
   policy fixture, prefer keeping that assertion in the primary infra suite only
   rather than expanding fixture scope.
4. **Validator module names:** suggested caller names are
   `tooling/governance/validate_python_bytecode_hygiene.py` and
   `tooling/governance/tests/test_python_bytecode_hygiene.py`; suggested infra
   name is `tests/test_python_cache_ignore.py`. Implementer may choose equivalent
   names in the same directories.
