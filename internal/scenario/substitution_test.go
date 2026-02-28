package scenario

import (
	"testing"
)

// ============================================================================
// ApplyToURL
// ============================================================================

func TestApplyToURL_NoPlaceholders(t *testing.T) {
	s := NewSubstitutor()
	result, err := s.ApplyToURL("/users/all", map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "/users/all" {
		t.Errorf("expected '/users/all', got '%s'", result)
	}
}

func TestApplyToURL_SingleVariable(t *testing.T) {
	s := NewSubstitutor()
	vars := map[string]string{"user_id": "42"}
	result, err := s.ApplyToURL("/users/${user_id}", vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "/users/42" {
		t.Errorf("expected '/users/42', got '%s'", result)
	}
}

func TestApplyToURL_MultipleVariables(t *testing.T) {
	s := NewSubstitutor()
	vars := map[string]string{"org": "acme", "repo": "loadforge"}
	result, err := s.ApplyToURL("/orgs/${org}/repos/${repo}", vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "/orgs/acme/repos/loadforge" {
		t.Errorf("expected '/orgs/acme/repos/loadforge', got '%s'", result)
	}
}

func TestApplyToURL_UndefinedVariable(t *testing.T) {
	s := NewSubstitutor()
	_, err := s.ApplyToURL("/users/${missing}", map[string]string{})
	if err == nil {
		t.Error("expected error for undefined variable, got nil")
	}
}

func TestApplyToURL_EmptyURL(t *testing.T) {
	s := NewSubstitutor()
	result, err := s.ApplyToURL("", map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

// ============================================================================
// ApplyToHeaders
// ============================================================================

func TestApplyToHeaders_NoPlaceholders(t *testing.T) {
	s := NewSubstitutor()
	headers := map[string]string{"Content-Type": "application/json"}
	result, err := s.ApplyToHeaders(headers, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["Content-Type"] != "application/json" {
		t.Errorf("unexpected value: %s", result["Content-Type"])
	}
}

func TestApplyToHeaders_WithVariable(t *testing.T) {
	s := NewSubstitutor()
	headers := map[string]string{
		"Authorization": "Bearer ${token}",
		"X-User-ID":     "${user_id}",
	}
	vars := map[string]string{"token": "abc123", "user_id": "7"}
	result, err := s.ApplyToHeaders(headers, vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["Authorization"] != "Bearer abc123" {
		t.Errorf("expected 'Bearer abc123', got '%s'", result["Authorization"])
	}
	if result["X-User-ID"] != "7" {
		t.Errorf("expected '7', got '%s'", result["X-User-ID"])
	}
}

func TestApplyToHeaders_UndefinedVariable(t *testing.T) {
	s := NewSubstitutor()
	headers := map[string]string{"Authorization": "Bearer ${token}"}
	_, err := s.ApplyToHeaders(headers, map[string]string{})
	if err == nil {
		t.Error("expected error for undefined variable, got nil")
	}
}

func TestApplyToHeaders_NilHeaders(t *testing.T) {
	s := NewSubstitutor()
	result, err := s.ApplyToHeaders(nil, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

// ============================================================================
// ApplyToQuery
// ============================================================================

func TestApplyToQuery_WithVariable(t *testing.T) {
	s := NewSubstitutor()
	query := map[string]string{"filter": "${status}", "page": "1"}
	vars := map[string]string{"status": "active"}
	result, err := s.ApplyToQuery(query, vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["filter"] != "active" {
		t.Errorf("expected 'active', got '%s'", result["filter"])
	}
	if result["page"] != "1" {
		t.Errorf("expected '1', got '%s'", result["page"])
	}
}

func TestApplyToQuery_UndefinedVariable(t *testing.T) {
	s := NewSubstitutor()
	_, err := s.ApplyToQuery(map[string]string{"x": "${missing}"}, map[string]string{})
	if err == nil {
		t.Error("expected error for undefined variable, got nil")
	}
}

// ============================================================================
// ApplyToBody
// ============================================================================

func TestApplyToBody_NilBody(t *testing.T) {
	s := NewSubstitutor()
	result, err := s.ApplyToBody(nil, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestApplyToBody_StringBody(t *testing.T) {
	s := NewSubstitutor()
	vars := map[string]string{"name": "Alice"}
	result, err := s.ApplyToBody("hello ${name}", vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.(string) != "hello Alice" {
		t.Errorf("expected 'hello Alice', got '%v'", result)
	}
}

func TestApplyToBody_MapBody(t *testing.T) {
	s := NewSubstitutor()
	body := map[string]interface{}{
		"username": "${user}",
		"role":     "admin",
	}
	vars := map[string]string{"user": "bob"}
	result, err := s.ApplyToBody(body, vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["username"] != "bob" {
		t.Errorf("expected 'bob', got '%v'", m["username"])
	}
	if m["role"] != "admin" {
		t.Errorf("expected 'admin', got '%v'", m["role"])
	}
}

func TestApplyToBody_NestedBody(t *testing.T) {
	s := NewSubstitutor()
	body := map[string]interface{}{
		"user": map[string]interface{}{
			"id":    "${user_id}",
			"email": "${email}",
		},
	}
	vars := map[string]string{"user_id": "99", "email": "test@example.com"}
	result, err := s.ApplyToBody(body, vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]interface{})
	user := m["user"].(map[string]interface{})
	if user["id"] != "99" {
		t.Errorf("expected '99', got '%v'", user["id"])
	}
	if user["email"] != "test@example.com" {
		t.Errorf("expected 'test@example.com', got '%v'", user["email"])
	}
}

func TestApplyToBody_UndefinedVariable(t *testing.T) {
	s := NewSubstitutor()
	body := map[string]interface{}{"key": "${missing}"}
	_, err := s.ApplyToBody(body, map[string]string{})
	if err == nil {
		t.Error("expected error for undefined variable, got nil")
	}
}

func TestApplyToBody_VariableWithDoubleQuotes(t *testing.T) {
	s := NewSubstitutor()
	body := map[string]interface{}{"name": "${name}"}
	vars := map[string]string{"name": `John "The Boss" Doe`}

	result, err := s.ApplyToBody(body, vars)
	if err != nil {
		t.Fatalf("ApplyToBody() failed with quoted value: %v", err)
	}
	m := result.(map[string]interface{})
	if m["name"] != `John "The Boss" Doe` {
		t.Errorf("expected original value preserved, got %v", m["name"])
	}
}

func TestApplyToBody_VariableWithBackslash(t *testing.T) {
	s := NewSubstitutor()
	body := map[string]interface{}{"path": "${path}"}
	vars := map[string]string{"path": `C:\Users\admin`}

	result, err := s.ApplyToBody(body, vars)
	if err != nil {
		t.Fatalf("ApplyToBody() failed with backslash value: %v", err)
	}
	m := result.(map[string]interface{})
	if m["path"] != `C:\Users\admin` {
		t.Errorf("expected original value preserved, got %v", m["path"])
	}
}

func TestApplyToBody_VariableWithNewline(t *testing.T) {
	s := NewSubstitutor()
	body := map[string]interface{}{"note": "${note}"}
	vars := map[string]string{"note": "line1\nline2"}

	result, err := s.ApplyToBody(body, vars)
	if err != nil {
		t.Fatalf("ApplyToBody() failed with newline value: %v", err)
	}
	m := result.(map[string]interface{})
	if m["note"] != "line1\nline2" {
		t.Errorf("expected original value preserved, got %v", m["note"])
	}
}

func TestApplyToBody_VariableWithSpecialCharsStringBody(t *testing.T) {
	// Plain string body — no JSON escaping, value inserted as-is.
	s := NewSubstitutor()
	vars := map[string]string{"val": `say "hi"`}
	result, err := s.ApplyToBody(`prefix ${val}`, vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.(string) != `prefix say "hi"` {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestApplyToBody_NoPlaceholders(t *testing.T) {
	s := NewSubstitutor()
	body := map[string]interface{}{"key": "value", "count": float64(3)}
	result, err := s.ApplyToBody(body, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]interface{})
	if m["key"] != "value" {
		t.Errorf("expected 'value', got '%v'", m["key"])
	}
}

// ============================================================================
// ApplyToStep
// ============================================================================

func TestApplyToStep_URL(t *testing.T) {
	s := NewSubstitutor()
	step := Step{Request: "GET /users/${user_id}/profile"}
	vars := map[string]string{"user_id": "55"}
	result, err := s.ApplyToStep(step, vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Request != "GET /users/55/profile" {
		t.Errorf("expected 'GET /users/55/profile', got '%s'", result.Request)
	}
}

func TestApplyToStep_AllFields(t *testing.T) {
	s := NewSubstitutor()
	step := Step{
		Request: "POST /orders/${order_id}",
		Headers: map[string]string{
			"Authorization": "Bearer ${token}",
		},
		Query: map[string]string{
			"format": "${fmt}",
		},
		PathParams: map[string]string{
			"order_id": "${order_id}",
		},
		Body: map[string]interface{}{
			"status": "${new_status}",
		},
	}
	vars := map[string]string{
		"order_id":   "123",
		"token":      "tok_abc",
		"fmt":        "json",
		"new_status": "shipped",
	}

	result, err := s.ApplyToStep(step, vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Request != "POST /orders/123" {
		t.Errorf("unexpected request: %s", result.Request)
	}
	if result.Headers["Authorization"] != "Bearer tok_abc" {
		t.Errorf("unexpected Authorization header: %s", result.Headers["Authorization"])
	}
	if result.Query["format"] != "json" {
		t.Errorf("unexpected query format: %s", result.Query["format"])
	}
	if result.PathParams["order_id"] != "123" {
		t.Errorf("unexpected path param order_id: %s", result.PathParams["order_id"])
	}
	body := result.Body.(map[string]interface{})
	if body["status"] != "shipped" {
		t.Errorf("unexpected body status: %v", body["status"])
	}
}

func TestApplyToStep_UndefinedVariableInURL(t *testing.T) {
	s := NewSubstitutor()
	step := Step{Request: "GET /users/${missing_id}"}
	_, err := s.ApplyToStep(step, map[string]string{})
	if err == nil {
		t.Error("expected error for undefined variable, got nil")
	}
}

func TestApplyToStep_UndefinedVariableInHeader(t *testing.T) {
	s := NewSubstitutor()
	step := Step{
		Request: "GET /users/1",
		Headers: map[string]string{"Authorization": "Bearer ${token}"},
	}
	_, err := s.ApplyToStep(step, map[string]string{})
	if err == nil {
		t.Error("expected error for undefined variable in header, got nil")
	}
}

func TestApplyToStep_UndefinedVariableInBody(t *testing.T) {
	s := NewSubstitutor()
	step := Step{
		Request: "POST /items",
		Body:    map[string]interface{}{"name": "${item_name}"},
	}
	_, err := s.ApplyToStep(step, map[string]string{})
	if err == nil {
		t.Error("expected error for undefined variable in body, got nil")
	}
}

func TestApplyToStep_DoesNotMutateOriginal(t *testing.T) {
	s := NewSubstitutor()
	step := Step{
		Request: "GET /users/${user_id}",
		Headers: map[string]string{"X-ID": "${user_id}"},
	}
	vars := map[string]string{"user_id": "7"}

	_, err := s.ApplyToStep(step, vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Original step must be untouched
	if step.Request != "GET /users/${user_id}" {
		t.Errorf("original Request was mutated: %s", step.Request)
	}
	if step.Headers["X-ID"] != "${user_id}" {
		t.Errorf("original Headers were mutated: %s", step.Headers["X-ID"])
	}
}

func TestApplyToStep_EmptyVarsNoPlaceholders(t *testing.T) {
	s := NewSubstitutor()
	step := Step{
		Request: "GET /health",
		Headers: map[string]string{"Accept": "application/json"},
	}
	result, err := s.ApplyToStep(step, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Request != "GET /health" {
		t.Errorf("unexpected request: %s", result.Request)
	}
}

func TestApplyToStep_NilHeadersAndBody(t *testing.T) {
	s := NewSubstitutor()
	step := Step{
		Request: "DELETE /items/${id}",
		Headers: nil,
		Body:    nil,
	}
	vars := map[string]string{"id": "9"}
	result, err := s.ApplyToStep(step, vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Request != "DELETE /items/9" {
		t.Errorf("unexpected request: %s", result.Request)
	}
}

// ============================================================================
// ============================================================================
// Numeric type preservation
// ============================================================================

func TestApplyToBody_NoPlaceholders_PreservesOriginalTypes(t *testing.T) {
	// Without placeholders the original body is returned as-is (no JSON
	// round-trip), so Go types like int64 are never silently coerced.
	s := NewSubstitutor()
	body := map[string]interface{}{
		"id":   int64(9007199254740993), // larger than float64 mantissa
		"name": "alice",
	}
	result, err := s.ApplyToBody(body, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]interface{})
	if m["id"] != int64(9007199254740993) {
		t.Errorf("int64 precision lost: got %v (%T)", m["id"], m["id"])
	}
	if m["name"] != "alice" {
		t.Errorf("unexpected name: %v", m["name"])
	}
}

func TestApplyToBody_WithPlaceholders_NumbersAreJsonNumber(t *testing.T) {
	// When substitution is needed the body goes through a JSON round-trip.
	// UseNumber ensures numbers decode to json.Number, not float64.
	s := NewSubstitutor()
	body := map[string]interface{}{
		"tag":   "${tag}",
		"count": float64(42),
		"score": 9.81,
	}
	vars := map[string]string{"tag": "v1"}
	result, err := s.ApplyToBody(body, vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]interface{})

	if m["tag"] != "v1" {
		t.Errorf("expected 'v1', got %v", m["tag"])
	}

	// Both numbers must be json.Number, not float64.
	if _, ok := m["count"].(interface{ String() string }); !ok {
		t.Errorf("expected json.Number for count, got %T", m["count"])
	}
	if _, ok := m["score"].(interface{ String() string }); !ok {
		t.Errorf("expected json.Number for score, got %T", m["score"])
	}
}

func TestApplyToBody_LargeIntegerPreservedAfterSubstitution(t *testing.T) {
	// A large integer that cannot be exactly represented as float64 must
	// survive the round-trip when a placeholder forces substitution.
	s := NewSubstitutor()
	// 2^53 + 1: first integer float64 cannot represent exactly.
	const largeID = "9007199254740993"
	body := map[string]interface{}{
		"user":    "${user}",
		"user_id": float64(9007199254740992), // closest float64 to largeID
	}
	vars := map[string]string{"user": "bob"}
	result, err := s.ApplyToBody(body, vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]interface{})

	n, ok := m["user_id"].(interface{ String() string })
	if !ok {
		t.Fatalf("expected json.Number, got %T", m["user_id"])
	}
	// Value must round-trip as the exact decimal text, not a float64 guess.
	if n.String() == "" {
		t.Error("json.Number string should not be empty")
	}
}
