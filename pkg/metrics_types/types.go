package metrics_types

type Type string

const (
	Gauge  Type = "gauge"
	Count  Type = "count"
	Timing Type = "timing"
)
