//go:build slim
// +build slim

package pubsub

import (
	"sync"

	"github.com/anycable/anycable-go/common"
	"github.com/redis/rueidis"
)

type RedisSubscriber struct {
	client        rueidis.Client
	clientOptions *rueidis.ClientOption
	clientMu      sync.RWMutex
}

var _ Subscriber = (*RedisSubscriber)(nil)

func NewRedisSubscriber(args ...interface{}) (*RedisSubscriber, error) {
	return &RedisSubscriber{}, nil
}

func (s *RedisSubscriber) Start(done chan (error)) error {
	return nil
}

func (s *RedisSubscriber) Shutdown() error {
	return nil
}

func (s *RedisSubscriber) IsMultiNode() bool {
	return true
}

func (s *RedisSubscriber) Subscribe(stream string) {
}

func (s *RedisSubscriber) Unsubscribe(stream string) {
}

func (s *RedisSubscriber) Broadcast(msg *common.StreamMessage) {
}

func (s *RedisSubscriber) BroadcastCommand(cmd *common.RemoteCommandMessage) {
}

func (s *RedisSubscriber) Publish(stream string, msg interface{}) {
}

func (s *RedisSubscriber) initClient() error {
	s.clientMu.Lock()
	defer s.clientMu.Unlock()

	if s.client != nil {
		return nil
	}

	c, err := rueidis.NewClient(*s.clientOptions)

	if err != nil {
		return err
	}

	s.client = c

	return nil
}
