# Claude Code review guidance

When reviewing a PR here — automatically on open (`.github/workflows/claude-code-review.yml`)
or on an `@claude` mention (`.github/workflows/claude-review.yml`) — focus on:

1. **Correctness** — does the diff do what the PR description says, with no
   introduced bugs, edge cases, or regressions?
2. **Security** — no secrets, no injection vectors, no unsafe handling of user input
   or auth.
3. **Tests** — meaningful coverage for the change; flag missing edge cases rather
   than demanding tests for everything.
4. **Consistency** — matches the surrounding code's existing patterns unless the PR
   has a good reason to diverge.
5. **Simplicity** — flag unnecessary complexity, unused code, or abstractions the
   change doesn't need.

Post findings as PR comments (`gh pr comment`), inline where possible. Don't modify
files, push commits, or approve/merge the PR yourself — reviewing is advisory, not a
merge gate. Report what you found; the person merging decides what to act on.
