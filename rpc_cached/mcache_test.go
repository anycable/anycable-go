// +build darwin,mrb linux,mrb

package rpc_cached

import (
	"testing"

	"github.com/anycable/anycable-go/mrb"

	"github.com/stretchr/testify/assert"
)

func init() {
	InitCache()
}

func TestMAction(t *testing.T) {
	maction, err := NewMAction(
		mrb.DefaultEngine(),
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
