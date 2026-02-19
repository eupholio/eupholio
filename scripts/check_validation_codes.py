#!/usr/bin/env python3
"""Ensure validation code docs are synchronized with CLI implementation."""

from __future__ import annotations

import re
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
RUST_FILE = ROOT / "eupholio-core/src/bin/eupholio-core-cli.rs"
DOC_FILE = ROOT / "eupholio-core/doc/09-validation-codes.md"

RUST_CODE_RE = re.compile(
    r"ValidationCode::([A-Za-z0-9_]+)\s*=>\s*(?:\{\s*)?\"([A-Z0-9_]+)\"",
    re.MULTILINE,
)
DOC_CODE_RE = re.compile(r"^\s*-\s*`([A-Z0-9_]+)`\s*$", re.MULTILINE)


def extract_rust_codes(text: str) -> dict[str, str]:
    pairs = RUST_CODE_RE.findall(text)
    if not pairs:
        raise ValueError("no ValidationCode::... => \"...\" mappings found in Rust file")

    variant_to_code: dict[str, str] = {}
    for variant, code in pairs:
        prev = variant_to_code.get(variant)
        if prev is not None and prev != code:
            raise ValueError(
                f"variant {variant} mapped to multiple codes: {prev!r} and {code!r}"
            )
        variant_to_code[variant] = code

    return variant_to_code


def extract_doc_codes(text: str) -> set[str]:
    codes = set(DOC_CODE_RE.findall(text))
    if not codes:
        raise ValueError("no markdown bullet validation codes found in doc file")
    return codes


def main() -> int:
    rust_text = RUST_FILE.read_text(encoding="utf-8")
    doc_text = DOC_FILE.read_text(encoding="utf-8")

    rust_codes_by_variant = extract_rust_codes(rust_text)
    rust_codes = set(rust_codes_by_variant.values())

    counts = Counter(rust_codes_by_variant.values())
    duplicates = sorted(code for code, count in counts.items() if count > 1)
    if duplicates:
        print("Validation code check failed: duplicate code strings in Rust mapping:")
        for code in duplicates:
            print(f"  - {code}")
        return 1

    doc_codes = extract_doc_codes(doc_text)

    missing_in_doc = sorted(rust_codes - doc_codes)
    missing_in_rust = sorted(doc_codes - rust_codes)

    if missing_in_doc or missing_in_rust:
        print("Validation code check failed: Rust and documentation are out of sync.")
        if missing_in_doc:
            print("Codes defined in Rust but missing in documentation:")
            for code in missing_in_doc:
                print(f"  - {code}")
        if missing_in_rust:
            print("Codes listed in documentation but missing in Rust:")
            for code in missing_in_rust:
                print(f"  - {code}")
        return 1

    print(
        f"Validation code check passed: {len(rust_codes)} codes are synchronized "
        f"between {RUST_FILE.relative_to(ROOT)} and {DOC_FILE.relative_to(ROOT)}."
    )
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:  # pragma: no cover - defensive for CI readability
        print(f"Validation code check failed with an unexpected error: {exc}")
        raise
