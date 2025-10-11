package store

import "sync"

type Message struct {
	Metadata map[string]any
	Data     string
}

type store struct {
	mu sync.Mutex

	s map[string][]Message
	// topics and the message counts
	metadata map[string]int
}

type Store interface {
	ListTopics() map[string]int
	Insert(topic string, msg Message)
}

func NewEphemeralStore(tables ...string) Store {
	s := make(map[string][]Message)
	md := make(map[string]int)
	for _, t := range tables {
		// pre allocate 100 for now
		s[t] = make([]Message, 0, 100)
		md[t] = 0
	}

	return &store{s: s, metadata: md}
}

func (s *store) ListTopics() map[string]int {
	return s.metadata
}

func (s *store) Insert(topic string, msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.s[topic] = append(s.s[topic], msg)
	s.metadata[topic] += 1
}
