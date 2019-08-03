package providers

import (
	"github.com/jchorl/splittrics/pkg/event"
)

type Provider interface {
	Send([]event.Event) error
}
