package unix

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"go.zoe.im/x"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/codec"
)

type Server struct {
	config    ServerConfig
	codec     codec.Codec
	server    *http.Server
	listener  net.Listener
	mux       *http.ServeMux
	endpoints []*talk.Endpoint
}

func NewServer(cfg x.TypedLazyConfig, opts ...Option) (*Server, error) {
	s := &Server{
		mux: http.NewServeMux(),
	}

	if err := cfg.Unmarshal(&s.config); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.codec == nil {
		s.codec = codec.MustGet("json")
	}

	return s, nil
}

func (s *Server) SetCodec(c codec.Codec) {
	s.codec = c
}

func (s *Server) String() string {
	return "unix"
}

func (s *Server) Serve(ctx context.Context, endpoints []*talk.Endpoint) error {
	s.endpoints = endpoints

	for _, ep := range endpoints {
		s.registerEndpoint(ep)
	}

	if err := os.Remove(s.config.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing socket: %w", err)
	}

	listener, err := net.Listen("unix", s.config.Path)
	if err != nil {
		return fmt.Errorf("failed to listen on unix socket: %w", err)
	}
	s.listener = listener

	s.server = &http.Server{
		Handler:      s.mux,
		ReadTimeout:  time.Duration(s.config.ReadTimeout),
		WriteTimeout: time.Duration(s.config.WriteTimeout),
		IdleTimeout:  time.Duration(s.config.IdleTimeout),
	}

	errCh := make(chan error, 1)
	go func() {
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
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
	err := s.server.Shutdown(ctx)
	os.Remove(s.config.Path)
	return err
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

		req = s.extractPathParams(r, ep, req)

		resp, err := ep.Handler(ctx, req)
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

		var req any
		req = s.extractPathParams(r, ep, req)

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

func (s *Server) extractPathParams(r *http.Request, ep *talk.Endpoint, req any) any {
	if strings.Contains(ep.Path, "{id}") {
		id := r.PathValue("id")
		if id != "" && req == nil {
			return id
		}
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

func init() {
	ServerFactory.Register("default", func(cfg x.TypedLazyConfig, opts ...Option) (ServerTransport, error) {
		return NewServer(cfg, opts...)
	})
}
