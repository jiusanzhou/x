package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"go.zoe.im/x"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/codec"
)

// Server implements talk.Transport using gRPC.
type Server struct {
	config    ServerConfig
	codec     codec.Codec
	server    *grpc.Server
	endpoints map[string]*talk.Endpoint
}

// NewServer creates a new gRPC server transport.
func NewServer(cfg x.TypedLazyConfig, opts ...Option) (*Server, error) {
	s := &Server{
		endpoints: make(map[string]*talk.Endpoint),
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
	return "grpc"
}

func (s *Server) Serve(ctx context.Context, endpoints []*talk.Endpoint) error {
	for _, ep := range endpoints {
		s.endpoints[ep.Name] = ep
	}

	var serverOpts []grpc.ServerOption

	if s.config.MaxRecvMsgSize > 0 {
		serverOpts = append(serverOpts, grpc.MaxRecvMsgSize(s.config.MaxRecvMsgSize))
	}
	if s.config.MaxSendMsgSize > 0 {
		serverOpts = append(serverOpts, grpc.MaxSendMsgSize(s.config.MaxSendMsgSize))
	}

	if s.config.TLSCertFile != "" && s.config.TLSKeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(s.config.TLSCertFile, s.config.TLSKeyFile)
		if err != nil {
			return fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		serverOpts = append(serverOpts, grpc.Creds(creds))
	}

	s.server = grpc.NewServer(serverOpts...)

	s.server.RegisterService(&grpc.ServiceDesc{
		ServiceName: "talk.Service",
		HandlerType: (*interface{})(nil),
		Methods:     s.buildUnaryMethods(),
		Streams:     s.buildStreamMethods(),
		Metadata:    "talk.proto",
	}, s)

	listener, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		if err := s.server.Serve(listener); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		s.server.GracefulStop()
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		s.server.GracefulStop()
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

func (s *Server) buildUnaryMethods() []grpc.MethodDesc {
	var methods []grpc.MethodDesc

	for name, ep := range s.endpoints {
		if ep.IsStreaming() {
			continue
		}

		epCopy := ep
		methods = append(methods, grpc.MethodDesc{
			MethodName: name,
			Handler:    s.createUnaryHandler(epCopy),
		})
	}

	return methods
}

func (s *Server) buildStreamMethods() []grpc.StreamDesc {
	var streams []grpc.StreamDesc

	for name, ep := range s.endpoints {
		if !ep.IsStreaming() {
			continue
		}

		epCopy := ep
		streams = append(streams, grpc.StreamDesc{
			StreamName:    name,
			Handler:       s.createStreamHandler(epCopy),
			ServerStreams: ep.StreamMode == talk.StreamServerSide || ep.StreamMode == talk.StreamBidirect,
			ClientStreams: ep.StreamMode == talk.StreamClientSide || ep.StreamMode == talk.StreamBidirect,
		})
	}

	return streams
}

func (s *Server) createUnaryHandler(ep *talk.Endpoint) func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	return func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
		var req any
		if ep.RequestType != nil {
			req = make(map[string]any)
			if err := dec(&req); err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}

		handler := func(ctx context.Context, req any) (any, error) {
			resp, err := ep.Handler(ctx, req)
			if err != nil {
				return nil, s.toGRPCError(err)
			}
			return resp, nil
		}

		if interceptor == nil {
			return handler(ctx, req)
		}

		info := &grpc.UnaryServerInfo{
			Server:     srv,
			FullMethod: "/talk.Service/" + ep.Name,
		}
		return interceptor(ctx, req, info, handler)
	}
}

func (s *Server) createStreamHandler(ep *talk.Endpoint) func(srv any, stream grpc.ServerStream) error {
	return func(srv any, stream grpc.ServerStream) error {
		talkStream := &grpcServerStream{
			ServerStream: stream,
			codec:        s.codec,
		}

		if ep.StreamHandler != nil {
			return ep.StreamHandler(stream.Context(), nil, talkStream)
		}

		return status.Error(codes.Unimplemented, "no stream handler configured")
	}
}

func (s *Server) toGRPCError(err error) error {
	if err == nil {
		return nil
	}

	if talkErr, ok := err.(*talk.Error); ok {
		return status.Error(codes.Code(talkErr.GRPCCode()), talkErr.Message)
	}

	return status.Error(codes.Unknown, err.Error())
}

type grpcServerStream struct {
	grpc.ServerStream
	codec codec.Codec
}

func (s *grpcServerStream) Context() context.Context {
	return s.ServerStream.Context()
}

func (s *grpcServerStream) Send(msg any) error {
	data, err := s.codec.Marshal(msg)
	if err != nil {
		return err
	}
	return s.ServerStream.SendMsg(data)
}

func (s *grpcServerStream) Recv(msg any) error {
	var data []byte
	if err := s.ServerStream.RecvMsg(&data); err != nil {
		return err
	}
	return s.codec.Unmarshal(data, msg)
}

func (s *grpcServerStream) Close() error {
	return nil
}

func init() {
	ServerFactory.Register("default", func(cfg x.TypedLazyConfig, opts ...Option) (ServerTransport, error) {
		return NewServer(cfg, opts...)
	})

	talk.RegisterServerTransport("grpc", func(cfg x.TypedLazyConfig) (talk.Transport, error) {
		return NewServer(cfg)
	})
}
