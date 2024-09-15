package controller

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	custom_errors "zadanie-6105/internal/custom-errors"
	"zadanie-6105/internal/model"
	"zadanie-6105/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/gookit/slog"
)

type tenderRoutes struct {
	tenderService service.ITender
}

func newTenderRoutes(g fiber.Router, tenderService service.ITender) {
	aR := &tenderRoutes{tenderService: tenderService}

	g.Get("/", aR.tenders)
	g.Post("/new", aR.tendersNew)
	g.Get("/my", aR.my)
	g.Patch("/:tenderId/edit", aR.edit)
	g.Get("/:tenderId/status", aR.status)
	g.Put("/:tenderId/status", aR.editStatus)
}

type tendersResponse struct {
	Tenders []tenderResponse `json:"tenders"`
}
type tenderResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ServiceType string    `json:"serviceType"`
	Status      string    `json:"status"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
}
type tenderParams struct {
	Tender model.Tender `json:"tender"`
}

func (tR *tenderRoutes) tenders(ctx *fiber.Ctx) error {
	path := "controller.tenders.tenders"
	m := ctx.Queries()

	limitStr := m["limit"]
	var limitInt int
	var err error

	if limitStr != "" {
		limitInt, err = strconv.Atoi(limitStr)
		if err != nil {
			slog.Errorf(path+".Atoi, error: {%s}", err)
			return wrapHttpError(ctx, 400, "Invalid limit parameter")
		}
	} else {
		limitInt = 5
	}
	offsetStr := m["offset"]
	var offsetInt int

	if offsetStr != "" {
		offsetInt, err = strconv.Atoi(offsetStr)
		if err != nil {
			slog.Errorf(path+".Atoi, error: {%s}", err)
			return wrapHttpError(ctx, 400, "Invalid offset parameter")
		}
	} else {
		offsetInt = 0
	}
	serviceTypes := m["serviceTypes"]
	var serviceTypesArr []string
	if serviceTypes != "" {
		serviceTypesArr = strings.Split(serviceTypes, ",") // Разделение по запятым
	}
	tenders, err := tR.tenderService.GetTenders(ctx.Context(), limitInt, offsetInt, serviceTypesArr)
	if err != nil {
		slog.Errorf(path+".Scan, error: {%s}", err)
		if errors.Is(err, custom_errors.ErrTenderNotFound) {
			return wrapHttpError(ctx, 400, err.Error())
		}
		return wrapHttpError(ctx, fiber.StatusInternalServerError, err.Error())
	}
	resp := tendersResponse{}

	for _, t := range tenders {
		resp.Tenders = append(resp.Tenders, tenderResponse{
			ID:          t.ID,
			Name:        t.Title,
			Description: t.Description,
			ServiceType: t.ServiceType,
			Status:      t.Status,
			Version:     t.Version,
			CreatedAt:   t.CreatedAt,
		})
	}
	err = httpResponse(ctx, fiber.StatusOK, resp)
	if err != nil {
		return wrapHttpError(ctx, fiber.StatusInternalServerError, "internal server error")
	}

	return nil
}

func (tR *tenderRoutes) tendersNew(ctx *fiber.Ctx) error {
	path := "controller.tenders.tendersNew"

	var tP tenderParams

	err := ctx.BodyParser(&tP.Tender)
	if err != nil {
		slog.Errorf(fmt.Errorf(path+".BodyParser, error: {%s}", err).Error())
		return wrapHttpError(ctx, 500, "internal error")
	}
	tP.Tender.Status = "Created"
	tender, err := tR.tenderService.CreateTender(ctx.Context(), tP.Tender)
	if err != nil {
		if errors.Is(err, custom_errors.ErrAccessDenied) {
			return wrapHttpError(ctx, 403, custom_errors.ErrAccessDenied.Error())
		}
		if errors.Is(err, custom_errors.ErrTenderAlreadyExists) {
			return wrapHttpError(ctx, 401, custom_errors.ErrTenderAlreadyExists.Error())
		}
		slog.Errorf(path+".CreateTender, error: {%s}", err.Error())
		return err
	}
	resp := tenderResponse{
		ID:          tender.ID,
		Name:        tender.Title,
		Description: tender.Description,
		ServiceType: tender.ServiceType,
		Status:      tender.Status,
		Version:     tender.Version,
		CreatedAt:   tender.CreatedAt,
	}

	err = httpResponse(ctx, fiber.StatusOK, resp)
	if err != nil {
		return wrapHttpError(ctx, fiber.StatusInternalServerError, "internal server error")
	}
	return nil
}

func (tR *tenderRoutes) my(ctx *fiber.Ctx) error {
	path := "controller.tenders.tendersNew"
	m := ctx.Queries()

	limitStr := m["limit"]
	var limitInt int
	var err error

	if limitStr != "" {
		limitInt, err = strconv.Atoi(limitStr)
		if err != nil {
			slog.Errorf(path+".Atoi, error: {%s}", err)
			return wrapHttpError(ctx, 400, custom_errors.ErrUnprocessableEntity.Error())
		}
	} else {
		limitInt = 5
	}
	offsetStr := m["offset"]
	var offsetInt int

	if offsetStr != "" {
		offsetInt, err = strconv.Atoi(offsetStr)
		if err != nil {
			slog.Errorf(path+".Atoi, error: {%s}", err)
			return wrapHttpError(ctx, 400, custom_errors.ErrUnprocessableEntity.Error())
		}
	} else {
		offsetInt = 0
	}
	user := ctx.Queries()
	tenders, err := tR.tenderService.GetTender(ctx.Context(), user["username"], limitInt, offsetInt)
	if err != nil {
		if errors.Is(err, custom_errors.ErrTenderNotFound) || len(tenders) == 0 {
			return wrapHttpError(ctx, 401, custom_errors.ErrTenderNotFound.Error())
		}
		slog.Errorf(path+".GetTender, error: {%s}", err)
		return err
	}
	resp := tendersResponse{}

	for _, t := range tenders {
		resp.Tenders = append(resp.Tenders, tenderResponse{
			ID:          t.ID,
			Name:        t.Title,
			Description: t.Description,
			ServiceType: t.ServiceType,
			Status:      t.Status,
			Version:     t.Version,
			CreatedAt:   t.CreatedAt,
		})
	}
	err = httpResponse(ctx, fiber.StatusOK, resp)
	if err != nil {
		return wrapHttpError(ctx, fiber.StatusInternalServerError, "internal server error")
	}
	return nil
}

func (tR *tenderRoutes) edit(ctx *fiber.Ctx) error {
	path := "controller.tenders.edit"

	var tP tenderParams
	username := ctx.Query("username")
	tP.Tender.CreatorUsername = username
	if tP.Tender.CreatorUsername == "" {
		return wrapHttpError(ctx, 400, custom_errors.ErrUnprocessableEntity.Error())
	}
	tenderIdStr := ctx.Params("tenderId")

	parsedID, err := uuid.Parse(tenderIdStr)
	if err != nil {
		return wrapHttpError(ctx, 400, "Invalid tenderId format")
	}
	tP.Tender.ID = parsedID
	err = ctx.BodyParser(&tP.Tender)
	if err != nil {
		slog.Errorf(fmt.Errorf(path+".BodyParser, error: {%s}", err).Error())
		return wrapHttpError(ctx, 500, "internal error")
	}
	res, err := tR.tenderService.UpdateTender(ctx.Context(), tP.Tender)
	if err != nil {
		slog.Errorf(path+".UpdateTender, error: {%s}", err.Error())

		if errors.Is(err, custom_errors.ErrTenderNotFound) {
			return wrapHttpError(ctx, 404, custom_errors.ErrTenderNotFound.Error())
		}

		if errors.Is(err, custom_errors.ErrUserNotFound) {
			return wrapHttpError(ctx, 400, custom_errors.ErrUserNotFound.Error())
		}
		if errors.Is(err, custom_errors.ErrAccessDenied) {
			return wrapHttpError(ctx, 403, custom_errors.ErrAccessDenied.Error())
		}
		if errors.Is(err, custom_errors.ErrUnprocessableEntity) {
			return wrapHttpError(ctx, 400, custom_errors.ErrUnprocessableEntity.Error())
		}
		return err
	}

	resp := tenderResponse{
		ID:          res.ID,
		Name:        res.Title,
		Description: res.Description,
		ServiceType: res.ServiceType,
		Status:      res.Status,
		Version:     res.Version,
		CreatedAt:   res.CreatedAt,
	}
	err = httpResponse(ctx, fiber.StatusOK, resp)
	if err != nil {
		return wrapHttpError(ctx, fiber.StatusInternalServerError, "internal server error")
	}
	return nil
}

func (tR *tenderRoutes) status(ctx *fiber.Ctx) error {
	path := "controller.tenders.status"

	tenderIdStr := ctx.Params("tenderId")
	tenderId, err := uuid.Parse(tenderIdStr)
	if err != nil {
		return wrapHttpError(ctx, 400, "Invalid tenderId format")
	}
	res, err := tR.tenderService.GetStatus(ctx.Context(), tenderId)
	if err != nil {
		if errors.Is(err, custom_errors.ErrTenderNotFound) {
			return wrapHttpError(ctx, 404, custom_errors.ErrTenderNotFound.Error())
		}
		slog.Errorf(path+".GetStatus, error: {%s}", err.Error())
		return err
	}
	resp := res
	err = httpResponse(ctx, fiber.StatusOK, resp)
	if err != nil {
		return wrapHttpError(ctx, fiber.StatusInternalServerError, "internal server error")
	}
	return nil
}

func (tR *tenderRoutes) editStatus(ctx *fiber.Ctx) error {
	path := "internal.controller.tenders.editStatus"
	var tP tenderParams

	username := ctx.Query("username")
	tP.Tender.CreatorUsername = username
	if tP.Tender.CreatorUsername == "" {
		return wrapHttpError(ctx, 400, custom_errors.ErrUnprocessableEntity.Error())
	}
	tenderIdStr := ctx.Params("tenderId")
	if tenderIdStr == "" {
		return wrapHttpError(ctx, 400, custom_errors.ErrUnprocessableEntity.Error())
	}

	parsedID, err := uuid.Parse(tenderIdStr)
	if err != nil {
		return wrapHttpError(ctx, 400, custom_errors.ErrUnprocessableEntity.Error())
	}
	tP.Tender.ID = parsedID

	status := ctx.Query("status")
	if status == "" {
		return wrapHttpError(ctx, 400, custom_errors.ErrUnprocessableEntity.Error())
	}
	tP.Tender.Status = status

	res, err := tR.tenderService.UpdateStatus(ctx.Context(), tP.Tender)
	if err != nil {
		slog.Errorf(path+".UpdateStatus, error: {%s}", err.Error())

		if errors.Is(err, custom_errors.ErrTenderNotFound) {
			return wrapHttpError(ctx, 404, custom_errors.ErrTenderNotFound.Error())
		}

		if errors.Is(err, custom_errors.ErrUserNotFound) {
			return wrapHttpError(ctx, 400, custom_errors.ErrUserNotFound.Error())
		}
		if errors.Is(err, custom_errors.ErrAccessDenied) {
			return wrapHttpError(ctx, 403, custom_errors.ErrAccessDenied.Error())
		}
		if errors.Is(err, custom_errors.ErrUnprocessableEntity) {
			return wrapHttpError(ctx, 400, custom_errors.ErrUnprocessableEntity.Error())
		}
		return wrapHttpError(ctx, 500, "Internal server error")
	}
	resp := tenderResponse{
		ID:          res.ID,
		Name:        res.Title,
		Description: res.Description,
		ServiceType: res.ServiceType,
		Status:      res.Status,
		Version:     res.Version,
		CreatedAt:   res.CreatedAt,
	}

	err = httpResponse(ctx, fiber.StatusOK, resp)
	if err != nil {
		return wrapHttpError(ctx, fiber.StatusInternalServerError, "internal server error")
	}
	return nil
}
