using System.Dynamic;
using FieldMark.Tests.Web.Fixtures;
using FieldMark.Tests.Web.Helpers;
using FluentAssertions;
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

public abstract class ComponentRenderFixture(PostgresFixture pg)
{
    protected static string CanonicalExamplePath(string component)
    {
        var startPath = AppContext.BaseDirectory;
        var dir = new DirectoryInfo(startPath);
        while (dir != null && !File.Exists(Path.Combine(dir.FullName, "Makefile")))
            dir = dir.Parent;
        if (dir is null)
            throw new InvalidOperationException(
                $"Repo root not found while walking up from '{startPath}'"
            );
        return Path.Combine(
            dir.FullName,
            "fieldmark_shared",
            "components",
            component,
            "canonical.html"
        );
    }

    protected static ExpandoObject Model(params (string Key, object? Value)[] values)
    {
        IDictionary<string, object?> model = new ExpandoObject();
        foreach (var (key, value) in values)
            model[key] = value;
        return (ExpandoObject)model;
    }

    protected async Task<string> RenderPartial(string partialName, object model)
    {
        using var scope = pg.CreateFactory().Services.CreateScope();
        var sp = scope.ServiceProvider;
        var httpContext = new DefaultHttpContext { RequestServices = sp };
        var actionContext = new ActionContext(httpContext, new RouteData(), new ActionDescriptor());
        var viewEngine = sp.GetRequiredService<ICompositeViewEngine>();
        var viewPath = $"/Pages/Shared/Components/_{partialName}.cshtml";
        var viewResult = viewEngine.GetView(executingFilePath: null, viewPath, isMainPage: false);
        viewResult
            .Success.Should()
            .BeTrue($"{partialName} partial must be found by the view engine");

        var viewData = new ViewDataDictionary<object>(
            new EmptyModelMetadataProvider(),
            new ModelStateDictionary()
        )
        {
            Model = model,
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

    protected static string RepoPath(params string[] parts)
    {
        var startPath = AppContext.BaseDirectory;
        var dir = new DirectoryInfo(startPath);
        while (dir != null && !File.Exists(Path.Combine(dir.FullName, "Makefile")))
            dir = dir.Parent;
        if (dir is null)
            throw new InvalidOperationException(
                $"Repo root not found while walking up from '{startPath}'"
            );
        return Path.Combine(new[] { dir.FullName }.Concat(parts).ToArray());
    }

    protected async Task AssertSnapshot(
        string component,
        string partialName,
        string variant,
        object model
    )
    {
        var actual = NormaliseHtml.NormaliseComponent(await RenderPartial(partialName, model));
        var canonical = NormaliseHtml.ExtractVariant(
            await File.ReadAllTextAsync(CanonicalExamplePath(component)),
            variant
        );
        actual.Should().Be(canonical, $"{component} {variant} must match the canonical fixture");
    }
}
