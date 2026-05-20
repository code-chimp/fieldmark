using FieldMark.Data.Context;
using Microsoft.AspNetCore.Antiforgery;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;
using Testcontainers.PostgreSql;

namespace FieldMark.Tests.Web.Fixtures;

/// <summary>
/// Collection fixture that starts a real Postgres container, applies EF auth
/// migrations, and seeds roles + dev users. Shared across all tests in the
/// <see cref="AuthTests"/> collection.
/// </summary>
public sealed class PostgresFixture : IAsyncLifetime
{
    private readonly PostgreSqlContainer _postgres = new PostgreSqlBuilder("postgres:17")
        .WithDatabase("fieldmark_test")
        .WithUsername("fieldmark_test")
        .WithPassword("fieldmark_test")
        .Build();

    public string ConnectionString { get; private set; } = "";

    public async Task InitializeAsync()
    {
        await _postgres.StartAsync();
        ConnectionString = _postgres.GetConnectionString();

        // Apply EF auth migrations and seed data using a throw-away factory.
        using var factory = CreateFactory();
        using var scope = factory.Services.CreateScope();

        var authCtx = scope.ServiceProvider.GetRequiredService<AuthDbContext>();
        await authCtx.Database.MigrateAsync();

        await FieldMark.Web.SeedData.RoleSeeder.SeedAsync(
            scope.ServiceProvider, CancellationToken.None);

        var env = scope.ServiceProvider.GetRequiredService<IWebHostEnvironment>();
        await FieldMark.Web.SeedData.DevUsersSeeder.SeedAsync(
            scope.ServiceProvider, env, CancellationToken.None);
    }

    public async Task DisposeAsync() => await _postgres.DisposeAsync();

    /// <summary>Creates a <see cref="WebApplicationFactory{Program}"/> wired to the test DB.</summary>
    public WebApplicationFactory<Program> CreateFactory() =>
        new WebApplicationFactory<Program>().WithWebHostBuilder(b =>
        {
            b.UseSetting("ConnectionStrings:FieldMark", ConnectionString);
            // Clear env var so Program.cs takes the connection-string path.
            b.UseSetting("FIELDMARK_DATABASE_URL_OVERRIDE_FOR_TESTS", "1");
            b.UseEnvironment("Testing");
            // Override the DB URL env var by configuring a custom connection factory.
            b.ConfigureServices(services =>
            {
                // Replace both DbContexts with test-DB-pointing instances.
                ReplaceDbContext<FieldMarkDbContext>(services, ConnectionString);
                ReplaceDbContext<AuthDbContext>(services, ConnectionString);
                // Bypass antiforgery validation so tests don't need to extract tokens.
                services.AddSingleton<IAntiforgery, NoOpAntiforgery>();
            });
        });

    private static void ReplaceDbContext<T>(IServiceCollection services, string cs)
        where T : Microsoft.EntityFrameworkCore.DbContext
    {
        var descriptor = services.SingleOrDefault(d => d.ServiceType == typeof(T));
        if (descriptor != null) services.Remove(descriptor);

        var optDescriptor = services.SingleOrDefault(
            d => d.ServiceType == typeof(Microsoft.EntityFrameworkCore.DbContextOptions<T>));
        if (optDescriptor != null) services.Remove(optDescriptor);

        services.AddDbContext<T>(o =>
            o.UseNpgsql(cs).UseSnakeCaseNamingConvention());
    }
}

// Antiforgery no-op for integration tests — always validates, emits a predictable token.
internal sealed class NoOpAntiforgery : IAntiforgery
{
    private static readonly AntiforgeryTokenSet _tokens =
        new("test-token", "test-token", "__RequestVerificationToken", "test");

    public AntiforgeryTokenSet GetAndStoreTokens(HttpContext httpContext) => _tokens;
    public AntiforgeryTokenSet GetTokens(HttpContext httpContext) => _tokens;
    public Task<bool> IsRequestValidAsync(HttpContext httpContext) => Task.FromResult(true);
    public void SetCookieTokenAndHeader(HttpContext httpContext) { }
    public Task ValidateRequestAsync(HttpContext httpContext) => Task.CompletedTask;
}

[CollectionDefinition(AuthTests.Name)]
public sealed class AuthTests : ICollectionFixture<PostgresFixture>
{
    public const string Name = "Auth";
}
