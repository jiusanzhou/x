package std

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

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
	pathPrefix string
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

// WithClientPathPrefix sets a path prefix for all derived endpoint paths.
func WithClientPathPrefix(prefix string) thttp.Option {
	return func(v any) {
		if c, ok := v.(*Client); ok {
			c.pathPrefix = strings.TrimSuffix(prefix, "/")
		}
	}
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
	var httpMethod string
	var path string

	if strings.HasPrefix(endpoint, "/") {
		// Direct path — passthrough
		httpMethod = http.MethodPost
		path = endpoint
	} else {
		// Method name — derive RESTful path and HTTP method
		httpMethod, path = deriveClientPath(endpoint)
		path = c.pathPrefix + path
	}

	url := c.baseURL + path

	var body io.Reader
	method := httpMethod

	// For methods with bodies (POST, PUT, PATCH), encode the request
	if req != nil && (method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch) {
		data, err := c.codec.Marshal(req)
		if err != nil {
			return talk.NewError(talk.InvalidArgument, "failed to encode request")
		}
		body = bytes.NewReader(data)
	}

	// For GET/DELETE with a simple ID request, append to path
	if req != nil && (method == http.MethodGet || method == http.MethodDelete) {
		if id, ok := req.(string); ok && strings.Contains(url, "{id}") {
			url = strings.Replace(url, "{id}", id, 1)
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, body)
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

// deriveClientPath converts a Go method name to an HTTP method + RESTful path,
// matching the server-side deriveMethodAndPath logic.
func deriveClientPath(name string) (httpMethod, path string) {
	resource := extractResource(name)

	switch {
	case strings.HasPrefix(name, "Get"):
		return "GET", "/" + resource + "/{id}"
	case strings.HasPrefix(name, "List"):
		return "GET", "/" + resource
	case strings.HasPrefix(name, "Create"):
		return "POST", "/" + resource
	case strings.HasPrefix(name, "Update"):
		return "PUT", "/" + resource + "/{id}"
	case strings.HasPrefix(name, "Delete"):
		return "DELETE", "/" + resource + "/{id}"
	case strings.HasPrefix(name, "Watch"):
		return "GET", "/" + resource + "/watch"
	default:
		return "POST", "/" + toKebabCase(name)
	}
}

func extractResource(name string) string {
	prefixes := []string{"Get", "List", "Create", "Update", "Delete", "Watch"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return strings.ToLower(name[len(prefix):])
		}
	}
	return strings.ToLower(name)
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

func init() {
	thttp.ClientFactory.Register("std", func(cfg x.TypedLazyConfig, opts ...thttp.Option) (thttp.ClientTransport, error) {
		return NewClient(cfg, opts...)
	}, "default")

	talk.RegisterTransport("http", &talk.TransportCreators{
		Server: func(cfg x.TypedLazyConfig) (talk.Transport, error) {
			return NewServer(cfg)
		},
		Client: func(cfg x.TypedLazyConfig) (talk.Transport, error) {
			return NewClient(cfg)
		},
	}, "http/std", "http/default")
}
