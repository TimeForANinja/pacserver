package internal

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// LaunchServer starts the HTTP server with request logging and panic recovery.
func LaunchServer() {
	conf := GetConfig()
	if conf == nil {
		log.Error("Server cannot start without a loaded config")
		os.Exit(1)
	}

	// Keep the server setup in one place: config, shared middleware, and route wiring.
	app := fiber.New(fiber.Config{
		EnablePrintRoutes: false,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			request := "request failed"
			if c != nil {
				request = fmt.Sprintf("request failed: %s %s from %s", c.Method(), c.OriginalURL(), c.IP())
			}
			LogUnexpectedError(request, err)
			if fiberErr, ok := err.(*fiber.Error); ok {
				return fiberErr
			}
			return fiber.NewError(fiber.StatusInternalServerError, "internal server error")
		},
	})

	// Install the safety middleware before any route-specific behavior.
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
	app.Use(logger.New(logger.Config{
		Format:     "${time} ${ip} ${status} - ${method} ${path}\n",
		TimeFormat: "2006-Jan-02 15:04:05",
		Output:     getAccessLogger(),
	}))

	// Route registration lives elsewhere so the bootstrap stays small and readable.
	registerRoutes(app)

	// Start the server only after all middleware and routes are in place.
	if err := app.Listen(fmt.Sprintf(":%d", conf.Port)); err != nil {
		LogUnexpectedError("server listen failed", err)
		os.Exit(1)
	}
}
