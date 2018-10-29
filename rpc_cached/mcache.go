// +build darwin,mrb linux,mrb

package rpc_cached

import (
	"strings"
	"sync"

	"github.com/anycable/anycable-go/mrb"
	"github.com/anycable/anycable-go/node"
	mruby "github.com/mitchellh/go-mruby"
)

// MCallResult is an interface to decode mruby call
type MCallResult struct {
	Streams        []string `mruby:"streams"`
	StopAllStreams bool     `mruby:"stop_all_streams"`
	Transmissions  []string `mruby:"transmissions"`
	Broadcasts     []string `mruby:"broadcasts"`
}

// MCache is a cache of mruby compiled channels methods
type MCache struct {
	engine *mrb.Engine
	store  map[string]map[string]*MAction
}

// NewMCache builds a new cache struct for mruby engine
func NewMCache(mengine *mrb.Engine) (*MCache, error) {
	// Build base channel class
	err := mengine.LoadString(
		`
	module AnyCable
		class Channel
			class Result
				attr_reader :streams, :transmissions, :broadcasts
				attr_accessor :stop_all_streams

				def initialize
					@stop_all_streams = false
					@transmissions = []
					@streams = []
					@broadcasts = []
				end

				def to_gomruby
					{
						stop_all_streams: stop_all_streams,
						streams: streams,
						transmissions: transmissions,
						broadcasts: broadcasts
					}
				end
			end

			class << self
				attr_reader :identifier

				def identify(channel)
					@identifier = { channel: channel }.to_json
				end

				def perform(json_data)
					data = JSON.parse(json_data)
					new.tap do |channel|
						channel.send(data.fetch('action'), data)
					end.result
				end

				def subscribe
					new.tap do |channel|
						channel.result.transmissions << {
							identifier: identifier,
							type: "confirm_subscription"
						}.to_json
						channel.subscribed
					end.result
				end

				def unsubscribe
					new.tap do |channel|
						channel.unsubscribed
						channel.stop_all_streams
					end.result
				end
			end

			attr_reader :result

			def initialize
				@result = Result.new
			end

			def transmit(data)
				result.transmissions << {
					identifier: self.class.identifier,
					message: data
				}.to_json
			end

			def stream_from(stream)
			  result.streams << stream
			end

			def stop_all_streams
			  result.stop_all_streams = true
			end

			def __broadcast__(stream, data)
			  result.broadcasts << {
					stream: stream,
					data: data.to_json
				}.to_json
			end
		end
	end
	`,
	)

	if err != nil {
		return nil, err
	}

	return &MCache{
		engine: mengine,
		store:  make(map[string]map[string]*MAction),
	}, nil
}

// Put compiles method and put it in the cache
func (c *MCache) Put(channel string, action string, source string) (err error) {
	// TODO: check for existence, add lock
	var maction *MAction

	maction, err = NewMAction(c, channel, source)

	if err != nil {
		return
	}

	if _, ok := c.store[channel]; !ok {
		c.store[channel] = make(map[string]*MAction)
	}

	c.store[channel][action] = maction
	return
}

// Get returns cached action for the channel if any
func (c *MCache) Get(channel string, action string) (maction *MAction) {
	if _, ok := c.store[channel]; !ok {
		return
	}

	maction, _ = c.store[channel][action]
	return
}

// MAction is a signle cached channel method
type MAction struct {
	compiled *mruby.MrbValue
	mu       sync.Mutex
	cache    *MCache
}

// NewMAction compiles a channel method within mruby VM
func NewMAction(cache *MCache, channel string, source string) (*MAction, error) {
	var buf strings.Builder

	channelClass := "CachedChannel_" + channel

	buf.WriteString(
		"class " + channelClass + " < AnyCable::Channel\n",
	)
	buf.WriteString("identify \"" + channel + "\"\n")
	buf.WriteString(source + "\n")
	buf.WriteString("end\n")

	engine := cache.engine

	err := engine.LoadString(buf.String())

	if err != nil {
		return nil, err
	}

	mchannel := engine.VM.Class(channelClass, nil)

	mchannelValue := mchannel.MrbValue(engine.VM)

	cache.engine.VM.IncrementalGC()

	return &MAction{compiled: mchannelValue, cache: cache}, nil
}

// Perform executes action within mruby
func (m *MAction) Perform(data string) (*node.CommandResult, error) {
	return m.exec(func() (v *mruby.MrbValue, err error) {
		return m.compiled.Call("perform", mruby.String(data))
	})
}

// Subscribe invokes "subscribed" action
func (m *MAction) Subscribe() (*node.CommandResult, error) {
	return m.exec(func() (v *mruby.MrbValue, err error) {
		return m.compiled.Call("subscribe")
	})
}

// Unubscribe invokes "unsubscribed" action
func (m *MAction) Unsubscribe() (*node.CommandResult, error) {
	return m.exec(func() (v *mruby.MrbValue, err error) {
		return m.compiled.Call("unsubscribe")
	})
}

func (m *MAction) exec(f func() (v *mruby.MrbValue, err error)) (res *node.CommandResult, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	arenaIdx := m.cache.engine.VM.ArenaSave()

	result, err := f()

	if err != nil {
		return nil, err
	}

	decoded := MCallResult{}

	err = mruby.Decode(&decoded, result)

	m.cache.engine.VM.ArenaRestore(arenaIdx)

	m.cache.engine.VM.IncrementalGC()

	if err != nil {
		return nil, err
	}

	res = &node.CommandResult{
		Transmissions:  decoded.Transmissions,
		StopAllStreams: decoded.StopAllStreams,
		Streams:        decoded.Streams,
		Broadcasts:     decoded.Broadcasts,
	}

	return res, nil
}
