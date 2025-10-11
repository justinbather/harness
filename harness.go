package harness

import (
	"context"
	"errors"
	"fmt"

	"github.com/justinbather/harness/internal/logger"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
	"go.uber.org/zap"
)

type Harness struct {
	kc     *kgo.Client
	topics []kmsg.MetadataRequestTopic
}

func New(brokers []string) (*Harness, error) {
	kc, err := kgo.NewClient(kgo.SeedBrokers(brokers...), kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()))
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	topics, err := getTopics(ctx, kc)
	if err != nil {
		return nil, err
	}

	registerTopics(kc, topics)

	h := &Harness{kc: kc}
	return h, nil
}

func registerTopics(kc *kgo.Client, topics []kmsg.MetadataResponseTopic) {
	var strTopics []string

	for _, t := range topics {
		if t.Topic != nil {
			// kafka internal topics
			if *t.Topic != "__consumer_offsets" && *t.Topic != "__transaction_state" {
				strTopics = append(strTopics, *t.Topic)
			}
		}
	}

	kc.AddConsumeTopics(strTopics...)
}

func getTopics(ctx context.Context, kc *kgo.Client) ([]kmsg.MetadataResponseTopic, error) {
	req := kmsg.NewMetadataRequest()
	req.Topics = nil // fetches all topics
	resp, err := req.RequestWith(ctx, kc)
	if err != nil {
		return nil, fmt.Errorf("getting topics: %w", err)
	}

	topics := resp.Topics

	if len(topics) == 0 {
		return nil, errors.New("no topics found")
	}

	return topics, nil
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
	}
}

func (h *Harness) Shutdown(ctx context.Context) {
	if h.kc != nil {
		h.kc.Close()
	}
}
