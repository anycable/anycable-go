package broker

import (
	"errors"
	"sync"

	"github.com/anycable/anycable-go/common"
)

// Broadcaster is responsible for fan-out message to the stream clients
type Broadcaster interface {
	Broadcast(msg *common.StreamMessage)
}

type StreamsTracker struct {
	store map[string]uint64

	mu sync.Mutex
}

func (s *StreamsTracker) Add(name string) (isNew bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.store[name]; !ok {
		s.store[name] = 1
		return true
	}

	s.store[name]++
	return false
}

func (s *StreamsTracker) Remove(name string) (isLast bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.store[name]; !ok {
		return false
	}

	if s.store[name] == 1 {
		delete(s.store, name)
		return true
	}

	return false
}

// Broker is responsible for:
// - Managing streams history.
// - Keeping client states for recovery.
// - Distributing broadcasts across nodes.
type Broker interface {
	Start() error
	Shutdown() error

	HandleBroadcast(msg *common.StreamMessage)

	// Registers the stream and returns its (short) unique identifier
	Subscribe(stream string) string
	// (Maybe) unregisters the stream and return its unique identifier
	Unsubscribe(stream string) string
	// Retrieves stream messages from history from the specified offset within the specified epoch
	HistoryFrom(stream string, epoch string, offset uint64) ([]common.StreamMessage, error)
	// Retrieves stream messages from history from the specified timestamp
	HistorySince(stream string, ts int) ([]common.StreamMessage, error)
}

// LegacyBroker preserves the v1 behaviour while implementing the Broker APIs.
// Thus, we can use it without breaking the older behaviour
type LegacyBroker struct {
	broadcaster Broadcaster
}

func NewLegacyBroker(broadcaster Broadcaster) *LegacyBroker {
	return &LegacyBroker{broadcaster: broadcaster}
}

func (LegacyBroker) Start() error {
	return nil
}

func (LegacyBroker) Shutdown() error {
	return nil
}

func (b *LegacyBroker) HandleBroadcast(msg *common.StreamMessage) {
	b.broadcaster.Broadcast(msg)
}

// Registring streams (for granular pub/sub)

func (LegacyBroker) Subscribe(stream string) string {
	return stream
}

func (LegacyBroker) Unsubscribe(stream string) string {
	return stream
}

func (LegacyBroker) HistoryFrom(stream string, epoch string, offset uint64) ([]common.StreamMessage, error) {
	return nil, errors.New("History not supported")
}

func (LegacyBroker) HistorySince(stream string, ts int) ([]common.StreamMessage, error) {
	return nil, errors.New("History not supported")
}
