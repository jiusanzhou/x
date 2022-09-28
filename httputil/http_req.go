/*
 * Copyright (c) 2022 wellwell.work, LLC by Zoe
 *
 * Licensed under the Apache License 2.0 (the "License");
 * You may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package httputil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

var (
	DefaultSession = &Session{}
)

type Session struct{}

type ReqObject struct {
	// for request
	rreq *http.Request

	client  *http.Client
	timeout time.Duration

	method string
	url    string

	// ====

	// query, data
	// TODO: support multiple key
	headers map[string]string
	query   map[string]interface{}
	data    interface{}

	body io.Reader

	// ===

	// for response
	rresp *http.Response
}

func (ro *ReqObject) update(opts ...ReqOption) *ReqObject {
	for _, o := range opts {
		o(ro)
	}
	return ro
}

func (ro *ReqObject) buildRequest() (*http.Request, error) {
	var err error

	// combine query to url params
	ro.url, err = CombineUrlAndQuery(ro.url, ro.query)

	// build body
	var body io.Reader
	if ro.body != nil {
		body = ro.body
	} else if ro.data != nil {
		// default with json encoding
		// TODO: add more
		bs, err := json.Marshal(ro.data)
		// 序列化失败返回错误
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(bs)
		ro.headers["Content-Type"] = "application/json"
	}

	ro.rreq, err = http.NewRequest(ro.method, ro.url, body)

	ro.updateClient()

	return ro.rreq, err
}

func (ro *ReqObject) updateClient() {
	// if any set session or timeout is set
	// we need to create a new client, DO NOT modify default one

	// set timeout for client
	if ro.timeout != 0 {
		ro.client.Timeout = ro.timeout
	}
}

func newReqObject(url string) *ReqObject {
	ro := &ReqObject{
		client: http.DefaultClient,
		url:    url,
		method: "GET",
	}
	return ro
}

type ReqOption func(*ReqObject)

func Request(url string, opts ...ReqOption) error {
	ro := newReqObject(url).update(opts...)

	// build the raw request
	rreq, err := ro.buildRequest()
	if err != nil {
		return err
	}

	rresp, err := ro.client.Do(rreq)

	return nil
}
