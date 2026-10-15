#!/usr/bin/env python3
"""Rewrite common Go comparisons to literal-first form under core/ and addons/."""
from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DIRS = [ROOT / "core", ROOT / "addons"]

SIMPLE_ERR_LEN = re.compile(
    r"if err != nil \|\| len\(([^)]+)\) == 0 \{\n(\s+)([^\n]+)\n\2\}",
)
REMAINING_ERR_LEN = re.compile(
    r"if err != nil \|\| len\(([^)]+)\) == 0",
)
LEN_EQ0 = re.compile(r"\blen\(([^)]+)\) == 0\b")
# Do not match struct fields (e.g. gid.Int64 == 0).
ID_LE0 = re.compile(r"(?<![.\w])([a-zA-Z_][a-zA-Z0-9_]*)\s*<=\s*0\b")
ID_EQ0 = re.compile(r"(?<![.\w])([a-zA-Z_][a-zA-Z0-9_]*)\s*==\s*0\b")

SKIP_EQ0 = frozenset({"err"})


def transform(source: str) -> str:
    def split_simple(match: re.Match[str]) -> str:
        inner = match.group(1)
        indent = match.group(2)
        body = match.group(3)
        return (
            f"if err != nil {{\n{indent}{body}\n{indent}}}\n"
            f"if 0 == len({inner}) {{\n{indent}{body}\n{indent}}}\n"
        )

    out = SIMPLE_ERR_LEN.sub(split_simple, source)
    out = REMAINING_ERR_LEN.sub(r"if err != nil || 0 == len(\1)", out)
    out = LEN_EQ0.sub(r"0 == len(\1)", out)
    out = ID_LE0.sub(lambda m: f"0 >= {m.group(1)}", out)

    def repl_eq0(match: re.Match[str]) -> str:
        name = match.group(1)
        if name in SKIP_EQ0:
            return match.group(0)
        return f"0 == {name}"

    out = ID_EQ0.sub(repl_eq0, out)
    return out


def main() -> int:
    changed = 0
    for base in DIRS:
        if not base.is_dir():
            continue
        for path in sorted(base.rglob("*.go")):
            text = path.read_text(encoding="utf-8")
            updated = transform(text)
            if updated != text:
                path.write_text(updated, encoding="utf-8")
                changed += 1
    print(f"updated {changed} Go files")
    return 0


if __name__ == "__main__":
    sys.exit(main())
