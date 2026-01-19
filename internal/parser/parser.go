package parser

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

type Endpoint struct {
	Path      string
	Method    string
	Tags      []string
	Responses any

	// optional
	OperationID string
	Summary     string
	Description string
}

type Parser struct {
	doc *openapi3.T
}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) ParseFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromData(data)
	if err != nil {
		return fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	ctx := context.Background()
	if err := doc.Validate(ctx); err != nil {
		return fmt.Errorf("invalid OpenAPI spec: %w", err)
	}

	p.doc = doc
	return nil
}

// ParseData loads and parses an OpenAPI specification from raw data
func (p *Parser) ParseData(data []byte) error {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromData(data)
	if err != nil {
		return fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	ctx := context.Background()
	if err := doc.Validate(ctx); err != nil {
		return fmt.Errorf("invalid OpenAPI spec: %w", err)
	}

	p.doc = doc
	return nil
}

// GetEndpoints extracts all endpoints from the parsed specification
func (p *Parser) GetEndpoints() ([]Endpoint, error) {
	if p.doc == nil {
		return nil, fmt.Errorf("no document loaded")
	}

	if p.doc.Paths == nil {
		return []Endpoint{}, nil
	}

	var endpoints []Endpoint

	for path, pathItem := range p.doc.Paths.Map() {
		if pathItem == nil {
			continue
		}

		operations := map[string]*openapi3.Operation{
			"GET":     pathItem.Get,
			"POST":    pathItem.Post,
			"PUT":     pathItem.Put,
			"PATCH":   pathItem.Patch,
			"DELETE":  pathItem.Delete,
			"HEAD":    pathItem.Head,
			"OPTIONS": pathItem.Options,
			"TRACE":   pathItem.Trace,
		}

		for method, operation := range operations {
			if operation == nil {
				continue
			}

			endpoint := Endpoint{
				Path:        path,
				Method:      method,
				OperationID: operation.OperationID,
				Summary:     operation.Summary,
				Description: operation.Description,
				Tags:        operation.Tags,
			}

			endpoints = append(endpoints, endpoint)
		}
	}

	return endpoints, nil
}

// GetEndpointsByTag filters endpoints by tag
func (p *Parser) GetEndpointsByTag(tag string) ([]Endpoint, error) {
	allEndpoints, err := p.GetEndpoints()
	if err != nil {
		return nil, err
	}

	var filtered []Endpoint
	for _, endpoint := range allEndpoints {
		if slices.Contains(endpoint.Tags, tag) {
			filtered = append(filtered, endpoint)
		}
	}

	return filtered, nil
}

// GetTags returns all unique tags from the specification
func (p *Parser) GetTags() ([]string, error) {
	if p.doc == nil {
		return nil, fmt.Errorf("no document loaded")
	}

	tagSet := make(map[string]bool)

	if p.doc.Tags != nil {
		for _, tag := range p.doc.Tags {
			if tag.Name != "" {
				tagSet[tag.Name] = true
			}
		}
	}

	endpoints, err := p.GetEndpoints()
	if err != nil {
		return nil, err
	}

	for _, endpoint := range endpoints {
		for _, tag := range endpoint.Tags {
			tagSet[tag] = true
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	return tags, nil
}

// GetInfo returns basic information about the API
func (p *Parser) GetInfo() (*openapi3.Info, error) {
	if p.doc == nil {
		return nil, fmt.Errorf("no document loaded")
	}

	return p.doc.Info, nil
}

// GetServerURLs returns all server URLs defined in the specification
func (p *Parser) GetServerURLs() ([]string, error) {
	if p.doc == nil {
		return nil, fmt.Errorf("no document loaded")
	}

	if p.doc.Servers == nil {
		return []string{}, nil
	}

	urls := make([]string, 0, len(p.doc.Servers))
	for _, server := range p.doc.Servers {
		urls = append(urls, server.URL)
	}

	return urls, nil
}
