package executor

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.doFunc != nil {
		return m.doFunc(req)
	}
	return nil, errors.New("mock client not configured")
}

func TestNew(t *testing.T) {
	executor, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if executor == nil {
		t.Fatal("New() returned nil executor")
	}
	if executor.client == nil {
		t.Error("Executor client should not be nil")
	}
	if executor.jar == nil {
		t.Error("Executor cookie jar should not be nil")
	}
}

func TestNewWithClient(t *testing.T) {
	mockClient := &mockHTTPClient{}
	executor := NewWithClient(mockClient)

	if executor == nil {
		t.Fatal("NewWithClient() returned nil")
	}
	if executor.client != mockClient {
		t.Error("Executor should use provided client")
	}
}

func TestExecute_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"success"}`))
	}))
	defer server.Close()

	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	req := &Request{
		Method: http.MethodGet,
		URL:    server.URL,
		Headers: map[string]string{
			"User-Agent": "test-agent",
		},
	}

	ctx := context.Background()
	resp, err := executor.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if string(resp.Body) != `{"message":"success"}` {
		t.Errorf("Unexpected response body: %s", string(resp.Body))
	}

	if resp.Duration == 0 {
		t.Error("Response duration should be greater than 0")
	}
}

func TestExecute_WithBody(t *testing.T) {
	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"123"}`))
	}))
	defer server.Close()

	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	requestBody := []byte(`{"name":"test"}`)
	req := &Request{
		Method: http.MethodPost,
		URL:    server.URL,
		Body:   requestBody,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	ctx := context.Background()
	resp, err := executor.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	if receivedBody != string(requestBody) {
		t.Errorf("Expected body %s, got %s", string(requestBody), receivedBody)
	}
}

func TestExecute_NilRequest(t *testing.T) {
	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx := context.Background()
	_, err = executor.Execute(ctx, nil)

	if err == nil {
		t.Error("Execute() should fail with nil request")
	}
}

func TestExecute_EmptyURL(t *testing.T) {
	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	req := &Request{
		Method: http.MethodGet,
		URL:    "",
	}

	ctx := context.Background()
	_, err = executor.Execute(ctx, req)

	if err == nil {
		t.Error("Execute() should fail with empty URL")
	}
}

func TestExecute_DefaultMethod(t *testing.T) {
	var receivedMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	req := &Request{
		URL: server.URL,
		// Method not set
	}

	ctx := context.Background()
	_, err = executor.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if receivedMethod != http.MethodGet {
		t.Errorf("Expected default method %s, got %s", http.MethodGet, receivedMethod)
	}
}

func TestExecute_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	req := &Request{
		Method:  http.MethodGet,
		URL:     server.URL,
		Timeout: 50 * time.Millisecond,
	}

	ctx := context.Background()
	_, err = executor.Execute(ctx, req)

	if err == nil {
		t.Error("Execute() should timeout")
	}

	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestExecute_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	req := &Request{
		Method: http.MethodGet,
		URL:    server.URL,
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err = executor.Execute(ctx, req)

	if err == nil {
		t.Error("Execute() should fail when context is cancelled")
	}

	if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected context canceled error, got: %v", err)
	}
}

func TestExecute_Headers(t *testing.T) {
	var receivedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	headers := map[string]string{
		"User-Agent":   "custom-agent",
		"Content-Type": "application/json",
		"X-Custom":     "test-value",
	}

	req := &Request{
		Method:  http.MethodGet,
		URL:     server.URL,
		Headers: headers,
	}

	ctx := context.Background()
	_, err = executor.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	for key, value := range headers {
		if receivedHeaders.Get(key) != value {
			t.Errorf("Expected header %s=%s, got %s", key, value, receivedHeaders.Get(key))
		}
	}
}

func TestGET(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GET response"))
	}))
	defer server.Close()

	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx := context.Background()
	resp, err := executor.GET(ctx, server.URL, nil)

	if err != nil {
		t.Fatalf("GET() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if string(resp.Body) != "GET response" {
		t.Errorf("Unexpected response body: %s", string(resp.Body))
	}
}

func TestPOST(t *testing.T) {
	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	body := []byte("POST data")
	ctx := context.Background()
	resp, err := executor.POST(ctx, server.URL, body, nil)

	if err != nil {
		t.Fatalf("POST() failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	if receivedBody != string(body) {
		t.Errorf("Expected body %s, got %s", string(body), receivedBody)
	}
}

func TestPUT(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx := context.Background()
	resp, err := executor.PUT(ctx, server.URL, []byte("PUT data"), nil)

	if err != nil {
		t.Fatalf("PUT() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestPATCH(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx := context.Background()
	resp, err := executor.PATCH(ctx, server.URL, []byte("PATCH data"), nil)

	if err != nil {
		t.Fatalf("PATCH() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestDELETE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx := context.Background()
	resp, err := executor.DELETE(ctx, server.URL, nil)

	if err != nil {
		t.Fatalf("DELETE() failed: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}

func TestHEAD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Errorf("Expected HEAD method, got %s", r.Method)
		}
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx := context.Background()
	resp, err := executor.HEAD(ctx, server.URL, nil)

	if err != nil {
		t.Fatalf("HEAD() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestOPTIONS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			t.Errorf("Expected OPTIONS method, got %s", r.Method)
		}
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx := context.Background()
	resp, err := executor.OPTIONS(ctx, server.URL, nil)

	if err != nil {
		t.Fatalf("OPTIONS() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestTRACE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodTrace {
			t.Errorf("Expected TRACE method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx := context.Background()
	resp, err := executor.TRACE(ctx, server.URL, nil)

	if err != nil {
		t.Fatalf("TRACE() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestCookieJar(t *testing.T) {
	cookieName := "session"
	cookieValue := "test-session-id"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/set-cookie" {
			http.SetCookie(w, &http.Cookie{
				Name:  cookieName,
				Value: cookieValue,
				Path:  "/",
			})
			w.WriteHeader(http.StatusOK)
		} else if r.URL.Path == "/check-cookie" {
			cookie, err := r.Cookie(cookieName)
			if err != nil {
				t.Errorf("Cookie not found: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if cookie.Value != cookieValue {
				t.Errorf("Expected cookie value %s, got %s", cookieValue, cookie.Value)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx := context.Background()

	_, err = executor.GET(ctx, server.URL+"/set-cookie", nil)
	if err != nil {
		t.Fatalf("Failed to set cookie: %v", err)
	}

	parsedURL, _ := url.Parse(server.URL)
	cookies := executor.jar.Cookies(parsedURL)
	found := false
	for _, cookie := range cookies {
		if cookie.Name == cookieName && cookie.Value == cookieValue {
			found = true
			break
		}
	}
	if !found {
		t.Error("Cookie not found in jar")
	}

	resp, err := executor.GET(ctx, server.URL+"/check-cookie", nil)
	if err != nil {
		t.Fatalf("Failed to check cookie: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Error("Cookie was not sent with second request")
	}
}

func TestGetCookieJar(t *testing.T) {
	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	jar := executor.GetCookieJar()
	if jar == nil {
		t.Error("GetCookieJar() returned nil")
	}

	if jar != executor.jar {
		t.Error("GetCookieJar() should return the executor's jar")
	}
}

func TestResponseDuration(t *testing.T) {
	delay := 100 * time.Millisecond
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx := context.Background()
	resp, err := executor.GET(ctx, server.URL, nil)

	if err != nil {
		t.Fatalf("GET() failed: %v", err)
	}

	if resp.Duration < delay {
		t.Errorf("Expected duration >= %v, got %v", delay, resp.Duration)
	}
}

func TestExecute_InvalidURL(t *testing.T) {
	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	req := &Request{
		Method: http.MethodGet,
		URL:    "://invalid-url",
	}

	ctx := context.Background()
	_, err = executor.Execute(ctx, req)

	if err == nil {
		t.Error("Execute() should fail with invalid URL")
	}
}

func TestExecute_NetworkError(t *testing.T) {
	executor, err := New()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	req := &Request{
		Method: http.MethodGet,
		URL:    "http://localhost:99999", // invalid port
	}

	ctx := context.Background()
	_, err = executor.Execute(ctx, req)

	if err == nil {
		t.Error("Execute() should fail with network error")
	}
}
