<div align="center">

# Utility library for Go

> `go.zoe.im/x`

[![Build Status](https://dev.azure.com/jiusanzhou/x/_apis/build/status/jiusanzhou.x?branchName=master)](https://dev.azure.com/jiusanzhou/x/_build/latest?definitionId=1&branchName=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/jiusanzhou/x)](https://goreportcard.com/report/github.com/jiusanzhou/x)
[![codecov](https://codecov.io/gh/jiusanzhou/x/branch/master/graph/badge.svg)](https://codecov.io/gh/jiusanzhou/x)
[![GoDoc](https://godoc.org/github.com/jiusanzhou/x?status.svg)](https://godoc.org/github.com/jiusanzhou/x)
[![On Sourcegraph](https://sourcegraph.com/github.com/jiusanzhou/x/-/badge.svg)](https://sourcegraph.com/github.com/jiusanzhou/x?badge)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)

</div>

| Don't repeat yourself, every piece of knowledge must have a single, unambiguous, authoritative representation within a system. |
| :----------------------------------------------------------------------------------------------------------------------------: |

## Table of Contents

- [Installation](#installation)
- [Array/Slice Utilities](#arrayslice-utilities)
- [Map Utilities](#map-utilities)
- [Concurrency](#concurrency)
- [Time/Duration](#timeduration)
- [UUID Generation](#uuid-generation)
- [Error Handling](#error-handling)
- [Backoff/Retry](#backoffretry)
  - [Retry with Backoff](#retry-with-backoff)
  - [Backoff Strategies](#backoff-strategies)
  - [Backoff Middleware](#backoff-middleware)
- [Rate Limiting](#rate-limiting)
- [Clock Interface](#clock-interface)
- [Configuration](#configuration)
- [Deep Copy](#deep-copy)
- [Container Detection](#container-detection)
- [Graceful Shutdown](#graceful-shutdown)
- [Home Directory](#home-directory)
- [IO Utilities](#io-utilities)
- [Key Management](#key-management)
- [Pattern Matching](#pattern-matching)
- [Selector/Operators](#selectoroperators)
- [String/Bytes Conversion](#stringbytes-conversion)
- [Generic Utilities](#generic-utilities)
- [Sub-packages](#sub-packages)

## Installation

```bash
go get go.zoe.im/x
```

## Array/Slice Utilities

### Contains

Reports whether target is present in items.

```go
nums := []int{1, 2, 3, 4, 5}
if x.Contains(nums, 3) {
    fmt.Println("found 3")
}
```

### ContainsFunc

Reports whether any element satisfies the predicate.

```go
hasAlice := x.ContainsFunc(users, func(u User) bool {
    return u.Name == "Alice"
})
```

### Filter

Returns elements that satisfy the predicate.

```go
evens := x.Filter(nums, func(n int) bool {
    return n%2 == 0
})
```

### Map

Returns transformed elements.

```go
doubled := x.Map(nums, func(n int) int {
    return n * 2
})
```

## Map Utilities

### Keys / Values

```go
m := map[string]int{"a": 1, "b": 2}
keys := x.Keys(m)     // ["a", "b"]
values := x.Values(m) // [1, 2]
```

### Range

Iterates with early termination.

```go
x.Range(m, func(k string, v int) bool {
    fmt.Printf("%s: %d\n", k, v)
    return true
})
```

### UpdateMap

Updates map and returns changed/deleted entries.

```go
changed, deleted := x.UpdateMap(original, inputs, convertFn, keyFn, false)
```

## Concurrency

### SyncMap

Type-safe generic concurrent map.

```go
var cache x.SyncMap[string, int]
cache.Store("key", 42)
value, ok := cache.Load("key")
```

## Time/Duration

### Duration

Wrapper with JSON/YAML marshaling support.

```go
type Config struct {
    Timeout x.Duration `json:"timeout"`
}
// JSON: {"timeout": "30s"}
```

### RunWithTimeout

Execute with timeout, returns true if timed out.

```go
timeout := x.RunWithTimeout(func(exit *bool) {
    for !*exit { /* work */ }
}, 5*time.Second)
```

## UUID Generation

```go
uuid := x.NewUUID()
fmt.Println(uuid.String())

parsed, err := x.ParseUUID("550e8400-e29b-41d4-a716-446655440000")
```

## Error Handling

### Errors

Aggregates multiple errors.

```go
var errs x.Errors
errs.Add(errors.New("error 1"))
errs.Add(errors.New("error 2"))
if !errs.IsNil() {
    fmt.Println(errs.Error())
}
```

## Backoff/Retry

### Retry with Backoff

Execute operations with configurable retry strategies and backoff algorithms.

```go
ctx := context.Background()

// Basic retry with exponential backoff
backoff := x.NewExponentialBackoff(100*time.Millisecond, 10*time.Second)
err := x.Retry(ctx, backoff, func(ctx context.Context) error {
    if err := doSomething(); err != nil {
        return x.RetryableError(err) // Mark as retryable
    }
    return nil // Success
})

// Convenience function
err = x.Exponential(ctx, 100*time.Millisecond, 10*time.Second, func(ctx context.Context) error {
    return x.RetryableError(db.Ping())
})
```

### Backoff Strategies

```go
// Constant: 1s -> 1s -> 1s -> 1s
b := x.NewConstantBackoff(1 * time.Second)

// Exponential: 1s -> 2s -> 4s -> 8s -> 10s (capped)
b = x.NewExponentialBackoff(1*time.Second, 10*time.Second)

// Fibonacci: 1s -> 1s -> 2s -> 3s -> 5s -> 8s -> 10s (capped)
b = x.NewFibonacciBackoff(1*time.Second, 10*time.Second)
```

### Backoff Middleware

```go
// Limit retries
b := x.WithMaxRetries(5, x.NewExponentialBackoff(100*time.Millisecond, 10*time.Second))

// Cap individual delay
b = x.WithCappedDuration(5*time.Second, b)

// Limit total retry time
b = x.WithMaxDuration(30*time.Second, b)

// Add jitter to prevent thundering herd
b = x.WithJitter(100*time.Millisecond, b)
b = x.WithJitterPercent(10, b) // +/- 10%
```

### Legacy Backoff (per-key tracking)

```go
backoff := x.NewBackOffWithJitter(100*time.Millisecond, 10*time.Second, 0.5)
delay := backoff.Get("operation-id")
backoff.Next("operation-id", time.Now())
backoff.Reset("operation-id")
```

## Rate Limiting

```go
limiter := x.NewTokenBucketRateLimiter(10.0, 5)
if limiter.TryAccept() { /* process */ }
limiter.Accept() // blocks until available
```

## Clock Interface

Injectable clock for testing.

```go
clock := x.RealClock{}
now := clock.Now()
clock.Sleep(time.Second)
```

## Configuration

### TypedLazyConfig

```go
config := &x.TypedLazyConfig{
    Name: "myconfig",
    Type: "database",
    Config: json.RawMessage(`{"host":"localhost"}`),
}
var dbConfig DatabaseConfig
config.Unmarshal(&dbConfig)
```

## Deep Copy

```go
original := map[string][]int{"a": {1, 2, 3}}
copied := x.DeepCopy(original)
```

## Container Detection

```go
inContainer, err := x.IsInContainer(os.Getpid())
```

## Graceful Shutdown

```go
err := x.GraceStart(func(stopCh x.GraceSignalChan) error {
    <-stopCh
    return nil
})

err := x.GraceRun(func() error {
    return http.ListenAndServe(":8080", nil)
})
```

## Home Directory

```go
home, err := x.HomeDir()
path, err := x.WithHomeDir("~/.config/app")
```

## IO Utilities

### LineWriter

```go
writer := x.LineWriter(func(line []byte) error {
    fmt.Println(string(line))
    return nil
}, true)
```

## Key Management

```go
keyPEM, err := x.MakeEllipticPrivateKeyPEM()
key, err := x.ParsePrivateKeyPEM(pemData)
```

## Pattern Matching

```go
pattern := x.Glob("*.txt")
if pattern.Match("readme.txt") { /* matched */ }
```

## Selector/Operators

```go
selectors := x.Selectors{
    {Key: ".Name", Operator: x.OperatorIn, Values: []string{"alice"}},
}
selectors.Init()
if selectors.Match(user) { /* matched */ }
```

Operators: `OperatorIn`, `OperatorNotIn`, `OperatorExists`, `OperatorNotExists`, `OperatorGt`, `OperatorLt`, `OperatorRange`

## String/Bytes Conversion

Zero-allocation conversions (unsafe).

```go
b := x.Str2Bytes("hello")
s := x.Bytes2Str(b)
```

## Generic Utilities

### Min/Max

```go
smaller := x.Min(int64(10), int64(20))
larger := x.Max(3.14, 2.71)
```

### Value

Fluent conditional values.

```go
result := x.V(config.Port).Or(8080).Value()
result := x.V(value).If(condition).Or(defaultValue).Value()
```

## Sub-packages

### factory

Generic factory pattern.

```go
f := factory.NewFactory[Plugin, Option]()
f.Register("example", creator)
plugin, err := f.Create(cfg)
```

### sh

Shell execution with mvdan.cc/sh.

```go
sh.Run("echo hello")
sh.Run("@script.sh")
```

### jsonmerge

Deep merge maps/structs.

```go
jsonmerge.Merge(&dst, src)
```

### httputil

HTTP utilities and API responses.

```go
httputil.NewResponse(w).WithData(result).Flush()
httputil.CloneRequest(req)
```

### service

System service management (darwin/linux/windows).

```go
svc, _ := service.New("myservice", "description")
svc.Install()
svc.Start()
```

### version

Semantic versioning.

```go
info := version.Get()
v1, _ := version.NewSemver("1.2.3")
```

### cgroup

Linux cgroup utilities.

```go
import "go.zoe.im/x/cgroup/automaxprocs"
automaxprocs.Set()
```

### cli

CLI builder with config support.

```go
cmd := cli.New(cli.Name("myapp"), cli.WithConfig(&Config{}))
cmd.Run()
```

## License

Apache License 2.0
