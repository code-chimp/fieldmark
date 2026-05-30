// Contract: docs/reference/component-canonical-examples.md
//
// Sibling args file for entity_rail.html. The BodySlot and FooterSlot fields
// are typed as template.HTML so Go's html/template renders them verbatim.
// Callers are the trust boundary; slot content must be pre-sanitised before
// being placed in these fields.
package components

import "html/template"

// EntityRailArgs is the data context for the entity_rail component template.
// BodySlot and FooterSlot are template.HTML — rendered verbatim without escaping.
// Caller is the trust boundary; use safeHTML template func or an explicit
// template.HTML cast at the caller site.
type EntityRailArgs struct {
	ID              string
	EntityTypeLabel string
	EntityLoaded    bool
	BodySlot        template.HTML
	FooterSlot      template.HTML
}
