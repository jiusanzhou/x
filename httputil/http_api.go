package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type status string

// Field ....
type Field func(r *Response)

var (
	// StatusSuccess present this request success
	StatusSuccess = "success"
	// StatusFailed repsent this request failed
	StatusFailed = "failed"
)

// Response present all fields for api response
type Response struct {
	Code   StatusCode  `json:"code"`
	Data   interface{} `json:"data,omitempty"`
	Status string      `json:"status"`
	Error  string      `json:"error,omitempty"`

	w   http.ResponseWriter // if present we need to flush imme
	err error
}

// Flush data to response writer
func (r *Response) Flush(fs ...Field) {

	for _, f := range fs {
		f(r)
	}

	// combine error
	if r.err != nil {
		if r.Error != "" {
			r.Error = fmt.Sprintf("%s; error: %s", r.Error, r.err)
		} else {
			r.Error = r.err.Error()
		}
	}

	if r.Code != CodeOK || r.Error != "" {
		r.Status = StatusFailed
		if r.Code == CodeOK {
			// set default code
			r.Code = CodeInternal
		}
	} else {
		r.Status = StatusSuccess
	}

	r.w.Header().Set("Content-Type", "application/json")
	r.w.WriteHeader(httpStatusFromCode(r.Code))

	enc := json.NewEncoder(r.w)
	// TODO: catch the error
	enc.Encode(r)
}

// WithCode set the response code
func (r *Response) WithCode(c uint32) *Response {
	r.Code = StatusCode(c)
	return r
}

// WithData set the response data
func (r *Response) WithData(d interface{}) *Response {
	r.Data = d
	return r
}

// WithDataOrErr set data and error
func (r *Response) WithDataOrErr(d interface{}, err error) *Response {
	r.Data = d
	r.err = err
	return r
}

// WithErrorf set errorf
func (r *Response) WithErrorf(msg string, a ...interface{}) *Response {
	r.Error = fmt.Sprintf(msg, a...)
	return r
}

// WithError set error
func (r *Response) WithError(err error) *Response {
	r.err = err
	return r
}

// Code ...
func Code(c uint32) Field {
	return func(r *Response) {
		r.Code = StatusCode(c)
	}
}

// Data ...
func Data(d interface{}) Field {
	return func(r *Response) {
		r.Data = d
	}
}

// Errorf ...
func Errorf(msg string, a ...interface{}) Field {
	return func(r *Response) {
		r.Error = fmt.Sprintf(msg, a...)
	}
}

// Error ...
func Error(err error) Field {
	return func(r *Response) {
		r.err = err
	}
}

// NewResponse ...
func NewResponse(w http.ResponseWriter, fs ...Field) *Response {
	r := &Response{
		w: w,
	}

	for _, f := range fs {
		f(r)
	}

	return r
}
