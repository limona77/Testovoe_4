package model

import (
	"time"

	"github.com/google/uuid"
)

type Tender struct {
	ID              uuid.UUID `json:"id"`
	OrganizationID  uuid.UUID `json:"organizationId"`
	Title           string    `json:"name"`
	Description     string    `json:"description"`
	ServiceType     string    `json:"serviceType"`
	Status          string    `json:"status"`
	Version         int       `son:"version"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	CreatorUsername string    `json:"creatorUsername"`
}
