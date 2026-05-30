# Blind Hunter Prompt — Story 2.8 Group 1

Role: You are the Blind Hunter reviewer. You receive only the diff excerpt below. No project context, no spec context.
Task: Find concrete defects, regressions, or risky assumptions visible from the diff alone.
Output: Markdown list of findings. For each finding include:
- One-line title
- Severity (`high`/`medium`/`low`)
- Evidence (file + behavior)
- Why it matters

## Review Scope (Group 1)
Core create-flow behavior files only:
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

## Diff Command
Use this exact command from repo root to capture the scope:
```bash
git diff main...HEAD -- \
  FieldMark/FieldMark.Domain/Entities/Project.cs \
  FieldMark/FieldMark.Domain/Entities/ProjectInspector.cs \
  FieldMark/FieldMark.Domain/Entities/ProjectTradeScope.cs \
  FieldMark/FieldMark.Web/Pages/Projects/Create.cshtml.cs \
  FieldMark/FieldMark.Web/Pages/Projects/Index.cshtml.cs \
  fieldmark_py/projects/forms.py \
  fieldmark_py/projects/models.py \
  fieldmark_py/projects/views.py \
  fieldmark-go/internal/domain/entities/project_create.go \
  fieldmark-go/internal/data/postgres/projectstore.go \
  fieldmark-go/internal/web/handlers/projects_create_handler.go
```
