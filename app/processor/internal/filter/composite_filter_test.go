package filter

import (
	"testing"

	"github.com/notifyx/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestCompositeFilter_Apply_NoCustomFilter(t *testing.T) {
	preferencesFilter := NewPreferencesFilter()
	registry := NewFilterRegistry()
	compositeFilter := NewCompositeFilter(preferencesFilter, registry)

	subscribers := []domain.Subscriber{
		{
			ID:         "sub-1",
			CustomerID: "customer-1",
			Email:      "test@example.com",
			Preferences: domain.SubscriberPrefs{
				Channels: map[domain.ChannelType]bool{
					domain.ChannelEmail: true,
				},
			},
		},
	}

	rule := domain.Rule{
		EventType:  "order.created",
		CustomerID: "customer-1",
		Channels:   []domain.ChannelType{domain.ChannelEmail},
		TemplateRefs: map[domain.ChannelType]string{
			domain.ChannelEmail: "template-1",
		},
		// No custom filter
	}

	message := map[string]interface{}{"orderId": "123"}

	result := compositeFilter.Apply(subscribers, rule, message)

	assert.Len(t, result, 1)
	assert.Equal(t, "sub-1", result[0].Subscriber.ID)
	assert.Contains(t, result[0].Channels, domain.ChannelEmail)
}

func TestCompositeFilter_Apply_WithCustomFilter(t *testing.T) {
	preferencesFilter := NewPreferencesFilter()
	registry := NewFilterRegistry()

	// Register a custom filter that filters based on order amount
	customFilter := &testCustomFilter{minAmount: 100.0}
	registry.Register("TestCustomFilter", customFilter)

	compositeFilter := NewCompositeFilter(preferencesFilter, registry)

	subscribers := []domain.Subscriber{
		{
			ID:         "sub-1",
			CustomerID: "customer-1",
			Email:      "test@example.com",
			Preferences: domain.SubscriberPrefs{
				Channels: map[domain.ChannelType]bool{
					domain.ChannelEmail: true,
				},
			},
		},
		{
			ID:         "sub-2",
			CustomerID: "customer-1",
			Email:      "test2@example.com",
			Preferences: domain.SubscriberPrefs{
				Channels: map[domain.ChannelType]bool{
					domain.ChannelEmail: true,
				},
			},
		},
	}

	rule := domain.Rule{
		EventType:  "order.created",
		CustomerID: "customer-1",
		Channels:   []domain.ChannelType{domain.ChannelEmail},
		TemplateRefs: map[domain.ChannelType]string{
			domain.ChannelEmail: "template-1",
		},
		CustomFilter: &domain.CustomFilterConfig{
			Type: "TestCustomFilter",
			Config: map[string]interface{}{
				"minOrderAmount": 100.0,
			},
		},
	}

	// Test with order amount >= 100 (both should pass)
	message := map[string]interface{}{"orderAmount": 150.0}
	result := compositeFilter.Apply(subscribers, rule, message)
	assert.Len(t, result, 2, "both subscribers should pass when order amount >= 100")

	// Test with order amount < 100 (both should be filtered out)
	message = map[string]interface{}{"orderAmount": 50.0}
	result = compositeFilter.Apply(subscribers, rule, message)
	assert.Len(t, result, 0, "both subscribers should be filtered out when order amount < 100")
}

func TestCompositeFilter_Apply_DirectRecipientsBypassCustomFilter(t *testing.T) {
	preferencesFilter := NewPreferencesFilter()
	registry := NewFilterRegistry()

	// Register a custom filter that would filter out everyone
	customFilter := &testCustomFilter{minAmount: 10000.0} // Very high threshold
	registry.Register("TestCustomFilter", customFilter)

	compositeFilter := NewCompositeFilter(preferencesFilter, registry)

	subscribers := []domain.Subscriber{
		{
			ID:         "", // Direct recipient (empty ID)
			CustomerID: "customer-1",
			Email:      "direct@example.com",
			Preferences: domain.SubscriberPrefs{
				Channels: map[domain.ChannelType]bool{
					domain.ChannelEmail: true,
				},
			},
		},
		{
			ID:         "sub-1", // Regular subscriber
			CustomerID: "customer-1",
			Email:      "test@example.com",
			Preferences: domain.SubscriberPrefs{
				Channels: map[domain.ChannelType]bool{
					domain.ChannelEmail: true,
				},
			},
		},
	}

	rule := domain.Rule{
		EventType:  "order.created",
		CustomerID: "customer-1",
		Channels:   []domain.ChannelType{domain.ChannelEmail},
		TemplateRefs: map[domain.ChannelType]string{
			domain.ChannelEmail: "template-1",
		},
		CustomFilter: &domain.CustomFilterConfig{
			Type: "TestCustomFilter",
			Config: map[string]interface{}{
				"minOrderAmount": 10000.0,
			},
		},
	}

	message := map[string]interface{}{"orderAmount": 50.0} // Low amount

	result := compositeFilter.Apply(subscribers, rule, message)

	// Direct recipient should pass (bypasses custom filter)
	// Regular subscriber should be filtered out by custom filter
	assert.Len(t, result, 1)
	assert.Empty(t, result[0].Subscriber.ID) // Direct recipient
	assert.Equal(t, "direct@example.com", result[0].Subscriber.Email)
}

func TestCompositeFilter_Apply_CustomFilterNotFound(t *testing.T) {
	preferencesFilter := NewPreferencesFilter()
	registry := NewFilterRegistry()
	compositeFilter := NewCompositeFilter(preferencesFilter, registry)

	subscribers := []domain.Subscriber{
		{
			ID:         "sub-1",
			CustomerID: "customer-1",
			Email:      "test@example.com",
			Preferences: domain.SubscriberPrefs{
				Channels: map[domain.ChannelType]bool{
					domain.ChannelEmail: true,
				},
			},
		},
	}

	rule := domain.Rule{
		EventType:  "order.created",
		CustomerID: "customer-1",
		Channels:   []domain.ChannelType{domain.ChannelEmail},
		TemplateRefs: map[domain.ChannelType]string{
			domain.ChannelEmail: "template-1",
		},
		CustomFilter: &domain.CustomFilterConfig{
			Type: "NonExistentFilter", // Filter not registered
		},
	}

	message := map[string]interface{}{"orderId": "123"}

	result := compositeFilter.Apply(subscribers, rule, message)

	// Should fall back to default filter results when custom filter not found
	assert.Len(t, result, 1)
	assert.Equal(t, "sub-1", result[0].Subscriber.ID)
}

// testCustomFilter is a test implementation of CustomFilter
type testCustomFilter struct {
	minAmount float64
}

func (f *testCustomFilter) Filter(subscriber domain.Subscriber, message map[string]interface{}, config map[string]interface{}) bool {
	// Get min amount from config if provided, otherwise use default
	minAmount := f.minAmount
	if val, ok := config["minOrderAmount"].(float64); ok {
		minAmount = val
	}

	// Check order amount
	if orderAmount, ok := message["orderAmount"].(float64); ok {
		return orderAmount >= minAmount
	}

	// If no order amount, allow by default for testing
	return true
}

func TestFilterRegistry_RegisterAndGet(t *testing.T) {
	registry := NewFilterRegistry()

	filter1 := &testCustomFilter{minAmount: 100.0}
	filter2 := &testCustomFilter{minAmount: 200.0}

	registry.Register("Filter1", filter1)
	registry.Register("Filter2", filter2)

	// Test Get
	retrieved1, ok := registry.Get("Filter1")
	assert.True(t, ok)
	assert.Equal(t, filter1, retrieved1)

	retrieved2, ok := registry.Get("Filter2")
	assert.True(t, ok)
	assert.Equal(t, filter2, retrieved2)

	// Test non-existent filter
	_, ok = registry.Get("NonExistent")
	assert.False(t, ok)
}

func TestRegisterCustomFilter(t *testing.T) {
	// Clear the default registry for testing
	DefaultFilterRegistry = NewFilterRegistry()

	filter := &testCustomFilter{minAmount: 100.0}
	RegisterCustomFilter(filter)

	// Should be registered with type name "testCustomFilter"
	retrieved, ok := DefaultFilterRegistry.Get("testCustomFilter")
	assert.True(t, ok)
	assert.Equal(t, filter, retrieved)
}

func TestRegisterCustomFilter_NilFilter(t *testing.T) {
	// Should not panic or error
	RegisterCustomFilter(nil)
	// No assertion needed, just ensure it doesn't panic
}

func TestGetTypeName(t *testing.T) {
	tests := []struct {
		name     string
		filter   CustomFilter
		expected string
	}{
		{
			name:     "pointer to struct",
			filter:   &testCustomFilter{},
			expected: "testCustomFilter",
		},
		{
			name:     "nil filter",
			filter:   nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTypeName(tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompositeFilter_Apply_EmptyCustomFilterType(t *testing.T) {
	preferencesFilter := NewPreferencesFilter()
	registry := NewFilterRegistry()
	compositeFilter := NewCompositeFilter(preferencesFilter, registry)

	subscribers := []domain.Subscriber{
		{
			ID:         "sub-1",
			CustomerID: "customer-1",
			Email:      "test@example.com",
			Preferences: domain.SubscriberPrefs{
				Channels: map[domain.ChannelType]bool{
					domain.ChannelEmail: true,
				},
			},
		},
	}

	rule := domain.Rule{
		EventType:  "order.created",
		CustomerID: "customer-1",
		Channels:   []domain.ChannelType{domain.ChannelEmail},
		TemplateRefs: map[domain.ChannelType]string{
			domain.ChannelEmail: "template-1",
		},
		CustomFilter: &domain.CustomFilterConfig{
			Type: "", // Empty type should be treated as no custom filter
		},
	}

	message := map[string]interface{}{"orderId": "123"}

	result := compositeFilter.Apply(subscribers, rule, message)

	// Should use default filter only
	assert.Len(t, result, 1)
}

