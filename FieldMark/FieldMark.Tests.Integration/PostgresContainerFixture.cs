using System.IO;
using System.Threading.Tasks;
using DotNet.Testcontainers.Configurations;
using Testcontainers.PostgreSql;

namespace FieldMark.Tests.Integration;

// Spins up a Postgres 17 container for the test assembly and runs the canonical
// init scripts from docker/postgres/init/ — the same scripts the live compose
// stack uses — so domain.* DDL and seed data come from a single source of truth.
//
// Container lifecycle is class-scoped via the xUnit ICollectionFixture wired up
// in PostgresCollection.cs; tests that need a real database join that collection
// and inject this fixture for the connection string.
public sealed class PostgresContainerFixture : IAsyncLifetime
{
    private const string Database = "fieldmark";
    private const string Username = "fieldmark";
    private const string Password = "fieldmark";

    private readonly PostgreSqlContainer _container;

    public PostgresContainerFixture()
    {
        var initDir = LocateInitDir();

        // Bind-mount the canonical init directory exactly the way docker-compose
        // does — Postgres runs every .sql file in /docker-entrypoint-initdb.d on
        // first boot, so the harness picks up new init scripts automatically.
        // Read-only because nothing here should mutate the host scripts.
        _container = new PostgreSqlBuilder("postgres:17")
            .WithDatabase(Database)
            .WithUsername(Username)
            .WithPassword(Password)
            .WithBindMount(initDir, "/docker-entrypoint-initdb.d", AccessMode.ReadOnly)
            .Build();
    }

    public string ConnectionString => _container.GetConnectionString();

    public Task InitializeAsync() => _container.StartAsync();

    public Task DisposeAsync() => _container.DisposeAsync().AsTask();

    // Walk up from the test bin/ directory until we find the repo root marker.
    // Cleaner than embedding the init scripts as project resources because the
    // canonical location is docker/postgres/init/ and it must not drift.
    private static string LocateInitDir()
    {
        var dir = new DirectoryInfo(Directory.GetCurrentDirectory());
        while (dir is not null)
        {
            var candidate = Path.Combine(dir.FullName, "docker", "postgres", "init");
            if (Directory.Exists(candidate))
            {
                return candidate;
            }
            dir = dir.Parent;
        }
        throw new DirectoryNotFoundException(
            "Could not locate docker/postgres/init relative to the test working directory.");
    }
}

[CollectionDefinition(Name)]
public sealed class PostgresCollection : ICollectionFixture<PostgresContainerFixture>
{
    public const string Name = "Postgres";
}
