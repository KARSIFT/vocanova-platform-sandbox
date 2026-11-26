#!/usr/bin/env python3
"""Build the public pronunciation index from an export and CDN receipts."""

import argparse
import hashlib
import json
from pathlib import Path
import re


def build_index(manifest, receipts):
    raw_manifest = manifest.read_bytes()
    manifest_hash = hashlib.sha256(raw_manifest).hexdigest()
    rows = {}
    cache_keys = set()
    for line in raw_manifest.splitlines():
        row = json.loads(line)
        key, checksum = row["cacheKey"], row["sha256"]
        if any(not re.fullmatch(r"[0-9a-f]{64}", value) for value in (key, checksum)):
            raise ValueError("Invalid recording hash")
        object_key = (
            f'meanings/{row["meaningId"]}/{row["meaningRevision"]}/{key}/{checksum}.mp3'
        )
        if row["path"] != f"public/{object_key}" or key in cache_keys:
            raise ValueError("Invalid object path or duplicate cache key")
        rows[object_key] = row
        cache_keys.add(key)

    confirmed = set()
    with receipts.open(encoding="utf-8") as stream:
        for line in stream:
            receipt = json.loads(line)
            key = receipt["objectKey"]
            row = rows.get(key)
            if (
                row is not None
                and receipt["verified"] is True
                and receipt["baseURL"] == "https://audio.vocanova.site"
                and receipt["manifestSHA256"] == manifest_hash
                and receipt["status"] == 200
                and receipt["contentType"] == "audio/mpeg"
                and receipt["observedBytes"] == row["bytes"]
                and receipt["observedSHA256"] == row["sha256"]
            ):
                confirmed.add(key)

    if not rows or len(confirmed) != len(rows):
        raise ValueError(f"CDN verified {len(confirmed)} of {len(rows)} manifest records")
    return b"".join(
        bytes.fromhex(row["cacheKey"] + row["sha256"])
        for row in sorted(rows.values(), key=lambda row: row["cacheKey"])
    )


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("manifest", type=Path)
    parser.add_argument("receipts", type=Path)
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    args.output.write_bytes(build_index(args.manifest, args.receipts))
