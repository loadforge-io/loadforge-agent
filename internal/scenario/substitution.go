package scenario

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// varPattern matches ${varName} placeholders.
var varPattern = regexp.MustCompile(`\${([^}]+)}`)

type Substitutor struct{}

func NewSubstitutor() *Substitutor {
	return &Substitutor{}
}

func substitute(s string, vars map[string]string) (string, error) {
	var firstErr error
	result := varPattern.ReplaceAllStringFunc(s, func(match string) string {
		if firstErr != nil {
			return match
		}
		name := match[2 : len(match)-1]
		val, ok := vars[name]
		if !ok {
			firstErr = fmt.Errorf("undefined variable %q", name)
			return match
		}
		return val
	})
	if firstErr != nil {
		return "", firstErr
	}
	return result, nil
}

// ApplyToURL substitutes variables in a URL path string (e.g. "/users/${user_id}").
func (s *Substitutor) ApplyToURL(url string, vars map[string]string) (string, error) {
	result, err := substitute(url, vars)
	if err != nil {
		return "", fmt.Errorf("url substitution failed: %w", err)
	}
	return result, nil
}

// ApplyToHeaders substitutes variables in header values.
func (s *Substitutor) ApplyToHeaders(headers map[string]string, vars map[string]string) (map[string]string, error) {
	result := make(map[string]string, len(headers))
	for k, v := range headers {
		replaced, err := substitute(v, vars)
		if err != nil {
			return nil, fmt.Errorf("header %q substitution failed: %w", k, err)
		}
		result[k] = replaced
	}
	return result, nil
}

// ApplyToQuery substitutes variables in query parameter values.
func (s *Substitutor) ApplyToQuery(query map[string]string, vars map[string]string) (map[string]string, error) {
	result := make(map[string]string, len(query))
	for k, v := range query {
		replaced, err := substitute(v, vars)
		if err != nil {
			return nil, fmt.Errorf("query param %q substitution failed: %w", k, err)
		}
		result[k] = replaced
	}
	return result, nil
}

// ApplyToBody substitutes variables in the request body.
func (s *Substitutor) ApplyToBody(body interface{}, vars map[string]string) (interface{}, error) {
	if body == nil {
		return nil, nil
	}

	if str, ok := body.(string); ok {
		result, err := substitute(str, vars)
		if err != nil {
			return nil, fmt.Errorf("body substitution failed: %w", err)
		}
		return result, nil
	}

	jsonVars := make(map[string]string, len(vars))
	for k, v := range vars {
		escaped, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to JSON-escape variable %q: %w", k, err)
		}
		jsonVars[k] = string(escaped[1 : len(escaped)-1])
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("body marshalling failed: %w", err)
	}

	substituted, err := substitute(string(raw), jsonVars)
	if err != nil {
		return nil, fmt.Errorf("body substitution failed: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal([]byte(substituted), &result); err != nil {
		return nil, fmt.Errorf("body unmarshalling after substitution failed: %w", err)
	}
	return result, nil
}

// ApplyToStep returns a copy of step with all ${var} placeholders resolved against vars
func (s *Substitutor) ApplyToStep(step Step, vars map[string]string) (Step, error) {
	result := step

	parts := strings.SplitN(step.Request, " ", 2)
	if len(parts) == 2 {
		substitutedPath, err := s.ApplyToURL(parts[1], vars)
		if err != nil {
			return Step{}, fmt.Errorf("request path substitution failed: %w", err)
		}
		result.Request = parts[0] + " " + substitutedPath
	}

	if step.Headers != nil {
		headers, err := s.ApplyToHeaders(step.Headers, vars)
		if err != nil {
			return Step{}, err
		}
		result.Headers = headers
	}

	if step.Query != nil {
		query, err := s.ApplyToQuery(step.Query, vars)
		if err != nil {
			return Step{}, err
		}
		result.Query = query
	}

	if step.PathParams != nil {
		pathParams, err := s.ApplyToQuery(step.PathParams, vars)
		if err != nil {
			return Step{}, fmt.Errorf("path_params substitution failed: %w", err)
		}
		result.PathParams = pathParams
	}

	if step.Body != nil {
		body, err := s.ApplyToBody(step.Body, vars)
		if err != nil {
			return Step{}, err
		}
		result.Body = body
	}

	return result, nil
}
