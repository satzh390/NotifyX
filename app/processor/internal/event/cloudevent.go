package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/notifyx/core/domain"
)

const (
	MaxSubscriberIDs = 5000
	MaxGroups        = 100
	MaxDirectEmails  = 5000
	MaxDirectPhones  = 5000
)

// CloudEventEnvelope models the CloudEvents v1.0 contract with NotifyX extensions.
type CloudEventEnvelope[T any] struct {
	ID              string            `json:"id"`
	Source          string            `json:"source"`
	SpecVersion     string            `json:"specversion"`
	Type            string            `json:"type"`
	Subject         string            `json:"subject,omitempty"`
	Time            time.Time         `json:"time"`
	DataContentType string            `json:"datacontenttype,omitempty"`
	DataSchema      string            `json:"dataschema,omitempty"`
	Data            T                 `json:"data"`
	CustomerID      string            `json:"customerId"`
	Recipients      domain.Recipients `json:"recipients"`
	Payload         map[string]any    `json:"payload,omitempty"`
	Meta            map[string]string `json:"meta,omitempty"`
	Extensions      map[string]any    `json:"extensions,omitempty"`
	RawData         json.RawMessage   `json:"-"`
	rawPayload      json.RawMessage   `json:"-"`
}

func (envelope *CloudEventEnvelope[T]) EventID() string {
	if envelope.ID != "" {
		return envelope.ID
	}
	return uuid.NewString()
}

func (envelope *CloudEventEnvelope[T]) Validate() error {
	switch {
	case envelope.CustomerID == "":
		return errors.New("event: customerId is required")
	case envelope.Type == "":
		return errors.New("event: type is required")
	case envelope.SpecVersion == "":
		return errors.New("event: specversion is required")
	case envelope.Source == "":
		return errors.New("event: source is required")
	}
	if envelope.Time.IsZero() {
		envelope.Time = time.Now().UTC()
	}

	// Validate recipient limits
	if len(envelope.Recipients.SubscriberIDs) > MaxSubscriberIDs {
		return fmt.Errorf("event: subscriberIds exceeds maximum of %d", MaxSubscriberIDs)
	}
	if len(envelope.Recipients.Groups) > MaxGroups {
		return fmt.Errorf("event: groups exceeds maximum of %d", MaxGroups)
	}
	if len(envelope.Recipients.DirectEmails) > MaxDirectEmails {
		return fmt.Errorf("event: directEmails exceeds maximum of %d", MaxDirectEmails)
	}
	if len(envelope.Recipients.DirectPhones) > MaxDirectPhones {
		return fmt.Errorf("event: directPhones exceeds maximum of %d", MaxDirectPhones)
	}

	// Validate at least one recipient definition exists
	if envelope.Recipients.Broadcast ||
		len(envelope.Recipients.Groups) > 0 ||
		len(envelope.Recipients.SubscriberIDs) > 0 ||
		len(envelope.Recipients.DirectEmails) > 0 ||
		len(envelope.Recipients.DirectPhones) > 0 {
		return nil
	}
	return errors.New("event: at least one recipient definition is required")
}

func ParseEnvelope(payload []byte) (CloudEventEnvelope[map[string]any], error) {
	var env CloudEventEnvelope[map[string]any]
	if err := json.Unmarshal(payload, &env); err != nil {
		return CloudEventEnvelope[map[string]any]{}, fmt.Errorf("event: decode: %w", err)
	}
	env.RawData = json.RawMessage(payload)
	if env.Payload == nil {
		env.Payload = env.Data
	}
	if err := env.Validate(); err != nil {
		return CloudEventEnvelope[map[string]any]{}, err
	}
	return env, nil
}
