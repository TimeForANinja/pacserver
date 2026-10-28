package internal

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/timeforaninja/pacserver/pkg/acl"
)

// LaunchServer starts the single listener that serves PAC lookups, the admin UI, and metrics.
func LaunchServer() {
	conf := GetConfig()
	if conf == nil {
		log.Error("Server cannot start without a loaded config")
		os.Exit(1)
	}

	app := newHTTPApp()
	installMiddlewares(app)
	// Fiber matches routes in registration order: the request instrumentation must
	// precede the admin routes so their traffic is counted as well, and the star
	// catch-all must be registered last or it would shadow the admin routes.
	trackPac := registerPrometheusMiddleware(app)
	// Build the source-network ACL once and share it across all protected
	// routes so they are gated by the same middleware instance.
	aclHandler, err := acl.Middleware(conf.AdminACLs)
	if err != nil {
		LogUnexpectedError("admin ACL configuration is invalid", err)
		os.Exit(1)
	}
	registerAdminRoutes(app, aclHandler)
	registerPrometheusEndpoint(app, aclHandler)
	registerCatchAllRoute(app, trackPac)

	if err := app.Listen(fmt.Sprintf(":%d", conf.Port)); err != nil {
		LogUnexpectedError("server listen failed", err)
		os.Exit(1)
	}
}

// newHTTPApp builds a Fiber app with the shared request limits and error handling.
func newHTTPApp() *fiber.App {
	return fiber.New(fiber.Config{
		EnablePrintRoutes: false,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			status := fiber.StatusInternalServerError
			message := "internal server error"
			if fiberErr, ok := err.(*fiber.Error); ok {
				status = fiberErr.Code
				message = fiberErr.Message
			}

			// A connection deadline has already expired, so there is no response left to
			// write. It is a normal client/network condition rather than an application fault.
			if status == fiber.StatusRequestTimeout {
				return nil
			}

			request := "request failed"
			if c != nil {
				request = fmt.Sprintf("request failed: %s %s from %s", c.Method(), c.OriginalURL(), c.IP())
			}
			if status >= fiber.StatusInternalServerError {
				LogUnexpectedError(request, err)
			} else {
				log.Warnf("%s: %v", request, err)
			}
			c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
			return c.Status(status).SendString(message)
		},
	})
}

// installMiddlewares wires panic recovery, compression, and access logging onto an app.
func installMiddlewares(app *fiber.App) {
	if app == nil {
		return
	}

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, panicValue interface{}) {
			request := "panic recovered"
			if c != nil {
				request = fmt.Sprintf("panic recovered during %s %s from %s", c.Method(), c.OriginalURL(), c.IP())
			}
			LogUnexpectedError(request, fmt.Errorf("%v", panicValue))
		},
	}))
	app.Use(compress.New())
	app.Use(accessLogMiddleware())
}
