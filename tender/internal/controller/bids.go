package controller

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	"zadanie-6105/helper"
	custom_errors "zadanie-6105/internal/custom-errors"
	"zadanie-6105/internal/model"
	"zadanie-6105/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/gookit/slog"
)

type bidsRoutes struct {
	bidsService service.IBids
}

func newBidsRoutes(g fiber.Router, bidsService service.IBids) {
	aR := &bidsRoutes{bidsService: bidsService}

	g.Post("/new", aR.newBids)
	g.Get("/my", aR.my)
	g.Get("/:tenderId/list", aR.bids)
	g.Patch("/:bidId/edit", aR.edit)
	g.Get("/:bidId/status", aR.status)
	g.Put("/:bidId/status", aR.editStatus)
	g.Put(":bidId/submit_decision", aR.submitDecision)
}

type bidsSliceResponse struct {
	Bids []bidsResponse `json:"bids"`
}
type bidsResponse struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Status          string    `json:"status"`
	Version         int       `json:"version"`
	CreatedAt       time.Time `json:"created_at"`
	CreatorUsername string    `json:"creatorUsername"`
}
type bidsParams struct {
	Bids *model.Bids `json:"bids"`
}

func (bR *bidsRoutes) newBids(ctx *fiber.Ctx) error {
	path := "internal.controller.bids.newBids"

	var bP bidsParams

	err := ctx.BodyParser(&bP.Bids)
	if err != nil {
		slog.Errorf(fmt.Errorf(path+".BodyParser, error: {%s}", err).Error())
		return wrapHttpError(ctx, 500, "internal error")
	}
	bP.Bids.Status = "Created"
	res, err := bR.bidsService.CreateBids(ctx.Context(), bP.Bids)
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

	resp := bidsResponse{
		ID:              res.ID,
		Name:            res.Title,
		Description:     res.Description,
		Status:          res.Status,
		Version:         res.Version,
		CreatedAt:       res.CreatedAt,
		CreatorUsername: res.CreatorUsername,
	}
	err = httpResponse(ctx, fiber.StatusOK, resp)
	if err != nil {
		return wrapHttpError(ctx, fiber.StatusInternalServerError, "internal server error")
	}
	return nil
}

func (bR *bidsRoutes) my(ctx *fiber.Ctx) error {
	path := "internal.controller.bids.my"
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
	res, err := bR.bidsService.GetBids(ctx.Context(), user["username"], limitInt, offsetInt)
	if err != nil {
		if errors.Is(err, custom_errors.ErrBidsNotFound) || len(res) == 0 {
			return wrapHttpError(ctx, 401, custom_errors.ErrBidsNotFound.Error())
		}
		slog.Errorf(path+".GetTender, error: {%s}", err)
		return err
	}
	resp := bidsSliceResponse{}

	for _, v := range res {
		resp.Bids = append(resp.Bids, bidsResponse{
			ID:              v.ID,
			Name:            v.Title,
			Description:     v.Description,
			Status:          v.Status,
			CreatorUsername: v.CreatorUsername,
			Version:         v.Version,
			CreatedAt:       v.CreatedAt,
		})
	}
	err = httpResponse(ctx, fiber.StatusOK, resp)
	if err != nil {
		return wrapHttpError(ctx, fiber.StatusInternalServerError, "internal server error")
	}
	return nil
}

func (bR *bidsRoutes) bids(ctx *fiber.Ctx) error {
	path := "internal.controller.bids.bids"
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
	if _, ok := user["username"]; !ok {
		return wrapHttpError(ctx, 400, custom_errors.ErrUnprocessableEntity.Error())
	}
	tenderId := ctx.Params("tenderId")
	parsedID, err := uuid.Parse(tenderId)
	if err != nil {
		return wrapHttpError(ctx, 400, "Invalid tenderId format")
	}
	res, err := bR.bidsService.GetBidsByTenderId(ctx.Context(), user["username"], parsedID, limitInt, offsetInt)
	if err != nil {
		slog.Errorf(path+".GetBidsByTenderId, error: {%s}", err.Error())
		if errors.Is(err, custom_errors.ErrTenderNotFound) {
			return wrapHttpError(ctx, 400, err.Error())
		}
		return err
	}
	resp := bidsSliceResponse{}

	for _, v := range res {
		resp.Bids = append(resp.Bids, bidsResponse{
			ID:              v.ID,
			Name:            v.Title,
			Description:     v.Description,
			Status:          v.Status,
			CreatorUsername: v.CreatorUsername,
			Version:         v.Version,
			CreatedAt:       v.CreatedAt,
		})
	}
	err = httpResponse(ctx, fiber.StatusOK, resp)
	if err != nil {
		return wrapHttpError(ctx, fiber.StatusInternalServerError, "internal server error")
	}
	return nil
}

func (bR *bidsRoutes) edit(ctx *fiber.Ctx) error {
	path := "internal.controller.bids.edit"

	var bP bidsParams
	err := ctx.BodyParser(&bP.Bids)
	if err != nil {
		slog.Errorf(fmt.Errorf(path+".BodyParser, error: {%s}", err).Error())
		return wrapHttpError(ctx, 400, "Invalid request format")
	}

	bidIdStr := ctx.Params("bidId")
	parsedID, err := uuid.Parse(bidIdStr)
	if err != nil {
		return wrapHttpError(ctx, 400, "Invalid bidId format")
	}

	bP.Bids.ID = parsedID

	username := ctx.Query("username")
	if username == "" {
		return wrapHttpError(ctx, 400, "Missing username parameter")
	}
	bP.Bids.CreatorUsername = username

	res, err := bR.bidsService.UpdateBids(ctx.Context(), bP.Bids)
	if err != nil {
		if errors.Is(err, custom_errors.ErrBidsNotFound) {
			return wrapHttpError(ctx, 404, "Bid not found")
		}
		if errors.Is(err, custom_errors.ErrAccessDenied) {
			return wrapHttpError(ctx, 403, "Insufficient permissions to edit the bid")
		}
		slog.Errorf(path+".UpdateBids, error: {%s}", err.Error())
		return wrapHttpError(ctx, 500, "Internal server error")
	}

	resp := bidsResponse{
		ID:              res.ID,
		Name:            res.Title,
		Description:     res.Description,
		Status:          res.Status,
		CreatorUsername: res.CreatorUsername,
		Version:         res.Version,
		CreatedAt:       res.CreatedAt,
	}
	err = httpResponse(ctx, fiber.StatusOK, resp)
	if err != nil {
		return wrapHttpError(ctx, fiber.StatusInternalServerError, "internal server error")
	}
	return nil
}

func (bR *bidsRoutes) status(c *fiber.Ctx) error {
	path := "internal.controller.bids.status"
	bidIdStr := c.Params("bidId")
	parsedID, err := uuid.Parse(bidIdStr)
	if err != nil {
		return wrapHttpError(c, 400, custom_errors.ErrUnprocessableEntity.Error())
	}
	user := c.Queries()
	if _, ok := user["username"]; !ok {
		return wrapHttpError(c, 400, custom_errors.ErrUnprocessableEntity.Error())
	}
	res, err := bR.bidsService.GetBidStatus(c.Context(), parsedID, user["username"])
	if err != nil {
		slog.Errorf(path+".GetStatus, error: {%s}", err.Error())
		return err
	}
	resp := res
	err = httpResponse(c, fiber.StatusOK, resp)
	if err != nil {
		return wrapHttpError(c, fiber.StatusInternalServerError, "internal server error")
	}
	return nil
}

func (bR *bidsRoutes) editStatus(ctx *fiber.Ctx) error {
	path := "internal.controller.bids.updateBidStatus"

	bidIdStr := ctx.Params("bidId")
	if bidIdStr == "" {
		return wrapHttpError(ctx, 400, custom_errors.ErrUnprocessableEntity.Error())
	}

	bidId, err := uuid.Parse(bidIdStr)
	if err != nil {
		return wrapHttpError(ctx, 400, custom_errors.ErrUnprocessableEntity.Error())
	}

	username := ctx.Query("username")
	if username == "" {
		return wrapHttpError(ctx, 400, custom_errors.ErrUnprocessableEntity.Error())
	}

	status := ctx.Query("status")
	if status == "" {
		return wrapHttpError(ctx, 400, custom_errors.ErrUnprocessableEntity.Error())
	}

	if !helper.IsValidBidsStatus(status) {
		return wrapHttpError(ctx, 400, custom_errors.ErrUnprocessableEntity.Error())
	}

	bid := model.Bids{
		ID:              bidId,
		Status:          status,
		CreatorUsername: username,
	}
	updatedBid, err := bR.bidsService.UpdateBidsStatus(ctx.Context(), bid)
	if err != nil {
		if errors.Is(err, custom_errors.ErrBidsNotFound) {
			return wrapHttpError(ctx, 404, custom_errors.ErrBidsNotFound.Error())
		}
		if errors.Is(err, custom_errors.ErrAccessDenied) {
			return wrapHttpError(ctx, 403, custom_errors.ErrAccessDenied.Error())
		}
		slog.Errorf(path+".UpdateBidsStatus, error: {%s}", err.Error())
		return wrapHttpError(ctx, 500, "Internal server error")
	}

	resp := bidsResponse{
		ID:              updatedBid.ID,
		Name:            updatedBid.Title,
		Status:          updatedBid.Status,
		CreatorUsername: updatedBid.CreatorUsername,
		Version:         updatedBid.Version,
		CreatedAt:       updatedBid.CreatedAt,
	}

	err = httpResponse(ctx, fiber.StatusOK, resp)
	if err != nil {
		return wrapHttpError(ctx, fiber.StatusInternalServerError, "internal server error")
	}
	return nil
}

func (bR *bidsRoutes) submitDecision(ctx *fiber.Ctx) error {
	path := "internal.controller.bids.submitDecision"

	bidIdStr := ctx.Params("bidId")
	decision := ctx.Query("decision")
	username := ctx.Query("username")
	if username == "" {
		return wrapHttpError(ctx, 400, custom_errors.ErrUnprocessableEntity.Error())
	}
	if decision == "" {
		return wrapHttpError(ctx, 400, custom_errors.ErrUnprocessableEntity.Error())
	}
	parsedID, err := uuid.Parse(bidIdStr)
	if err != nil {
		return wrapHttpError(ctx, 400, custom_errors.ErrUnprocessableEntity.Error())
	}
	res, err := bR.bidsService.UpdateBidsDecision(ctx.Context(), parsedID, decision, username)
	if err != nil {
		slog.Errorf(path+".UpdateBidsDecision, error: {%s}", err.Error())
		return wrapHttpError(ctx, 400, err.Error())
	}

	resp := bidsResponse{
		ID:              res.ID,
		Name:            res.Title,
		Description:     res.Description,
		Status:          res.Status,
		CreatorUsername: res.CreatorUsername,
		Version:         res.Version,
		CreatedAt:       res.CreatedAt,
	}
	err = httpResponse(ctx, fiber.StatusOK, resp)
	if err != nil {
		return wrapHttpError(ctx, fiber.StatusInternalServerError, "internal server error")
	}
	return nil
}
