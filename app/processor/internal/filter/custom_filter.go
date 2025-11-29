package filter

import (
	"github.com/notifyx/core/domain"
)

// CustomFilter defines the interface that custom filters must implement.
// Custom filters are used to filter subscribers based on custom business logic.
// The Filter method should return true if the subscriber should receive the notification,
// false otherwise.
type CustomFilter interface {
	// Filter determines if a subscriber should receive a notification based on
	// the subscriber data, the event message payload, and the filter configuration.
	// The config parameter contains the filter-specific configuration from the rule.
	// Returns true if the subscriber should be included, false otherwise.
	Filter(subscriber domain.Subscriber, message map[string]interface{}, config map[string]interface{}) bool
}

// FilterRegistry manages custom filter implementations.
// It allows registering and retrieving custom filters by type name.
type FilterRegistry struct {
	filters map[string]CustomFilter
}

// NewFilterRegistry creates a new filter registry.
func NewFilterRegistry() *FilterRegistry {
	return &FilterRegistry{
		filters: make(map[string]CustomFilter),
	}
}

// Register registers a custom filter implementation with the given type name.
func (r *FilterRegistry) Register(filterType string, filter CustomFilter) {
	r.filters[filterType] = filter
}

// Get retrieves a custom filter by type name.
// Returns the filter and true if found, nil and false otherwise.
func (r *FilterRegistry) Get(filterType string) (CustomFilter, bool) {
	filter, ok := r.filters[filterType]
	return filter, ok
}

// DefaultFilterRegistry is the global default registry instance.
// It can be used for registering filters at application startup.
var DefaultFilterRegistry = NewFilterRegistry()
