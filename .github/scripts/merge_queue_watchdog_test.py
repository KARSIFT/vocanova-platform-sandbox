#!/usr/bin/env python3
import json
import unittest
from datetime import UTC, datetime, timedelta

from merge_queue_watchdog import (
    DEQUEUE_MUTATION,
    ENQUEUE_MUTATION,
    QUEUE_QUERY,
    Entry,
    MergeQueue,
    RecoveryError,
    recover,
)


NOW = datetime(2026, 9, 5, tzinfo=UTC)


def entry(number: int, minutes_old: int, state: str = "AWAITING_CHECKS") -> Entry:
    return Entry(number, f"PR_{number}", state, NOW - timedelta(minutes=minutes_old))


class FakeGitHub:
    def __init__(self, entries: list[Entry], failures: dict[tuple[str, int], str] | None = None):
        self.entries = entries[:]
        self.failures = failures or {}
        self.calls: list[tuple[str, int]] = []

    def run(self, query: str, variables: dict[str, str], token: str | None) -> str:
        if query == QUEUE_QUERY:
            return json.dumps(
                {
                    "data": {
                        "repository": {
                            "mergeQueue": {
                                "entries": {
                                    "nodes": [
                                        {
                                            "state": item.state,
                                            "enqueuedAt": item.enqueued_at.isoformat().replace(
                                                "+00:00", "Z"
                                            ),
                                            "pullRequest": {
                                                "number": item.number,
                                                "id": item.node_id,
                                            },
                                        }
                                        for item in self.entries
                                    ]
                                }
                            }
                        }
                    }
                }
            )
        number = int(variables["id"].removeprefix("PR_"))
        operation = "dequeue" if query == DEQUEUE_MUTATION else "enqueue"
        self.calls.append((operation, number))
        if (operation, number) in self.failures:
            raise RecoveryError(self.failures[(operation, number)])
        if operation == "dequeue":
            self.entries = [item for item in self.entries if item.number != number]
        else:
            self.entries.append(entry(number, 0, "QUEUED"))
        return "{}"


class MergeQueueWatchdogTest(unittest.TestCase):
    def queue(self, fake: FakeGitHub, mutation_token: str = "token") -> MergeQueue:
        return MergeQueue("KARSIFT", "vocanova-platform-sandbox", "read-token", mutation_token, fake.run)

    def test_resets_all_stale_entries_in_original_fifo_order(self) -> None:
        fake = FakeGitHub([entry(1196, 20), entry(1198, 19), entry(1199, 18), entry(1201, 17)])
        messages: list[str] = []

        recover(self.queue(fake), NOW, 10, messages.append)

        self.assertEqual(fake.calls, [
            ("dequeue", 1196), ("dequeue", 1198), ("dequeue", 1199), ("dequeue", 1201),
            ("enqueue", 1196), ("enqueue", 1198), ("enqueue", 1199), ("enqueue", 1201),
        ])
        self.assertEqual([item.number for item in fake.entries], [1196, 1198, 1199, 1201])
        self.assertIn("FIFO order: #1196, #1198, #1199, #1201", messages[0])

    def test_non_head_stale_entry_does_not_reset_healthy_work(self) -> None:
        fake = FakeGitHub([entry(1, 2, "QUEUED"), entry(2, 20), entry(3, 1, "AWAITING_CHECKS")])

        recover(self.queue(fake), NOW, 10)

        self.assertEqual(fake.calls, [("dequeue", 2), ("enqueue", 2)])
        self.assertEqual([item.number for item in fake.entries], [1, 3, 2])

    def test_single_stale_entry_keeps_existing_recovery(self) -> None:
        fake = FakeGitHub([entry(1201, 20)])

        recover(self.queue(fake), NOW, 10)

        self.assertEqual(fake.calls, [("dequeue", 1201), ("enqueue", 1201)])

    def test_dequeue_failure_identifies_affected_pr_and_stops(self) -> None:
        fake = FakeGitHub([entry(1, 20), entry(2, 20)], {("dequeue", 2): "forbidden"})

        with self.assertRaisesRegex(RecoveryError, r"PR #2: dequeue failed"):
            recover(self.queue(fake), NOW, 10)
        self.assertEqual(fake.calls, [("dequeue", 1), ("dequeue", 2)])

    def test_enqueue_failure_identifies_affected_pr_and_stops(self) -> None:
        fake = FakeGitHub([entry(1, 20), entry(2, 20)], {("enqueue", 2): "forbidden"})

        with self.assertRaisesRegex(RecoveryError, r"PR #2: enqueue failed"):
            recover(self.queue(fake), NOW, 10)
        self.assertEqual(fake.calls, [("dequeue", 1), ("dequeue", 2), ("enqueue", 1), ("enqueue", 2)])

    def test_missing_mutation_token_fails_with_setup_instructions(self) -> None:
        fake = FakeGitHub([entry(1201, 20)])

        with self.assertRaisesRegex(RecoveryError, r"MERGE_QUEUE_WATCHDOG_TOKEN"):
            recover(self.queue(fake, ""), NOW, 10)


if __name__ == "__main__":
    unittest.main()
