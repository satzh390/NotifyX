package mongo

import (
	"testing"

	"github.com/notifyx/core/domain"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestBuildGroupFilter_WithSubscribedEventTypes(t *testing.T) {
	opts := domain.ListOptions{
		Filter: map[string]string{
			"customerId":           "cust-1",
			"subscribedEventTypes": "order.created, order.shipped ,",
		},
	}

	filter := buildGroupFilter(opts)

	require.Equal(t, bson.M{
		"customerId": "cust-1",
		"subscribedEventTypes": bson.M{
			"$in": []string{"order.created", "order.shipped"},
		},
	}, filter)
}
