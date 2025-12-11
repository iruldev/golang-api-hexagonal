# Story 5.3: Return 503 on Dependency Failure

Status: done

## Story

As a SRE,
I want 503 when database is down,
So that load balancer routes traffic elsewhere.

## Acceptance Criteria

### AC1: Readiness returns 503 when DB down
**Given** database connection is lost
**When** I request `GET /readyz`
**Then** response status is 503
**And** body shows `{"database": "unavailable"}`

---

## Tasks / Subtasks

- [x] **All tasks completed in Story 4.7**

---

## Dev Notes

> **NOTE:** This story was already completed as part of **Story 4.7: Add Database Readiness Check**!

See [Story 4.7](file:///docs/sprint-artifacts/4-7-add-database-readiness-check.md) for implementation details.

### Current Implementation

**Endpoint:** `GET /readyz`  
**Handler:** `ReadyzHandler.ServeHTTP`  
**Response (DB down):**
```json
{"success": false, "error": {"code": "ERR_SERVICE_UNAVAILABLE", "message": "database unavailable"}}
```
