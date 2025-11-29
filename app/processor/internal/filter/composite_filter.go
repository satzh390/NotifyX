package filter

import (
	"github.com/notifyx/core/domain"
)

// CompositeFilter combines the default preferences filter with optional custom filters.
// It ensures that:
// 1. Default preferences filter is always applied
// 2. Custom filters are applied only to non-direct recipients (subscribers with IDs)
// 3. Direct recipients (empty ID) bypass custom filters
type CompositeFilter struct {
	defaultFilter *PreferencesFilter
	registry      *FilterRegistry
}

// NewCompositeFilter creates a new composite filter that combines default and custom filters.
func NewCompositeFilter(defaultFilter *PreferencesFilter, registry *FilterRegistry) *CompositeFilter {
	return &CompositeFilter{
		defaultFilter: defaultFilter,
		registry:      registry,
	}
}

// Apply applies both default and custom filters to the subscribers.
// Custom filters are only applied to non-direct recipients (subscribers with IDs).
func (cf *CompositeFilter) Apply(subscribers []domain.Subscriber, rule domain.Rule, message map[string]interface{}) []FilteredSubscriber {
	// First, apply the default preferences filter
	filtered := cf.defaultFilter.Apply(subscribers, rule, message)

	// If no custom filter is configured, return the default filter results
	if rule.CustomFilter == nil || rule.CustomFilter.Type == "" {
		return filtered
	}

	// Get the custom filter from the registry
	customFilter, ok := cf.registry.Get(rule.CustomFilter.Type)
	if !ok {
		// If custom filter is not found in registry, return default filter results
		// In production, you might want to log a warning here
		return filtered
	}

	// Apply custom filter only to non-direct recipients
	result := make([]FilteredSubscriber, 0, len(filtered))
	for _, entry := range filtered {
		// Direct recipients (empty ID) bypass custom filters
		isDirect := entry.Subscriber.ID == ""
		if isDirect {
			// Include direct recipients without applying custom filter
			result = append(result, entry)
			continue
		}

		// Apply custom filter to non-direct recipients
		// Pass the filter configuration from the rule
		filterConfig := make(map[string]interface{})
		if rule.CustomFilter != nil && rule.CustomFilter.Config != nil {
			filterConfig = rule.CustomFilter.Config
		}
		if customFilter.Filter(entry.Subscriber, message, filterConfig) {
			result = append(result, entry)
		}
	}

	return result
}
