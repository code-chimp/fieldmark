using Microsoft.AspNetCore.Mvc.RazorPages;
using Microsoft.AspNetCore.Routing;

namespace FieldMark.Web.Tools;

internal static class DumpRoutes
{
    internal static void Run(WebApplication app)
    {
        // IEndpointRouteBuilder.DataSources holds every endpoint source registered
        // via MapRazorPages(), MapGet(), etc. — populated at build time before app.Run().
        var builder = (IEndpointRouteBuilder)app;
        var endpoints = builder.DataSources.SelectMany(ds => ds.Endpoints).OfType<RouteEndpoint>();

        var lines = endpoints
            // Only include application Razor Page endpoints; static assets and error pages are excluded.
            .Where(ep => ep.Metadata.GetMetadata<PageActionDescriptor>() is not null)
            .SelectMany(ep =>
            {
                var descriptor = ep.Metadata.GetMetadata<PageActionDescriptor>()!;

                // Exclude framework internals: Error page and Admin area.
                if (
                    descriptor.ViewEnginePath.Equals("/Error", StringComparison.OrdinalIgnoreCase)
                    || descriptor.AreaName is not null
                )
                    return (IEnumerable<string>)[];

                var pattern = ep.RoutePattern.RawText ?? string.Empty;

                // Normalize to leading-slash, lowercase, no trailing slash.
                // An empty pattern is the root route (/).
                var path = pattern.Length == 0 ? "/" : "/" + pattern.TrimStart('/');
                if (path.Length > 1)
                    path = path.TrimEnd('/');
                path = path.ToLowerInvariant();

                // Suppress /index alias — the root page is already emitted as /.
                if (path.EndsWith("/index", StringComparison.Ordinal))
                    return [];

                return (IEnumerable<string>)[$"get {path}"];
            })
            .Distinct()
            .OrderBy(l => l)
            .ToList();

        foreach (var line in lines)
            Console.WriteLine(line);
    }
}
