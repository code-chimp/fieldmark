using FieldMark.Domain.Services;
using FluentAssertions;

namespace FieldMark.Tests.Domain.Services;

public class AvatarInitialsTests
{
    [Fact]
    public void EmptyFullName_FallsBackToUsername_FirstTwoChars()
    {
        AvatarInitials.From("", "alice").Should().Be("AL");
    }

    [Fact]
    public void NullFullName_FallsBackToUsername_FirstTwoChars()
    {
        AvatarInitials.From(null, "bob").Should().Be("BO");
    }

    [Fact]
    public void SingleTokenFullName_FirstTwoCharsUppercased()
    {
        AvatarInitials.From("Alice", "alice").Should().Be("AL");
    }

    [Fact]
    public void TwoTokenFullName_FirstAndLastInitialsUppercased()
    {
        AvatarInitials.From("Alice Admin", "alice").Should().Be("AA");
    }

    [Fact]
    public void ThreePlusTokenFullName_FirstAndLastInitials()
    {
        AvatarInitials.From("Alice Marie Admin", "alice").Should().Be("AA");
    }

    [Fact]
    public void UnicodeCharactersPreservedAsIs()
    {
        AvatarInitials.From("Ää Öö", "aao").Should().Be("ÄÖ");
    }

    [Fact]
    public void SingleTokenUnicode_FirstTwoChars()
    {
        AvatarInitials.From("李明", "user").Should().Be("李明");
    }

    [Fact]
    public void NonBmpLeadingCodePoint_TwoTokenName_NotSplitAtSurrogate()
    {
        // 𝕳 is U+1D573 (non-BMP) encoded as surrogate pair — must not yield a lone surrogate
        AvatarInitials.From("𝕳ello World", "user").Should().Be("𝕳W");
    }

    [Fact]
    public void BothEmpty_ReturnsFallbackToken()
    {
        AvatarInitials.From(null, null).Should().Be("??");
    }

    [Fact]
    public void EmptyFullNameAndEmptyUsername_ReturnsFallbackToken()
    {
        AvatarInitials.From("", "").Should().Be("??");
    }
}
