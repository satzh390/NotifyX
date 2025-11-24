package fanout

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"

	"github.com/notifyx/core/domain"
)

type Envelope struct {
	Channel domain.ChannelType
	Task    domain.DeliveryTask
}

type Publisher interface {
	Publish(ctx context.Context, envelopes []Envelope) error
	Close(ctx context.Context) error
}

type KafkaPublisher struct {
	writers map[domain.ChannelType]*kafka.Writer
}

func NewKafkaPublisher(brokers []string, topics map[domain.ChannelType]string) *KafkaPublisher {
	writers := make(map[domain.ChannelType]*kafka.Writer, len(topics))
	for channel, topic := range topics {
		writers[channel] = &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		}
	}
	return &KafkaPublisher{writers: writers}
}

func (publisher *KafkaPublisher) Publish(ctx context.Context, envelopes []Envelope) error {
	if len(envelopes) == 0 {
		return nil
	}

	grouped := map[domain.ChannelType][]kafka.Message{}
	for _, env := range envelopes {
		if _, ok := publisher.writers[env.Channel]; !ok {
			return fmt.Errorf("fanout: missing topic for channel %s", env.Channel)
		}
		payload, err := json.Marshal(env.Task)
		if err != nil {
			return fmt.Errorf("fanout: marshal task: %w", err)
		}
		grouped[env.Channel] = append(grouped[env.Channel], kafka.Message{
			Key:   []byte(env.Task.TaskID),
			Value: payload,
		})
	}

	for channel, messages := range grouped {
		writer := publisher.writers[channel]
		if err := writer.WriteMessages(ctx, messages...); err != nil {
			return fmt.Errorf("fanout: write channel %s: %w", channel, err)
		}
	}
	return nil
}

func (publisher *KafkaPublisher) Close(ctx context.Context) error {
	var firstErr error
	for _, writer := range publisher.writers {
		if err := writer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
