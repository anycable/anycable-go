// +build darwin,mrb linux,mrb

package metrics

import (
	"sync"

	"github.com/anycable/anycable-go/mrb"
	"github.com/apex/log"
	"github.com/mitchellh/go-mruby"
)

// RubyPrinter contains refs to mruby vm and code
type RubyPrinter struct {
	mrbModule *mruby.MrbValue
	engine    *mrb.Engine
	mu        sync.Mutex
}

// NewCustomPrinter generates log formatter from the provided (as path)
// Ruby script
func NewCustomPrinter(path string) (*RubyPrinter, error) {
	engine := mrb.NewEngine()

	if err := engine.LoadFile(path); err != nil {
		return nil, err
	}

	mod := engine.VM.Module("MetricsFormatter")

	modValue := mod.MrbValue(engine.VM)

	return &RubyPrinter{mrbModule: modValue, engine: engine}, nil
}

// Print calls Ruby script to format the output and prints it to the log
func (printer *RubyPrinter) Print(snapshot map[string]int64) {
	printer.mu.Lock()
	defer printer.mu.Unlock()

	arenaIdx := printer.engine.VM.ArenaSave()

	// It turned out that LoadString("{}") generates non-GC trash,
	// while calling Hash.new doesn't
	rhash, _ := printer.engine.VM.Class("Hash", nil).MrbValue(printer.engine.VM).Call("new")

	hash := rhash.Hash()

	for k, v := range snapshot {
		hash.Set(mruby.String(k), mruby.Int(v))
	}

	result, err := printer.mrbModule.Call("call", rhash)

	if err != nil {
		log.WithField("context", "metrics").Error(err.Error())
		return
	}

	log.Info(result.String())

	printer.engine.VM.ArenaRestore(arenaIdx)
	printer.engine.VM.IncrementalGC()
}
