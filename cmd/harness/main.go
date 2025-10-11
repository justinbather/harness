package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/justinbather/harness/internal/harness"
	"github.com/justinbather/harness/internal/logger"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	log, ctx := logger.FromCtx(ctx)

	harness, err := harness.New("localhost:9092")
	if err != nil {
		panic(err)
	}

	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt)

	go pollDB(harness)

	harness.Start(ctx)

	<-sigs
	log.Info("got shutdown signal")
	done := make(chan struct{})
	go func() {
		defer close(done)
		harness.Shutdown(ctx)
	}()

	select {
	case <-sigs:
		log.Info("got second interrupt; forcing shutdown")
	case <-done:
	}
}

func pollDB(harness *harness.Harness) {
	log, _ := logger.FromCtx(context.Background())
	for {
		time.Sleep(2 * time.Second)
		topics := harness.ListTopics()
		for t, size := range topics {
			log.Info("topic", zap.String("topic", t), zap.Int("size", size))
		}
	}
}
