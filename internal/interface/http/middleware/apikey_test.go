package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// =============================================================================
// APIKeyAuthenticator Tests
// =============================================================================

func TestNewAPIKeyAuthenticator(t *testing.T) {
	tests := []struct {
		name      string
		validator KeyValidator
		opts      []APIKeyOption
		wantErr   error
	}{
		{
			name:      "valid validator",
			validator: &MapKeyValidator{Keys: map[string]*KeyInfo{}},
			wantErr:   nil,
		},
		{
			name:      "nil validator returns error",
			validator: nil,
			wantErr:   ErrValidatorRequired,
		},
		{
			name:      "with custom header name",
			validator: &MapKeyValidator{Keys: map[string]*KeyInfo{}},
			opts:      []APIKeyOption{WithHeaderName("Authorization")},
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewAPIKeyAuthenticator(tt.validator, tt.opts...)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("NewAPIKeyAuthenticator() error = nil, want %v", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("NewAPIKeyAuthenticator() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("NewAPIKeyAuthenticator() unexpected error = %v", err)
				return
			}

			if auth == nil {
				t.Error("NewAPIKeyAuthenticator() returned nil authenticator")
			}
		})
	}
}

func TestAPIKeyAuthenticator_Authenticate(t *testing.T) {
	validator := &MapKeyValidator{
		Keys: map[string]*KeyInfo{
			"valid-key": {
				ServiceID:   "svc-test",
				Roles:       []string{"service", "admin"},
				Permissions: []string{"read", "write"},
				Metadata:    map[string]string{"env": "prod"},
			},
			"minimal-key": {
				ServiceID: "svc-minimal",
			},
		},
	}

	auth, err := NewAPIKeyAuthenticator(validator)
	if err != nil {
		t.Fatalf("failed to create authenticator: %v", err)
	}

	tests := []struct {
		name            string
		headerName      string
		headerValue     string
		wantErr         error
		wantUserID      string
		wantRoles       []string
		wantPermissions []string
	}{
		{
			name:            "valid key with full info",
			headerName:      "X-API-Key",
			headerValue:     "valid-key",
			wantErr:         nil,
			wantUserID:      "svc-test",
			wantRoles:       []string{"service", "admin"},
			wantPermissions: []string{"read", "write"},
		},
		{
			name:        "minimal key gets default service role",
			headerName:  "X-API-Key",
			headerValue: "minimal-key",
			wantErr:     nil,
			wantUserID:  "svc-minimal",
			wantRoles:   []string{"service"},
		},
		{
			name:        "missing header returns ErrUnauthenticated",
			headerName:  "",
			headerValue: "",
			wantErr:     ErrUnauthenticated,
		},
		{
			name:        "empty header value returns ErrUnauthenticated",
			headerName:  "X-API-Key",
			headerValue: "",
			wantErr:     ErrUnauthenticated,
		},
		{
			name:        "invalid key returns ErrTokenInvalid",
			headerName:  "X-API-Key",
			headerValue: "wrong-key",
			wantErr:     ErrTokenInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.headerName != "" && tt.headerValue != "" {
				req.Header.Set(tt.headerName, tt.headerValue)
			}

			claims, err := auth.Authenticate(req)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Authenticate() error = nil, want %v", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("Authenticate() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Authenticate() unexpected error = %v", err)
				return
			}

			if claims.UserID != tt.wantUserID {
				t.Errorf("Authenticate() UserID = %v, want %v", claims.UserID, tt.wantUserID)
			}

			if len(claims.Roles) != len(tt.wantRoles) {
				t.Errorf("Authenticate() Roles = %v, want %v", claims.Roles, tt.wantRoles)
			} else {
				for i, role := range claims.Roles {
					if role != tt.wantRoles[i] {
						t.Errorf("Authenticate() Roles[%d] = %v, want %v", i, role, tt.wantRoles[i])
					}
				}
			}

			if tt.wantPermissions != nil {
				if len(claims.Permissions) != len(tt.wantPermissions) {
					t.Errorf("Authenticate() Permissions = %v, want %v", claims.Permissions, tt.wantPermissions)
				}
			}
		})
	}
}

func TestAPIKeyAuthenticator_CustomHeader(t *testing.T) {
	validator := &MapKeyValidator{
		Keys: map[string]*KeyInfo{
			"custom-key": {ServiceID: "custom-svc"},
		},
	}

	auth, err := NewAPIKeyAuthenticator(validator, WithHeaderName("X-Custom-Key"))
	if err != nil {
		t.Fatalf("failed to create authenticator: %v", err)
	}

	tests := []struct {
		name       string
		headerName string
		wantErr    error
	}{
		{
			name:       "custom header works",
			headerName: "X-Custom-Key",
			wantErr:    nil,
		},
		{
			name:       "default header does not work",
			headerName: "X-API-Key",
			wantErr:    ErrUnauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set(tt.headerName, "custom-key")

			_, err := auth.Authenticate(req)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Authenticate() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Authenticate() unexpected error = %v", err)
			}
		})
	}
}

func TestAPIKeyAuthenticator_MetadataMapping(t *testing.T) {
	validator := &MapKeyValidator{
		Keys: map[string]*KeyInfo{
			"meta-key": {
				ServiceID: "svc-meta",
				Metadata: map[string]string{
					"region": "us-east-1",
					"tier":   "premium",
				},
			},
		},
	}

	auth, err := NewAPIKeyAuthenticator(validator)
	if err != nil {
		t.Fatalf("failed to create authenticator: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "meta-key")

	claims, err := auth.Authenticate(req)
	if err != nil {
		t.Fatalf("Authenticate() unexpected error = %v", err)
	}

	if claims.Metadata == nil {
		t.Fatal("Authenticate() Metadata is nil")
	}

	if claims.Metadata["region"] != "us-east-1" {
		t.Errorf("Authenticate() Metadata[region] = %v, want us-east-1", claims.Metadata["region"])
	}

	if claims.Metadata["tier"] != "premium" {
		t.Errorf("Authenticate() Metadata[tier] = %v, want premium", claims.Metadata["tier"])
	}
}

// =============================================================================
// EnvKeyValidator Tests
// =============================================================================

func TestEnvKeyValidator_Validate(t *testing.T) {
	// Setup: Set environment variable for testing
	const envVar = "TEST_API_KEYS"
	originalValue := os.Getenv(envVar)
	defer func() {
		if originalValue == "" {
			os.Unsetenv(envVar)
		} else {
			os.Setenv(envVar, originalValue)
		}
	}()

	tests := []struct {
		name        string
		envValue    string
		key         string
		wantErr     error
		wantService string
	}{
		{
			name:        "valid key",
			envValue:    "abc123:svc-payments,xyz789:svc-inventory",
			key:         "abc123",
			wantErr:     nil,
			wantService: "svc-payments",
		},
		{
			name:        "second valid key",
			envValue:    "abc123:svc-payments,xyz789:svc-inventory",
			key:         "xyz789",
			wantErr:     nil,
			wantService: "svc-inventory",
		},
		{
			name:     "invalid key",
			envValue: "abc123:svc-payments",
			key:      "wrong-key",
			wantErr:  ErrTokenInvalid,
		},
		{
			name:     "empty env var",
			envValue: "",
			key:      "any-key",
			wantErr:  ErrTokenInvalid,
		},
		{
			name:        "whitespace handling",
			envValue:    " abc123 : svc-payments , xyz789 : svc-inventory ",
			key:         "abc123",
			wantErr:     nil,
			wantService: "svc-payments",
		},
		{
			name:     "malformed entry is skipped",
			envValue: "abc123:svc-payments,invalid-entry,xyz789:svc-inventory",
			key:      "invalid-entry",
			wantErr:  ErrTokenInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(envVar, tt.envValue)
			validator := NewEnvKeyValidator(envVar)

			keyInfo, err := validator.Validate(context.Background(), tt.key)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Validate() error = nil, want %v", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Validate() unexpected error = %v", err)
				return
			}

			if keyInfo.ServiceID != tt.wantService {
				t.Errorf("Validate() ServiceID = %v, want %v", keyInfo.ServiceID, tt.wantService)
			}

			// EnvKeyValidator always sets "service" role
			if len(keyInfo.Roles) != 1 || keyInfo.Roles[0] != "service" {
				t.Errorf("Validate() Roles = %v, want [service]", keyInfo.Roles)
			}
		})
	}
}

func TestEnvKeyValidator_UnsetEnvVar(t *testing.T) {
	// Test with env var that doesn't exist
	validator := NewEnvKeyValidator("NONEXISTENT_API_KEYS_12345")

	_, err := validator.Validate(context.Background(), "any-key")
	if err != ErrTokenInvalid {
		t.Errorf("Validate() error = %v, want %v", err, ErrTokenInvalid)
	}
}

// =============================================================================
// MapKeyValidator Tests
// =============================================================================

func TestMapKeyValidator_Validate(t *testing.T) {
	validator := &MapKeyValidator{
		Keys: map[string]*KeyInfo{
			"key1": {
				ServiceID:   "svc1",
				Roles:       []string{"service"},
				Permissions: []string{"read"},
			},
			"key2": {
				ServiceID: "svc2",
				Roles:     []string{"admin"},
			},
		},
	}

	tests := []struct {
		name        string
		key         string
		wantErr     error
		wantService string
	}{
		{
			name:        "valid key1",
			key:         "key1",
			wantErr:     nil,
			wantService: "svc1",
		},
		{
			name:        "valid key2",
			key:         "key2",
			wantErr:     nil,
			wantService: "svc2",
		},
		{
			name:    "invalid key",
			key:     "nonexistent",
			wantErr: ErrTokenInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyInfo, err := validator.Validate(context.Background(), tt.key)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Validate() unexpected error = %v", err)
				return
			}

			if keyInfo.ServiceID != tt.wantService {
				t.Errorf("Validate() ServiceID = %v, want %v", keyInfo.ServiceID, tt.wantService)
			}
		})
	}
}

func TestMapKeyValidator_NilMap(t *testing.T) {
	validator := &MapKeyValidator{}

	_, err := validator.Validate(context.Background(), "any-key")
	if err != ErrTokenInvalid {
		t.Errorf("Validate() error = %v, want %v", err, ErrTokenInvalid)
	}
}

// =============================================================================
// Nil KeyInfo Edge Case Test
// =============================================================================

// nilKeyInfoValidator is a mock that returns nil KeyInfo without error
type nilKeyInfoValidator struct{}

func (v *nilKeyInfoValidator) Validate(ctx context.Context, key string) (*KeyInfo, error) {
	return nil, nil // Returns nil KeyInfo without error - edge case
}

func TestAPIKeyAuthenticator_NilKeyInfoReturnsError(t *testing.T) {
	// Arrange: Create validator that returns nil KeyInfo without error
	validator := &nilKeyInfoValidator{}
	auth, err := NewAPIKeyAuthenticator(validator)
	if err != nil {
		t.Fatalf("failed to create authenticator: %v", err)
	}

	// Act: Authenticate with any key
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "any-key")
	_, err = auth.Authenticate(req)

	// Assert: Should return ErrTokenInvalid, not panic
	if err != ErrTokenInvalid {
		t.Errorf("Authenticate() error = %v, want %v", err, ErrTokenInvalid)
	}
}

// =============================================================================
// EnvKeyValidator.KeyCount Tests
// =============================================================================

func TestEnvKeyValidator_KeyCount(t *testing.T) {
	const envVar = "TEST_KEYCOUNT_API_KEYS"
	originalValue := os.Getenv(envVar)
	defer func() {
		if originalValue == "" {
			os.Unsetenv(envVar)
		} else {
			os.Setenv(envVar, originalValue)
		}
	}()

	tests := []struct {
		name      string
		envValue  string
		wantCount int
	}{
		{
			name:      "empty env var",
			envValue:  "",
			wantCount: 0,
		},
		{
			name:      "single key",
			envValue:  "key1:svc1",
			wantCount: 1,
		},
		{
			name:      "multiple keys",
			envValue:  "key1:svc1,key2:svc2,key3:svc3",
			wantCount: 3,
		},
		{
			name:      "skips malformed entries",
			envValue:  "key1:svc1,invalid,key2:svc2",
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(envVar, tt.envValue)
			validator := NewEnvKeyValidator(envVar)

			if got := validator.KeyCount(); got != tt.wantCount {
				t.Errorf("KeyCount() = %v, want %v", got, tt.wantCount)
			}
		})
	}
}

// =============================================================================
// Integration with AuthMiddleware
// =============================================================================

func TestAPIKeyAuthenticator_WithAuthMiddleware(t *testing.T) {
	validator := &MapKeyValidator{
		Keys: map[string]*KeyInfo{
			"service-key": {ServiceID: "svc-test"},
		},
	}

	auth, err := NewAPIKeyAuthenticator(validator)
	if err != nil {
		t.Fatalf("failed to create authenticator: %v", err)
	}

	// Handler that extracts claims from context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := FromContext(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(claims.UserID))
	})

	middleware := AuthMiddleware(auth)
	wrapped := middleware(handler)

	tests := []struct {
		name           string
		apiKey         string
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "valid key passes through middleware",
			apiKey:         "service-key",
			wantStatusCode: http.StatusOK,
			wantBody:       "svc-test",
		},
		{
			name:           "invalid key returns 401",
			apiKey:         "wrong-key",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "missing key returns 401",
			apiKey:         "",
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Errorf("StatusCode = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("Body = %v, want %v", rec.Body.String(), tt.wantBody)
			}
		})
	}
}
