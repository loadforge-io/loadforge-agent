package extractor

import (
	"testing"
)

func TestNew(t *testing.T) {
	extractor := New()
	if extractor == nil {
		t.Fatal("New() returned nil")
	}
}

// ============================================================================
// Extract() Tests - String values
// ============================================================================

func TestExtract_StringValue(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"user": {"id": "12345", "name": "John Doe"}}`)

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "extract user id",
			path:     "user.id",
			expected: "12345",
		},
		{
			name:     "extract user name",
			path:     "user.name",
			expected: "John Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractor.Extract(jsonData, tt.path)
			if err != nil {
				t.Fatalf("Extract() failed: %v", err)
			}

			// Type assertion
			strResult, ok := result.(string)
			if !ok {
				t.Fatalf("Expected string type, got %T", result)
			}

			if strResult != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, strResult)
			}
		})
	}
}

// ============================================================================
// Extract() Tests - Numeric values
// ============================================================================

func TestExtract_IntegerValue(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"count": 42, "age": 25}`)

	tests := []struct {
		name     string
		path     string
		expected int64
	}{
		{
			name:     "extract count",
			path:     "count",
			expected: 42,
		},
		{
			name:     "extract age",
			path:     "age",
			expected: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractor.Extract(jsonData, tt.path)
			if err != nil {
				t.Fatalf("Extract() failed: %v", err)
			}

			// gjson returns float64 for all numbers
			floatResult, ok := result.(float64)
			if !ok {
				t.Fatalf("Expected float64 type, got %T", result)
			}

			if int64(floatResult) != tt.expected {
				t.Errorf("Expected %d, got %v", tt.expected, floatResult)
			}
		})
	}
}

func TestExtract_FloatValue(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"price": 19.99, "rating": 4.5}`)

	tests := []struct {
		name     string
		path     string
		expected float64
	}{
		{
			name:     "extract price",
			path:     "price",
			expected: 19.99,
		},
		{
			name:     "extract rating",
			path:     "rating",
			expected: 4.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractor.Extract(jsonData, tt.path)
			if err != nil {
				t.Fatalf("Extract() failed: %v", err)
			}

			floatResult, ok := result.(float64)
			if !ok {
				t.Fatalf("Expected float64 type, got %T", result)
			}

			if floatResult != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, floatResult)
			}
		})
	}
}

// ============================================================================
// Extract() Tests - Boolean values
// ============================================================================

func TestExtract_BooleanValue(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"active": true, "verified": false}`)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "extract active (true)",
			path:     "active",
			expected: true,
		},
		{
			name:     "extract verified (false)",
			path:     "verified",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractor.Extract(jsonData, tt.path)
			if err != nil {
				t.Fatalf("Extract() failed: %v", err)
			}

			boolResult, ok := result.(bool)
			if !ok {
				t.Fatalf("Expected bool type, got %T", result)
			}

			if boolResult != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, boolResult)
			}
		})
	}
}

// ============================================================================
// Extract() Tests - Null values
// ============================================================================

func TestExtract_NullValue(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"user": {"id": "123", "email": null}}`)

	result, err := extractor.Extract(jsonData, "user.email")
	if err != nil {
		t.Fatalf("Extract() failed: %v", err)
	}

	// gjson returns nil for null values
	if result != nil {
		t.Errorf("Expected nil for null value, got %v (%T)", result, result)
	}
}

// ============================================================================
// Extract() Tests - Array and Object values
// ============================================================================

func TestExtract_ArrayValue(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"tags": ["go", "testing", "api"]}`)

	result, err := extractor.Extract(jsonData, "tags")
	if err != nil {
		t.Fatalf("Extract() failed: %v", err)
	}

	// gjson returns []interface{} for arrays
	arrayResult, ok := result.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{} type, got %T", result)
	}

	if len(arrayResult) != 3 {
		t.Errorf("Expected array length 3, got %d", len(arrayResult))
	}

	// Check first element
	if arrayResult[0].(string) != "go" {
		t.Errorf("Expected first element 'go', got '%v'", arrayResult[0])
	}
}

func TestExtract_ObjectValue(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"user": {"id": "123", "name": "John"}}`)

	result, err := extractor.Extract(jsonData, "user")
	if err != nil {
		t.Fatalf("Extract() failed: %v", err)
	}

	// gjson returns map[string]interface{} for objects
	objResult, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{} type, got %T", result)
	}

	if objResult["id"].(string) != "123" {
		t.Errorf("Expected id '123', got '%v'", objResult["id"])
	}

	if objResult["name"].(string) != "John" {
		t.Errorf("Expected name 'John', got '%v'", objResult["name"])
	}
}

// ============================================================================
// Extract() Tests - Nested paths
// ============================================================================

func TestExtract_NestedFields(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{
		"response": {
			"data": {
				"user": {
					"profile": {
						"email": "test@example.com"
					}
				}
			}
		}
	}`)

	result, err := extractor.Extract(jsonData, "response.data.user.profile.email")
	if err != nil {
		t.Fatalf("Extract() failed: %v", err)
	}

	email, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string type, got %T", result)
	}

	expected := "test@example.com"
	if email != expected {
		t.Errorf("Expected '%s', got '%s'", expected, email)
	}
}

// ============================================================================
// Extract() Tests - Array access
// ============================================================================

func TestExtract_ArrayAccess(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"users": [{"id": "1", "name": "Alice"}, {"id": "2", "name": "Bob"}]}`)

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "first user id",
			path:     "users.0.id",
			expected: "1",
		},
		{
			name:     "second user name",
			path:     "users.1.name",
			expected: "Bob",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractor.Extract(jsonData, tt.path)
			if err != nil {
				t.Fatalf("Extract() failed: %v", err)
			}

			strResult, ok := result.(string)
			if !ok {
				t.Fatalf("Expected string type, got %T", result)
			}

			if strResult != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, strResult)
			}
		})
	}
}

func TestExtract_ArrayWildcard(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{
		"users": [
			{"name": "Alice", "age": 30},
			{"name": "Bob", "age": 25},
			{"name": "Charlie", "age": 35}
		]
	}`)

	result, err := extractor.Extract(jsonData, "users.#.name")
	if err != nil {
		t.Fatalf("Extract() failed: %v", err)
	}

	// gjson returns []interface{} for wildcard results
	names, ok := result.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{} type, got %T", result)
	}

	if len(names) != 3 {
		t.Fatalf("Expected 3 names, got %d", len(names))
	}

	expectedNames := []string{"Alice", "Bob", "Charlie"}
	for i, expectedName := range expectedNames {
		if names[i].(string) != expectedName {
			t.Errorf("Expected name[%d] '%s', got '%v'", i, expectedName, names[i])
		}
	}
}

func TestExtract_ArrayCount(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"users": [{"id": "1"}, {"id": "2"}, {"id": "3"}]}`)

	result, err := extractor.Extract(jsonData, "users.#")
	if err != nil {
		t.Fatalf("Extract() failed: %v", err)
	}

	// Array count returns float64
	count, ok := result.(float64)
	if !ok {
		t.Fatalf("Expected float64 type, got %T", result)
	}

	if int(count) != 3 {
		t.Errorf("Expected count 3, got %v", count)
	}
}

// ============================================================================
// Extract() Tests - Error cases
// ============================================================================

func TestExtract_PathNotFound(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"user": {"id": "123"}}`)

	_, err := extractor.Extract(jsonData, "user.nonexistent")
	if err == nil {
		t.Error("Extract() should fail when path doesn't exist")
	}

	expectedError := "path 'user.nonexistent' not found in JSON"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestExtract_EmptyJSON(t *testing.T) {
	extractor := New()

	_, err := extractor.Extract([]byte{}, "user.id")
	if err == nil {
		t.Error("Extract() should fail with empty JSON data")
	}

	expectedError := "json data cannot be empty"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestExtract_EmptyPath(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"user": {"id": "123"}}`)

	_, err := extractor.Extract(jsonData, "")
	if err == nil {
		t.Error("Extract() should fail with empty path")
	}

	expectedError := "path cannot be empty"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestExtract_InvalidJSON(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{invalid json}`)

	// gjson is lenient and won't fail on invalid JSON, but paths won't exist
	_, err := extractor.Extract(jsonData, "user.id")
	if err == nil {
		t.Error("Extract() should fail when path doesn't exist in invalid JSON")
	}
}

// ============================================================================
// Exists() Tests
// ============================================================================

func TestExists_PathExists(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"user": {"id": "123", "name": "John"}}`)

	tests := []struct {
		name   string
		path   string
		exists bool
	}{
		{
			name:   "existing simple path",
			path:   "user.id",
			exists: true,
		},
		{
			name:   "existing nested path",
			path:   "user.name",
			exists: true,
		},
		{
			name:   "non-existing path",
			path:   "user.email",
			exists: false,
		},
		{
			name:   "non-existing root path",
			path:   "profile",
			exists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.Exists(jsonData, tt.path)
			if result != tt.exists {
				t.Errorf("Expected exists=%v, got %v", tt.exists, result)
			}
		})
	}
}

func TestExists_EmptyData(t *testing.T) {
	extractor := New()

	result := extractor.Exists([]byte{}, "user.id")
	if result {
		t.Error("Exists() should return false for empty JSON data")
	}
}

func TestExists_EmptyPath(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"user": {"id": "123"}}`)

	result := extractor.Exists(jsonData, "")
	if result {
		t.Error("Exists() should return false for empty path")
	}
}

func TestExists_NullValue(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"user": {"email": null}}`)

	// Null values exist, they just have null value
	result := extractor.Exists(jsonData, "user.email")
	if !result {
		t.Error("Exists() should return true for null values")
	}
}

// ============================================================================
// ExtractWithDefault() Tests - String defaults
// ============================================================================

func TestExtractWithDefault_PathExists_String(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"user": {"id": "123"}}`)

	result := extractor.ExtractWithDefault(jsonData, "user.id", "default-id")

	strResult, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string type, got %T", result)
	}

	expected := "123"
	if strResult != expected {
		t.Errorf("Expected '%s', got '%s'", expected, strResult)
	}
}

func TestExtractWithDefault_PathNotFound_String(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"user": {"id": "123"}}`)

	defaultValue := "default-email"
	result := extractor.ExtractWithDefault(jsonData, "user.email", defaultValue)

	strResult, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string type, got %T", result)
	}

	if strResult != defaultValue {
		t.Errorf("Expected default value '%s', got '%s'", defaultValue, strResult)
	}
}

// ============================================================================
// ExtractWithDefault() Tests - Numeric defaults
// ============================================================================

func TestExtractWithDefault_PathExists_Number(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"count": 42}`)

	result := extractor.ExtractWithDefault(jsonData, "count", 0)

	// gjson returns float64 for numbers
	numResult, ok := result.(float64)
	if !ok {
		t.Fatalf("Expected float64 type, got %T", result)
	}

	if int(numResult) != 42 {
		t.Errorf("Expected 42, got %v", numResult)
	}
}

func TestExtractWithDefault_PathNotFound_Number(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"count": 42}`)

	defaultValue := 100
	result := extractor.ExtractWithDefault(jsonData, "missing", defaultValue)

	intResult, ok := result.(int)
	if !ok {
		t.Fatalf("Expected int type, got %T", result)
	}

	if intResult != defaultValue {
		t.Errorf("Expected default value %d, got %d", defaultValue, intResult)
	}
}

// ============================================================================
// ExtractWithDefault() Tests - Boolean defaults
// ============================================================================

func TestExtractWithDefault_PathExists_Bool(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"active": true}`)

	result := extractor.ExtractWithDefault(jsonData, "active", false)

	boolResult, ok := result.(bool)
	if !ok {
		t.Fatalf("Expected bool type, got %T", result)
	}

	if !boolResult {
		t.Error("Expected true, got false")
	}
}

func TestExtractWithDefault_PathNotFound_Bool(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"active": true}`)

	defaultValue := false
	result := extractor.ExtractWithDefault(jsonData, "verified", defaultValue)

	boolResult, ok := result.(bool)
	if !ok {
		t.Fatalf("Expected bool type, got %T", result)
	}

	if boolResult != defaultValue {
		t.Errorf("Expected default value %v, got %v", defaultValue, boolResult)
	}
}

// ============================================================================
// ExtractWithDefault() Tests - Empty data
// ============================================================================

func TestExtractWithDefault_EmptyJSON(t *testing.T) {
	extractor := New()

	defaultValue := "default-value"
	result := extractor.ExtractWithDefault([]byte{}, "user.id", defaultValue)

	strResult, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string type, got %T", result)
	}

	if strResult != defaultValue {
		t.Errorf("Expected default value '%s', got '%s'", defaultValue, strResult)
	}
}

// ============================================================================
// Complex integration tests
// ============================================================================

func TestExtract_ComplexNestedStructure(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{
		"response": {
			"status": "success",
			"data": {
				"users": [
					{
						"id": 1,
						"name": "Alice",
						"active": true,
						"metadata": {
							"created_at": "2024-01-01",
							"score": 95.5
						}
					},
					{
						"id": 2,
						"name": "Bob",
						"active": false,
						"metadata": {
							"created_at": "2024-01-02",
							"score": 87.3
						}
					}
				]
			}
		}
	}`)

	tests := []struct {
		name     string
		path     string
		validate func(t *testing.T, result any)
	}{
		{
			name: "extract status string",
			path: "response.status",
			validate: func(t *testing.T, result any) {
				if result.(string) != "success" {
					t.Errorf("Expected 'success', got '%v'", result)
				}
			},
		},
		{
			name: "extract first user id (number)",
			path: "response.data.users.0.id",
			validate: func(t *testing.T, result any) {
				if int(result.(float64)) != 1 {
					t.Errorf("Expected 1, got %v", result)
				}
			},
		},
		{
			name: "extract second user active status (bool)",
			path: "response.data.users.1.active",
			validate: func(t *testing.T, result any) {
				if result.(bool) != false {
					t.Errorf("Expected false, got %v", result)
				}
			}},
		{
			name: "extract all user names (array)",
			path: "response.data.users.#.name",
			validate: func(t *testing.T, result any) {
				names := result.([]interface{})
				if len(names) != 2 {
					t.Errorf("Expected 2 names, got %d", len(names))
				}
				if names[0].(string) != "Alice" || names[1].(string) != "Bob" {
					t.Errorf("Unexpected names: %v", names)
				}
			},
		},
		{
			name: "extract nested metadata score (float)",
			path: "response.data.users.0.metadata.score",
			validate: func(t *testing.T, result any) {
				if result.(float64) != 95.5 {
					t.Errorf("Expected 95.5, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractor.Extract(jsonData, tt.path)
			if err != nil {
				t.Fatalf("Extract() failed: %v", err)
			}
			tt.validate(t, result)
		})
	}
}

func TestExtract_EmptyString(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{"user": {"id": "", "name": "John"}}`)

	result, err := extractor.Extract(jsonData, "user.id")
	if err != nil {
		t.Fatalf("Extract() failed: %v", err)
	}

	strResult, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string type, got %T", result)
	}

	if strResult != "" {
		t.Errorf("Expected empty string, got '%s'", strResult)
	}
}

func TestExtract_ZeroValues(t *testing.T) {
	extractor := New()
	jsonData := []byte(`{
		"count": 0,
		"price": 0.0,
		"active": false,
		"name": ""
	}`)

	tests := []struct {
		name     string
		path     string
		expected any
	}{
		{"zero integer", "count", float64(0)},
		{"zero float", "price", float64(0)},
		{"false bool", "active", false},
		{"empty string", "name", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractor.Extract(jsonData, tt.path)
			if err != nil {
				t.Fatalf("Extract() failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %v (%T), got %v (%T)", tt.expected, tt.expected, result, result)
			}
		})
	}
}
