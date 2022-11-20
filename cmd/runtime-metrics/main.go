package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ab22/gometrics/internal/metrics"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// MetricsNamespace for all app specific metrics.
	metricsNamespace = "gometrics"

	// Addr for the prometheus endpoint.
	addr = ":8080"

	// Interval used to report metrics.
	interval = 5 * time.Second

	// shutdownTimeout specifies how long we should
	// wait for the collector to shutdown.
	shutdownTimeout = 10 * time.Second
)

func MustBuildLogger() *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := cfg.Build()

	if err != nil {
		panic("could not created zap logger: " + err.Error())
	}

	return logger
}

func main() {
	var (
		signals        = make(chan os.Signal, 1)
		logger         = MustBuildLogger()
		ctx, cancel    = context.WithCancel(context.Background())
		collector, err = metrics.NewCollector(metrics.CollectorOpts{
			Logger:          logger,
			Addr:            addr,
			MetricNamespace: metricsNamespace,
			MetricInterval:  interval,
		})
	)

	logger.Info("Starting Runtime Metrics...",
		zap.String("collector_addr", addr),
		zap.String("collector_interval", interval.String()))

	defer cancel()

	if err != nil {
		logger.Fatal("failed to create collector", zap.Error(err))
	}

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Starting collector...",
		zap.String("interval", interval.String()))
	collector.Start()

	sig, sigOK := <-signals
	logger.Info("Interrupt signal received",
		zap.String("signal", sig.String()),
		zap.Bool("signalOK", sigOK))

	collectorCtx, collectorCancel := context.WithTimeout(ctx, shutdownTimeout)
	defer collectorCancel()

	logger.Info("Stopping metrics collector...",
		zap.String("timeout", shutdownTimeout.String()))
	if err = collector.Stop(collectorCtx); err != nil {
		logger.Fatal("could not stop collector", zap.Error(err))
	}

	logger.Info("Shutting down collector.")
}
