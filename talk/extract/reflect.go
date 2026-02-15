package extract

import (
	"context"
	"reflect"
	"strings"
	"unicode"

	"go.zoe.im/x/talk"
)

var (
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
)

func init() {
	talk.SetDefaultExtractor(NewReflectExtractor())
}

// ReflectExtractor extracts endpoints from a service using reflection.
type ReflectExtractor struct {
	opts Options
}

// NewReflectExtractor creates a new reflection-based extractor.
func NewReflectExtractor(opts ...Option) *ReflectExtractor {
	e := &ReflectExtractor{}
	for _, opt := range opts {
		opt(&e.opts)
	}
	return e
}

func (e *ReflectExtractor) Extract(service any) ([]*talk.Endpoint, error) {
	svcValue := reflect.ValueOf(service)
	svcType := svcValue.Type()

	var endpoints []*talk.Endpoint

	for i := 0; i < svcType.NumMethod(); i++ {
		method := svcType.Method(i)

		if !method.IsExported() {
			continue
		}

		endpoint, ok := e.extractMethod(svcValue, method)
		if !ok {
			continue
		}

		if e.opts.PathPrefix != "" {
			endpoint.Path = e.opts.PathPrefix + endpoint.Path
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}

func (e *ReflectExtractor) extractMethod(svcValue reflect.Value, method reflect.Method) (*talk.Endpoint, bool) {
	methodType := method.Type

	// Minimum: receiver, context -> (response, error)
	if methodType.NumIn() < 2 || methodType.NumOut() < 1 {
		return nil, false
	}

	// First param (after receiver) must be context.Context
	if !methodType.In(1).Implements(contextType) {
		return nil, false
	}

	// Last return must be error
	if !methodType.Out(methodType.NumOut() - 1).Implements(errorType) {
		return nil, false
	}

	name := method.Name
	httpMethod, path := e.deriveMethodAndPath(name, methodType)
	streamMode := e.detectStreamMode(methodType)

	// Check for explicit mapping
	if mapping, ok := e.opts.MethodMapping[name]; ok {
		if mapping.Path != "" {
			path = mapping.Path
		}
		if mapping.HTTPMethod != "" {
			httpMethod = mapping.HTTPMethod
		}
		if mapping.StreamMode != talk.StreamNone {
			streamMode = mapping.StreamMode
		}
	}

	endpoint := &talk.Endpoint{
		Name:       name,
		Path:       path,
		Method:     httpMethod,
		StreamMode: streamMode,
		Metadata:   make(map[string]any),
	}

	// Extract request/response types
	if methodType.NumIn() > 2 {
		endpoint.RequestType = methodType.In(2)
	}
	if methodType.NumOut() > 1 {
		respType := methodType.Out(0)
		if respType.Kind() == reflect.Chan {
			respType = respType.Elem()
		}
		endpoint.ResponseType = respType
	}

	// Create handler that wraps the method
	methodValue := svcValue.Method(method.Index)
	endpoint.Handler = e.createHandler(methodValue, methodType)

	return endpoint, true
}

func (e *ReflectExtractor) deriveMethodAndPath(name string, methodType reflect.Type) (httpMethod, path string) {
	resource := e.extractResource(name)
	hasIDParam := methodType.NumIn() > 2 && isSimpleType(methodType.In(2))

	switch {
	case strings.HasPrefix(name, "Get"):
		httpMethod = "GET"
		if hasIDParam {
			path = "/" + resource + "/{id}"
		} else {
			path = "/" + resource
		}
	case strings.HasPrefix(name, "List"):
		httpMethod = "GET"
		path = "/" + resource
	case strings.HasPrefix(name, "Create"):
		httpMethod = "POST"
		path = "/" + resource
	case strings.HasPrefix(name, "Update"):
		httpMethod = "PUT"
		path = "/" + resource + "/{id}"
	case strings.HasPrefix(name, "Delete"):
		httpMethod = "DELETE"
		path = "/" + resource + "/{id}"
	case strings.HasPrefix(name, "Watch"):
		httpMethod = "GET"
		path = "/" + resource + "/watch"
	default:
		httpMethod = "POST"
		path = "/" + toKebabCase(name)
	}

	return
}

func (e *ReflectExtractor) extractResource(methodName string) string {
	prefixes := []string{"Get", "List", "Create", "Update", "Delete", "Watch"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(methodName, prefix) {
			resource := methodName[len(prefix):]
			return strings.ToLower(resource)
		}
	}
	return strings.ToLower(methodName)
}

func (e *ReflectExtractor) detectStreamMode(methodType reflect.Type) talk.StreamMode {
	hasInputChan := false
	hasOutputChan := false

	// Check input params for chan (skip receiver and context)
	for i := 2; i < methodType.NumIn(); i++ {
		if methodType.In(i).Kind() == reflect.Chan {
			hasInputChan = true
			break
		}
	}

	// Check output for chan (skip error)
	for i := 0; i < methodType.NumOut()-1; i++ {
		if methodType.Out(i).Kind() == reflect.Chan {
			hasOutputChan = true
			break
		}
	}

	switch {
	case hasInputChan && hasOutputChan:
		return talk.StreamBidirect
	case hasInputChan:
		return talk.StreamClientSide
	case hasOutputChan:
		return talk.StreamServerSide
	default:
		return talk.StreamNone
	}
}

func (e *ReflectExtractor) createHandler(methodValue reflect.Value, methodType reflect.Type) talk.EndpointFunc {
	return func(ctx context.Context, request any) (any, error) {
		args := []reflect.Value{reflect.ValueOf(ctx)}

		// Add request argument if method expects one
		if methodType.NumIn() > 2 {
			if request != nil {
				args = append(args, reflect.ValueOf(request))
			} else {
				args = append(args, reflect.Zero(methodType.In(2)))
			}
		}

		results := methodValue.Call(args)

		var resp any
		var err error

		if len(results) > 0 {
			lastIdx := len(results) - 1
			if !results[lastIdx].IsNil() {
				err = results[lastIdx].Interface().(error)
			}
			if lastIdx > 0 && !results[0].IsNil() {
				resp = results[0].Interface()
			}
		}

		return resp, err
	}
}

func isSimpleType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func toKebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('-')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func pluralize(s string) string {
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") ||
		strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	if strings.HasSuffix(s, "y") && len(s) > 1 {
		c := s[len(s)-2]
		if c != 'a' && c != 'e' && c != 'i' && c != 'o' && c != 'u' {
			return s[:len(s)-1] + "ies"
		}
	}
	return s + "s"
}

// FromService extracts endpoints from a service using reflection.
func FromService(service any, opts ...Option) ([]*talk.Endpoint, error) {
	return NewReflectExtractor(opts...).Extract(service)
}
