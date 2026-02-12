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
Package clientlb provides client-side load balancing for HTTP requests.

This package implements an http.RoundTripper that distributes requests
across multiple backend endpoints with automatic health checking and
failover support.

# Basic Usage

Create a load balancer and use it with an HTTP client:

	bl, err := clientlb.NewBalancer(
	    clientlb.AddEndpoints("http://backend1:8080", "http://backend2:8080"),
	    clientlb.SetRoundTripper(http.DefaultTransport),
	    clientlb.SetHealthChecker(clientlb.NewSimpleHealthCheck("GET", "/healthz", "ok")),
	)

	client := &http.Client{Transport: bl}
	resp, err := client.Get("http://any-host/api/v1/resource")

# Features

  - Random endpoint selection (round-robin planned)
  - Automatic health checking with configurable interval
  - Unhealthy endpoint recovery
  - Request retrying on failure

# Health Checking

The balancer periodically checks unhealthy endpoints and restores them
when they become healthy again. Configure the health check behavior:

	clientlb.SetHealthChecker(func(ep *Endpoint, req *http.Request, resp *http.Response) bool {
	    if resp == nil {
	        return false
	    }
	    return resp.StatusCode == 200
	})

# RoundTripper Integration

Use NewBalancerRoundTripper for easy integration:

	transport := clientlb.NewBalancerRoundTripper("http://host1:8080", "http://host2:8080")
	client := &http.Client{Transport: transport(http.DefaultTransport)}
*/
package clientlb
