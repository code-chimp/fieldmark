# Component Canonical Examples

## Status

Live — populated by Story 2.4, 2026-05-28.

## Why

This document indexes FieldMark component fixtures that define the cross-stack rendered-markup contract. The contract follows the root `CLAUDE.md` Cross-Stack Architecture Principle: cross-stack invariants live as documentation plus native implementation and native conformance tests, not as shared runtime code.

## Component Index

| Component | Canonical example | README | .NET wrapper | Django wrapper | Go wrapper | .NET test | Django test | Go test |
|---|---|---|---|---|---|---|---|---|
| StatusBadge | `fieldmark_shared/components/status_badge/canonical.html` | `fieldmark_shared/components/status_badge/README.md` | `FieldMark/FieldMark.Web/Pages/Shared/Components/_StatusBadge.cshtml` | `fieldmark_py/templates/components/_status_badge.html` | `fieldmark-go/internal/web/templates/components/status_badge.html` | `FieldMark/FieldMark.Tests.Web/Components/StatusBadgeSnapshotTests.cs` | `fieldmark_py/fieldmark/tests/test_status_badge_snapshot.py` | `fieldmark-go/internal/web/templates/components/status_badge_test.go` |
| InlineAlert | `fieldmark_shared/components/inline_alert/canonical.html` | `fieldmark_shared/components/inline_alert/README.md` | `FieldMark/FieldMark.Web/Pages/Shared/Components/_InlineAlert.cshtml` | `fieldmark_py/templates/components/_inline_alert.html` | `fieldmark-go/internal/web/templates/components/inline_alert.html` | `FieldMark/FieldMark.Tests.Web/Components/InlineAlertSnapshotTests.cs` | `fieldmark_py/fieldmark/tests/test_inline_alert_snapshot.py` | `fieldmark-go/internal/web/templates/components/inline_alert_test.go` |
| AuditRow | `fieldmark_shared/components/audit_row/canonical.html` | `fieldmark_shared/components/audit_row/README.md` | `FieldMark/FieldMark.Web/Pages/Shared/Components/_AuditRow.cshtml` | `fieldmark_py/templates/components/_audit_row.html` | `fieldmark-go/internal/web/templates/components/audit_row.html` | `FieldMark/FieldMark.Tests.Web/Components/AuditRowSnapshotTests.cs` | `fieldmark_py/fieldmark/tests/test_audit_row_snapshot.py` | `fieldmark-go/internal/web/templates/components/audit_row_test.go` |
| DashboardTile | `fieldmark_shared/components/dashboard_tile/canonical.html` | `fieldmark_shared/components/dashboard_tile/README.md` | `FieldMark/FieldMark.Web/Pages/Shared/Components/_DashboardTile.cshtml` | `fieldmark_py/templates/components/_dashboard_tile.html` | `fieldmark-go/internal/web/templates/components/dashboard_tile.html` | `FieldMark/FieldMark.Tests.Web/Components/DashboardTileSnapshotTests.cs` | `fieldmark_py/fieldmark/tests/test_dashboard_tile_snapshot.py` | `fieldmark-go/internal/web/templates/components/dashboard_tile_test.go` |
| ComplianceTile | `fieldmark_shared/components/compliance_tile/canonical.html` | `fieldmark_shared/components/compliance_tile/README.md` | `FieldMark/FieldMark.Web/Pages/Shared/Components/_ComplianceTile.cshtml` | `fieldmark_py/templates/components/_compliance_tile.html` | `fieldmark-go/internal/web/templates/components/compliance_tile.html` | `FieldMark/FieldMark.Tests.Web/Components/ComplianceTileSnapshotTests.cs` | `fieldmark_py/fieldmark/tests/test_compliance_tile_snapshot.py` | `fieldmark-go/internal/web/templates/components/compliance_tile_test.go` |
| EntityRail | `fieldmark_shared/components/entity_rail/canonical.html` | `fieldmark_shared/components/entity_rail/README.md` | `FieldMark/FieldMark.Web/Pages/Shared/Components/_EntityRail.cshtml` | `fieldmark_py/templates/components/_entity_rail.html` | `fieldmark-go/internal/web/templates/components/entity_rail.html` | `FieldMark/FieldMark.Tests.Web/Components/EntityRailSnapshotTests.cs` | `fieldmark_py/fieldmark/tests/test_entity_rail_snapshot.py` | `fieldmark-go/internal/web/templates/components/entity_rail_test.go` |
| TabStrip | `fieldmark_shared/components/tab_strip/canonical.html` | `fieldmark_shared/components/tab_strip/README.md` | `FieldMark/FieldMark.Web/Pages/Shared/Components/_TabStrip.cshtml` | `fieldmark_py/templates/components/_tab_strip.html` | `fieldmark-go/internal/web/templates/components/tab_strip.html` | `FieldMark/FieldMark.Tests.Web/Components/TabStripSnapshotTests.cs` | `fieldmark_py/fieldmark/tests/test_tab_strip_snapshot.py` | `fieldmark-go/internal/web/templates/components/tab_strip_test.go` |

## Snapshot-Test Pipeline

The canonical pipeline is documented in `fieldmark_shared/CLAUDE.md` §"Snapshot-test pipeline". Story 2.4 adds no component-specific normalization deviation; each stack reads the matching `canonical.html`, extracts the named variant block, normalizes comments/whitespace/entity spelling through its existing helper, and asserts rendered wrapper output equals the fixture.

## Change Procedure

Adding a component is a four-step change:

1. Author `fieldmark_shared/components/<name>/canonical.html` and `fieldmark_shared/components/<name>/README.md`.
2. Implement the native wrapper in each stack.
3. Implement the per-stack snapshot tests against the canonical fixture variants.
4. Append one row to the Component Index above.
