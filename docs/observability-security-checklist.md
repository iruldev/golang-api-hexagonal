# Security Checklist for Observability Stories

This checklist should be applied when creating or reviewing stories related to metrics, logging, or tracing.

---

## Metrics Security Checklist

### Label Safety (Cardinality & PII)

- [ ] **No user identifiers in labels** - UUIDs, user IDs, emails must not appear
- [ ] **Route patterns use placeholders** - Use `/users/{id}` not `/users/123`
- [ ] **Unmatched routes use static label** - Use `"unmatched"` for 404s
- [ ] **HTTP methods are whitelisted** - Non-standard methods map to `"OTHER"`
- [ ] **No query parameters in labels** - Strip query strings from paths
- [ ] **No JWT tokens or secrets** - Never log or label tokens

### Cardinality Protection

- [ ] **Bounded label values** - All labels have finite, known values
- [ ] **Fallback for unknown values** - Use static fallback like `"unknown"`
- [ ] **Tested with malicious input** - Verify random paths don't create new series

### Integration Test Requirements

- [ ] **Scrape /metrics in test** - Validate actual output
- [ ] **Assert no forbidden patterns** - Check UUIDs, emails, tokens
- [ ] **Assert route pattern usage** - Verify placeholders are used

---

## Logging Security Checklist

### PII Redaction

- [ ] **Emails redacted** - Use `[REDACTED]` or partial masking
- [ ] **Passwords never logged** - Filter at struct level
- [ ] **JWT tokens masked** - Show only first/last few characters if needed
- [ ] **IP addresses reviewed** - Consider GDPR implications

### Context Propagation

- [ ] **request_id injected** - Use `LoggerFromContext`
- [ ] **trace_id injected** - Include when tracing enabled
- [ ] **span_id injected** - Include for distributed tracing
- [ ] **Zero IDs filtered** - Don't log `00000000...`

### Sensitive Detection Test

- [ ] **Unit test for redaction** - Verify PII is masked
- [ ] **Log output assertions** - Check actual log entries

---

## Tracing Security Checklist

### Span Attribute Safety

- [ ] **No PII in span names** - Use operation names not user data
- [ ] **No secrets in attributes** - Never attach tokens
- [ ] **Request body not in spans** - Unless explicitly redacted

---

## RFC7807 Error Response Checklist

### Extension Fields

- [ ] **request_id included** - Always when available
- [ ] **trace_id included** - When tracing enabled
- [ ] **Zero IDs filtered** - Omit empty/zero values

### Fallback Response

- [ ] **Fallback JSON includes IDs** - Emergency 500 response too
- [ ] **No injection vulnerabilities** - Marshal IDs securely

### Test Coverage

- [ ] **Test ID presence** - When context has IDs
- [ ] **Test graceful degradation** - When context empty
- [ ] **Test both IDs together** - request_id AND trace_id

---

## Story Template Addition

When creating observability stories, add this section to the story file:

```markdown
## Security Checklist

- [ ] Reviewed against [metrics-audit-checklist.md](../docs/metrics-audit-checklist.md)
- [ ] Reviewed against [observability-security-checklist.md](../docs/observability-security-checklist.md)
- [ ] No PII in labels/logs/spans
- [ ] Cardinality bounded
- [ ] Test coverage for security assertions
```

---

## References

- [Prometheus Best Practices: Labels](https://prometheus.io/docs/practices/naming/#labels)
- [OWASP Logging Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html)
- [GDPR and Logging](https://gdpr.eu/article-25-data-protection-by-design/)
