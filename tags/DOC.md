# http_ctxtags
--
    import "github.com/mwitkow/go-httpwares/tags"


## Usage

```go
const (
	// TagForCallService is a string naming the ctxtag identifying a "service" grouping for an http.Request (e.g. "github")
	TagForCallService = "http.call.service"

	// TagForHandlerGroup is a string naming the ctxtag identifying a name of the grouping of http.Handlers (e.g. auth).
	TagForHandlerGroup = "http.handler.group"
	// TagForHandlerName is a string naming the ctxtag identifying a logical name for the http.Handler (e.g. exchange_token).
	TagForHandlerName = "http.handler.name"
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

By default it will treat the route pattern as the handler name.

#### func  DefaultServiceNameDetector

```go
func DefaultServiceNameDetector(req *http.Request) string
```
DefaultServiceNameDetector is the default detector of services from URLs.

#### func  HandlerName

```go
func HandlerName(handlerName string) httpwares.Middleware
```
HandlerName is a piece of middleware that is meant to be used right around an
htpt.Handler that will tag it with a given service name and method name.

This tag will be used for tracing, logging and monitoring purposes. This *needs*
to be set in a chain of Middleware that has `http_ctxtags.Middleware` before it.

#### func  Middleware

```go
func Middleware(handlerGroupName string, opts ...Option) httpwares.Middleware
```
Middleware returns a http.Handler middleware values for request tags.

handlerGroupName specifies a logical name for a group of handlers.

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
WithServiceName is an option for client-side wares that explicitly states the
name of the service called.

This option takes precedence over the WithServiceNameDetector values.

For example WithServiceName("github").

#### func  WithServiceNameDetector

```go
func WithServiceNameDetector(fn serviceNameDetectorFunc) Option
```
WithServiceNameDetector allows you to customize the function for automatically
detecting the service name from URLs.

By default it uses the `DefaultServiceNameDetector`.

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
