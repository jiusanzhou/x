// Package swagger provides OpenAPI/Swagger documentation generation for talk endpoints.
package swagger

import (
	"encoding/json"
	"reflect"
	"strings"

	"go.zoe.im/x/talk"
)

// Config holds configuration for Swagger documentation.
type Config struct {
	// Enabled indicates whether Swagger documentation is enabled.
	Enabled bool `json:"enabled,omitempty" yaml:"enabled"`
	// Path is the base path for Swagger endpoints (default: /swagger).
	Path string `json:"path,omitempty" yaml:"path"`
	// Title is the API title.
	Title string `json:"title,omitempty" yaml:"title"`
	// Description is the API description.
	Description string `json:"description,omitempty" yaml:"description"`
	// Version is the API version.
	Version string `json:"version,omitempty" yaml:"version"`
	// BasePath is the base path for all API endpoints.
	BasePath string `json:"base_path,omitempty" yaml:"base_path"`
	// Host is the API host (e.g., "localhost:8080").
	Host string `json:"host,omitempty" yaml:"host"`
	// Schemes are the supported schemes (e.g., ["http", "https"]).
	Schemes []string `json:"schemes,omitempty" yaml:"schemes"`
}

// DefaultConfig returns a default swagger configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:     false,
		Path:        "/swagger",
		Title:       "API Documentation",
		Description: "Auto-generated API documentation",
		Version:     "1.0.0",
		BasePath:    "/",
		Schemes:     []string{"http"},
	}
}

// OpenAPI represents an OpenAPI 3.0 specification.
type OpenAPI struct {
	OpenAPI    string              `json:"openapi"`
	Info       Info                `json:"info"`
	Servers    []Server            `json:"servers,omitempty"`
	Paths      map[string]PathItem `json:"paths"`
	Components *Components         `json:"components,omitempty"`
	Tags       []Tag               `json:"tags,omitempty"`
}

// Info represents the info section of an OpenAPI spec.
type Info struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

// Server represents a server in an OpenAPI spec.
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// PathItem represents a path in an OpenAPI spec.
type PathItem struct {
	Get     *Operation `json:"get,omitempty"`
	Post    *Operation `json:"post,omitempty"`
	Put     *Operation `json:"put,omitempty"`
	Delete  *Operation `json:"delete,omitempty"`
	Patch   *Operation `json:"patch,omitempty"`
	Options *Operation `json:"options,omitempty"`
}

// Operation represents an operation in an OpenAPI spec.
type Operation struct {
	Tags        []string            `json:"tags,omitempty"`
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
	OperationID string              `json:"operationId,omitempty"`
	Parameters  []Parameter         `json:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"`
}

// Parameter represents a parameter in an OpenAPI spec.
type Parameter struct {
	Name        string  `json:"name"`
	In          string  `json:"in"` // query, header, path, cookie
	Description string  `json:"description,omitempty"`
	Required    bool    `json:"required,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`
}

// RequestBody represents a request body in an OpenAPI spec.
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required,omitempty"`
	Content     map[string]MediaType `json:"content"`
}

// Response represents a response in an OpenAPI spec.
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// MediaType represents a media type in an OpenAPI spec.
type MediaType struct {
	Schema *Schema `json:"schema,omitempty"`
}

// Schema represents a schema in an OpenAPI spec.
type Schema struct {
	Type        string             `json:"type,omitempty"`
	Format      string             `json:"format,omitempty"`
	Description string             `json:"description,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Ref         string             `json:"$ref,omitempty"`
}

// Components represents the components section of an OpenAPI spec.
type Components struct {
	Schemas map[string]*Schema `json:"schemas,omitempty"`
}

// Tag represents a tag in an OpenAPI spec.
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Generator generates OpenAPI specs from talk endpoints.
type Generator struct {
	config  Config
	schemas map[string]*Schema
}

// NewGenerator creates a new OpenAPI generator.
func NewGenerator(cfg Config) *Generator {
	return &Generator{
		config:  cfg,
		schemas: make(map[string]*Schema),
	}
}

// Generate generates an OpenAPI spec from the given endpoints.
func (g *Generator) Generate(endpoints []*talk.Endpoint) *OpenAPI {
	spec := &OpenAPI{
		OpenAPI: "3.0.3",
		Info: Info{
			Title:       g.config.Title,
			Description: g.config.Description,
			Version:     g.config.Version,
		},
		Paths: make(map[string]PathItem),
	}

	// Add servers if host is specified
	if g.config.Host != "" {
		for _, scheme := range g.config.Schemes {
			spec.Servers = append(spec.Servers, Server{
				URL: scheme + "://" + g.config.Host + g.config.BasePath,
			})
		}
	}

	// Process each endpoint
	tags := make(map[string]bool)
	for _, ep := range endpoints {
		g.addEndpoint(spec, ep)

		// Extract tag from path
		tag := g.extractTag(ep.Path)
		if tag != "" {
			tags[tag] = true
		}
	}

	// Add tags
	for tag := range tags {
		spec.Tags = append(spec.Tags, Tag{Name: tag})
	}

	// Add component schemas
	if len(g.schemas) > 0 {
		spec.Components = &Components{
			Schemas: g.schemas,
		}
	}

	return spec
}

// GenerateJSON generates an OpenAPI spec as JSON.
func (g *Generator) GenerateJSON(endpoints []*talk.Endpoint) ([]byte, error) {
	spec := g.Generate(endpoints)
	return json.MarshalIndent(spec, "", "  ")
}

func (g *Generator) addEndpoint(spec *OpenAPI, ep *talk.Endpoint) {
	path := g.normalizePath(ep.Path)

	op := &Operation{
		Summary:     ep.Name,
		OperationID: strings.ToLower(ep.Name),
		Responses: map[string]Response{
			"200": {
				Description: "Successful response",
			},
		},
	}

	// Add tag
	tag := g.extractTag(ep.Path)
	if tag != "" {
		op.Tags = []string{tag}
	}

	// Extract path parameters
	params := g.extractPathParams(ep.Path)
	op.Parameters = append(op.Parameters, params...)

	// Add request body for POST/PUT/PATCH
	if ep.RequestType != nil && (ep.Method == "POST" || ep.Method == "PUT" || ep.Method == "PATCH") {
		schema := g.typeToSchema(ep.RequestType)
		op.RequestBody = &RequestBody{
			Required: true,
			Content: map[string]MediaType{
				"application/json": {
					Schema: schema,
				},
			},
		}
	}

	// Add response schema
	if ep.ResponseType != nil {
		schema := g.typeToSchema(ep.ResponseType)
		op.Responses["200"] = Response{
			Description: "Successful response",
			Content: map[string]MediaType{
				"application/json": {
					Schema: schema,
				},
			},
		}
	}

	// Handle streaming endpoints
	if ep.IsStreaming() {
		op.Description = "Streaming endpoint (" + ep.StreamMode.String() + ")"
		if ep.StreamMode == talk.StreamServerSide {
			op.Responses["200"] = Response{
				Description: "Server-sent events stream",
				Content: map[string]MediaType{
					"text/event-stream": {
						Schema: &Schema{Type: "string"},
					},
				},
			}
		}
	}

	// Add to path item
	pathItem, exists := spec.Paths[path]
	if !exists {
		pathItem = PathItem{}
	}

	switch ep.Method {
	case "GET":
		pathItem.Get = op
	case "POST":
		pathItem.Post = op
	case "PUT":
		pathItem.Put = op
	case "DELETE":
		pathItem.Delete = op
	case "PATCH":
		pathItem.Patch = op
	default:
		pathItem.Post = op
	}

	spec.Paths[path] = pathItem
}

func (g *Generator) normalizePath(path string) string {
	// Convert {param} to OpenAPI style (already correct)
	return path
}

func (g *Generator) extractTag(path string) string {
	// Extract first segment after leading slash
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return ""
}

func (g *Generator) extractPathParams(path string) []Parameter {
	var params []Parameter

	// Find all {param} patterns
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			name := strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}")
			params = append(params, Parameter{
				Name:     name,
				In:       "path",
				Required: true,
				Schema: &Schema{
					Type: "string",
				},
			})
		}
	}

	return params
}

func (g *Generator) typeToSchema(t reflect.Type) *Schema {
	if t == nil {
		return nil
	}

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check if we already have this schema
	typeName := t.Name()
	if typeName != "" {
		if _, exists := g.schemas[typeName]; exists {
			return &Schema{Ref: "#/components/schemas/" + typeName}
		}
	}

	schema := g.buildSchema(t)

	// Register named types as components
	if typeName != "" && t.Kind() == reflect.Struct {
		g.schemas[typeName] = schema
		return &Schema{Ref: "#/components/schemas/" + typeName}
	}

	return schema
}

func (g *Generator) buildSchema(t reflect.Type) *Schema {
	switch t.Kind() {
	case reflect.Bool:
		return &Schema{Type: "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return &Schema{Type: "integer", Format: "int32"}
	case reflect.Int64:
		return &Schema{Type: "integer", Format: "int64"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return &Schema{Type: "integer", Format: "int32"}
	case reflect.Uint64:
		return &Schema{Type: "integer", Format: "int64"}
	case reflect.Float32:
		return &Schema{Type: "number", Format: "float"}
	case reflect.Float64:
		return &Schema{Type: "number", Format: "double"}
	case reflect.String:
		return &Schema{Type: "string"}
	case reflect.Slice, reflect.Array:
		return &Schema{
			Type:  "array",
			Items: g.typeToSchema(t.Elem()),
		}
	case reflect.Map:
		return &Schema{
			Type: "object",
		}
	case reflect.Struct:
		return g.structToSchema(t)
	case reflect.Interface:
		return &Schema{Type: "object"}
	default:
		return &Schema{Type: "object"}
	}
}

func (g *Generator) structToSchema(t reflect.Type) *Schema {
	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON tag name
		name := field.Name
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				name = parts[0]
			} else if parts[0] == "-" {
				continue
			}
		}

		fieldSchema := g.typeToSchema(field.Type)
		if fieldSchema != nil {
			schema.Properties[name] = fieldSchema
		}

		// Check for required fields
		if !strings.Contains(jsonTag, "omitempty") {
			schema.Required = append(schema.Required, name)
		}
	}

	return schema
}
