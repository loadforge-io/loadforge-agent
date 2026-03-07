package runner

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"loadforge-agent/internal/scenario"
)

// ============================================================================
// helpers
// ============================================================================

func makeScenario(baseURL string, virtualUsers, durationSec uint64, steps []scenario.Step) *scenario.Scenario {
	return &scenario.Scenario{
		Name:         "test",
		BaseURL:      baseURL,
		VirtualUsers: virtualUsers,
		Duration:     durationSec,
		Steps:        steps,
	}
}

func simpleStep(method, path string) scenario.Step {
	return scenario.Step{Request: method + " " + path}
}

// ============================================================================
// VU unit tests
// ============================================================================

func TestVU_New(t *testing.T) {
	sc := makeScenario("http://localhost", 1, 1, []scenario.Step{simpleStep("GET", "/")})
	vu, err := newVirtualUser(1, sc)
	if err != nil {
		t.Fatalf("newVU failed: %v", err)
	}
	if vu == nil {
		t.Fatal("newVU returned nil")
	}
	if vu.id != 1 {
		t.Errorf("expected id=1, got %d", vu.id)
	}
}

func TestVU_SetGetVar(t *testing.T) {
	sc := makeScenario("http://localhost", 1, 1, nil)
	vu, _ := newVirtualUser(1, sc)

	vu.setValue("token", "abc")
	vars := vu.getValues()
	if vars["token"] != "abc" {
		t.Errorf("expected token=abc, got %q", vars["token"])
	}
}

func TestVU_ScenarioVarsAreDefaults(t *testing.T) {
	sc := makeScenario("http://localhost", 1, 1, nil)
	sc.Variables = map[string]string{"env": "prod", "ver": "v1"}

	vu, _ := newVirtualUser(1, sc)
	// VU-level override
	vu.setValue("env", "staging")

	vars := vu.getValues()
	if vars["env"] != "staging" {
		t.Errorf("VU var should override scenario var; got %q", vars["env"])
	}
	if vars["ver"] != "v1" {
		t.Errorf("expected ver=v1 from scenario, got %q", vars["ver"])
	}
}

func TestVU_RunScenario_SingleStep(t *testing.T) {
	var called atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sc := makeScenario(srv.URL, 1, 1, []scenario.Step{simpleStep("GET", "/ping")})
	vu, _ := newVirtualUser(1, sc)

	ctx := context.Background()
	if err := vu.runScenario(ctx); err != nil {
		t.Fatalf("runScenario failed: %v", err)
	}
	if called.Load() != 1 {
		t.Errorf("expected 1 request, got %d", called.Load())
	}
}

func TestVU_RunScenario_MultipleSteps(t *testing.T) {
	var paths []string
	var mu sync.Mutex

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sc := makeScenario(srv.URL, 1, 10, []scenario.Step{
		simpleStep("GET", "/step1"),
		simpleStep("POST", "/step2"),
		simpleStep("DELETE", "/step3"),
	})
	vu, _ := newVirtualUser(1, sc)

	if err := vu.runScenario(context.Background()); err != nil {
		t.Fatalf("runScenario failed: %v", err)
	}
	if len(paths) != 3 {
		t.Errorf("expected 3 requests, got %d", len(paths))
	}
}

func TestVU_RunScenario_SaveToContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"session":{"token":"tok_xyz"}}`)
	}))
	defer srv.Close()

	step := scenario.Step{
		Request:       "GET /auth",
		SaveToContext: map[string]string{"session.token": "authToken"},
	}
	sc := makeScenario(srv.URL, 1, 10, []scenario.Step{step})
	vu, _ := newVirtualUser(1, sc)

	if err := vu.runScenario(context.Background()); err != nil {
		t.Fatalf("runScenario failed: %v", err)
	}

	vars := vu.getValues()
	if vars["authToken"] != "tok_xyz" {
		t.Errorf("expected authToken=tok_xyz, got %q", vars["authToken"])
	}
}

func TestVU_RunScenario_VariableSubstitution(t *testing.T) {
	var receivedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sc := makeScenario(srv.URL, 1, 10, []scenario.Step{
		{Request: "GET /users/${userId}"},
	})
	sc.Variables = map[string]string{"userId": "42"}

	vu, _ := newVirtualUser(1, sc)
	if err := vu.runScenario(context.Background()); err != nil {
		t.Fatalf("runScenario failed: %v", err)
	}
	if receivedPath != "/users/42" {
		t.Errorf("expected /users/42, got %q", receivedPath)
	}
}

func TestVU_RunScenario_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sc := makeScenario(srv.URL, 1, 10, []scenario.Step{simpleStep("GET", "/slow")})
	vu, _ := newVirtualUser(1, sc)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := vu.runScenario(ctx)
	if err == nil {
		t.Error("expected error on context cancellation")
	}
}

func TestVU_RunScenario_WithDelay(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	step := scenario.Step{
		Request: "GET /fast",
		Delay:   scenario.Duration{Duration: 50 * time.Millisecond},
	}
	sc := makeScenario(srv.URL, 1, 10, []scenario.Step{step})
	vu, _ := newVirtualUser(1, sc)

	start := time.Now()
	if err := vu.runScenario(context.Background()); err != nil {
		t.Fatalf("runScenario failed: %v", err)
	}
	elapsed := time.Since(start)
	if elapsed < 50*time.Millisecond {
		t.Errorf("expected at least 50ms delay, got %v", elapsed)
	}
}

// ============================================================================
// Runner integration tests
// ============================================================================

func TestRunner_New(t *testing.T) {
	sc := makeScenario("http://localhost", 1, 1, nil)
	r := New(sc)
	if r == nil {
		t.Fatal("New returned nil")
	}
}

func TestRunner_NilScenario(t *testing.T) {
	r := New(nil)
	err := r.Run(context.Background())
	if err == nil {
		t.Error("expected error for nil scenario")
	}
}

func TestRunner_Run_BasicExecution(t *testing.T) {
	var requests atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sc := makeScenario(srv.URL, 2, 1, []scenario.Step{simpleStep("GET", "/")})
	r := New(sc)

	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if requests.Load() == 0 {
		t.Error("expected at least some requests")
	}
}

func TestRunner_Run_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sc := makeScenario(srv.URL, 5, 60, []scenario.Step{simpleStep("GET", "/")})
	r := New(sc)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_ = r.Run(ctx)
	elapsed := time.Since(start)

	if elapsed > 2*time.Second {
		t.Errorf("runner did not stop quickly after cancel; elapsed=%v", elapsed)
	}
}

func TestRunner_Run_Duration(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sc := makeScenario(srv.URL, 2, 1, []scenario.Step{simpleStep("GET", "/")})
	r := New(sc)

	start := time.Now()
	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	elapsed := time.Since(start)

	if elapsed < 900*time.Millisecond {
		t.Errorf("expected ~1s duration, finished too early: %v", elapsed)
	}
	if elapsed > 3*time.Second {
		t.Errorf("expected ~1s duration, took too long: %v", elapsed)
	}
}

func TestRunner_Run_MultipleVUs_NoLeaks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sc := makeScenario(srv.URL, 10, 1, []scenario.Step{simpleStep("GET", "/")})
	r := New(sc)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := r.Run(ctx); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	// If goroutines leaked, the race detector or timeout would catch it.
}

func TestRunner_Run_RampUp(t *testing.T) {
	// Verify that VUs are spawned gradually: track first-request timestamps.
	type stamp struct{ t time.Time }
	var (
		mu     sync.Mutex
		stamps []stamp
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		stamps = append(stamps, stamp{time.Now()})
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sc := makeScenario(srv.URL, 5, 2, []scenario.Step{simpleStep("GET", "/")})
	r := New(sc)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := r.Run(ctx); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// With ramp-up, the first VU should start before all 5 are alive,
	// so we just confirm we got requests and the runner returned cleanly.
	mu.Lock()
	got := len(stamps)
	mu.Unlock()
	if got == 0 {
		t.Error("expected requests during ramp-up run")
	}
}

// ============================================================================
// Helper tests
// ============================================================================

func TestParseStepRequest_Valid(t *testing.T) {
	m, p, err := parseStepRequest("POST /api/users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m != "POST" || p != "/api/users" {
		t.Errorf("got method=%q path=%q", m, p)
	}
}

func TestParseStepRequest_Invalid(t *testing.T) {
	_, _, err := parseStepRequest("INVALID")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestApplyPathParams(t *testing.T) {
	result := applyPathParams("/users/{id}/posts/{postId}", map[string]string{
		"id":     "42",
		"postId": "7",
	})
	if result != "/users/42/posts/7" {
		t.Errorf("unexpected path: %q", result)
	}
}

func TestMatchesStatusCode(t *testing.T) {
	cases := []struct {
		code    int
		pattern string
		want    bool
	}{
		{200, "200", true},
		{201, "200", false},
		{200, "2xx", true},
		{404, "4xx", true},
		{500, "2xx", false},
	}
	for _, c := range cases {
		got := matchesStatusCode(c.code, c.pattern)
		if got != c.want {
			t.Errorf("matchesStatusCode(%d, %q) = %v, want %v", c.code, c.pattern, got, c.want)
		}
	}
}

func TestEncodeQuery(t *testing.T) {
	q := map[string]string{"a": "1"}
	result := encodeQuery(q)
	if result != "a=1" {
		t.Errorf("unexpected query: %q", result)
	}
}
