using FieldMark.Data.Context;
using Microsoft.EntityFrameworkCore;
using Npgsql;

var builder = WebApplication.CreateBuilder(args);

// Add services to the container.
builder.Services.AddRazorPages();

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
        Host     = uri.Host,
        Port     = uri.Port > 0 ? uri.Port : 5432,
        Database = Uri.UnescapeDataString(uri.AbsolutePath.TrimStart('/')),
        Username = rawParts.Length > 0 ? Uri.UnescapeDataString(rawParts[0]) : string.Empty,
        Password = rawParts.Length > 1 ? Uri.UnescapeDataString(rawParts[1]) : string.Empty,
    };
    // Forward all URL query parameters (sslmode, connect_timeout, application_name,
    // etc.) so the URL is a complete, lossless source of connection configuration.
    if (!string.IsNullOrEmpty(uri.Query))
    {
        foreach (var param in uri.Query.TrimStart('?').Split('&', StringSplitOptions.RemoveEmptyEntries))
        {
            var kv = param.Split('=', 2);
            if (kv.Length == 2)
            {
                var key   = Uri.UnescapeDataString(kv[0]);
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
{
    options.UseNpgsql(connectionString);
});

var app = builder.Build();

// Configure the HTTP request pipeline.
if (!app.Environment.IsDevelopment())
{
    app.UseExceptionHandler("/Error");
    // The default HSTS value is 30 days. You may want to change this for production scenarios, see https://aka.ms/aspnetcore-hsts.
    app.UseHsts();
}

app.UseHttpsRedirection();

app.UseRouting();

app.UseAuthorization();

app.MapStaticAssets();
app.MapRazorPages().WithStaticAssets();

app.Run();
