package rpc_cached

import (
	"testing"

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
	controller := &Controller{cache: cache}

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

func BenchmarkCachedActionPerform(b *testing.B) {
	controller := &Controller{cache: cache}

	for i := 0; i < b.N; i++ {
		controller.Perform(
			"test",
			"",
			"{\"channel\":\"BenchmarkChannel\"}",
			"{\"action\":\"echo\",\"text\":\"hello\"}",
		)
	}
}
