// Package httputil contains lots of helper functions
package httputil

import "net/http"

// CloneRequest creates a shallow copy of the request along with a deep copy of the Headers.
func CloneRequest(req *http.Request) *http.Request {
	r := new(http.Request)

	// shallow clone
	*r = *req

	// deep copy headers
	r.Header = CloneHeader(req.Header)

	return r
}

// CloneHeader creates a deep copy of an http.Header.
func CloneHeader(in http.Header) http.Header {
	out := make(http.Header, len(in))
	for key, values := range in {
		newValues := make([]string, len(values))
		copy(newValues, values)
		out[key] = newValues
	}
	return out
}
