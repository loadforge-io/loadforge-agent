package extractor

import (
	"fmt"

	"github.com/tidwall/gjson"
)

type Extractor struct{}

func New() *Extractor {
	return &Extractor{}
}

// Extract extracts a value from JSON data using a JSONPath expression
// It uses gjson syntax which is similar to JSONPath but simpler
// Examples:
//   - "user.id" extracts the id field from user object
//   - "users.0.name" extracts the name of the first user
//   - "users.#.name" extracts all user names as an array
func (e *Extractor) Extract(jsonData []byte, path string) (any, error) {
	if len(jsonData) == 0 {
		return nil, fmt.Errorf("json data cannot be empty")
	}

	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	result := gjson.GetBytes(jsonData, path)

	if !result.Exists() {
		return nil, fmt.Errorf("path '%s' not found in JSON", path)
	}

	return result.Value(), nil
}

func (e *Extractor) Exists(jsonData []byte, path string) bool {
	if len(jsonData) == 0 || path == "" {
		return false
	}

	result := gjson.GetBytes(jsonData, path)
	return result.Exists()
}

func (e *Extractor) ExtractWithDefault(jsonData []byte, path string, defaultValue any) any {
	value, err := e.Extract(jsonData, path)
	if err != nil {
		return defaultValue
	}
	return value
}
