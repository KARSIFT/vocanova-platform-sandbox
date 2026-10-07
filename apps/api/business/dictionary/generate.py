#!/usr/bin/env python3
"""Create the bounded WordNet lookup index from the pinned, local NLTK ZIP.

No network access or third-party Python packages. See SOURCE.md for provenance.
Usage: python3 generate.py /path/to/wordnet.zip
"""
import gzip
import hashlib
import json
from pathlib import Path
import re
import sys
import zipfile

SOURCE_SHA256 = "cbda5ea6eef7f36a97a43d4a75f85e07fccbb4f23657d27b4ccbc93e2646ab59"
DEST = Path(__file__).resolve().parent
WORD = re.compile(r"[a-z]+(?:['-][a-z]+)*\Z")


def generate(source):
    raw = Path(source).read_bytes()
    if hashlib.sha256(raw).hexdigest() != SOURCE_SHA256:
        raise ValueError("WordNet ZIP checksum differs from reviewed source")
    entries, senses, exceptions = {}, [], {}
    with zipfile.ZipFile(source) as archive:
        notice = archive.read("wordnet/LICENSE").decode("utf-8")
        for file_pos, label in [("noun", "noun"), ("verb", "verb"), ("adj", "adjective"), ("adv", "adverb")]:
            records = {}
            for line in archive.read(f"wordnet/data.{file_pos}").decode("utf-8").splitlines():
                if not line or not line[0].isdigit():
                    continue
                metadata, gloss = line.split("|", 1)
                # Preserve definitions verbatim. Only terminal, quoted examples
                # separated with semicolons are split into their own UI field.
                gloss = gloss.strip()
                example = ""
                match = re.search(r'; "([^"\n]+)"(?:; "[^"\n]+")*$', gloss)
                if match:
                    example = match.group(1)
                    gloss = gloss[:match.start()].rstrip()
                records[metadata.split()[0]] = (label, gloss, example)
            positions = {}
            for line in archive.read(f"wordnet/index.{file_pos}").decode("utf-8").splitlines():
                if not line or line.startswith(" "):
                    continue
                fields = line.split()
                lemma = fields[0]
                if len(lemma) > 48 or not WORD.fullmatch(lemma):
                    continue
                count, pointer_count = int(fields[2]), int(fields[3])
                offsets = fields[6 + pointer_count:6 + pointer_count + count]
                for offset in offsets[:2]:
                    if offset not in positions:
                        positions[offset] = len(senses)
                        senses.append(records[offset])
                    entries.setdefault(lemma, []).append(positions[offset])
            for line in archive.read(f"wordnet/{file_pos}.exc").decode("utf-8").splitlines():
                fields = line.split()
                if not fields or not WORD.fullmatch(fields[0]):
                    continue
                for base in fields[1:]:
                    if base in entries:
                        target = exceptions.setdefault(fields[0], [])
                        if base not in target:
                            target.append(base)
        payload = {"license": notice, "entries": entries, "senses": senses, "exceptions": exceptions}
        encoded = json.dumps(payload, ensure_ascii=False, sort_keys=True, separators=(",", ":")).encode()
        packed = gzip.compress(encoded, compresslevel=9, mtime=0)
        (DEST / "wordnet.json.gz").write_bytes(packed)
        (DEST / "LICENSE.wordnet").write_text(notice, encoding="utf-8")
        print(json.dumps({"entries": len(entries), "senses": len(senses), "exceptions": len(exceptions), "jsonBytes": len(encoded), "gzipBytes": len(packed), "gzipSha256": hashlib.sha256(packed).hexdigest()}))


if __name__ == "__main__":
    generate(sys.argv[1])
