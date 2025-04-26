package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

// GaugeVecFunc implements the prometheus.Collector interface
//
// unfortunately, there's no NewGaugeVecFunc build-in, that would allow a simple callback to collect metrics
// so we can either do it an interval-based (e.g., every second) or
// we have to implement the Collector interface ourselves
type GaugeVecFunc struct {
	metric   *prometheus.Desc
	callback func() map[string]float64
}

func NewGaugeVecFunc(
	opts prometheus.GaugeOpts,
	labelNames []string,
	callback func() map[string]float64,
) *GaugeVecFunc {
	desc := prometheus.V2.NewDesc(
		prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		opts.Help,
		prometheus.UnconstrainedLabels(labelNames),
		opts.ConstLabels,
	)
	return &GaugeVecFunc{
		metric:   desc,
		callback: callback,
	}
}

// Describe sends the super-set of all possible descriptors of metrics
func (c *GaugeVecFunc) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.metric
}

// Collect is called by the Prometheus registry when collecting metrics
func (c *GaugeVecFunc) Collect(ch chan<- prometheus.Metric) {
	values := c.callback()
	for label, value := range values {
		ch <- prometheus.MustNewConstMetric(c.metric, prometheus.GaugeValue, value, label)
	}
}
