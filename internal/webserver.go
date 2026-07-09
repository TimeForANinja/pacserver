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

func LaunchServer() {
	// Keep the server setup in one place: config, shared middleware, and route wiring.
	app := fiber.New(fiber.Config{
		EnablePrintRoutes: false,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	})

	// Install the safety middleware before any route-specific behavior.
	app.Use(recover.New())
	app.Use(compress.New())
	app.Use(logger.New(logger.Config{
		Format:     "${time} ${ip} ${status} - ${method} ${path}\n",
		TimeFormat: "2006-Jan-02 15:04:05",
		Output:     getAccessLogger(),
	}))

	// Route registration lives elsewhere so the bootstrap stays small and readable.
	registerRoutes(app)

	// Start the server only after all middleware and routes are in place.
	if err := app.Listen(fmt.Sprintf(":%d", GetConfig().Port)); err != nil {
		log.Errorf("Server error: %v", err)
		os.Exit(1)
	}
}
