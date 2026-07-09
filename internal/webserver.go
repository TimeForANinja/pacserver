package internal

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// LaunchServer starts the prod and admin listeners.
func LaunchServer() {
	conf := GetConfig()
	if conf == nil {
		log.Error("Server cannot start without a loaded config")
		os.Exit(1)
	}

	prodApp := newHTTPApp()
	adminApp := newHTTPApp()

	installMiddlewares(prodApp)
	installMiddlewares(adminApp)

	registerProdRoutes(prodApp)
	registerAdminRoutes(adminApp)

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	start := func(name string, app *fiber.App, port uint16) {
		if app == nil || port == 0 {
			return
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil {
				errCh <- fmt.Errorf("%s listener failed: %w", name, err)
			}
		}()
	}

	start("prod", prodApp, conf.Port)
	start("admin", adminApp, conf.AdminPort)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case err := <-errCh:
		LogUnexpectedError("server listen failed", err)
		os.Exit(1)
	case <-done:
		os.Exit(0)
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
