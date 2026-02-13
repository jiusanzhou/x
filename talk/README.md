# Talk - Transport Abstraction Layer Kit

> `go.zoe.im/x/talk`

Talk 是一个传输层抽象库，让使用者只需定义业务方法，无需关心底层连接实现。通过配置即可切换 HTTP、gRPC、WebSocket 等传输协议。

## 设计目标

1. **协议无关** - 业务代码不依赖具体传输协议
2. **工厂模式** - 使用 `x.Factory` 注册和创建传输实现
3. **可插拔** - 每种协议可有多种实现（如 HTTP: net/http, Gin, Echo）
4. **配置驱动** - 通过配置切换传输方案，无需改代码
5. **双向支持** - 同时支持 Server 和 Client 模式

## 架构设计

```
┌─────────────────────────────────────────────────────────────────┐
│                        User Code                                │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  type UserService interface {                            │   │
│  │      // @talk path=/users/{id} method=GET                │   │
│  │      GetUser(ctx, id string) (*User, error)              │   │
│  │                                                          │   │
│  │      // 流式：chan 输入/输出                              │   │
│  │      Watch(ctx, req) (<-chan *Event, error)              │   │
│  │  }                                                       │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
            ┌─────────────────┴─────────────────┐
            ▼                                   ▼
   ┌─────────────────┐                ┌─────────────────┐
   │  反射注册        │                │  代码生成       │
   │  (runtime)      │                │  (go generate)  │
   └─────────────────┘                └─────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Talk Core                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │   Endpoint   │  │    Codec     │  │   Transport Factory  │  │
│  │  Extractor   │  │   Factory    │  │   (x.Factory)        │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Error Mapping                          │  │
│  │   AppError ←→ HTTP Status ←→ gRPC Code ←→ WS CloseCode   │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Transport Layer                              │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐    │
│  │      HTTP      │  │      gRPC      │  │   WebSocket    │    │
│  ├────────────────┤  ├────────────────┤  ├────────────────┤    │
│  │ Middleware:    │  │ Interceptor:   │  │ Middleware:    │    │
│  │ - Gin MW       │  │ - gRPC Native  │  │ - gorilla      │    │
│  │ - Echo MW      │  │                │  │ - nhooyr       │    │
│  │ - std Handler  │  │                │  │                │    │
│  └────────────────┘  └────────────────┘  └────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

## 目录结构

```
talk/
├── README.md              # 本文档
├── talk.go                # 核心入口
├── endpoint.go            # Endpoint 定义与提取
├── server.go              # Server 抽象
├── client.go              # Client 抽象
├── options.go             # 配置选项
├── errors.go              # 统一错误定义与映射
├── stream.go              # 流式支持
│
├── codec/                 # 编解码器（可插拔）
│   ├── codec.go           # Codec 接口与工厂
│   ├── json.go            # JSON 编解码
│   ├── protobuf.go        # Protobuf 编解码
│   └── msgpack.go         # MsgPack 编解码
│
├── extract/               # Endpoint 提取器
│   ├── extract.go         # 提取器接口
│   ├── reflect.go         # 反射提取（运行时）
│   └── comment.go         # 注释解析
│
├── gen/                   # 代码生成
│   └── gen.go             # go generate 工具
│
└── transport/             # 传输实现
    ├── transport.go       # Transport 接口
    ├── http/              # HTTP 传输
    │   ├── http.go        # HTTP Transport 接口
    │   ├── std/           # net/http 实现（使用 net/http 中间件）
    │   ├── gin/           # Gin 实现（使用 Gin 中间件）
    │   └── echo/          # Echo 实现（使用 Echo 中间件）
    ├── grpc/              # gRPC 传输（使用 gRPC Interceptor）
    │   └── grpc.go
    └── websocket/         # WebSocket 传输
        └── websocket.go
```

## 核心接口设计

### 1. Endpoint - 端点定义

```go
// Endpoint 表示一个服务端点
type Endpoint struct {
    Name        string            // 方法名，如 "GetUser"
    Path        string            // 路径，如 "/users/{id}"
    Method      string            // HTTP 方法，如 "GET"
    Handler     EndpointFunc      // 处理函数
    StreamMode  StreamMode        // 流模式
    Metadata    map[string]any    // 额外元数据
}

// EndpointFunc 是端点处理函数的通用签名
type EndpointFunc func(ctx context.Context, request any) (response any, err error)

// StreamMode 流模式（通过函数签名中的 chan 自动识别）
type StreamMode int

const (
    StreamNone       StreamMode = iota // 普通请求-响应
    StreamClientSide                   // 客户端流（参数含 <-chan）
    StreamServerSide                   // 服务端流（返回值含 <-chan）
    StreamBidirect                     // 双向流（参数和返回值都含 chan）
)
```

### 2. Endpoint 提取 - 两种方式

#### 方式一：反射提取（运行时）

从函数名自动推导 HTTP Method 和 Path：

```go
type UserService interface {
    GetUser(ctx context.Context, id string) (*User, error)      // GET /user/{id}
    CreateUser(ctx context.Context, req *CreateReq) (*User, error) // POST /user
    ListUsers(ctx context.Context, req *ListReq) (*ListResp, error) // GET /users
}

// 反射提取
endpoints := extract.FromInterface((*UserService)(nil), &userServiceImpl{})
```

**函数名推导规则：**

| 函数名前缀 | HTTP Method | Path 规则 |
|-----------|-------------|-----------|
| `Get`     | GET         | `/{resource}` 或 `/{resource}/{id}` |
| `List`    | GET         | `/{resources}` (复数) |
| `Create`  | POST        | `/{resource}` |
| `Update`  | PUT         | `/{resource}/{id}` |
| `Delete`  | DELETE      | `/{resource}/{id}` |
| `Watch`   | GET (SSE/WS)| `/{resource}/watch` |
| 其他      | POST        | `/{method_name}` |

#### 方式二：注释定义 + 代码生成

```go
type UserService interface {
    // @talk path=/users/{id} method=GET
    GetUser(ctx context.Context, id string) (*User, error)
    
    // @talk path=/users method=POST
    CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
    
    // @talk path=/users/{id}/events
    // 返回 <-chan，自动识别为服务端流
    WatchUser(ctx context.Context, id string) (<-chan *Event, error)
}

//go:generate go run go.zoe.im/x/talk/gen -type=UserService
```

### 3. 流式支持 - 通过 chan 自动识别

```go
type StreamService interface {
    // 服务端流：返回 <-chan
    // HTTP: SSE, gRPC: server stream, WS: server push
    Watch(ctx context.Context, req *WatchRequest) (<-chan *Event, error)
    
    // 客户端流：参数含 <-chan
    // HTTP: chunked upload, gRPC: client stream
    Upload(ctx context.Context, chunks <-chan *Chunk) (*UploadResult, error)
    
    // 双向流：参数和返回值都有 chan
    // gRPC: bidi stream, WS: full duplex
    Chat(ctx context.Context, in <-chan *Message) (<-chan *Message, error)
}
```

### 4. Codec - 可插拔编解码

```go
// Codec 编解码器接口
type Codec interface {
    Name() string
    Marshal(v any) ([]byte, error)
    Unmarshal(data []byte, v any) error
    ContentType() string  // 如 "application/json"
}

// Codec 工厂
var CodecFactory = factory.NewFactory[Codec, CodecOption]()

func init() {
    CodecFactory.Register("json", NewJSONCodec)
    CodecFactory.Register("protobuf", NewProtobufCodec)
    CodecFactory.Register("msgpack", NewMsgPackCodec)
}
```

### 5. 统一错误映射

```go
// Error 统一错误类型
type Error struct {
    Code    ErrorCode
    Message string
    Details any
}

// ErrorCode 错误码（参考 gRPC 错误码设计）
type ErrorCode int

const (
    OK                 ErrorCode = 0
    Cancelled          ErrorCode = 1
    Unknown            ErrorCode = 2
    InvalidArgument    ErrorCode = 3
    NotFound           ErrorCode = 4
    AlreadyExists      ErrorCode = 5
    PermissionDenied   ErrorCode = 6
    Unauthenticated    ErrorCode = 7
    ResourceExhausted  ErrorCode = 8
    Internal           ErrorCode = 9
    Unavailable        ErrorCode = 10
    // ...
)

// 错误码自动映射
func (c ErrorCode) HTTPStatus() int {
    // InvalidArgument -> 400
    // NotFound -> 404
    // Unauthenticated -> 401
    // PermissionDenied -> 403
    // Internal -> 500
    // ...
}

func (c ErrorCode) GRPCCode() codes.Code {
    // 直接映射到 gRPC codes
}

func (c ErrorCode) WSCloseCode() int {
    // 映射到 WebSocket close code
}

// 从各协议错误转换
func FromHTTPStatus(status int) ErrorCode
func FromGRPCCode(code codes.Code) ErrorCode
```

### 6. Transport - 传输接口

```go
// Transport 传输层接口
type Transport interface {
    String() string
    
    // Server
    Serve(ctx context.Context, endpoints []Endpoint) error
    Shutdown(ctx context.Context) error
    
    // Client
    Invoke(ctx context.Context, endpoint string, req any, resp any) error
    InvokeStream(ctx context.Context, endpoint string, req any) (Stream, error)
    Close() error
}

// Stream 流接口
type Stream interface {
    Send(msg any) error
    Recv(msg any) error
    Close() error
}

// 传输工厂 - 外层
var TransportFactory = factory.NewFactory[Transport, TransportOption]()

// HTTP 实现工厂 - 内层（中间件由各框架自己处理）
var HTTPFactory = factory.NewFactory[HTTPTransport, HTTPOption]()

func init() {
    HTTPFactory.Register("std", NewStdHTTP)    // net/http 中间件
    HTTPFactory.Register("gin", NewGinHTTP)    // Gin 中间件
    HTTPFactory.Register("echo", NewEchoHTTP)  // Echo 中间件
}
```

## 使用示例

### 定义服务

```go
type UserService interface {
    // @talk path=/users/{id} method=GET
    GetUser(ctx context.Context, id string) (*User, error)
    
    // @talk path=/users method=POST
    CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
    
    // 返回 chan，自动识别为服务端流
    // @talk path=/users/{id}/events
    WatchUser(ctx context.Context, id string) (<-chan *UserEvent, error)
}

type userServiceImpl struct{}

func (s *userServiceImpl) GetUser(ctx context.Context, id string) (*User, error) {
    return &User{ID: id, Name: "Alice"}, nil
}

func (s *userServiceImpl) WatchUser(ctx context.Context, id string) (<-chan *UserEvent, error) {
    ch := make(chan *UserEvent)
    go func() {
        defer close(ch)
        for i := 0; i < 10; i++ {
            ch <- &UserEvent{Type: "update", UserID: id}
            time.Sleep(time.Second)
        }
    }()
    return ch, nil
}
```

### Server 端

```go
// 配置
cfg := x.TypedLazyConfig{
    Type: "http",
    Config: json.RawMessage(`{
        "addr": ":8080",
        "implementation": "gin"
    }`),
}

// 创建 Server
server, err := talk.NewServer(cfg,
    talk.WithCodec("json"),
)

// 方式1: 反射注册
server.Register(&userServiceImpl{})

// 方式2: 代码生成（编译时已生成）
server.RegisterEndpoints(userServiceEndpoints...)

// 启动
server.Serve(ctx)
```

### Client 端

```go
cfg := x.TypedLazyConfig{
    Type: "http",
    Config: json.RawMessage(`{
        "addr": "http://localhost:8080"
    }`),
}

client, err := talk.NewClient(cfg)

// 方式1: 直接调用
var user User
err = client.Call(ctx, "GetUser", "123", &user)

// 方式2: 类型安全客户端（反射绑定）
var userSvc UserService
client.Bind(&userSvc)
user, err := userSvc.GetUser(ctx, "123")

// 方式3: 流式调用
stream, err := client.Stream(ctx, "WatchUser", "123")
for {
    var event UserEvent
    if err := stream.Recv(&event); err != nil {
        break
    }
    fmt.Println(event)
}
```

### 切换协议 - 只需改配置

```yaml
# HTTP -> gRPC
server:
  transport:
    type: grpc      # 改这里即可
    config:
      addr: ":9090"

# HTTP -> WebSocket
server:
  transport:
    type: websocket # 改这里即可
    config:
      addr: ":8080"
      path: /ws
```

## 设计决策总结

| 问题 | 决策 |
|------|------|
| 服务注册 | 支持 **反射** 和 **代码生成** 两种方式 |
| Endpoint 定义 | **函数名推导** + **注释声明** (`@talk path=... method=...`) |
| 编解码 | **可插拔**，通过 `CodecFactory` 注册 |
| 中间件 | **各框架自己处理**（Gin 用 Gin MW，gRPC 用 Interceptor） |
| 错误处理 | **统一映射**，ErrorCode ↔ HTTP Status ↔ gRPC Code ↔ WS Code |
| 流式支持 | **chan 自动识别**：参数/返回值含 `<-chan` 即为流式 |

## 实现计划

### Phase 1: 核心框架
- [ ] 定义核心接口 (Endpoint, Transport, Codec, Error)
- [ ] 实现反射提取器 (extract/reflect.go)
- [ ] 实现注释解析器 (extract/comment.go)
- [ ] 实现 JSON Codec
- [ ] 实现统一错误映射

### Phase 2: HTTP 传输
- [ ] 实现 net/http Transport (std)
- [ ] 实现 Gin Transport
- [ ] SSE 流式支持（服务端流）

### Phase 3: 代码生成
- [ ] 注释解析 + Endpoint 生成
- [ ] go generate 工具

### Phase 4: gRPC 传输
- [ ] 实现 gRPC Transport
- [ ] 全部流模式支持

### Phase 5: WebSocket
- [ ] 实现 WebSocket Transport
- [ ] 双向流支持

### Phase 6: 增强
- [ ] 更多 Codec (Protobuf, MsgPack)
- [ ] 服务发现集成
- [ ] 链路追踪支持
