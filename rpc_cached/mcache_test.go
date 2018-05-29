// +build darwin linux

package rpc_cached

import (
	"log"
	"testing"

	"github.com/anycable/anycable-go/mrb"

	"github.com/stretchr/testify/assert"
)

var (
	cache *MCache
)

func init() {
	var err error
	cache, err = NewMCache(mrb.DefaultEngine())

	if err != nil {
		log.Fatalf("Failed to initialize mruby cache: %s", err)
	}
}

func TestMAction(t *testing.T) {
	maction, err := NewMAction(
		cache,
		"BenchmarkChannel",
		`
		def echo(data)
			transmit response: data
		end
		`,
	)

	assert.Nil(t, err)

	res, err := maction.Perform("{\"action\":\"echo\",\"text\":\"hello\"}")

	assert.Nil(t, err)

	identifier := "{\\\"channel\\\":\\\"BenchmarkChannel\\\"}"

	assert.Empty(t, res.Streams)
	assert.False(t, res.Disconnect)
	assert.False(t, res.StopAllStreams)
	assert.Equal(t, []string{"{\"identifier\":\"" + identifier + "\",\"message\":{\"response\":{\"action\":\"echo\",\"text\":\"hello\"}}}"}, res.Transmissions)
}
