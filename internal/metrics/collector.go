package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Collector interface {
	// Starts a new goroutine which will collect and send all
	// metrics. Multiple calls to this method are possible but
	// the collector can only be started once. Once the collector
	// has been stopped, it cannot be restarted.
	Start()

	// Stop performs a graceful shutdown. It blocks until all
	// resources have been cleaned up or returns an error.
	Stop(ctx context.Context) error
}

type collector struct {
	logger   *zap.Logger
	interval time.Duration
	server   *http.Server

	// Start/Stop controls
	stopCh  chan struct{}
	started atomic.Bool
	stopped atomic.Bool

	// Prometheus Metrics
	mGoroutines     prometheus.Gauge
	mHeapAlloc      prometheus.Gauge
	mSys            prometheus.Gauge
	mPauseTotalNano prometheus.Gauge
	mNumGCs         prometheus.Gauge
}

type CollectorOpts struct {
	Logger          *zap.Logger
	Addr            string
	MetricNamespace string
	MetricInterval  time.Duration
}

func NewCollector(opts CollectorOpts) (Collector, error) {
	if opts.Logger == nil {
		return nil, fmt.Errorf("could not create collector: %w", ErrNilLogger)
	} else if opts.MetricInterval == 0 {
		opts.MetricInterval = 5 * time.Second
	}

	r := mux.NewRouter()
	r.Handle("/metrics", promhttp.Handler())

	c := &collector{
		logger:   opts.Logger,
		stopCh:   make(chan struct{}),
		interval: opts.MetricInterval,
		mGoroutines: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: opts.MetricNamespace,
			Name:      "goroutines",
		}),
		mHeapAlloc: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: opts.MetricNamespace,
			Name:      "heap_alloc",
		}),
		mSys: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: opts.MetricNamespace,
			Name:      "heap_sys",
		}),
		mPauseTotalNano: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: opts.MetricNamespace,
			Name:      "gc_pause_total_ns",
		}),
		mNumGCs: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: opts.MetricNamespace,
			Name:      "gc_completed_cycles",
		}),
		server: &http.Server{
			Addr:    opts.Addr,
			Handler: r,
		},
	}

	if err := c.registerMetrics(); err != nil {
		return nil, fmt.Errorf("could not create collector: %w", err)
	}

	return c, nil
}

func (c *collector) registerMetrics() error {
	var err error

	collectors := []prometheus.Collector{
		c.mGoroutines,
		c.mHeapAlloc,
		c.mSys,
		c.mPauseTotalNano,
		c.mNumGCs,
	}

	for _, collector := range collectors {
		if err = prometheus.Register(collector); err != nil {
			return err
		}
	}

	return nil
}

func (c *collector) Start() {
	if swapped := c.started.CompareAndSwap(false, true); !swapped {
		return
	}

	go func() {
		if err := c.server.ListenAndServe(); err != nil {
			c.logger.Error("failed to stop prometheus server", zap.Error(err))
		}
	}()

	go func() {
		for {
			select {
			case <-c.stopCh:
				return

			case <-time.After(c.interval):
				c.CollectRuntimeMetrics()
			}
		}
	}()
}

func (c *collector) Stop(ctx context.Context) error {
	if swapped := c.stopped.CompareAndSwap(false, true); !swapped {
		return nil
	}

	select {
	case c.stopCh <- struct{}{}:
	case <-ctx.Done():
		return fmt.Errorf("failed to stop collector: %w", ctx.Err())
	}

	if err := c.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to stop collector: %w", err)
	}

	return nil
}
