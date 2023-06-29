//go:build slim
// +build slim

package broadcast

import (
	_ "github.com/redis/rueidis"
)

type RedisBroadcaster struct {
}

var _ Broadcaster = (*RedisBroadcaster)(nil)

func NewRedisBroadcaster(args ...interface{}) *RedisBroadcaster {
	return &RedisBroadcaster{}
}

func (s *RedisBroadcaster) IsFanout() bool {
	return false
}

func (s *RedisBroadcaster) Start(done chan error) error {
	return nil
}

func (s *RedisBroadcaster) Shutdown() error {
	return nil
}
