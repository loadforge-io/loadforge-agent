package scenario

import (
	"fmt"
	"time"
)

type Scenario struct {
	Name         string            `yaml:"name"`
	BaseURL      string            `yaml:"base_url"`
	VirtualUsers uint64            `yaml:"virtual_users"`
	Duration     uint64            `yaml:"duration"`
	Variables    map[string]string `yaml:"variables,omitempty"`
	Steps        []Step            `yaml:"steps"`
}

type Step struct {
	Request       string            `yaml:"request"`
	Headers       map[string]string `yaml:"headers,omitempty"`
	Query         map[string]string `yaml:"query,omitempty"`
	PathParams    map[string]string `yaml:"path_params,omitempty"`
	Body          interface{}       `yaml:"body,omitempty"`
	Delay         Duration          `yaml:"delay,omitempty"`
	SaveToContext map[string]string `yaml:"save_to_context,omitempty"`
	NextSteps     []NextStep        `yaml:"next_steps,omitempty"`
}

type NextStep struct {
	Request     string            `yaml:"request"`
	StatusCodes []string          `yaml:"status_codes"`
	Map         map[string]string `yaml:"map,omitempty"`
}

type Duration struct {
	time.Duration
}

func (d *Duration) IsZero() bool {
	return d.Duration == 0
}

func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw string
	if err := unmarshal(&raw); err != nil {
		var ns int64
		if err2 := unmarshal(&ns); err2 != nil {
			return fmt.Errorf("delay must be a duration string (e.g. '2s', '500ms'): %w", err)
		}
		d.Duration = time.Duration(ns)
		return nil
	}

	if raw == "" {
		d.Duration = 0
		return nil
	}

	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return fmt.Errorf("invalid delay %q: %w", raw, err)
	}
	d.Duration = parsed
	return nil
}

func (d *Duration) MarshalYAML() (interface{}, error) {
	return d.Duration.String(), nil
}
