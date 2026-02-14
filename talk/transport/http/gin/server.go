// Package gin provides a Gin-based HTTP transport implementation for talk.
package gin

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"go.zoe.im/x"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/codec"
	thttp "go.zoe.im/x/talk/transport/http"
)

// Server implements talk.Transport using Gin.
type Server struct {
	config    thttp.ServerConfig
	codec     codec.Codec
	server    *http.Server
	engine    *gin.Engine
	endpoints []*talk.Endpoint
}

// NewServer creates a new Gin HTTP server transport.
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

	// Set Gin to release mode by default
	gin.SetMode(gin.ReleaseMode)
	s.engine = gin.New()
	s.engine.Use(gin.Recovery())

	return s, nil
}

// SetCodec sets the codec for the server.
func (s *Server) SetCodec(c codec.Codec) {
	s.codec = c
}

// Engine returns the underlying Gin engine for adding custom middleware.
func (s *Server) Engine() *gin.Engine {
	return s.engine
}

// String returns the transport name.
func (s *Server) String() string {
	return "http/gin"
}

// Serve starts the server and blocks until context is cancelled.
func (s *Server) Serve(ctx context.Context, endpoints []*talk.Endpoint) error {
	s.endpoints = endpoints

	for _, ep := range endpoints {
		s.registerEndpoint(ep)
	}

	s.server = &http.Server{
		Addr:         s.config.Addr,
		Handler:      s.engine,
		ReadTimeout:  time.Duration(s.config.ReadTimeout),
		WriteTimeout: time.Duration(s.config.WriteTimeout),
		IdleTimeout:  time.Duration(s.config.IdleTimeout),
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

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// Invoke is not supported on server.
func (s *Server) Invoke(ctx context.Context, endpoint string, req any, resp any) error {
	return talk.NewError(talk.Unimplemented, "server does not support Invoke")
}

// InvokeStream is not supported on server.
func (s *Server) InvokeStream(ctx context.Context, endpoint string, req any) (talk.Stream, error) {
	return nil, talk.NewError(talk.Unimplemented, "server does not support InvokeStream")
}

// Close closes the server.
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

		// Extract path parameters
		req = s.extractPathParams(c, ep, req)

		resp, err := ep.Handler(ctx, req)
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
		req = s.extractPathParams(c, ep, req)

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

func (s *Server) extractPathParams(c *gin.Context, ep *talk.Endpoint, req any) any {
	// For simple ID parameter extraction
	if strings.Contains(ep.Path, "{id}") {
		id := c.Param("id")
		if id != "" && req == nil {
			return id
		}
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

// convertPathParams converts {param} to :param for Gin routing
func convertPathParams(path string) string {
	// Replace {param} with :param
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
}
