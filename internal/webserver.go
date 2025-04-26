package internal

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"strconv"
	"strings"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/cakturk/go-netstat/netstat"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/timeforaninja/pacserver/pkg/IP"
)

// Custom metrics for Prometheus
var (
	// Response time metrics
	responseTimeHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "app_response_time_hist_seconds",
		Help:    "Response time distribution in seconds",
		Buckets: []float64{1e-7, 5e-7, 1e-6, 5e-6, 1e-5, 5e-5, 0.0001, 0.0005, 0.001, 0.005, 0.01},
	})

	responseTimeSummary = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "app_response_time_summary_seconds",
		Help: "Response time distribution in seconds",
		Objectives: map[float64]float64{
			0.5:   0.05,   // 50th percentile (median) with 5% error
			0.9:   0.01,   // 90th percentile with 1% error
			0.99:  0.001,  // 99th percentile with 0.1% error
			0.999: 0.0001, // 99.9th percentile with 0.01% error
		},
	})

	// Response time metrics
	openSocketCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "app_socket_states",
			Help: "number of active sockets by state",
		},
		[]string{"state"},
	)

	// HTTP status code metrics
	httpErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_http_errors_total",
			Help: "Total number of HTTP status codes",
		},
		[]string{"status_code"},
	)

	// Response pac file metric
	pacFileCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_pac_file",
			Help: "Number of PAC files server.",
		},
		[]string{"file"},
	)

	// Data I/O metrics
	dataInCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "app_bytes_in",
		Help: "Total bytes received",
	})
	dataOutCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "app_data_out",
		Help: "Total bytes sent",
	})

	// various other metrics are already tracked by fiberprometheus by default
	// those include cgo, memory and cpu times
)

func init() {
	// Register custom metrics with Prometheus
	prometheus.MustRegister(responseTimeHistogram)
	prometheus.MustRegister(responseTimeSummary)
	prometheus.MustRegister(openSocketCounter)
	prometheus.MustRegister(httpErrorCounter)
	prometheus.MustRegister(pacFileCounter)
	prometheus.MustRegister(dataInCounter)
	prometheus.MustRegister(dataOutCounter)
}

// updateResourceMetrics periodically updates CPU, memory, and thread metrics
func updateResourceMetrics() {
	for {
		// list all the TCP sockets for your HTTP server
		tabs, err := netstat.TCPSocks(netstat.NoopFilter)
		if err == nil {
			// Create a map and count states
			stateCounts := make(map[string]float64)
			for _, tab := range tabs {
				state := tab.State.String()
				stateCounts[state]++
			}

			// Set all values at once, which should perform slightly better
			for state, count := range stateCounts {
				openSocketCounter.WithLabelValues(state).Set(count)
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func LaunchServer() {
	app := fiber.New(fiber.Config{
		// Enable tracking of response sizes for Prometheus metrics
		EnablePrintRoutes: false,

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

	// Start resource metrics collection in a separate goroutine
	go updateResourceMetrics()

	// Enable transport Compression
	app.Use(compress.New())

	// middleware to write access log
	app.Use(logger.New(logger.Config{
		// For more options, see the Config section
		Format:     "${time} ${ip} ${status} - ${method} ${path}\n",
		TimeFormat: "2006-Jan-02 15:04:05",
		Output:     getAccessLogger(),
	}))

	// Add middleware to track response times and errors
	app.Use(func(c *fiber.Ctx) error {
		// Record request size
		dataInCounter.Add(float64(len(c.Request().Body())))

		// Start timer for response time
		startTime := time.Now()

		// Process request
		err := c.Next()

		// Record response time
		duration := time.Since(startTime).Seconds()
		responseTimeHistogram.Observe(duration)
		responseTimeSummary.Observe(duration)

		// Record response size
		dataOutCounter.Add(float64(len(c.Response().Body())))

		// Track HTTP codes
		httpErrorCounter.WithLabelValues(strconv.Itoa(c.Response().StatusCode())).Inc()

		return err
	})

	// Setup Prometheus middleware if enabled
	if GetConfig().PrometheusEnabled {
		prom := fiberprometheus.New("pacserver")
		prom.RegisterAt(app, GetConfig().PrometheusPath)
	}

	// Define (testing) route where the IP is passed as a parameter
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

	// Start the server
	log.Fatal(app.Listen(fmt.Sprintf(":%d", GetConfig().Port)))
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

	// Track which PAC file was served
	pacFileCounter.WithLabelValues(pac.IPMap.Filename).Inc()

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
			log.Errorf("Error marshaling debug JSON: %v", err)
			return err
		}

		return c.SendString(string(jsonData) + "\n\n---------------------------------------\n\n" + pac.getVariant())
	} else {
		c.Type("application/x-ns-proxy-autoconfig")
		return c.SendString(pac.getVariant())
	}
}
