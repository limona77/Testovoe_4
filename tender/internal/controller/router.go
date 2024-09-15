package controller

import (
	"zadanie-6105/internal/service"

	"github.com/gofiber/fiber/v2"
)

func NewRouter(app *fiber.App, services *service.Services) {
	tenders := app.Group("/api/tenders")
	newTenderRoutes(tenders, services.ITender)
	ping := app.Group("/api/ping")
	newPingRoutes(ping)
	bids := app.Group("/api/bids")
	newBidsRoutes(bids, services.IBids)
}
