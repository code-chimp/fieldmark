package entities

import "github.com/google/uuid"

type JobSite struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Label     string
	Address   *string
}
