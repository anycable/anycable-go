//go:build slim
// +build slim

package telemetry

type Tracker struct{}

func NewTracker(args ...interface{}) *Tracker {
	return &Tracker{}
}

func (t *Tracker) Announce() {
}

func (t *Tracker) Collect() {
}

func (t *Tracker) Shutdown() error {
	return nil
}

func (t *Tracker) Send(event string, props map[string]interface{}) {
}
