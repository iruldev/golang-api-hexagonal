// Package contract provides HTTP transport layer contracts including
// RFC 7807 Problem Details for machine-readable error responses.
//
// This package implements the transport layer's error handling contracts,
// providing a standardized way to communicate errors to API clients following
// RFC 7807 (Problem Details for HTTP APIs).
//
// # RFC 7807 Problem Details
//
// All error responses from this API use the RFC 7807 format with the following structure:
//
//	{
//	    "type": "https://api.example.com/problems/validation-error",
//	    "title": "Validation Error",
//	    "status": 400,
//	    "detail": "Email format is invalid",
//	    "code": "VAL-002",
//	    "request_id": "req_abc123",
//	    "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736"
//	}
//
// # Error Code Taxonomy
//
// Error codes follow the format {CATEGORY}-{NUMBER} with the following categories:
//
//	| Category | Code Prefix | HTTP Status | Description                        |
//	|----------|-------------|-------------|------------------------------------|
//	| AUTH     | AUTH-       | 401         | Authentication (token issues)      |
//	| AUTHZ    | AUTHZ-      | 403         | Authorization (permission denied)  |
//	| VAL      | VAL-        | 400         | Validation (input errors)          |
//	| USR      | USR-        | 4xx         | User domain (not found, conflict)  |
//	| DB       | DB-         | 500         | Database (connection, query)       |
//	| SYS      | SYS-        | 500/503     | System (internal, unavailable)     |
//	| RATE     | RATE-       | 429         | Rate limiting                      |
//	| RES      | RES-        | 503         | Resilience (circuit breaker, etc.) |
//
// # Content-Type
//
// All error responses use Content-Type: application/problem+json as required
// by RFC 7807.
//
// # Usage
//
// Creating a problem response from an application error:
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//	    user, err := service.GetUser(ctx, id)
//	    if err != nil {
//	        contract.WriteProblemJSON(w, r, err)
//	        return
//	    }
//	    // success response...
//	}
//
// Creating a validation error response:
//
//	errors := []contract.ValidationError{
//	    {Field: "email", Message: "must be a valid email address"},
//	}
//	contract.WriteValidationError(w, r, errors)
//
// Creating a problem with custom fields:
//
//	problem := contract.NewProblem(http.StatusBadRequest, "Invalid Input", "The request body is malformed")
//	problem.Code = contract.CodeValInvalidFormat
//	contract.WriteProblem(w, problem)
//
// # Thread Safety
//
// Problem instances are not safe for concurrent modification.
// Create a new Problem for each error response.
//
// # Legacy Code Support
//
// The registry supports legacy ERR_* format codes from domain/app layers
// via the TranslateLegacyCode function for backward compatibility.
//
// # See Also
//
//   - RFC 7807: https://tools.ietf.org/html/rfc7807
//   - ADR-004: RFC 7807 Error Handling
package contract
