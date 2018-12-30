// +build darwin,mrb linux,mrb

package metrics

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/anycable/anycable-go/mrb"
)

func TestPrintGC(t *testing.T) {
	engine := mrb.NewEngine()

	engine.LoadString(
		`
		module MetricsFormatter
			def self.new_hash; {}; end

			def self.call(data)
				filtered_data = data.select do |k, v|
					!k.start_with?("_")
				end

				filtered_data.to_json
			end
		end
		`,
	)

	mod := engine.VM.Module("MetricsFormatter")

	modValue := mod.MrbValue(engine.VM)

	printer := &RubyPrinter{mrbModule: modValue, engine: engine}

	engine.VM.FullGC()

	origObjects := engine.VM.LiveObjectCount()

	snapshot := map[string]int64{"a": 123, "_b": 432, "_f": 321}

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	printer.Print(snapshot)

	assert.Contains(t, buf.String(), "{\"a\":123}")

	newObjects := engine.VM.LiveObjectCount()

	if origObjects != newObjects {
		t.Fatalf("Object count was not what was expected after action call: %d %d", origObjects, newObjects)
	}
}
