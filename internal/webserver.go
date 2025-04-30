package internal

/**
 * the webserver includes everything required for the webserver
 * this means creation of the webserver, registering routes,
 * and the main "getFileForIP" function to reply to user requests
 */

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/timeforaninja/pacserver/pkg/IP"
)

func LaunchServer() {
	// Write PID file for signal-based reloading
	if err := WritePidFile(); err != nil {
		log.Errorf("Failed to write PID file: %v", err)
	}

	app := fiber.New(fiber.Config{
		// Enable tracking of response sizes for Prometheus metrics
		EnablePrintRoutes: false,

		// TODO: check if setting "Prefork: true," improves performance. First tests look like it doesn't

		// ReadTimeout: Maximum duration for reading the entire request (including body).
		// If a client takes longer than this to send their request, the connection is closed.
		ReadTimeout: 5 * time.Second,
		// WriteTimeout: Maximum duration for writing the response to the client.
		// If the server takes longer than this to send the complete response, the connection is terminated.
		WriteTimeout: 10 * time.Second,
		// IdleTimeout: Maximum time to wait for the next request when keep-alive is enabled.
		// Connection is closed if no new request is received within this duration.
		IdleTimeout: 120 * time.Second,
	})

	// Set up signal handling for SIGHUP, SIGINT, and SIGTERM
	setupSignalHandling(app)

	// Enable transport Compression
	app.Use(compress.New())

	// middleware to write access log
	app.Use(logger.New(logger.Config{
		// For more options, see the Config section
		Format:     "${time} ${ip} ${status} - ${method} ${path}\n",
		TimeFormat: "2006-Jan-02 15:04:05",
		Output:     getAccessLogger(),
	}))

	trackPac := setupPrometheus(app)

	// Route for serving wpad.dat file
	app.Get("/wpad.dat", func(c *fiber.Ctx) error {
		log.Debug("Received for /wpad.dat")
		return servePAC(
			c,
			wpadPAC,
			make([]*LookupElement, 0),
			&IP.Net{},
			"", 0,
			trackPac,
		)
	})

	// Define (testing) route where the IP is passed as a parameter
	app.Get("/:ip", func(c *fiber.Ctx) error {
		ip := c.Params("ip")
		log.Debugf("Received for /:ip with ip=%s", ip)

		// check the ip syntax
		// if it fails we default to the "/" route
		if !IP.IsValidPartialIP(ip) {
			return c.Next()
		}

		// Split the IP into octets
		octets := strings.Split(ip, ".")
		cidr := len(octets) * 8

		// Pad the IP to always be 4 octets
		for len(octets) < 4 {
			octets = append(octets, "0")
		}

		return serveFromIP(c, strings.Join(octets, "."), cidr, trackPac)
	})

	// second testing route allowing for ip and cidr
	app.Get("/:ip/:cidr", func(c *fiber.Ctx) error {
		ip := c.Params("ip")
		log.Debugf("Received for /:ip/:cidr with ip=%s and cidr=%s", ip, c.Params("cidr"))

		// (try to) read cidr
		// if it fails we default to the "/:ip" route
		cidr, err := strconv.Atoi(c.Params("cidr"))
		if err != nil {
			return c.Next()
		}

		// check the ip syntax
		// if it fails we default to the "/:ip" and then the "/" route
		if !IP.IsValidPartialIP(ip) {
			return c.Next()
		}

		// Pad the IP to always be 4 octets
		octets := strings.Split(ip, ".")
		for len(octets) < 4 {
			octets = append(octets, "0")
		}

		return serveFromIP(c, strings.Join(octets, "."), cidr, trackPac)
	})

	// Default route for handling requests with no path
	// use the requesters source ip
	app.Get("/", func(c *fiber.Ctx) error {
		log.Debug("Received for /")
		return serveFromIP(c, c.IP(), 32, trackPac)
	})

	// Start the server
	err := app.Listen(fmt.Sprintf(":%d", GetConfig().Port))
	if err != nil {
		log.Errorf("Server error: %v", err)
		// Ensure PID file is removed before exiting
		err2 := RemovePidFile()
		if err2 != nil {
			log.Errorf("Failed to remove PID file: %v", err2)
		}
		os.Exit(1)
	}
}

// getFileForIP is the main function that resolves the PAC file for a given IP
func serveFromIP(c *fiber.Ctx, ipStr string, networkBits int, trackPac func(pac *LookupElement)) error {
	log.Debugf("Received request for IP: %s, Bits: %d", ipStr, networkBits)
	pac, ipNet, stackTrace := findPAC(ipStr, networkBits)

	return servePAC(c, pac, stackTrace, ipNet, ipStr, networkBits, trackPac)
}

func servePAC(
	c *fiber.Ctx,
	pac *LookupElement,
	stackTrace []*LookupElement,
	ipNet *IP.Net,
	ipStr string, networkBits int,
	trackPac func(pac *LookupElement),
) error {
	// Track which PAC file was served
	trackPac(pac)

	// Check for any case variation of "debug"
	hasDebug := false
	for key := range c.Queries() {
		if strings.EqualFold(key, "debug") {
			hasDebug = true
			break
		}
	}
	if hasDebug {
		pacMeta, err := json.MarshalIndent(fiber.Map{
			"requested":        fmt.Sprintf("%s/%d", ipStr, networkBits),
			"parsed_requested": ipNet.ToString(),
			"pac":              pac._stringify(),
		}, "", "\t")
		if err != nil {
			log.Errorf("Error marshaling debug JSON: %v", err)
			return err
		}

		treeMeta := _stringifyLookupStack(stackTrace)

		c.Set("content-type", "text/plain")
		return c.SendString(
			strings.Join([]string{
				string(pacMeta),
				treeMeta,
				pac.getVariant(),
			},
				"\n\n---------------------------------------\n\n",
			))
	} else {
		c.Set("content-type", "application/x-ns-proxy-autoconfig")
		return c.SendString(pac.getVariant())
	}
}

func findPAC(ipStr string, networkBits int) (*LookupElement, *IP.Net, []*LookupElement) {
	ipNet, err := IP.NewIPNetFromMixed(ipStr, networkBits)
	if err != nil {
		// fallback to the root/default node with the default pac
		return lookupTree.data, &IP.Net{}, make([]*LookupElement, 0)
	}

	// search db for best pac
	pac, stackTrace := findInTree(lookupTree, &ipNet)

	return pac, &ipNet, stackTrace
}
