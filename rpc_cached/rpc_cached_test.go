package rpc_cached

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/anycable/anycable-go/config"
	"github.com/anycable/anycable-go/metrics"
	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
)

func init() {
	err := cache.Put(
		"BenchmarkChannel",
		"echo",
		`
		def echo(data)
			transmit response: data
		end
		`,
	)

	if err != nil {
		log.Fatalf("Failed to compile channel method: %s", err)
	}
}

func TestCachedPerform(t *testing.T) {
	m := metrics.NewMetrics(nil, 10)
	c := &config.Config{}
	controller := NewController(c, m)
	controller.cache = cache

	res, err := controller.Perform(
		"test",
		"",
		"{\"channel\":\"BenchmarkChannel\"}",
		"{\"action\":\"echo\",\"text\":\"hello\"}",
	)

	assert.Nil(t, err)

	identifier := "{\\\"channel\\\":\\\"BenchmarkChannel\\\"}"

	assert.Equal(
		t,
		[]string{"{\"identifier\":\"" + identifier + "\",\"message\":{\"response\":{\"action\":\"echo\",\"text\":\"hello\"}}}"},
		res.Transmissions,
	)
}

func TestConcurrentCachedPerform(t *testing.T) {
	m := metrics.NewMetrics(nil, 10)
	c := &config.Config{}
	controller := NewController(c, m)
	controller.cache = cache

	var wg sync.WaitGroup

	errorsCount := int64(0)

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			_, err := controller.Perform(
				"test",
				"",
				"{\"channel\":\"BenchmarkChannel\"}",
				"{\"action\":\"echo\",\"text\":\"hello\"}",
			)

			if err != nil {
				atomic.AddInt64(&errorsCount, 1)
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, int64(0), errorsCount)
}

func BenchmarkCachedActionEcho(b *testing.B) {
	m := metrics.NewMetrics(nil, 10)
	c := &config.Config{}
	controller := NewController(c, m)
	controller.cache = cache

	for i := 0; i < b.N; i++ {
		controller.Perform(
			"test",
			"",
			"{\"channel\":\"BenchmarkChannel\"}",
			"{\"action\":\"echo\",\"text\":\"hello\"}",
		)
	}
}
