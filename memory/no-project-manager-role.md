---
name: no-project-manager-role
description: There is no distinct ProjectManager role in the seeded role set; PM-persona duties map to ADMIN
metadata:
  type: project
---

FieldMark's seeded conceptual role set is exactly `{ADMIN, EXECUTIVE, INSPECTOR, SITE_SUPERVISOR, COMPLIANCE_OFFICER}`. There is **no distinct `ProjectManager` role**, even though epic/UX personas (e.g. Aisha, "Project Manager") and epic story titles (e.g. Story 2.8 "Project create form (PM/Admin)") use that language.

PM-persona duties are fulfilled by the **ADMIN** role. Evidence: Story 2.8 granted `project.create` → ADMIN; Story 2.11 registered `project.place_on_hold`/`resume`/`close` → ADMIN only; the 2.11 trichotomy test enumerates exactly the five roles above (ADMIN permitted, the other four not).

**Why:** Story epics say "As a Project Manager" but the implemented authz has no such role — re-deriving this each story risks inventing a role or mis-granting permissions.

**How to apply:** When a story says "Project Manager", map it to ADMIN. Do not register a new role or broaden grants beyond ADMIN unless a story explicitly ratifies it via the Change Procedure. See [[story-2-12-place-on-hold]] (Decision 3).
