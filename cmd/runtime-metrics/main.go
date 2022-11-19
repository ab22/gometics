package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func BuildLogger() (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	return cfg.Build()
}

func main() {
	logger, err := BuildLogger()

	if err != nil {
		panic("could not created zap logger: " + err.Error())
	}

	logger.Info("Starting Runtime Metrics...")

	var (
		signals   = make(chan os.Signal, 1)
		_, cancel = context.WithCancel(context.Background())
	)

	defer cancel()
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	sig, sigOK := <-signals
	logger.Info("Interrupt signal received",
		zap.String("signal", sig.String()),
		zap.Bool("signalOK", sigOK))
	logger.Info("Shutting down...")
}
