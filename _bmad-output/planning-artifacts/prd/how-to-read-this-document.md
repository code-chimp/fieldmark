# How to Read This Document

This PRD is the **capability contract** and **architectural-binding document** for FieldMark. It serves three audiences:

1. **The author** — as a synthesis of long-percolating intuitions about server-driven architecture, made examinable.
2. **AI agents** (BMAD method, Claude Code, others) — as the authoritative source for what FieldMark is, what it must do, and what it must not do, when implementing or reviewing code.
3. **Future contributors and external readers** — as a complete, standalone description of the artifact's intent.

**Reading order.** Sections progress from vision (Executive Summary) through strategy (Success Criteria, Product Scope) to commitments (Architectural Constraints, Functional and Non-Functional Requirements). Earlier sections frame; later sections bind.

**Cross-references.** Where a topic has a canonical home within this document, other sections reference it rather than restate it. For example, performance targets are locked in §Success Criteria → Measurable Outcomes; §Web App Specific Requirements and §Non-Functional Requirements reference that section rather than duplicate the numbers. Architectural rules are in §Architectural Constraints (PRD-Binding); the Functional Requirements section assumes them.

**Source of truth.** This PRD, together with the downstream BMAD artifacts (architecture document and epics/stories, produced in subsequent phases), is the authoritative description of FieldMark. Earlier draft documents in `_bmad-output/planning-artifacts/research/` were seeding inputs for synthesis; they are not maintained going forward and should not be relied on by agents.

**Out-of-scope items** are documented explicitly in §Non-Functional Requirements → Non-Goals. If something is not in this PRD and not in the Non-Goals list, treat its inclusion as an open question requiring an answer before implementation.
