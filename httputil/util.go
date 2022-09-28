// Package httputil contains lots of helper functions
package httputil

import (
	"fmt"
	"net/http"
	"net/url"
)

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

func CombineUrlAndQuery(targetUrl string, query map[string]interface{}) (finalUrl string, err error) {
	// query has higher level
	var u *url.URL
	u, err = url.Parse(targetUrl)
	if err != nil {
		return
	}

	var mQuery url.Values
	mQuery, err = url.ParseQuery(u.RawQuery)
	if err != nil {
		return
	}

	if query != nil {
		for k, v := range query {
			mQuery.Set(k, fmt.Sprintf("%v", v))
		}
	}

	u.RawQuery = mQuery.Encode()

	finalUrl = u.String()

	return
}
