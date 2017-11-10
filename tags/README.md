# http_ctxtags
`import "github.com/improbable-eng/go-httpwares/tags"`

* [Overview](#pkg-overview)
* [Imported Packages](#pkg-imports)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>
`http_ctxtags` adds a Tag object to the request's context that identifies it for other wares.

### Request Context Tags
Tags describe information about the request, and can be set and used by other middleware. Tags are used for logging
and tracing of requests. This extends to both client-side (Tripperware) and server-side (Middleware) libraries.

### Service Tags
`http_ctxtags` introduces a concept of services for client-side calls. This makes it easy to identify outbound requests
to both internal and external services.

For calling external services a `http_ctx.Tripperware()` will automatically try to guess the service name from the URL,
e.g. calling "www.googleapis.com" will yield `googleapis` as service name.

However, for calling internal services, it is recommended to explicitly state the service name by using `WithServiceName("myservice")`
and reusing that particular client for all subsequent calls.

### Handler Names and Groups
For server-side purposes handlers can be named (e.g. "token_exchange") and placed in a group (e.g. "auth"). This allows
easy organisation of HTTP endpoints for logging and monitoring purposes.

See `TagFor*` consts below.

### Custom Tags
You can provide a `WithTagExtractor` function that will populate tags server-side and client-side. Each `http.Request`
passing through the Tripperware/Middleware will be passed through this function and new tags will be added.

Tags fields are typed, and shallow and should follow the OpenTracing semantics convention (be prefixed with `http.`):
<a href="https://github.com/opentracing/specification/blob/master/semantic_conventions.md">https://github.com/opentracing/specification/blob/master/semantic_conventions.md</a>

## <a name="pkg-imports">Imported Packages</a>

- [github.com/improbable-eng/go-httpwares](./..)

## <a name="pkg-index">Index</a>
* [Constants](#pkg-constants)
* [Variables](#pkg-variables)
* [func DefaultServiceNameDetector(req \*http.Request) string](#DefaultServiceNameDetector)
* [func HandlerName(handlerName string) httpwares.Middleware](#HandlerName)
* [func Middleware(handlerGroupName string, opts ...Option) httpwares.Middleware](#Middleware)
* [func Tripperware(opts ...Option) httpwares.Tripperware](#Tripperware)
* [type Option](#Option)
  * [func WithServiceName(serviceName string) Option](#WithServiceName)
  * [func WithServiceNameDetector(fn serviceNameDetectorFunc) Option](#WithServiceNameDetector)
  * [func WithTagExtractor(f RequestTagExtractorFunc) Option](#WithTagExtractor)
* [type RequestTagExtractorFunc](#RequestTagExtractorFunc)
* [type Tags](#Tags)
  * [func ExtractInbound(req \*http.Request) \*Tags](#ExtractInbound)
  * [func ExtractInboundFromCtx(ctx context.Context) \*Tags](#ExtractInboundFromCtx)
  * [func ExtractOutbound(req \*http.Request) \*Tags](#ExtractOutbound)
  * [func ExtractOutboundFromCtx(ctx context.Context) \*Tags](#ExtractOutboundFromCtx)
  * [func (t \*Tags) Has(key string) bool](#Tags.Has)
  * [func (t \*Tags) Set(key string, value interface{}) \*Tags](#Tags.Set)
  * [func (t \*Tags) Values() map[string]interface{}](#Tags.Values)

#### <a name="pkg-files">Package files</a>
[context.go](./context.go) [doc.go](./doc.go) [middleware.go](./middleware.go) [options.go](./options.go) [tripperware.go](./tripperware.go) 

## <a name="pkg-constants">Constants</a>
``` go
const (
    // TagForCallService is a string naming the ctxtag identifying a "service" grouping for an http.Request (e.g. "github")
    TagForCallService = "http.call.service"

    // TagForHandlerGroup is a string naming the ctxtag identifying a name of the grouping of http.Handlers (e.g. auth).
    TagForHandlerGroup = "http.handler.group"
    // TagForHandlerName is a string naming the ctxtag identifying a logical name for the http.Handler (e.g. exchange_token).
    TagForHandlerName = "http.handler.name"
)
```

## <a name="pkg-variables">Variables</a>
``` go
var (
    DefaultServiceName = "unspecified"
)
```

## <a name="DefaultServiceNameDetector">func</a> [DefaultServiceNameDetector](./options.go#L82)
``` go
func DefaultServiceNameDetector(req *http.Request) string
```
DefaultServiceNameDetector is the default detector of services from URLs.

## <a name="HandlerName">func</a> [HandlerName](./middleware.go#L49)
``` go
func HandlerName(handlerName string) httpwares.Middleware
```
HandlerName is a piece of middleware that is meant to be used right around an htpt.Handler that will tag it with a
given service name and method name.

This tag will be used for tracing, logging and monitoring purposes. This *needs* to be set in a chain of
Middleware that has `http_ctxtags.Middleware` before it.

## <a name="Middleware">func</a> [Middleware](./middleware.go#L17)
``` go
func Middleware(handlerGroupName string, opts ...Option) httpwares.Middleware
```
Middleware returns a http.Handler middleware values for request tags.

handlerGroupName specifies a logical name for a group of handlers.

## <a name="Tripperware">func</a> [Tripperware](./tripperware.go#L10)
``` go
func Tripperware(opts ...Option) httpwares.Tripperware
```
Tripperware returns a new client-side ware that injects tags about the request.

## <a name="Option">type</a> [Option](./options.go#L47)
``` go
type Option func(*options)
```

### <a name="WithServiceName">func</a> [WithServiceName](./options.go#L66)
``` go
func WithServiceName(serviceName string) Option
```
WithServiceName is an option for client-side wares that explicitly states the name of the service called.

This option takes precedence over the WithServiceNameDetector values.

For example WithServiceName("github").

### <a name="WithServiceNameDetector">func</a> [WithServiceNameDetector](./options.go#L75)
``` go
func WithServiceNameDetector(fn serviceNameDetectorFunc) Option
```
WithServiceNameDetector allows you to customize the function for automatically detecting the service name from URLs.

By default it uses the `DefaultServiceNameDetector`.

### <a name="WithTagExtractor">func</a> [WithTagExtractor](./options.go#L55)
``` go
func WithTagExtractor(f RequestTagExtractorFunc) Option
```
WithTagExtractor adds another request tag extractor, allowing you to customize what tags get prepopulated from the request.

## <a name="RequestTagExtractorFunc">type</a> [RequestTagExtractorFunc](./options.go#L52)
``` go
type RequestTagExtractorFunc func(req *http.Request) map[string]interface{}
```
RequestTagExtractorFunc is a signature of user-customizeable functions for extracting tags from requests.

## <a name="Tags">type</a> [Tags](./context.go#L23-L25)
``` go
type Tags struct {
    // contains filtered or unexported fields
}
```
Tags is the struct used for storing request tags between Context calls.
This object is *not* thread safe, and should be handled only in the context of the request.

### <a name="ExtractInbound">func</a> [ExtractInbound](./context.go#L47)
``` go
func ExtractInbound(req *http.Request) *Tags
```
ExtractInbound returns a pre-existing Tags object in the request's Context meant for server-side.
If the context wasn't set in the Middleware, a no-op Tag storage is returned that will *not* be propagated in context.

### <a name="ExtractInboundFromCtx">func</a> [ExtractInboundFromCtx](./context.go#L53)
``` go
func ExtractInboundFromCtx(ctx context.Context) *Tags
```
ExtractInbounfFromCtx returns a pre-existing Tags object in the request's Context.
If the context wasn't set in a tag interceptor, a no-op Tag storage is returned that will *not* be propagated in context.

### <a name="ExtractOutbound">func</a> [ExtractOutbound](./context.go#L67)
``` go
func ExtractOutbound(req *http.Request) *Tags
```
ExtractOutbound returns a pre-existing Tags object in the request's Context meant for server-side.
If the context wasn't set in the Middleware, a no-op Tag storage is returned that will *not* be propagated in context.

### <a name="ExtractOutboundFromCtx">func</a> [ExtractOutboundFromCtx](./context.go#L73)
``` go
func ExtractOutboundFromCtx(ctx context.Context) *Tags
```
ExtractInbounfFromCtx returns a pre-existing Tags object in the request's Context.
If the context wasn't set in a tag interceptor, a no-op Tag storage is returned that will *not* be propagated in context.

### <a name="Tags.Has">func</a> (\*Tags) [Has](./context.go#L34)
``` go
func (t *Tags) Has(key string) bool
```
Has checks if the given key exists.

### <a name="Tags.Set">func</a> (\*Tags) [Set](./context.go#L28)
``` go
func (t *Tags) Set(key string, value interface{}) *Tags
```
Set sets the given key in the metadata tags.

### <a name="Tags.Values">func</a> (\*Tags) [Values](./context.go#L41)
``` go
func (t *Tags) Values() map[string]interface{}
```
Values returns a map of key to values.
Do not modify the underlying map, please use Set instead.

- - -
Generated by [godoc2ghmd](https://github.com/GandalfUK/godoc2ghmd)