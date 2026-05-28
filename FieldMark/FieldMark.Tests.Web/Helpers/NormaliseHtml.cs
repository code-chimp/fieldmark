using System.Text.RegularExpressions;

namespace FieldMark.Tests.Web.Helpers;

/// <summary>
/// Whitespace normalisation and per-stack noise stripping for cross-stack HTML snapshot tests.
/// Strips antiforgery tokens and csrfmiddlewaretoken inputs before comparison so only the
/// canonical form markup is compared across .NET and Django.
/// </summary>
public static partial class NormaliseHtml
{
    [GeneratedRegex(@"\s+")]
    private static partial Regex WhitespaceRun();

    [GeneratedRegex(
        @"<input[^>]*name=""__RequestVerificationToken""[^>]*>",
        RegexOptions.IgnoreCase
    )]
    private static partial Regex AntiforgeryInput();

    [GeneratedRegex(@"<input[^>]*name=""csrfmiddlewaretoken""[^>]*>", RegexOptions.IgnoreCase)]
    private static partial Regex CsrfInput();

    [GeneratedRegex(
        @"<form[^>]*id=""login-form""[^>]*>.*?</form>",
        RegexOptions.Singleline | RegexOptions.IgnoreCase
    )]
    private static partial Regex LoginFormBlock();

    [GeneratedRegex(
        @"<div[^>]*id=""login-errors""[^>]*>.*?</div>",
        RegexOptions.Singleline | RegexOptions.IgnoreCase
    )]
    private static partial Regex LoginErrorBlock();

    /// <summary>
    /// Extracts and normalises the <c>&lt;form id="login-form"&gt;...&lt;/form&gt;</c> block.
    /// Strips per-stack antiforgery noise before comparison.
    /// </summary>
    public static string ExtractLoginForm(string html)
    {
        var match = LoginFormBlock().Match(html);
        return match.Success ? Normalise(match.Value) : "";
    }

    /// <summary>
    /// Extracts and normalises the <c>&lt;div id="login-errors"&gt;...&lt;/div&gt;</c> block.
    /// </summary>
    public static string ExtractLoginErrorRegion(string html)
    {
        var match = LoginErrorBlock().Match(html);
        return match.Success ? Normalise(match.Value) : "";
    }

    [GeneratedRegex(@"<!--.*?-->", RegexOptions.Singleline)]
    private static partial Regex HtmlComment();

    /// <summary>
    /// Extracts the content of a named variant block from a canonical component example file.
    /// Blocks are delimited by <c>&lt;!-- variant: name ... --&gt;</c> comment lines.
    /// </summary>
    public static string ExtractVariant(string exampleFileContent, string variantName)
    {
        var lines = exampleFileContent.Split('\n');
        var startMarker = $"<!-- variant: {variantName}";
        var inBlock = false;
        var sb = new System.Text.StringBuilder();

        foreach (var line in lines)
        {
            var trimmed = line.TrimEnd();
            if (
                string.Equals(trimmed, $"{startMarker} -->", StringComparison.OrdinalIgnoreCase)
                || trimmed.StartsWith($"{startMarker} ", StringComparison.OrdinalIgnoreCase)
            )
            {
                inBlock = true;
                continue;
            }
            if (inBlock && trimmed.StartsWith("<!-- variant:", StringComparison.OrdinalIgnoreCase))
                break;
            if (inBlock)
                sb.AppendLine(trimmed);
        }

        return NormaliseComponent(sb.ToString());
    }

    /// <summary>
    /// Normalises component HTML for cross-stack snapshot comparison:
    /// strips HTML comments, collapses whitespace, trims.
    /// </summary>
    public static string NormaliseComponent(string html)
    {
        html = HtmlComment().Replace(html, "");
        html = html.Replace("&#34;", "&quot;", StringComparison.Ordinal);
        html = html.Replace("&#x2014;", "—", StringComparison.OrdinalIgnoreCase);
        html = WhitespaceRun().Replace(html, " ").Trim();
        return html;
    }

    private static string Normalise(string html)
    {
        // Strip per-stack antiforgery noise — excluded from the snapshot contract.
        html = AntiforgeryInput().Replace(html, "");
        html = CsrfInput().Replace(html, "");

        // Collapse whitespace runs (including newlines) to a single space.
        html = WhitespaceRun().Replace(html, " ").Trim();

        return html;
    }
}
