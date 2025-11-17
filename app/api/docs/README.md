# NotifyX API Documentation

## Overview

The NotifyX API provides REST endpoints for managing subscribers, groups, rules, and templates for an event-driven notification system.

## Base URL

```
http://localhost:3000/api/v1
```

## Authentication

All endpoints require JWT authentication via the `Authorization` header:

```
Authorization: Bearer <access_token>
```

The JWT must contain:
- `orgId` claim (string) - Organization identifier for tenant partitioning
- `scope` or `scp` claim (string or array) - Space-delimited scopes: `notify:read`, `notify:write`

## Authorization

- **Read operations** require `notify:read` scope
- **Write operations** (POST, PUT, DELETE) require `notify:write` scope

## Common Query Parameters

### Pagination

- `page` (integer, default: 1) - Page number (1-based)
- `pageSize` (integer, default: 20, max: 100) - Items per page

### Sorting

- `sortBy` (string) - Sort specification
  - Format: `field:direction` or `field` (defaults to asc)
  - Multiple fields: `field1:asc,field2:desc`
  - Examples: `createdAt:desc`, `name:asc,createdAt:desc`

### Filtering

- `groupId` (string) - Filter subscribers by group ID (subscribers endpoint only)

## Response Format

### Success Responses

- `200 OK` - Successful GET/PUT request
- `201 Created` - Successful POST request
- `204 No Content` - Successful DELETE request

### Error Responses

- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

Error response format:
```json
{
  "message": "Error description"
}
```

## Paginated List Response

List endpoints return paginated results:

```json
{
  "items": [...],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "totalCount": 150,
    "totalPages": 8
  }
}
```

## Update Operations (JSON Merge Patch)

Update operations (PUT) use JSON Merge Patch semantics (RFC 7396):

- Include only fields you want to update
- Nested objects are merged recursively
- Set a field to `null` to remove it
- Omitted fields remain unchanged

Example:
```json
PATCH /api/v1/groups/123
{
  "name": "New Name",
  "description": "Updated description"
}
```

## Endpoints

### Subscribers

- `GET /subscribers` - List subscribers (with pagination, sorting, group filter)
- `POST /subscribers` - Create subscriber
- `GET /subscribers/{id}` - Get subscriber
- `PUT /subscribers/{id}` - Update subscriber
- `DELETE /subscribers/{id}` - Delete subscriber

### Groups

- `GET /groups` - List groups (with pagination, sorting)
- `POST /groups` - Create group
- `GET /groups/{id}` - Get group
- `PUT /groups/{id}` - Update group
- `DELETE /groups/{id}` - Delete group

### Rules

- `GET /rules` - List rules (with pagination, sorting)
- `POST /rules` - Create rule
- `GET /rules/{eventType}` - Get rule by event type
- `PUT /rules/{eventType}` - Update rule
- `DELETE /rules/{eventType}` - Delete rule

### Templates

- `POST /templates` - Create template
- `GET /templates/{id}` - Get template
- `PUT /templates/{id}` - Update template
- `DELETE /templates/{id}` - Delete template

## OpenAPI Specification

A complete OpenAPI 3.0 specification is available at `openapi.yaml`. You can use tools like Swagger UI or Postman to import and explore the API.

## Examples

### Create a Subscriber

```bash
curl -X POST http://localhost:3000/api/v1/subscribers \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "preferences": {
      "channels": {"email": true},
      "language": "en",
      "allowedDays": ["monday", "tuesday", "wednesday"],
      "notificationWindow": {
        "start": "09:00",
        "end": "17:00"
      }
    }
  }'
```

### List Subscribers with Pagination

```bash
curl -X GET "http://localhost:3000/api/v1/subscribers?page=1&pageSize=20&sortBy=createdAt:desc" \
  -H "Authorization: Bearer <token>"
```

### Update a Group

```bash
curl -X PUT http://localhost:3000/api/v1/groups/group-123 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Group Name",
    "description": "New description"
  }'
```

