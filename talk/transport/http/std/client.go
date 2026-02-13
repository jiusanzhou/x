package std

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"go.zoe.im/x"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/codec"
	thttp "go.zoe.im/x/talk/transport/http"
)

// Client implements talk.Transport for HTTP client operations.
type Client struct {
	config     thttp.ClientConfig
	codec      codec.Codec
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new HTTP client transport.
func NewClient(cfg x.TypedLazyConfig, opts ...thttp.Option) (*Client, error) {
	c := &Client{}

	if err := cfg.Unmarshal(&c.config); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.codec == nil {
		c.codec = codec.MustGet("json")
	}

	c.baseURL = strings.TrimSuffix(c.config.Addr, "/")
	c.httpClient = &http.Client{
		Timeout: time.Duration(c.config.Timeout),
	}

	return c, nil
}

func (c *Client) SetCodec(cd codec.Codec) {
	c.codec = cd
}

func (c *Client) String() string {
	return "http/std/client"
}

func (c *Client) Serve(ctx context.Context, endpoints []*talk.Endpoint) error {
	return talk.NewError(talk.Unimplemented, "client does not support Serve")
}

func (c *Client) Shutdown(ctx context.Context) error {
	return nil
}

func (c *Client) Invoke(ctx context.Context, endpoint string, req any, resp any) error {
	url := c.baseURL + "/" + strings.TrimPrefix(endpoint, "/")

	var body io.Reader
	if req != nil {
		data, err := c.codec.Marshal(req)
		if err != nil {
			return talk.NewError(talk.InvalidArgument, "failed to encode request")
		}
		body = bytes.NewReader(data)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return talk.NewError(talk.Internal, err.Error())
	}

	httpReq.Header.Set("Content-Type", c.codec.ContentType())
	httpReq.Header.Set("Accept", c.codec.ContentType())

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return talk.NewError(talk.Unavailable, err.Error())
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return talk.NewError(talk.Internal, "failed to read response")
	}

	if httpResp.StatusCode >= 400 {
		var talkErr talk.Error
		if err := c.codec.Unmarshal(respBody, &talkErr); err == nil && talkErr.Code != talk.OK {
			return &talkErr
		}
		return talk.NewError(talk.FromHTTPStatus(httpResp.StatusCode), string(respBody))
	}

	if resp != nil && len(respBody) > 0 {
		if err := c.codec.Unmarshal(respBody, resp); err != nil {
			return talk.NewError(talk.Internal, "failed to decode response")
		}
	}

	return nil
}

func (c *Client) InvokeStream(ctx context.Context, endpoint string, req any) (talk.Stream, error) {
	url := c.baseURL + "/" + strings.TrimPrefix(endpoint, "/")

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, talk.NewError(talk.Internal, err.Error())
	}

	httpReq.Header.Set("Accept", "text/event-stream")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, talk.NewError(talk.Unavailable, err.Error())
	}

	if httpResp.StatusCode >= 400 {
		defer httpResp.Body.Close()
		body, _ := io.ReadAll(httpResp.Body)
		return nil, talk.NewError(talk.FromHTTPStatus(httpResp.StatusCode), string(body))
	}

	return &sseClientStream{
		ctx:    ctx,
		resp:   httpResp,
		reader: bufio.NewReader(httpResp.Body),
		codec:  c.codec,
	}, nil
}

func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

type sseClientStream struct {
	ctx    context.Context
	resp   *http.Response
	reader *bufio.Reader
	codec  codec.Codec
	closed bool
}

func (s *sseClientStream) Context() context.Context {
	return s.ctx
}

func (s *sseClientStream) Send(msg any) error {
	return talk.NewError(talk.Unimplemented, "SSE client stream is receive-only")
}

func (s *sseClientStream) Recv(msg any) error {
	if s.closed {
		return io.EOF
	}

	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		default:
		}

		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				s.closed = true
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if err := s.codec.Unmarshal([]byte(data), msg); err != nil {
				return err
			}
			return nil
		}

		if strings.HasPrefix(line, "event: error") {
			nextLine, _ := s.reader.ReadString('\n')
			if strings.HasPrefix(nextLine, "data: ") {
				errMsg := strings.TrimPrefix(strings.TrimSpace(nextLine), "data: ")
				return talk.NewError(talk.Internal, errMsg)
			}
		}
	}
}

func (s *sseClientStream) Close() error {
	s.closed = true
	return s.resp.Body.Close()
}

func init() {
	thttp.ClientFactory.Register("std", func(cfg x.TypedLazyConfig, opts ...thttp.Option) (thttp.ClientTransport, error) {
		return NewClient(cfg, opts...)
	})
}
