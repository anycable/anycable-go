// +build !darwin,!linux !mrb

package rpc_cached

import (
	"log"

	"github.com/anycable/anycable-go/config"
	"github.com/anycable/anycable-go/metrics"
	"github.com/anycable/anycable-go/rpc"
)

// NewController builds new Controller from config
func NewController(config *config.Config, metrics *metrics.Metrics) *rpc.Controller {
	log.Fatal("Cached (mRuby) RPC controller is not supported!")
	return rpc.NewController(config, metrics)
}
