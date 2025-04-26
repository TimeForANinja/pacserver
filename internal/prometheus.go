package internal

/**
 * This class handles all setup required
 * to expose Prometheus metrics
 */

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/cakturk/go-netstat/netstat"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	myPrometheus "github.com/timeforaninja/pacserver/pkg/prometheus"
	"strconv"
	"time"
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
	openSocketCounter = myPrometheus.NewGaugeVecFunc(
		prometheus.GaugeOpts{
			Name: "app_socket_states",
			Help: "number of active sockets by state",
		},
		[]string{"state"},
		func() map[string]float64 {
			// list all the TCP sockets for your HTTP server
			tabs, err := netstat.TCPSocks(netstat.NoopFilter)
			if err != nil {
				return make(map[string]float64)
			}
			// Create a map and count states
			stateCounts := make(map[string]float64)
			for _, tab := range tabs {
				state := tab.State.String()
				stateCounts[state]++
			}
			return stateCounts
		},
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

func setupPrometheus(app *fiber.App) func(pac *LookupElement) {
	// skip Prometheus setup if not enabled
	if !GetConfig().PrometheusEnabled {
		return func(pac *LookupElement) {}
	}

	// Register custom metrics with Prometheus
	prometheus.MustRegister(responseTimeHistogram)
	prometheus.MustRegister(responseTimeSummary)
	prometheus.MustRegister(openSocketCounter)
	prometheus.MustRegister(httpErrorCounter)
	prometheus.MustRegister(pacFileCounter)
	prometheus.MustRegister(dataInCounter)
	prometheus.MustRegister(dataOutCounter)

	// register prometheus app route
	prom := fiberprometheus.New("pacserver")
	prom.RegisterAt(app, GetConfig().PrometheusPath)

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

	// return a func to track PAC Files chosen
	return func(pac *LookupElement) {
		if pac == nil {
			pacFileCounter.WithLabelValues("default").Inc()
		} else {
			pacFileCounter.WithLabelValues(pac.IPMap.Filename).Inc()
		}
	}
}
