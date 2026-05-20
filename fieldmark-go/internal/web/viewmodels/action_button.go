package viewmodels

// ActionButtonVM is the data context for the action_button component template
// (internal/web/templates/components/action_button.html). Build it in handlers
// using auth.Can to populate Permission and an entity-method-derived bool for
// StateAllows. Templates never call Can directly.
type ActionButtonVM struct {
	ID             string
	Permission     bool
	StateAllows    bool
	Label          string
	HxPost         string
	HxTarget       string
	DisabledReason string
}
