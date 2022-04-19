package broker

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/anycable/anycable-go/common"
	nanoid "github.com/matoous/go-nanoid"
)

type entry struct {
	timestamp int64
	offset    uint64
	data      string
}

type memstream struct {
	offset uint64
	// The lowest available offset in the stream
	low   uint64
	data  []*entry
	ttl   int64
	limit int

	mu sync.RWMutex
}

func (ms *memstream) add(data string) uint64 {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ts := time.Now().Unix()

	ms.offset++

	entry := &entry{
		offset:    ms.offset,
		timestamp: ts,
		data:      data,
	}

	ms.data = append(ms.data, entry)

	if len(ms.data) > ms.limit {
		ms.data = ms.data[1:]
		ms.low = ms.data[0].offset
	}

	return ms.offset
}

func (ms *memstream) expire() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	cutIndex := 0

	now := time.Now().Unix()
	deadline := now - ms.ttl

	for _, entry := range ms.data {
		if entry.timestamp < deadline {
			cutIndex++
			continue
		}

		break
	}

	if cutIndex < 0 {
		return
	}

	ms.data = ms.data[cutIndex:]
	ms.low = ms.data[0].offset
}

func (ms *memstream) filterByOffset(offset uint64, callback func(e *entry)) error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.low > offset {
		return fmt.Errorf("Requested offset couldn't be found: %d, lowest: %d", offset, ms.low)
	}

	start := (offset - ms.low) + 1

	for _, v := range ms.data[start:] {
		callback(v)
	}

	return nil
}

func (ms *memstream) filterByTime(since int64, callback func(e *entry)) error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	for _, v := range ms.data {
		if v.timestamp >= since {
			callback(v)
		}
	}

	return nil
}

type Memory struct {
	broadcaster Broadcaster
	config      *Config
	tracker     *StreamsTracker
	streams     map[string]*memstream
	epoch       string

	mu sync.RWMutex
}

func NewMemoryBroker(node Broadcaster, config *Config) *Memory {
	epoch, _ := nanoid.Nanoid(4)

	return &Memory{
		broadcaster: node,
		config:      config,
		tracker:     NewStreamsTracker(),
		streams:     make(map[string]*memstream),
		epoch:       epoch,
	}
}

func (b *Memory) Start() error {
	go b.expireLoop()

	return nil
}

func (b *Memory) Shutdown() error {
	return nil
}

func (b *Memory) HandleBroadcast(msg *common.StreamMessage) {
	offset := b.add(msg.Stream, msg.Data)

	msg.Epoch = b.epoch
	msg.Offset = offset

	if b.tracker.Has(msg.Stream) {
		b.broadcaster.Broadcast(msg)
	}
}

// Registring streams (for granular pub/sub)

func (b *Memory) Subscribe(stream string) string {
	return stream
}

func (b *Memory) Unsubscribe(stream string) string {
	return stream
}

func (b *Memory) HistoryFrom(name string, epoch string, offset uint64) ([]common.StreamMessage, error) {
	if b.epoch != epoch {
		return nil, fmt.Errorf("Unknown epoch: %s, current: %s", epoch, b.epoch)
	}

	stream := b.get(name)

	if stream == nil {
		return nil, errors.New("Stream not found")
	}

	history := []common.StreamMessage{}

	err := stream.filterByOffset(offset, func(entry *entry) {
		history = append(history, common.StreamMessage{
			Stream: name,
			Data:   entry.data,
			Offset: entry.offset,
			Epoch:  b.epoch,
		})
	})

	if err != nil {
		return nil, err
	}

	return history, nil
}

func (b *Memory) HistorySince(name string, ts int64) ([]common.StreamMessage, error) {
	stream := b.get(name)

	if stream == nil {
		return nil, nil
	}

	history := []common.StreamMessage{}

	err := stream.filterByTime(ts, func(entry *entry) {
		history = append(history, common.StreamMessage{
			Stream: name,
			Data:   entry.data,
			Offset: entry.offset,
			Epoch:  b.epoch,
		})
	})

	if err != nil {
		return nil, err
	}

	return history, nil
}

func (b *Memory) CommitSession(sid string, session Cacheable) error {
	return nil
}

func (b *Memory) RestoreSession(from string, to string) (string, error) {
	return "", nil
}

func (b *Memory) add(name string, data string) uint64 {
	b.mu.Lock()

	if _, ok := b.streams[name]; !ok {
		b.streams[name] = &memstream{
			data:  []*entry{},
			ttl:   b.config.HistoryTTL,
			limit: b.config.HistoryLimit,
		}
	}

	stream := b.streams[name]

	b.mu.Unlock()

	return stream.add(data)
}

func (b *Memory) get(name string) *memstream {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.streams[name]
}

func (b *Memory) expireLoop() {
	for {
		time.Sleep(time.Second)
		b.expire()
	}
}

func (b *Memory) expire() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, stream := range b.streams {
		stream.expire()
	}
}
