# Hard Rules (All Stacks)

These cannot be relaxed without an ADR amendment:

- **Backend authority.** Domain rules, transitions, validation, authorization — server only.
- **Infrastructure-owned domain schema.** `domain` created by SQL init scripts; frameworks only touch their auth schemas.
- **No fat service layers.** Handlers call entity methods directly.
- **No repository or Unit-of-Work abstractions.** Use DbContext/ORM/explicit SQL directly.
- **No CQRS, MediatR, or in-process buses.**
- **No client-side state stores.** No Redux, Zustand, etc.
- **No AutoMapper.** Project to view models manually.
- **No SQLite in tests.** Real PostgreSQL only.
- **AuditEntry writes are non-optional** and in same transaction.
- **Stack symmetry** on routes, HTMX IDs, AG Grid contracts, audit strings, method names.
- **Casing is canonical at wire/DB** (`snake_case`); code uses language idiom.

Stack-specific rules live in each project's CLAUDE.md.
