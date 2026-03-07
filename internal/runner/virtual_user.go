package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"loadforge-agent/internal/executor"
	"loadforge-agent/internal/scenario"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

type VirtualUser struct {
	id       uint64
	scenario *scenario.Scenario

	vars map[string]string
	mu   sync.Mutex

	exec        *executor.Executor
	substitutor *scenario.Substitutor
}

func newVirtualUser(id uint64, sc *scenario.Scenario) (*VirtualUser, error) {
	exec, err := executor.New()
	if err != nil {
		return nil, fmt.Errorf("virtual user %d: failed to create executor: %w", id, err)
	}
	return &VirtualUser{
		id:          id,
		scenario:    sc,
		vars:        make(map[string]string),
		exec:        exec,
		substitutor: scenario.NewSubstitutor(),
	}, nil
}

func (v *VirtualUser) setValue(key, value string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.vars[key] = value
}

func (v *VirtualUser) getValues() map[string]string {
	v.mu.Lock()
	defer v.mu.Unlock()
	copiedVars := make(map[string]string, len(v.vars)+len(v.scenario.Variables))
	for key, value := range v.scenario.Variables {
		copiedVars[key] = value
	}
	for key, value := range v.vars {
		copiedVars[key] = value
	}
	return copiedVars
}

func (v *VirtualUser) runScenario(ctx context.Context) error {
	vars := v.getValues()

	for i := range v.scenario.Steps {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		step, err := v.substitutor.ApplyToStep(v.scenario.Steps[i], vars)
		if err != nil {
			return fmt.Errorf("vu %d step %d: substitution failed: %w", v.id, i, err)
		}

		if err := v.executeStep(ctx, &step, vars); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slog.Error("failed to execute step", "vu", v.id, "step", i, "err", err)
		}

		if !step.Delay.IsZero() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(step.Delay.Duration):
			}
		}
	}
	return nil
}

func (v *VirtualUser) executeStep(ctx context.Context, step *scenario.Step, vars map[string]string) error {
	method, path, err := parseStepRequest(step.Request)
	if err != nil {
		return err
	}

	url := v.scenario.BaseURL + applyPathParams(path, step.PathParams)
	if len(step.Query) > 0 {
		url += "?" + encodeQuery(step.Query)
	}

	var body []byte
	if step.Body != nil {
		body, err = json.Marshal(step.Body)
		if err != nil {
			return fmt.Errorf("failed to marshal body: %w", err)
		}
	}

	req := &executor.Request{
		Method:  method,
		URL:     url,
		Headers: step.Headers,
		Body:    body,
	}

	resp, err := v.exec.Execute(ctx, req)
	if err != nil {
		return fmt.Errorf("request %s failed: %w", step.Request, err)
	}

	for jsonPath, varName := range step.SaveToContext {
		if len(resp.Body) == 0 {
			continue
		}
		result := gjson.GetBytes(resp.Body, jsonPath)
		if result.Exists() {
			v.setValue(varName, result.String())
			vars[varName] = result.String()
		}
	}

	return nil
}

func parseStepRequest(request string) (method, path string, err error) {
	parts := strings.SplitN(request, " ", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid request format: %q", request)
	}
	return parts[0], parts[1], nil
}

func applyPathParams(path string, params map[string]string) string {
	for k, v := range params {
		path = strings.ReplaceAll(path, "{"+k+"}", v)
	}
	return path
}

func encodeQuery(q map[string]string) string {
	var buf bytes.Buffer
	first := true
	for k, v := range q {
		if !first {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(v)
		first = false
	}
	return buf.String()
}

func statusClass(code int) string {
	return fmt.Sprintf("%dxx", code/100)
}

func matchesStatusCode(code int, pattern string) bool {
	if len(pattern) == 3 && pattern[1:] == "xx" {
		return statusClass(code) == pattern
	}
	return fmt.Sprintf("%d", code) == pattern
}

func matchesAnyStatus(code int, patterns []string) bool {
	for _, p := range patterns {
		if matchesStatusCode(code, p) {
			return true
		}
	}
	return false
}

func resolveNextStep(step *scenario.Step, respCode int) *scenario.NextStep {
	for i := range step.NextSteps {
		ns := &step.NextSteps[i]
		if len(ns.StatusCodes) == 0 || matchesAnyStatus(respCode, ns.StatusCodes) {
			return ns
		}
	}
	return nil
}

func httpStatusFromResponse(resp *executor.Response) int {
	if resp == nil {
		return http.StatusInternalServerError
	}
	return resp.StatusCode
}
