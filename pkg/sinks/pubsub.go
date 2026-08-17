package sinks

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub/v2"
	pubsubpb "cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/blaxel-ai/kubernetes-event-exporter/pkg/kube"
	"github.com/rs/zerolog/log"
)

type PubsubConfig struct {
	GcloudProjectId string `yaml:"gcloud_project_id"`
	Topic           string `yaml:"topic"`
	CreateTopic     bool   `yaml:"create_topic"`
}

type PubsubSink struct {
	cfg          *PubsubConfig
	pubsubClient *pubsub.Client
	publisher    *pubsub.Publisher
}

func NewPubsubSink(cfg *PubsubConfig) (Sink, error) {
	ctx := context.Background()
	pubsubClient, err := pubsub.NewClient(ctx, cfg.GcloudProjectId)
	if err != nil {
		return nil, err
	}

	topicName := fmt.Sprintf("projects/%s/topics/%s", cfg.GcloudProjectId, cfg.Topic)

	if cfg.CreateTopic {
		_, err = pubsubClient.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{
			Name: topicName,
		})
		if err != nil {
			return nil, err
		}
		log.Info().Msgf("pubsub: created topic: %s", cfg.Topic)
	}

	publisher := pubsubClient.Publisher(topicName)

	return &PubsubSink{
		pubsubClient: pubsubClient,
		publisher:    publisher,
		cfg:          cfg,
	}, nil
}

func (ps *PubsubSink) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	msg := &pubsub.Message{
		Data: ev.ToJSON(),
	}
	_, err := ps.publisher.Publish(ctx, msg).Get(ctx)
	return err
}

func (ps *PubsubSink) Close() {
	log.Info().Msgf("pubsub: Closing publisher...")
	ps.publisher.Stop()
	ps.pubsubClient.Close()
}
