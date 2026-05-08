// Package app contains use-case orchestration for FieldMark.
//
// It holds application services, persistence port interfaces, and DTOs that
// cross the app/web boundary. It must NOT import Fiber, render HTML, or
// contain SQL queries. It depends on domain and on port interfaces only —
// never on concrete persistence adapters.
package app
