package repository

import (
	"context"
	"zadanie-6105/internal/model"
	"zadanie-6105/pkg/postgres"

	"github.com/google/uuid"
)

type ITender interface {
	GetTenders(ctx context.Context, limit int, offset int, serviceTypesArr []string) ([]model.Tender, error)
	CreateTender(ctx context.Context, tender model.Tender) (model.Tender, error)
	GetTender(ctx context.Context, user string, limit int, offset int) ([]model.Tender, error)
	UpdateTender(ctx context.Context, tender model.Tender) (model.Tender, error)
	GetStatus(ctx context.Context, tenderId uuid.UUID) (string, error)
	UpdateStatus(ctx context.Context, tender model.Tender) (model.Tender, error)
	IsUserResponsibleForOrganization(ctx context.Context, tender model.Tender) (bool, error)
}
type IBids interface {
	CreateBids(ctx context.Context, bids *model.Bids) (model.Bids, error)
	GetBids(ctx context.Context, user string, limit, offset int) ([]model.Bids, error)
	GetBidsByTenderId(ctx context.Context, user string, tenderId uuid.UUID, limit, offset int) ([]model.Bids, error)
	UpdateBids(ctx context.Context, bids *model.Bids) (model.Bids, error)
	IsUserAuthorizedToCreateBid(ctx context.Context, bids *model.Bids) (bool, error)
	IsTenderValid(ctx context.Context, tenderID uuid.UUID) (bool, error)
	GetBidStatus(ctx context.Context, bidId uuid.UUID, user string) (string, error)
	CheckUserExists(ctx context.Context, username string) (bool, error)
	UpdateBidsStatus(ctx context.Context, bids model.Bids) (model.Bids, error)
	UpdateBidsDecision(ctx context.Context, bidId uuid.UUID, decision, username string) (model.Bids, error)
}
type Repositories struct {
	ITender
	IBids
}

func NewRepositories(db *postgres.DB) *Repositories {
	return &Repositories{NewTenderRepository(db), NewBidsRepository(db)}
}
