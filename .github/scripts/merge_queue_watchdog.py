#!/usr/bin/env python3
"""Recover stale GitHub merge-queue entries without changing their order.

The Actions ``GITHUB_TOKEN`` is safe for reading the queue, but GitHub does
not emit workflows for events it causes.  A separate, minimally scoped token
is therefore required only when this script needs to mutate the queue.
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
from dataclasses import dataclass
from datetime import UTC, datetime
from typing import Callable


QUEUE_QUERY = """
query($owner: String!, $repo: String!) {
  repository(owner: $owner, name: $repo) {
    mergeQueue(branch: "main") {
      entries(first: 100) {
        nodes {
          state
          enqueuedAt
          pullRequest { number id }
        }
      }
    }
  }
}
"""

DEQUEUE_MUTATION = """
mutation($id: ID!) {
  dequeuePullRequest(input: {id: $id}) { clientMutationId }
}
"""

ENQUEUE_MUTATION = """
mutation($id: ID!) {
  enqueuePullRequest(input: {pullRequestId: $id}) {
    mergeQueueEntry { state position }
  }
}
"""


class RecoveryError(RuntimeError):
    """An incomplete recovery that needs an operator to inspect the queue."""


@dataclass(frozen=True)
class Entry:
    number: int
    node_id: str
    state: str
    enqueued_at: datetime

    @classmethod
    def from_api(cls, raw: dict[str, object]) -> "Entry":
        pull_request = raw["pullRequest"]
        if not isinstance(pull_request, dict):
            raise RecoveryError("merge queue returned an entry without a pull request")
        timestamp = raw["enqueuedAt"]
        if not isinstance(timestamp, str):
            raise RecoveryError("merge queue returned an entry without an enqueue time")
        return cls(
            number=int(pull_request["number"]),
            node_id=str(pull_request["id"]),
            state=str(raw["state"]),
            enqueued_at=datetime.fromisoformat(timestamp.replace("Z", "+00:00")),
        )


Runner = Callable[[str, dict[str, str], str | None], str]


def gh_runner(query: str, variables: dict[str, str], token: str | None) -> str:
    env = os.environ.copy()
    if token:
        env["GH_TOKEN"] = token
    result = subprocess.run(
        [
            "gh",
            "api",
            "graphql",
            "-f",
            f"query={query}",
            *[
                item
                for pair in variables.items()
                for item in ("-f", f"{pair[0]}={pair[1]}")
            ],
        ],
        check=False,
        capture_output=True,
        text=True,
        env=env,
    )
    if result.returncode:
        raise RecoveryError(result.stderr.strip() or "GitHub API request failed")
    return result.stdout


class MergeQueue:
    def __init__(
        self,
        owner: str,
        repo: str,
        read_token: str,
        mutation_token: str,
        runner: Runner = gh_runner,
    ):
        self.owner = owner
        self.repo = repo
        self.read_token = read_token
        self.mutation_token = mutation_token
        self.runner = runner

    def entries(self) -> list[Entry]:
        response = json.loads(
            self.runner(
                QUEUE_QUERY,
                {"owner": self.owner, "repo": self.repo},
                self.read_token,
            )
        )
        nodes = response["data"]["repository"]["mergeQueue"]["entries"]["nodes"]
        return [Entry.from_api(node) for node in nodes]

    def mutate(self, query: str, entry: Entry, operation: str) -> None:
        if not self.mutation_token:
            raise RecoveryError(
                f"PR #{entry.number}: {operation} requires MERGE_QUEUE_WATCHDOG_TOKEN. "
                "Configure a fine-grained token with repository Contents and Pull requests read/write access; "
                "GITHUB_TOKEN cannot dispatch the required merge_group workflows."
            )
        try:
            self.runner(query, {"id": entry.node_id}, self.mutation_token)
        except RecoveryError as error:
            raise RecoveryError(
                f"PR #{entry.number}: {operation} failed; queue recovery stopped. "
                f"Inspect the merge queue and re-enqueue its original order manually. GitHub API: {error}"
            ) from error

    def dequeue(self, entry: Entry) -> None:
        self.mutate(DEQUEUE_MUTATION, entry, "dequeue")

    def enqueue(self, entry: Entry) -> None:
        self.mutate(ENQUEUE_MUTATION, entry, "enqueue")


def stale(entry: Entry, now: datetime, stale_minutes: int) -> bool:
    return entry.state == "AWAITING_CHECKS" and (now - entry.enqueued_at).total_seconds() >= stale_minutes * 60


def recover(
    queue: MergeQueue,
    now: datetime,
    stale_minutes: int,
    out: Callable[[str], None] = print,
) -> None:
    entries = queue.entries()
    stale_entries = [entry for entry in entries if stale(entry, now, stale_minutes)]
    if not stale_entries:
        out("No stale AWAITING_CHECKS merge-queue entries found.")
        return

    # When every entry is stale, independent re-kicks only rotate the queue.
    # Empty it first, then enqueue the captured FIFO order in one recovery.
    if len(entries) > 1 and len(stale_entries) == len(entries):
        order = ", ".join(f"#{entry.number}" for entry in entries)
        out(f"All {len(entries)} queue entries are stale; resetting FIFO order: {order}.")
        for entry in entries:
            queue.dequeue(entry)

        remaining = queue.entries()
        if remaining:
            remaining_numbers = ", ".join(f"#{entry.number}" for entry in remaining)
            raise RecoveryError(
                f"Queue was not empty after reset dequeue (still queued: {remaining_numbers}); "
                "no entries were re-enqueued. Inspect the queue before retrying."
            )

        for entry in entries:
            queue.enqueue(entry)
        out(f"Reset complete; re-enqueued FIFO order: {order}.")
        return

    # A stale entry among otherwise healthy work gets the established single
    # re-kick, avoiding a disruptive full reset.
    for entry in stale_entries:
        out(
            f"PR #{entry.number}: stale AWAITING_CHECKS entry; "
            "re-kicking without resetting healthy work."
        )
        queue.dequeue(entry)
        queue.enqueue(entry)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repository", required=True, help="owner/repository")
    parser.add_argument("--stale-minutes", type=int, default=10)
    args = parser.parse_args()
    owner, separator, repo = args.repository.partition("/")
    if not separator or not owner or not repo:
        parser.error("--repository must be owner/repository")
    if args.stale_minutes < 1:
        parser.error("--stale-minutes must be at least 1")

    queue = MergeQueue(
        owner,
        repo,
        os.environ.get("GH_TOKEN", ""),
        os.environ.get("MERGE_QUEUE_WATCHDOG_TOKEN", ""),
    )
    try:
        recover(queue, datetime.now(UTC), args.stale_minutes)
    except (RecoveryError, KeyError, ValueError, json.JSONDecodeError) as error:
        print(f"merge-queue watchdog: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
