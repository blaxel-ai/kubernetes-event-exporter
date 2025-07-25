package sinks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/blaxel-ai/kubernetes-event-exporter/pkg/kube"
	"github.com/rs/zerolog/log"
)

type EventBridgeConfig struct {
	DetailType   string                 `yaml:"detailType"`
	Detail       map[string]interface{} `yaml:"detail"`
	Source       string                 `yaml:"source"`
	EventBusName string                 `yaml:"eventBusName"`
	Region       string                 `yaml:"region"`
}

type EventBridgeSink struct {
	cfg *EventBridgeConfig
	svc *eventbridge.EventBridge
}

func NewEventBridgeSink(cfg *EventBridgeConfig) (Sink, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region),
		Retryer: client.DefaultRetryer{
			NumMaxRetries:    client.DefaultRetryerMaxNumRetries,
			MinRetryDelay:    client.DefaultRetryerMinRetryDelay,
			MinThrottleDelay: client.DefaultRetryerMinThrottleDelay,
			MaxRetryDelay:    client.DefaultRetryerMaxRetryDelay,
			MaxThrottleDelay: client.DefaultRetryerMaxThrottleDelay,
		},
	},
	)
	if err != nil {
		return nil, err
	}

	svc := eventbridge.New(sess)
	return &EventBridgeSink{
		cfg: cfg,
		svc: svc,
	}, nil
}

func (s *EventBridgeSink) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	log.Info().Msg("Sending event to EventBridge ")
	var toSend string
	if s.cfg.Detail != nil {
		res, err := convertLayoutTemplate(s.cfg.Detail, ev)
		if err != nil {
			return err
		}

		b, err := json.Marshal(res)
		toSend = string(b)
		if err != nil {
			return err
		}
	} else {
		toSend = string(ev.ToJSON())
	}
	tym := time.Now()
	inputRequest := eventbridge.PutEventsRequestEntry{
		Detail:       &toSend,
		DetailType:   &s.cfg.DetailType,
		Time:         &tym,
		Source:       &s.cfg.Source,
		EventBusName: &s.cfg.EventBusName,
	}
	log.Info().Str("InputEvent", inputRequest.String()).Msg("Request")

	req, _ := s.svc.PutEventsRequest(&eventbridge.PutEventsInput{Entries: []*eventbridge.PutEventsRequestEntry{&inputRequest}})
	// TODO: Retry failed events
	err := req.Send()
	if err != nil {
		log.Error().Err(err).Msg("EventBridge Error")
		return err
	}
	return nil
}

func (s *EventBridgeSink) Close() {
}
