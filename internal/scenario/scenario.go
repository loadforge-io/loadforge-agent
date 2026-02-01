package scenario

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
	SaveToContext map[string]string `yaml:"save_to_context,omitempty"`
	NextSteps     []NextStep        `yaml:"next_steps,omitempty"`
}

type NextStep struct {
	Request     string            `yaml:"request"`
	StatusCodes []string          `yaml:"status_codes"`
	Map         map[string]string `yaml:"map,omitempty"`
}
