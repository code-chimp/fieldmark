# Future Plan: Java Stack Candidate

**Date:** 2026-05-30  
**Status:** Future option, not committed scope  
**Related project:** FieldMark server-authoritative HTMX reference implementation

## Purpose

After the initial FieldMark implementation proves the architecture across .NET, Django, and Go/Fiber, a Java implementation may be worth adding to reach a wider backend-centric developer audience.

The Java version would not change the FieldMark thesis. It would test whether the same server-authoritative HTMX architecture remains credible in the mainstream enterprise Java ecosystem:

- Server-rendered pages and HTMX partials, not SPA routing.
- AG Grid Enterprise SSRM as a scoped JavaScript island.
- Business rules, validation, authorization, state transitions, audit writes, and compliance scoring owned by the server.
- Shared PostgreSQL `domain` schema, owned by infrastructure SQL, not by framework migrations.
- Route, HTMX target, AG Grid contract, audit string, and domain-method parity with the existing stacks.

## Recommended Stack

Use a conventional Spring enterprise stack:

| Concern | Recommendation | Rationale |
|---|---|---|
| Application framework | Spring Boot | Most recognizable mainstream Java backend choice. |
| Web framework | Spring MVC | Fits request-response, server-rendered HTMX architecture better than WebFlux. |
| Templates | Thymeleaf | Common Spring server-rendered HTML pairing. |
| Interactivity | HTMX | Preserve FieldMark's hypermedia-centered interaction model. |
| Auth/authz | Spring Security | Enterprise Java-standard security vocabulary. |
| Core ORM | Spring Data JPA + Hibernate | Familiar to senior Java developers; avoids looking like a non-idiomatic Java port. |
| Grid/read queries | jOOQ | Strong fit for AG Grid SSRM, reporting, audit lists, and projection-heavy SQL. |
| Schema migration | Flyway or Liquibase for `java_auth` only | Preserve infrastructure ownership of `domain`; allow stack-local auth schema management. |
| Build tool | Maven | Most broadly recognizable Spring enterprise build choice; keeps the Java stack conventional for reviewers. |
| Tests | JUnit, Testcontainers, MockMvc | Common Spring testing shape with real PostgreSQL. |

## Persistence Strategy

Use a hybrid persistence strategy rather than forcing one tool everywhere.

### Transactional Domain Workflows

Use Spring Data JPA/Hibernate for aggregate loading and persistence in state-changing workflows:

```text
Spring MVC Controller
  -> thin transactional application service
  -> Spring Data repository loads aggregate
  -> entity/domain method performs transition
  -> audit entry + compliance recomputation in same transaction
  -> Thymeleaf partial/full page render
```

This is the Java stack's closest equivalent to the project's existing domain-centric pattern. It gives Java reviewers the expected JPA/Hibernate vocabulary while keeping FieldMark's key rule intact: business decisions live in the domain path, not in the browser.

Guardrails:

- Map JPA entities to `domain.*` tables.
- Set Hibernate schema behavior to validate-only, not create/update.
- Do not generate or apply JPA/Flyway/Liquibase migrations against `domain`.
- Keep repositories persistence-focused; do not let repository methods encode workflow rules.
- Keep application services thin: authorize, start transaction, load aggregate, invoke domain method, append audit entry, recompute compliance score, persist, render.

### AG Grid SSRM And Read-Heavy Screens

Use jOOQ for AG Grid SSRM and projection-heavy reads:

- Project lists.
- Violation grids.
- Audit/history views.
- Dashboard/reporting projections.
- Server-side sorting, filtering, pagination, and row loading.

This avoids contorting JPA criteria/query abstractions around dynamic grid contracts while giving Java reviewers a credible SQL-first tool for data-dense enterprise screens.

## Proposed Module Shape

A possible Spring Boot structure:

```text
fieldmark-java/
├── pom.xml
├── src/main/java/dev/fieldmark/
│   ├── FieldMarkJavaApplication.java
│   ├── domain/              # Entities, enums, domain exceptions, transition methods
│   ├── persistence/
│   │   ├── jpa/             # Spring Data repositories and JPA mappings
│   │   └── jooq/            # Grid/read-model query adapters
│   ├── app/                 # Thin transactional orchestration services
│   ├── web/
│   │   ├── controllers/     # Spring MVC controllers
│   │   ├── viewmodels/      # Manual projections for Thymeleaf
│   │   └── grid/            # AG Grid SSRM endpoints
│   ├── security/            # Spring Security config and role mapping
│   └── config/              # DataSource, jOOQ, transaction, template config
├── src/main/resources/
│   ├── templates/           # Full pages and HTMX partials
│   ├── static/              # Symlinked shared assets where practical
│   └── db/migration/        # `java_auth` only if Flyway/Liquibase is used
└── src/test/java/           # JUnit, Testcontainers, MockMvc
```

Exact package boundaries should be revisited when the stack is implemented. The important part is that Spring idioms do not become an excuse to weaken FieldMark's shared contracts.

## Scope Boundaries

The Java stack should be considered only after the initial three-stack implementation is complete enough to demonstrate the thesis.

Recommended entry criteria:

- .NET, Django, and Go/Fiber stacks have completed the anchor workflows.
- AG Grid Enterprise SSRM is working in the existing stacks.
- Parity tooling can compare the original three stacks reliably.
- The landing page and README explain the project as a reference implementation, not a product roadmap.

Recommended non-goals:

- Do not replace any existing stack.
- Do not add a Java stack before the first architecture proof is complete.
- Do not introduce a client-side JavaScript framework.
- Do not let Java-specific persistence tooling own the shared `domain` schema.
- Do not use WebFlux unless the project explicitly pivots to demonstrating reactive Java, which is currently unrelated to the thesis.

## Research Notes

This direction is intended to avoid a credibility gap with senior Java developers. A purely explicit-SQL Spring implementation would be coherent with FieldMark's infrastructure-owned schema, but it could read as intentionally avoiding the mainstream enterprise Java ORM path. Spring Data JPA/Hibernate is the more recognizable default for transactional domain persistence.

jOOQ is recommended as a complementary tool, not a replacement for JPA in the whole stack. It is better suited to the dynamic, projection-heavy, server-side query work required by AG Grid SSRM.

Official references used:

- [Spring Boot documentation](https://docs.spring.io/spring-boot/documentation.html)
- [Spring MVC with Spring Boot](https://docs.spring.io/spring-boot/how-to/spring-mvc.html)
- [Spring Framework Thymeleaf integration](https://docs.spring.io/spring-framework/reference/web/webmvc-view/mvc-thymeleaf.html)
- [Spring Data JPA reference](https://docs.spring.io/spring-data/jpa/reference/index.html)
- [Spring Security authorization reference](https://docs.spring.io/spring-security/reference/features/authorization/index.html)
- [Spring JDBC `JdbcClient` reference](https://docs.spring.io/spring-framework/reference/data-access/jdbc/core.html)
- [Spring Boot SQL database documentation, including jOOQ and migration tooling](https://docs.spring.io/spring-boot/reference/data/sql.html)
- [jOOQ manual](https://www.jooq.org/doc/latest/manual/)

## Open Questions For Future Design

- Whether `java_auth` should use Spring Security JDBC tables, JPA-managed auth entities, or Flyway-managed custom tables aligned with the other stacks' conceptual-role model.
- How much JPA entity behavior is acceptable before Java reviewers expect separate domain objects. FieldMark should prefer entity/domain behavior, but the exact Java idiom deserves deliberate design.
- Whether jOOQ code generation should target only read-model tables/views or the full `domain` schema. If code generation is used, configure it through Maven rather than introducing a second build system.
- How parity tooling should name the fourth stack: `java`, `spring`, or `spring-boot`.
