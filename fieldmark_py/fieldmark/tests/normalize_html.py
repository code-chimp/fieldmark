"""Shared HTML normalization helper for cross-stack snapshot tests."""

from __future__ import annotations

import re

_COMMENT_RE = re.compile(r"<!--.*?-->", re.DOTALL)
_WHITESPACE_RE = re.compile(r"\s+")


def normalise_component(html: str) -> str:
    """Strip HTML comments, collapse all whitespace to single spaces, trim."""
    html = _COMMENT_RE.sub("", html)
    html = _WHITESPACE_RE.sub(" ", html).strip()
    return html


def extract_variant(example_content: str, variant_name: str) -> str:
    """Extract a named variant block from a canonical component example file.

    Blocks are delimited by ``<!-- variant: <name> ... -->`` comment lines.
    Returns the normalised content between that delimiter and the next one.
    """
    lines = example_content.splitlines()
    start_marker = f"<!-- variant: {variant_name}"
    in_block = False
    collected: list[str] = []

    for line in lines:
        stripped = line.rstrip()
        if stripped.startswith(start_marker):
            in_block = True
            continue
        if in_block and stripped.startswith("<!-- variant:"):
            break
        if in_block:
            collected.append(stripped)

    return normalise_component("\n".join(collected))
