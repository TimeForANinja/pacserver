package internal

import (
	"github.com/gofiber/fiber/v2"
	"github.com/timeforaninja/pacserver/internal/storage"
	"github.com/timeforaninja/pacserver/pkg/admin"
)

// registerAdminRoutes wires the admin dashboard, login, reload, and metrics routes.
func registerAdminRoutes(app *fiber.App) {
	if app == nil {
		return
	}

	conf := GetConfig()
	if conf == nil {
		return
	}

	// The admin listener serves the UI and reload controls; it does not collect request metrics.
	admin.RegisterAdminUIRoute(app, conf.AdminSecret, conf.PrometheusPath, conf.Port)
	admin.RegisterAdminLoginRoute(app, conf.AdminSecret)
	registerPrometheusEndpoint(app)

	// Admin reload updates the LUT in place without restarting the process.
	admin.RegisterAdminReloadRoute(app, conf.AdminSecret, func() error {
		storage.UpdateLookupTree(conf.ToStorageConfig())
		return nil
	})

	// fallback default route, doing the same PAC Lookup as prod
	app.Get("*", func(c *fiber.Ctx) error {
		return serveLookupRequest(c, nil, true)
	})
}
