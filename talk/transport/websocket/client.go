package websocket

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"

	"go.zoe.im/x"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/codec"
)

// Client implements talk.Transport for WebSocket client operations.
type Client struct {
	config  ClientConfig
	codec   codec.Codec
	conn    *websocket.Conn
	mu      sync.Mutex
	reqID   uint64
	pending sync.Map
	closed  bool
}

// NewClient creates a new WebSocket client transport.
func NewClient(cfg x.TypedLazyConfig, opts ...Option) (*Client, error) {
	c := &Client{}

	if err := cfg.Unmarshal(&c.config); err != nil {
		return nil, err
	}

	if c.config.Path == "" {
		c.config.Path = "/ws"
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.codec == nil {
		c.codec = codec.MustGet("json")
	}

	wsConfig, err := websocket.NewConfig("ws://"+c.config.Addr+c.config.Path, "http://localhost")
	if err != nil {
		return nil, err
	}

	conn, err := websocket.DialConfig(wsConfig)
	if err != nil {
		return nil, talk.NewError(talk.Unavailable, err.Error())
	}
	c.conn = conn

	go c.readLoop()

	return c, nil
}

func (c *Client) SetCodec(cd codec.Codec) {
	c.codec = cd
}

func (c *Client) String() string {
	return "websocket/client"
}

func (c *Client) Serve(ctx context.Context, endpoints []*talk.Endpoint) error {
	return talk.NewError(talk.Unimplemented, "client does not support Serve")
}

func (c *Client) Shutdown(ctx context.Context) error {
	return nil
}

func (c *Client) Invoke(ctx context.Context, endpoint string, req any, resp any) error {
	id := c.nextID()

	reqData, err := json.Marshal(req)
	if err != nil {
		return talk.NewError(talk.InvalidArgument, "failed to encode request")
	}

	msg := wsMessage{
		ID:     id,
		Method: endpoint,
		Params: reqData,
	}

	respCh := make(chan *wsResponse, 1)
	c.pending.Store(id, respCh)
	defer c.pending.Delete(id)

	c.mu.Lock()
	err = websocket.JSON.Send(c.conn, msg)
	c.mu.Unlock()

	if err != nil {
		return talk.NewError(talk.Unavailable, err.Error())
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case response := <-respCh:
		if response.Error != nil {
			return talk.NewError(talk.ErrorCode(response.Error.Code), response.Error.Message)
		}
		if resp != nil && response.Result != nil {
			data, err := json.Marshal(response.Result)
			if err != nil {
				return talk.NewError(talk.Internal, "failed to encode result")
			}
			if err := c.codec.Unmarshal(data, resp); err != nil {
				return talk.NewError(talk.Internal, "failed to decode response")
			}
		}
		return nil
	}
}

func (c *Client) InvokeStream(ctx context.Context, endpoint string, req any) (talk.Stream, error) {
	return &wsClientStream{
		client:   c,
		endpoint: endpoint,
		ctx:      ctx,
	}, nil
}

func (c *Client) Close() error {
	c.closed = true
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) nextID() string {
	id := atomic.AddUint64(&c.reqID, 1)
	return string(rune('a' + (id % 26)))
}

func (c *Client) readLoop() {
	for !c.closed {
		var response wsResponse
		if err := websocket.JSON.Receive(c.conn, &response); err != nil {
			if err == io.EOF || c.closed {
				return
			}
			continue
		}

		if ch, ok := c.pending.Load(response.ID); ok {
			if respCh, ok := ch.(chan *wsResponse); ok {
				select {
				case respCh <- &response:
				default:
				}
			}
		}
	}
}

type wsClientStream struct {
	client   *Client
	endpoint string
	ctx      context.Context
}

func (s *wsClientStream) Context() context.Context {
	return s.ctx
}

func (s *wsClientStream) Send(msg any) error {
	reqData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	wsMsg := wsMessage{
		ID:     s.client.nextID(),
		Method: s.endpoint,
		Params: reqData,
	}

	s.client.mu.Lock()
	defer s.client.mu.Unlock()
	return websocket.JSON.Send(s.client.conn, wsMsg)
}

func (s *wsClientStream) Recv(msg any) error {
	var response wsResponse
	if err := websocket.JSON.Receive(s.client.conn, &response); err != nil {
		return err
	}

	if response.Error != nil {
		return talk.NewError(talk.ErrorCode(response.Error.Code), response.Error.Message)
	}

	if response.Result != nil {
		data, err := json.Marshal(response.Result)
		if err != nil {
			return err
		}
		return s.client.codec.Unmarshal(data, msg)
	}

	return nil
}

func (s *wsClientStream) Close() error {
	return nil
}

func init() {
	ClientFactory.Register("default", func(cfg x.TypedLazyConfig, opts ...Option) (ClientTransport, error) {
		return NewClient(cfg, opts...)
	})
}

var _ = time.Second
