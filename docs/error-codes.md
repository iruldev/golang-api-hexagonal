# Error Code Taxonomy

This document defines the error code taxonomy used in API error responses. All error codes follow the RFC 7807 Problem Details format.

## Error Code Format

Error codes follow the format `{CATEGORY}-{NNN}` where:
- **CATEGORY**: A short identifier for the error domain (e.g., `AUTH`, `VAL`)
- **NNN**: A 3-digit number within the category's reserved range

### Example

```json
{
  "type": "https://api.example.com/problems/validation-error",
  "title": "Validation Error",
  "status": 400,
  "detail": "The email address format is invalid",
  "code": "VAL-003",
  "request_id": "req_abc123",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736"
}
```

## Error Categories

| Category | Prefix | Range     | HTTP Status | Description                        |
|----------|--------|-----------|-------------|------------------------------------|
| AUTH     | AUTH   | 001-099   | 401         | Authentication (token, credentials)|
| AUTHZ    | AUTHZ  | 001-099   | 403         | Authorization (permissions)        |
| VAL      | VAL    | 001-199   | 400         | Validation (input errors)          |
| USR      | USR    | 001-099   | 400/404/409 | User domain specific               |
| DB       | DB     | 001-099   | 500/503     | Database operations                |
| SYS      | SYS    | 001-099   | 500/503     | System/infrastructure              |
| RATE     | RATE   | 001-099   | 429         | Rate limiting                      |
| RES      | RES    | 001-099   | 503         | Resilience (circuit breaker, etc.) |

---

## AUTH - Authentication Errors (HTTP 401)

Authentication errors occur when the provided credentials or token is invalid, expired, or missing.

| Code       | Title                | Description                              |
|------------|----------------------|------------------------------------------|
| AUTH-001   | Token Expired        | The authentication token has expired     |
| AUTH-002   | Invalid Token        | The authentication token is invalid      |
| AUTH-003   | Missing Token        | No authentication token was provided     |
| AUTH-004   | Invalid Credentials  | The username or password is incorrect    |

### Resolution

- For `AUTH-001`: Refresh your token or re-authenticate
- For `AUTH-002`: Ensure token format is correct and hasn't been tampered with
- For `AUTH-003`: Include `Authorization: Bearer <token>` header
- For `AUTH-004`: Verify username and password are correct

---

## AUTHZ - Authorization Errors (HTTP 403)

Authorization errors occur when the authenticated user lacks permission to perform the requested action.

| Code       | Title                      | Description                                    |
|------------|----------------------------|------------------------------------------------|
| AUTHZ-001  | Forbidden                  | Access to this resource is forbidden           |
| AUTHZ-002  | Insufficient Permissions   | You do not have permission to perform this action |

### Resolution

- Contact your administrator to request the necessary permissions
- Verify you're accessing the correct resource

---

## VAL - Validation Errors (HTTP 400)

Validation errors occur when the request body or parameters fail validation rules.

| Code       | Title                | Description                                    |
|------------|----------------------|------------------------------------------------|
| VAL-001    | Required Field Missing | A required field is missing                   |
| VAL-002    | Invalid Format       | The value has an invalid format                |
| VAL-003    | Invalid Email        | The email address format is invalid            |
| VAL-004    | Value Too Short      | The value is shorter than the minimum length   |
| VAL-005    | Value Too Long       | The value exceeds the maximum length           |
| VAL-006    | Value Out of Range   | The value is outside the allowed range         |
| VAL-007    | Invalid Type         | The value has an invalid type                  |
| VAL-008    | Invalid JSON         | The request body is not valid JSON             |
| VAL-009    | Request Too Large    | The request body exceeds the size limit (HTTP 413) |
| VAL-010    | Invalid UUID         | The UUID format is invalid                     |
| VAL-100    | Idempotency Conflict | The idempotency key already exists with different request data (HTTP 409) |

### Reserved Ranges

- **VAL-001 to VAL-099**: General validation errors
- **VAL-100 to VAL-199**: Feature-specific validation errors (e.g., idempotency)

### Resolution

- Check the `errors` array in the response for field-specific details
- Correct the invalid fields and retry the request

---

## USR - User Domain Errors

User domain errors are specific to user-related operations.

| Code       | Title                | HTTP Status | Description                            |
|------------|----------------------|-------------|----------------------------------------|
| USR-001    | User Not Found       | 404         | The requested user was not found       |
| USR-002    | Email Already Exists | 409         | The email address is already registered |
| USR-003    | Invalid User Field   | 400         | A user field has an invalid value      |

### Resolution

- For `USR-001`: Verify the user ID is correct
- For `USR-002`: Use a different email address or recover the existing account
- For `USR-003`: Check the specific field validation error in the `errors` array

---

## DB - Database Errors (HTTP 500/503)

Database errors indicate issues with database operations. These are typically transient and may be resolved by retrying after a delay.

| Code       | Title                     | HTTP Status | Description                            |
|------------|---------------------------|-------------|----------------------------------------|
| DB-001     | Database Connection Failed | 503         | Unable to connect to the database      |
| DB-002     | Database Query Failed     | 500         | An error occurred while executing the database query |
| DB-003     | Transaction Failed        | 500         | The database transaction failed        |

### Resolution

- Retry the request after a short delay
- If the error persists, contact support

---

## SYS - System Errors (HTTP 500/503)

System errors indicate infrastructure or configuration issues.

| Code       | Title                | HTTP Status | Description                            |
|------------|----------------------|-------------|----------------------------------------|
| SYS-001    | Internal Server Error | 500         | An internal error occurred             |
| SYS-002    | Service Unavailable  | 503         | The service is temporarily unavailable |
| SYS-003    | Configuration Error  | 500         | A configuration error occurred         |

### Resolution

- Retry the request after a short delay
- If the error persists, contact support with the `request_id` and `trace_id`

---

## RATE - Rate Limit Errors (HTTP 429)

Rate limit errors occur when you exceed the allowed request rate.

| Code       | Title               | Description                            |
|------------|---------------------|----------------------------------------|
| RATE-001   | Rate Limit Exceeded | You have exceeded the rate limit       |

### Response Headers

When rate limited, the following headers are included:

- `X-RateLimit-Limit`: Maximum requests allowed in the window
- `X-RateLimit-Remaining`: Requests remaining in the current window
- `X-RateLimit-Reset`: Unix timestamp when the limit resets
- `Retry-After`: Seconds until you can retry

### Resolution

- Wait for the `Retry-After` duration before retrying
- Implement exponential backoff in your client

---

## RES - Resilience Errors (HTTP 503)

Resilience errors occur when circuit breakers, bulkheads, or retry mechanisms are activated to protect the system.

| Code       | Title                | Description                                              |
|------------|----------------------|----------------------------------------------------------|
| RES-001    | Circuit Breaker Open | The service is temporarily unavailable due to circuit breaker activation |
| RES-002    | Service Overloaded   | The service is currently overloaded (bulkhead full)      |
| RES-003    | Operation Timeout    | The operation timed out                                  |
| RES-004    | Retry Limit Exceeded | The operation failed after maximum retry attempts        |

### Resolution

- These errors indicate the system is under stress
- Retry with exponential backoff
- If errors persist, the underlying dependency may be unavailable

---

## Legacy Error Codes

> ⚠️ **Deprecated**: The following `ERR_*` format codes are deprecated. New code should use the new `{CATEGORY}-{NNN}` format.

For backward compatibility, legacy codes are automatically mapped to the new taxonomy:

| Legacy Code           | New Code    | Description                    |
|----------------------|-------------|--------------------------------|
| ERR_USER_NOT_FOUND   | USR-001     | User not found                 |
| ERR_USER_EMAIL_EXISTS | USR-002    | Email already exists           |
| ERR_VALIDATION       | VAL-001     | Validation error               |
| ERR_UNAUTHORIZED     | AUTH-001    | Authentication required        |
| ERR_FORBIDDEN        | AUTHZ-001   | Access forbidden               |
| ERR_INTERNAL         | SYS-001     | Internal server error          |
| ERR_REQUEST_TOO_LARGE | VAL-009    | Request too large              |
| ERR_RATE_LIMIT_EXCEEDED | RATE-001 | Rate limit exceeded            |

### Migration Guide

1. Update error handling code to expect new format codes
2. The API will return new format codes for new error conditions
3. Legacy codes are still supported for backward compatibility
4. Check for both formats during the transition period

---

## Best Practices

### For API Consumers

1. **Always check the `code` field** for programmatic error handling
2. **Use `request_id` and `trace_id`** when reporting issues
3. **Check the `errors` array** for field-level validation details
4. **Implement retry logic** for 5xx errors with exponential backoff
5. **Respect rate limits** by checking response headers

### For API Developers

1. **Use predefined error codes** from the registry
2. **Never expose internal details** in error messages
3. **Always include `request_id` and `trace_id`** for debugging
4. **Log errors with context** before sending the response
5. **Keep error codes stable** - never change the meaning of existing codes

---

## References

- [RFC 7807: Problem Details for HTTP APIs](https://www.rfc-editor.org/rfc/rfc7807.html)
- [Error Code Implementation](../internal/transport/http/contract/codes.go)
- [Problem Details Implementation](../internal/transport/http/contract/problem.go)
