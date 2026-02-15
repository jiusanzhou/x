package talk

import (
	"context"
	"reflect"
	"regexp"
	"strings"
	"unicode"
)

var (
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
)

// MethodAnnotations allows services to provide custom endpoint configuration.
type MethodAnnotations interface {
	TalkAnnotations() map[string]string
}

type reflectExtractor struct{}

func (e *reflectExtractor) Extract(service any) ([]*Endpoint, error) {
	svcValue := reflect.ValueOf(service)
	svcType := svcValue.Type()

	var annotations map[string]string
	if ma, ok := service.(MethodAnnotations); ok {
		annotations = ma.TalkAnnotations()
	}

	var endpoints []*Endpoint

	for i := 0; i < svcType.NumMethod(); i++ {
		method := svcType.Method(i)

		if !method.IsExported() {
			continue
		}

		var ann *annotation
		if annotations != nil {
			if comment, ok := annotations[method.Name]; ok {
				ann = parseAnnotation(comment)
				if ann != nil && ann.skip {
					continue
				}
			}
		}

		endpoint, ok := e.extractMethod(svcValue, method, ann)
		if !ok {
			continue
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}

func (e *reflectExtractor) extractMethod(svcValue reflect.Value, method reflect.Method, ann *annotation) (*Endpoint, bool) {
	methodType := method.Type

	if methodType.NumIn() < 2 || methodType.NumOut() < 1 {
		return nil, false
	}

	if !methodType.In(1).Implements(contextType) {
		return nil, false
	}

	if !methodType.Out(methodType.NumOut() - 1).Implements(errorType) {
		return nil, false
	}

	name := method.Name
	httpMethod, path := e.deriveMethodAndPath(name, methodType)
	streamMode := e.detectStreamMode(methodType)

	if ann != nil {
		if ann.path != "" {
			path = ann.path
		}
		if ann.method != "" {
			httpMethod = ann.method
		}
		if ann.streamMode != StreamNone {
			streamMode = ann.streamMode
		}
	}

	endpoint := &Endpoint{
		Name:       name,
		Path:       path,
		Method:     httpMethod,
		StreamMode: streamMode,
		Metadata:   make(map[string]any),
	}

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

	methodValue := svcValue.Method(method.Index)

	if streamMode != StreamNone {
		endpoint.StreamHandler = e.createStreamHandler(methodValue, methodType)
	} else {
		endpoint.Handler = e.createHandler(methodValue, methodType)
	}

	return endpoint, true
}

func (e *reflectExtractor) deriveMethodAndPath(name string, methodType reflect.Type) (httpMethod, path string) {
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

func (e *reflectExtractor) extractResource(methodName string) string {
	prefixes := []string{"Get", "List", "Create", "Update", "Delete", "Watch"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(methodName, prefix) {
			resource := methodName[len(prefix):]
			return strings.ToLower(resource)
		}
	}
	return strings.ToLower(methodName)
}

func (e *reflectExtractor) detectStreamMode(methodType reflect.Type) StreamMode {
	hasInputChan := false
	hasOutputChan := false

	for i := 2; i < methodType.NumIn(); i++ {
		if methodType.In(i).Kind() == reflect.Chan {
			hasInputChan = true
			break
		}
	}

	for i := 0; i < methodType.NumOut()-1; i++ {
		if methodType.Out(i).Kind() == reflect.Chan {
			hasOutputChan = true
			break
		}
	}

	switch {
	case hasInputChan && hasOutputChan:
		return StreamBidirect
	case hasInputChan:
		return StreamClientSide
	case hasOutputChan:
		return StreamServerSide
	default:
		return StreamNone
	}
}

func (e *reflectExtractor) createHandler(methodValue reflect.Value, methodType reflect.Type) EndpointFunc {
	return func(ctx context.Context, request any) (any, error) {
		args := []reflect.Value{reflect.ValueOf(ctx)}

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

func (e *reflectExtractor) createStreamHandler(methodValue reflect.Value, methodType reflect.Type) StreamEndpointFunc {
	return func(ctx context.Context, request any, stream Stream) error {
		args := []reflect.Value{reflect.ValueOf(ctx)}

		if methodType.NumIn() > 2 {
			if request != nil {
				args = append(args, reflect.ValueOf(request))
			} else {
				args = append(args, reflect.Zero(methodType.In(2)))
			}
		}

		results := methodValue.Call(args)

		if len(results) == 0 {
			return nil
		}

		lastIdx := len(results) - 1
		if !results[lastIdx].IsNil() {
			return results[lastIdx].Interface().(error)
		}

		if lastIdx > 0 && results[0].Kind() == reflect.Chan {
			ch := results[0]
			ctxDone := reflect.ValueOf(ctx.Done())

			cases := []reflect.SelectCase{
				{Dir: reflect.SelectRecv, Chan: ctxDone},
				{Dir: reflect.SelectRecv, Chan: ch},
			}

			for {
				chosen, val, ok := reflect.Select(cases)
				if chosen == 0 {
					return ctx.Err()
				}
				if !ok {
					break
				}
				if err := stream.Send(val.Interface()); err != nil {
					return err
				}
			}
		}

		return nil
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

type annotation struct {
	path       string
	method     string
	streamMode StreamMode
	skip       bool
}

var annotationRegex = regexp.MustCompile(`@talk\s*(.*)`)
var kvRegex = regexp.MustCompile(`(\w+)=([^\s]+)`)

func parseAnnotation(comment string) *annotation {
	match := annotationRegex.FindStringSubmatch(comment)
	if match == nil {
		return nil
	}

	ann := &annotation{}
	content := strings.TrimSpace(match[1])

	if content == "skip" || content == "ignore" || content == "-" {
		ann.skip = true
		return ann
	}

	kvMatches := kvRegex.FindAllStringSubmatch(content, -1)
	for _, kv := range kvMatches {
		key := strings.ToLower(kv[1])
		value := kv[2]

		switch key {
		case "path":
			ann.path = value
		case "method":
			ann.method = strings.ToUpper(value)
		case "stream":
			ann.streamMode = parseStreamMode(value)
		case "skip", "ignore":
			ann.skip = value == "true" || value == "1"
		}
	}

	return ann
}

func parseStreamMode(s string) StreamMode {
	switch strings.ToLower(s) {
	case "server", "server-side", "sse":
		return StreamServerSide
	case "client", "client-side":
		return StreamClientSide
	case "bidi", "bidirectional", "duplex":
		return StreamBidirect
	default:
		return StreamNone
	}
}

func init() {
	defaultExtractor = &reflectExtractor{}
}
