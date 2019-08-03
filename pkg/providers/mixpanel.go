package providers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/jchorl/splittrics/pkg/event"
)

var _ Provider = (*mixpanelProvider)(nil)

type MixpanelOpts struct {
	Token string
}

type mixpanelProvider struct {
	url   string
	token string
}

func Mixpanel(opts MixpanelOpts) Provider {
	return &mixpanelProvider{
		url:   "http://api.mixpanel.com/track/",
		token: opts.Token,
	}
}

func (p *mixpanelProvider) Send(events []event.Event) error {
	mixpanelEvents := []mixpanelEvent{}
	for _, e := range events {
		properties := map[string]interface{}{
			"token": p.token,
			"time":  e.Time,
			"value": e.Value,
		}

		for k, v := range e.Tags {
			if k == "token" || k == "time" || k == "value" {
				return fmt.Errorf("cannot specify reserved keywork tag %s", k)
			}
			properties[k] = v
		}

		mixpanelEvents = append(mixpanelEvents, mixpanelEvent{
			Event:      e.Name,
			Properties: properties,
		})
	}

	jsonEncoded, err := json.Marshal(events)
	if err != nil {
		return errors.Wrap(err, "error sending events")
	}

	b64Encoded := make([]byte, base64.StdEncoding.EncodedLen(len(jsonEncoded)))
	base64.StdEncoding.Encode(b64Encoded, jsonEncoded)

	body := bytes.NewBuffer(append([]byte("data="), b64Encoded...))
	resp, err := http.Post(p.url, "application/x-www-form-urlencoded", body)
	if err != nil {
		return errors.Wrap(err, "error sending events")
	}

	defer resp.Body.Close()

	return nil
}

type mixpanelEvent struct {
	Event      string                 `json:"event"`
	Properties map[string]interface{} `json:"properties"`
}
