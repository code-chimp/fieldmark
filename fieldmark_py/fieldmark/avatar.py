"""Avatar initials helper (AC #3, Story 1.13).

Algorithm:
  - Two+ whitespace-separated tokens → first char of first + first char of last, uppercased.
  - Single-token full name → first two characters, uppercased.
  - Empty or None full_name → first two characters of username_fallback, uppercased.
  - Unicode characters are preserved as-is (no transliteration).
"""


def initials(full_name: str | None, username_fallback: str | None) -> str:
    name = (full_name or "").strip()
    if name:
        tokens = name.split()
        if len(tokens) >= 2:
            return (tokens[0][0] + tokens[-1][0]).upper()
        if len(tokens[0]) >= 2:
            return (tokens[0][0] + tokens[0][1]).upper()
        return tokens[0].upper()

    fallback = (username_fallback or "").strip()
    if len(fallback) >= 2:
        return (fallback[0] + fallback[1]).upper()
    if fallback:
        return fallback.upper()
    return "??"
