package postgres

import "errors"

// ErrProjectNotFound is returned by ProjectStore.Load when no row matches.
// Handler stories (2.11) translate this to HTTP 404; do not translate it in
// the data layer.
var ErrProjectNotFound = errors.New("project not found")
