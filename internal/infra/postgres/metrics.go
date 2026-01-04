package postgres

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// DBMetrics implements prometheus.Collector to scrape pgxpool stats.
type DBMetrics struct {
	pool Pooler
	log  *slog.Logger

	// Descriptors
	connectionsTotal         *prometheus.Desc
	connectionsInUse         *prometheus.Desc
	connectionsIdle          *prometheus.Desc
	connectionsMaxOpen       *prometheus.Desc
	waitCountTotal           *prometheus.Desc
	waitDurationSecondsTotal *prometheus.Desc
}

// NewDBMetrics creates a new database metrics collector.
func NewDBMetrics(pool Pooler, log *slog.Logger) *DBMetrics {
	const ns = "db"
	const sub = "pool"

	return &DBMetrics{
		pool: pool,
		log:  log,
		connectionsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "connections_total"),
			"Total number of connections in the pool.",
			nil, nil,
		),
		connectionsInUse: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "connections_in_use"),
			"Number of connections currently in use.",
			nil, nil,
		),
		connectionsIdle: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "connections_idle"),
			"Number of idle connections in the pool.",
			nil, nil,
		),
		connectionsMaxOpen: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "connections_max_open"),
			"Maximum number of open connections allowed.",
			nil, nil,
		),
		waitCountTotal: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "wait_count_total"),
			"Total number of times waited for a connection.",
			nil, nil,
		),
		waitDurationSecondsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "wait_duration_seconds_total"),
			"Total time waited for a connection in seconds.",
			nil, nil,
		),
	}
}

// Describe implements prometheus.Collector.
func (c *DBMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.connectionsTotal
	ch <- c.connectionsInUse
	ch <- c.connectionsIdle
	ch <- c.connectionsMaxOpen
	ch <- c.waitCountTotal
	ch <- c.waitDurationSecondsTotal
}

// Collect implements prometheus.Collector.
func (c *DBMetrics) Collect(ch chan<- prometheus.Metric) {
	// Safely get the underlying pool
	pgxPool := c.pool.Pool()
	if pgxPool == nil {
		// Underlying pool not initialized or closed, report nothing or zeros?
		// Reporting nothing is safer as 0 might be misleading for "MaxOpen"
		return
	}

	stats := pgxPool.Stat()

	ch <- prometheus.MustNewConstMetric(c.connectionsTotal, prometheus.GaugeValue, float64(stats.TotalConns()))
	ch <- prometheus.MustNewConstMetric(c.connectionsInUse, prometheus.GaugeValue, float64(stats.AcquiredConns()))
	ch <- prometheus.MustNewConstMetric(c.connectionsIdle, prometheus.GaugeValue, float64(stats.IdleConns()))
	ch <- prometheus.MustNewConstMetric(c.connectionsMaxOpen, prometheus.GaugeValue, float64(stats.MaxConns()))
	ch <- prometheus.MustNewConstMetric(c.waitCountTotal, prometheus.CounterValue, float64(stats.AcquireCount()))
	ch <- prometheus.MustNewConstMetric(c.waitDurationSecondsTotal, prometheus.CounterValue, stats.AcquireDuration().Seconds())
}
