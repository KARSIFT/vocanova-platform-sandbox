#!/usr/bin/env python3
"""Verify both publicly served images belong to the intended deployment."""
import datetime
import json
import os
import time
import urllib.error
import urllib.request


def matches_identity(actual, expected):
    if any(actual.get(key) != expected[key] for key in ("version", "commit", "environment")):
        return False
    try:
        deployed = datetime.datetime.fromisoformat(actual["builtAt"].replace("Z", "+00:00"))
        built = datetime.datetime.fromisoformat(expected["builtAt"].replace("Z", "+00:00"))
        return deployed == built
    except (KeyError, TypeError, ValueError, AttributeError):
        return False


def main():
    expected = {
        "version": os.environ["RELEASE_VERSION"],
        "commit": os.environ["RELEASE_COMMIT"],
        "environment": os.environ["RELEASE_ENVIRONMENT"],
        "builtAt": os.environ["RELEASE_BUILT_AT"],
    }
    for service in ("API", "WEB"):
        url = os.environ[f"RELEASE_{service}_URL"].rstrip("/") + "/version"
        last_failure = "release metadata mismatch"
        for attempt in range(12):
            try:
                request = urllib.request.Request(url, headers={"Cache-Control": "no-cache", "User-Agent": "VocaNova-Release-Check/1.0"})
                with urllib.request.urlopen(request, timeout=5) as response:
                    actual = json.loads(response.read(16384))
                last_failure = "release metadata mismatch"
                if isinstance(actual, dict) and matches_identity(actual, expected):
                    print(f"{service} release identity matches the deployment")
                    break
            except urllib.error.HTTPError as error:
                last_failure = f"HTTP {error.code}"
            except (urllib.error.URLError, TimeoutError, ValueError):
                last_failure = "request failed or returned invalid JSON"
            if attempt == 11:
                raise SystemExit(f"{service} release identity is unavailable or does not match the deployment ({last_failure})")
            time.sleep(5)


if __name__ == "__main__":
    main()
