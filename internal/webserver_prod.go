package internal

import (
	"github.com/gofiber/fiber/v2"
	"github.com/timeforaninja/pacserver/internal/storage"
)

// registerCatchAllRoute registers the open PAC lookup that answers every other GET.
// It must be registered last: Fiber matches routes in registration order and a star
// route matches any path, so an earlier catch-all would shadow the admin and metrics
// routes.
func registerCatchAllRoute(app *fiber.App, trackPac func(pac *storage.LookupEntry)) {
	if app == nil || GetConfig() == nil {
		return
	}

	app.Get("*", func(c *fiber.Ctx) error {
		return serveLookupRequest(c, trackPac, false)
	})
}
