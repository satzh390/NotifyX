# Custom Filters

This directory contains custom filter implementations that are automatically registered when the processor starts.

## How It Works

1. Create a new Go file in this directory (e.g., `my_custom_filter.go`)
2. Implement the `CustomFilter` interface
3. Add an `init()` function that calls `filter.RegisterCustomFilter(&YourFilter{})`
4. The filter will be automatically registered using the struct type name

## Example

```go
package custom

import (
	"github.com/notifyx/core/domain"
	"github.com/notifyx/processor/internal/filter"
)

// MyCustomFilter filters subscribers based on custom logic
type MyCustomFilter struct {
	// Add any fields you need
}

// Filter implements the CustomFilter interface
func (f *MyCustomFilter) Filter(subscriber domain.Subscriber, message map[string]interface{}, config map[string]interface{}) bool {
	// Your custom filtering logic here
	// Return true if subscriber should receive notification, false otherwise
	return true
}

// init registers this filter automatically
func init() {
	filter.RegisterCustomFilter(&MyCustomFilter{})
}
```

## Usage in Rules

Once registered, use the filter in a rule by specifying the type name (same as the struct name):

```json
{
  "eventType": "order.created",
  "channels": ["email"],
  "customFilter": {
    "type": "MyCustomFilter",
    "config": {
      "yourConfigKey": "yourConfigValue"
    }
  }
}
```

## Important Notes

- The filter type name in the rule must match the struct name exactly (e.g., "MyCustomFilter")
- Custom filters are **NOT** applied to direct emails and phone numbers (subscribers with empty ID)
- The `config` parameter in the Filter method contains the configuration from `rule.CustomFilter.Config`
- All filters in this package are automatically imported and registered when the processor starts

