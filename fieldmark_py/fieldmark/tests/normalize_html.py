"""Shared HTML normalization helper for cross-stack snapshot tests."""

from __future__ import annotations

import re

_COMMENT_RE = re.compile(r"<!--.*?-->", re.DOTALL)
_WHITESPACE_RE = re.compile(r"\s+")
# strips <input type="hidden" ...> so CSRF tokens don't break cross-stack parity diffs
_HIDDEN_INPUT_RE = re.compile(r'<input[^>]+type="hidden"[^>]*>', re.IGNORECASE)


def normalise_component(html: str) -> str:
    """Strip HTML comments, collapse all whitespace to single spaces, trim."""
    html = _COMMENT_RE.sub("", html)
    html = html.replace("&#34;", "&quot;")
    html = _WHITESPACE_RE.sub(" ", html).strip()
    return html


def normalise_for_parity(html: str) -> str:
    """Like normalise_component but also strips hidden inputs so CSRF tokens
    do not break cross-stack snapshot comparisons."""
    html = _HIDDEN_INPUT_RE.sub("", html)
    return normalise_component(html)


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
        if stripped == f"{start_marker} -->" or stripped.startswith(f"{start_marker} "):
            in_block = True
            continue
        if in_block and stripped.startswith("<!-- variant:"):
            break
        if in_block:
            collected.append(stripped)

    return normalise_component("\n".join(collected))
