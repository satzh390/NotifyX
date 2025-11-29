package custom

import (
	"github.com/notifyx/core/domain"
	"github.com/notifyx/processor/internal/filter"
)

// ExampleCustomFilter is an example implementation of a custom filter.
// This demonstrates how to create a custom filter that will be automatically registered.
//
// The filter type name will be "ExampleCustomFilter" (same as the struct name).
//
// Configure it in a rule:
//   {
//     "eventType": "order.created",
//     "customFilter": {
//       "type": "ExampleCustomFilter",
//       "config": {
//         "minOrderAmount": 100.0
//       }
//     }
//   }
type ExampleCustomFilter struct {
	// This example filter filters based on order amount
}

// Filter implements the CustomFilter interface.
// This example checks if the order amount in the message payload meets a threshold.
func (f *ExampleCustomFilter) Filter(subscriber domain.Subscriber, message map[string]interface{}, config map[string]interface{}) bool {
	// Get threshold from config, default to 100.0 if not specified
	minOrderAmount := 100.0
	if val, ok := config["minOrderAmount"].(float64); ok {
		minOrderAmount = val
	}

	// Check if order amount meets threshold
	if orderAmount, ok := message["orderAmount"].(float64); ok {
		return orderAmount >= minOrderAmount
	}

	// If order amount not found or doesn't meet criteria, filter out
	return false
}

// init registers this filter automatically on package import
func init() {
	filter.RegisterCustomFilter(&ExampleCustomFilter{})
}

