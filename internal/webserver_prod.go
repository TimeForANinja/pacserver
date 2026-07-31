package internal

import (
	"github.com/gofiber/fiber/v2"
)

// registerProdRoutes wires the PAC listener, including the WPAD and lookup routes.
func registerProdRoutes(app *fiber.App) {
	if app == nil {
		return
	}

	if GetConfig() == nil {
		return
	}

	trackPac := registerPrometheusMiddleware(app)

	app.Get("*", func(c *fiber.Ctx) error {
		return serveLookupRequest(c, trackPac, false)
	})
}
