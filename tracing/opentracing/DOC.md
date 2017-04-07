# http_opentracing
--
    import "github.com/mwitkow/go-httpwares/tracing/opentracing"


## Usage

```go
const (
	TagTraceId = "trace.traceid"
	TagSpanId  = "trace.spanid"
)
```

#### func  DefaultStatusCodeIsError

```go
func DefaultStatusCodeIsError(statusCode int) bool
```

#### func  Middleware

```go
func Middleware(opts ...Option) httpwares.Middleware
```
Middleware returns a http.Handler middleware values for request tags.

#### func  Tripperware

```go
func Tripperware(opts ...Option) httpwares.Tripperware
```
UnaryClientInterceptor returns a new unary server interceptor for OpenTracing.

#### type FilterFunc

```go
type FilterFunc func(req *http.Request) bool
```

FilterFunc allows users to provide a function that filters out certain methods
from being traced.

If it returns false, the given request will not be traced.

#### type Option

```go
type Option func(*options)
```


#### func  WithFilterFunc

```go
func WithFilterFunc(f FilterFunc) Option
```
WithFilterFunc customizes the function used for deciding whether a given call is
traced or not.

#### func  WithStatusCodeIsError

```go
func WithStatusCodeIsError(f StatusCodeIsError) Option
```
WithStatusCodeIsError customizes the function used for deciding whether a given
call was an error

#### func  WithTracer

```go
func WithTracer(tracer opentracing.Tracer) Option
```
WithTracer sets a custom tracer to be used for this middleware, otherwise the
opentracing.GlobalTracer is used.

#### type StatusCodeIsError

```go
type StatusCodeIsError func(statusCode int) bool
```

StatusCodeIsError allows the customization of which requests are considered
errors in the tracing system.
