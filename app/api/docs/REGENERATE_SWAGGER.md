# Regenerating Swagger Documentation

After making changes to API handlers, regenerate the Swagger documentation using the `swag` tool.

## Install Swag (if not already installed)

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

## Regenerate Swagger Docs

From the `app/api` directory:

**Windows:**
```cmd
cd app\api
%USERPROFILE%\go\bin\swag.exe init -g cmd/main.go -o docs
```

**Linux/Mac:**
```bash
cd app/api
$(go env GOPATH)/bin/swag init -g cmd/main.go -o docs
```

This will regenerate:
- `docs/docs.go`
- `docs/swagger.json`
- `docs/swagger.yaml`

## What Was Updated

The Swagger documentation has been updated to include:

1. **AppConfig Endpoints** (`/app-configs`):
   - POST `/app-configs` - Create app configuration
   - GET `/app-configs` - List app configurations
   - GET `/app-configs/{id}` - Get app configuration by ID
   - PUT `/app-configs/{id}` - Create or update app configuration
   - PATCH `/app-configs/{id}` - Partially update app configuration
   - DELETE `/app-configs/{id}` - Delete app configuration

2. **Updated Definitions**:
   - `domain.AppConfig` - App configuration entity
   - `domain.APNSConfig` - APNS provider configuration
   - `domain.FirebaseConfig` - Firebase provider configuration
   - `domain.Rule` - Added `metadata` field
   - `domain.Subscriber` - Added `pushTokens` field (map of appId → token)
   - `internal_handlers_rule.ruleRequest` - Added `metadata` and `customFilter` fields
   - `internal_handlers_subscriber.subscriberRequest` - Added `pushTokens` field

3. **New Request/Response Types**:
   - `internal_handlers_appconfig.appConfigRequest`
   - `internal_handlers_appconfig.appConfigPatchRequest`
   - `internal_handlers_appconfig.aPNSConfigRequest`
   - `internal_handlers_appconfig.firebaseConfigRequest`

## Verification

After regeneration, verify the Swagger UI at:
```
http://localhost:8080/swagger/index.html
```

All AppConfig endpoints should be visible under the "app-configs" tag.

