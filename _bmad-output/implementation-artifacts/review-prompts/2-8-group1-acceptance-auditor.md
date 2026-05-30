# Acceptance Auditor Prompt — Story 2.8 Group 1

You are an Acceptance Auditor.
Review this diff against the spec and context docs.
Check for: violations of acceptance criteria, deviations from spec intent, missing implementation of specified behavior, contradictions between spec constraints and actual code.
Output findings as a Markdown list. Each finding must include:
- one-line title
- which AC/constraint it violates
- evidence from the diff
- severity (`high`/`medium`/`low`)

## Spec file
- `_bmad-output/implementation-artifacts/2-8-project-create-form-pm-admin.md`

## Context docs
- `_bmad-output/project-context.md`
- `docs/reference/hard-rules.md`
- `docs/reference/security-defaults.md`
- `docs/reference/component-edge-case-checklist.md`
- `_bmad-output/implementation-artifacts/deferred-work.md`
- `docs/reference/project-create-form-contract.md`

## Review Scope (Group 1 files)
- FieldMark/FieldMark.Domain/Entities/Project.cs
- FieldMark/FieldMark.Domain/Entities/ProjectInspector.cs
- FieldMark/FieldMark.Domain/Entities/ProjectTradeScope.cs
- FieldMark/FieldMark.Web/Pages/Projects/Create.cshtml.cs
- FieldMark/FieldMark.Web/Pages/Projects/Index.cshtml.cs
- fieldmark_py/projects/forms.py
- fieldmark_py/projects/models.py
- fieldmark_py/projects/views.py
- fieldmark-go/internal/domain/entities/project_create.go
- fieldmark-go/internal/data/postgres/projectstore.go
- fieldmark-go/internal/web/handlers/projects_create_handler.go

## Diff command
Use the same command from `2-8-group1-blind-hunter.md`.
