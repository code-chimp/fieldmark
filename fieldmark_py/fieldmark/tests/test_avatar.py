"""Unit tests for fieldmark.avatar.initials (AC #3, Story 1.13)."""

from fieldmark.avatar import initials


def test_empty_full_name_falls_back_to_username():
    assert initials("", "alice") == "AL"


def test_none_full_name_falls_back_to_username():
    assert initials(None, "bob") == "BO"


def test_single_token_full_name_first_two_chars_uppercased():
    assert initials("Alice", "alice") == "AL"


def test_two_token_full_name_first_and_last_initials():
    assert initials("Alice Admin", "alice") == "AA"


def test_three_plus_token_full_name_first_and_last_initials():
    assert initials("Alice Marie Admin", "alice") == "AA"


def test_unicode_characters_preserved_as_is():
    assert initials("Ää Öö", "aao") == "ÄÖ"


def test_single_token_unicode_first_two_chars():
    assert initials("李明", "user") == "李明"


def test_both_empty_returns_fallback_token():
    assert initials(None, None) == "??"


def test_empty_full_name_and_empty_username_returns_fallback_token():
    assert initials("", "") == "??"
