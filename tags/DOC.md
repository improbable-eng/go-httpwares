# http_ctxtags
--
    import "github.com/mwitkow/go-httpwares/tags"


## Usage

```go
const (
	// TagForCallService is a string naming the ctxtag identifying a "service" grouping for an http.Request (e.g. "github")
	TagForCallService = "http.call.service"
	// TagForCallMethod is a string naming the ctxtag identifying a "method" in a "service" for an http.Request (e.g. "login")
	TagForCallMethod = "http.call.method"

	// TagForHandlerService is a string naming the ctxtag identifying a "service" grouping of http.Handlers (e.g. auth).
	TagForHandlerService = "http.handler.service"
	// TagForHandlerMethod is a string naming the ctxtag identifying a logical "method" of the http.Handler (e.g. exchange_token).
	TagForHandlerMethod = "http.handler.method"
)
```

```go
var (
	DefaultServiceName = "unspecified"
)
```

#### func  ChiRouteTagExtractor

```go
func ChiRouteTagExtractor(req *http.Request) map[string]interface{}
```
ChiRouteTagExtractor extracts chi router information and puts them into tags.

By default it will treat the route pattern as a method name.

#### func  Middleware

```go
func Middleware(opts ...Option) httpwares.Middleware
```
Middleware returns a http.Handler middleware values for request tags.

#### func  TagHandler

```go
func TagHandler(serviceName string, methodName string, handler http.Handler) http.Handler
```
TagHandler is a helper wrapper for http.Handler that will tag it with a given
service name and method name.

This tag will be used for tracing, logging and monitoring purposes. This *needs*
to be set in a chain of Middleware that has `http_ctxtags.Middleware` before it.

You can pass in an empty serviceName, in which case it will inherit it from the
`http_ctxtags.Middleware` configuration.

#### func  TagRequest

```go
func TagRequest(req *http.Request, serviceName string, methodName string) *http.Request
```
TagRequest is a helper that identifies an `http.Request` service name and
method.

This is useful to add "service/method" semantics to your external calls. This
tag will be used for tracing, logging and monitoring purposes. In order for it
to work, the invoked `http.Client` needs to have `http_ctxtags.Tripperware` in
its Roundtripper chain.

You can pass in an empty serviceName, in which case it will inherit it from the
`http_ctxtags.Tripperware` configuration.

It returns a new Request object.

#### func  Tripperware

```go
func Tripperware(opts ...Option) httpwares.Tripperware
```
Tripperware returns a new client-side ware that injects tags about the request.

#### type Option

```go
type Option func(*options)
```


#### func  WithServiceName

```go
func WithServiceName(serviceName string) Option
```
WithServiceName is an option that allows you to track requests to different URL
under the same service name.

For client side requests, you can track external, and internal service names by
using WithServiceName("github"). For server side you can track logical groups of
http.Handlers into a single service.

#### func  WithTagExtractor

```go
func WithTagExtractor(f RequestTagExtractorFunc) Option
```
WithTagExtractor adds another request tag extractor, allowing you to customize
what tags get prepopulated from the request.

#### type RequestTagExtractorFunc

```go
type RequestTagExtractorFunc func(req *http.Request) map[string]interface{}
```

RequestTagExtractorFunc is a signature of user-customizeable functions for
extracting tags from requests.

#### type Tags

```go
type Tags struct {
}
```

Tags is the struct used for storing request tags between Context calls. This
object is *not* thread safe, and should be handled only in the context of the
request.

#### func  ExtractInbound

```go
func ExtractInbound(req *http.Request) *Tags
```
ExtractInbound returns a pre-existing Tags object in the request's Context meant
for server-side. If the context wasn't set in the Middleware, a no-op Tag
storage is returned that will *not* be propagated in context.

#### func  ExtractInboundFromCtx

```go
func ExtractInboundFromCtx(ctx context.Context) *Tags
```
ExtractInbounfFromCtx returns a pre-existing Tags object in the request's
Context. If the context wasn't set in a tag interceptor, a no-op Tag storage is
returned that will *not* be propagated in context.

#### func  ExtractOutbound

```go
func ExtractOutbound(req *http.Request) *Tags
```
ExtractOutbound returns a pre-existing Tags object in the request's Context
meant for server-side. If the context wasn't set in the Middleware, a no-op Tag
storage is returned that will *not* be propagated in context.

#### func  ExtractOutboundFromCtx

```go
func ExtractOutboundFromCtx(ctx context.Context) *Tags
```
ExtractInbounfFromCtx returns a pre-existing Tags object in the request's
Context. If the context wasn't set in a tag interceptor, a no-op Tag storage is
returned that will *not* be propagated in context.

#### func (*Tags) Has

```go
func (t *Tags) Has(key string) bool
```
Has checks if the given key exists.

#### func (*Tags) Set

```go
func (t *Tags) Set(key string, value interface{}) *Tags
```
Set sets the given key in the metadata tags.

#### func (*Tags) Values

```go
func (t *Tags) Values() map[string]interface{}
```
Values returns a map of key to values. Do not modify the underlying map, please
use Set instead.
