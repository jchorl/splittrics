package event

import (
	"time"

	"github.com/jchorl/splittrics/pkg/metrics_types"
)

type Event struct {
	Name  string
	Time  time.Time
	Value interface{}
	Type  metrics_types.Type
	Tags  map[string]string
}
