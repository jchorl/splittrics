package splittrics

import (
	"errors"
	"sync"
	"time"

	"github.com/jchorl/splittrics/pkg/event"
	"github.com/jchorl/splittrics/pkg/metrics_types"
	"github.com/jchorl/splittrics/pkg/providers"
)

type Client interface {
	Gauge(name string, value float64, tags map[string]string) error
	Count(name string, value int64, tags map[string]string) error
	Decr(name string, tags map[string]string) error
	Incr(name string, tags map[string]string) error
	Timing(name string, value time.Duration, tags map[string]string) error

	Close() error
}

var _ Client = (*client)(nil)

type client struct {
	Provider providers.Provider

	sync.Mutex
	maxBufferSize int
	buffer        []event.Event
	closed        bool
}

func New(provider providers.Provider) Client {
	maxBufferSize := 50

	return &client{
		Provider:      provider,
		maxBufferSize: maxBufferSize,
	}
}

// Gauge measures the value of a metric at a particular time.
func (c *client) Gauge(name string, value float64, tags map[string]string) error {
	return c.Send(event.Event{
		Name:  name,
		Time:  time.Now(),
		Value: value,
		Type:  metrics_types.Gauge,
		Tags:  tags,
	})
}

// Count tracks how many times something happened per second.
func (c *client) Count(name string, value int64, tags map[string]string) error {
	return c.Send(event.Event{
		Name:  name,
		Time:  time.Now(),
		Value: value,
		Type:  metrics_types.Count,
		Tags:  tags,
	})
}

// Decr is just Count of -1
func (c *client) Decr(name string, tags map[string]string) error {
	return c.Send(event.Event{
		Name:  name,
		Time:  time.Now(),
		Value: -1,
		Type:  metrics_types.Count,
		Tags:  tags,
	})
}

// Incr is just Count of 1
func (c *client) Incr(name string, tags map[string]string) error {
	return c.Send(event.Event{
		Name:  name,
		Time:  time.Now(),
		Value: 1,
		Type:  metrics_types.Count,
		Tags:  tags,
	})
}

// Timing sends timing information, it is an alias for TimeInMilliseconds
func (c *client) Timing(name string, value time.Duration, tags map[string]string) error {
	return c.Send(event.Event{
		Name:  name,
		Time:  time.Now(),
		Value: value,
		Type:  metrics_types.Timing,
		Tags:  tags,
	})
}

func (c *client) Close() error {
	c.Lock()
	defer c.Unlock()

	if c.closed {
		return errors.New("client is already closed")
	}

	err := c.flush()
	if err != nil {
		return err
	}

	c.closed = true
	return nil
}

// send sends the metrics to the provider
func (c *client) Send(e event.Event) error {
	c.Lock()
	defer c.Unlock()

	if c.closed {
		return errors.New("client is already closed")
	}

	c.buffer = append(c.buffer, e)
	if c.shouldFlush() {
		return c.flush()
	}

	return nil
}

// shouldFlush determines whether to flush the buffer to the provider
// it assumes the caller already has a lock on the buffer
func (c *client) shouldFlush() bool {
	return len(c.buffer) == c.maxBufferSize
}

// flush flushes the buffer to the provider
// it assumes the caller already has a lock on the buffer
func (c *client) flush() error {
	err := c.Provider.Send(c.buffer)
	if err != nil {
		return err
	}

	c.buffer = nil
	return nil
}
