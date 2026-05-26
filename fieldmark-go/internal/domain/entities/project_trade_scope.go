package entities

import "github.com/google/uuid"

type ProjectTradeScope struct {
	ProjectID   uuid.UUID
	TradeTypeID uuid.UUID
}
