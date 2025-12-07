//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	mongoadapter "github.com/notifyx/core/adapters/mongo"
	"github.com/notifyx/core/storage"
	"github.com/stretchr/testify/require"
)

// SetupTestMongoDB connects to test MongoDB and returns stores and cleanup function
func SetupTestMongoDB(t *testing.T) (storage.Stores, func()) {
	ctx := context.Background()

	stores, cleanup, err := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})
	require.NoError(t, err, "Failed to connect to test MongoDB")

	return stores, func() {
		if cleanup != nil {
			_ = cleanup(ctx)
		}
	}
}
