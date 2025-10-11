package harness

import (
	"context"
	"errors"

	"github.com/justinbather/harness/internal/logger"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type Harness struct {
	kc *kgo.Client
}

func New(brokers, topics []string) (*Harness, error) {
	kc, err := kgo.NewClient(kgo.SeedBrokers(brokers...), kgo.ConsumeTopics(topics...), kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()), kgo.ConsumerGroup("harness"))
	if err != nil {
		return nil, err
	}
	h := &Harness{kc: kc}

	return h, nil
}

func (h *Harness) Start(ctx context.Context) error {
	if h.kc == nil {
		return errors.New("harness is in a bad state: kafka client")
	}

	go h.consume(ctx)

	return nil
}

func (h *Harness) consume(ctx context.Context) {

	log, _ := logger.FromCtx(ctx)
	log.Info("consuming")
	for {

		fetches := h.kc.PollFetches(ctx)
		fetches.EachRecord(func(r *kgo.Record) {
			log.Info("consumed", zap.String("topic", r.Topic), zap.Any("data", r.Value))
		})
		if err := h.kc.CommitUncommittedOffsets(ctx); err != nil {
			log.Error("committing records", zap.Error(err))
			continue
		}
	}
}

func (h *Harness) Shutdown(ctx context.Context) {
	if h.kc != nil {
		h.kc.CommitUncommittedOffsets(ctx)
		h.kc.Close()
	}
}
