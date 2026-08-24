#!/usr/bin/env python3
"""Prepare a roles.yml stored model string for Cursor CLI --model invocation.

Stored bindings use the cursor/<model>[param=value,...] convention from
config/roles.yml. Workflows must pass the result to `agent --model` without
stripping bracket parameters or silently substituting another vendor/model.
"""

from __future__ import annotations

import argparse
import os
import sys

CURSOR_PREFIX = "cursor/"


class CursorModelError(Exception):
    """Fail-closed model routing refusal with a safe reason code."""


def prepare_cursor_model(
    stored: str,
    *,
    require_api_key: bool = False,
) -> str:
    if require_api_key and not os.environ.get("CURSOR_API_KEY"):
        raise CursorModelError("missing_cursor_api_key")

    if not stored.startswith(CURSOR_PREFIX):
        raise CursorModelError("unsupported_provider_prefix")

    cli_model = stored[len(CURSOR_PREFIX) :]
    if not cli_model:
        raise CursorModelError("empty_cursor_model")

    return cli_model


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "stored_model",
        help="Model string from roles.yml (e.g. cursor/grok-4.6[fast=false])",
    )
    parser.add_argument(
        "--require-api-key",
        action="store_true",
        help="Fail closed when CURSOR_API_KEY is unset (never print its value).",
    )
    parser.add_argument(
        "--check-prefix",
        action="store_true",
        help="Validate cursor/ prefix only; do not require CURSOR_API_KEY.",
    )
    args = parser.parse_args(argv)

    try:
        cli_model = prepare_cursor_model(
            args.stored_model,
            require_api_key=args.require_api_key and not args.check_prefix,
        )
    except CursorModelError as exc:
        print(f"prepare-cursor-model: {exc}", file=sys.stderr)
        return 1

    sys.stdout.write(cli_model)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
