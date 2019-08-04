package providers

import (
	"context"
	"fmt"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"

	"github.com/jchorl/splittrics/pkg/event"
	"github.com/jchorl/splittrics/pkg/metrics_types"
)

var _ Provider = (*stackdriverProvider)(nil)

const StackdriverBufferSize = 200

type StackdriverOpts struct {
	ProjectID string
}

type stackdriverProvider struct {
	projectID string
	client    *monitoring.MetricClient
}

func Stackdriver(opts StackdriverOpts) (Provider, error) {
	ctx := context.Background()
	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error creating stackdriver client")
	}

	return &stackdriverProvider{
		client:    c,
		projectID: opts.ProjectID,
	}, nil
}

func (p *stackdriverProvider) Send(events []event.Event) error {
	timeSeries := []*monitoringpb.TimeSeries{}

	for _, e := range events {
		metricType := fmt.Sprintf("custom.googleapis.com/%s", e.Name)
		if e.Type == metrics_types.Timing {
			metricType = metricType + "_millis"
		}

		ptime := &timestamp.Timestamp{
			Seconds: e.Time.Unix(),
		}

		value, err := getEventValue(e)
		if err != nil {
			return err
		}

		timeSeries = append(timeSeries, &monitoringpb.TimeSeries{
			Metric: &metricpb.Metric{
				Type:   metricType,
				Labels: e.Tags,
			},
			Resource: &monitoredres.MonitoredResource{
				Type: "global",
			},
			Points: []*monitoringpb.Point{{
				Interval: &monitoringpb.TimeInterval{
					StartTime: ptime,
					EndTime:   ptime,
				},
				Value: value,
			}},
		})
	}

	req := &monitoringpb.CreateTimeSeriesRequest{
		Name:       "projects/" + p.projectID,
		TimeSeries: timeSeries,
	}

	err := p.client.CreateTimeSeries(context.Background(), req)
	if err != nil {
		return errors.Wrap(err, "could not write time series value")
	}

	return nil
}

func getEventValue(e event.Event) (*monitoringpb.TypedValue, error) {
	switch e.Type {
	case metrics_types.Gauge:
		parsedVal, ok := e.Value.(float64)
		if !ok {
			return nil, errors.New("metrics with type Gauge should contain values of type float64")
		}

		return &monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_DoubleValue{
				DoubleValue: parsedVal,
			},
		}, nil

	case metrics_types.Count:
		parsedVal, ok := e.Value.(int64)
		if !ok {
			return nil, errors.New("metrics with type Gauge should contain values of type int64")
		}

		return &monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_Int64Value{
				Int64Value: parsedVal,
			},
		}, nil

	case metrics_types.Timing:
		parsedVal, ok := e.Value.(time.Duration)
		if !ok {
			return nil, errors.New("metrics with type Timing should contain values of type time.Duration")
		}

		return &monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_Int64Value{
				Int64Value: parsedVal.Nanoseconds() / 1000,
			},
		}, nil

	default:
		return nil, errors.Errorf("unrecognized event type: %s", e.Type)
	}
}
