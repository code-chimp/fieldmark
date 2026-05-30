using System.Reflection;
using Microsoft.AspNetCore.Mvc.Infrastructure;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace FieldMark.Web.Tools;

internal static class DumpRoutes
{
    private static readonly string[] HttpMethodNames =
    [
        "GET",
        "POST",
        "PUT",
        "DELETE",
        "PATCH",
        "HEAD",
        "OPTIONS",
    ];

    internal static void Run(WebApplication app)
    {
        // IActionDescriptorCollectionProvider provides fully resolved PageActionDescriptors.
        var provider = app.Services.GetRequiredService<IActionDescriptorCollectionProvider>();

        var lines = provider
            .ActionDescriptors.Items.OfType<PageActionDescriptor>()
            .SelectMany(descriptor =>
            {
                // Exclude framework internals: Error page and Admin area.
                if (
                    descriptor.ViewEnginePath.Equals("/Error", StringComparison.OrdinalIgnoreCase)
                    || descriptor.AreaName is not null
                )
                    return (IEnumerable<string>)[];

                // Prefer explicit route template (e.g. "@page "/preferences/theme""),
                // otherwise derive the path from the view engine path.
                var rawPath = descriptor.AttributeRouteInfo?.Template ?? descriptor.ViewEnginePath;
                var path = rawPath.Length == 0 ? "/" : "/" + rawPath.TrimStart('/');
                if (path.Length > 1)
                    path = path.TrimEnd('/');
                path = path.ToLowerInvariant();
                // Normalize Razor Page path params {name:type} or {name} → :name
                // for cross-stack parity with Django (<type:name>) and Fiber (:name).
                path = System.Text.RegularExpressions.Regex.Replace(
                    path,
                    @"\{([a-z][a-z0-9_]*)(?::[^}]*)?\}",
                    m => ":" + m.Groups[1].Value
                );

                // Suppress /index alias — root page already emitted as /.
                if (path.EndsWith("/index", StringComparison.Ordinal))
                    return [];

                // Derive HTTP methods by reflecting on the PageModel type: look for
                // public instance methods matching On{Method}[Async] (e.g. OnPost, OnGetAsync).
                var methods = HttpMethodsFromPageModel(descriptor.ViewEnginePath);
                if (methods.Count == 0)
                    return (IEnumerable<string>)[$"get {path}"];

                return methods.Select(m => $"{m} {path}");
            })
            .Distinct()
            .OrderBy(l => l)
            .ToList();

        foreach (var line in lines)
            Console.WriteLine(line);
    }

    private static List<string> HttpMethodsFromPageModel(string viewEnginePath)
    {
        // Resolve the PageModel type by naming convention:
        // "/Preferences/Theme" → "FieldMark.Web.Pages.Preferences.ThemeModel"
        // HttpMethodMetadata is not populated on Razor Page endpoints in .NET 10;
        // the On{Method}[Async] handler naming convention is a stable Razor Pages
        // contract (unchanged since ASP.NET Core 2.0).
        var segments = viewEnginePath.TrimStart('/').Split('/');
        var typeName =
            "FieldMark.Web.Pages."
            + string.Join('.', segments.Take(segments.Length - 1).Append(segments[^1] + "Model"));

        // Search all loaded assemblies so the dumper works correctly if pages are
        // ever split into a separate assembly; verify the resolved type is a PageModel
        // to rule out accidental name collisions.
        var type = AppDomain
            .CurrentDomain.GetAssemblies()
            .Select(a => a.GetType(typeName))
            .FirstOrDefault(t => t is not null && typeof(PageModel).IsAssignableFrom(t));

        if (type is null)
            return [];

        var result = new HashSet<string>(StringComparer.OrdinalIgnoreCase);
        foreach (var method in type.GetMethods(BindingFlags.Public | BindingFlags.Instance))
        {
            var name = method.Name;
            // Match On{HttpMethod}[Async] or On{HttpMethod}{Handler}[Async]
            if (!name.StartsWith("On", StringComparison.Ordinal))
                continue;
            var withoutOn = name[2..];
            if (withoutOn.EndsWith("Async", StringComparison.Ordinal))
                withoutOn = withoutOn[..^5];
            // Extract the leading HTTP method segment (GET, POST, etc.)
            foreach (var http in HttpMethodNames)
            {
                if (withoutOn.StartsWith(http, StringComparison.OrdinalIgnoreCase))
                {
                    result.Add(http.ToLowerInvariant());
                    break;
                }
            }
        }

        return [.. result];
    }
}
