package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

type collector struct {
	pool *pgxpool.Pool

	acquire                 *prometheus.Desc
	acquireSeconds          *prometheus.Desc
	acquiredConnections     *prometheus.Desc
	canceledAcquire         *prometheus.Desc
	constructingConnections *prometheus.Desc
	emptyAcquire            *prometheus.Desc
	idleConnections         *prometheus.Desc
	maxConnections          *prometheus.Desc
	totalConnections        *prometheus.Desc
	newConnections          *prometheus.Desc
	maxLifetimeDestroy      *prometheus.Desc
	maxIdleDestroy          *prometheus.Desc
	emptyAcquireWaitSeconds *prometheus.Desc
}

var _ prometheus.Collector = (*collector)(nil)

// NewCollector creates a prometheus collector that exports metrics about the given [pgxpool.Pool] labeled by given `db_name`.
func NewCollector(pool *pgxpool.Pool, dbName string) prometheus.Collector {
	fqName := func(name string) string {
		return prometheus.BuildFQName("pgx", "pool", name)
	}

	return &collector{
		pool: pool,
		acquire: prometheus.NewDesc(
			fqName("acquire_total"),
			"The cumulative count of successful acquires from the pool.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		acquireSeconds: prometheus.NewDesc(
			fqName("acquire_seconds_total"),
			"The total duration of all successful acquires from the pool.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		acquiredConnections: prometheus.NewDesc(
			fqName("acquired_connections"),
			"The number of currently acquired connections in the pool.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		canceledAcquire: prometheus.NewDesc(
			fqName("canceled_acquire_total"),
			"The cumulative count of acquires from the pool that were canceled by a context.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		constructingConnections: prometheus.NewDesc(
			fqName("constructing_connections"),
			"The number of conns with construction in progress in the pool.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		emptyAcquire: prometheus.NewDesc(
			fqName("empty_acquire_total"),
			"The cumulative count of successful acquires from the pool that waited for a resource to be released or constructed because the pool was empty.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		idleConnections: prometheus.NewDesc(
			fqName("idle_connections"),
			"The number of currently idle conns in the pool.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		maxConnections: prometheus.NewDesc(
			fqName("max_connections"),
			"The maximum size of the pool.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		totalConnections: prometheus.NewDesc(
			fqName("total_connections"),
			"The total number of resources currently in the pool. The value is the sum of ConstructingConns, AcquiredConns, and IdleConns.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		newConnections: prometheus.NewDesc(
			fqName("new_connections_total"),
			"The cumulative count of new connections opened.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		maxLifetimeDestroy: prometheus.NewDesc(
			fqName("max_lifetime_destroy_total"),
			"The cumulative count of connections destroyed because they exceeded MaxConnLifetime.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		maxIdleDestroy: prometheus.NewDesc(
			fqName("max_idle_destroy_total"),
			"The cumulative count of connections destroyed because they exceeded MaxConnIdleTime.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		emptyAcquireWaitSeconds: prometheus.NewDesc(
			fqName("empty_acquire_wait_seconds_total"),
			"The cumulative time waited for successful acquires from the pool for a resource to be released or constructed because the pool was empty.",
			nil, prometheus.Labels{"db_name": dbName},
		),
	}
}

// Describe implements [prometheus.Collector].
func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.acquire
	ch <- c.acquireSeconds
	ch <- c.acquiredConnections
	ch <- c.canceledAcquire
	ch <- c.constructingConnections
	ch <- c.emptyAcquire
	ch <- c.idleConnections
	ch <- c.maxConnections
	ch <- c.totalConnections
	ch <- c.newConnections
	ch <- c.maxLifetimeDestroy
	ch <- c.maxIdleDestroy
	ch <- c.emptyAcquireWaitSeconds
}

// Collect implements [prometheus.Collector].
func (c *collector) Collect(ch chan<- prometheus.Metric) {
	stat := c.pool.Stat()
	ch <- prometheus.MustNewConstMetric(c.acquire, prometheus.CounterValue, float64(stat.AcquireCount()))
	ch <- prometheus.MustNewConstMetric(c.acquireSeconds, prometheus.CounterValue, float64(stat.AcquireDuration().Seconds()))
	ch <- prometheus.MustNewConstMetric(c.acquiredConnections, prometheus.GaugeValue, float64(stat.AcquiredConns()))
	ch <- prometheus.MustNewConstMetric(c.canceledAcquire, prometheus.CounterValue, float64(stat.CanceledAcquireCount()))
	ch <- prometheus.MustNewConstMetric(c.constructingConnections, prometheus.GaugeValue, float64(stat.ConstructingConns()))
	ch <- prometheus.MustNewConstMetric(c.emptyAcquire, prometheus.CounterValue, float64(stat.EmptyAcquireCount()))
	ch <- prometheus.MustNewConstMetric(c.idleConnections, prometheus.GaugeValue, float64(stat.IdleConns()))
	ch <- prometheus.MustNewConstMetric(c.maxConnections, prometheus.GaugeValue, float64(stat.MaxConns()))
	ch <- prometheus.MustNewConstMetric(c.totalConnections, prometheus.GaugeValue, float64(stat.TotalConns()))
	ch <- prometheus.MustNewConstMetric(c.newConnections, prometheus.CounterValue, float64(stat.NewConnsCount()))
	ch <- prometheus.MustNewConstMetric(c.maxLifetimeDestroy, prometheus.CounterValue, float64(stat.MaxLifetimeDestroyCount()))
	ch <- prometheus.MustNewConstMetric(c.maxIdleDestroy, prometheus.CounterValue, float64(stat.MaxIdleDestroyCount()))
	ch <- prometheus.MustNewConstMetric(c.emptyAcquireWaitSeconds, prometheus.CounterValue, float64(stat.EmptyAcquireWaitTime().Seconds()))
}
