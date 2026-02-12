// Copyright (c) 2020 wellwell.work, LLC by Zoe
//
// Licensed under the Apache License 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
Package httputil provides HTTP utilities for building REST APIs.

This package includes helpers for constructing JSON API responses with
consistent status codes, error handling, and data serialization.

# Response Builder

Create and send JSON responses:

	func handler(w http.ResponseWriter, r *http.Request) {
	    result, err := doSomething()

	    httputil.NewResponse(w).
	        WithDataOrErr(result, err).
	        Flush()
	}

# Response Methods

The Response type provides a fluent interface:

	// Set response data
	resp.WithData(data)

	// Set data and error together
	resp.WithDataOrErr(data, err)

	// Set error message
	resp.WithError(err)
	resp.WithErrorf("failed: %v", err)

	// Set status code
	resp.WithCode(httputil.CodeOK)

# Status Codes

The package defines common status codes that map to HTTP status codes:

	CodeOK           // 200 OK
	CodeInvalidParam // 400 Bad Request
	CodeUnauthorized // 401 Unauthorized
	CodeForbidden    // 403 Forbidden
	CodeNotFound     // 404 Not Found
	CodeInternal     // 500 Internal Server Error

# Response Format

All responses follow a consistent JSON structure:

	{
	    "code": 0,
	    "status": "success",
	    "data": {...},
	    "error": ""
	}
*/
package httputil
