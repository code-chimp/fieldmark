# FieldMark Documentation Index

This folder contains the canonical, published documentation for developers, contributors, and operators. It is organized by the **[Diátaxis](https://diataxis.fr/)** framework — four quadrants matched to four kinds of user need.

| Quadrant | When to read | What's here |
|----------|--------------|-------------|
| **[Tutorials](tutorials/)** — learning-oriented | You are new and want to get something running | [getting-started.md](tutorials/getting-started.md) |
| **[How-to guides](how-to/)** — problem-oriented | You know the system and need to accomplish a specific task | [basecoat-upgrade-checklist.md](how-to/basecoat-upgrade-checklist.md) |
| **[Reference](reference/)** — information-oriented | You need precise, lookup-style facts | [hard-rules.md](reference/hard-rules.md), [domain-model.md](reference/domain-model.md), [persistence-schema.md](reference/persistence-schema.md) |
| **[Explanation](explanation/)** — understanding-oriented | You want background, rationale, and the bigger picture | [overview.md](explanation/overview.md), [architecture.md](explanation/architecture.md), [request-flow.md](explanation/request-flow.md) |

## Related References

- Root [CLAUDE.md](../CLAUDE.md) — AI/agent guidance with links to these docs (progressive disclosure).
- Stack-specific CLAUDE.md files in `FieldMark/`, `fieldmark_py/`, and `fieldmark-go/`.
- `_bmad-output/planning-artifacts/` — Historical planning artifacts (not maintained post-kickoff).

## Notes on the four quadrants

The four kinds of doc are kept in separate folders so each can stay true to its purpose. Tutorials minimize explanation; how-to guides assume basic knowledge; reference avoids instruction; explanation avoids step-by-step. When adding a new document, pick the quadrant first — if a draft mixes two, split it.
