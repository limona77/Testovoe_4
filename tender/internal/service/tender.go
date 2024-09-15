package service

import (
	"context"
	"fmt"
	"zadanie-6105/helper"
	custom_errors "zadanie-6105/internal/custom-errors"
	"zadanie-6105/internal/model"
	"zadanie-6105/internal/repository"

	"github.com/google/uuid"
)

type TenderService struct {
	tenderRepository repository.ITender
}

func NewTenderService(tenderRepository repository.ITender) *TenderService {
	return &TenderService{
		tenderRepository: tenderRepository,
	}
}

func (tS *TenderService) GetTenders(
	ctx context.Context,
	limit int,
	offset int,
	serviceTypesArr []string,
) ([]model.Tender, error) {
	return tS.tenderRepository.GetTenders(ctx, limit, offset, serviceTypesArr)
}

func (tS *TenderService) CreateTender(ctx context.Context, tender model.Tender) (model.Tender, error) {
	path := "service.tender.CreateTender"
	isResponsible, err := tS.tenderRepository.IsUserResponsibleForOrganization(ctx, tender)
	if err != nil {
		return model.Tender{}, fmt.Errorf(path+".IsUserResponsibleForOrganization, error: {%s}",
			custom_errors.ErrAccessDenied.Error())
	}
	if !isResponsible {
		return model.Tender{}, fmt.Errorf("пользователь не связан с организацией")
	}
	return tS.tenderRepository.CreateTender(ctx, tender)
}

func (tS *TenderService) GetTender(ctx context.Context, user string, limit int, offset int) ([]model.Tender, error) {
	return tS.tenderRepository.GetTender(ctx, user, limit, offset)
}

func (tS *TenderService) UpdateTender(ctx context.Context, tender model.Tender) (model.Tender, error) {
	if tender.Status != "" {
		flag := helper.IsValidTenderStatus(tender.Status)
		if !flag {
			return model.Tender{}, custom_errors.ErrUnprocessableEntity
		}
	}

	return tS.tenderRepository.UpdateTender(ctx, tender)
}

func (tS *TenderService) GetStatus(ctx context.Context, tenderId uuid.UUID) (string, error) {
	return tS.tenderRepository.GetStatus(ctx, tenderId)
}

func (tS *TenderService) UpdateStatus(ctx context.Context, tender model.Tender) (model.Tender, error) {
	if tender.Status != "" {
		flag := helper.IsValidTenderStatus(tender.Status)
		if !flag {
			return model.Tender{}, custom_errors.ErrUnprocessableEntity
		}
	}
	return tS.tenderRepository.UpdateStatus(ctx, tender)
}
