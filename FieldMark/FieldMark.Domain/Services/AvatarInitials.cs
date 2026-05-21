using System.Text;

namespace FieldMark.Domain.Services;

/// <summary>
/// Derives display initials from a user's full name or username fallback.
/// Algorithm (AC #3, Story 1.13):
///   - Two+ whitespace-separated tokens → first rune of first + first rune of last, uppercased.
///   - Single-token full name → first two runes, uppercased.
///   - Empty or null full name → first two runes of username, uppercased.
///   - Unicode characters are preserved as-is (no transliteration).
///   - Rune boundaries respected to avoid splitting surrogate pairs for non-BMP characters.
/// </summary>
public static class AvatarInitials
{
    public static string From(string? fullName, string? usernameFallback)
    {
        var name = (fullName ?? string.Empty).Trim();
        if (name.Length > 0)
        {
            var tokens = name.Split(' ', StringSplitOptions.RemoveEmptyEntries);
            if (tokens.Length >= 2)
                return FirstRuneUppercase(tokens[0]) + FirstRuneUppercase(tokens[^1]);
            return TwoRunesUppercase(tokens[0]);
        }

        var fallback = (usernameFallback ?? string.Empty).Trim();
        if (fallback.Length > 0)
            return TwoRunesUppercase(fallback);
        return "??";
    }

    private static string FirstRuneUppercase(string s)
    {
        Rune.DecodeFromUtf16(s.AsSpan(), out var rune, out _);
        return Rune.ToUpperInvariant(rune).ToString();
    }

    private static string TwoRunesUppercase(string s)
    {
        var span = s.AsSpan();
        Rune.DecodeFromUtf16(span, out var first, out var consumed);
        span = span[consumed..];
        if (span.IsEmpty)
            return Rune.ToUpperInvariant(first).ToString();
        Rune.DecodeFromUtf16(span, out var second, out _);
        return Rune.ToUpperInvariant(first).ToString() + Rune.ToUpperInvariant(second).ToString();
    }
}
