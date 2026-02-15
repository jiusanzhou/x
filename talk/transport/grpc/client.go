package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"go.zoe.im/x"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/codec"
)

// Client implements talk.Transport for gRPC client operations.
type Client struct {
	config ClientConfig
	codec  codec.Codec
	conn   *grpc.ClientConn
}

// NewClient creates a new gRPC client transport.
func NewClient(cfg x.TypedLazyConfig, opts ...Option) (*Client, error) {
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

	var dialOpts []grpc.DialOption

	if c.config.Insecure {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	} else if c.config.TLSCertFile != "" {
		creds, err := credentials.NewClientTLSFromFile(c.config.TLSCertFile, "")
		if err != nil {
			return nil, err
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	if c.config.WaitForReady {
		dialOpts = append(dialOpts, grpc.WithDefaultCallOptions(grpc.WaitForReady(true)))
	}

	conn, err := grpc.Dial(c.config.Addr, dialOpts...)
	if err != nil {
		return nil, err
	}
	c.conn = conn

	return c, nil
}

func (c *Client) SetCodec(cd codec.Codec) {
	c.codec = cd
}

func (c *Client) String() string {
	return "grpc/client"
}

func (c *Client) Serve(ctx context.Context, endpoints []*talk.Endpoint) error {
	return talk.NewError(talk.Unimplemented, "client does not support Serve")
}

func (c *Client) Shutdown(ctx context.Context) error {
	return nil
}

func (c *Client) Invoke(ctx context.Context, endpoint string, req any, resp any) error {
	method := "/talk.Service/" + endpoint

	var callOpts []grpc.CallOption
	if c.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(c.config.Timeout))
		defer cancel()
	}

	reqData, err := c.codec.Marshal(req)
	if err != nil {
		return talk.NewError(talk.InvalidArgument, "failed to encode request")
	}

	var respData []byte
	err = c.conn.Invoke(ctx, method, reqData, &respData, callOpts...)
	if err != nil {
		return c.fromGRPCError(err)
	}

	if resp != nil && len(respData) > 0 {
		if err := c.codec.Unmarshal(respData, resp); err != nil {
			return talk.NewError(talk.Internal, "failed to decode response")
		}
	}

	return nil
}

func (c *Client) InvokeStream(ctx context.Context, endpoint string, req any) (talk.Stream, error) {
	method := "/talk.Service/" + endpoint

	streamDesc := &grpc.StreamDesc{
		StreamName:    endpoint,
		ServerStreams: true,
		ClientStreams: true,
	}

	clientStream, err := c.conn.NewStream(ctx, streamDesc, method)
	if err != nil {
		return nil, c.fromGRPCError(err)
	}

	return &grpcClientStream{
		ClientStream: clientStream,
		codec:        c.codec,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) fromGRPCError(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return talk.NewError(talk.Unknown, err.Error())
	}

	return talk.NewError(talk.ErrorCode(st.Code()), st.Message())
}

type grpcClientStream struct {
	grpc.ClientStream
	codec codec.Codec
}

func (s *grpcClientStream) Context() context.Context {
	return s.ClientStream.Context()
}

func (s *grpcClientStream) Send(msg any) error {
	data, err := s.codec.Marshal(msg)
	if err != nil {
		return err
	}
	return s.ClientStream.SendMsg(data)
}

func (s *grpcClientStream) Recv(msg any) error {
	var data []byte
	if err := s.ClientStream.RecvMsg(&data); err != nil {
		return err
	}
	return s.codec.Unmarshal(data, msg)
}

func (s *grpcClientStream) Close() error {
	return s.ClientStream.CloseSend()
}

func init() {
	ClientFactory.Register("default", func(cfg x.TypedLazyConfig, opts ...Option) (ClientTransport, error) {
		return NewClient(cfg, opts...)
	})

	talk.RegisterTransport("grpc", &talk.TransportCreators{
		Server: func(cfg x.TypedLazyConfig) (talk.Transport, error) {
			return NewServer(cfg)
		},
		Client: func(cfg x.TypedLazyConfig) (talk.Transport, error) {
			return NewClient(cfg)
		},
	})
}

// ErrorCodeFromGRPC converts a gRPC code to a talk ErrorCode.
func ErrorCodeFromGRPC(code codes.Code) talk.ErrorCode {
	return talk.ErrorCode(code)
}
