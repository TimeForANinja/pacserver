package internal

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/timeforaninja/pacserver/pkg/IP"
)

func LaunchServer() {
	app := fiber.New()

	// Enable transport Compression
	app.Use(compress.New())

	// middleware to write access log
	app.Use(logger.New(logger.Config{
		// For more options, see the Config section
		Format:     "${time} ${ip} ${status} - ${method} ${path}\n",
		TimeFormat: "2006-Jan-02 15:04:05",
		Output:     getAccessLogger(),
	}))

	// Define (testing) route where the IP is passed as parameter
	app.Get("/:ip", func(c *fiber.Ctx) error {
		ip := c.Params("ip")

		// check the ip syntax
		// if it fails we default to the / route
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

		return getFileForIP(c, strings.Join(octets, "."), cidr)
	})

	// second testing route allowing for ip and cidr
	app.Get("/:ip/:cidr", func(c *fiber.Ctx) error {
		ip := c.Params("ip")

		// (try to) read cidr
		// if it fails we default to the /:ip route
		cidr, err := strconv.Atoi(c.Params("cidr"))
		if err != nil {
			return c.Next()
		}

		// check the ip syntax
		// if it fails we default to the /:ip and then the / route
		if !IP.IsValidPartialIP(ip) {
			return c.Next()
		}

		// Pad the IP to always be 4 octets
		octets := strings.Split(ip, ".")
		for len(octets) < 4 {
			octets = append(octets, "0")
		}

		return getFileForIP(c, strings.Join(octets, "."), cidr)
	})

	// Default route for handling requests with no path
	// use the requesters source ip
	app.Get("/", func(c *fiber.Ctx) error {
		return getFileForIP(c, c.IP(), 32)
	})

	app.Listen(fmt.Sprintf(":%d", conf.Port))
}

func getFileForIP(c *fiber.Ctx, ipStr string, networkBits int) error {
	ipNet, err := IP.NewIPNetFromMixed(ipStr, networkBits)
	if err != nil {
		// TODO: fallback to default PAC
		return err
	}

	// search db for best pac
	pac := findInTree(lookupTree, &ipNet)

	// TODO: fallback to default PAC
	if pac == nil {
		pac = &LookupElement{IPMap: &ipMap{}}
	}

	if _, isDebug := c.Queries()["debug"]; isDebug {
		jsonData, err := json.MarshalIndent(fiber.Map{
			"raw_requester": fiber.Map{
				"ip":   ipStr,
				"cidr": networkBits,
			},
			"parsed_requester": ipNet,
			"pac":              pac,
		}, "", "\t")
		if err != nil {
			return err
		}

		return c.SendString(string(jsonData) + "\n\n---------------------------------------\n\n" + pac.getVariant())
	} else {
		c.Type("application/x-ns-proxy-autoconfig")
		return c.SendString(pac.getVariant())
	}
}
