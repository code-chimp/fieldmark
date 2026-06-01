using FieldMark.Data.Context;
using FieldMark.Data.Reference;
using FieldMark.Domain.ValueObjects;
using FieldMark.Web.Dashboard;
using FieldMark.Web.Authorization;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Identity;
using Microsoft.EntityFrameworkCore;
using Npgsql;

var builder = WebApplication.CreateBuilder(args);

// Add services to the container.
builder.Services.AddRazorPages(options =>
{
    options.Conventions.AllowAnonymousToPage("/Account/Login");
    options.Conventions.AllowAnonymousToPage("/Account/Logout");
    // Theme toggle must be callable while unauthenticated so it works on /login.
    options.Conventions.AllowAnonymousToFolder("/Preferences");
    options.Conventions.AddPageRoute("/Projects/Detail", "/projects/{id:guid}/tabs/{tab}");
});

// FIELDMARK_DATABASE_URL takes precedence when non-blank/whitespace.
// The value is trimmed before parsing so leading/trailing spaces do not cause
// URI parse failures. A postgres:// URI is converted to Npgsql key-value
// format: credentials are split before URL-decoding (safe for encoded colons),
// the database path segment is URL-decoded, and all query parameters are
// forwarded so the URL is a lossless source of connection configuration.
var rawDbUrl = Environment.GetEnvironmentVariable("FIELDMARK_DATABASE_URL")?.Trim();
string connectionString;
if (!string.IsNullOrWhiteSpace(rawDbUrl))
{
    var uri = new Uri(rawDbUrl);
    // Split on the raw UserInfo before URL-decoding so that an encoded colon
    // in the username (e.g. user%3Aname) is not treated as the separator.
    var rawParts = uri.UserInfo.Split(':', 2);
    var csb = new NpgsqlConnectionStringBuilder
    {
        Host = uri.Host,
        Port = uri.Port > 0 ? uri.Port : 5432,
        Database = Uri.UnescapeDataString(uri.AbsolutePath.TrimStart('/')),
        Username = rawParts.Length > 0 ? Uri.UnescapeDataString(rawParts[0]) : string.Empty,
        Password = rawParts.Length > 1 ? Uri.UnescapeDataString(rawParts[1]) : string.Empty,
    };
    // Forward all URL query parameters (sslmode, connect_timeout, application_name,
    // etc.) so the URL is a complete, lossless source of connection configuration.
    if (!string.IsNullOrEmpty(uri.Query))
    {
        foreach (
            var param in uri.Query.TrimStart('?').Split('&', StringSplitOptions.RemoveEmptyEntries)
        )
        {
            var kv = param.Split('=', 2);
            if (kv.Length == 2)
            {
                var key = Uri.UnescapeDataString(kv[0]);
                var value = Uri.UnescapeDataString(kv[1]);
                csb[key] = value;
            }
        }
    }
    connectionString = csb.ConnectionString;
}
else
{
    connectionString =
        builder.Configuration.GetConnectionString("FieldMark")
        ?? "Host=localhost;Database=fieldmark;Username=fieldmark;Password=fieldmark";
}

builder.Services.AddDbContext<FieldMarkDbContext>(options =>
    options.UseNpgsql(connectionString).UseSnakeCaseNamingConvention()
);

builder.Services.AddDbContext<AuthDbContext>(options =>
    options.UseNpgsql(connectionString).UseSnakeCaseNamingConvention()
);

// Audit helper — shares the request-scoped FieldMarkDbContext so the Append
// participates in the handler's transaction (FR39). No handler consumes it
// until Story 2.8 (ProjectCreated emission).
builder.Services.AddScoped<
    FieldMark.Data.Auditing.IAuditAppender,
    FieldMark.Data.Auditing.AuditAppender
>();
builder.Services.AddScoped<IReferenceReader, ReferenceReader>();
builder.Services.AddScoped<DashboardStatsReader>();

builder
    .Services.AddIdentityCore<IdentityUser<Guid>>(options =>
    {
        options.Password.RequireDigit = true;
        options.Password.RequireLowercase = true;
        options.Password.RequireUppercase = true;
        options.Password.RequireNonAlphanumeric = false;
        options.Password.RequiredLength = 10;
    })
    .AddRoles<IdentityRole<Guid>>()
    .AddEntityFrameworkStores<AuthDbContext>()
    .AddSignInManager()
    .AddDefaultTokenProviders();

builder
    .Services.AddAuthentication(IdentityConstants.ApplicationScheme)
    .AddCookie(
        IdentityConstants.ApplicationScheme,
        options =>
        {
            options.LoginPath = "/login";
            options.LogoutPath = "/logout";
            options.AccessDeniedPath = "/login";
            options.Events.OnRedirectToAccessDenied = async context =>
            {
                context.Response.StatusCode = StatusCodes.Status403Forbidden;
                await context.Response.WriteAsync(
                    "You do not have permission to access this page."
                );
            };
            options.ReturnUrlParameter = "return_url";
            options.ExpireTimeSpan = TimeSpan.FromDays(14);
            options.SlidingExpiration = true;
            options.Cookie.SameSite = SameSiteMode.Lax;
            options.Cookie.SecurePolicy = CookieSecurePolicy.SameAsRequest;
            options.Cookie.HttpOnly = true;
        }
    );

builder.Services.AddAuthorization(options =>
{
    options.FallbackPolicy = new AuthorizationPolicyBuilder().RequireAuthenticatedUser().Build();
});

// Register domain action → role permissions.
DomainPolicies.RegisterAction("project.create", Role.Admin);
// Story 2.9: project.read granted to all five roles (portfolio list visible to any authenticated user).
DomainPolicies.RegisterAction(
    "project.read",
    Role.Admin, Role.ComplianceOfficer, Role.Inspector, Role.SiteSupervisor, Role.Executive
);
DomainPolicies.RegisterAction("project.place_on_hold", Role.Admin);
DomainPolicies.RegisterAction("project.resume", Role.Admin);
DomainPolicies.RegisterAction("project.close", Role.Admin);
DomainPolicies.RegisterAction(
    "dashboard.view",
    Role.Admin, Role.ComplianceOfficer, Role.Inspector, Role.SiteSupervisor, Role.Executive
);

var app = builder.Build();

// Configure the HTTP request pipeline.
if (!app.Environment.IsDevelopment())
{
    app.UseExceptionHandler("/Error");
    // The default HSTS value is 30 days. You may want to change this for production scenarios, see https://aka.ms/aspnetcore-hsts.
    app.UseHsts();
}

app.UseHttpsRedirection();

app.UseStaticFiles();

app.UseRouting();

app.UseAuthentication();

app.UseAuthorization();

app.MapStaticAssets().AllowAnonymous();
app.MapRazorPages().WithStaticAssets();

if (args.Contains("--dump-routes"))
{
    FieldMark.Web.Tools.DumpRoutes.Run(app);
    return; // allows the host to dispose cleanly; no Environment.Exit needed
}

if (args.Contains("--seed-dev-users"))
{
    using var scope = app.Services.CreateScope();
    await scope.ServiceProvider.GetRequiredService<AuthDbContext>().Database.MigrateAsync();
    await FieldMark.Web.SeedData.RoleSeeder.SeedAsync(
        scope.ServiceProvider,
        CancellationToken.None
    );
    await FieldMark.Web.SeedData.DevUsersSeeder.SeedAsync(
        scope.ServiceProvider,
        app.Environment,
        CancellationToken.None
    );
    Console.WriteLine("✓ Roles and dev users seeded.");
    return;
}

using (var scope = app.Services.CreateScope())
{
    await scope.ServiceProvider.GetRequiredService<AuthDbContext>().Database.MigrateAsync();
    await FieldMark.Web.SeedData.RoleSeeder.SeedAsync(
        scope.ServiceProvider,
        CancellationToken.None
    );
    await FieldMark.Web.SeedData.DevUsersSeeder.SeedAsync(
        scope.ServiceProvider,
        app.Environment,
        CancellationToken.None
    );
}

await app.RunAsync();

// Expose Program to WebApplicationFactory in test projects.
public partial class Program { }
