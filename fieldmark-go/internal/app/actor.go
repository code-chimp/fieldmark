// Package app is the THIN coordinator — wiring only (Deps struct, env
// config, Actor type). It must not import fiber/v3, and no business
// rules live here. See fieldmark-go/CLAUDE.md.
package app

import "github.com/google/uuid"

// Actor is the resolved request principal. Constructed by the auth
// middleware (internal/web/auth) and read by handlers via the web
// layer's ActorFromCtx helper. Lives in app/ so future packages
// (e.g., domain audit-entry helpers) can take an Actor parameter
// without depending on web/.
type Actor struct {
	ID          uuid.UUID
	Username    string
	DisplayName string
	// Role is "" for anonymous; otherwise one of the five canonical role
	// values enumerated in domain.AllRoles. Typed as string to avoid a
	// refactor cascade; domain.Role* constants are the authoritative names.
	Role string
}

// Anonymous returns the sentinel actor representing an unauthenticated
// request. ID is uuid.Nil. Role is the empty string.
func Anonymous() *Actor {
	return &Actor{Username: "anonymous"}
}

// IsAnonymous is true when the Actor has no resolved identity.
func (a *Actor) IsAnonymous() bool {
	return a == nil || a.ID == uuid.Nil
}
