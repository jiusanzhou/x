// Package gin provides a Gin-based HTTP transport implementation for talk.
package gin

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

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
	engine         *gin.Engine
	endpoints      []*talk.Endpoint
	swaggerHandler *swagger.Handler
	externalEngine bool
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

	if s.engine == nil {
		gin.SetMode(gin.ReleaseMode)
		s.engine = gin.New()
		s.engine.Use(gin.Recovery())
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

func WithEngine(engine *gin.Engine) thttp.Option {
	return func(v any) {
		if s, ok := v.(*Server); ok {
			s.engine = engine
			s.externalEngine = true
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

func (s *Server) Engine() *gin.Engine {
	return s.engine
}

func (s *Server) Handler() http.Handler {
	return s.engine
}

func (s *Server) RegisterEndpoints(endpoints []*talk.Endpoint) {
	s.endpoints = endpoints
	for _, ep := range endpoints {
		s.registerEndpoint(ep)
	}
	if s.swaggerHandler != nil {
		s.swaggerHandler.SetEndpoints(endpoints)
		s.engine.Any(s.swaggerHandler.BasePath()+"/*filepath", gin.WrapH(s.swaggerHandler))
	}
}

func (s *Server) String() string {
	return "http/gin"
}

func (s *Server) Serve(ctx context.Context, endpoints []*talk.Endpoint) error {
	s.RegisterEndpoints(endpoints)

	if s.externalEngine {
		<-ctx.Done()
		return nil
	}

	if s.server == nil {
		s.server = &http.Server{
			Addr:         s.config.Addr,
			Handler:      s.engine,
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
	path := convertPathParams(ep.Path)

	handler := s.createHandler(ep)

	switch ep.Method {
	case "GET":
		s.engine.GET(path, handler)
	case "POST":
		s.engine.POST(path, handler)
	case "PUT":
		s.engine.PUT(path, handler)
	case "DELETE":
		s.engine.DELETE(path, handler)
	case "PATCH":
		s.engine.PATCH(path, handler)
	default:
		s.engine.POST(path, handler)
	}
}

func (s *Server) createHandler(ep *talk.Endpoint) gin.HandlerFunc {
	if ep.IsStreaming() && ep.StreamMode == talk.StreamServerSide {
		return s.createSSEHandler(ep)
	}
	return s.createJSONHandler(ep)
}

func (s *Server) createJSONHandler(ep *talk.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var req any
		if ep.RequestType != nil && c.Request.ContentLength > 0 {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				s.writeError(c, talk.NewError(talk.InvalidArgument, "failed to read body"))
				return
			}

			reqVal := reflect.New(ep.RequestType).Interface()
			if err := s.codec.Unmarshal(body, reqVal); err != nil {
				s.writeError(c, talk.NewError(talk.InvalidArgument, "failed to decode request"))
				return
			}
			req = reflect.ValueOf(reqVal).Elem().Interface()
		}

		// Ensure request struct is instantiated for struct types even without body
		if req == nil && ep.RequestType != nil && ep.RequestType.Kind() == reflect.Struct {
			req = reflect.New(ep.RequestType).Elem().Interface()
		}

		// Extract path parameters and query parameters
		req = s.extractParams(c, ep, req)

		handler := ep.WrappedHandler()
		resp, err := handler(ctx, req)
		if err != nil {
			s.writeError(c, talk.ToError(err))
			return
		}

		s.writeJSON(c, http.StatusOK, resp)
	}
}

func (s *Server) createSSEHandler(ep *talk.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		var req any
		if ep.RequestType != nil && ep.RequestType.Kind() == reflect.Struct {
			req = reflect.New(ep.RequestType).Elem().Interface()
		}
		req = s.extractParams(c, ep, req)

		stream := &sseServerStream{
			ctx:   ctx,
			c:     c,
			codec: s.codec,
		}

		if ep.StreamHandler != nil {
			if err := ep.StreamHandler(ctx, req, stream); err != nil {
				c.SSEvent("error", err.Error())
			}
		}
	}
}

// pathParamRegex matches {paramName} patterns in endpoint paths.
var pathParamRegex = regexp.MustCompile(`\{(\w+)\}`)

// extractParams populates the request with path parameters and query parameters.
// It supports both struct types (via `path` and `query` struct tags) and simple
// string types (backward-compatible {id} extraction).
func (s *Server) extractParams(c *gin.Context, ep *talk.Endpoint, req any) any {
	// extract {id} as a raw value.
	if ep.RequestType == nil || isSimpleType(ep.RequestType) {
		if strings.Contains(ep.Path, "{id}") {
			id := c.Param("id")
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
	structVal := reflect.New(ep.RequestType).Elem()
	structVal.Set(reflect.ValueOf(req))

	t := ep.RequestType
	changed := false

	// Extract path parameters using `path` struct tag
	paramNames := pathParamRegex.FindAllStringSubmatch(ep.Path, -1)
	for _, match := range paramNames {
		paramName := match[1]
		paramValue := c.Param(paramName)
		if paramValue == "" {
			continue
		}
		if setStructFieldByTag(structVal, t, "path", paramName, paramValue) {
			changed = true
		}
	}

	// Extract query parameters using `query` struct tag, falling back to `json` tag.
	// This allows GET endpoints to bind ?page=1&size=10 to struct fields
	// tagged with either `query:"page"` or `json:"page"`.
	queryValues := c.Request.URL.Query()
	if len(queryValues) > 0 {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			tag := field.Tag.Get("query")
			if tag == "" || tag == "-" {
				// Fallback to json tag name for query binding
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

func (s *Server) writeJSON(c *gin.Context, status int, data any) {
	c.Header("Content-Type", s.codec.ContentType())
	if data != nil {
		body, _ := s.codec.Marshal(data)
		c.Data(status, s.codec.ContentType(), body)
	} else {
		c.Status(status)
	}
}

func (s *Server) writeError(c *gin.Context, err *talk.Error) {
	body, _ := s.codec.Marshal(err)
	c.Data(err.HTTPStatus(), s.codec.ContentType(), body)
}

func convertPathParams(path string) string {
	result := path
	for {
		start := strings.Index(result, "{")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		param := result[start+1 : start+end]
		result = result[:start] + ":" + param + result[start+end+1:]
	}
	return result
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

type sseServerStream struct {
	ctx    context.Context
	c      *gin.Context
	codec  codec.Codec
	closed bool
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

	s.c.SSEvent("message", string(data))
	s.c.Writer.Flush()
	return nil
}

func (s *sseServerStream) Recv(msg any) error {
	return talk.NewError(talk.Unimplemented, "SSE is server-push only")
}

func (s *sseServerStream) Close() error {
	s.closed = true
	return nil
}

func init() {
	thttp.ServerFactory.Register("gin", func(cfg x.TypedLazyConfig, opts ...thttp.Option) (thttp.ServerTransport, error) {
		return NewServer(cfg, opts...)
	})

	talk.RegisterTransport("http/gin", &talk.TransportCreators{
		Server: func(cfg x.TypedLazyConfig) (talk.Transport, error) {
			return NewServer(cfg)
		},
	})
}
