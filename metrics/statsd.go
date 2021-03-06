package metrics

import "github.com/smira/go-statsd"

const (
	statsdPrefix = "anycable_go."
	statsdMaxPacketSize = 1400
)

func (m *Metrics) InitStatsdClient(host string) {
	m.statsdClient = statsd.NewClient(
		host,
		statsd.MaxPacketSize(statsdMaxPacketSize),
		statsd.MetricPrefix(statsdPrefix))
}

func (m *Metrics) StatsdSend() {
	m.EachCounter(func(counter *Counter) {
		m.statsdClient.Incr(counter.Name(), counter.IntervalValue())
	})

	m.EachGauge(func(gauge *Gauge) {
		m.statsdClient.Gauge(gauge.Name(), gauge.Value())
	})
}
