package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const (
	prometheusNamespace = "anycable_go"
)

// Prometheus returns metrics info in Prometheus format
func (m *Metrics) Prometheus() string {
	var buf strings.Builder

	for _, counter := range m.Counters() {
		name := prometheusNamespace + `_` + counter.Name()

		buf.WriteString(
			"\n# HELP " + name + " " + counter.Desc() + "\n",
		)
		buf.WriteString("# TYPE " + name + " counter\n")
		buf.WriteString(name + " " + strconv.FormatInt(counter.Value(), 10) + "\n")
	}

	for _, gauge := range m.Gauges() {
		name := prometheusNamespace + `_` + gauge.Name()

		buf.WriteString(
			"\n# HELP " + name + " " + gauge.Desc() + "\n",
		)
		buf.WriteString("# TYPE " + name + " gauge\n")
		buf.WriteString(name + " " + strconv.FormatInt(gauge.Value(), 10) + "\n")
	}

	return buf.String()
}

// PrometheusHandler is provide metrics to the world
func (m *Metrics) PrometheusHandler(w http.ResponseWriter, r *http.Request) {
	metricsData := m.Prometheus()

	fmt.Fprint(w, metricsData)
}
