# Authentication & authorization — planning primer

**Audience:** BMAD agents, architects, and contributors defining identity and access across Django, .NET, and Fiber without collapsing FieldMark into a single auth implementation.

**Normative ADR:** [`architecture-decisions.md`](architecture-decisions.md) — **ADR-012: Authentication Is Framework-Local; Authorization Is Domain-Driven.** This primer orients and summarizes; it **does not replace** ADR-012.

**Stack guardrails:** [`django-reference.md`](django-reference.md), [`dotnet-reference.md`](dotnet-reference.md), [`fiber-reference.md`](fiber-reference.md) — authentication & authorization sections.

**Domain linkage:** Opaque user references on `domain.*` rows — ADR-012 and [`domain-model.md`](domain-model.md) (identity boundaries).

---

## Thesis

**One authoritative domain, multiple replaceable application projections.** Identity plumbing stays **framework-local**; **business roles and permission semantics** stay aligned to the **domain** vocabulary below.

---

## Principles

| Principle | Meaning |
|-----------|---------|
| **Authentication is framework-local** | Each stack owns its identity store and login mechanics. No shared users table and no cross-framework SSO for demo scope. |
| **Authorization is domain-driven** | Roles reflect **business intent**; stacks map them to native auth APIs (Django groups/permissions, ASP.NET policies, Fiber middleware). UI does not invent new authorization meanings. |
| **Schemas enforce boundaries** | `django_auth`, `dotnet_auth`, `fiber_auth` hold framework identity data; **`domain`** holds business data (ADR-013). |
| **Domain stays portable** | `domain.*` does **not** foreign-key any `*_auth` table. Store user linkage as **opaque UUIDs** (e.g. `created_by_user_id`, `actor_id`); meaning is local to the active stack (ADR-012). |

---

## Postgres (auth-related)

Single database **`fieldmark`**. Auth data for each stack lives only in its schema — see ADR-012/013. Schemas are **infrastructure** (created under **`docker/postgres/init/`**, not by framework migrations); **who may act** is enforced in **server** code, not by FK bridges between **`domain`** and **`*_auth`**.

---

## Trade-offs accepted (ADR-012)

Deliberate costs of framework-local identity:

- **No shared login UX** across Django / .NET / Fiber demos (acceptable for the teaching artifact).
- **Opaque user IDs on `domain.*`** are not enforced by foreign keys to auth tables — **referential discipline is application-level** per stack.
- **Role-to-permission mapping** is implemented **separately** in each backend; parity of *meaning* is required, not identical tables.

---

## Shared role vocabulary (conceptual)

These names are **product language**, not ORM class names. Each framework implements mapping to its own groups/roles/claims.

| Role | Typical capabilities (illustrative) |
|------|--------------------------------------|
| **Administrator** | System configuration, compliance rules, user management |
| **Compliance Officer** | Review/resolve violations, audit visibility |
| **Inspector** | Perform inspections, record findings |
| **Site Supervisor** | Project compliance visibility, respond to violations |
| **Executive Viewer** | Read-only dashboards and reports |

Refine per [`prd.md`](prd.md) / product specs if role names or duties change; keep **parity of meaning** across stacks.

---

## Authorization model (MVP posture)

- **Role-based** (not fine-grained ABAC) with **coarse, explicit** permissions.
- **Server-side only** — no client-side authority; HTMX and templates reflect **backend** allow/deny.
- Examples of intent: only Administrators change compliance rules; only Compliance Officers approve resolution paths Inspectors cannot reach admin surfaces — encode in handlers/middleware per stack.

---

## Framework responsibilities (summary)

| Stack | Authentication |
|-------|------------------|
| **Django** | Built-in auth + Admin; tables in **`django_auth`**; map roles to permissions in Django terms. |
| **.NET** | ASP.NET Core Identity when adopted; tables in **`dotnet_auth`**; policies/roles align to vocabulary above. |
| **Fiber** | Deferred until needed; **`fiber_auth`** when implemented; middleware enforces role checks. |

No stack reads or writes another stack’s auth schema.

---

## Explicit non-goals (demo scope)

- Single sign-on across stacks  
- Shared users table  
- Cross-framework login session  
- External IdP integration  

May be revisited later; out of scope for the teaching artifact unless product scope changes.

---

## Lifecycle

- **Architecture** (framework-local auth + domain-driven authorization) is fixed **now**.
- **Implementations** land **incrementally** — Django auth/admin often earliest; .NET and Fiber add identity when features require it, without redesigning **`domain`**.

---

## Agent checklist

1. Do **not** introduce a **shared** identity schema or FK from **`domain.*`** to **`*_auth.*`**.
2. Map the **five conceptual roles** consistently when enforcing rules in views/handlers (opaque UUID on domain rows comes from **that** stack’s user store).
3. Keep authorization decisions **server-side**; reject client-only gates for business actions.
4. Use stack **`CLAUDE.md`** + stack reference for deferrals (e.g. Fiber auth not scaffolded until needed).

---

## References

- [architecture-decisions.md](architecture-decisions.md) — ADR-012 (full rationale and consequences); see **Cross-reference — data layer and identity** for ADRs 012–014 together.
- [domain-schema-ownership-primer.md](domain-schema-ownership-primer.md) — domain vs auth schema separation and DDL workflow.
- [domain-model.md](domain-model.md) — opaque user references on domain entities.

---

**Status:** Planning primer — pairs with ADR-012–014 and `domain-schema-ownership-primer.md` as the consolidated story for auth + init + domain ownership.
