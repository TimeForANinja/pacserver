package internal

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/timeforaninja/pacserver/internal/admin"
	"github.com/timeforaninja/pacserver/internal/storage"
)

// registerAdminRoutes wires the admin UI, login, reload, and debug-lookup
// routes onto the shared listener. The caller builds the source-network ACL
// middleware once and passes it in so every protected route shares the same gate.
func registerAdminRoutes(app *fiber.App, aclHandler fiber.Handler) {
	if app == nil {
		return
	}

	conf := GetConfig()
	if conf == nil {
		return
	}

	// All admin UI routes live under /admin so the public catch-all stays the PAC lookup.
	adminGroup := app.Group("/admin", aclHandler)
	admin.RegisterAdminUIRoute(adminGroup, conf.AdminSecret, conf.PrometheusPath)
	admin.RegisterAdminLoginRoute(adminGroup, conf.AdminSecret)

	// Admin reload updates the LUT in place without restarting the process.
	admin.RegisterAdminReloadRoute(adminGroup, conf.AdminSecret, func() error {
		storage.UpdateLookupTree(conf.ToStorageConfig())
		return nil
	})

	// Expose debug PAC lookups below /pac while keeping the generic resolver
	// unaware of the admin-only prefix.
	app.Get("/pac/*", aclHandler, func(c *fiber.Ctx) error {
		originalPath := c.Path()
		lookupPath := strings.TrimPrefix(originalPath, "/pac")
		if lookupPath == "" {
			lookupPath = "/"
		}
		c.Path(lookupPath)
		defer c.Path(originalPath)
		return serveLookupRequest(c, nil, true)
	})
}
