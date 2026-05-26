# Story 2.2: Map `domain.audit_entry` and provide a per-stack `append_audit_entry()` helper

Status: ready-for-dev

Epic: 2 — Project Lifecycle & Compliance Dashboard
Source AC: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.2
Canonical DDL: [docker/postgres/init/010_domain_tables.sql:190–211](../../docker/postgres/init/010_domain_tables.sql)
Canonical contract doc (to populate): [docs/reference/audit-actions.md](../../docs/reference/audit-actions.md) — currently a skeleton (Epic 1 retro action item A4); this story owns the population.

## Story

As a handler author across all three stacks,
I want a single helper per stack that appends an `AuditEntry` row inside the surrounding DB transaction, using a canonical, stack-native enum/constants set whose values are pinned to a single documentation contract,
So that FR39 (audit-on-every-mutation) and FR40 (no inventing action variants) are mechanically satisfied for every transition Epic 2+ introduces, with cross-stack drift impossible.

**Scope boundary:** this story produces (a) the `AuditEntry` mapping per stack, (b) the `append_audit_entry()` helper per stack, (c) the canonical audit-action enum/constants per stack, (d) the populated `docs/reference/audit-actions.md` contract, and (e) per-stack conformance + transactional tests. **Out of scope:** any consuming handler call (Story 2.8 emits the first real `ProjectCreated`), the project audit-log tab UI (Story 2.13), the `AuditRow` component (Story 2.4). Do not pre-wire `Deps.AuditEntries` into a handler; just expose the helper so 2.8 can pick it up.

## Acceptance Criteria

### AC1 — Canonical contract populated at `docs/reference/audit-actions.md`

**Given** the Cross-Stack Architecture Principle (root [CLAUDE.md](../../CLAUDE.md) §Cross-Stack Architecture Principle) and Epic 1 retro action item A4
**When** I open [docs/reference/audit-actions.md](../../docs/reference/audit-actions.md)
**Then** the skeleton TODOs are replaced with:

- A **Canonical Action List** table containing **every** audit-action string emitted in the MVP, with columns: `Action`, `Entity`, `Emitted when` (1-line trigger), `Story that introduces emission`, `Notes`. The list MUST be derivable from [architecture.md:603](../planning-artifacts/architecture.md) (line: "Audit action strings (canonical): … ") plus the `ProjectCreated` ADR amendment recorded in the epic file at [epic-2 §Story 2.8 note](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md). **Reconcile the count:** the architecture line enumerates 13 strings; the epic + audit-actions skeleton say "14 + `ProjectCreated` = 15 total". Resolve by cross-walking PRD FR9–FR15 + Epic 4/5/6 story emissions, picking the missing string (most-likely candidate: `ViolationResolved` per FR-driven closure flow). Document the resolution rationale in the doc's "Change Procedure" section so the next reviewer can audit the call.
- A **Casing Convention** section stating: strings are `PascalCase`, present-tense past-form (e.g. `ProjectPlacedOnHold`, not `placeOnHold` / `PROJECT_PLACED_ON_HOLD`). The convention is binding across stacks; per-stack symbol names may differ (`AuditAction.ProjectPlacedOnHold` in C#, `AuditAction.PROJECT_PLACED_ON_HOLD` constant in Python, `AuditActionProjectPlacedOnHold` const in Go) but the **persisted string** in `domain.audit_entry.action` is the PascalCase form verbatim.
- A **Per-Stack Native Implementations** section listing the file path of each stack's enum/constants module (per AC2/AC3/AC4 below).
- A **Conformance Test Contract** section specifying: each stack ships a unit test that reads the canonical action list from a checked-in JSON fixture at `docs/reference/audit-actions.json` (see AC6) and asserts the stack's native set matches exactly — no extras, no missing. Test names: `.NET ` `AuditActionConformanceTests.cs`, Django `audit/tests/test_action_conformance.py`, Go `audit_action_conformance_test.go`.
- A **Change Procedure** subsection: adding/removing an action requires (1) an ADR amendment recorded in the epic file *and* this doc, (2) updating `audit-actions.json`, (3) re-running the three conformance tests, (4) green `make parity` (action-string drift is a parity concern even though `pg_indexes` won't catch it — flag in dev notes if A1's review-churn analysis adds a parity script for this).

**Given** the doc is populated
**When** I read it
**Then** it is the **single source of truth** — no shared code package, no symlinked manifest, no generated stubs. Each stack's native enum is the implementation, the JSON fixture is the conformance gate, this doc is the contract.

### AC2 — .NET mapping + helper + enum

**Given** the .NET stack
**When** I inspect `FieldMark/FieldMark.Domain/Entities/`
**Then** `AuditEntry.cs` exists as an immutable property bag with:
- properties: `Id (Guid)`, `OccurredAt (DateTimeOffset)`, `ActorId (Guid)`, `Action (string)` — stored as `string` (not the enum) so unrecognized DB values surface as a deserialization failure on read rather than corrupting the row — `EntityType (string)`, `EntityId (Guid)`, `ProjectId (Guid?)`, `BeforeState (JsonDocument?)`, `AfterState (JsonDocument?)`, `Metadata (JsonDocument?)`
- a public constructor taking all fields except `Id` / `OccurredAt` (both server-defaulted: `Id` = `Guid.NewGuid()` at construction; `OccurredAt` defaults to DDL `now()` via the EF Core `ValueGeneratedOnAdd()` configuration — see below)
- a parameterless private constructor for EF Core
- no behavior methods. `AuditEntry` is write-once value object per [architecture.md:1038](../planning-artifacts/architecture.md).

**Given** the .NET stack
**When** I inspect `FieldMark/FieldMark.Domain/ValueObjects/AuditAction.cs`
**Then** an `enum AuditAction` is declared with one member per canonical action (PascalCase symbol matching the persisted string). The top-of-file XML doc comment references `docs/reference/audit-actions.md` as the source of truth. An `AuditActionExtensions.AsString(this AuditAction action)` returns the symbol's name via `nameof`-style mapping (a static `Dictionary<AuditAction, string>` is acceptable; a `switch` expression is acceptable; `Enum.GetName` is acceptable as long as the test in AC6 proves all 15 entries round-trip).

**Given** the .NET stack
**When** I inspect `FieldMark/FieldMark.Data/Configuration/AuditEntryConfiguration.cs`
**Then** it:
- `builder.ToTable("audit_entry", "domain")`
- `builder.HasKey(a => a.Id)`
- `builder.Property(a => a.OccurredAt).HasDefaultValueSql("now()").ValueGeneratedOnAdd()` — the DDL has `DEFAULT now()`; this lets EF Core know not to send the value on insert when the property is unset and to read back the server-assigned value
- `builder.Property(a => a.Action).HasMaxLength(64).IsRequired()`
- `builder.Property(a => a.EntityType).HasMaxLength(64).IsRequired()`
- JSONB columns: declare via `builder.Property(a => a.BeforeState).HasColumnType("jsonb")` — Npgsql maps `System.Text.Json.JsonDocument` to `jsonb` natively. Do **not** introduce Newtonsoft.Json.
- **No `HasIndex` calls** — the DDL declares `idx_audit_entity` and `idx_audit_project` already. Mirror them in the model with `builder.HasIndex(...).HasDatabaseName("idx_audit_entity")` and `.IsDescending(false, true)` only if EF Core's "phantom-index" detection complains during `make parity`. Settle this empirically: run `make parity` after wiring; if the diff is zero, leave the indexes purely DDL-owned (preferred — same pattern as Story 2.1's resolution of `Project.code UNIQUE`).
- `builder.Property(a => a.ProjectId).IsRequired(false)` — nullable per DDL.

**Given** the .NET stack
**When** I inspect `FieldMark/FieldMark.Data/Auditing/AuditAppender.cs`
**Then** it exposes a stateless service:

```csharp
public interface IAuditAppender
{
    void Append(
        Guid actorId,
        AuditAction action,
        string entityType,
        Guid entityId,
        Guid? projectId = null,
        JsonDocument? beforeState = null,
        JsonDocument? afterState = null,
        JsonDocument? metadata = null);
}

public sealed class AuditAppender(FieldMarkDbContext db) : IAuditAppender { ... }
```

The implementation calls `db.AuditEntries.Add(new AuditEntry(actorId, action.AsString(), entityType, entityId, projectId, beforeState, afterState, metadata))`. **It does not call `SaveChangesAsync` and does not open a transaction.** The surrounding handler owns transaction lifecycle (per [architecture.md §Canonical Request Flow:733–746](../planning-artifacts/architecture.md)). Registered as `services.AddScoped<IAuditAppender, AuditAppender>()` in `FieldMark.Web/Program.cs` so it shares the request-scoped `FieldMarkDbContext`.

**DbContext wiring:** add `DbSet<AuditEntry> AuditEntries` to `FieldMarkDbContext`. The `OnModelCreating` change from Story 2.1 (`ApplyConfigurationsFromAssembly`) picks up the new configuration automatically.

### AC3 — Django mapping + helper + constants

**Given** the Django stack
**When** I inspect `fieldmark_py/audit/models.py`
**Then** `AuditEntry` is declared with:
- `Meta.managed = False`, `Meta.db_table = 'domain"."audit_entry'` (embedded double-quote pattern matching Story 2.1)
- fields matching DDL: `id = UUIDField(primary_key=True, default=uuid.uuid4)`, `occurred_at = DateTimeField()` (no `default` — DB sets it; `Meta.managed = False` means migrations don't apply anyway, but be explicit), `actor_id = UUIDField()`, `action = CharField(max_length=64)`, `entity_type = CharField(max_length=64)`, `entity_id = UUIDField()`, `project_id = UUIDField(null=True)` (declared as raw UUID, not `ForeignKey(Project)` — keeping the audit module free of an import dependency on `projects/`; the DB FK is DDL-owned), `before_state = models.JSONField(null=True)`, `after_state = models.JSONField(null=True)`, `metadata = models.JSONField(null=True)`.
- `class Meta` also declares `default_permissions = ()` to suppress Django's automatic CRUD permission rows for an unmanaged table.

**Given** the Django stack
**When** I inspect `fieldmark_py/audit/actions.py`
**Then** `AuditAction` is a `class AuditAction(models.TextChoices)` with one member per canonical action. The choice value is the persisted PascalCase string; the symbol name is `SCREAMING_SNAKE_CASE` per Python convention (e.g. `PROJECT_PLACED_ON_HOLD = "ProjectPlacedOnHold", "ProjectPlacedOnHold"`). Top-of-file docstring references `docs/reference/audit-actions.md`.

**Given** the Django stack
**When** I inspect `fieldmark_py/audit/append.py`
**Then** it exposes:

```python
def append_audit_entry(
    *,
    actor_id: uuid.UUID,
    action: AuditAction,
    entity_type: str,
    entity_id: uuid.UUID,
    project_id: uuid.UUID | None = None,
    before_state: dict | None = None,
    after_state: dict | None = None,
    metadata: dict | None = None,
) -> AuditEntry:
    """Append an AuditEntry inside the caller's open `transaction.atomic()` block.

    Caller is responsible for the transaction; this function only issues the INSERT.
    """
    return AuditEntry.objects.create(
        actor_id=actor_id,
        action=action.value,    # persist the PascalCase string verbatim — FR40
        entity_type=entity_type,
        entity_id=entity_id,
        project_id=project_id,
        before_state=before_state,
        after_state=after_state,
        metadata=metadata,
    )
```

Keyword-only signature (the `*,` is intentional — six similar UUID/JSON arguments would silently mis-bind positionally). The function MUST NOT call `transaction.atomic()` itself; callers wrap with `with transaction.atomic():` per [architecture.md:802–831](../planning-artifacts/architecture.md).

### AC4 — Go mapping + helper + constants

**Given** the Go stack
**When** I inspect `fieldmark-go/internal/domain/entities/audit_entry.go`
**Then** `AuditEntry` is a plain struct with exported fields: `ID uuid.UUID`, `OccurredAt time.Time`, `ActorID uuid.UUID`, `Action string`, `EntityType string`, `EntityID uuid.UUID`, `ProjectID *uuid.UUID` (nullable), `BeforeState json.RawMessage` (nullable — `nil` if no payload), `AfterState json.RawMessage`, `Metadata json.RawMessage`. No methods this story.

**Given** the Go stack
**When** I inspect `fieldmark-go/internal/domain/enums/audit_action.go`
**Then** a string-typed `type AuditAction string` is declared with one `const` per canonical action — `AuditActionProjectPlacedOnHold AuditAction = "ProjectPlacedOnHold"`, etc. A `var AllAuditActions = []AuditAction{...}` exhaustive slice is declared so the conformance test in AC6 has a single iteration target. Top-of-file `// audit-actions.md` doc comment.

**Given** the Go stack
**When** I inspect `fieldmark-go/internal/data/postgres/auditentrystore.go`
**Then** it exposes:

```go
type AuditEntryStore interface {
    // Append inserts the entry using the supplied tx. Caller owns the
    // transaction lifecycle (per architecture.md canonical request flow).
    // The entry's ID and OccurredAt are server-assigned on insert and
    // populated on the passed entry pointer on success.
    Append(ctx context.Context, tx pgx.Tx, entry *domain.AuditEntry) error
}

type auditEntryStorePg struct{}

func NewAuditEntryStore() AuditEntryStore { return &auditEntryStorePg{} }
```

Implementation uses explicit `pgx` SQL with an enumerated column list and `RETURNING id, occurred_at`:

```go
const insertSQL = `
    INSERT INTO domain.audit_entry (
        id, actor_id, action, entity_type, entity_id, project_id,
        before_state, after_state, metadata
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    RETURNING id, occurred_at
`
```

If `entry.ID == uuid.Nil`, the implementation pre-fills `entry.ID = uuid.New()` before insert (DDL has no `DEFAULT` on `id` — it's required input). The `RETURNING` clause writes back `OccurredAt` so callers see the server-assigned timestamp. No `Deps.AuditEntries` wiring in `app/deps.go` this story (no handler consumes it yet — wiring lands in 2.8). The constructor is exposed so a future `Deps` field can pick it up without ceremony.

**No conversion of `AuditAction` to string at the boundary:** callers pass the string form already (`string(enums.AuditActionProjectPlacedOnHold)`). The helper does not depend on the `enums` package, keeping `internal/data/` free of a `domain/enums` import cycle if one is later introduced.

### AC5 — Transactional integrity (per-stack integration test)

**Given** any of the three stacks
**When** a handler-style test:
1. Opens a transaction
2. Inserts a `domain.project` row (use the Story 2.1 mapping or a raw SQL insert against `domain.project`)
3. Calls `append_audit_entry(..., entity_type="Project", entity_id=<project_id>, project_id=<project_id>, ...)` inside that same transaction
4. **Rolls back** (do not commit)

**Then** after a fresh connection / reopened transaction:
- `SELECT count(*) FROM domain.project WHERE id = <project_id>` returns `0`
- `SELECT count(*) FROM domain.audit_entry WHERE entity_id = <project_id>` returns `0`

No orphaned audit row. This is the load-bearing test for FR39 + FR57 (audit-in-same-transaction). One per-stack integration test is required:
- **.NET:** `FieldMark.Tests.Integration/AuditAppenderRollbackTests.cs` using `[Collection(PostgresCollection.Name)]` and the existing `PostgresContainerFixture` — open `FieldMarkDbContext` against the Testcontainer connection string, call `IAuditAppender.Append`, do not commit, assert via raw `NpgsqlCommand` on a fresh connection.
- **Django:** `fieldmark_py/audit/tests/test_append_audit_entry.py` marked `@pytest.mark.integration`. Use the `domain_db` fixture from [conftest.py](../../fieldmark_py/conftest.py) (rolls back on teardown) — but because `append_audit_entry` writes via the Django ORM through the `default` connection rather than the fixture cursor, the test must wrap the whole sequence in `with transaction.atomic():` and raise to force rollback, or use `transaction.atomic()` + manual `transaction.set_rollback(True)`. Document the chosen approach in the test docstring; the existing `test_rollback_leaves_no_trace` in [audit/tests/test_db_rollback.py](../../fieldmark_py/audit/tests/test_db_rollback.py) demonstrates the two-phase pattern.
- **Go:** `fieldmark-go/internal/data/postgres/auditentrystore_test.go` with `//go:build integration`. Use `openPool(t)` from the existing [integration_test.go](../../fieldmark-go/internal/data/postgres/integration_test.go); `Begin`, insert project row + call `AuditEntryStore.Append`, `Rollback`, then assert via fresh pool connection.

Additionally each stack ships a positive **commit** test in the same file: open tx, insert project + audit entry, commit, assert both rows are present after a fresh-connection query, then clean up with `DELETE FROM domain.audit_entry WHERE entity_id = $1; DELETE FROM domain.project WHERE id = $1` (audit_entry first — `idx_audit_project` references the project, but the DDL leaves `project_id` as nullable with no `ON DELETE CASCADE`, so order matters). **Append-only at app level:** the cleanup `DELETE` lives in the test, not in any application path, mirroring the DDL comment at [010_domain_tables.sql:187–189](../../docker/postgres/init/010_domain_tables.sql).

### AC6 — Per-stack action-set conformance test

**Given** a checked-in JSON fixture at `docs/reference/audit-actions.json`
**When** I inspect it
**Then** it is a single JSON object of the form:

```json
{
  "actions": [
    "ProjectCreated",
    "ProjectPlacedOnHold",
    "ProjectResumed",
    "..."
  ]
}
```

Listing **every** canonical PascalCase action string exactly once, in the same order as the doc table. The fixture is the conformance gate; the doc table is human-readable. **Both must agree** — the doc's "Change Procedure" section makes the dual-update explicit.

**Given** each stack
**When** I run that stack's conformance test
**Then** the test:
1. Reads `docs/reference/audit-actions.json` from a path resolved relative to the repo root (walk up from the test working directory until finding `docs/reference/audit-actions.json`, mirroring the `LocateInitDir` pattern in [PostgresContainerFixture.cs](../../FieldMark/FieldMark.Tests.Integration/PostgresContainerFixture.cs)).
2. Extracts the stack's native action set: `Enum.GetNames<AuditAction>().Select(name => name)` for .NET (and verifies the string form via `AsString()` matches the symbol name verbatim — a regression guard for accidental override), `AuditAction.values` for Django (Python: `[c.value for c in AuditAction]`), `AllAuditActions` for Go (cast to `[]string`).
3. Asserts the two sets are equal (no extras, no missing). On failure, prints the symmetric diff — `expected ∖ actual` and `actual ∖ expected` — so the developer sees both directions at once.

This is a **unit test, not an integration test** — no DB required. Placement:
- **.NET:** `FieldMark/FieldMark.Tests/AuditActionConformanceTests.cs` (the existing unit-test project). If `FieldMark.Tests/` does not exist or has been pruned, place at `FieldMark/FieldMark.Domain.Tests/` — confirm the actual layout before writing.
- **Django:** `fieldmark_py/audit/tests/test_action_conformance.py` (no `@pytest.mark.integration` — pure unit test).
- **Go:** `fieldmark-go/internal/domain/enums/audit_action_conformance_test.go` (no build tag — pure unit test).

### AC7 — `make parity` clean, no phantom indexes, no new routes

**Given** all three stacks now map `domain.audit_entry`
**When** I run `make parity` from the repo root
**Then** `pg_indexes` for `domain.*` shows **zero diff** against the baseline at [_parity-snapshots/](_parity-snapshots/). The two existing audit indexes (`idx_audit_entity`, `idx_audit_project`) are DDL-owned; no stack mapping introduces a duplicate. If a phantom index appears (most likely from EF Core inferring an index on `(entity_type, entity_id)`), resolve by either dropping the EF declaration or naming it to match (`HasDatabaseName("idx_audit_entity")`) per AC2.

**Given** this story
**When** I inspect the diff
**Then** no new route is introduced on any stack. The `append_audit_entry` helper has no HTTP surface.

### AC8 — Cross-stack architecture principle guard rail

**Given** the Cross-Stack Architecture Principle (root [CLAUDE.md](../../CLAUDE.md))
**When** I inspect this story's diff
**Then**:
- The canonical contract lives at `docs/reference/audit-actions.md` + the derived `docs/reference/audit-actions.json` fixture — **and nowhere else**.
- No file in `fieldmark_shared/` lists audit-action strings. No generated stub. No symlinked manifest.
- Each stack's enum/constants file is the native implementation: `FieldMark.Domain/ValueObjects/AuditAction.cs`, `fieldmark_py/audit/actions.py`, `fieldmark-go/internal/domain/enums/audit_action.go`. A developer working inside one stack reads only their stack's file + the top-of-file comment pointing to `docs/reference/audit-actions.md`.

### AC9 — Build, type, lint, and test gates green on every stack

- **.NET:** `cd FieldMark && dotnet csharpier check . && dotnet build && dotnet test && dotnet test FieldMark.Tests.Integration/FieldMark.Tests.Integration.csproj` — clean.
- **Django:** `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest && uv run pytest -m integration` — clean. Verify `uv run python manage.py makemigrations audit` outputs `No changes detected` (or only `django_auth`-scoped output) — `Meta.managed = False` must hold.
- **Go:** `cd fieldmark-go && make check && go test ./... && go test -tags=integration ./internal/data/postgres/...` — clean.
- From repo root: `make parity` exits 0 (AC7).

## Tasks / Subtasks

- [ ] **Task 1: Resolve the canonical action set** (AC: #1)
  - [ ] 1.1 Read [architecture.md:603](../planning-artifacts/architecture.md) — enumerate the 13 strings literally listed.
  - [ ] 1.2 Read the [epic 2 file Story 2.8 note](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) — confirms `ProjectCreated` is added by ADR amendment, total 15.
  - [ ] 1.3 Read PRD FR9–FR15 and Epic 4/5/6 story emissions to find the missing 14th string from the architecture line (likely `ViolationResolved` per the corrective-action approval flow; verify against [architecture.md:1329](../planning-artifacts/architecture.md)'s flow diagram and `_bmad-output/planning-artifacts/prd/`). If genuinely indeterminate, raise it to Tim before populating the doc — do not guess silently.
  - [ ] 1.4 Write the resolution into the doc's **Change Procedure** section so the next reviewer can audit the call.

- [ ] **Task 2: Populate the contract** (AC: #1, #6, #8)
  - [ ] 2.1 Rewrite [docs/reference/audit-actions.md](../../docs/reference/audit-actions.md) replacing every `TODO` with the populated sections per AC1. Keep the existing top-matter intact (the `Status:` line should be updated to `Status: live — populated by Story 2.2`).
  - [ ] 2.2 Create `docs/reference/audit-actions.json` per AC6's schema. **Single line per action, sorted to match the doc table order, not alphabetical.** Mismatch between this file and the doc table is a reviewer red flag.
  - [ ] 2.3 Walk the cross-stack principle (root CLAUDE.md): confirm zero non-doc/non-fixture artifact lists audit actions. Specifically: no entry in `fieldmark_shared/`.

- [ ] **Task 3: .NET — entity, configuration, enum, helper, tests** (AC: #2, #5, #6, #7, #9)
  - [ ] 3.1 Add `FieldMark/FieldMark.Domain/ValueObjects/AuditAction.cs` per AC2. Add a `static class AuditActionExtensions { public static string AsString(this AuditAction a) => a.ToString(); }`. The conformance test in 3.7 proves this is correct for every value.
  - [ ] 3.2 Add `FieldMark/FieldMark.Domain/Entities/AuditEntry.cs` per AC2. Use `System.Text.Json.JsonDocument` for the three JSONB columns.
  - [ ] 3.3 Add `FieldMark/FieldMark.Data/Configuration/AuditEntryConfiguration.cs` per AC2. Run `make parity` immediately after wiring; resolve any phantom-index diff before proceeding.
  - [ ] 3.4 Add `DbSet<AuditEntry> AuditEntries` to `FieldMark.Data/Context/FieldMarkDbContext.cs`. The existing `ApplyConfigurationsFromAssembly` call from Story 2.1 picks the new config up automatically — verify by inspecting `OnModelCreating`.
  - [ ] 3.5 Add `FieldMark/FieldMark.Data/Auditing/IAuditAppender.cs` and `AuditAppender.cs` per AC2. The implementation does NOT call `SaveChangesAsync` and does NOT open a transaction.
  - [ ] 3.6 Register `services.AddScoped<IAuditAppender, AuditAppender>()` in `FieldMark.Web/Program.cs`. No handler consumes it yet; Story 2.8 will pick it up. Verify the scoping by reading the existing `FieldMarkDbContext` registration in `Program.cs` — both must be `Scoped` (or `AddDbContextPool`-style — confirm compatibility) so the same `DbContext` instance is shared inside a request.
  - [ ] 3.7 Add `FieldMark/FieldMark.Tests/AuditActionConformanceTests.cs` per AC6 (locate the unit-test project — if it doesn't exist as `FieldMark.Tests/`, place under `FieldMark.Domain.Tests/`; confirm the layout before writing). The test walks the repo root up from `AppContext.BaseDirectory`, loads `docs/reference/audit-actions.json`, and asserts set equality.
  - [ ] 3.8 Add `FieldMark/FieldMark.Tests.Integration/AuditAppenderRollbackTests.cs` per AC5 (rollback test + commit-and-cleanup test).
  - [ ] 3.9 Run `dotnet csharpier format .`, `dotnet build`, `dotnet test`, `dotnet test FieldMark.Tests.Integration` — all green.

- [ ] **Task 4: Django — model, constants, helper, tests** (AC: #3, #5, #6, #7, #9)
  - [ ] 4.1 Populate `fieldmark_py/audit/models.py` per AC3 (currently a single `# Create your models here.` comment). Use `models.JSONField` (works against psycopg's native JSONB adapter on Django ≥ 3.1; verify `pyproject.toml`).
  - [ ] 4.2 Add `fieldmark_py/audit/actions.py` per AC3. Use `models.TextChoices` (Django idiom).
  - [ ] 4.3 Add `fieldmark_py/audit/append.py` per AC3 — keyword-only signature, returns the created `AuditEntry`.
  - [ ] 4.4 Run `uv run python manage.py makemigrations audit` — assert `No changes detected` (or only `django_auth`-scoped output). If Django emits a migration for `domain.audit_entry`, the `Meta` flags are wrong.
  - [ ] 4.5 Add `fieldmark_py/audit/tests/test_action_conformance.py` per AC6 (pure unit test — no `integration` marker). Walk up from `__file__` to find `docs/reference/audit-actions.json`.
  - [ ] 4.6 Add `fieldmark_py/audit/tests/test_append_audit_entry.py` per AC5 (rollback + commit-and-cleanup). Marked `@pytest.mark.integration`. Reuse the project fixture pattern from `fieldmark_py/projects/tests/test_mapping.py` if Story 2.1 has landed by the time this story is implemented; otherwise insert the parent project row inline via raw SQL.
  - [ ] 4.7 Run `uv run ruff check .`, `uv run mypy .`, `uv run pytest`, `uv run pytest -m integration` — all green.

- [ ] **Task 5: Go — entity, enum, store, tests** (AC: #4, #5, #6, #7, #9)
  - [ ] 5.1 Add `fieldmark-go/internal/domain/entities/audit_entry.go` per AC4. Use `json.RawMessage` (`encoding/json`) for the three JSONB columns — pgx scans `jsonb` into `[]byte` / `json.RawMessage` directly.
  - [ ] 5.2 Add `fieldmark-go/internal/domain/enums/audit_action.go` per AC4. Declare `AllAuditActions` slice in the same file as the iteration target for the conformance test.
  - [ ] 5.3 Add `fieldmark-go/internal/data/postgres/auditentrystore.go` per AC4. Use the SQL constant block; `pgx.Tx.QueryRow(ctx, insertSQL, ...).Scan(&entry.ID, &entry.OccurredAt)` to capture the server-side `OccurredAt`.
  - [ ] 5.4 Add `fieldmark-go/internal/domain/enums/audit_action_conformance_test.go` per AC6 — pure unit test, no build tag. Walk up from `runtime.Caller` to find `docs/reference/audit-actions.json`.
  - [ ] 5.5 Add `fieldmark-go/internal/data/postgres/auditentrystore_test.go` per AC5 with `//go:build integration`. Reuse `openPool(t)`.
  - [ ] 5.6 No `app/deps.go` wiring this story — Story 2.8 introduces the first handler consumer and will plumb `Deps.AuditEntries`. (If contentious, wire `Deps.AuditEntries = NewAuditEntryStore()` now; YAGNI default applies.)
  - [ ] 5.7 Run `make check && go test ./... && go test -tags=integration ./internal/data/postgres/...` — all green.

- [ ] **Task 6: Parity and cross-stack verification** (AC: #1, #7, #8)
  - [ ] 6.1 Run `make parity` from repo root. `pg_indexes` diff for `domain.*` MUST remain at zero. Resolve any phantom audit-entry index before merging (see AC2 resolution path).
  - [ ] 6.2 Verify `docs/reference/audit-actions.json` and the `docs/reference/audit-actions.md` table list the same actions in the same order. A trivial reviewer check; trivially regressable.
  - [ ] 6.3 Grep `fieldmark_shared/` for any audit-action string — must return zero hits (AC8 guard).
  - [ ] 6.4 Run all three conformance tests and confirm the assertion message format (symmetric diff) is readable on intentional drift — temporarily add a 16th member to one stack's enum, watch the test fail with both `expected ∖ actual` and `actual ∖ expected` populated, then revert.

- [ ] **Task 7: Story sign-off** (AC: all)
  - [ ] 7.1 Populate the Sign-off block (date, review-round count, reviewer verdict, deferred-work link if any).
  - [ ] 7.2 `dev-story` workflow flips `sprint-status.yaml` `2-2-...` to `review` — no manual edit required here.

## Dev Notes

### Critical context (read before writing code)

- **Doc-first, code-second.** Populating `docs/reference/audit-actions.md` + `audit-actions.json` (Task 2) is a precondition for the conformance tests in every stack. Do Task 1 + 2 before Task 3/4/5 so the tests have a fixture to read.
- **Reconcile the action count (13 vs 14+1) in Task 1.3 before populating the doc.** This is the single substantive judgment call in the story. If the missing 14th string is genuinely indeterminate from the existing artifacts, raise it to Tim rather than guessing — silent invention of an audit action would be a worse failure than asking.
- **Append-only at app level.** The DDL has no `UPDATE` or `DELETE` permission on `audit_entry` in production (see [010_domain_tables.sql:187–189](../../docker/postgres/init/010_domain_tables.sql)). Application code MUST NOT issue UPDATE or DELETE against `domain.audit_entry`. The cleanup `DELETE` in the per-stack commit test lives in test code only, runs against a transient row with a synthetic UUID, and is the only allowed exception.
- **Helper does not own the transaction.** Every stack's helper is the *non-trivial inner step* of the canonical request flow ([architecture.md §Process Patterns:733–746](../planning-artifacts/architecture.md)). The handler opens `BeginTransactionAsync` / `transaction.atomic()` / `WithTx`; the helper writes the row using the handler's connection/context. Helpers that open their own transactions break FR39 (audit-on-every-mutation): a rollback elsewhere in the handler would leave an orphan audit row.
- **Persist the string, not the enum.** Per FR40 ("no inventing variants") and the conformance test contract: the column stores the PascalCase string verbatim. The enum is a developer ergonomic, not a storage type. .NET avoids `HasConversion<AuditAction>()`; Django stores `action.value`; Go callers pass `string(enumConst)`. Conformance ensures the symbols never drift from the persisted strings.
- **JSONB nullability is meaningful.** `before_state` is `NULL` for `Created` actions (no prior state), `after_state` is `NULL` for `Deleted`/`Voided` actions if the entity is being terminated. The mapping must permit `NULL`; do not coalesce to `'null'::jsonb` or `'{}'::jsonb`. Empty payload and absence-of-payload are semantically different.
- **`docs/reference/audit-actions.json` is the conformance fixture, not the contract.** The Markdown table is human-canonical. The JSON exists because parsing a Markdown table in three different stacks' test runners is fragile; the JSON is a derived, machine-readable mirror that the Change Procedure mandates be regenerated whenever the doc changes.

### Edge cases (per [docs/reference/component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md))

Walked the nine categories. **None apply** — no user-facing component, no JS, no fonts, no overlays, no toasters. This is a data-layer + library-helper story. If a future reviewer suggests Category 1 (unknown enum values) applies to `AuditAction`: on read, an unrecognized DB value would fail enum deserialization in C#, raise a `ValueError` in Python (`AuditAction(value)` constructor), or surface as a plain string in Go (the type alias accepts any string). The conformance tests in AC6 prevent this *prospectively* on the write side. On the read side, the story doesn't read audit entries through the enum at all — only Story 2.13 (Project audit log tab) does, and its dev notes should add a "treat unknown value as opaque pass-through and log a warning" branch.

### Security defaults (per [docs/reference/security-defaults.md](../../docs/reference/security-defaults.md))

Walked the seven categories. **None apply** — no forms, cookies, redirects, user input, or filesystem writes. The `actor_id` parameter is taken from the authenticated `Actor`/`User` in the handler (not from request input), which is the security control point — and lives in the handler stories, not here.

### Cross-stack contract three-deliverable check

The Cross-Stack Architecture Principle's three-deliverable rule **fully applies** to this story:

| Deliverable | Where |
|---|---|
| 1. Documentation contract | `docs/reference/audit-actions.md` (this story populates) + `docs/reference/audit-actions.json` fixture |
| 2. Native implementation per stack | `.NET FieldMark.Domain/ValueObjects/AuditAction.cs` + `FieldMark.Data/Auditing/AuditAppender.cs`; Django `audit/actions.py` + `audit/append.py`; Go `internal/domain/enums/audit_action.go` + `internal/data/postgres/auditentrystore.go` |
| 3. Per-stack conformance test | .NET `AuditActionConformanceTests.cs`; Django `audit/tests/test_action_conformance.py`; Go `audit_action_conformance_test.go` |

Plus the transactional-integrity test (AC5) in each stack, which is a behavioral conformance gate rather than a contract gate.

### Files this story modifies vs creates

| File | New / Modified | Purpose |
|---|---|---|
| `docs/reference/audit-actions.md` | MODIFY (populate skeleton) | canonical contract |
| `docs/reference/audit-actions.json` | NEW | conformance fixture |
| `FieldMark/FieldMark.Domain/ValueObjects/AuditAction.cs` | NEW | enum + `AsString()` |
| `FieldMark/FieldMark.Domain/Entities/AuditEntry.cs` | NEW | entity property bag |
| `FieldMark/FieldMark.Data/Configuration/AuditEntryConfiguration.cs` | NEW | EF mapping |
| `FieldMark/FieldMark.Data/Auditing/IAuditAppender.cs` | NEW | helper interface |
| `FieldMark/FieldMark.Data/Auditing/AuditAppender.cs` | NEW | helper implementation |
| `FieldMark/FieldMark.Data/Context/FieldMarkDbContext.cs` | MODIFY | add `DbSet<AuditEntry>` |
| `FieldMark/FieldMark.Web/Program.cs` | MODIFY | register `IAuditAppender` |
| `FieldMark/FieldMark.Tests/AuditActionConformanceTests.cs` *(or `FieldMark.Domain.Tests/`)* | NEW | AC6 conformance |
| `FieldMark/FieldMark.Tests.Integration/AuditAppenderRollbackTests.cs` | NEW | AC5 rollback + commit |
| `fieldmark_py/audit/models.py` | MODIFY | replace placeholder with `AuditEntry` |
| `fieldmark_py/audit/actions.py` | NEW | `AuditAction` TextChoices |
| `fieldmark_py/audit/append.py` | NEW | `append_audit_entry()` helper |
| `fieldmark_py/audit/tests/test_action_conformance.py` | NEW | AC6 conformance |
| `fieldmark_py/audit/tests/test_append_audit_entry.py` | NEW | AC5 rollback + commit |
| `fieldmark-go/internal/domain/entities/audit_entry.go` | NEW | struct |
| `fieldmark-go/internal/domain/enums/audit_action.go` | NEW | typed constants + `AllAuditActions` |
| `fieldmark-go/internal/domain/enums/audit_action_conformance_test.go` | NEW | AC6 conformance |
| `fieldmark-go/internal/data/postgres/auditentrystore.go` | NEW | interface + pgx impl |
| `fieldmark-go/internal/data/postgres/auditentrystore_test.go` | NEW | AC5 rollback + commit |

Anything outside this list — handlers, routes, view models, templates, `app/deps.go` wiring, `AuditRow` component, audit-log tab UI, project audit-log reads, `docs/how-to/*` — is out of scope. Resist the urge.

### Files to read fully before editing

- [docker/postgres/init/010_domain_tables.sql:190–211](../../docker/postgres/init/010_domain_tables.sql) — `domain.audit_entry` DDL + the two indexes + the append-only comment. Binding.
- [docs/reference/audit-actions.md](../../docs/reference/audit-actions.md) — current skeleton; you are populating it.
- [_bmad-output/planning-artifacts/architecture.md](../planning-artifacts/architecture.md) §**Audit Trail (FR39–FR43)** (line 57), §**Audit action strings (canonical)** (line 603), §**The Canonical Request Flow** (lines 733–878) — the helper is step 5 of the canonical flow; the existing C#/Python/Go stubs are the implementation template.
- [_bmad-output/implementation-artifacts/2-1-map-domain-project-and-supporting-tables-into-each-stacks-data-layer.md](2-1-map-domain-project-and-supporting-tables-into-each-stacks-data-layer.md) — same shape of mapping story; reuse the conventions decided there (snake-case naming, no `HasIndex` calls for DDL-owned constraints, `Meta.managed = False`, flat `internal/data/postgres/`).
- [FieldMark/FieldMark.Tests.Integration/PostgresContainerFixture.cs](../../FieldMark/FieldMark.Tests.Integration/PostgresContainerFixture.cs) and [DomainRollbackSmokeTests.cs](../../FieldMark/FieldMark.Tests.Integration/DomainRollbackSmokeTests.cs) — .NET integration harness pattern.
- [fieldmark_py/conftest.py](../../fieldmark_py/conftest.py) and [fieldmark_py/audit/tests/test_db_rollback.py](../../fieldmark_py/audit/tests/test_db_rollback.py) — Django integration harness pattern (especially the two-phase rollback verification).
- [fieldmark-go/internal/data/postgres/integration_test.go](../../fieldmark-go/internal/data/postgres/integration_test.go) — Go integration harness pattern; `openPool(t)` helper.
- [fieldmark-go/internal/app/actor.go](../../fieldmark-go/internal/app/actor.go) — the `Actor` shape, for handler stories that come later; not needed inside the helper but useful background.
- [_bmad-output/implementation-artifacts/epic-1-retro-2026-05-25.md](epic-1-retro-2026-05-25.md) §**Story AC Amendments Landed During Retro** — confirms 2.2 carries the three-deliverable rule.

### Project Structure Notes

- The Django `audit` app exists with placeholder `models.py` (`# Create your models here.`) and a populated `tests/` subpackage (Epic 1 retro A3). This story replaces the placeholder; the `tests/__init__.py` already exists.
- The Go `internal/domain/enums/` package exists with `role.go` per Story 1.12; add `audit_action.go` alongside.
- The .NET `FieldMark.Domain/Entities/` directory may not exist yet (Story 2.1 creates it). If 2.1 has not yet landed when this story starts, create it.
- The .NET test project layout: confirm whether unit tests live at `FieldMark/FieldMark.Tests/` or `FieldMark/FieldMark.Domain.Tests/` before placing `AuditActionConformanceTests.cs`. The conformance test only needs a reference to `FieldMark.Domain`; it is fastest to place in whichever project already has that reference. If neither exists, create a new test project — but check Story 2.1's resolution of the same question first.

### References

- AC source: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.2
- DDL: [docker/postgres/init/010_domain_tables.sql:190–211](../../docker/postgres/init/010_domain_tables.sql)
- Contract: [docs/reference/audit-actions.md](../../docs/reference/audit-actions.md) (skeleton to be populated)
- Canonical action list source: [architecture.md:603](../planning-artifacts/architecture.md) (13 strings) + epic 2 file Story 2.8 note (`ProjectCreated` added by ADR amendment)
- Canonical request flow (helper is step 5): [architecture.md:733–878](../planning-artifacts/architecture.md)
- Cross-Stack Architecture Principle: root [CLAUDE.md](../../CLAUDE.md) §Cross-Stack Architecture Principle (ratified Epic 1 retro 2026-05-25)
- Previous story (Project mapping — same shape pattern): [2-1-map-domain-project-and-supporting-tables-into-each-stacks-data-layer.md](2-1-map-domain-project-and-supporting-tables-into-each-stacks-data-layer.md)
- Stack rules: [FieldMark/CLAUDE.md](../../FieldMark/CLAUDE.md), [fieldmark_py/CLAUDE.md](../../fieldmark_py/CLAUDE.md), [fieldmark-go/CLAUDE.md](../../fieldmark-go/CLAUDE.md)
- Epic 1 retro three-deliverable rule: [epic-1-retro-2026-05-25.md](epic-1-retro-2026-05-25.md) §Story AC Amendments Landed During Retro (Story 2.2 row)

## Dev Agent Record

### Agent Model Used

_to be filled at implementation time_

### Debug Log References

### Completion Notes List

### File List

## Sign-off

| Field | Value |
|---|---|
| Final review date | _pending_ |
| Total review rounds | _pending_ |
| Final reviewer verdict | _pending_ |
| Deferred-work entries | _pending_ |
| Dev-notes divergences from epic AC | _pending_ — record here if the Task 1 action-count reconciliation lands on a 14th string different from `ViolationResolved`, and the rationale. |
