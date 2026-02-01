package openapi

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// Test OpenAPI spec samples
const (
	validOpenAPISpec = `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
servers:
  - url: https://api.example.com
tags:
  - name: users
    description: User operations
  - name: products
    description: Product operations
paths:
  /users:
    get:
      operationId: listUsers
      summary: List users
      description: Returns a list of users
      tags:
        - users
      responses:
        '200':
          description: Success
    post:
      operationId: createUser
      summary: Create user
      tags:
        - users
      responses:
        '201':
          description: Created
  /users/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
    get:
      operationId: getUser
      summary: Get user by ID
      tags:
        - users
      responses:
        '200':
          description: Success
    put:
      operationId: updateUser
      summary: Update user
      tags:
        - users
      responses:
        '200':
          description: Updated
    delete:
      operationId: deleteUser
      summary: Delete user
      tags:
        - users
      responses:
        '204':
          description: Deleted
  /products:
    get:
      operationId: listProducts
      summary: List products
      tags:
        - products
      responses:
        '200':
          description: Success
    post:
      operationId: createProduct
      summary: Create product
      tags:
        - products
      responses:
        '201':
          description: Created
  /products/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
    patch:
      operationId: updateProduct
      summary: Update product
      tags:
        - products
      responses:
        '200':
          description: Updated
    delete:
      operationId: deleteProduct
      summary: Delete product
      tags:
        - products
      responses:
        '204':
          description: Deleted
  /health:
    head:
      operationId: healthCheck
      summary: Health check
      responses:
        '200':
          description: Healthy
    options:
      operationId: healthOptions
      summary: Health options
      responses:
        '200':
          description: Options
    trace:
      operationId: healthTrace
      summary: Health trace
      responses:
        '200':
          description: Trace`

	invalidOpenAPISpec = `openapi: 3.0.3
info:
  title: Invalid API
paths:
  /test:
    get:
      responses: "invalid"`

	emptyPathsSpec = `openapi: 3.0.3
info:
  title: Empty Paths API
  version: 1.0.0
paths: {}`

	noTagsSpec = `openapi: 3.0.3
info:
  title: No Tags API
  version: 1.0.0
paths:
  /test:
    get:
      operationId: testOp
      responses:
        '200':
          description: Success`

	multipleServersSpec = `openapi: 3.0.3
info:
  title: Multi Server API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
    description: Production
  - url: https://staging.example.com/v1
    description: Staging
  - url: http://localhost:3000
    description: Development
paths:
  /test:
    get:
      responses:
        '200':
          description: Success`
)

func TestNew(t *testing.T) {
	parser := New()
	if parser == nil {
		t.Fatal("New() returned nil")
	}
	if parser.doc != nil {
		t.Error("New parser should have nil document")
	}
}

func TestParseData_ValidSpec(t *testing.T) {
	parser := New()
	err := parser.ParseData([]byte(validOpenAPISpec))
	if err != nil {
		t.Fatalf("ParseData() failed with valid spec: %v", err)
	}
	if parser.doc == nil {
		t.Error("Document should not be nil after successful parse")
	}
}

func TestParseData_InvalidSpec(t *testing.T) {
	parser := New()
	err := parser.ParseData([]byte(invalidOpenAPISpec))
	if err == nil {
		t.Error("ParseData() should fail with invalid spec")
	}
}

func TestParseData_EmptyData(t *testing.T) {
	parser := New()
	err := parser.ParseData([]byte{})
	if err == nil {
		t.Error("ParseData() should fail with empty data")
	}
}

func TestParseData_MalformedYAML(t *testing.T) {
	parser := New()
	err := parser.ParseData([]byte("not: valid: yaml: data"))
	if err == nil {
		t.Error("ParseData() should fail with malformed YAML")
	}
}

func TestParseFile_ValidFile(t *testing.T) {
	// Create temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "openapi.yaml")
	err := os.WriteFile(tmpFile, []byte(validOpenAPISpec), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	parser := New()
	err = parser.ParseFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseFile() failed: %v", err)
	}
	if parser.doc == nil {
		t.Error("Document should not be nil after successful parse")
	}
}

func TestParseFile_NonExistentFile(t *testing.T) {
	parser := New()
	err := parser.ParseFile("/nonexistent/path/to/file.yaml")
	if err == nil {
		t.Error("ParseFile() should fail with non-existent file")
	}
}

func TestParseFile_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.yaml")
	err := os.WriteFile(tmpFile, []byte(invalidOpenAPISpec), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	parser := New()
	err = parser.ParseFile(tmpFile)
	if err == nil {
		t.Error("ParseFile() should fail with invalid spec")
	}
}

func TestGetEndpoints_ValidSpec(t *testing.T) {
	parser := New()
	err := parser.ParseData([]byte(validOpenAPISpec))
	if err != nil {
		t.Fatalf("ParseData() failed: %v", err)
	}

	endpoints, err := parser.GetEndpoints()
	if err != nil {
		t.Fatalf("GetEndpoints() failed: %v", err)
	}

	expectedCount := 12
	if len(endpoints) != expectedCount {
		t.Errorf("Expected %d endpoints, got %d", expectedCount, len(endpoints))
	}

	foundGetUsers := false
	foundPostProducts := false
	foundDeleteUser := false

	for _, ep := range endpoints {
		if ep.Path == "/users" && ep.Method == "GET" {
			foundGetUsers = true
			if ep.OperationID != "listUsers" {
				t.Errorf("Expected operationId 'listUsers', got '%s'", ep.OperationID)
			}
			if ep.Summary != "List users" {
				t.Errorf("Expected summary 'List users', got '%s'", ep.Summary)
			}
			if len(ep.Tags) != 1 || ep.Tags[0] != "users" {
				t.Errorf("Expected tags ['users'], got %v", ep.Tags)
			}
		}
		if ep.Path == "/products" && ep.Method == "POST" {
			foundPostProducts = true
		}
		if ep.Path == "/users/{id}" && ep.Method == "DELETE" {
			foundDeleteUser = true
		}
	}

	if !foundGetUsers {
		t.Error("Expected to find GET /users endpoint")
	}
	if !foundPostProducts {
		t.Error("Expected to find POST /products endpoint")
	}
	if !foundDeleteUser {
		t.Error("Expected to find DELETE /users/{id} endpoint")
	}
}

func TestGetEndpoints_AllHTTPMethods(t *testing.T) {
	parser := New()
	err := parser.ParseData([]byte(validOpenAPISpec))
	if err != nil {
		t.Fatalf("ParseData() failed: %v", err)
	}

	endpoints, err := parser.GetEndpoints()
	if err != nil {
		t.Fatalf("GetEndpoints() failed: %v", err)
	}

	methods := make(map[string]bool)
	for _, ep := range endpoints {
		methods[ep.Method] = true
	}

	expectedMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "TRACE"}
	for _, method := range expectedMethods {
		if !methods[method] {
			t.Errorf("Expected to find %s method in endpoints", method)
		}
	}
}

func TestGetEndpoints_NoDocument(t *testing.T) {
	parser := New()
	_, err := parser.GetEndpoints()
	if err == nil {
		t.Error("GetEndpoints() should fail when no document is loaded")
	}
}

func TestGetEndpoints_EmptyPaths(t *testing.T) {
	parser := New()
	err := parser.ParseData([]byte(emptyPathsSpec))
	if err != nil {
		t.Fatalf("ParseData() failed: %v", err)
	}

	endpoints, err := parser.GetEndpoints()
	if err != nil {
		t.Fatalf("GetEndpoints() failed: %v", err)
	}

	if len(endpoints) != 0 {
		t.Errorf("Expected 0 endpoints for empty paths, got %d", len(endpoints))
	}
}

func TestGetEndpointsByTag_ValidTag(t *testing.T) {
	parser := New()
	err := parser.ParseData([]byte(validOpenAPISpec))
	if err != nil {
		t.Fatalf("ParseData() failed: %v", err)
	}

	endpoints, err := parser.GetEndpointsByTag("users")
	if err != nil {
		t.Fatalf("GetEndpointsByTag() failed: %v", err)
	}

	expectedCount := 5 // GET, POST /users and GET, PUT, DELETE /users/{id}
	if len(endpoints) != expectedCount {
		t.Errorf("Expected %d endpoints with tag 'users', got %d", expectedCount, len(endpoints))
	}

	for _, ep := range endpoints {
		found := slices.Contains(ep.Tags, "users")
		if !found {
			t.Errorf("Endpoint %s %s should have 'users' tag", ep.Method, ep.Path)
		}
	}
}

func TestGetEndpointsByTag_NonExistentTag(t *testing.T) {
	parser := New()
	err := parser.ParseData([]byte(validOpenAPISpec))
	if err != nil {
		t.Fatalf("ParseData() failed: %v", err)
	}

	endpoints, err := parser.GetEndpointsByTag("nonexistent")
	if err != nil {
		t.Fatalf("GetEndpointsByTag() failed: %v", err)
	}

	if len(endpoints) != 0 {
		t.Errorf("Expected 0 endpoints for non-existent tag, got %d", len(endpoints))
	}
}

func TestGetEndpointsByTag_NoDocument(t *testing.T) {
	parser := New()
	_, err := parser.GetEndpointsByTag("users")
	if err == nil {
		t.Error("GetEndpointsByTag() should fail when no document is loaded")
	}
}

func TestGetTags_ValidSpec(t *testing.T) {
	parser := New()
	err := parser.ParseData([]byte(validOpenAPISpec))
	if err != nil {
		t.Fatalf("ParseData() failed: %v", err)
	}

	tags, err := parser.GetTags()
	if err != nil {
		t.Fatalf("GetTags() failed: %v", err)
	}

	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}

	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}

	if !tagSet["users"] {
		t.Error("Expected to find 'users' tag")
	}
	if !tagSet["products"] {
		t.Error("Expected to find 'products' tag")
	}
}

func TestGetTags_NoTags(t *testing.T) {
	parser := New()
	err := parser.ParseData([]byte(noTagsSpec))
	if err != nil {
		t.Fatalf("ParseData() failed: %v", err)
	}

	tags, err := parser.GetTags()
	if err != nil {
		t.Fatalf("GetTags() failed: %v", err)
	}

	if len(tags) != 0 {
		t.Errorf("Expected 0 tags, got %d: %v", len(tags), tags)
	}
}

func TestGetTags_NoDocument(t *testing.T) {
	parser := New()
	_, err := parser.GetTags()
	if err == nil {
		t.Error("GetTags() should fail when no document is loaded")
	}
}

func TestGetInfo_ValidSpec(t *testing.T) {
	parser := New()
	err := parser.ParseData([]byte(validOpenAPISpec))
	if err != nil {
		t.Fatalf("ParseData() failed: %v", err)
	}

	info, err := parser.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo() failed: %v", err)
	}

	if info == nil {
		t.Fatal("Info should not be nil")
	}

	if info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got '%s'", info.Title)
	}

	if info.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", info.Version)
	}
}

func TestGetInfo_NoDocument(t *testing.T) {
	parser := New()
	_, err := parser.GetInfo()
	if err == nil {
		t.Error("GetInfo() should fail when no document is loaded")
	}
}

func TestGetServerURLs_ValidSpec(t *testing.T) {
	parser := New()
	err := parser.ParseData([]byte(validOpenAPISpec))
	if err != nil {
		t.Fatalf("ParseData() failed: %v", err)
	}

	urls, err := parser.GetServerURLs()
	if err != nil {
		t.Fatalf("GetServerURLs() failed: %v", err)
	}

	if len(urls) != 1 {
		t.Errorf("Expected 1 server URL, got %d", len(urls))
	}

	if urls[0] != "https://api.example.com" {
		t.Errorf("Expected URL 'https://api.example.com', got '%s'", urls[0])
	}
}

func TestGetServerURLs_MultipleServers(t *testing.T) {
	parser := New()
	err := parser.ParseData([]byte(multipleServersSpec))
	if err != nil {
		t.Fatalf("ParseData() failed: %v", err)
	}

	urls, err := parser.GetServerURLs()
	if err != nil {
		t.Fatalf("GetServerURLs() failed: %v", err)
	}

	if len(urls) != 3 {
		t.Errorf("Expected 3 server URLs, got %d", len(urls))
	}

	expectedURLs := map[string]bool{
		"https://api.example.com/v1":     true,
		"https://staging.example.com/v1": true,
		"http://localhost:3000":          true,
	}

	for _, url := range urls {
		if !expectedURLs[url] {
			t.Errorf("Unexpected URL: %s", url)
		}
	}
}

func TestGetServerURLs_NoDocument(t *testing.T) {
	parser := New()
	_, err := parser.GetServerURLs()
	if err == nil {
		t.Error("GetServerURLs() should fail when no document is loaded")
	}
}

func TestEndpointStruct(t *testing.T) {
	parser := New()
	err := parser.ParseData([]byte(validOpenAPISpec))
	if err != nil {
		t.Fatalf("ParseData() failed: %v", err)
	}

	endpoints, err := parser.GetEndpoints()
	if err != nil {
		t.Fatalf("GetEndpoints() failed: %v", err)
	}

	// Find GET /users endpoint to verify all fields
	var getUsersEndpoint *Endpoint
	for _, ep := range endpoints {
		if ep.Path == "/users" && ep.Method == "GET" {
			getUsersEndpoint = &ep
			break
		}
	}

	if getUsersEndpoint == nil {
		t.Fatal("Could not find GET /users endpoint")
	}

	if getUsersEndpoint.Path != "/users" {
		t.Errorf("Expected path '/users', got '%s'", getUsersEndpoint.Path)
	}

	if getUsersEndpoint.Method != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", getUsersEndpoint.Method)
	}

	if getUsersEndpoint.OperationID != "listUsers" {
		t.Errorf("Expected operationId 'listUsers', got '%s'", getUsersEndpoint.OperationID)
	}

	if getUsersEndpoint.Summary != "List users" {
		t.Errorf("Expected summary 'List users', got '%s'", getUsersEndpoint.Summary)
	}

	if getUsersEndpoint.Description != "Returns a list of users" {
		t.Errorf("Expected description 'Returns a list of users', got '%s'", getUsersEndpoint.Description)
	}

	if len(getUsersEndpoint.Tags) != 1 || getUsersEndpoint.Tags[0] != "users" {
		t.Errorf("Expected tags ['users'], got %v", getUsersEndpoint.Tags)
	}
}
