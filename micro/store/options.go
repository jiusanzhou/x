package store

// NewListOptions returns
func NewListOptions(ops ...ListOption) *ListOptions {
	opts := &ListOptions{}
	for _, o := range ops {
		o(opts)
	}
	return opts
}

// ListOption ia a function to set ListOptions
type ListOption func(opts *ListOptions)

// ListOptions is the query options.
type ListOptions struct {
	// TODO: query, order, skip?
	Query, Order string // use type wth interface{}
	Skip         int64

	// Watch for changes to the described resources and return them as a stream of
	// add, update, and remove notifications. Specify resourceVersion.
	Watch bool

	// Timeout duration in seconds for the call.
	// This limits the duration of the call, regardless of any activity or inactivity.
	Timeout int64

	// limit is a maximum number of responses to return for a list call. If more items exist, the
	// server will set the `continue` field on the list metadata to a value that can be used with the
	// same initial query to retrieve the next set of results. Setting a limit may return fewer than
	// the requested amount of items (up to zero items) in the event all requested objects are
	// filtered out and clients should only use the presence of the continue field to determine whether
	// more results are available. Servers may choose not to support the limit argument and will return
	// all of the available results. If limit is specified and the continue field is empty, clients may
	// assume that no more results are available. This field is not supported if watch is true.
	//
	// The server guarantees that the objects returned when using continue will be identical to issuing
	// a single list call without a limit - that is, no objects created, modified, or deleted after the
	// first request is issued will be included in any subsequent continued requests. This is sometimes
	// referred to as a consistent snapshot, and ensures that a client that is using limit to receive
	// smaller chunks of a very large result can ensure they see all possible objects. If objects are
	// updated during a chunked list the version of the object that was present at the time the first list
	// result was calculated is returned.
	Limit int64
	// The continue option should be set when retrieving more results from the server. Since this value is
	// server defined, clients may only use the continue value from a previous query result with identical
	// query parameters (except for the value of continue) and the server may reject a continue value it
	// does not recognize. If the specified continue value is no longer valid whether due to expiration
	// (generally five to fifteen minutes) or a configuration change on the server, the server will
	// respond with a 410 ResourceExpired error together with a continue token. If the client needs a
	// consistent list, it must restart their list without the continue field. Otherwise, the client may
	// send another list request with the token received with the 410 error, the server will respond with
	// a list starting from the next key, but from the latest snapshot, which is inconsistent from the
	// previous list results - objects that are created, modified, or deleted after the first list request
	// will be included in the response, as long as their keys are after the "next key".
	//
	// This field is not supported when watch is true. Clients may start a watch from the last
	// resourceVersion value returned by the server and not miss any modifications.
	Continue string
}

// NewGetOptions returns
func NewGetOptions(ops ...GetOption) *GetOptions {
	opts := &GetOptions{}
	for _, o := range ops {
		o(opts)
	}
	return opts
}

// GetOption is a function to set GetOptions
type GetOption func(opts *GetOptions)

// GetOptions is the standard query options.
type GetOptions struct {
	// When specified:
	// - if unset, then the result is returned from remote storage based on quorum-read flag;
	// - if it's 0, then we simply return what we currently have in cache, no guarantee;
	// - if set to non zero, then the result is at least as fresh as given rv.
	// TODO: Version
	Version string

	// Timeout duration in seconds for the call.
	// This limits the duration of the call, regardless of any activity or inactivity.
	Timeout int64
}

// NewCreateOptions returns
func NewCreateOptions(ops ...CreateOption) *CreateOptions {
	opts := &CreateOptions{}
	for _, o := range ops {
		o(opts)
	}
	return opts
}

// CreateOption is a function to set CreateOptions
type CreateOption func(opts *CreateOptions)

// CreateOptions may be provided when creating an object.
type CreateOptions struct {
	// Timeout duration in seconds for the call.
	// This limits the duration of the call, regardless of any activity or inactivity.
	Timeout int64

	// Over write is record exits
	Force bool

	// When present, indicates that modifications should not be
	// persisted. An invalid or unrecognized dryRun directive will
	// result in an error response and no further processing of the
	DryRun bool
}

// NewUpdateOptions returns ...
func NewUpdateOptions(ops ...UpdateOption) *UpdateOptions {
	opts := &UpdateOptions{}
	for _, o := range ops {
		o(opts)
	}
	return opts
}

// UpdateOption is a function to set UpdateOptions
type UpdateOption func(opts *UpdateOptions)

// UpdateOptions may be provided when updating an object.
type UpdateOptions struct {
	// Timeout duration in seconds for the call.
	// This limits the duration of the call, regardless of any activity or inactivity.
	Timeout int64

	// When present, indicates that modifications should not be
	// persisted. An invalid or unrecognized dryRun directive will
	// result in an error response and no further processing of the
	DryRun bool
}

// NewDeleteOptions returns ...
func NewDeleteOptions(ops ...DeleteOption) *DeleteOptions {
	opts := &DeleteOptions{}
	for _, o := range ops {
		o(opts)
	}
	return opts
}

// DeleteOption is a function to set DeleteOptions
type DeleteOption func(opts *DeleteOptions)

// DeleteOptions may be provided when deleting an object.
type DeleteOptions struct {
	// Timeout duration in seconds for the call.
	// This limits the duration of the call, regardless of any activity or inactivity.
	Timeout int64

	// The duration in seconds before the object should be deleted.
	// Defaults to a per object value if not specified. zero means delete immediately.
	GracePeriod int64

	// When present, indicates that modifications should not be
	// persisted. An invalid or unrecognized dryRun directive will
	// result in an error response and no further processing of the
	DryRun bool
}