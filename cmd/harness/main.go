package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/justinbather/harness"
	"github.com/justinbather/harness/internal/logger"
)

func main() {
	ctx := context.Background()
	logger, ctx := logger.FromCtx(ctx)

	brokers := []string{"localhost:9092"}
	harness, err := harness.New(brokers)
	if err != nil {
		panic(err)
	}

	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt)

	harness.Start(ctx)

	<-sigs
	logger.Info("got shutdown signal")
	done := make(chan struct{})
	go func() {
		defer close(done)
		harness.Shutdown(ctx)
	}()

	select {
	case <-sigs:
		logger.Info("got second interrupt; forcing shutdown")
	case <-done:
	}
}
