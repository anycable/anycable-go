package rpc

import (
	"os"
	"testing"

	"github.com/anycable/anycable-go/config"
	"github.com/anycable/anycable-go/metrics"
	"github.com/stretchr/testify/assert"
)

var (
	rpcHost = os.Getenv("TEST_RPC")
)

func TestActionPerform(t *testing.T) {
	if rpcHost == "" {
		t.Skip("RPC server is not running")
	}

	m := metrics.NewMetrics(nil, 10)
	c := &config.Config{RPCHost: "0.0.0.0:50051"}
	controller := NewController(c, m)
	controller.Start()

	_, err := controller.Perform(
		"test",
		"{}",
		"{\"channel\":\"BenchmarkChannel\"}",
		"{\"action\":\"echo\",\"text\":\"hello\"}",
	)

	assert.Nil(t, err)
}

func BenchmarkActionPerform(b *testing.B) {
	if rpcHost == "" {
		b.Skip("RPC server is not running")
	}

	m := metrics.NewMetrics(nil, 10)
	c := &config.Config{RPCHost: "0.0.0.0:50051"}
	controller := NewController(c, m)

	controller.Start()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		controller.Perform(
			"test",
			"{}",
			"{\"channel\":\"BenchmarkChannel\"}",
			"{\"action\":\"echo\",\"text\":\"hello\"}",
		)
	}
}
