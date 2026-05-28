package entities

import (
	"encoding/json"

	"github.com/google/uuid"
)

type TradeType struct {
	ID          uuid.UUID
	Code        string
	Name        string
	Description *string
	Active      bool
}

type ViolationCategory struct {
	ID              uuid.UUID
	Code            string
	Name            string
	TradeTypeID     *uuid.UUID
	DefaultSeverity string
	Description     *string
	Active          bool
}

type ComplianceRule struct {
	ID          uuid.UUID
	Code        string
	Name        string
	Description string
	RuleKind    string
	Parameters  json.RawMessage
	Active      bool
}
