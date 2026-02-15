# Talk - Transport Abstraction Layer Kit

> `go.zoe.im/x/talk`

Talk 是一个传输层抽象库，让使用者只需定义业务方法，无需关心底层连接实现。通过配置即可切换 HTTP、gRPC、WebSocket、Unix Socket 等传输协议。

## 安装

```bash
go get go.zoe.im/x/talk
```

## 快速开始

### 1. 定义服务

```go
type UserService interface {
    GetUser(ctx context.Context, id string) (*User, error)
    CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
    ListUsers(ctx context.Context) ([]*User, error)
    WatchUsers(ctx context.Context) (<-chan *UserEvent, error) // 流式
}

type userServiceImpl struct {
    users map[string]*User
}

func (s *userServiceImpl) GetUser(ctx context.Context, id string) (*User, error) {
    user, ok := s.users[id]
    if !ok {
        return nil, talk.NewError(talk.NotFound, "user not found")
    }
    return user, nil
}
// ... 实现其他方法
```

### 2. 创建 Server

```go
import (
    "go.zoe.im/x"
    "go.zoe.im/x/talk"
    _ "go.zoe.im/x/talk/extract"               // 注册默认 Extractor
    _ "go.zoe.im/x/talk/transport/http/std"    // 注册 HTTP 传输
)

func main() {
    userSvc := NewUserService()

    // 创建 Server，设置默认路径前缀
    cfg := x.TypedLazyConfig{
        Type:   "http",
        Config: json.RawMessage(`{"addr": ":8080"}`),
    }
    server, _ := talk.NewServerFromConfig(cfg, talk.WithPathPrefix("/api/v1"))
    
    // 注册服务（自动使用默认前缀）
    server.Register(userSvc)

    // 启动
    ctx := context.Background()
    server.Serve(ctx)
}
```

### 3. 创建 Client

```go
cfg := x.TypedLazyConfig{
    Type:   "http",
    Config: json.RawMessage(`{"addr": "http://localhost:8080"}`),
}

client, _ := talk.NewClientFromConfig(cfg)
defer client.Close()

var user User
err := client.Call(ctx, "/api/v1/user/123", nil, &user)
```

## 支持的传输协议

| Type | 别名 | 说明 | Import |
|------|------|------|--------|
| `http` | `http/std`, `http/default` | HTTP (net/http) | `_ "go.zoe.im/x/talk/transport/http/std"` |
| `http/gin` | - | HTTP (Gin) | `_ "go.zoe.im/x/talk/transport/http/gin"` |
| `grpc` | - | gRPC | `_ "go.zoe.im/x/talk/transport/grpc"` |
| `websocket` | `ws` | WebSocket | `_ "go.zoe.im/x/talk/transport/websocket"` |
| `unix` | `unix-socket` | Unix Domain Socket | `_ "go.zoe.im/x/talk/transport/unix"` |

## 配置示例

### HTTP

```json
{
    "addr": ":8080",
    "read_timeout": "30s",
    "write_timeout": "30s",
    "swagger": {
        "enabled": true,
        "path": "/swagger",
        "title": "My API",
        "version": "1.0.0"
    }
}
```

### HTTP (Gin)

```json
{
    "addr": ":8080",
    "swagger": {
        "enabled": true
    }
}
```

### WebSocket

```json
{
    "addr": ":8081",
    "path": "/ws"
}
```

### gRPC

```json
{
    "addr": ":9090"
}
```

### Unix Socket

```json
{
    "path": "/var/run/myapp.sock"
}
```

## Swagger 文档

HTTP 传输（std 和 Gin）支持自动生成 Swagger/OpenAPI 文档：

```go
cfg := x.TypedLazyConfig{
    Type: "http",
    Config: json.RawMessage(`{
        "addr": ":8080",
        "swagger": {
            "enabled": true,
            "path": "/swagger",
            "title": "User Service API",
            "description": "API for managing users",
            "version": "1.0.0"
        }
    }`),
}

server, _ := talk.NewServerFromConfig(cfg, talk.WithPathPrefix("/api/v1"))
server.Register(userSvc)
server.Serve(ctx)

// Swagger UI: http://localhost:8080/swagger/
// OpenAPI spec: http://localhost:8080/swagger/openapi.json
```

## 切换协议

只需更改配置，无需改代码：

```go
// HTTP -> WebSocket: 改 Type 即可
cfg := x.TypedLazyConfig{
    Type:   "websocket", // 改这里
    Config: json.RawMessage(`{"addr": ":8081", "path": "/ws"}`),
}

// HTTP -> Unix Socket
cfg := x.TypedLazyConfig{
    Type:   "unix",
    Config: json.RawMessage(`{"path": "/var/run/app.sock"}`),
}
```

## Endpoint 提取

### 反射提取（推荐）

使用 `Register` 方法自动提取并注册 Endpoint。导入 `extract` 包会自动注册 `ReflectExtractor` 作为默认提取器：

```go
import _ "go.zoe.im/x/talk/extract"  // 注册默认 Extractor

// 方式一：Server 级别设置默认前缀
server, _ := talk.NewServerFromConfig(cfg, talk.WithPathPrefix("/api/v1"))
server.Register(&userServiceImpl{})  // 自动使用 /api/v1 前缀

// 方式二：Register 时指定前缀（覆盖默认）
server.Register(&otherServiceImpl{}, talk.WithPrefix("/api/v2"))
```

**函数名推导规则：**

| 前缀 | HTTP Method | Path |
|------|-------------|------|
| `Get` | GET | `/{resource}/{id}` |
| `List` | GET | `/{resources}` |
| `Create` | POST | `/{resource}` |
| `Update` | PUT | `/{resource}/{id}` |
| `Delete` | DELETE | `/{resource}/{id}` |
| `Watch` | GET (SSE) | `/{resource}/watch` |

### 使用注解自定义 Endpoint

服务可以实现 `MethodAnnotations` 接口来自定义 Endpoint 配置或跳过某些方法：

```go
type userService struct{}

func (s *userService) GetUser(ctx context.Context, id string) (*User, error) { ... }
func (s *userService) InternalMethod(ctx context.Context) error { ... }
func (s *userService) CustomEndpoint(ctx context.Context) (*Result, error) { ... }

// TalkAnnotations 返回方法注解
func (s *userService) TalkAnnotations() map[string]string {
    return map[string]string{
        "InternalMethod": "@talk skip",                           // 跳过注册
        "CustomEndpoint": "@talk path=/custom method=PUT",        // 自定义路径和方法
    }
}
```

**注解格式：**
- `@talk skip` 或 `@talk ignore` - 跳过该方法，不注册为 Endpoint
- `@talk path=/custom/path` - 自定义路径
- `@talk method=PUT` - 自定义 HTTP 方法
- `@talk stream=server` - 设置流模式 (server/client/bidi)

### 手动注册

```go
server.RegisterEndpoints(
    talk.NewEndpoint("ping", func(ctx context.Context, req any) (any, error) {
        return map[string]string{"status": "ok"}, nil
    }, talk.WithPath("/ping"), talk.WithMethod("GET")),
)
```

## 流式支持

### Server-Side Streaming (SSE)

```go
type EventService interface {
    // 返回 <-chan 自动识别为服务端流
    WatchEvents(ctx context.Context) (<-chan *Event, error)
}

func (s *eventServiceImpl) WatchEvents(ctx context.Context) (<-chan *Event, error) {
    ch := make(chan *Event)
    go func() {
        defer close(ch)
        for i := 0; i < 10; i++ {
            ch <- &Event{ID: i, Type: "update"}
            time.Sleep(time.Second)
        }
    }()
    return ch, nil
}
```

### 手动流式 Endpoint

```go
ep := talk.NewStreamEndpoint(
    "WatchEvents",
    func(ctx context.Context, req any, stream talk.Stream) error {
        for i := 0; i < 10; i++ {
            if err := stream.Send(&Event{ID: i}); err != nil {
                return err
            }
        }
        return nil
    },
    talk.StreamServerSide,
    talk.WithPath("/events/watch"),
)
```

## 错误处理

```go
// 创建错误
err := talk.NewError(talk.NotFound, "user not found")
err := talk.NewErrorWithDetails(talk.InvalidArgument, "validation failed", details)

// 错误码自动映射
err.HTTPStatus() // 404
err.GRPCCode()   // codes.NotFound

// 检查错误
if talkErr, ok := talk.IsError(err); ok {
    fmt.Printf("Code: %s, Message: %s\n", talkErr.Code, talkErr.Message)
}
```

**错误码映射：**

| ErrorCode | HTTP Status | gRPC Code |
|-----------|-------------|-----------|
| NotFound | 404 | NotFound |
| InvalidArgument | 400 | InvalidArgument |
| Unauthenticated | 401 | Unauthenticated |
| PermissionDenied | 403 | PermissionDenied |
| Internal | 500 | Internal |

## 注册自定义传输

```go
talk.RegisterTransport("custom", &talk.TransportCreators{
    Server: func(cfg x.TypedLazyConfig) (talk.Transport, error) {
        return NewCustomServer(cfg)
    },
    Client: func(cfg x.TypedLazyConfig) (talk.Transport, error) {
        return NewCustomClient(cfg)
    },
}, "custom-alias")
```

## 目录结构

```
talk/
├── talk.go                # Server/Client 抽象
├── endpoint.go            # Endpoint 定义
├── errors.go              # 统一错误处理
├── stream.go              # 流式支持
├── config.go              # 统一传输注册
│
├── codec/                 # 编解码器
│   ├── codec.go           # Codec 接口
│   └── json.go            # JSON 实现
│
├── extract/               # Endpoint 提取器
│   ├── extract.go         # 接口定义
│   └── reflect.go         # 反射提取
│
├── gen/                   # 代码生成
│   └── gen.go
│
├── swagger/               # Swagger 文档生成
│   ├── swagger.go         # OpenAPI 生成器
│   └── handler.go         # HTTP Handler
│
└── transport/             # 传输实现
    ├── transport.go       # Transport 接口
    ├── http/
    │   ├── http.go        # HTTP 配置
    │   ├── std/           # net/http 实现
    │   └── gin/           # Gin 实现
    ├── grpc/              # gRPC 实现
    ├── websocket/         # WebSocket 实现
    └── unix/              # Unix Socket 实现
```

## License

Apache License 2.0
