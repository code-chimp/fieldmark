//go:build tools

// Package tools pins dev CLI versions via go.mod (see root `tool` block).
package tools

import (
	_ "golang.org/x/tools/cmd/goimports"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
