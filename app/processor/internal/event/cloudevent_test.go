package event

import (
	"encoding/json"
	"testing"

	"github.com/notifyx/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestCloudEventEnvelope_Validate_Success(t *testing.T) {
	envelope := CloudEventEnvelope[map[string]any]{
		ID:              "event-123",
		Source:          "test-source",
		SpecVersion:     "1.0",
		Type:            "order.created",
		CustomerID:      "customer-456",
		Recipients: domain.Recipients{
			SubscriberIDs: []string{"sub-1"},
		},
		Payload: map[string]any{"orderId": "123"},
	}

	err := envelope.Validate()
	assert.NoError(t, err)
}

func TestCloudEventEnvelope_Validate_MissingCustomerID(t *testing.T) {
	envelope := CloudEventEnvelope[map[string]any]{
		ID:          "event-123",
		Source:      "test-source",
		SpecVersion: "1.0",
		Type:        "order.created",
		CustomerID:  "", // Missing
		Recipients: domain.Recipients{
			SubscriberIDs: []string{"sub-1"},
		},
	}

	err := envelope.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "customerId is required")
}

func TestCloudEventEnvelope_Validate_MissingType(t *testing.T) {
	envelope := CloudEventEnvelope[map[string]any]{
		ID:          "event-123",
		Source:      "test-source",
		SpecVersion: "1.0",
		Type:        "", // Missing
		CustomerID:  "customer-456",
		Recipients: domain.Recipients{
			SubscriberIDs: []string{"sub-1"},
		},
	}

	err := envelope.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "type is required")
}

func TestCloudEventEnvelope_Validate_MissingRecipients(t *testing.T) {
	envelope := CloudEventEnvelope[map[string]any]{
		ID:          "event-123",
		Source:      "test-source",
		SpecVersion: "1.0",
		Type:        "order.created",
		CustomerID:  "customer-456",
		Recipients:  domain.Recipients{}, // Empty
	}

	err := envelope.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one recipient definition is required")
}

func TestCloudEventEnvelope_Validate_RecipientLimits(t *testing.T) {
	tests := []struct {
		name      string
		recipients domain.Recipients
		shouldErr bool
	}{
		{
			name: "too many subscriber IDs",
			recipients: domain.Recipients{
				SubscriberIDs: make([]string, MaxSubscriberIDs+1),
			},
			shouldErr: true,
		},
		{
			name: "too many groups",
			recipients: domain.Recipients{
				Groups: make([]string, MaxGroups+1),
			},
			shouldErr: true,
		},
		{
			name: "too many direct emails",
			recipients: domain.Recipients{
				DirectEmails: make([]string, MaxDirectEmails+1),
			},
			shouldErr: true,
		},
		{
			name: "too many direct phones",
			recipients: domain.Recipients{
				DirectPhones: make([]string, MaxDirectPhones+1),
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envelope := CloudEventEnvelope[map[string]any]{
				ID:          "event-123",
				Source:      "test-source",
				SpecVersion: "1.0",
				Type:        "order.created",
				CustomerID:  "customer-456",
				Recipients:  tt.recipients,
			}

			err := envelope.Validate()
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCloudEventEnvelope_EventID(t *testing.T) {
	envelope := CloudEventEnvelope[map[string]any]{
		ID: "custom-event-id",
	}

	assert.Equal(t, "custom-event-id", envelope.EventID())

	envelope2 := CloudEventEnvelope[map[string]any]{}
	eventID := envelope2.EventID()
	assert.NotEmpty(t, eventID)
}

func TestParseEnvelope_Success(t *testing.T) {
	payload := map[string]any{
		"id":              "event-123",
		"source":          "test-source",
		"specversion":     "1.0",
		"type":            "order.created",
		"customerId":      "customer-456",
		"recipients": map[string]any{
			"subscriberIds": []string{"sub-1"},
		},
		"payload": map[string]any{"orderId": "123"},
	}

	payloadJSON, _ := json.Marshal(payload)
	envelope, err := ParseEnvelope(payloadJSON)

	assert.NoError(t, err)
	assert.Equal(t, "event-123", envelope.ID)
	assert.Equal(t, "customer-456", envelope.CustomerID)
	assert.Equal(t, "order.created", envelope.Type)
}

func TestParseEnvelope_InvalidJSON(t *testing.T) {
	invalidJSON := []byte("invalid json")
	_, err := ParseEnvelope(invalidJSON)

	assert.Error(t, err)
}

func TestParseEnvelope_ValidationError(t *testing.T) {
	payload := map[string]any{
		"id":          "event-123",
		"source":      "test-source",
		"specversion": "1.0",
		"type":        "order.created",
		// Missing customerId
		"recipients": map[string]any{
			"subscriberIds": []string{"sub-1"},
		},
	}

	payloadJSON, _ := json.Marshal(payload)
	_, err := ParseEnvelope(payloadJSON)

	assert.Error(t, err)
}

