package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// HTTPClient defines the interface for making HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Request represents an HTTP request to be executed
type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
	Timeout time.Duration
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Status     string
	Headers    map[string][]string
	Body       []byte
	Duration   time.Duration
}

// Executor handles HTTP request execution
type Executor struct {
	client HTTPClient
	jar    http.CookieJar
}

// New creates a new Executor with default settings
func New() (*Executor, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	return &Executor{
		client: client,
		jar:    jar,
	}, nil
}

// NewWithClient creates a new Executor with a custom HTTP client
func NewWithClient(client HTTPClient) *Executor {
	return &Executor{
		client: client,
	}
}

// Execute performs an HTTP request and returns the response
func (e *Executor) Execute(ctx context.Context, req *Request) (*Response, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if req.URL == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	if req.Method == "" {
		req.Method = http.MethodGet
	}

	var bodyReader io.Reader
	if req.Body != nil {
		bodyReader = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
		httpReq = httpReq.WithContext(ctx)
	}

	start := time.Now()
	httpResp, err := e.client.Do(httpReq)
	duration := time.Since(start)

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	response := &Response{
		StatusCode: httpResp.StatusCode,
		Status:     httpResp.Status,
		Headers:    httpResp.Header,
		Body:       respBody,
		Duration:   duration,
	}

	return response, nil
}

func (e *Executor) GET(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	req := &Request{
		Method:  http.MethodGet,
		URL:     url,
		Headers: headers,
	}
	return e.Execute(ctx, req)
}

func (e *Executor) POST(ctx context.Context, url string, body []byte, headers map[string]string) (*Response, error) {
	req := &Request{
		Method:  http.MethodPost,
		URL:     url,
		Body:    body,
		Headers: headers,
	}
	return e.Execute(ctx, req)
}

func (e *Executor) PUT(ctx context.Context, url string, body []byte, headers map[string]string) (*Response, error) {
	req := &Request{
		Method:  http.MethodPut,
		URL:     url,
		Body:    body,
		Headers: headers,
	}
	return e.Execute(ctx, req)
}

func (e *Executor) PATCH(ctx context.Context, url string, body []byte, headers map[string]string) (*Response, error) {
	req := &Request{
		Method:  http.MethodPatch,
		URL:     url,
		Body:    body,
		Headers: headers,
	}
	return e.Execute(ctx, req)
}

func (e *Executor) DELETE(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	req := &Request{
		Method:  http.MethodDelete,
		URL:     url,
		Headers: headers,
	}
	return e.Execute(ctx, req)
}

func (e *Executor) HEAD(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	req := &Request{
		Method:  http.MethodHead,
		URL:     url,
		Headers: headers,
	}
	return e.Execute(ctx, req)
}

func (e *Executor) OPTIONS(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	req := &Request{
		Method:  http.MethodOptions,
		URL:     url,
		Headers: headers,
	}
	return e.Execute(ctx, req)
}

func (e *Executor) TRACE(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	req := &Request{
		Method:  http.MethodTrace,
		URL:     url,
		Headers: headers,
	}
	return e.Execute(ctx, req)
}

func (e *Executor) GetCookieJar() http.CookieJar {
	return e.jar
}
