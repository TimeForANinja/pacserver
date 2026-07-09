package internal

/**
 * This class handles all setup required
 * to expose Prometheus metrics
 */

import (
	"strconv"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/cakturk/go-netstat/netstat"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/timeforaninja/pacserver/internal/storage"
	myPrometheus "github.com/timeforaninja/pacserver/pkg/prometheus"
)

// Custom metrics for Prometheus
var (
	// Response time metrics
	// Histogram of request durations by fixed buckets.
	// Use this for fleet-wide aggregation and PromQL percentile calculations.
	responseTimeHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "app_response_time_hist_seconds",
		Help:    "Response time distribution in seconds",
		Buckets: []float64{1e-7, 5e-7, 1e-6, 5e-6, 1e-5, 5e-5, 0.0001, 0.0005, 0.001, 0.005, 0.01},
	})
	// Summary of request durations with client-side quantile estimation.
	// Useful for per-process percentiles but not aggregatable across instances.
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
			// Pull the current TCP socket table each scrape so the metric reflects live state.
			tabs, err := netstat.TCPSocks(netstat.NoopFilter)
			if err != nil {
				return make(map[string]float64)
			}
			// Fold the raw socket list into label counts for Prometheus.
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
		Name: "app_bytes_out",
		Help: "Total bytes sent",
	})

	// various other metrics are already tracked by fiberprometheus by default
	// those include cgo, memory and cpu times
)

// registerPrometheusMiddleware attaches the request counters to the prod listener.
func registerPrometheusMiddleware(app *fiber.App) func(pac *storage.LookupEntry) {
	// Return a no-op tracker when metrics are disabled so callers do not need branching.
	if app == nil || GetConfig() == nil || !GetConfig().PrometheusEnabled {
		return func(pac *storage.LookupEntry) {}
	}

	// Register the custom collectors once during startup even if we wire the middleware to multiple apps.
	prometheus.MustRegister(responseTimeHistogram)
	prometheus.MustRegister(responseTimeSummary)
	prometheus.MustRegister(openSocketCounter)
	prometheus.MustRegister(httpErrorCounter)
	prometheus.MustRegister(pacFileCounter)
	prometheus.MustRegister(dataInCounter)
	prometheus.MustRegister(dataOutCounter)

	// Measure request size, duration, response size, and final status for every request.
	app.Use(func(c *fiber.Ctx) error {
		// Record request size before the handler mutates the response.
		dataInCounter.Add(float64(len(c.Request().Body())))

		// Start timing as close to the handler invocation as possible.
		startTime := time.Now()

		// Run the handler chain and observe the result afterwards.
		err := c.Next()

		// Record response time after the handler completes.
		duration := time.Since(startTime).Seconds()
		responseTimeHistogram.Observe(duration)
		responseTimeSummary.Observe(duration)

		// Record response size from the final response body.
		dataOutCounter.Add(float64(len(c.Response().Body())))

		// Count the final HTTP status for error tracking and dashboards.
		httpErrorCounter.WithLabelValues(strconv.Itoa(c.Response().StatusCode())).Inc()

		return err
	})

	// Return a PAC tracker so the route handlers can attribute responses to file names.
	return func(pac *storage.LookupEntry) {
		if pac == nil {
			pacFileCounter.WithLabelValues("default").Inc()
		} else {
			pacFileCounter.WithLabelValues(pac.IPMap.Filename).Inc()
		}
	}
}

// registerPrometheusEndpoint exposes the scrape endpoint on the admin listener.
func registerPrometheusEndpoint(app *fiber.App) {
	if app == nil || GetConfig() == nil || !GetConfig().PrometheusEnabled {
		return
	}

	// Expose the scrape endpoint on the admin listener only.
	prom := fiberprometheus.New("pacserver")
	prom.RegisterAt(app, GetConfig().PrometheusPath)
}
