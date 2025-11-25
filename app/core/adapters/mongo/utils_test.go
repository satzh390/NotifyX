package mongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCommaSeparatedString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single value",
			input:    "order.created",
			expected: []string{"order.created"},
		},
		{
			name:     "multiple values with spaces",
			input:    " order.created ,order.shipped , ,order.delivered ",
			expected: []string{"order.created", "order.shipped", "order.delivered"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, ParseCommaSeparatedString(tc.input))
		})
	}
}
