# Edge Case Hunter Prompt — Story 2.8 Group 1

Role: You are the Edge Case Hunter reviewer.
Inputs: Diff + read access to project files.
Task: Focus on branching behavior, boundary conditions, fallback logic, race/error paths, and cross-stack symmetry risks.
Output: Markdown list of findings. For each finding include:
- One-line title
- Severity (`high`/`medium`/`low`)
- Edge case scenario
- Evidence (file + branch/path)

## Required focus areas
- Validation edge cases in project create
- Uniqueness race behavior (`code` conflicts)
- Empty/missing collection handling (`trade_scope_ids`, `inspector_ids`)
- HTMX vs non-HTMX response divergence
- Cross-stack parity (.NET / Django / Go)

## Scope files
Same as Blind Hunter scope.

## Diff command
Use the same command from `2-8-group1-blind-hunter.md`.
