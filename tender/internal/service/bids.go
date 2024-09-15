package service

import (
	"context"
	"errors"
	"fmt"
	custom_errors "zadanie-6105/internal/custom-errors"
	"zadanie-6105/internal/model"
	"zadanie-6105/internal/repository"

	"github.com/google/uuid"
)

type BidsService struct {
	bidsRepository repository.IBids
}

func NewBidsService(bidsRepository repository.IBids) *BidsService {
	return &BidsService{
		bidsRepository: bidsRepository,
	}
}

func (bS *BidsService) CreateBids(ctx context.Context, bids *model.Bids) (model.Bids, error) {
	path := "service.bidss.CreateBids"
	isAuthorized, err := bS.bidsRepository.IsUserAuthorizedToCreateBid(ctx, bids)
	if err != nil {
		return model.Bids{}, fmt.Errorf(path+".IsUserAuthorizedToCreateBid, error: {%w}", err)
	}
	if !isAuthorized {
		return model.Bids{}, fmt.Errorf(path+".IsUserAuthorizedToCreateBid, error: {%w}", custom_errors.ErrAccessDenied)
	}

	isValidTender, err := bS.bidsRepository.IsTenderValid(ctx, bids.TenderID)
	if err != nil {
		return model.Bids{}, fmt.Errorf(path+".IsTenderValid, error: {%w}", err)
	}
	if !isValidTender {
		return model.Bids{}, fmt.Errorf(path+".IsUserAuthorizedToCreateBid, error: {%w}", errors.New("tender not found"))
	}
	bids.Status = "CREATED"
	return bS.bidsRepository.CreateBids(ctx, bids)
}

func (bs *BidsService) GetBids(ctx context.Context, user string, limit, offset int) ([]model.Bids, error) {
	return bs.bidsRepository.GetBids(ctx, user, limit, offset)
}

func (bs *BidsService) GetBidsByTenderId(
	ctx context.Context,
	user string,
	tenderId uuid.UUID,
	limit, offset int,
) ([]model.Bids, error) {
	exists, err := bs.bidsRepository.CheckUserExists(ctx, user)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []model.Bids{}, custom_errors.ErrUserNotFound
	}
	return bs.bidsRepository.GetBidsByTenderId(ctx, user, tenderId, limit, offset)
}

func (bs *BidsService) UpdateBids(ctx context.Context, bids *model.Bids) (model.Bids, error) {
	return bs.bidsRepository.UpdateBids(ctx, bids)
}

func (bs *BidsService) GetBidStatus(ctx context.Context, bidId uuid.UUID, user string) (string, error) {
	return bs.bidsRepository.GetBidStatus(ctx, bidId, user)
}

func (bs *BidsService) UpdateBidsStatus(ctx context.Context, bids model.Bids) (model.Bids, error) {
	return bs.bidsRepository.UpdateBidsStatus(ctx, bids)
}

func (bs *BidsService) UpdateBidsDecision(
	ctx context.Context,
	bidId uuid.UUID,
	decision, username string,
) (model.Bids, error) {
	if decision != "Approved" && decision != "Rejected" {
		return model.Bids{}, custom_errors.ErrUnprocessableEntity
	}
	return bs.bidsRepository.UpdateBidsDecision(ctx, bidId, decision, username)
}
