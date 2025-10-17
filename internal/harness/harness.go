package harness

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/justinbather/harness/internal/store"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
)

type Harness struct {
	Brokers []string
	kc      *kgo.Client
	// should create a seperate querier iface to prevent external mutation
	// or just add methods to Harness
	store.Store
	topics []kmsg.MetadataRequestTopic
}

func New(brokers ...string) (*Harness, error) {
	ctx := context.Background()

	kc, err := kgo.NewClient(kgo.SeedBrokers(brokers...), kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()))
	if err != nil {
		return nil, err
	}

	topics, err := fetchTopics(ctx, kc)
	if err != nil {
		return nil, err
	}

	var topicNames []string
	for _, topic := range topics {
		topicNames = append(topicNames, topic.Name)
	}

	kc.AddConsumeTopics(topicNames...)

	store := store.NewEphemeralStore(topics...)

	h := &Harness{kc: kc, Brokers: brokers, Store: store}
	return h, nil
}

func (h *Harness) Start(ctx context.Context) error {
	if h.kc == nil {
		return errors.New("harness is in a bad state: kafka client")
	}

	go h.consume(ctx)

	return nil
}

func (h *Harness) Shutdown(ctx context.Context) {
	if h.kc != nil {
		h.kc.Close()
	}
}

func fetchTopics(ctx context.Context, kc *kgo.Client) ([]*store.Topic, error) {
	var topics []*store.Topic

	req := kmsg.NewMetadataRequest()
	req.Topics = nil // fetches all topics

	resp, err := req.RequestWith(ctx, kc)
	if err != nil {
		return nil, fmt.Errorf("getting topics: %w", err)
	}

	for _, topic := range resp.Topics {
		if topic.Topic == nil {
			continue
		}

		if *topic.Topic == "__consumer_offsets" || *topic.Topic == "__transaction_state" {
			continue
		}

		t := &store.Topic{
			Name:         *topic.Topic,
			Partitions:   len(topic.Partitions),
			MessageCount: 0,
		}

		topics = append(topics, t)
	}

	if len(topics) == 0 {
		return nil, errors.New("no topics found")
	}

	return topics, nil
}

// should spawn a goroutine per topic
// this needs to save into an ephemeral storage in json?
func (h *Harness) consume(ctx context.Context) {
	for {
		fetches := h.kc.PollFetches(ctx)
		fetches.EachRecord(func(r *kgo.Record) {
			msg := store.Message{
				Data:      r.Value,
				Partition: strconv.Itoa(int(r.Partition)),
				Offset:    strconv.Itoa(int(r.Offset)),
			}
			h.Store.Insert(r.Topic, msg)
		})
	}
}
