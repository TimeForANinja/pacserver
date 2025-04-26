package internal

/**
 * this file handles OS signals
 *
 * we support two signals:
 *  - SIGHUP: reload PACs
 *  - SIGINT/SIGTERM: gracefully shut down the server
 *
 * SIGINT/SIGTERM is mainly required to properly delete the PID file
 */

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"os"
	"os/signal"
	"syscall"
)

// setupSignalHandling sets up a goroutine to handle OS signals
func setupSignalHandling(app *fiber.App) {
	// Create a channel to receive signals
	sigs := make(chan os.Signal, 1)

	// Register for SIGHUP, SIGINT, and SIGTERM
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	// Start a goroutine to handle signals
	go func() {
		for {
			sig := <-sigs
			log.Infof("Received signal: %v", sig)

			if sig == syscall.SIGHUP {
				log.Info("Reloading PACs due to SIGHUP")

				// Reload PAC Zone & Files
				updateLookupTree()

				log.Info("PACs reloaded successfully")
			} else if sig == syscall.SIGINT || sig == syscall.SIGTERM {
				log.Info("Shutting down server due to signal")

				// Clean up PID file before exiting
				if err := RemovePidFile(); err != nil {
					log.Errorf("Failed to remove PID file: %v", err)
				} else {
					log.Info("PID file removed successfully")
				}

				// Gracefully shut down the server
				if err := app.Shutdown(); err != nil {
					log.Errorf("Error shutting down server: %v", err)
					os.Exit(1)
				}

				log.Info("Server has been gracefully shut down")
				os.Exit(0)
			}
		}
	}()
}
