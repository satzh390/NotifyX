# Multi-App Push Notification Support - Design Document

## Overview
This document describes the changes to support multiple push notification apps per Organization.

## Changes Summary

### 1. Domain Model Changes

#### Subscriber
- **Changed**: `PushToken string` → `PushTokens map[string]string` (appId → token)
- **Backward compatibility**: Keep `PushToken` field (deprecated) for migration period

#### Rule
- **Added**: `Metadata map[string]string` field
- **Usage**: Store `appId` in `Rule.Metadata["appId"]` to specify which app to use for push notifications

#### New: AppConfig Entity
- Bound to Organization (not Customer)
- Contains push provider configuration (APNS or Firebase)
- Fields:
  - `ID` - App identifier
  - `OrgID` - Organization ID
  - `Name` - App name
  - `Provider` - "apns", "firebase", or "mock"
  - `APNS` - APNS configuration (if provider is "apns")
  - `Firebase` - Firebase configuration (if provider is "firebase")
  - `Metadata` - Additional metadata
  - `CreatedAt`, `UpdatedAt`

### 2. Storage Changes

#### New: AppConfigStore Interface
```go
type AppConfigStore interface {
    Put(ctx context.Context, appConfig domain.AppConfig) error
    Get(ctx context.Context, orgID, appID string) (domain.AppConfig, error)
    List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.AppConfig], error)
    Delete(ctx context.Context, orgID, appID string) error
}
```

#### MongoDB Implementation
- New collection: `appConfigs`
- Index: `{orgId: 1, id: 1}`

### 3. Processor Changes

#### Extract appId from Rule.Metadata
- When processing push channel, check `Rule.Metadata["appId"]`
- Pass `appId` to `DeliveryTask.Metadata["appId"]`

### 4. Push Worker Changes

#### ProviderManager
- **Purpose**: Manage multiple push providers (one per app)
- **Caching**: Cache providers per appId to avoid recreating
- **Methods**:
  - `GetProvider(ctx, orgID, appID) (Provider, error)` - Get or create provider
  - `Close()` - Cleanup all providers

#### Push Worker Flow
1. Extract `appId` from `task.Metadata["appId"]`
2. Get `orgID` from Customer (via CustomerID lookup)
3. Load AppConfig from store using `orgID` and `appID`
4. Get provider from ProviderManager (cached)
5. Get push token from `task.Subscriber.PushTokens[appId]`
6. Send notification using correct provider

### 5. API Changes

#### New Endpoints: AppConfig CRUD
- `POST /api/v1/app-configs` - Create app config
- `GET /api/v1/app-configs/:id` - Get app config
- `PUT /api/v1/app-configs/:id` - Update app config
- `DELETE /api/v1/app-configs/:id` - Delete app config
- `GET /api/v1/app-configs` - List app configs

### 6. Migration Notes

#### Subscriber Migration
- Existing `PushToken` field will continue to work
- Migration script needed to convert `PushToken` → `PushTokens["default"]`
- Or handle both fields in code during transition

#### Rule Migration
- Existing rules without `Metadata["appId"]` will need default app
- Consider fallback to default app if `appId` not specified

## Usage Example

### 1. Create AppConfig
```json
POST /api/v1/app-configs
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

### 2. Create Rule with appId
```json
POST /api/v1/rules
{
  "eventType": "order.created",
  "customerId": "customer-123",
  "channels": ["push"],
  "templateRefs": {
    "push": "template-123"
  },
  "metadata": {
    "appId": "ios-app-1"
  }
}
```

### 3. Register Subscriber with Multiple Tokens
```json
POST /api/v1/subscribers
{
  "subscriberId": "user-123",
  "customerId": "customer-123",
  "pushTokens": {
    "ios-app-1": "apns-token-123",
    "android-app-1": "fcm-token-456"
  }
}
```

### 4. Send Event
```json
POST /api/v1/events
{
  "customerId": "customer-123",
  "type": "order.created",
  "payload": {
    "orderId": "12345"
  }
}
```

## Flow Diagram

```
Event → Processor → Rule (with appId in Metadata)
  ↓
DeliveryTask (with appId in Metadata)
  ↓
Push Worker → Extract appId from task.Metadata
  ↓
Get orgID from Customer
  ↓
Load AppConfig(orgID, appID) from Store
  ↓
Get Provider from ProviderManager (cached)
  ↓
Get token from Subscriber.PushTokens[appId]
  ↓
Send via Provider
```

