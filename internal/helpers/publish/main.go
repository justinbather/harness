package main

import (
	"context"
	"fmt"
	"os"

	"github.com/justinbather/harness/internal/logger"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	log, ctx := logger.FromCtx(ctx)

	topics := []string{"user-events", "order-events", "payment-events", "notification-events", "audit-logs"}
	brokers := []string{"localhost:9092"}

	client, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
	if err != nil {
		log.Error("creating client", logger.Err(err))
		os.Exit(1)
	}

	defer client.Close()

	log.Info("publishing...")
	for i := range 10 {
		for _, topic := range topics {
			r := &kgo.Record{
				Value: []byte(fmt.Sprintf("message %d", i)),
				Topic: topic,
			}
			client.Produce(ctx, r, func(r *kgo.Record, err error) {
				if err != nil {
					log.Error("producing", zap.Error(err))
				} else {
					log.Info("produced message", logger.F("topic", topic))

				}
			})
		}
	}

	log.Info("flushing...")

	if err := client.Flush(ctx); err != nil {
		log.Error("flushing", zap.Error(err))
	}

	log.Info("done publishing")
}
