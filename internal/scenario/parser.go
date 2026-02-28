package scenario

import (
	"fmt"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Parser struct {
	scenario *Scenario
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	return p.ParseData(data)
}

func (p *Parser) ParseData(data []byte) error {
	var scenario Scenario
	if err := yaml.Unmarshal(data, &scenario); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	p.scenario = &scenario
	return nil
}

func (p *Parser) GetScenario() (*Scenario, error) {
	if p.scenario == nil {
		return nil, fmt.Errorf("no scenario loaded")
	}
	return p.scenario, nil
}

func (s *Scenario) FindStep(request string) *Step {
	for i := range s.Steps {
		if s.Steps[i].Request == request {
			return &s.Steps[i]
		}
	}
	return nil
}

const maxDelay = 10 * time.Minute

func (p *Parser) Validate() error {
	if p.scenario == nil {
		return fmt.Errorf("no scenario loaded")
	}

	if p.scenario.Name == "" {
		return fmt.Errorf("scenario.name is required")
	}

	if p.scenario.BaseURL == "" {
		return fmt.Errorf("scenario.base_url is required")
	}

	if p.scenario.VirtualUsers <= 0 {
		return fmt.Errorf("scenario.virtual_users must be greater than 0")
	}

	if p.scenario.Duration <= 0 {
		return fmt.Errorf("scenario.duration must be greater than 0")
	}

	if p.scenario.Duration > 31_556_952 {
		return fmt.Errorf("scenario.duration must be less than 1 year (31556952 seconds)")
	}

	if len(p.scenario.Steps) == 0 {
		return fmt.Errorf("scenario.steps: at least one step is required")
	}

	uniqueRequests := make(map[string]struct{})

	for i := range p.scenario.Steps {
		step := &p.scenario.Steps[i]

		if step.Request == "" {
			return fmt.Errorf("step[%d]: request field is required", i)
		}

		if _, exists := uniqueRequests[step.Request]; exists {
			return fmt.Errorf("step[%d]: duplicate request '%s'", i, step.Request)
		}
		uniqueRequests[step.Request] = struct{}{}

		httpMethod, _, err := parseRequest(step.Request)
		if err != nil {
			return fmt.Errorf("step[%d]: %w", i, err)
		}

		if (httpMethod == http.MethodGet || httpMethod == http.MethodHead) &&
			step.Body != nil {
			return fmt.Errorf("step[%d] (%s): GET and HEAD requests cannot have a body",
				i, step.Request)
		}

		if step.Delay.Duration < 0 {
			return fmt.Errorf("step[%d] (%s): delay must be non-negative", i, step.Request)
		}

		if step.Delay.Duration > maxDelay {
			return fmt.Errorf("step[%d] (%s): delay must not exceed %s", i, step.Request, maxDelay)
		}

		for j := range step.NextSteps {
			nextStep := &step.NextSteps[j]

			if nextStep.Request == "" {
				return fmt.Errorf("step[%d], next_step[%d]: request field is required", i, j)
			}

			_, _, err := parseRequest(nextStep.Request)
			if err != nil {
				return fmt.Errorf("step[%d], next_step[%d]: %w", i, j, err)
			}

			targetStep := p.scenario.FindStep(nextStep.Request)
			if targetStep == nil {
				return fmt.Errorf("step[%d], next_step[%d]: target step '%s' not found",
					i, j, nextStep.Request)
			}

			for k, code := range nextStep.StatusCodes {
				if err := validateStatusCode(code); err != nil {
					return fmt.Errorf("step[%d], next_step[%d], status_code[%d]: %w",
						i, j, k, err)
				}
			}

			for mapSource, mapTarget := range nextStep.Map {
				if err := validateMapping(mapSource, mapTarget); err != nil {
					return fmt.Errorf("step[%d], next_step[%d]: invalid mapping '%s' -> '%s': %w",
						i, j, mapSource, mapTarget, err)
				}
			}
		}
	}

	return nil
}

func parseRequest(request string) (method string, path string, err error) {
	if request == "" {
		return "", "", fmt.Errorf("request cannot be empty")
	}

	parts := strings.SplitN(request, " ", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid request format '%s', expected 'METHOD /path'", request)
	}

	method = parts[0]
	path = parts[1]

	validMethods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodHead,
	}

	if !slices.Contains(validMethods, method) {
		return "", "", fmt.Errorf("invalid HTTP method '%s', must be one of: %v",
			method, validMethods)
	}

	if !strings.HasPrefix(path, "/") {
		return "", "", fmt.Errorf("path must start with '/', got: %s", path)
	}

	return method, path, nil
}

func validateStatusCode(code string) error {
	if code == "" {
		return fmt.Errorf("status code cannot be empty")
	}

	if len(code) == 3 && code[1:] == "xx" {
		firstChar := code[0]
		if firstChar < '1' || firstChar > '5' {
			return fmt.Errorf("wildcard must be 1xx-5xx, got: %s", code)
		}
		return nil
	}

	statusCode, err := strconv.Atoi(code)
	if err != nil {
		return fmt.Errorf("invalid status code format '%s'", code)
	}

	if statusCode < 100 || statusCode >= 600 {
		return fmt.Errorf("status code must be 100-599, got: %d", statusCode)
	}

	return nil
}

func validateMapping(source, target string) error {
	validSources := []string{"response", "headers", "query", "path_params", "body", "cookies", "variables"}
	validTargets := []string{"headers", "query", "path_params", "body", "cookies", "variables"}

	sourceParts := strings.SplitN(source, ".", 2)
	if len(sourceParts) != 2 {
		return fmt.Errorf("invalid source format, expected 'source.field', got: %s", source)
	}

	if !slices.Contains(validSources, sourceParts[0]) {
		return fmt.Errorf("invalid source '%s', must be one of: %v",
			sourceParts[0], validSources)
	}

	targetParts := strings.SplitN(target, ".", 2)
	if len(targetParts) != 2 {
		return fmt.Errorf("invalid target format, expected 'target.field', got: %s", target)
	}

	if !slices.Contains(validTargets, targetParts[0]) {
		return fmt.Errorf("invalid target '%s', must be one of: %v",
			targetParts[0], validTargets)
	}

	return nil
}
