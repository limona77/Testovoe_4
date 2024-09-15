package model

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationType string

const (
	OrganizationTypeIE  OrganizationType = "IE"
	OrganizationTypeLLC OrganizationType = "LLC"
	OrganizationTypeJSC OrganizationType = "JSC"
)

type Organization struct {
	ID          uuid.UUID                 `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Type        OrganizationType          `json:"type"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
	Responsible []OrganizationResponsible `json:"responsible"`
}

type OrganizationResponsible struct {
	ID             uuid.UUID    `json:"id"`
	OrganizationID uuid.UUID    `json:"organization_id"`
	UserID         uuid.UUID    `json:"user_id"`
	Organization   Organization `json:"organization"`
	Employee       Employee     `json:"user"`
}
