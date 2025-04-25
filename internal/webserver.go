package internal

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
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
		Name:    "pacserver_response_time_seconds",
		Help:    "Response time distribution in seconds",
		Buckets: prometheus.DefBuckets,
	})

	// CPU and Memory metrics
	cpuUsageGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "pacserver_cpu_usage_percent",
		Help: "Current CPU usage percentage",
	})

	memUsageGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "pacserver_memory_usage_bytes",
		Help: "Current memory usage in bytes",
	})

	// Thread count metric
	threadCountGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "pacserver_thread_count",
		Help: "Current number of goroutines",
	})

	// HTTP error rate metric
	httpErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pacserver_http_errors_total",
			Help: "Total number of HTTP errors",
		},
		[]string{"status_code"},
	)

	// Data I/O metrics
	dataInCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pacserver_data_in_bytes_total",
		Help: "Total bytes received",
	})

	dataOutCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pacserver_data_out_bytes_total",
		Help: "Total bytes sent",
	})
)

func init() {
	// Register custom metrics with Prometheus
	prometheus.MustRegister(responseTimeHistogram)
	prometheus.MustRegister(cpuUsageGauge)
	prometheus.MustRegister(memUsageGauge)
	prometheus.MustRegister(threadCountGauge)
	prometheus.MustRegister(httpErrorCounter)
	prometheus.MustRegister(dataInCounter)
	prometheus.MustRegister(dataOutCounter)
}

// updateResourceMetrics periodically updates CPU, memory, and thread metrics
func updateResourceMetrics() {
	for {
		// Update thread count
		threadCountGauge.Set(float64(runtime.NumGoroutine()))

		// Update memory usage
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		memUsageGauge.Set(float64(memStats.Alloc))

		// CPU usage is more complex and would require additional libraries
		// For simplicity, we're not implementing actual CPU usage here
		// In a production environment, you might use a library like gopsutil

		time.Sleep(15 * time.Second)
	}
}

func LaunchServer() {
	app := fiber.New(fiber.Config{
		// Enable tracking of response sizes for Prometheus metrics
		EnablePrintRoutes: false,
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

		// Record response size
		dataOutCounter.Add(float64(len(c.Response().Body())))

		// Record HTTP errors
		if c.Response().StatusCode() >= 400 {
			httpErrorCounter.WithLabelValues(strconv.Itoa(c.Response().StatusCode())).Inc()
		}

		return err
	})

	// Setup Prometheus middleware if enabled
	if conf.PrometheusEnabled {
		prometheus := fiberprometheus.New("pacserver")
		prometheus.RegisterAt(app, conf.PrometheusPath)
	}

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
