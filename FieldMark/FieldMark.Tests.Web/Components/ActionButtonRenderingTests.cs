using FieldMark.Tests.Web.Fixtures;
using FieldMark.Tests.Web.Helpers;
using FieldMark.Web.ViewModels.Components;
using FluentAssertions;
using HtmlAgilityPack;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.Abstractions;
using Microsoft.AspNetCore.Mvc.ModelBinding;
using Microsoft.AspNetCore.Mvc.Rendering;
using Microsoft.AspNetCore.Mvc.ViewEngines;
using Microsoft.AspNetCore.Mvc.ViewFeatures;
using Microsoft.AspNetCore.Routing;
using Microsoft.Extensions.DependencyInjection;

namespace FieldMark.Tests.Web.Components;

[Collection(AuthTests.Name)]
public sealed class ActionButtonRenderingTests(PostgresFixture pg)
{
    // Canonical fixture values matching action_button.example.html.
    private static readonly ActionButtonVm FixtureDisabled = new(
        Id: "ab-fixture-1",
        Permission: true,
        StateAllows: false,
        Label: "Approve Resolution",
        HxPost: "/violations/00000000-0000-0000-0000-000000000001/corrective-actions/00000000-0000-0000-0000-000000000002/approve",
        HxTarget: "#violation-detail",
        DisabledReason: "Awaiting review"
    );

    private static readonly ActionButtonVm FixturePresent = FixtureDisabled with
    {
        StateAllows = true,
    };
    private static readonly ActionButtonVm FixtureAbsent = FixtureDisabled with
    {
        Permission = false,
    };

    private static string CanonicalExamplePath()
    {
        var dir = new DirectoryInfo(AppContext.BaseDirectory);
        while (dir != null && !File.Exists(Path.Combine(dir.FullName, "Makefile")))
            dir = dir.Parent;
        return Path.Combine(
            dir?.FullName ?? throw new InvalidOperationException("Repo root not found"),
            "fieldmark_shared",
            "components",
            "action_button.example.html"
        );
    }

    private async Task<string> RenderPartial(ActionButtonVm vm)
    {
        using var scope = pg.CreateFactory().Services.CreateScope();
        var sp = scope.ServiceProvider;

        var httpContext = new DefaultHttpContext { RequestServices = sp };
        var actionContext = new ActionContext(httpContext, new RouteData(), new ActionDescriptor());

        var viewEngine = sp.GetRequiredService<ICompositeViewEngine>();
        var viewResult = viewEngine.FindView(actionContext, "_ActionButton", isMainPage: false);
        viewResult
            .Success.Should()
            .BeTrue(because: "_ActionButton partial must be found by the view engine");

        var viewData = new ViewDataDictionary<ActionButtonVm>(
            new EmptyModelMetadataProvider(),
            new ModelStateDictionary()
        )
        {
            Model = vm,
        };
        var tempData = sp.GetRequiredService<ITempDataDictionaryFactory>().GetTempData(httpContext);

        using var writer = new StringWriter();
        var viewContext = new ViewContext(
            actionContext,
            viewResult.View!,
            viewData,
            tempData,
            writer,
            new HtmlHelperOptions()
        );

        await viewResult.View!.RenderAsync(viewContext);
        return writer.ToString();
    }

    [Fact]
    public async Task ActionButton_PermissionFalse_RendersEmpty()
    {
        var html = NormaliseHtml.NormaliseComponent(await RenderPartial(FixtureAbsent));
        html.Should()
            .BeEmpty(because: "absent variant must produce empty output when permission=false");
    }

    [Fact]
    public async Task ActionButton_DisabledVariant_MatchesCanonicalSnapshot()
    {
        var actual = NormaliseHtml.NormaliseComponent(await RenderPartial(FixtureDisabled));
        var canonical = NormaliseHtml.ExtractVariant(
            await File.ReadAllTextAsync(CanonicalExamplePath()),
            "disabled"
        );

        actual
            .Should()
            .Be(canonical, because: "disabled variant must be byte-identical to canonical example");
    }

    [Fact]
    public async Task ActionButton_PresentVariant_MatchesCanonicalSnapshot()
    {
        var actual = NormaliseHtml.NormaliseComponent(await RenderPartial(FixturePresent));
        var canonical = NormaliseHtml.ExtractVariant(
            await File.ReadAllTextAsync(CanonicalExamplePath()),
            "present"
        );

        actual
            .Should()
            .Be(canonical, because: "present variant must be byte-identical to canonical example");
    }

    [Fact]
    public async Task ActionButton_DisabledVariant_HasScreenReaderReason()
    {
        var html = await RenderPartial(FixtureDisabled);

        var doc = new HtmlDocument();
        doc.LoadHtml(html);

        var button = doc.DocumentNode.SelectSingleNode("//button");
        button.Should().NotBeNull(because: "disabled variant must render a button element");
        button!
            .Attributes["disabled"]
            .Should()
            .NotBeNull(because: "button must carry the disabled attribute");
        button.GetAttributeValue("aria-disabled", "").Should().Be("true");
        button.GetAttributeValue("tabindex", "").Should().Be("0");
        button.GetAttributeValue("data-tooltip", "").Should().Be("Awaiting review");

        var describedBy = button.GetAttributeValue("aria-describedby", "");
        describedBy.Should().Be("ab-fixture-1-reason");

        var srSpan = doc.DocumentNode.SelectSingleNode($"//span[@id='{describedBy}']");
        srSpan.Should().NotBeNull(because: "sr-only reason span must be present in the DOM");
        srSpan!.GetAttributeValue("class", "").Should().Contain("sr-only");
        srSpan.InnerText.Trim().Should().Be("Awaiting review");
    }
}
