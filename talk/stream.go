package talk

import (
	"context"
	"io"
)

// Stream provides bidirectional communication for streaming endpoints.
// Implementations handle protocol-specific details (HTTP/2 streams, WebSocket, gRPC streams).
type Stream interface {
	// Send transmits a message to the remote peer.
	Send(msg any) error

	// Recv receives a message from the remote peer.
	// Returns io.EOF when the stream is closed by the sender.
	Recv(msg any) error

	// Close terminates the stream.
	Close() error

	// Context returns the stream's context.
	Context() context.Context
}

// ServerStream is used by server handlers to send responses to clients.
type ServerStream interface {
	Stream
	// SendHeader sends response headers (for protocols that support it).
	SendHeader(metadata map[string]string) error
}

// ClientStream is used by clients to send requests and receive responses.
type ClientStream interface {
	Stream
	// CloseAndRecv closes the send side and receives the final response.
	CloseAndRecv(resp any) error
	// CloseSend closes the send side of the stream.
	CloseSend() error
}

// streamBase provides common stream functionality.
type streamBase struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *streamBase) Context() context.Context {
	return s.ctx
}

// ChanStream implements Stream using Go channels.
// Useful for in-process communication and testing.
type ChanStream[T any] struct {
	streamBase
	sendCh chan T
	recvCh chan T
	closed bool
}

// NewChanStream creates a bidirectional channel-based stream.
func NewChanStream[T any](ctx context.Context, bufSize int) *ChanStream[T] {
	ctx, cancel := context.WithCancel(ctx)
	return &ChanStream[T]{
		streamBase: streamBase{ctx: ctx, cancel: cancel},
		sendCh:     make(chan T, bufSize),
		recvCh:     make(chan T, bufSize),
	}
}

func (s *ChanStream[T]) Send(msg any) error {
	if s.closed {
		return io.ErrClosedPipe
	}
	v, ok := msg.(T)
	if !ok {
		return NewError(InvalidArgument, "invalid message type")
	}
	select {
	case s.sendCh <- v:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

func (s *ChanStream[T]) Recv(msg any) error {
	if s.closed {
		return io.EOF
	}
	select {
	case v, ok := <-s.recvCh:
		if !ok {
			return io.EOF
		}
		// Type assertion to set the value
		if ptr, ok := msg.(*T); ok {
			*ptr = v
			return nil
		}
		return NewError(InvalidArgument, "msg must be a pointer to the expected type")
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

func (s *ChanStream[T]) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true
	s.cancel()
	close(s.sendCh)
	return nil
}

// SendChan returns the send channel for direct access.
func (s *ChanStream[T]) SendChan() chan<- T {
	return s.sendCh
}

// RecvChan returns the receive channel for direct access.
func (s *ChanStream[T]) RecvChan() <-chan T {
	return s.recvCh
}

// SetRecvChan sets the receive channel (used to wire up bidirectional streams).
func (s *ChanStream[T]) SetRecvChan(ch chan T) {
	s.recvCh = ch
}
