// Package x provides a collection of utility functions and types for Go applications.
//
// This package follows the DRY (Don't Repeat Yourself) principle, providing common
// utilities that can be reused across projects. It includes:
//
//   - Array/Slice utilities: Contains, ContainsFunc, Filter, Map
//   - Backoff: Exponential backoff with jitter for retry logic
//   - Clock: Injectable clock interface for testing
//   - Configuration: TypedLazyConfig for lazy configuration loading
//   - Deep Copy: Generic deep copy using reflection
//   - Docker: Container detection utilities
//   - Errors: Multi-error aggregation
//   - Grace: Graceful shutdown utilities
//   - Home Directory: Cross-platform home directory detection
//   - IO: Line-based writer utilities
//   - Key Management: PEM key parsing and generation
//   - Map utilities: Keys, Values, Range, UpdateMap
//   - Pattern Matching: Glob pattern caching
//   - Operators: Kubernetes-style label selector operators
//   - Rate Limiting: Token bucket rate limiter
//   - Selectors: Template-based object field selection
//   - String/Bytes: Zero-allocation conversions
//   - SyncMap: Type-safe generic concurrent map
//   - Time: Duration with JSON/YAML marshaling, timeout utilities
//   - Types: Generic Min/Max functions
//   - UUID: RFC 4122 UUID generation and parsing
//   - Value: Fluent conditional value handling
//
// # Installation
//
//	go get go.zoe.im/x
//
// # Example
//
//	package main
//
//	import "go.zoe.im/x"
//
//	func main() {
//	    // Check if slice contains a value
//	    if x.Contains([]int{1, 2, 3}, 2) {
//	        println("found!")
//	    }
//
//	    // Generate a new UUID
//	    uuid := x.NewUUID()
//	    println(uuid.String())
//	}
package x
