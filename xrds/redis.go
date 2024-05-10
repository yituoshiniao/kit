package xrds

import (
	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"strconv"
	"time"
)

var MetricsEnable bool

func recordMetrics(conf Config, rds *redis.Client) {
	maxSize := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "redis_pool_max_size",
		Help: "max size of pool",
		ConstLabels: map[string]string{
			"redis_db":   strconv.Itoa(conf.DB),
			"redis_addr": conf.Addr,
		},
	})

	totalConns := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "redis_pool_total_conns",
		Help: "number of total connections in the pool",
		ConstLabels: map[string]string{
			"redis_db":   strconv.Itoa(conf.DB),
			"redis_addr": conf.Addr,
		},
	})
	idleConns := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "redis_pool_idle_conns",
		Help: "number of idle connections in the pool",
		ConstLabels: map[string]string{
			"redis_db":   strconv.Itoa(conf.DB),
			"redis_addr": conf.Addr,
		},
	})
	staleConns := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "redis_pool_stale_conns",
		Help: "number of stale connections removed from the pool",
		ConstLabels: map[string]string{
			"redis_db":   strconv.Itoa(conf.DB),
			"redis_addr": conf.Addr,
		},
	})

	hits := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "redis_pool_hits",
		Help: "number of times free connection was found in the pool",
		ConstLabels: map[string]string{
			"redis_db":   strconv.Itoa(conf.DB),
			"redis_addr": conf.Addr,
		},
	})

	misses := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "redis_pool_misses",
		Help: "number of times free connection was NOT found in the pool",
		ConstLabels: map[string]string{
			"redis_db":   strconv.Itoa(conf.DB),
			"redis_addr": conf.Addr,
		},
	})

	timeouts := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "redis_pool_timeouts",
		Help: "number of times a wait timeout occurred",
		ConstLabels: map[string]string{
			"redis_db":   strconv.Itoa(conf.DB),
			"redis_addr": conf.Addr,
		},
	})

	go func() {
		for range time.Tick(5 * time.Second) {
			totalConns.Set(float64(rds.PoolStats().TotalConns))
			idleConns.Set(float64(rds.PoolStats().IdleConns))
			staleConns.Set(float64(rds.PoolStats().StaleConns))
			hits.Set(float64(rds.PoolStats().Hits))
			misses.Set(float64(rds.PoolStats().Misses))
			timeouts.Set(float64(rds.PoolStats().Timeouts))
			maxSize.Set(float64(conf.PoolSize))
		}
	}()
}

func Open(conf Config) *redis.Client {
	rds := redis.NewClient(&redis.Options{
		Addr:       conf.Addr,
		Password:   conf.Password,
		DB:         conf.DB,
		MaxRetries: conf.MaxRetries,
		PoolSize:   conf.PoolSize,
	})

	MetricsEnable = conf.MetricsEnable
	recordMetrics(conf, rds)
	return rds
}
