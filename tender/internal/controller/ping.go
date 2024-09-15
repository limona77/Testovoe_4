package controller

import (
	"github.com/gofiber/fiber/v2"
)

type PingRoutes struct{}

func newPingRoutes(g fiber.Router) {
	aR := &tenderRoutes{}

	g.Get("/", aR.ping)
}

func (tR *tenderRoutes) ping(c *fiber.Ctx) error {
	return c.SendString("ok")
}
