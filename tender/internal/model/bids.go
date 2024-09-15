package model

import (
	"time"

	"github.com/google/uuid"
)

type BidsStatus string

type Bids struct {
	ID              uuid.UUID `json:"id"`
	TenderID        uuid.UUID `json:"tenderId"`
	OrganizationID  uuid.UUID `json:"organizationId"`
	Title           string    `json:"name"`
	Description     string    `json:"description"`
	Status          string    `json:"status"`
	Version         int       `son:"version"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	CreatorUsername string    `json:"creatorUsername"`
}
