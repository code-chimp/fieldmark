# Shared domain schema ownership — planning primer

**Audience:** BMAD agents, architects, and contributors who need a **single entry point** for how FieldMark’s **`domain`** schema is owned and evolved across Django, .NET, and Fiber.

**Normative detail:** This primer **does not replace** the ADRs or the canonical DDL narrative. Use it as orientation and checklist; defer to:

- [`architecture-decisions.md`](architecture-decisions.md) — **ADR-012** (framework-local auth), **ADR-013** (schemas created by infrastructure), **ADR-014** (domain schema infrastructure-owned)
- [`domain-model.md`](domain-model.md) — entity catalog, naming (`snake_case`), and **§9** / DDL for actual table names

**Together with** [`authentication-authorization-primer.md`](authentication-authorization-primer.md), ADRs **012–014** describe the full **data-layer contract**: who owns schemas, where identity lives, and how shared business tables evolve.

---

## Principles (summary)

1. **The domain is not owned by a framework** — no stack is the system of record for `domain.*` DDL.
2. **The database expresses architecture** — shared tables are explicit, reviewable, and evolved deliberately.
3. **Infrastructure precedes app migrations** — schemas exist before framework migrations; framework migrations stay inside **authorized** schemas only.
4. **Stacks are replaceable** — projections change; **`domain`** stays the contract.

---

## Database and schemas

| Item | Convention |
|------|----------------|
| Database | Single Postgres DB **`fieldmark`** (local/demo simplicity) |
| **`domain`** | Authoritative business tables — **infrastructure-owned** |
| **`django_auth`**, **`dotnet_auth`**, **`fiber_auth`** | Framework-local identity/admin — each framework’s migrations only |
| **`infra`** | Reserved cross-stack metadata / infra concerns (per init scripts; optional early on) |

Frameworks **must not** create or rename **schemas** via application code.

**Where schemas come from:** Postgres **`docker-entrypoint-initdb.d`** scripts in **`docker/postgres/init/`** at repo root. That runs on **first** container init; changing schemas later may require volume reset (see ADR-013 consequences in [`architecture-decisions.md`](architecture-decisions.md)).

---

## Why not code-first for `domain.*`

ORM migrations are appropriate for **single-stack** apps; for a **shared multi-backend** domain they implicitly crown one framework as owner, bias naming to ORM defaults, and block parallel evolution. FieldMark trades some ORM migration convenience for **neutrality and clarity** (expanded rationale: ADR-014).

---

## Acceptable ORM / data access (by stack)

| Stack | Rule |
|-------|------|
| **.NET** | EF Core maps to `domain.*` via explicit fluent config; migrations **only** touch `dotnet_auth` |
| **Django** | Models may map `domain.*` rows with **`Meta.managed = False`** (and explicit `db_table` / schema); migrations **only** touch `django_auth` |
| **Fiber** | Explicit SQL (or narrow stores) against `domain.*`; **no** DDL ownership of `domain` |

---

## Domain ↔ authentication boundary

- No **foreign keys** from `domain.*` to any `*_auth` table.
- Store user linkage as **opaque UUIDs** (e.g. `created_by_user_id`, `actor_id`) — see ADR-012 and `domain-model.md` §3.11–§3.13.

---

## Change workflow for `domain` tables

When shared domain DDL must change:

1. Update the **ERD / domain documentation** ([`domain-model.md`](domain-model.md)).
2. Author **infrastructure SQL** under `docker/postgres/init/` (or numbered migrations **outside** framework tooling), reviewed like application code.
3. Review **cross-stack impact** (.NET mappings, Django models, Fiber SQL).
4. Update **framework mappings** to match; **never** let EF/Django “discover” drift into `domain.*`.
5. Apply via your documented DB bootstrap (e.g. `docker compose` / volume lifecycle per ADR-013 notes).

**Prohibited:** silent drift from ORM auto-migrations; framework-specific “extras” on shared tables without coordinated infra SQL; breaking changes without updating all consumers.

---

## Agent and contributor checklist

1. Do **not** generate **`domain.*`** tables via Django / EF / Go migration tooling.
2. Do **not** alter **`domain.*`** from application migration pipelines.
3. Treat **`domain`** schema as a **published contract** alongside `domain-model.md`.
4. When infra DDL changes, update **all three** mapping layers as needed.
5. Escalate unclear ownership (domain vs auth vs infra) before inventing schema.

---

## Relation to other research

- **ADRs 012–014** together define auth + init + domain ownership — see **Cross-reference — data layer and identity** in [`architecture-decisions.md`](architecture-decisions.md).
- Stack guardrails repeat mapping discipline in [`django-reference.md`](django-reference.md), [`dotnet-reference.md`](dotnet-reference.md), [`fiber-reference.md`](fiber-reference.md).
- **Data & persistence** constraints are summarized again in [`architecture-decisions.md`](architecture-decisions.md) Part 2.

---

**Status:** Consolidated primer — supersedes standalone “shared domain ownership strategy” drafts when paired with ADRs and `domain-model.md`.
