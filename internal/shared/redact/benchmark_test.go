package redact_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/shared/redact"
)

func BenchmarkRedactAndMarshal(b *testing.B) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	// 1. Small Payload
	smallPayload := map[string]any{
		"event": "user.login",
		"email": "user@example.com",
		"id":    "u-123",
	}

	// 2. Medium Payload (Nested)
	mediumPayload := map[string]any{
		"event":     "payment.processed",
		"timestamp": "2024-01-01T12:00:00Z",
		"user": map[string]any{
			"id":    "u-123",
			"email": "user@example.com",
			"name":  "John Doe",
		},
		"transaction": map[string]any{
			"id":     "tx-999",
			"amount": 100.50,
			"card":   map[string]any{"number": "4111-1111-1111-1111", "cvv": "123"},
		},
	}

	// 3. Large Payload (Deep)
	largePayload := makeDeepMap(50)

	b.Run("Small_Map", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = redact.RedactAndMarshal(r, smallPayload)
		}
	})

	b.Run("Medium_Map", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = redact.RedactAndMarshal(r, mediumPayload)
		}
	})

	b.Run("Large_Map", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = redact.RedactAndMarshal(r, largePayload)
		}
	})

	// 4. Struct Input
	type UserStruct struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	structPayload := UserStruct{ID: "u-1", Email: "me@test.com", Password: "secret"}

	b.Run("Struct_Input", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = redact.RedactAndMarshal(r, structPayload)
		}
	})

	// 5. JSON Input
	jsonBytes, _ := json.Marshal(mediumPayload)
	b.Run("JSON_Input", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = redact.RedactAndMarshal(r, jsonBytes)
		}
	})
}

func makeDeepMap(depth int) map[string]any {
	root := make(map[string]any)
	curr := root
	for i := 0; i < depth; i++ {
		next := make(map[string]any)
		curr[fmt.Sprintf("level-%d", i)] = next
		curr["data"] = "some data"
		if i%5 == 0 {
			curr["password"] = "secret"
		}
		curr = next
	}
	return root
}
