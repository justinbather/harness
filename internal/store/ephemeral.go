package store

import "sync"

type Message struct {
	Data      string
	Partition string
	Offset    string
}

type Topic struct {
	Name         string
	Partitions   int
	MessageCount int
}

type store struct {
	mu sync.Mutex

	messageStorage map[string][]Message
	// topics and the message counts
	topicStorage map[string]*Topic
}

type Store interface {
	ListTopics() map[string]*Topic

	Insert(topic string, msg Message)

	ListMessages(topic string) []Message
}

func NewEphemeralStore(topics ...*Topic) Store {
	messageStore := make(map[string][]Message)
	topicStore := make(map[string]*Topic)
	for _, t := range topics {
		// pre allocate 100 for now
		messageStore[t.Name] = make([]Message, 0, 100)
		copy := *t
		topicStore[t.Name] = &copy
	}

	return &store{messageStorage: messageStore, topicStorage: topicStore}
}

func (s *store) ListTopics() map[string]*Topic {
	return s.topicStorage
}

func (s *store) Insert(topic string, msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messageStorage[topic] = append(s.messageStorage[topic], msg)
	s.topicStorage[topic].MessageCount += 1
}

func (s *store) ListMessages(topic string) []Message {
	return s.messageStorage[topic]
}
