# FieldMark Per‑Language QA & Static Analysis Checklist

## Purpose

This document defines a **per‑language quality assurance (QA) and static analysis checklist** for the FieldMark monorepo.

The intent is:
- to use **idiomatic, best‑of‑breed tooling per ecosystem**
- to avoid forcing cross‑language symmetry
- to ensure high signal‑to‑noise feedback for humans and agents
- to align tooling with backend‑authority and low‑ceremony principles

This document assumes primary development in **JetBrains IDEs** (Rider, PyCharm, GoLand), with **VS Code** and **NeoVim** as secondary editors.

---

## Guiding Principles

1. Formatting is deterministic and automatic
2. Linting focuses on correctness and maintainability, not style bikeshedding
3. Static analysis should surface real bugs and design smells
4. IDEs should reinforce rules automatically
5. CI enforcement can be added later; local feedback comes first

---

## .NET / C# QA Tooling

### Formatting

✅ **CSharpier**

- Opinionated, deterministic formatter
- Comparable to Prettier / Black
- Should run automatically on save via IDE integration

Recommended:
- Enable CSharpier plugin in Rider
- Avoid mixing with other formatters

---

### Static Analysis (Roslyn Analyzers)

Roslyn analyzers are the primary linting mechanism for C#.

They operate at **compile time** and integrate deeply with IDEs.

#### Recommended Analyzer Sets

✅ **Microsoft.CodeAnalysis.NetAnalyzers** (built‑in)
- Enabled by default in modern .NET SDKs
- Covers:
  - correctness
  - performance
  - reliability
  - security

Optional additions:
- **StyleCop.Analyzers** (if you want explicit style rules)
- **SonarAnalyzer.CSharp** (if deeper bug detection is desired)

---

### Enabling and Configuring Roslyn Analyzers

#### 1. Enable analyzers in the project

In your `.csproj`:

```xml
<PropertyGroup>
  <EnableNETAnalyzers>true</EnableNETAnalyzers>
  <AnalysisLevel>latest</AnalysisLevel>
</PropertyGroup>
```

This enables Microsoft’s built‑in analyzers at the latest rule set.

#### 2. Control severity via `.editorconfig`

```ini
[*.cs]
# Treat selected rules as warnings or errors
dotnet_diagnostic.CA1062.severity = warning
dotnet_diagnostic.CA2000.severity = warning
dotnet_diagnostic.CA1822.severity = suggestion
```

Start with **warnings**, not errors.

---

### IDE Notes (Rider)

- Rider automatically surfaces Roslyn diagnostics inline
- CSharpier integrates cleanly and should be the only formatter
- Inspections can be tuned per solution

Avoid enabling overlapping inspections that fight Roslyn rules.

---

## Python / Django QA Tooling

### Formatting

✅ **Black**

- Deterministic, opinionated formatter
- Minimal configuration
- Equivalent to CSharpier / Prettier

Enable format‑on‑save in PyCharm or VS Code.

---

### Linting & Code Smells

✅ **Ruff** (primary tool)

Ruff replaces:
- flake8
- pyflakes
- pycodestyle
- isort
- parts of pylint

It detects:
- unused imports
- unreachable code
- shadowed variables
- suspicious logic
- excessive complexity

Use Ruff as the **single source of linting truth**.

---

### Type Checking (Optional but Recommended)

✅ **mypy** with **django‑stubs**

- Enforces static typing where present
- Helps validate domain and service interfaces

Guidance:
- Do not enable strict mode initially
- Treat typing as assistive, not punitive

---

### Minimal Python QA Stack

✅ Black
✅ Ruff
✅ mypy (+ django‑stubs)
✅ pytest

This stack is widely accepted in mature Django codebases.

---

### IDE Notes (PyCharm / VS Code / NeoVim)

- PyCharm integrates Black, Ruff, and mypy natively
- VS Code works well with Ruff + Black extensions
- NeoVim users typically rely on:
  - null‑ls / conform
  - ruff‑lsp or pylsp

---

## Go / Fiber QA Tooling

Go emphasizes correctness via the compiler and standard tooling.

---

### Formatting

✅ **gofmt** (mandatory)
✅ **goimports** (recommended)

- gofmt is non‑negotiable
- goimports additionally manages imports

Most editors run these automatically on save.

---

### Static Analysis

✅ **go vet** (baseline)

- Part of the Go toolchain
- Finds suspicious constructs the compiler allows

✅ **staticcheck** (strongly recommended)

Detects:
- subtle bugs
- nil misuses
- ineffective assignments
- deprecated APIs
- concurrency issues

staticcheck is the closest Go equivalent to ESLint‑level correctness checks.

---

### Aggregated Linting (Optional)

✅ **golangci‑lint**

- Runs multiple analyzers together
- Useful as a single entry point
- Keep enabled linters minimal to avoid noise

---

### Minimal Go QA Stack

✅ gofmt
✅ goimports
✅ go vet
✅ staticcheck
✅ go test

This is idiomatic, conservative, and widely respected.

---

### IDE Notes (GoLand / VS Code / NeoVim)

- GoLand integrates gofmt, go vet, and staticcheck seamlessly
- VS Code Go extension supports all core tools
- NeoVim users typically integrate:
  - gopls
  - gofmt/goimports on save

---

## Cross‑Language Comparison (Conceptual)

| Concern | .NET | Python | Go |
|------|------|--------|----|
| Formatter | CSharpier | Black | gofmt |
| Linter | Roslyn | Ruff | staticcheck |
| Type checks | Compiler + analyzers | mypy | Compiler |
| Tests | xUnit | pytest | go test |

Each stack uses **native tooling**, but enforces the same architectural intent.

---

## Recommended Adoption Order

1. Enable formatters in IDEs
2. Enable linters at warning level
3. Run tools locally before committing
4. Add CI enforcement later if desired

Do not block early development with overly strict rules.

---

## Rules for Contributors and Agents

- Use the dominant tooling of the language
- Do not introduce alternate linters without justification
- Prefer warnings over errors initially
- Treat static analysis findings as design feedback, not noise

---

## Status

Drafted – FieldMark per‑language QA & static analysis checklist
