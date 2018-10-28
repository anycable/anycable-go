// +build darwin,mrb linux,mrb

package rpc_cached

import (
	"encoding/json"

	"github.com/anycable/anycable-go/mrb"

	"github.com/anycable/anycable-go/config"
	"github.com/anycable/anycable-go/metrics"
	"github.com/anycable/anycable-go/node"
	"github.com/anycable/anycable-go/rpc"
	"github.com/apex/log"
)

const (
	metricsCacheHit = "rpc_cache_hit"
)

// Identifier is used to extract channel id from message JSON
type Identifier struct {
	Channel string `json:"channel"`
}

// ActionMessage is used to extract action name from message JSON
type ActionMessage struct {
	Action string `json:"action"`
}

// Controller implements node.Controller interface for gRPC
type Controller struct {
	rpc     *rpc.Controller
	cache   *MCache
	metrics *metrics.Metrics
	log     *log.Entry
}

// NewController builds new Controller from config
func NewController(config *config.Config, metrics *metrics.Metrics) *Controller {
	metrics.RegisterCounter(metricsCacheHit, "The total number of RPC cache hits")

	rpc := rpc.NewController(config, metrics)
	return &Controller{log: log.WithField("context", "rpc"), metrics: metrics, rpc: rpc}
}

// Start initializes RPC connection pool
func (c *Controller) Start() (err error) {
	err = c.rpc.Start()
	c.cache, err = NewMCache(mrb.DefaultEngine())
	return
}

// Shutdown closes connections
func (c *Controller) Shutdown() error {
	return c.rpc.Shutdown()
}

// Authenticate performs Connect RPC call
func (c *Controller) Authenticate(path string, headers *map[string]string) (string, []string, error) {
	return c.rpc.Authenticate(path, headers)
}

// Subscribe performs Command RPC call with "subscribe" command
func (c *Controller) Subscribe(sid string, id string, channel string) (*node.CommandResult, error) {
	return c.rpc.Subscribe(sid, id, channel)
}

// Unsubscribe performs Command RPC call with "unsubscribe" command
func (c *Controller) Unsubscribe(sid string, id string, channel string) (*node.CommandResult, error) {
	return c.rpc.Unsubscribe(sid, id, channel)
}

// Perform performs Command RPC call with "perform" command
func (c *Controller) Perform(sid string, id string, channel string, data string) (res *node.CommandResult, err error) {
	identifier := Identifier{}

	err = json.Unmarshal([]byte(channel), &identifier)

	if err != nil {
		return
	}

	msg := ActionMessage{}
	err = json.Unmarshal([]byte(data), &msg)

	if err != nil {
		return
	}

	if maction := c.cache.Get(identifier.Channel, msg.Action); maction != nil {
		c.metrics.Counter(metricsCacheHit).Inc()
		log.WithFields(
			log.Fields{
				"sid":     sid,
				"context": "rpc",
				"channel": identifier.Channel,
				"action":  msg.Action,
			}).Debugf("cache hit")
		return maction.Perform(data)
	}

	return c.rpc.Perform(sid, id, channel, data)
}

// Disconnect performs disconnect RPC call
func (c *Controller) Disconnect(sid string, id string, subscriptions []string, path string, headers *map[string]string) error {
	return c.rpc.Disconnect(sid, id, subscriptions, path, headers)
}
