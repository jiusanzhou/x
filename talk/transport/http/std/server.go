// Package std provides a net/http based transport implementation.
package std

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.zoe.im/x"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/codec"
	"go.zoe.im/x/talk/swagger"
	thttp "go.zoe.im/x/talk/transport/http"
)

type Server struct {
	config         thttp.ServerConfig
	codec          codec.Codec
	server         *http.Server
	mux            *http.ServeMux
	endpoints      []*talk.Endpoint
	swaggerHandler *swagger.Handler
	externalMux    bool
	externalServer bool
}

func NewServer(cfg x.TypedLazyConfig, opts ...thttp.Option) (*Server, error) {
	s := &Server{}

	if err := cfg.Unmarshal(&s.config); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.codec == nil {
		s.codec = codec.MustGet("json")
	}

	if s.mux == nil {
		s.mux = http.NewServeMux()
	}

	if s.config.Swagger.Enabled {
		swaggerCfg := s.config.Swagger
		if swaggerCfg.Path == "" {
			swaggerCfg.Path = "/swagger"
		}
		if swaggerCfg.Title == "" {
			swaggerCfg.Title = "API Documentation"
		}
		if swaggerCfg.Version == "" {
			swaggerCfg.Version = "1.0.0"
		}
		if swaggerCfg.Host == "" {
			swaggerCfg.Host = s.config.Addr
		}
		s.swaggerHandler = swagger.NewHandler(swaggerCfg)
	}

	return s, nil
}

func WithServeMux(mux *http.ServeMux) thttp.Option {
	return func(v any) {
		if s, ok := v.(*Server); ok {
			s.mux = mux
			s.externalMux = true
		}
	}
}

func WithHTTPServer(server *http.Server) thttp.Option {
	return func(v any) {
		if s, ok := v.(*Server); ok {
			s.server = server
			s.externalServer = true
		}
	}
}

func (s *Server) SetCodec(c codec.Codec) {
	s.codec = c
}

func (s *Server) String() string {
	return "http/std"
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) ServeMux() *http.ServeMux {
	return s.mux
}

func (s *Server) RegisterEndpoints(endpoints []*talk.Endpoint) {
	s.endpoints = endpoints
	for _, ep := range endpoints {
		s.registerEndpoint(ep)
	}
	if s.swaggerHandler != nil {
		s.swaggerHandler.SetEndpoints(endpoints)
		s.mux.Handle(s.swaggerHandler.BasePath()+"/", s.swaggerHandler)
	}
}

func (s *Server) Serve(ctx context.Context, endpoints []*talk.Endpoint) error {
	s.RegisterEndpoints(endpoints)

	if s.externalMux {
		<-ctx.Done()
		return nil
	}

	if s.server == nil {
		s.server = &http.Server{
			Addr:         s.config.Addr,
			Handler:      s.mux,
			ReadTimeout:  time.Duration(s.config.ReadTimeout),
			WriteTimeout: time.Duration(s.config.WriteTimeout),
			IdleTimeout:  time.Duration(s.config.IdleTimeout),
		}
	}

	if s.externalServer {
		<-ctx.Done()
		return nil
	}

	errCh := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		return s.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

func (s *Server) Invoke(ctx context.Context, endpoint string, req any, resp any) error {
	return talk.NewError(talk.Unimplemented, "server does not support Invoke")
}

func (s *Server) InvokeStream(ctx context.Context, endpoint string, req any) (talk.Stream, error) {
	return nil, talk.NewError(talk.Unimplemented, "server does not support InvokeStream")
}

func (s *Server) Close() error {
	return nil
}

func (s *Server) registerEndpoint(ep *talk.Endpoint) {
	pattern := s.buildPattern(ep)

	handler := s.createHandler(ep)
	s.mux.HandleFunc(pattern, handler)
}

func (s *Server) buildPattern(ep *talk.Endpoint) string {
	path := ep.Path

	// Convert {param} to Go 1.22+ pattern syntax
	path = convertPathParams(path)

	// Add method prefix for Go 1.22+ routing
	if ep.Method != "" {
		return ep.Method + " " + path
	}
	return path
}

func (s *Server) createHandler(ep *talk.Endpoint) http.HandlerFunc {
	if ep.IsStreaming() && ep.StreamMode == talk.StreamServerSide {
		return s.createSSEHandler(ep)
	}
	return s.createJSONHandler(ep)
}

func (s *Server) createJSONHandler(ep *talk.Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req any

		// Parse body if present
		if ep.RequestType != nil && r.ContentLength > 0 {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				s.writeError(w, talk.NewError(talk.InvalidArgument, "failed to read body"))
				return
			}

			reqVal := reflect.New(ep.RequestType).Interface()
			if err := s.codec.Unmarshal(body, reqVal); err != nil {
				s.writeError(w, talk.NewError(talk.InvalidArgument, "failed to decode request"))
				return
			}
			req = reflect.ValueOf(reqVal).Elem().Interface()
		}

		// Ensure request struct is instantiated for struct types even without body
		if req == nil && ep.RequestType != nil && ep.RequestType.Kind() == reflect.Struct {
			req = reflect.New(ep.RequestType).Elem().Interface()
		}

		// Extract path parameters and query parameters
		req = s.extractParams(r, ep, req)

		handler := ep.WrappedHandler()
		resp, err := handler(ctx, req)
		if err != nil {
			s.writeError(w, talk.ToError(err))
			return
		}

		s.writeJSON(w, http.StatusOK, resp)
	}
}

func (s *Server) createSSEHandler(ep *talk.Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		flusher, ok := w.(http.Flusher)
		if !ok {
			s.writeError(w, talk.NewError(talk.Internal, "streaming not supported"))
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		var req any
		if ep.RequestType != nil && ep.RequestType.Kind() == reflect.Struct {
			req = reflect.New(ep.RequestType).Elem().Interface()
		}
		req = s.extractParams(r, ep, req)

		stream := &sseServerStream{
			ctx:     ctx,
			w:       w,
			flusher: flusher,
			codec:   s.codec,
		}

		if ep.StreamHandler != nil {
			if err := ep.StreamHandler(ctx, req, stream); err != nil {
				fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
				flusher.Flush()
			}
		}
	}
}

// pathParamRegex matches {paramName} patterns in endpoint paths.
var pathParamRegex = regexp.MustCompile(`\{(\w+)\}`)

// extractParams populates the request with path parameters and query parameters.
// It supports both struct types (via `path` and `query` struct tags) and simple
// string types (backward-compatible {id} extraction).
func (s *Server) extractParams(r *http.Request, ep *talk.Endpoint, req any) any {
	// For simple types (string, int, etc.), maintain backward compatibility:
	// extract {id} as a raw value.
	if ep.RequestType == nil || isSimpleType(ep.RequestType) {
		if strings.Contains(ep.Path, "{id}") {
			id := r.PathValue("id")
			if id != "" && req == nil {
				return id
			}
		}
		return req
	}

	if req == nil {
		return req
	}

	// For struct types, use reflection to populate fields from path/query params
	v := reflect.ValueOf(&req).Elem()
	// req is an interface{} holding a struct value; we need a pointer to modify it
	structVal := reflect.New(ep.RequestType).Elem()
	// Copy existing values from req into structVal
	structVal.Set(reflect.ValueOf(req))

	t := ep.RequestType
	changed := false

	// Extract path parameters using `path` struct tag
	paramNames := pathParamRegex.FindAllStringSubmatch(ep.Path, -1)
	for _, match := range paramNames {
		paramName := match[1]
		paramValue := r.PathValue(paramName)
		if paramValue == "" {
			continue
		}
		if setStructFieldByTag(structVal, t, "path", paramName, paramValue) {
			changed = true
		}
	}

	// Extract query parameters using `query` struct tag, falling back to `json` tag.
	// JSON fallback only applies to methods without a body (GET, DELETE, HEAD, OPTIONS)
	// to prevent query params from overriding body-parsed values on POST/PUT/PATCH.
	bodylessMethod := r.Method == "GET" || r.Method == "DELETE" || r.Method == "HEAD" || r.Method == "OPTIONS"
	queryValues := r.URL.Query()
	if len(queryValues) > 0 {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			tag := field.Tag.Get("query")
			if tag == "" || tag == "-" {
				if !bodylessMethod {
					continue
				}
				// Fallback to json tag name only for bodyless methods
				tag = field.Tag.Get("json")
				if idx := strings.Index(tag, ","); idx != -1 {
					tag = tag[:idx]
				}
			}
			if tag == "" || tag == "-" {
				continue
			}
			qv := queryValues.Get(tag)
			if qv == "" {
				continue
			}
			if setFieldValue(structVal.Field(i), qv) {
				changed = true
			}
		}
	}

	if changed {
		v.Set(structVal)
		return v.Interface()
	}
	return req
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", s.codec.ContentType())
	w.WriteHeader(status)
	if data != nil {
		body, _ := s.codec.Marshal(data)
		w.Write(body)
	}
}

func (s *Server) writeError(w http.ResponseWriter, err *talk.Error) {
	w.Header().Set("Content-Type", s.codec.ContentType())
	w.WriteHeader(err.HTTPStatus())
	body, _ := s.codec.Marshal(err)
	w.Write(body)
}

type sseServerStream struct {
	ctx     context.Context
	w       http.ResponseWriter
	flusher http.Flusher
	codec   codec.Codec
	closed  bool
}

func (s *sseServerStream) Context() context.Context {
	return s.ctx
}

func (s *sseServerStream) Send(msg any) error {
	if s.closed {
		return io.ErrClosedPipe
	}

	data, err := s.codec.Marshal(msg)
	if err != nil {
		return err
	}

	fmt.Fprintf(s.w, "data: %s\n\n", data)
	s.flusher.Flush()
	return nil
}

func (s *sseServerStream) Recv(msg any) error {
	return talk.NewError(talk.Unimplemented, "SSE is server-push only")
}

func (s *sseServerStream) Close() error {
	s.closed = true
	return nil
}

func convertPathParams(path string) string {
	return path
}

// setStructFieldByTag finds a struct field by the given tag name and key,
// and sets its value from the string parameter.
func setStructFieldByTag(v reflect.Value, t reflect.Type, tagName, tagValue, paramValue string) bool {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(tagName)
		if tag == tagValue {
			return setFieldValue(v.Field(i), paramValue)
		}
		// Fallback: match by json tag name
		if tag == "" {
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" {
				name := strings.Split(jsonTag, ",")[0]
				if name == tagValue {
					return setFieldValue(v.Field(i), paramValue)
				}
			}
		}
	}
	return false
}

// setFieldValue sets a reflect.Value from a string, supporting common types.
func setFieldValue(field reflect.Value, value string) bool {
	if !field.CanSet() {
		return false
	}
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if n, err := strconv.ParseInt(value, 10, 64); err == nil {
			field.SetInt(n)
			return true
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if n, err := strconv.ParseUint(value, 10, 64); err == nil {
			field.SetUint(n)
			return true
		}
	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			field.SetFloat(f)
			return true
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(value); err == nil {
			field.SetBool(b)
			return true
		}
	}
	return false
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

func init() {
	thttp.ServerFactory.Register("std", func(cfg x.TypedLazyConfig, opts ...thttp.Option) (thttp.ServerTransport, error) {
		return NewServer(cfg, opts...)
	}, "default")
}
