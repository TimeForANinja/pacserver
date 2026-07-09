package internal

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/timeforaninja/pacserver/internal/storage"
	"github.com/timeforaninja/pacserver/pkg/IP"
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

	app.Get("/wpad.dat", func(c *fiber.Ctx) error {
		log.Debug("Received GET for /wpad.dat")
		ipStr, _ := extractIP(c)
		return servePAC(
			c,
			storage.WPAD(),
			make([]*storage.LookupEntry, 0),
			&IP.Net{},
			ipStr,
			32,
			trackPac,
		)
	})

	app.Get("/*", func(c *fiber.Ctx) error {
		return serveLookupRequest(c, trackPac)
	})
}
