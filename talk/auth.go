package talk

import (
	"context"
	"strings"
)

// AuthLevel defines the authentication requirement for an endpoint.
type AuthLevel string

const (
	AuthNone  AuthLevel = "none"  // No authentication required
	AuthToken AuthLevel = "token" // Requires a valid token
	AuthAdmin AuthLevel = "admin" // Requires admin privileges
)

// AuthFunc validates a request and returns the authenticated identity or an error.
// The returned identity (e.g. user ID, role) is stored in context for downstream use.
type AuthFunc func(ctx context.Context, req any) (identity string, err error)

// AuthMiddleware creates a MiddlewareFunc that enforces authentication based on
// the endpoint's "auth" metadata. Endpoints without auth metadata or with
// auth=none are passed through without checking.
//
// Usage:
//
//	server := talk.NewServer(transport,
//	    talk.WithServerMiddleware(talk.AuthMiddleware(myAuthFunc)),
//	)
func AuthMiddleware(authFn AuthFunc) MiddlewareFunc {
	return func(next EndpointFunc) EndpointFunc {
		return func(ctx context.Context, req any) (any, error) {
			// Check if endpoint has auth metadata
			ep := EndpointFromContext(ctx)
			if ep == nil {
				return next(ctx, req)
			}

			level := authLevelFromEndpoint(ep)
			if level == AuthNone || level == "" {
				return next(ctx, req)
			}

			identity, err := authFn(ctx, req)
			if err != nil {
				return nil, NewError(Unauthenticated, "authentication failed: "+err.Error())
			}

			ctx = context.WithValue(ctx, ctxKeyIdentity, identity)
			ctx = context.WithValue(ctx, ctxKeyAuthLevel, level)

			return next(ctx, req)
		}
	}
}

// IdentityFromContext returns the authenticated identity from context.
func IdentityFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxKeyIdentity).(string)
	return v, ok
}

// AuthLevelFromContext returns the required auth level from context.
func AuthLevelFromContext(ctx context.Context) AuthLevel {
	v, _ := ctx.Value(ctxKeyAuthLevel).(AuthLevel)
	return v
}

func authLevelFromEndpoint(ep *Endpoint) AuthLevel {
	if ep.Metadata == nil {
		return ""
	}
	v, ok := ep.Metadata["auth"]
	if !ok {
		return ""
	}
	switch s := v.(type) {
	case string:
		return AuthLevel(strings.ToLower(s))
	case AuthLevel:
		return s
	default:
		return ""
	}
}

type ctxKey int

const (
	ctxKeyIdentity ctxKey = iota
	ctxKeyAuthLevel
	ctxKeyEndpoint
)

// WithEndpointContext returns a new context carrying the endpoint.
func WithEndpointContext(ctx context.Context, ep *Endpoint) context.Context {
	return context.WithValue(ctx, ctxKeyEndpoint, ep)
}

// EndpointFromContext returns the endpoint from context, if set.
func EndpointFromContext(ctx context.Context) *Endpoint {
	ep, _ := ctx.Value(ctxKeyEndpoint).(*Endpoint)
	return ep
}
