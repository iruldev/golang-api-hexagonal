# API Contracts

**Project:** golang-api-hexagonal  
**Version:** v1  
**Base Path:** `/api/v1`  

## Overview

This document describes the HTTP API contracts for the golang-api-hexagonal service.

## Authentication

Authentication is **optional** and controlled by the `JWT_ENABLED` environment variable.

When enabled:
- All `/api/v1/*` endpoints require a valid JWT token
- Token must be passed in the `Authorization` header: `Bearer <token>`
- Token must use HS256 algorithm and be signed with `JWT_SECRET`

## Common Response Format

### Success Response

```json
{
  "data": { ... }
}
```

### Paginated Response

```json
{
  "data": [ ... ],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "total": 100,
    "totalPages": 5
  }
}
```

### Error Response (RFC 7807)

```json
{
  "type": "https://api.example.com/problems/validation-error",
  "title": "Validation Error",
  "status": 400,
  "detail": "The email field is required",
  "instance": "/api/v1/users"
}
```

---

## Health Endpoints

### GET /health

**Description:** Liveness probe - checks if the service is running.

**Authentication:** None

**Response:**

```json
{
  "data": {
    "status": "ok"
  }
}
```

**Status Codes:**
- `200 OK` - Service is alive

---

### GET /ready

**Description:** Readiness probe - checks if the service can handle requests.

**Authentication:** None

**Response:**

```json
{
  "data": {
    "status": "ready",
    "checks": {
      "database": "ok"
    }
  }
}
```

**Status Codes:**
- `200 OK` - Service is ready
- `503 Service Unavailable` - Service is not ready (e.g., database down)

---

## Metrics Endpoint

### GET /metrics

**Description:** Prometheus metrics in text format.

**Authentication:** None

**Response:** Prometheus text format

**Metrics Available:**
- `http_requests_total{method, route, status}` - Request counter
- `http_request_duration_seconds{method, route}` - Request latency histogram
- `go_goroutines` - Number of goroutines
- `go_memstats_*` - Memory statistics

---

## Users Endpoints

### POST /api/v1/users

**Description:** Create a new user.

**Authentication:** Required (if JWT_ENABLED)

**Request Body:**

```json
{
  "email": "john@example.com",
  "firstName": "John",
  "lastName": "Doe"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `email` | string | ✅ | Valid email format |
| `firstName` | string | ✅ | Non-empty |
| `lastName` | string | ✅ | Non-empty |

**Response (201 Created):**

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "john@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "createdAt": "2024-12-24T10:00:00Z",
    "updatedAt": "2024-12-24T10:00:00Z"
  }
}
```

**Status Codes:**
- `201 Created` - User created successfully
- `400 Bad Request` - Validation error
- `401 Unauthorized` - Missing or invalid JWT token
- `409 Conflict` - Email already exists
- `429 Too Many Requests` - Rate limit exceeded

**Side Effects:**
- Creates audit event: `user.created`

---

### GET /api/v1/users/{id}

**Description:** Get a user by ID.

**Authentication:** Required (if JWT_ENABLED)

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | User ID |

**Response (200 OK):**

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "john@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "createdAt": "2024-12-24T10:00:00Z",
    "updatedAt": "2024-12-24T10:00:00Z"
  }
}
```

**Status Codes:**
- `200 OK` - User found
- `400 Bad Request` - Invalid UUID format
- `401 Unauthorized` - Missing or invalid JWT token
- `404 Not Found` - User not found

---

### GET /api/v1/users

**Description:** List users with pagination.

**Authentication:** Required (if JWT_ENABLED)

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | int | 1 | Page number (1-indexed) |
| `pageSize` | int | 20 | Items per page (max 100) |

**Response (200 OK):**

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "john@example.com",
      "firstName": "John",
      "lastName": "Doe",
      "createdAt": "2024-12-24T10:00:00Z",
      "updatedAt": "2024-12-24T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "total": 1,
    "totalPages": 1
  }
}
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid pagination parameters
- `401 Unauthorized` - Missing or invalid JWT token

---

## Request Headers

### Standard Headers

| Header | Required | Description |
|--------|----------|-------------|
| `Content-Type` | Yes (POST/PUT) | `application/json` |
| `Authorization` | Conditional | `Bearer <token>` (if JWT_ENABLED) |
| `X-Request-ID` | No | Client-provided request ID (UUID format) |

### Response Headers

| Header | Description |
|--------|-------------|
| `X-Request-ID` | Request correlation ID |
| `X-Content-Type-Options` | `nosniff` |
| `X-Frame-Options` | `DENY` |
| `Content-Security-Policy` | `default-src 'none'` |

---

## Rate Limiting

Rate limiting is applied to `/api/v1/*` endpoints.

**Configuration:**
- `RATE_LIMIT_RPS` - Requests per second (default: 100)
- `TRUST_PROXY` - Trust X-Forwarded-For header (default: false)

**Response when rate limited (429):**

```json
{
  "type": "https://api.example.com/problems/rate-limit-exceeded",
  "title": "Rate Limit Exceeded",
  "status": 429,
  "detail": "You have exceeded the rate limit. Please try again later."
}
```

---

## Error Types

| Type Suffix | HTTP Status | Description |
|-------------|-------------|-------------|
| `validation-error` | 400 | Request validation failed |
| `invalid-request` | 400 | Malformed request |
| `unauthorized` | 401 | Authentication required |
| `forbidden` | 403 | Permission denied |
| `not-found` | 404 | Resource not found |
| `conflict` | 409 | Resource conflict (e.g., duplicate) |
| `rate-limit-exceeded` | 429 | Rate limit exceeded |
| `internal-error` | 500 | Internal server error |
