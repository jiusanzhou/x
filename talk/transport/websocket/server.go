package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/websocket"

	"go.zoe.im/x"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/codec"
)

// Server implements talk.Transport using WebSocket.
type Server struct {
	config    ServerConfig
	codec     codec.Codec
	server    *http.Server
	endpoints map[string]*talk.Endpoint
	conns     sync.Map
}

// NewServer creates a new WebSocket server transport.
func NewServer(cfg x.TypedLazyConfig, opts ...Option) (*Server, error) {
	s := &Server{
		endpoints: make(map[string]*talk.Endpoint),
	}

	if err := cfg.Unmarshal(&s.config); err != nil {
		return nil, err
	}

	if s.config.Path == "" {
		s.config.Path = "/ws"
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
	return "websocket"
}

func (s *Server) Serve(ctx context.Context, endpoints []*talk.Endpoint) error {
	for _, ep := range endpoints {
		s.endpoints[ep.Name] = ep
	}

	mux := http.NewServeMux()
	mux.Handle(s.config.Path, websocket.Handler(s.handleConnection))

	s.server = &http.Server{
		Addr:    s.config.Addr,
		Handler: mux,
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
	s.conns.Range(func(key, value any) bool {
		if conn, ok := value.(*websocket.Conn); ok {
			conn.Close()
		}
		return true
	})

	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
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

func (s *Server) handleConnection(conn *websocket.Conn) {
	connID := fmt.Sprintf("%p", conn)
	s.conns.Store(connID, conn)
	defer func() {
		s.conns.Delete(connID)
		conn.Close()
	}()

	stream := &wsStream{
		conn:  conn,
		codec: s.codec,
	}

	for {
		var msg wsMessage
		if err := websocket.JSON.Receive(conn, &msg); err != nil {
			if err == io.EOF {
				return
			}
			s.sendError(conn, "", talk.NewError(talk.InvalidArgument, err.Error()))
			continue
		}

		ep, ok := s.endpoints[msg.Method]
		if !ok {
			s.sendError(conn, msg.ID, talk.NewError(talk.NotFound, "method not found: "+msg.Method))
			continue
		}

		go s.handleRequest(conn, stream, ep, &msg)
	}
}

func (s *Server) handleRequest(conn *websocket.Conn, stream *wsStream, ep *talk.Endpoint, msg *wsMessage) {
	ctx := context.Background()

	if ep.IsStreaming() && ep.StreamHandler != nil {
		if err := ep.StreamHandler(ctx, msg.Params, stream); err != nil {
			s.sendError(conn, msg.ID, talk.ToError(err))
		}
		return
	}

	if ep.Handler == nil {
		s.sendError(conn, msg.ID, talk.NewError(talk.Unimplemented, "no handler configured"))
		return
	}

	resp, err := ep.Handler(ctx, msg.Params)
	if err != nil {
		s.sendError(conn, msg.ID, talk.ToError(err))
		return
	}

	s.sendResponse(conn, msg.ID, resp)
}

func (s *Server) sendResponse(conn *websocket.Conn, id string, result any) {
	response := wsResponse{
		ID:     id,
		Result: result,
	}
	websocket.JSON.Send(conn, response)
}

func (s *Server) sendError(conn *websocket.Conn, id string, err *talk.Error) {
	response := wsResponse{
		ID: id,
		Error: &wsError{
			Code:    int(err.Code),
			Message: err.Message,
		},
	}
	websocket.JSON.Send(conn, response)
}

type wsMessage struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

type wsResponse struct {
	ID     string   `json:"id"`
	Result any      `json:"result,omitempty"`
	Error  *wsError `json:"error,omitempty"`
}

type wsError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type wsStream struct {
	conn  *websocket.Conn
	codec codec.Codec
	mu    sync.Mutex
}

func (s *wsStream) Context() context.Context {
	return context.Background()
}

func (s *wsStream) Send(msg any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.codec.Marshal(msg)
	if err != nil {
		return err
	}

	response := wsResponse{
		Result: json.RawMessage(data),
	}
	return websocket.JSON.Send(s.conn, response)
}

func (s *wsStream) Recv(msg any) error {
	var wsMsg wsMessage
	if err := websocket.JSON.Receive(s.conn, &wsMsg); err != nil {
		return err
	}
	return s.codec.Unmarshal(wsMsg.Params, msg)
}

func (s *wsStream) Close() error {
	return s.conn.Close()
}

func init() {
	ServerFactory.Register("default", func(cfg x.TypedLazyConfig, opts ...Option) (ServerTransport, error) {
		return NewServer(cfg, opts...)
	})

	talk.RegisterServerTransport("websocket", func(cfg x.TypedLazyConfig) (talk.Transport, error) {
		return NewServer(cfg)
	}, "ws")
}

var _ = time.Second
