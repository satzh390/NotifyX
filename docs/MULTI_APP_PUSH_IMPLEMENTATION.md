# Multi-App Push Notification - Implementation Summary

## ✅ Completed Changes

### 1. Domain Model (`app/core/domain/types.go`)
- ✅ Added `PushTokens map[string]string` to `Subscriber` (appId → token)
- ✅ Kept `PushToken string` for backward compatibility (deprecated)
- ✅ Added `Metadata map[string]string` to `Rule` (stores appId)
- ✅ Added `AppConfig` entity bound to Organization
- ✅ Added `PushProviderType` enum (apns, firebase, mock)
- ✅ Added `APNSConfig` and `FirebaseConfig` structs

### 2. Storage Layer
- ✅ Added `AppConfigStore` interface (`app/core/storage/storage.go`)
- ✅ Added MongoDB implementation (`app/core/adapters/mongo/appconfig_repository.go`)
- ✅ Updated `Stores` struct to include `AppConfigs`
- ✅ Updated `mongo_store.go` to initialize AppConfig repository

### 3. Processor (`app/processor/internal/pipeline/processor.go`)
- ✅ Extracts `appId` from `Rule.Metadata["appId"]` for push channel
- ✅ Passes `appId` to `DeliveryTask.Metadata["appId"]`

### 4. Push Worker
- ✅ Created `ProviderManager` (`app/workers/worker-push/internal/provider/provider_manager.go`)
  - Caches providers per appId
  - Loads AppConfig from store
  - Creates providers dynamically based on AppConfig
- ✅ Updated `PushWorker` (`app/workers/worker-push/internal/worker/push_worker.go`)
  - Uses ProviderManager instead of single provider
  - Extracts appId from task metadata
  - Selects token from `PushTokens[appId]` with fallback to legacy `PushToken`
- ✅ Updated `main.go` (`app/workers/worker-push/cmd/main.go`)
  - Creates ProviderManager instead of single provider
  - Passes AppConfigStore and CustomerStore to ProviderManager

## ⏳ Pending Changes

### 1. API Endpoints (`app/api`)
- ⏳ AppConfig CRUD endpoints:
  - `POST /api/v1/app-configs` - Create
  - `GET /api/v1/app-configs/:id` - Get
  - `PUT /api/v1/app-configs/:id` - Update
  - `DELETE /api/v1/app-configs/:id` - Delete
  - `GET /api/v1/app-configs` - List

### 2. Configuration
- ⏳ Remove single provider config from worker config (now uses AppConfig store)
- ⏳ Update Helm charts/configmaps

## 🔄 Migration Notes

### Subscriber Migration
Existing subscribers with `PushToken` will continue to work:
- Code checks `PushTokens[appId]` first
- Falls back to `PushToken` if not found
- Migration script can convert: `PushToken` → `PushTokens["default"]`

### Rule Migration
- Existing rules without `Metadata["appId"]` will fail
- Need to add `appId` to Rule.Metadata for push rules
- Consider adding default app fallback logic if needed

## 📋 Usage Flow

1. **Create AppConfig** (via API - pending)
   ```json
   {
     "id": "ios-app-1",
     "orgId": "org-123",
     "name": "iOS App",
     "provider": "apns",
     "apns": {
       "keyId": "ABC123",
       "teamId": "TEAM123",
       "bundleId": "com.company.app",
       "keyPath": "/secrets/apns-app1.p8",
       "production": false
     }
   }
   ```

2. **Create Rule with appId**
   ```json
   {
     "eventType": "order.created",
     "customerId": "customer-123",
     "channels": ["push"],
     "templateRefs": {"push": "template-123"},
     "metadata": {"appId": "ios-app-1"}
   }
   ```

3. **Register Subscriber with Multiple Tokens**
   ```json
   {
     "subscriberId": "user-123",
     "customerId": "customer-123",
     "pushTokens": {
       "ios-app-1": "apns-token-123",
       "android-app-1": "fcm-token-456"
     }
   }
   ```

4. **Send Event** (appId automatically extracted from Rule)

## 🎯 Key Design Decisions

1. **AppConfig bound to Organization** (not Customer)
   - Allows sharing apps across customers in same org
   - Matches multi-tenant SaaS model

2. **appId from Rule.Metadata**
   - Rule-level configuration
   - Can have different apps for different event types

3. **Provider Caching**
   - Providers cached per appId
   - Avoids recreating providers on every request
   - Thread-safe with sync.RWMutex

4. **Backward Compatibility**
   - Legacy `PushToken` field still supported
   - Falls back if `PushTokens[appId]` not found

## 🔍 Testing Checklist

- [ ] Test ProviderManager with multiple apps
- [ ] Test provider caching
- [ ] Test backward compatibility with legacy PushToken
- [ ] Test error handling (missing appId, missing AppConfig, etc.)
- [ ] Test processor extracts appId correctly
- [ ] Integration test with real providers

