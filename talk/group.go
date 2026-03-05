package talk

// Group creates a sub-group with its own prefix and middleware.
// Endpoints registered through the group inherit the group's prefix
// and middleware, in addition to any server-level settings.
//
// Usage:
//
//	admin := server.Group("/admin", adminAuthMiddleware)
//	admin.Register(adminService)
//
//	public := server.Group("/public")
//	public.Register(publicService)
type Group struct {
	server     *Server
	pathPrefix string
	middleware []MiddlewareFunc
}

// Group creates a new endpoint group with the given prefix and middleware.
func (s *Server) Group(prefix string, mw ...MiddlewareFunc) *Group {
	return &Group{
		server:     s,
		pathPrefix: s.pathPrefix + prefix,
		middleware: mw,
	}
}

// Register extracts endpoints from a service and registers them with
// the group's prefix and middleware.
func (g *Group) Register(service any, opts ...RegisterOption) error {
	if g.server.extractor == nil {
		return NewError(FailedPrecondition, "no extractor configured")
	}

	cfg := &registerConfig{
		pathPrefix: g.pathPrefix,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	endpoints, err := g.server.extractor.Extract(service)
	if err != nil {
		return err
	}

	for _, ep := range endpoints {
		if cfg.pathPrefix != "" {
			ep.Path = cfg.pathPrefix + ep.Path
		}
		// Apply: server middleware → group middleware → endpoint middleware
		combined := make([]MiddlewareFunc, 0, len(g.server.middleware)+len(g.middleware)+len(ep.Middleware))
		combined = append(combined, g.server.middleware...)
		combined = append(combined, g.middleware...)
		combined = append(combined, ep.Middleware...)
		ep.Middleware = combined
	}

	g.server.endpoints = append(g.server.endpoints, endpoints...)
	return nil
}

// RegisterEndpoints adds pre-defined endpoints with the group's prefix and middleware.
func (g *Group) RegisterEndpoints(endpoints ...*Endpoint) {
	for _, ep := range endpoints {
		ep.Path = g.pathPrefix + ep.Path
		combined := make([]MiddlewareFunc, 0, len(g.server.middleware)+len(g.middleware)+len(ep.Middleware))
		combined = append(combined, g.server.middleware...)
		combined = append(combined, g.middleware...)
		combined = append(combined, ep.Middleware...)
		ep.Middleware = combined
	}
	g.server.endpoints = append(g.server.endpoints, endpoints...)
}

// Group creates a nested sub-group.
func (g *Group) Group(prefix string, mw ...MiddlewareFunc) *Group {
	combined := make([]MiddlewareFunc, 0, len(g.middleware)+len(mw))
	combined = append(combined, g.middleware...)
	combined = append(combined, mw...)
	return &Group{
		server:     g.server,
		pathPrefix: g.pathPrefix + prefix,
		middleware: combined,
	}
}
